package domain_test

import (
	"testing"

	"github.com/straye-as/relation-api/internal/domain"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// OfferPhase Tests
// =============================================================================

func TestOfferPhase_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		phase    domain.OfferPhase
		expected bool
	}{
		{"draft is valid", domain.OfferPhaseDraft, true},
		{"in_progress is valid", domain.OfferPhaseInProgress, true},
		{"sent is valid", domain.OfferPhaseSent, true},
		{"order is valid", domain.OfferPhaseOrder, true},
		{"completed is valid", domain.OfferPhaseCompleted, true},
		{"lost is valid", domain.OfferPhaseLost, true},
		{"expired is valid", domain.OfferPhaseExpired, true},
		{"invalid phase", domain.OfferPhase("invalid"), false},
		{"empty phase", domain.OfferPhase(""), false},
		{"active is invalid", domain.OfferPhase("active"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.IsValid())
		})
	}
}

func TestOfferPhase_IsActivePhase(t *testing.T) {
	tests := []struct {
		name     string
		phase    domain.OfferPhase
		expected bool
	}{
		{"draft is not active", domain.OfferPhaseDraft, false},
		{"in_progress is not active", domain.OfferPhaseInProgress, false},
		{"sent is not active", domain.OfferPhaseSent, false},
		{"order IS active", domain.OfferPhaseOrder, true},
		{"completed is not active", domain.OfferPhaseCompleted, false},
		{"lost is not active", domain.OfferPhaseLost, false},
		{"expired is not active", domain.OfferPhaseExpired, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.IsActivePhase())
		})
	}
}

func TestOfferPhase_IsClosedPhase(t *testing.T) {
	tests := []struct {
		name     string
		phase    domain.OfferPhase
		expected bool
	}{
		{"draft is not closed", domain.OfferPhaseDraft, false},
		{"in_progress is not closed", domain.OfferPhaseInProgress, false},
		{"sent is not closed", domain.OfferPhaseSent, false},
		{"order is not closed", domain.OfferPhaseOrder, false},
		{"completed IS closed", domain.OfferPhaseCompleted, true},
		{"lost IS closed", domain.OfferPhaseLost, true},
		{"expired IS closed", domain.OfferPhaseExpired, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.IsClosedPhase())
		})
	}
}

func TestOfferPhase_IsSalesPhase(t *testing.T) {
	tests := []struct {
		name     string
		phase    domain.OfferPhase
		expected bool
	}{
		{"draft IS sales phase", domain.OfferPhaseDraft, true},
		{"in_progress IS sales phase", domain.OfferPhaseInProgress, true},
		{"sent IS sales phase", domain.OfferPhaseSent, true},
		{"order is not sales phase", domain.OfferPhaseOrder, false},
		{"completed is not sales phase", domain.OfferPhaseCompleted, false},
		{"lost is not sales phase", domain.OfferPhaseLost, false},
		{"expired is not sales phase", domain.OfferPhaseExpired, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.IsSalesPhase())
		})
	}
}

func TestOfferPhase_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     domain.OfferPhase
		to       domain.OfferPhase
		expected bool
	}{
		// Same phase transitions (always valid)
		{"draft to draft", domain.OfferPhaseDraft, domain.OfferPhaseDraft, true},
		{"in_progress to in_progress", domain.OfferPhaseInProgress, domain.OfferPhaseInProgress, true},
		{"sent to sent", domain.OfferPhaseSent, domain.OfferPhaseSent, true},
		{"order to order", domain.OfferPhaseOrder, domain.OfferPhaseOrder, true},
		{"completed to completed", domain.OfferPhaseCompleted, domain.OfferPhaseCompleted, true},
		{"lost to lost", domain.OfferPhaseLost, domain.OfferPhaseLost, true},
		{"expired to expired", domain.OfferPhaseExpired, domain.OfferPhaseExpired, true},

		// From draft
		{"draft to in_progress", domain.OfferPhaseDraft, domain.OfferPhaseInProgress, true},
		{"draft to lost", domain.OfferPhaseDraft, domain.OfferPhaseLost, true},
		{"draft to sent (invalid)", domain.OfferPhaseDraft, domain.OfferPhaseSent, false},
		{"draft to order (invalid)", domain.OfferPhaseDraft, domain.OfferPhaseOrder, false},
		{"draft to completed (invalid)", domain.OfferPhaseDraft, domain.OfferPhaseCompleted, false},
		{"draft to expired (invalid)", domain.OfferPhaseDraft, domain.OfferPhaseExpired, false},

		// From in_progress
		{"in_progress to sent", domain.OfferPhaseInProgress, domain.OfferPhaseSent, true},
		{"in_progress to draft", domain.OfferPhaseInProgress, domain.OfferPhaseDraft, true},
		{"in_progress to lost", domain.OfferPhaseInProgress, domain.OfferPhaseLost, true},
		{"in_progress to order (invalid)", domain.OfferPhaseInProgress, domain.OfferPhaseOrder, false},
		{"in_progress to completed (invalid)", domain.OfferPhaseInProgress, domain.OfferPhaseCompleted, false},
		{"in_progress to expired (invalid)", domain.OfferPhaseInProgress, domain.OfferPhaseExpired, false},

		// From sent
		{"sent to order", domain.OfferPhaseSent, domain.OfferPhaseOrder, true},
		{"sent to lost", domain.OfferPhaseSent, domain.OfferPhaseLost, true},
		{"sent to expired", domain.OfferPhaseSent, domain.OfferPhaseExpired, true},
		{"sent to in_progress", domain.OfferPhaseSent, domain.OfferPhaseInProgress, true},
		{"sent to draft (invalid)", domain.OfferPhaseSent, domain.OfferPhaseDraft, false},
		{"sent to completed (invalid)", domain.OfferPhaseSent, domain.OfferPhaseCompleted, false},

		// From order
		{"order to completed", domain.OfferPhaseOrder, domain.OfferPhaseCompleted, true},
		{"order to lost", domain.OfferPhaseOrder, domain.OfferPhaseLost, true},
		{"order to draft (invalid)", domain.OfferPhaseOrder, domain.OfferPhaseDraft, false},
		{"order to in_progress (invalid)", domain.OfferPhaseOrder, domain.OfferPhaseInProgress, false},
		{"order to sent (invalid)", domain.OfferPhaseOrder, domain.OfferPhaseSent, false},
		{"order to expired (invalid)", domain.OfferPhaseOrder, domain.OfferPhaseExpired, false},

		// From completed (can reopen)
		{"completed to order", domain.OfferPhaseCompleted, domain.OfferPhaseOrder, true},
		{"completed to draft (invalid)", domain.OfferPhaseCompleted, domain.OfferPhaseDraft, false},
		{"completed to in_progress (invalid)", domain.OfferPhaseCompleted, domain.OfferPhaseInProgress, false},
		{"completed to sent (invalid)", domain.OfferPhaseCompleted, domain.OfferPhaseSent, false},
		{"completed to lost (invalid)", domain.OfferPhaseCompleted, domain.OfferPhaseLost, false},
		{"completed to expired (invalid)", domain.OfferPhaseCompleted, domain.OfferPhaseExpired, false},

		// From lost (can restart)
		{"lost to draft", domain.OfferPhaseLost, domain.OfferPhaseDraft, true},
		{"lost to in_progress (invalid)", domain.OfferPhaseLost, domain.OfferPhaseInProgress, false},
		{"lost to sent (invalid)", domain.OfferPhaseLost, domain.OfferPhaseSent, false},
		{"lost to order (invalid)", domain.OfferPhaseLost, domain.OfferPhaseOrder, false},
		{"lost to completed (invalid)", domain.OfferPhaseLost, domain.OfferPhaseCompleted, false},
		{"lost to expired (invalid)", domain.OfferPhaseLost, domain.OfferPhaseExpired, false},

		// From expired (can restart)
		{"expired to draft", domain.OfferPhaseExpired, domain.OfferPhaseDraft, true},
		{"expired to in_progress (invalid)", domain.OfferPhaseExpired, domain.OfferPhaseInProgress, false},
		{"expired to sent (invalid)", domain.OfferPhaseExpired, domain.OfferPhaseSent, false},
		{"expired to order (invalid)", domain.OfferPhaseExpired, domain.OfferPhaseOrder, false},
		{"expired to completed (invalid)", domain.OfferPhaseExpired, domain.OfferPhaseCompleted, false},
		{"expired to lost (invalid)", domain.OfferPhaseExpired, domain.OfferPhaseLost, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.from.CanTransitionTo(tt.to))
		})
	}
}

// =============================================================================
// ProjectPhase Tests
// =============================================================================

func TestProjectPhase_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		phase    domain.ProjectPhase
		expected bool
	}{
		{"tilbud is valid", domain.ProjectPhaseTilbud, true},
		{"working is valid", domain.ProjectPhaseWorking, true},
		{"on_hold is valid", domain.ProjectPhaseOnHold, true},
		{"completed is valid", domain.ProjectPhaseCompleted, true},
		{"cancelled is valid", domain.ProjectPhaseCancelled, true},
		{"invalid phase", domain.ProjectPhase("invalid"), false},
		{"empty phase", domain.ProjectPhase(""), false},
		{"active is invalid", domain.ProjectPhase("active"), false},
		{"planning is invalid", domain.ProjectPhase("planning"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.IsValid())
		})
	}
}

func TestProjectPhase_IsEditablePhase(t *testing.T) {
	tests := []struct {
		name     string
		phase    domain.ProjectPhase
		expected bool
	}{
		{"tilbud IS editable", domain.ProjectPhaseTilbud, true},
		{"working IS editable", domain.ProjectPhaseWorking, true},
		{"on_hold IS editable", domain.ProjectPhaseOnHold, true},
		{"completed is not editable", domain.ProjectPhaseCompleted, false},
		{"cancelled is not editable", domain.ProjectPhaseCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.IsEditablePhase())
		})
	}
}

func TestProjectPhase_IsActivePhase(t *testing.T) {
	tests := []struct {
		name     string
		phase    domain.ProjectPhase
		expected bool
	}{
		{"tilbud is not active", domain.ProjectPhaseTilbud, false},
		{"working IS active", domain.ProjectPhaseWorking, true},
		{"on_hold is not active", domain.ProjectPhaseOnHold, false},
		{"completed is not active", domain.ProjectPhaseCompleted, false},
		{"cancelled is not active", domain.ProjectPhaseCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.IsActivePhase())
		})
	}
}

func TestProjectPhase_IsClosedPhase(t *testing.T) {
	tests := []struct {
		name     string
		phase    domain.ProjectPhase
		expected bool
	}{
		{"tilbud is not closed", domain.ProjectPhaseTilbud, false},
		{"working is not closed", domain.ProjectPhaseWorking, false},
		{"on_hold is not closed", domain.ProjectPhaseOnHold, false},
		{"completed IS closed", domain.ProjectPhaseCompleted, true},
		{"cancelled IS closed", domain.ProjectPhaseCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.phase.IsClosedPhase())
		})
	}
}

func TestProjectPhase_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     domain.ProjectPhase
		to       domain.ProjectPhase
		expected bool
	}{
		// Same phase transitions (always valid)
		{"tilbud to tilbud", domain.ProjectPhaseTilbud, domain.ProjectPhaseTilbud, true},
		{"working to working", domain.ProjectPhaseWorking, domain.ProjectPhaseWorking, true},
		{"on_hold to on_hold", domain.ProjectPhaseOnHold, domain.ProjectPhaseOnHold, true},
		{"completed to completed", domain.ProjectPhaseCompleted, domain.ProjectPhaseCompleted, true},
		{"cancelled to cancelled", domain.ProjectPhaseCancelled, domain.ProjectPhaseCancelled, true},

		// From tilbud
		{"tilbud to working", domain.ProjectPhaseTilbud, domain.ProjectPhaseWorking, true},
		{"tilbud to on_hold", domain.ProjectPhaseTilbud, domain.ProjectPhaseOnHold, true},
		{"tilbud to cancelled", domain.ProjectPhaseTilbud, domain.ProjectPhaseCancelled, true},
		{"tilbud to completed (invalid)", domain.ProjectPhaseTilbud, domain.ProjectPhaseCompleted, false},

		// From working
		{"working to on_hold", domain.ProjectPhaseWorking, domain.ProjectPhaseOnHold, true},
		{"working to completed", domain.ProjectPhaseWorking, domain.ProjectPhaseCompleted, true},
		{"working to cancelled", domain.ProjectPhaseWorking, domain.ProjectPhaseCancelled, true},
		{"working to tilbud", domain.ProjectPhaseWorking, domain.ProjectPhaseTilbud, true},

		// From on_hold
		{"on_hold to working", domain.ProjectPhaseOnHold, domain.ProjectPhaseWorking, true},
		{"on_hold to cancelled", domain.ProjectPhaseOnHold, domain.ProjectPhaseCancelled, true},
		{"on_hold to completed", domain.ProjectPhaseOnHold, domain.ProjectPhaseCompleted, true},
		{"on_hold to tilbud (invalid)", domain.ProjectPhaseOnHold, domain.ProjectPhaseTilbud, false},

		// From completed (can reopen)
		{"completed to working", domain.ProjectPhaseCompleted, domain.ProjectPhaseWorking, true},
		{"completed to tilbud (invalid)", domain.ProjectPhaseCompleted, domain.ProjectPhaseTilbud, false},
		{"completed to on_hold (invalid)", domain.ProjectPhaseCompleted, domain.ProjectPhaseOnHold, false},
		{"completed to cancelled (invalid)", domain.ProjectPhaseCompleted, domain.ProjectPhaseCancelled, false},

		// From cancelled (terminal state)
		{"cancelled to tilbud (invalid)", domain.ProjectPhaseCancelled, domain.ProjectPhaseTilbud, false},
		{"cancelled to working (invalid)", domain.ProjectPhaseCancelled, domain.ProjectPhaseWorking, false},
		{"cancelled to on_hold (invalid)", domain.ProjectPhaseCancelled, domain.ProjectPhaseOnHold, false},
		{"cancelled to completed (invalid)", domain.ProjectPhaseCancelled, domain.ProjectPhaseCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.from.CanTransitionTo(tt.to))
		})
	}
}

// =============================================================================
// OfferHealth Tests
// =============================================================================

func TestOfferHealth_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		health   domain.OfferHealth
		expected bool
	}{
		{"on_track is valid", domain.OfferHealthOnTrack, true},
		{"at_risk is valid", domain.OfferHealthAtRisk, true},
		{"delayed is valid", domain.OfferHealthDelayed, true},
		{"over_budget is valid", domain.OfferHealthOverBudget, true},
		{"invalid health", domain.OfferHealth("invalid"), false},
		{"empty health", domain.OfferHealth(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.health.IsValid())
		})
	}
}

// =============================================================================
// Phase Lifecycle Tests (integration-style)
// =============================================================================

func TestOfferPhase_TypicalSalesLifecycle(t *testing.T) {
	// Test a typical successful sales lifecycle: draft -> in_progress -> sent -> order -> completed
	phases := []domain.OfferPhase{
		domain.OfferPhaseDraft,
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
		domain.OfferPhaseOrder,
		domain.OfferPhaseCompleted,
	}

	for i := 0; i < len(phases)-1; i++ {
		from := phases[i]
		to := phases[i+1]
		t.Run(string(from)+" to "+string(to), func(t *testing.T) {
			assert.True(t, from.CanTransitionTo(to),
				"Expected valid transition from %s to %s", from, to)
		})
	}
}

func TestOfferPhase_LostAtAnyStage(t *testing.T) {
	// Test that offers can be marked as lost from any sales phase
	salesPhases := []domain.OfferPhase{
		domain.OfferPhaseDraft,
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
		domain.OfferPhaseOrder,
	}

	for _, phase := range salesPhases {
		t.Run(string(phase)+" to lost", func(t *testing.T) {
			assert.True(t, phase.CanTransitionTo(domain.OfferPhaseLost),
				"Expected %s to transition to lost", phase)
		})
	}
}

func TestOfferPhase_RestartFromTerminal(t *testing.T) {
	// Test that lost and expired offers can be restarted
	terminalPhases := []domain.OfferPhase{
		domain.OfferPhaseLost,
		domain.OfferPhaseExpired,
	}

	for _, phase := range terminalPhases {
		t.Run(string(phase)+" can restart to draft", func(t *testing.T) {
			assert.True(t, phase.CanTransitionTo(domain.OfferPhaseDraft),
				"Expected %s to be restartable to draft", phase)
		})
	}
}

func TestProjectPhase_TypicalLifecycle(t *testing.T) {
	// Test a typical project lifecycle: tilbud -> working -> completed
	phases := []domain.ProjectPhase{
		domain.ProjectPhaseTilbud,
		domain.ProjectPhaseWorking,
		domain.ProjectPhaseCompleted,
	}

	for i := 0; i < len(phases)-1; i++ {
		from := phases[i]
		to := phases[i+1]
		t.Run(string(from)+" to "+string(to), func(t *testing.T) {
			assert.True(t, from.CanTransitionTo(to),
				"Expected valid transition from %s to %s", from, to)
		})
	}
}

func TestProjectPhase_OnHoldAndResume(t *testing.T) {
	// Test putting a project on hold and resuming it
	t.Run("working to on_hold", func(t *testing.T) {
		assert.True(t, domain.ProjectPhaseWorking.CanTransitionTo(domain.ProjectPhaseOnHold))
	})

	t.Run("on_hold to working", func(t *testing.T) {
		assert.True(t, domain.ProjectPhaseOnHold.CanTransitionTo(domain.ProjectPhaseWorking))
	})
}

func TestProjectPhase_CancelledIsTerminal(t *testing.T) {
	// Cancelled is a terminal state - no transitions allowed
	targetPhases := []domain.ProjectPhase{
		domain.ProjectPhaseTilbud,
		domain.ProjectPhaseWorking,
		domain.ProjectPhaseOnHold,
		domain.ProjectPhaseCompleted,
	}

	for _, target := range targetPhases {
		t.Run("cancelled cannot transition to "+string(target), func(t *testing.T) {
			assert.False(t, domain.ProjectPhaseCancelled.CanTransitionTo(target),
				"Cancelled should not transition to %s", target)
		})
	}
}

func TestProjectPhase_CompletedCanReopen(t *testing.T) {
	// Completed projects can be reopened to working
	t.Run("completed to working", func(t *testing.T) {
		assert.True(t, domain.ProjectPhaseCompleted.CanTransitionTo(domain.ProjectPhaseWorking))
	})

	// But not to other phases
	invalidTargets := []domain.ProjectPhase{
		domain.ProjectPhaseTilbud,
		domain.ProjectPhaseOnHold,
		domain.ProjectPhaseCancelled,
	}

	for _, target := range invalidTargets {
		t.Run("completed cannot transition to "+string(target), func(t *testing.T) {
			assert.False(t, domain.ProjectPhaseCompleted.CanTransitionTo(target))
		})
	}
}
