package domain

import (
	"testing"
)

func TestValidType(t *testing.T) {
	valid := []string{TypeApproval, TypeReview, TypeEscalation}
	for _, v := range valid {
		if !ValidType(v) {
			t.Errorf("ValidType(%q) = false, want true", v)
		}
	}
	invalid := []string{"", "unknown", "APPROVAL"}
	for _, v := range invalid {
		if ValidType(v) {
			t.Errorf("ValidType(%q) = true, want false", v)
		}
	}
}

func TestValidStatus(t *testing.T) {
	valid := []string{StatusPending, StatusActive, StatusApproved, StatusRejected, StatusEscalated, StatusCancelled}
	for _, v := range valid {
		if !ValidStatus(v) {
			t.Errorf("ValidStatus(%q) = false, want true", v)
		}
	}
	invalid := []string{"", "done", "PENDING", "completed"}
	for _, v := range invalid {
		if ValidStatus(v) {
			t.Errorf("ValidStatus(%q) = true, want false", v)
		}
	}
}

func TestValidAction(t *testing.T) {
	valid := []string{ActionApprove, ActionReject, ActionEscalate}
	for _, v := range valid {
		if !ValidAction(v) {
			t.Errorf("ValidAction(%q) = false, want true", v)
		}
	}
	invalid := []string{"", "cancel", "APPROVE", "complete"}
	for _, v := range invalid {
		if ValidAction(v) {
			t.Errorf("ValidAction(%q) = true, want false", v)
		}
	}
}

func TestWorkflowInstanceStepProgression(t *testing.T) {
	instance := &WorkflowInstance{
		CurrentStep: 1,
		Status:      StatusActive,
	}

	instance.CurrentStep++
	if instance.CurrentStep != 2 {
		t.Errorf("CurrentStep = %d, want 2", instance.CurrentStep)
	}
}
