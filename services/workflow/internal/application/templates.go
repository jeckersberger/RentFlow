package application

import (
	"encoding/json"

	"github.com/jeckersberger/EquipFlow/services/workflow/internal/domain"
)

// WorkflowTemplate represents a pre-built workflow template that can be
// instantiated into a full WorkflowDefinition for a tenant.
type WorkflowTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	TriggerType string `json:"trigger_type"`
	TriggerInfo string `json:"trigger_info"`
	Steps       json.RawMessage `json:"steps"`
}

// GetDefaultTemplates returns the set of pre-built workflow templates.
func GetDefaultTemplates() []WorkflowTemplate {
	return []WorkflowTemplate{
		{
			Name:        "Rechnung bei Projektabschluss",
			Description: "Erstellt automatisch eine Rechnung, wenn ein Projekt abgeschlossen wird.",
			Type:        domain.TypeApproval,
			TriggerType: "event",
			TriggerInfo: "project.completed",
			Steps: json.RawMessage(`[
				{"step": 1, "name": "Rechnungsdaten pruefen", "assignee": "manager", "action": "approve"},
				{"step": 2, "name": "Rechnung versenden", "assignee": "system", "action": "approve"}
			]`),
		},
		{
			Name:        "Mahnlauf taeglich",
			Description: "Taeglich um 08:00 Uhr ueberfaellige Rechnungen pruefen und Mahnungen versenden.",
			Type:        domain.TypeEscalation,
			TriggerType: "cron",
			TriggerInfo: "0 8 * * *",
			Steps: json.RawMessage(`[
				{"step": 1, "name": "Ueberfaellige Rechnungen ermitteln", "assignee": "system", "action": "approve"},
				{"step": 2, "name": "Mahnung versenden", "assignee": "system", "action": "approve"},
				{"step": 3, "name": "Eskalation bei 3. Mahnung", "assignee": "manager", "action": "escalate"}
			]`),
		},
		{
			Name:        "E-Check Erinnerung",
			Description: "Woechentlich montags um 09:00 Uhr Equipment auf faellige E-Checks pruefen.",
			Type:        domain.TypeReview,
			TriggerType: "cron",
			TriggerInfo: "0 9 * * 1",
			Steps: json.RawMessage(`[
				{"step": 1, "name": "Faellige E-Checks ermitteln", "assignee": "system", "action": "approve"},
				{"step": 2, "name": "Techniker benachrichtigen", "assignee": "system", "action": "approve"},
				{"step": 3, "name": "E-Check durchfuehren", "assignee": "technician", "action": "approve"}
			]`),
		},
		{
			Name:        "Willkommens-E-Mail",
			Description: "Sendet eine Willkommens-E-Mail, wenn ein neuer Benutzer erstellt wird.",
			Type:        domain.TypeApproval,
			TriggerType: "event",
			TriggerInfo: "user.created",
			Steps: json.RawMessage(`[
				{"step": 1, "name": "Willkommens-E-Mail senden", "assignee": "system", "action": "approve"}
			]`),
		},
		{
			Name:        "Projekt-Erinnerung 7 Tage",
			Description: "Taeglich um 08:00 Uhr pruefen, ob Projekte in 7 Tagen starten und erinnern.",
			Type:        domain.TypeReview,
			TriggerType: "cron",
			TriggerInfo: "0 8 * * *",
			Steps: json.RawMessage(`[
				{"step": 1, "name": "Projekte mit Start in 7 Tagen ermitteln", "assignee": "system", "action": "approve"},
				{"step": 2, "name": "Projektleiter benachrichtigen", "assignee": "system", "action": "approve"},
				{"step": 3, "name": "Equipment-Verfuegbarkeit pruefen", "assignee": "warehouse", "action": "approve"}
			]`),
		},
	}
}
