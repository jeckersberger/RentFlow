package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/crew/internal/domain"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockSkillMatchRepo struct {
	findBySkillFn      func(ctx context.Context, tenantID uuid.UUID, skill string) ([]domain.SkillMatchRow, error)
	blockedMemberIDsFn func(ctx context.Context, tenantID uuid.UUID, from, to string) (map[uuid.UUID]bool, error)
	assignedMemberIDsFn func(ctx context.Context, tenantID uuid.UUID, from, to string, excludeProjectID *uuid.UUID) (map[uuid.UUID]bool, error)
}

func (m *mockSkillMatchRepo) FindBySkill(ctx context.Context, tenantID uuid.UUID, skill string) ([]domain.SkillMatchRow, error) {
	if m.findBySkillFn != nil {
		return m.findBySkillFn(ctx, tenantID, skill)
	}
	return nil, nil
}

func (m *mockSkillMatchRepo) BlockedMemberIDs(ctx context.Context, tenantID uuid.UUID, from, to string) (map[uuid.UUID]bool, error) {
	if m.blockedMemberIDsFn != nil {
		return m.blockedMemberIDsFn(ctx, tenantID, from, to)
	}
	return map[uuid.UUID]bool{}, nil
}

func (m *mockSkillMatchRepo) AssignedMemberIDs(ctx context.Context, tenantID uuid.UUID, from, to string, excludeProjectID *uuid.UUID) (map[uuid.UUID]bool, error) {
	if m.assignedMemberIDsFn != nil {
		return m.assignedMemberIDsFn(ctx, tenantID, from, to, excludeProjectID)
	}
	return map[uuid.UUID]bool{}, nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func newTestService(repo *mockSkillMatchRepo) *SkillMatchService {
	logger := zerolog.Nop()
	return NewSkillMatchService(repo, logger)
}

func TestMatchSkills_EmptyRequiredSkills(t *testing.T) {
	svc := newTestService(&mockSkillMatchRepo{})

	_, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: nil,
		DateFrom:       "2026-04-10",
		DateTo:         "2026-04-12",
	})
	if !errors.Is(err, domain.ErrNoSkillsRequested) {
		t.Errorf("expected ErrNoSkillsRequested, got: %v", err)
	}
}

func TestMatchSkills_InvalidCount(t *testing.T) {
	svc := newTestService(&mockSkillMatchRepo{})

	_, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Rigger", Count: 0},
		},
		DateFrom: "2026-04-10",
		DateTo:   "2026-04-12",
	})
	if !errors.Is(err, domain.ErrInvalidSkillCount) {
		t.Errorf("expected ErrInvalidSkillCount, got: %v", err)
	}
}

func TestMatchSkills_MissingDateFrom(t *testing.T) {
	svc := newTestService(&mockSkillMatchRepo{})

	_, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Rigger", Count: 1},
		},
		DateFrom: "",
		DateTo:   "2026-04-12",
	})
	if !errors.Is(err, domain.ErrDateFromRequired) {
		t.Errorf("expected ErrDateFromRequired, got: %v", err)
	}
}

func TestMatchSkills_MissingDateTo(t *testing.T) {
	svc := newTestService(&mockSkillMatchRepo{})

	_, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Rigger", Count: 1},
		},
		DateFrom: "2026-04-10",
		DateTo:   "",
	})
	if !errors.Is(err, domain.ErrDateToRequired) {
		t.Errorf("expected ErrDateToRequired, got: %v", err)
	}
}

func TestMatchSkills_EndBeforeStart(t *testing.T) {
	svc := newTestService(&mockSkillMatchRepo{})

	_, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Rigger", Count: 1},
		},
		DateFrom: "2026-04-15",
		DateTo:   "2026-04-10",
	})
	if !errors.Is(err, domain.ErrEndBeforeStart) {
		t.Errorf("expected ErrEndBeforeStart, got: %v", err)
	}
}

func TestMatchSkills_InvalidDateFormat(t *testing.T) {
	svc := newTestService(&mockSkillMatchRepo{})

	_, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Rigger", Count: 1},
		},
		DateFrom: "10.04.2026",
		DateTo:   "12.04.2026",
	})
	if !errors.Is(err, domain.ErrInvalidDateFormat) {
		t.Errorf("expected ErrInvalidDateFormat, got: %v", err)
	}
}

func TestMatchSkills_HappyPath_AllAvailable(t *testing.T) {
	memberA := uuid.New()
	memberB := uuid.New()

	repo := &mockSkillMatchRepo{
		findBySkillFn: func(_ context.Context, _ uuid.UUID, skill string) ([]domain.SkillMatchRow, error) {
			if skill == "Tontechniker" {
				return []domain.SkillMatchRow{
					{CrewMemberID: memberA, FirstName: "Max", LastName: "Mueller", ExactMatch: true},
					{CrewMemberID: memberB, FirstName: "Lisa", LastName: "Schmidt", ExactMatch: true},
				}, nil
			}
			return nil, nil
		},
	}

	svc := newTestService(repo)
	resp, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Tontechniker", Count: 2},
		},
		DateFrom: "2026-04-10",
		DateTo:   "2026-04-12",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches) != 1 {
		t.Fatalf("Matches length = %d, want 1", len(resp.Matches))
	}
	if resp.Matches[0].Skill != "Tontechniker" {
		t.Errorf("Skill = %q, want %q", resp.Matches[0].Skill, "Tontechniker")
	}
	if resp.Matches[0].Required != 2 {
		t.Errorf("Required = %d, want 2", resp.Matches[0].Required)
	}
	if len(resp.Matches[0].Matched) != 2 {
		t.Fatalf("Matched length = %d, want 2", len(resp.Matches[0].Matched))
	}
	if resp.Matches[0].Matched[0].Score != 1.0 {
		t.Errorf("Score = %f, want 1.0", resp.Matches[0].Matched[0].Score)
	}
	if !resp.Matches[0].Matched[0].Available {
		t.Error("candidate should be available")
	}
	if len(resp.UnmatchedSkills) != 0 {
		t.Errorf("UnmatchedSkills = %v, want empty", resp.UnmatchedSkills)
	}
}

func TestMatchSkills_BlockedMember(t *testing.T) {
	memberA := uuid.New()

	repo := &mockSkillMatchRepo{
		findBySkillFn: func(_ context.Context, _ uuid.UUID, _ string) ([]domain.SkillMatchRow, error) {
			return []domain.SkillMatchRow{
				{CrewMemberID: memberA, FirstName: "Tom", LastName: "Weber", ExactMatch: true},
			}, nil
		},
		blockedMemberIDsFn: func(_ context.Context, _ uuid.UUID, _, _ string) (map[uuid.UUID]bool, error) {
			return map[uuid.UUID]bool{memberA: true}, nil
		},
	}

	svc := newTestService(repo)
	resp, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Rigger", Count: 1},
		},
		DateFrom: "2026-04-10",
		DateTo:   "2026-04-12",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches[0].Matched) != 1 {
		t.Fatalf("Matched length = %d, want 1", len(resp.Matches[0].Matched))
	}
	if resp.Matches[0].Matched[0].Available {
		t.Error("candidate should NOT be available (blocked)")
	}
	if resp.Matches[0].Matched[0].Score != 0.0 {
		t.Errorf("Score = %f, want 0.0 (blocked)", resp.Matches[0].Matched[0].Score)
	}
	if len(resp.UnmatchedSkills) != 1 {
		t.Errorf("UnmatchedSkills length = %d, want 1", len(resp.UnmatchedSkills))
	}
}

func TestMatchSkills_AssignedMember(t *testing.T) {
	memberA := uuid.New()

	repo := &mockSkillMatchRepo{
		findBySkillFn: func(_ context.Context, _ uuid.UUID, _ string) ([]domain.SkillMatchRow, error) {
			return []domain.SkillMatchRow{
				{CrewMemberID: memberA, FirstName: "Anna", LastName: "Braun", ExactMatch: true},
			}, nil
		},
		assignedMemberIDsFn: func(_ context.Context, _ uuid.UUID, _, _ string, _ *uuid.UUID) (map[uuid.UUID]bool, error) {
			return map[uuid.UUID]bool{memberA: true}, nil
		},
	}

	svc := newTestService(repo)
	resp, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Lichtdesigner", Count: 1},
		},
		DateFrom: "2026-04-10",
		DateTo:   "2026-04-12",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Matches[0].Matched[0].Available {
		t.Error("candidate should NOT be available (assigned)")
	}
	if resp.Matches[0].Matched[0].Score != 0.0 {
		t.Errorf("Score = %f, want 0.0 (assigned)", resp.Matches[0].Matched[0].Score)
	}
}

func TestMatchSkills_PartialMatch(t *testing.T) {
	memberA := uuid.New()

	repo := &mockSkillMatchRepo{
		findBySkillFn: func(_ context.Context, _ uuid.UUID, _ string) ([]domain.SkillMatchRow, error) {
			return []domain.SkillMatchRow{
				{CrewMemberID: memberA, FirstName: "Jan", LastName: "Klein", ExactMatch: false},
			}, nil
		},
	}

	svc := newTestService(repo)
	resp, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Ton", Count: 1},
		},
		DateFrom: "2026-04-10",
		DateTo:   "2026-04-12",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Matches[0].Matched[0].Score != 0.8 {
		t.Errorf("Score = %f, want 0.8 (partial match)", resp.Matches[0].Matched[0].Score)
	}
}

func TestMatchSkills_NoMatchesForSkill(t *testing.T) {
	repo := &mockSkillMatchRepo{
		findBySkillFn: func(_ context.Context, _ uuid.UUID, _ string) ([]domain.SkillMatchRow, error) {
			return nil, nil
		},
	}

	svc := newTestService(repo)
	resp, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Pyrotechniker", Count: 1},
		},
		DateFrom: "2026-04-10",
		DateTo:   "2026-04-12",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches[0].Matched) != 0 {
		t.Errorf("Matched length = %d, want 0", len(resp.Matches[0].Matched))
	}
	if len(resp.UnmatchedSkills) != 1 {
		t.Fatalf("UnmatchedSkills length = %d, want 1", len(resp.UnmatchedSkills))
	}
	if resp.UnmatchedSkills[0] != "Pyrotechniker" {
		t.Errorf("UnmatchedSkills[0] = %q, want %q", resp.UnmatchedSkills[0], "Pyrotechniker")
	}
}

func TestMatchSkills_MultipleSkills(t *testing.T) {
	memberA := uuid.New()
	memberB := uuid.New()
	memberC := uuid.New()

	repo := &mockSkillMatchRepo{
		findBySkillFn: func(_ context.Context, _ uuid.UUID, skill string) ([]domain.SkillMatchRow, error) {
			switch skill {
			case "Tontechniker":
				return []domain.SkillMatchRow{
					{CrewMemberID: memberA, FirstName: "Max", LastName: "Mueller", ExactMatch: true},
					{CrewMemberID: memberB, FirstName: "Lisa", LastName: "Schmidt", ExactMatch: true},
				}, nil
			case "Rigger":
				return []domain.SkillMatchRow{
					{CrewMemberID: memberC, FirstName: "Tom", LastName: "Weber", ExactMatch: true},
				}, nil
			}
			return nil, nil
		},
	}

	svc := newTestService(repo)
	resp, err := svc.MatchSkills(context.Background(), uuid.New(), MatchSkillsRequest{
		RequiredSkills: []domain.SkillRequirement{
			{Skill: "Tontechniker", Count: 2},
			{Skill: "Rigger", Count: 1},
		},
		DateFrom: "2026-04-10",
		DateTo:   "2026-04-12",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Matches) != 2 {
		t.Fatalf("Matches length = %d, want 2", len(resp.Matches))
	}
	if resp.Matches[0].Skill != "Tontechniker" {
		t.Errorf("Matches[0].Skill = %q", resp.Matches[0].Skill)
	}
	if len(resp.Matches[0].Matched) != 2 {
		t.Errorf("Tontechniker matched = %d, want 2", len(resp.Matches[0].Matched))
	}
	if resp.Matches[1].Skill != "Rigger" {
		t.Errorf("Matches[1].Skill = %q", resp.Matches[1].Skill)
	}
	if len(resp.Matches[1].Matched) != 1 {
		t.Errorf("Rigger matched = %d, want 1", len(resp.Matches[1].Matched))
	}
	if len(resp.UnmatchedSkills) != 0 {
		t.Errorf("UnmatchedSkills = %v, want empty", resp.UnmatchedSkills)
	}
}
