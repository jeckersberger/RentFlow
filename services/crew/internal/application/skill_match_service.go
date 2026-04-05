package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/crew/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTO
// ---------------------------------------------------------------------------

// MatchSkillsRequest holds the incoming payload for the skill-matching endpoint.
type MatchSkillsRequest struct {
	RequiredSkills []domain.SkillRequirement `json:"required_skills"`
	DateFrom       string                    `json:"date_from"`
	DateTo         string                    `json:"date_to"`
	ProjectID      *string                   `json:"project_id,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// SkillMatchService implements the skill-matching use case.
type SkillMatchService struct {
	skillRepo domain.SkillMatchRepository
	logger    zerolog.Logger
}

// NewSkillMatchService constructs a new SkillMatchService.
func NewSkillMatchService(
	skillRepo domain.SkillMatchRepository,
	logger zerolog.Logger,
) *SkillMatchService {
	return &SkillMatchService{
		skillRepo: skillRepo,
		logger:    logger.With().Str("service", "skill_match").Logger(),
	}
}

// MatchSkills finds crew members matching the requested skills and checks
// their availability for the given date range.
func (s *SkillMatchService) MatchSkills(
	ctx context.Context,
	tenantID uuid.UUID,
	req MatchSkillsRequest,
) (*domain.SkillMatchResponse, error) {
	// --- Validation ---
	if len(req.RequiredSkills) == 0 {
		return nil, domain.ErrNoSkillsRequested
	}
	for _, rs := range req.RequiredSkills {
		if rs.Count < 1 {
			return nil, domain.ErrInvalidSkillCount
		}
	}

	if req.DateFrom == "" {
		return nil, domain.ErrDateFromRequired
	}
	if req.DateTo == "" {
		return nil, domain.ErrDateToRequired
	}

	dateFrom, err := time.Parse("2006-01-02", req.DateFrom)
	if err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	dateTo, err := time.Parse("2006-01-02", req.DateTo)
	if err != nil {
		return nil, domain.ErrInvalidDateFormat
	}
	if dateTo.Before(dateFrom) {
		return nil, domain.ErrEndBeforeStart
	}

	// --- Parse optional project ID ---
	var excludeProjectID *uuid.UUID
	if req.ProjectID != nil && *req.ProjectID != "" {
		pid, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("invalid project_id: %w", err)
		}
		excludeProjectID = &pid
	}

	// --- Fetch unavailability sets (blocks + assignments) ---
	blockedIDs, err := s.skillRepo.BlockedMemberIDs(ctx, tenantID, req.DateFrom, req.DateTo)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to fetch blocked member IDs")
		return nil, fmt.Errorf("match skills - blocked lookup: %w", err)
	}

	assignedIDs, err := s.skillRepo.AssignedMemberIDs(ctx, tenantID, req.DateFrom, req.DateTo, excludeProjectID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to fetch assigned member IDs")
		return nil, fmt.Errorf("match skills - assigned lookup: %w", err)
	}

	// --- Match each required skill ---
	var results []domain.SkillMatchResult
	var unmatchedSkills []string

	for _, rs := range req.RequiredSkills {
		rows, err := s.skillRepo.FindBySkill(ctx, tenantID, rs.Skill)
		if err != nil {
			s.logger.Error().Err(err).
				Str("skill", rs.Skill).
				Msg("failed to find members by skill")
			return nil, fmt.Errorf("match skills - find %q: %w", rs.Skill, err)
		}

		var candidates []domain.SkillMatchCandidate
		for _, row := range rows {
			available := !blockedIDs[row.CrewMemberID] && !assignedIDs[row.CrewMemberID]

			var score float64
			switch {
			case row.ExactMatch && available:
				score = 1.0
			case row.ExactMatch && !available:
				score = 0.0
			case !row.ExactMatch && available:
				score = 0.8
			default:
				score = 0.0
			}

			candidates = append(candidates, domain.SkillMatchCandidate{
				CrewMemberID: row.CrewMemberID,
				Name:         row.FirstName + " " + row.LastName,
				Score:        score,
				Available:    available,
			})
		}

		result := domain.SkillMatchResult{
			Skill:    rs.Skill,
			Required: rs.Count,
			Matched:  candidates,
		}
		if result.Matched == nil {
			result.Matched = []domain.SkillMatchCandidate{}
		}
		results = append(results, result)

		// Count how many available candidates we have for this skill.
		availableCount := 0
		for _, c := range candidates {
			if c.Available {
				availableCount++
			}
		}
		if availableCount < rs.Count {
			unmatchedSkills = append(unmatchedSkills, rs.Skill)
		}
	}

	if unmatchedSkills == nil {
		unmatchedSkills = []string{}
	}

	s.logger.Info().
		Str("tenant_id", tenantID.String()).
		Int("skills_requested", len(req.RequiredSkills)).
		Int("unmatched", len(unmatchedSkills)).
		Str("date_from", req.DateFrom).
		Str("date_to", req.DateTo).
		Msg("skill matching completed")

	return &domain.SkillMatchResponse{
		Matches:         results,
		UnmatchedSkills: unmatchedSkills,
	}, nil
}
