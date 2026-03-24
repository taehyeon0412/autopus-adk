package sigmap

// GoExtractor implements Extractor for Go source files by delegating to the
// existing Extract function.
type GoExtractor struct{}

// Language returns "go".
func (g *GoExtractor) Language() string {
	return "go"
}

// Extract extracts exported Go signatures from the given directory.
func (g *GoExtractor) Extract(dir string) (*SignatureMap, error) {
	return Extract(dir)
}
