package routing

import "log"

// Router determines which model to use based on message complexity and provider.
type Router struct {
	config     RoutingConfig
	classifier *MessageComplexityClassifier
}

// NewRouter creates a Router with the given configuration.
func NewRouter(config RoutingConfig) *Router {
	return &Router{
		config:     config,
		classifier: NewClassifier(config.Thresholds),
	}
}

// Route returns the model name to use for the given provider and message.
// Returns empty string if routing is disabled (REQ-ROUTE-06) or provider not configured.
func (r *Router) Route(provider, message string) string {
	if !r.config.Enabled {
		return ""
	}

	complexity, signals := r.classifier.Classify(message)

	// REQ-ROUTE-07: Log routing decision.
	log.Printf("[routing] provider=%s complexity=%s chars=%d code_blocks=%v keywords=%v score=%d",
		provider, complexity, signals.CharCount, signals.HasCodeBlocks, signals.Keywords, signals.Score)

	models, ok := r.config.Models[provider]
	if !ok {
		log.Printf("[routing] no model mapping for provider %q, using default", provider)
		return ""
	}

	switch complexity {
	case ComplexitySimple:
		return models.Simple
	case ComplexityMedium:
		return models.Medium
	case ComplexityComplex:
		return models.Complex
	default:
		return models.Medium
	}
}
