package application

type CreateDocumentCommand struct {
	TenantID   string
	Name       string
	Type       string
	EntityType string
	EntityID   string
	FileRef    string
	MimeType   string
	Size       int64
	Checksum   string
	CreatedBy  string
}

type GenerateDocumentCommand struct {
	TenantID   string
	TemplateID string
	EntityType string
	EntityID   string
	Variables  map[string]string
	CreatedBy  string
}

type CreateTemplateCommand struct {
	TenantID  string
	Name      string
	Type      string
	Content   string
	Variables []string
	CreatedBy string
}

type UpdateTemplateCommand struct {
	ID        string
	TenantID  string
	Name      string
	Type      string
	Content   string
	Variables []string
}

type SetTemplateDefaultCommand struct {
	ID        string
	TenantID  string
	IsDefault bool
}
