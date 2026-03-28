package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

type ExpenseCategory struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type Expense struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	ProjectID     uuid.UUID  `json:"project_id,omitempty"`
	CategoryID    uuid.UUID  `json:"category_id,omitempty"`
	Description   string     `json:"description"`
	Amount        int64      `json:"amount"`
	Currency      string     `json:"currency"`
	ExpenseDate   string     `json:"expense_date"`
	ReceiptNumber string     `json:"receipt_number"`
	Vendor        string     `json:"vendor"`
	Status        string     `json:"status"`
	ApprovedBy    uuid.UUID  `json:"approved_by,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	Notes         string     `json:"notes"`
	CreatedBy     uuid.UUID  `json:"created_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ExpenseReceipt struct {
	ID         uuid.UUID `json:"id"`
	ExpenseID  uuid.UUID `json:"expense_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int       `json:"file_size"`
	MimeType   string    `json:"mime_type"`
	UploadedBy uuid.UUID `json:"uploaded_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type ExpenseFilter struct {
	Page      int    `json:"page"`
	PerPage   int    `json:"per_page"`
	ProjectID string `json:"project_id,omitempty"`
	Status    string `json:"status,omitempty"`
}
