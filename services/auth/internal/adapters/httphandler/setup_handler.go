package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/auth/internal/application"
)

// SetupHandler handles initial system setup HTTP endpoints.
type SetupHandler struct {
	setupService *application.SetupService
	authService  *application.AuthService
	logger       zerolog.Logger
}

// NewSetupHandler creates a new SetupHandler.
func NewSetupHandler(
	setupService *application.SetupService,
	authService *application.AuthService,
	logger zerolog.Logger,
) *SetupHandler {
	return &SetupHandler{
		setupService: setupService,
		authService:  authService,
		logger:       logger.With().Str("handler", "setup").Logger(),
	}
}

// Status returns the current setup state (complete or not).
func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]interface{}{
		"is_complete": true,
		"version":     "1.0.0",
	})
}

// Init checks whether the system is ready for initial setup.
func (h *SetupHandler) Init(w http.ResponseWriter, r *http.Request) {
	var req application.SetupInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body — Init may be called without parameters.
		req = application.SetupInitRequest{}
	}

	ready, err := h.setupService.Init(r.Context(), req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, map[string]bool{"ready": ready})
}

// Complete performs the initial system setup and returns tenant, user, and tokens.
func (h *SetupHandler) Complete(w http.ResponseWriter, r *http.Request) {
	var req application.SetupCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if req.Company.Name == "" || req.Company.Slug == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest,
			"company.name und company.slug sind erforderlich"))
		return
	}
	if req.Admin.Email == "" || req.Admin.Password == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest,
			"admin.email und admin.password sind erforderlich"))
		return
	}

	tenant, user, err := h.setupService.Complete(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("setup complete failed")
		errors.HandleError(w, err)
		return
	}

	// Generate token pair for the new admin user via login.
	loginReq := application.LoginRequest{
		Email:    req.Admin.Email,
		Password: req.Admin.Password,
	}

	tokens, _, loginErr := h.authService.Login(
		r.Context(), tenant.ID, loginReq, r.RemoteAddr, r.UserAgent(),
	)
	if loginErr != nil {
		// Setup succeeded but auto-login failed — still return the setup data.
		h.logger.Warn().Err(loginErr).Msg("auto-login after setup failed")
		response.Created(w, map[string]interface{}{
			"tenant": tenant,
			"user":   toUserResponse(user),
		})
		return
	}

	response.Created(w, map[string]interface{}{
		"tenant": tenant,
		"user":   toUserResponse(user),
		"tokens": tokens,
	})
}
