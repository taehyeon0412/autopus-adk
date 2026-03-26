package pipeline

import (
	"context"
	"fmt"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// LayoutPlan describes the sequential split strategy for team panes.
// The initial pane is the dashboard; subsequent splits create one pane per teammate.
type LayoutPlan struct {
	// Roles lists teammate roles in split order.
	Roles []string
	// SplitCount is the number of Vertical splits needed (== len(Roles)).
	SplitCount int
}

// planLayout creates a LayoutPlan from the given teammate list.
// Dashboard uses the initial pane; each teammate gets one Vertical split.
func planLayout(teammates []string) LayoutPlan {
	return LayoutPlan{
		Roles:      teammates,
		SplitCount: len(teammates),
	}
}

// LayoutResult holds the pane IDs created by applyLayout.
type LayoutResult struct {
	// DashboardPaneID is the initial pane (not split-created).
	DashboardPaneID terminal.PaneID
	// TeammatePaneIDs maps each role to its pane ID, in order.
	TeammatePaneIDs []terminal.PaneID
}

// @AX:NOTE [AUTO] @AX:REASON: design choice — uses Vertical split per SPEC-TEAMPANE-001 R4; all teammates are stacked vertically below dashboard
// applyLayout executes the layout plan by performing sequential Vertical splits.
// Each SplitPane call produces a vertical stack layout.
//
// Layout for 3 teammates:
//
//	+---------------------------+
//	| Dashboard (initial pane)  |
//	+---------------------------+
//	| Lead                      |
//	+---------------------------+
//	| Builder                   |
//	+---------------------------+
//	| Guardian                  |
//	+---------------------------+
//
// On SplitPane failure, already-created panes are cleaned up (S9).
func applyLayout(ctx context.Context, term terminal.Terminal, plan LayoutPlan) (*LayoutResult, error) {
	if plan.SplitCount == 0 {
		return &LayoutResult{}, nil
	}

	paneIDs := make([]terminal.PaneID, 0, plan.SplitCount)

	for i := 0; i < plan.SplitCount; i++ {
		paneID, err := term.SplitPane(ctx, terminal.Vertical)
		if err != nil {
			// S9: cleanup already-created panes on failure.
			cleanupLayoutPanes(ctx, term, paneIDs)
			return nil, fmt.Errorf("split pane for %s (split %d/%d): %w",
				plan.Roles[i], i+1, plan.SplitCount, err)
		}
		paneIDs = append(paneIDs, paneID)
	}

	return &LayoutResult{
		TeammatePaneIDs: paneIDs,
	}, nil
}

// cleanupLayoutPanes closes panes that were created before a split failure.
func cleanupLayoutPanes(ctx context.Context, term terminal.Terminal, paneIDs []terminal.PaneID) {
	for _, id := range paneIDs {
		_ = term.Close(ctx, string(id))
	}
}
