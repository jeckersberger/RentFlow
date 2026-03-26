package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/pkg/common/middleware"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
)

// IndustryProfile represents an industry profile definition
type IndustryProfile struct {
	ID               string            `json:"id"`
	Label            string            `json:"label"`
	Group            string            `json:"group"`
	Labels           map[string]string `json:"labels"`
	Features         map[string]bool   `json:"features"`
	ScannerHomeCards []string          `json:"scanner_home_cards"`
	Categories       []string          `json:"categories"`
}

// IndustryHandlers holds references to industry-related handlers
type IndustryHandlers struct {
	configService *application.ConfigService
	logger        logger.Logger
}

// NewIndustryHandlers creates a new industry handlers instance
func NewIndustryHandlers(configService *application.ConfigService, log logger.Logger) *IndustryHandlers {
	return &IndustryHandlers{
		configService: configService,
		logger:        log,
	}
}

// GetIndustryProfile handles GET /api/v1/config/industry
// Returns the industry profile for the current tenant
func (h *IndustryHandlers) GetIndustryProfile(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	// Try to get the industry_profile config for this tenant
	value, err := h.configService.GetConfig(r.Context(), tenantID, "industry_profile")
	if err != nil {
		h.logger.Error("get industry profile config error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	profileID := "event_tech" // default
	if value != nil {
		// Config value is stored as JSON, so it might be a quoted string
		var raw string
		if err := json.Unmarshal(value, &raw); err == nil && raw != "" {
			profileID = raw
		}
	}

	profile, ok := AllIndustryProfiles[profileID]
	if !ok {
		profile = AllIndustryProfiles["event_tech"]
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    profile,
		"message": "Industry profile retrieved successfully",
	})
}

// GetAllIndustryProfiles handles GET /api/v1/config/industries
// Returns all available industry profiles (for setup wizard / settings)
func (h *IndustryHandlers) GetAllIndustryProfiles(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantIDFromClaims(r.Context())
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated")
		return
	}

	// Build grouped list
	type GroupedProfiles struct {
		Group    string            `json:"group"`
		Profiles []IndustryProfile `json:"profiles"`
	}

	groupOrder := []string{"Technik", "Maschinen & Werkzeuge", "Ausstattung & Mobiliar", "Fahrzeuge & Outdoor", "Sonstiges"}
	groupMap := make(map[string][]IndustryProfile)

	for _, p := range industryProfileList {
		groupMap[p.Group] = append(groupMap[p.Group], p)
	}

	var grouped []GroupedProfiles
	for _, g := range groupOrder {
		if profiles, ok := groupMap[g]; ok {
			grouped = append(grouped, GroupedProfiles{
				Group:    g,
				Profiles: profiles,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":    grouped,
		"message": "Industry profiles retrieved successfully",
	})
}

// Default labels used when a profile doesn't override a key
var defaultLabels = map[string]string{
	"project":  "Projekt",
	"checkout": "Ausgabe",
	"checkin":  "Rueckgabe",
	"location": "Lager",
	"crew":     "Crew",
	"worker":   "Mitarbeiter",
}

// Default scanner home cards
var defaultScannerCards = []string{"quick_scan", "checkout", "checkin", "inventory"}

// industryProfileList maintains ordered list of all profiles
var industryProfileList = []IndustryProfile{
	{
		ID:    "event_tech",
		Label: "Veranstaltungstechnik",
		Group: "Technik",
		Labels: map[string]string{
			"project":        "Veranstaltung",
			"project_plural": "Veranstaltungen",
			"checkout":       "Check-Out",
			"checkin":        "Check-In",
			"location":       "Veranstaltungsort",
			"crew":           "Crew",
			"worker":         "Techniker",
		},
		Features: map[string]bool{
			"packing_lists":   true,
			"rfid_assign":     true,
			"operating_hours": false,
			"fuel_tracking":   false,
			"cleaning_status": false,
			"crew_planning":   true,
			"flight_cases":    true,
			"cable_tracking":  true,
			"power_distro":    true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory", "rfid_assign", "new_project"},
		Categories:       []string{"Licht", "Ton", "Video", "Rigging", "Strom", "Buehne", "Kabel", "Zubehoer"},
	},
	{
		ID:    "film_tech",
		Label: "Film & TV",
		Group: "Technik",
		Labels: map[string]string{
			"project":  "Produktion",
			"checkout": "Ausgabe",
			"checkin":  "Rueckgabe",
			"location": "Fundus",
			"crew":     "Team",
			"worker":   "Techniker",
		},
		Features: map[string]bool{
			"packing_lists":   true,
			"operating_hours": false,
			"fuel_tracking":   false,
			"cleaning_status": false,
			"crew_planning":   true,
			"flight_cases":    true,
			"cable_tracking":  true,
			"serial_tracking": true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "packing_list", "inventory"},
		Categories:       []string{"Kamera", "Objektive", "Licht", "Ton", "Grip", "Strom", "Kabel", "Zubehoer"},
	},
	{
		ID:    "it_rental",
		Label: "IT & Buerotechnik",
		Group: "Technik",
		Labels: map[string]string{
			"project":  "Auftrag",
			"checkout": "Ausgabe",
			"checkin":  "Rueckgabe",
			"location": "Lager",
			"crew":     "Team",
			"worker":   "Techniker",
		},
		Features: map[string]bool{
			"packing_lists":   true,
			"operating_hours": false,
			"fuel_tracking":   false,
			"cleaning_status": false,
			"serial_tracking": true,
			"asset_tags":      true,
			"software_config": true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory"},
		Categories:       []string{"Laptops", "Monitore", "Drucker", "Netzwerk", "Server", "Zubehoer", "Kabel", "Moebel"},
	},
	{
		ID:    "construction",
		Label: "Baumaschinen",
		Group: "Maschinen & Werkzeuge",
		Labels: map[string]string{
			"project":  "Baustelle",
			"checkout": "Lieferung",
			"checkin":  "Abholung",
			"location": "Depot",
			"crew":     "Kolonne",
			"worker":   "Maschinist",
		},
		Features: map[string]bool{
			"packing_lists":   false,
			"operating_hours": true,
			"fuel_tracking":   true,
			"cleaning_status": true,
			"damage_report":   true,
			"maintenance":     true,
			"weight_tracking": true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "damage_report", "operating_hours"},
		Categories:       []string{"Bagger", "Radlader", "Krane", "Verdichter", "Generatoren", "Anhaenger", "Kleingeraete", "Zubehoer"},
	},
	{
		ID:    "tool_rental",
		Label: "Werkzeugverleih",
		Group: "Maschinen & Werkzeuge",
		Labels: map[string]string{
			"project":  "Auftrag",
			"checkout": "Ausgabe",
			"checkin":  "Rueckgabe",
			"location": "Lager",
			"crew":     "Team",
			"worker":   "Mitarbeiter",
		},
		Features: map[string]bool{
			"packing_lists":   false,
			"operating_hours": true,
			"fuel_tracking":   false,
			"cleaning_status": true,
			"maintenance":     true,
			"calibration":     true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory"},
		Categories:       []string{"Bohrmaschinen", "Saegen", "Schleifer", "Messgeraete", "Elektrowerkzeug", "Handwerkzeug", "Zubehoer"},
	},
	{
		ID:    "scaffolding",
		Label: "Geruestbau",
		Group: "Maschinen & Werkzeuge",
		Labels: map[string]string{
			"project":  "Baustelle",
			"checkout": "Lieferung",
			"checkin":  "Abholung",
			"location": "Lager",
			"crew":     "Kolonne",
			"worker":   "Geruestbauer",
		},
		Features: map[string]bool{
			"packing_lists":    false,
			"operating_hours":  false,
			"fuel_tracking":    false,
			"cleaning_status":  false,
			"piece_counting":   true,
			"weight_tracking":  true,
			"safety_inspection": true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory", "piece_count"},
		Categories:       []string{"Rahmen", "Riegel", "Diagonalen", "Boeden", "Kupplungen", "Konsolen", "Treppen", "Zubehoer"},
	},
	{
		ID:    "landscaping",
		Label: "Landschaftsbau",
		Group: "Maschinen & Werkzeuge",
		Labels: map[string]string{
			"project":  "Baustelle",
			"checkout": "Ausgabe",
			"checkin":  "Rueckgabe",
			"location": "Lager",
			"crew":     "Kolonne",
			"worker":   "Gaertner",
		},
		Features: map[string]bool{
			"packing_lists":   false,
			"operating_hours": true,
			"fuel_tracking":   true,
			"cleaning_status": true,
			"maintenance":     true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "operating_hours"},
		Categories:       []string{"Maeher", "Heckenscheren", "Motorsaegen", "Haecksler", "Pumpen", "Anhaenger", "Handgeraete", "Zubehoer"},
	},
	{
		ID:    "electrical",
		Label: "Elektrotechnik",
		Group: "Maschinen & Werkzeuge",
		Labels: map[string]string{
			"project":  "Baustelle",
			"checkout": "Ausgabe",
			"checkin":  "Rueckgabe",
			"location": "Lager",
			"crew":     "Team",
			"worker":   "Elektriker",
		},
		Features: map[string]bool{
			"packing_lists":   true,
			"operating_hours": false,
			"fuel_tracking":   false,
			"cleaning_status": false,
			"calibration":     true,
			"safety_inspection": true,
			"cable_tracking":  true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory"},
		Categories:       []string{"Messgeraete", "Werkzeuge", "Kabel", "Verteiler", "Steckverbinder", "Sicherheitstechnik", "Zubehoer"},
	},
	{
		ID:    "party_rental",
		Label: "Party & Eventausstattung",
		Group: "Ausstattung & Mobiliar",
		Labels: map[string]string{
			"project":  "Event",
			"checkout": "Lieferung",
			"checkin":  "Abholung",
			"location": "Lager",
			"crew":     "Team",
			"worker":   "Mitarbeiter",
		},
		Features: map[string]bool{
			"packing_lists":   true,
			"operating_hours": false,
			"fuel_tracking":   false,
			"cleaning_status": true,
			"piece_counting":  true,
			"damage_report":   true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "packing_list", "piece_count"},
		Categories:       []string{"Tische", "Stuehle", "Zelte", "Geschirr", "Besteck", "Dekoration", "Textilien", "Zubehoer"},
	},
	{
		ID:    "exhibition",
		Label: "Messebau",
		Group: "Ausstattung & Mobiliar",
		Labels: map[string]string{
			"project":  "Messe",
			"checkout": "Lieferung",
			"checkin":  "Abholung",
			"location": "Lager",
			"crew":     "Team",
			"worker":   "Monteur",
		},
		Features: map[string]bool{
			"packing_lists":   true,
			"operating_hours": false,
			"fuel_tracking":   false,
			"cleaning_status": true,
			"modular_systems": true,
			"floor_plans":     true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "packing_list", "inventory"},
		Categories:       []string{"Waende", "Boeden", "Tresen", "Vitrinen", "Beleuchtung", "Grafik", "Elektrik", "Zubehoer"},
	},
	{
		ID:    "gastro",
		Label: "Gastro-Equipment",
		Group: "Ausstattung & Mobiliar",
		Labels: map[string]string{
			"project":  "Event",
			"checkout": "Lieferung",
			"checkin":  "Abholung",
			"location": "Lager",
			"crew":     "Team",
			"worker":   "Mitarbeiter",
		},
		Features: map[string]bool{
			"packing_lists":    false,
			"operating_hours":  false,
			"fuel_tracking":    false,
			"cleaning_status":  true,
			"piece_counting":   true,
			"hygiene_tracking": true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "piece_count"},
		Categories:       []string{"Kuehlgeraete", "Kochgeraete", "Spuelmaschinen", "Warmhaltung", "Zapfanlagen", "Geschirr", "Moebel", "Zubehoer"},
	},
	{
		ID:    "textile",
		Label: "Textil & Kostuem",
		Group: "Ausstattung & Mobiliar",
		Labels: map[string]string{
			"project":  "Produktion",
			"checkout": "Ausgabe",
			"checkin":  "Rueckgabe",
			"location": "Fundus",
			"crew":     "Team",
			"worker":   "Mitarbeiter",
		},
		Features: map[string]bool{
			"packing_lists":   true,
			"operating_hours": false,
			"fuel_tracking":   false,
			"cleaning_status": true,
			"size_tracking":   true,
			"condition_notes": true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory"},
		Categories:       []string{"Kostueme", "Anzuege", "Kleider", "Schuhe", "Accessoires", "Stoffe", "Requisiten", "Zubehoer"},
	},
	{
		ID:    "staging",
		Label: "Buehnen & Tribuenen",
		Group: "Ausstattung & Mobiliar",
		Labels: map[string]string{
			"project":  "Veranstaltung",
			"checkout": "Lieferung",
			"checkin":  "Abholung",
			"location": "Lager",
			"crew":     "Kolonne",
			"worker":   "Monteur",
		},
		Features: map[string]bool{
			"packing_lists":    false,
			"operating_hours":  false,
			"fuel_tracking":    false,
			"cleaning_status":  false,
			"piece_counting":   true,
			"weight_tracking":  true,
			"safety_inspection": true,
			"static_calc":      true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory", "piece_count"},
		Categories:       []string{"Buehnenelemente", "Tribuenen", "Podeste", "Treppen", "Gelaender", "Verkleidungen", "Daecher", "Zubehoer"},
	},
	{
		ID:    "vehicle_rental",
		Label: "Fahrzeugvermietung",
		Group: "Fahrzeuge & Outdoor",
		Labels: map[string]string{
			"project":  "Buchung",
			"checkout": "Uebergabe",
			"checkin":  "Ruecknahme",
			"location": "Standort",
			"crew":     "Team",
			"worker":   "Mitarbeiter",
		},
		Features: map[string]bool{
			"packing_lists":    false,
			"operating_hours":  true,
			"fuel_tracking":    true,
			"cleaning_status":  true,
			"damage_report":    true,
			"mileage_tracking": true,
			"insurance":        true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "damage_report", "operating_hours"},
		Categories:       []string{"PKW", "Transporter", "LKW", "Anhaenger", "Spezialfahrzeuge", "E-Fahrzeuge", "Zubehoer"},
	},
	{
		ID:    "camping",
		Label: "Camping & Wohnmobile",
		Group: "Fahrzeuge & Outdoor",
		Labels: map[string]string{
			"project":  "Buchung",
			"checkout": "Uebergabe",
			"checkin":  "Ruecknahme",
			"location": "Standort",
			"crew":     "Team",
			"worker":   "Mitarbeiter",
		},
		Features: map[string]bool{
			"packing_lists":    false,
			"operating_hours":  true,
			"fuel_tracking":    true,
			"cleaning_status":  true,
			"damage_report":    true,
			"mileage_tracking": true,
			"inventory_check":  true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "damage_report", "inventory"},
		Categories:       []string{"Wohnmobile", "Wohnwagen", "Dachzelte", "Campingmoebel", "Kochausruestung", "Outdoor", "Zubehoer"},
	},
	{
		ID:    "sports",
		Label: "Sport & Freizeit",
		Group: "Fahrzeuge & Outdoor",
		Labels: map[string]string{
			"project":  "Buchung",
			"checkout": "Ausgabe",
			"checkin":  "Rueckgabe",
			"location": "Lager",
			"crew":     "Team",
			"worker":   "Mitarbeiter",
		},
		Features: map[string]bool{
			"packing_lists":   false,
			"operating_hours": false,
			"fuel_tracking":   false,
			"cleaning_status": true,
			"size_tracking":   true,
			"damage_report":   true,
			"maintenance":     true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory"},
		Categories:       []string{"Ski", "Snowboard", "Fahrraeder", "Boote", "Surfbretter", "Klettern", "Fitness", "Zubehoer"},
	},
	{
		ID:    "container",
		Label: "Container & Raumsysteme",
		Group: "Sonstiges",
		Labels: map[string]string{
			"project":  "Baustelle",
			"checkout": "Lieferung",
			"checkin":  "Abholung",
			"location": "Depot",
			"crew":     "Team",
			"worker":   "Fahrer",
		},
		Features: map[string]bool{
			"packing_lists":    false,
			"operating_hours":  false,
			"fuel_tracking":    false,
			"cleaning_status":  true,
			"damage_report":    true,
			"location_tracking": true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "damage_report"},
		Categories:       []string{"Buerocontainer", "Lagercontainer", "Sanitaercontainer", "Wohncontainer", "Spezialcontainer", "Zubehoer"},
	},
	{
		ID:    "medical",
		Label: "Medizintechnik",
		Group: "Sonstiges",
		Labels: map[string]string{
			"project":  "Einsatz",
			"checkout": "Ausgabe",
			"checkin":  "Rueckgabe",
			"location": "Lager",
			"crew":     "Team",
			"worker":   "Techniker",
		},
		Features: map[string]bool{
			"packing_lists":    true,
			"operating_hours":  false,
			"fuel_tracking":    false,
			"cleaning_status":  true,
			"serial_tracking":  true,
			"calibration":      true,
			"hygiene_tracking": true,
			"safety_inspection": true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory"},
		Categories:       []string{"Diagnostik", "Therapie", "Monitoring", "Beatmung", "Chirurgie", "Labor", "Mobilitaet", "Zubehoer"},
	},
	{
		ID:    "general",
		Label: "Allgemeiner Verleih",
		Group: "Sonstiges",
		Labels: map[string]string{
			"project":  "Auftrag",
			"checkout": "Ausgabe",
			"checkin":  "Rueckgabe",
			"location": "Lager",
			"crew":     "Team",
			"worker":   "Mitarbeiter",
		},
		Features: map[string]bool{
			"packing_lists":   true,
			"operating_hours": false,
			"fuel_tracking":   false,
			"cleaning_status": true,
		},
		ScannerHomeCards: []string{"scan_info", "checkout", "checkin", "inventory"},
		Categories:       []string{"Kategorie 1", "Kategorie 2", "Kategorie 3", "Zubehoer"},
	},
}

// AllIndustryProfiles is a map for quick lookup by ID
var AllIndustryProfiles = func() map[string]IndustryProfile {
	m := make(map[string]IndustryProfile, len(industryProfileList))
	for _, p := range industryProfileList {
		m[p.ID] = p
	}
	return m
}()
