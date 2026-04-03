package learn

import (
	"sort"
	"strings"
	"time"
)

const (
	fileMatchWeight    = 3.0
	packageMatchWeight = 2.0
	keywordMatchWeight = 1.0
	recencyBonusDays   = 30
	recencyBonus       = 1.0
)

// MatchRelevance scores how relevant an entry is to the given query.
// Returns 0.0 if no match.
func MatchRelevance(entry LearningEntry, query RelevanceQuery) float64 {
	if len(query.Files) == 0 && len(query.Packages) == 0 && len(query.Keywords) == 0 {
		return 0.0
	}

	var score float64

	// File path matching
	for _, qf := range query.Files {
		for _, ef := range entry.Files {
			if qf == ef {
				score += fileMatchWeight
			}
		}
	}

	// Package matching
	for _, qp := range query.Packages {
		for _, ep := range entry.Packages {
			if qp == ep {
				score += packageMatchWeight
			}
		}
	}

	// Keyword matching in Pattern and Resolution
	combined := strings.ToLower(entry.Pattern + " " + entry.Resolution)
	for _, kw := range query.Keywords {
		if strings.Contains(combined, strings.ToLower(kw)) {
			score += keywordMatchWeight
		}
	}

	if score == 0.0 {
		return 0.0
	}

	// Recency bonus: entries within recencyBonusDays get a boost
	if !entry.Timestamp.IsZero() {
		age := time.Since(entry.Timestamp)
		if age < time.Duration(recencyBonusDays)*24*time.Hour {
			score += recencyBonus
		}
	}

	return score
}

// QueryRelevant returns entries with relevance score above threshold,
// sorted by score descending.
func QueryRelevant(store *Store, query RelevanceQuery, threshold float64) ([]LearningEntry, error) {
	entries, err := store.Read()
	if err != nil {
		return nil, err
	}

	type scored struct {
		entry LearningEntry
		score float64
	}
	var results []scored
	for _, e := range entries {
		s := MatchRelevance(e, query)
		if s >= threshold {
			results = append(results, scored{e, s})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	out := make([]LearningEntry, len(results))
	for i, r := range results {
		out[i] = r.entry
	}
	return out, nil
}
