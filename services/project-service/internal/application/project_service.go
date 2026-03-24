package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
)

type ProjectService struct {
	repo                   ports.ProjectRepository
	packlistRepo           ports.PacklistRepository
	logger                 logger.Logger
	notificationServiceURL string
}

func NewProjectService(repo ports.ProjectRepository, logger logger.Logger) *ProjectService {
	return &ProjectService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ProjectService) SetPacklistRepository(repo ports.PacklistRepository) {
	s.packlistRepo = repo
}

func (s *ProjectService) SetNotificationServiceURL(url string) {
	s.notificationServiceURL = url
}

func (s *ProjectService) CreateProject(ctx context.Context, cmd CreateProjectCommand) (*ProjectDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "project name is required", nil)
	}
	projectID := fmt.Sprintf("proj_%d", hashString(cmd.TenantID+cmd.Name))

	project := domain.NewProject(projectID, cmd.TenantID, cmd.Name, cmd.ClientName, cmd.CreatedByUserID)
	project.Description = cmd.Description
	project.ClientEmail = cmd.ClientEmail
	project.ClientPhone = cmd.ClientPhone
	project.ClientAddress = domain.Address{
		Street:      cmd.ClientAddress.Street,
		City:        cmd.ClientAddress.City,
		State:       cmd.ClientAddress.State,
		PostalCode:  cmd.ClientAddress.PostalCode,
		Country:     cmd.ClientAddress.Country,
		Coordinates: cmd.ClientAddress.Coordinates,
	}
	project.VenueAddress = domain.Address{
		Street:      cmd.VenueAddress.Street,
		City:        cmd.VenueAddress.City,
		State:       cmd.VenueAddress.State,
		PostalCode:  cmd.VenueAddress.PostalCode,
		Country:     cmd.VenueAddress.Country,
		Coordinates: cmd.VenueAddress.Coordinates,
	}
	project.StartDate = cmd.StartDate
	project.EndDate = cmd.EndDate
	project.SetupDate = cmd.SetupDate
	project.TeardownDate = cmd.TeardownDate
	project.Budget = cmd.Budget
	project.Currency = cmd.Currency
	project.Notes = cmd.Notes
	project.Tags = cmd.Tags

	if err := project.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create project", err)
	}

	s.logger.Info("Project created", "id", project.ID, "tenant_id", cmd.TenantID)
	return ProjectToDTO(project), nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, cmd UpdateProjectCommand) (*ProjectDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	project, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	project.Name = cmd.Name
	project.Description = cmd.Description
	project.ClientName = cmd.ClientName
	project.ClientEmail = cmd.ClientEmail
	project.ClientPhone = cmd.ClientPhone
	project.ClientAddress = domain.Address{
		Street:      cmd.ClientAddress.Street,
		City:        cmd.ClientAddress.City,
		State:       cmd.ClientAddress.State,
		PostalCode:  cmd.ClientAddress.PostalCode,
		Country:     cmd.ClientAddress.Country,
		Coordinates: cmd.ClientAddress.Coordinates,
	}
	project.VenueAddress = domain.Address{
		Street:      cmd.VenueAddress.Street,
		City:        cmd.VenueAddress.City,
		State:       cmd.VenueAddress.State,
		PostalCode:  cmd.VenueAddress.PostalCode,
		Country:     cmd.VenueAddress.Country,
		Coordinates: cmd.VenueAddress.Coordinates,
	}
	project.StartDate = cmd.StartDate
	project.EndDate = cmd.EndDate
	project.SetupDate = cmd.SetupDate
	project.TeardownDate = cmd.TeardownDate
	project.Budget = cmd.Budget
	project.Currency = cmd.Currency
	project.Notes = cmd.Notes
	project.Tags = cmd.Tags

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, domain.NewDomainError("UPDATE_ERROR", "failed to update project", err)
	}

	s.logger.Info("Project updated", "id", cmd.ID, "tenant_id", cmd.TenantID)
	return ProjectToDTO(project), nil
}

func (s *ProjectService) GetProject(ctx context.Context, tenantID, projectID string) (*ProjectDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	project, err := s.repo.GetByID(ctx, tenantID, projectID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	return ProjectToDTO(project), nil
}

func (s *ProjectService) ListProjects(ctx context.Context, query ListProjectsQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	var result *ports.ProjectListResult
	var err error

	if query.Filter != "" {
		result, err = s.repo.ListWithFilter(ctx, query.TenantID, query.Filter, query.Limit, query.Offset)
	} else {
		result, err = s.repo.List(ctx, query.TenantID, query.Limit, query.Offset)
	}
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list projects", err)
	}

	dtos := make([]*ProjectDTO, len(result.Items))
	for i, project := range result.Items {
		dtos[i] = ProjectToDTO(project)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *ProjectService) SearchProjects(ctx context.Context, query SearchProjectsQuery) (*PaginatedResult, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	result, err := s.repo.Search(ctx, query.TenantID, query.SearchTerm, query.Limit, query.Offset)
	if err != nil {
		return nil, domain.NewDomainError("SEARCH_ERROR", "failed to search projects", err)
	}

	dtos := make([]*ProjectDTO, len(result.Items))
	for i, project := range result.Items {
		dtos[i] = ProjectToDTO(project)
	}

	return &PaginatedResult{
		Data:   dtos,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}, nil
}

func (s *ProjectService) ChangeStatus(ctx context.Context, cmd ChangeProjectStatusCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	project, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	oldStatus := string(project.Status)
	if err := project.ChangeStatus(cmd.Status); err != nil {
		return domain.NewDomainError("INVALID_STATUS", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, project); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to update status", err)
	}

	s.logger.Info("Project status changed", "id", cmd.ID, "status", cmd.Status)

	// Send notification in background (non-blocking)
	if s.notificationServiceURL != "" {
		go s.sendStatusChangeNotification(project, oldStatus)
	}

	return nil
}

func (s *ProjectService) SetProjectManager(ctx context.Context, cmd SetProjectManagerCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	project, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	if err := project.SetProjectManager(cmd.ManagerID); err != nil {
		return domain.NewDomainError("INVALID_INPUT", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, project); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to update project manager", err)
	}

	s.logger.Info("Project manager set", "id", cmd.ID, "manager_id", cmd.ManagerID)
	return nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, tenantID, projectID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	_, err := s.repo.GetByID(ctx, tenantID, projectID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	if err := s.repo.Delete(ctx, tenantID, projectID); err != nil {
		return domain.NewDomainError("DELETE_ERROR", "failed to delete project", err)
	}

	s.logger.Info("Project deleted", "id", projectID, "tenant_id", tenantID)
	return nil
}

func (s *ProjectService) GetCalendar(ctx context.Context, query GetCalendarQuery) ([]CalendarEventDTO, error) {
	projects, err := s.repo.ListByDateRange(ctx, query.TenantID, query.StartDate, query.EndDate)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to fetch calendar", err)
	}

	events := make([]CalendarEventDTO, len(projects))
	for i, p := range projects {
		color := getColorByStatus(p.Status)
		events[i] = CalendarEventDTO{
			ID:         p.ID,
			Name:       p.Name,
			ClientName: p.ClientName,
			Status:     string(p.Status),
			StartDate:  p.StartDate,
			EndDate:    p.EndDate,
			Color:      color,
		}
	}

	return events, nil
}

func (s *ProjectService) CopyProject(ctx context.Context, projectID, tenantID, newStartDateStr, newEndDateStr string) (*ProjectDTO, error) {
	project, err := s.repo.GetByID(ctx, tenantID, projectID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	newProject := domain.NewProject(
		fmt.Sprintf("proj_%d", hashString(tenantID+project.Name+newStartDateStr)),
		tenantID,
		project.Name+" (Copy)",
		project.ClientName,
		project.CreatedByUserID,
	)

	newProject.Description = project.Description
	newProject.ClientEmail = project.ClientEmail
	newProject.ClientPhone = project.ClientPhone
	newProject.ClientAddress = project.ClientAddress
	newProject.VenueAddress = project.VenueAddress
	newProject.Status = domain.ProjectDraft
	newProject.Budget = project.Budget
	newProject.Currency = project.Currency
	newProject.Notes = project.Notes
	newProject.Tags = project.Tags

	// Parse new dates
	startTime, err := parseDate(newStartDateStr)
	if err != nil {
		return nil, domain.NewDomainError("INVALID_DATE", "invalid start date", err)
	}
	endTime, err := parseDate(newEndDateStr)
	if err != nil {
		return nil, domain.NewDomainError("INVALID_DATE", "invalid end date", err)
	}

	newProject.StartDate = startTime
	newProject.EndDate = endTime

	if err := newProject.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.repo.Create(ctx, newProject); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to copy project", err)
	}

	s.logger.Info("Project copied", "original_id", projectID, "new_id", newProject.ID, "tenant_id", tenantID)
	return ProjectToDTO(newProject), nil
}

func (s *ProjectService) GeneratePackingListJSON(ctx context.Context, tenantID, projectID string) (*PackingListDTO, error) {
	project, err := s.repo.GetByID(ctx, tenantID, projectID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	var packlists []*domain.Packlist
	if s.packlistRepo != nil {
		result, err := s.packlistRepo.ListByProjectID(ctx, tenantID, projectID, 1000, 0)
		if err != nil {
			s.logger.Warn("Failed to fetch packlists for JSON generation", "error", err)
		} else {
			packlists = result.Items
		}
	}

	// Group items by storage location
	locationMap := make(map[string][]PackingListItemDTO)
	totalItems := 0

	for _, packlist := range packlists {
		for _, item := range packlist.Items {
			loc := item.StorageLocation
			if loc == "" {
				loc = "Nicht zugeordnet"
			}
			locationMap[loc] = append(locationMap[loc], PackingListItemDTO{
				Name:           item.EquipmentName,
				Quantity:       item.Quantity,
				QuantityPacked: item.QuantityPacked,
				Status:         string(item.Status),
				PacklistName:   packlist.Name,
			})
			totalItems++
		}
	}

	// Build sorted locations slice
	var locations []PackingListLocationDTO
	for loc, items := range locationMap {
		locations = append(locations, PackingListLocationDTO{
			Location: loc,
			Items:    items,
		})
	}

	// Sort locations alphabetically
	for i := 0; i < len(locations); i++ {
		for j := i + 1; j < len(locations); j++ {
			if locations[i].Location > locations[j].Location {
				locations[i], locations[j] = locations[j], locations[i]
			}
		}
	}

	projectDates := fmt.Sprintf("%s - %s",
		project.StartDate.Format("02.01.2006"),
		project.EndDate.Format("02.01.2006"))

	return &PackingListDTO{
		ProjectName:  project.Name,
		ProjectDates: projectDates,
		Locations:    locations,
		TotalItems:   totalItems,
	}, nil
}

func (s *ProjectService) GeneratePackingListHTML(ctx context.Context, tenantID, projectID string) (string, error) {
	project, err := s.repo.GetByID(ctx, tenantID, projectID)
	if err != nil {
		return "", domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	var packlists []*domain.Packlist
	if s.packlistRepo != nil {
		result, err := s.packlistRepo.ListByProjectID(ctx, tenantID, projectID, 1000, 0)
		if err != nil {
			s.logger.Warn("Failed to fetch packlists for HTML generation", "error", err)
		} else {
			packlists = result.Items
		}
	}

	html := buildPackingListHTML(project, packlists)
	return html, nil
}

// GetProjectEquipment returns the equipment list (Soll-Liste) for a project based on packlist items.
func (s *ProjectService) GetProjectEquipment(ctx context.Context, tenantID, projectID string) (*ProjectEquipmentDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	_, err := s.repo.GetByID(ctx, tenantID, projectID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "project not found", err)
	}

	var items []ProjectEquipmentItemDTO

	if s.packlistRepo != nil {
		result, err := s.packlistRepo.ListByProjectID(ctx, tenantID, projectID, 1000, 0)
		if err != nil {
			s.logger.Warn("Failed to fetch packlists for equipment list", "error", err)
		} else {
			// Aggregate items across all packlists by equipment_id
			equipmentMap := make(map[string]*ProjectEquipmentItemDTO)
			for _, packlist := range result.Items {
				for _, item := range packlist.Items {
					if existing, ok := equipmentMap[item.EquipmentID]; ok {
						existing.Quantity += item.Quantity
						existing.CheckedOut += item.QuantityPacked
					} else {
						equipmentMap[item.EquipmentID] = &ProjectEquipmentItemDTO{
							ID:         item.EquipmentID,
							Name:       item.EquipmentName,
							SKU:        "", // SKU not stored in packlist items
							Quantity:   item.Quantity,
							CheckedOut: item.QuantityPacked,
						}
					}
				}
			}
			for _, v := range equipmentMap {
				items = append(items, *v)
			}
		}
	}

	if items == nil {
		items = []ProjectEquipmentItemDTO{}
	}

	return &ProjectEquipmentDTO{Items: items}, nil
}

func (s *ProjectService) sendStatusChangeNotification(project *domain.Project, oldStatus string) {
	notificationPayload := map[string]interface{}{
		"tenant_id": project.TenantID,
		"type":      "project_status_changed",
		"subject":   fmt.Sprintf("Projekt-Status geändert: %s", project.Name),
		"body": fmt.Sprintf("Das Projekt '%s' hat den Status von '%s' auf '%s' geändert.",
			project.Name, oldStatus, string(project.Status)),
		"recipients": []string{},
	}

	payload, err := json.Marshal(notificationPayload)
	if err != nil {
		s.logger.Error("Failed to marshal notification payload", err)
		return
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/api/v1/notifications", s.notificationServiceURL),
		bytes.NewBuffer(payload))
	if err != nil {
		s.logger.Error("Failed to create notification request", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send status change notification", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		s.logger.Warn("Notification service returned error", "status", resp.StatusCode)
	}
}

func getStatusColor(status string) string {
	switch status {
	case "pending":
		return "#ffc107"
	case "packed":
		return "#28a745"
	case "loaded":
		return "#17a2b8"
	case "returned":
		return "#6c757d"
	case "missing":
		return "#dc3545"
	case "damaged":
		return "#e83e8c"
	default:
		return "#cccccc"
	}
}

func getColorByStatus(status domain.ProjectStatus) string {
	switch status {
	case domain.ProjectDraft:
		return "#cccccc"
	case domain.ProjectQuoted:
		return "#ffc107"
	case domain.ProjectConfirmed:
		return "#17a2b8"
	case domain.ProjectInProgress:
		return "#28a745"
	case domain.ProjectCompleted:
		return "#6c757d"
	case domain.ProjectInvoiced:
		return "#007bff"
	case domain.ProjectCancelled:
		return "#dc3545"
	default:
		return "#cccccc"
	}
}

func parseDate(dateStr string) (time.Time, error) {
	// Try RFC3339 format first
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func buildPackingListHTML(project *domain.Project, packlists []*domain.Packlist) string {
	// Build table rows for all packlist items
	tableRows := ""
	if len(packlists) == 0 {
		tableRows = `            <tr>
                <td colspan="6" style="text-align: center; background: #f0f0f0;">No equipment data available</td>
            </tr>`
	} else {
		for _, packlist := range packlists {
			for _, item := range packlist.Items {
				statusColor := getStatusColor(string(item.Status))
				tableRows += fmt.Sprintf(`            <tr>
                <td>%s</td>
                <td>%s</td>
                <td>%d</td>
                <td>%d / %d</td>
                <td>%s</td>
                <td><span style="background-color: %s; padding: 4px 8px; border-radius: 3px; color: white; font-size: 0.85em;">%s</span></td>
            </tr>
`, packlist.Name, item.EquipmentName, item.Quantity, item.QuantityPacked, item.Quantity, item.StorageLocation, statusColor, item.Status)
			}
		}
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Packing List - ` + project.Name + `</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { margin-bottom: 30px; }
        .project-info { background: #f5f5f5; padding: 15px; margin-bottom: 20px; border-radius: 5px; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .status-badge { padding: 4px 8px; border-radius: 3px; color: white; font-size: 0.85em; }
        @media print { body { margin: 0; } }
    </style>
</head>
<body>
    <div class="header">
        <h1>Packing List</h1>
        <h2>` + project.Name + `</h2>
    </div>
    <div class="project-info">
        <p><strong>Client:</strong> ` + project.ClientName + `</p>
        <p><strong>Dates:</strong> ` + project.StartDate.Format("2006-01-02") + ` to ` + project.EndDate.Format("2006-01-02") + `</p>
        <p><strong>Location:</strong> ` + project.VenueAddress.City + `, ` + project.VenueAddress.Country + `</p>
    </div>
    <table>
        <thead>
            <tr>
                <th>Packlist</th>
                <th>Equipment</th>
                <th>Quantity</th>
                <th>Packed</th>
                <th>Storage Location</th>
                <th>Status</th>
            </tr>
        </thead>
        <tbody>
` + tableRows + `
        </tbody>
    </table>
</body>
</html>`
	return html
}

func hashString(s string) int64 {
	h := int64(5381)
	for _, c := range s {
		h = ((h << 5) + h) + int64(c)
	}
	return h & 0x7FFFFFFFFFFFFFFF
}
