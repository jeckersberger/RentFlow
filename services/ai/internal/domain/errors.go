package domain

import "errors"

var (
	ErrPredictionNotFound = errors.New("ai prediction not found")
	ErrSuggestionNotFound = errors.New("ai suggestion not found")
	ErrTypeRequired       = errors.New("type is required")
	ErrTitleRequired      = errors.New("title is required")
	ErrAIDisabled         = errors.New("KI-Funktionen sind deaktiviert")
	ErrOllamaUnavailable  = errors.New("Ollama-Server nicht erreichbar")
	ErrNoCandidates       = errors.New("candidates list is empty")
	ErrProjectIDRequired  = errors.New("project_id is required")
)
