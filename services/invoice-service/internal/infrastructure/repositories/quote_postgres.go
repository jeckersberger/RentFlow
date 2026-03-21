package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/invoice-service/internal/ports"
)

type QuotePostgres struct {
	db *database.PostgresPool
}

func NewQuotePostgres(db *database.PostgresPool) *QuotePostgres {
	return &QuotePostgres{db: db}
}

func (r *QuotePostgres) Create(ctx context.Context, quote *domain.Quote) error {
	query := `
		INSERT INTO invoice.quotes (
			id, tenant_id, quote_number, project_id, client_name, client_address_street,
			client_address_city, client_address_postcode, client_address_country,
			client_email, sub_total, tax_rate, tax_amount, total, currency,
			status, valid_until, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20
		)
	`

	_, err := r.db.Exec(ctx, query,
		quote.ID, quote.TenantID, quote.QuoteNumber, quote.ProjectID, quote.ClientName,
		quote.ClientAddress.Street, quote.ClientAddress.City, quote.ClientAddress.PostCode,
		quote.ClientAddress.Country, quote.ClientEmail, quote.SubTotal, quote.TaxRate,
		quote.TaxAmount, quote.Total, quote.Currency, string(quote.Status),
		quote.ValidUntil, quote.Notes, quote.CreatedAt, quote.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create quote: %w", err)
	}

	// Insert items
	for _, item := range quote.Items {
		itemQuery := `
			INSERT INTO invoice.quote_items (
				id, quote_id, description, quantity, unit, unit_price, total_price, tax_rate, equipment_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err := r.db.Exec(ctx, itemQuery,
			item.ID, quote.ID, item.Description, item.Quantity, item.Unit,
			item.UnitPrice, item.TotalPrice, item.TaxRate, item.EquipmentID,
		)
		if err != nil {
			return fmt.Errorf("failed to create quote item: %w", err)
		}
	}

	return nil
}

func (r *QuotePostgres) GetByID(ctx context.Context, tenantID, quoteID string) (*domain.Quote, error) {
	query := `
		SELECT id, tenant_id, quote_number, project_id, client_name, client_address_street,
		       client_address_city, client_address_postcode, client_address_country,
		       client_email, sub_total, tax_rate, tax_amount, total, currency,
		       status, valid_until, notes, created_at, updated_at
		FROM invoice.quotes
		WHERE id = $1 AND tenant_id = $2
	`

	row := r.db.QueryRow(ctx, query, quoteID, tenantID)

	quote := &domain.Quote{}
	err := row.Scan(
		&quote.ID, &quote.TenantID, &quote.QuoteNumber, &quote.ProjectID, &quote.ClientName,
		&quote.ClientAddress.Street, &quote.ClientAddress.City, &quote.ClientAddress.PostCode,
		&quote.ClientAddress.Country, &quote.ClientEmail, &quote.SubTotal, &quote.TaxRate,
		&quote.TaxAmount, &quote.Total, &quote.Currency, &quote.Status,
		&quote.ValidUntil, &quote.Notes, &quote.CreatedAt, &quote.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrQuoteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	items, err := r.getQuoteItems(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to load quote items: %w", err)
	}
	quote.Items = items

	return quote, nil
}

func (r *QuotePostgres) GetByNumber(ctx context.Context, tenantID, quoteNumber string) (*domain.Quote, error) {
	query := `
		SELECT id, tenant_id, quote_number, project_id, client_name, client_address_street,
		       client_address_city, client_address_postcode, client_address_country,
		       client_email, sub_total, tax_rate, tax_amount, total, currency,
		       status, valid_until, notes, created_at, updated_at
		FROM invoice.quotes
		WHERE quote_number = $1 AND tenant_id = $2
	`

	row := r.db.QueryRow(ctx, query, quoteNumber, tenantID)

	quote := &domain.Quote{}
	err := row.Scan(
		&quote.ID, &quote.TenantID, &quote.QuoteNumber, &quote.ProjectID, &quote.ClientName,
		&quote.ClientAddress.Street, &quote.ClientAddress.City, &quote.ClientAddress.PostCode,
		&quote.ClientAddress.Country, &quote.ClientEmail, &quote.SubTotal, &quote.TaxRate,
		&quote.TaxAmount, &quote.Total, &quote.Currency, &quote.Status,
		&quote.ValidUntil, &quote.Notes, &quote.CreatedAt, &quote.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrQuoteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	items, err := r.getQuoteItems(ctx, quote.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load quote items: %w", err)
	}
	quote.Items = items

	return quote, nil
}

func (r *QuotePostgres) List(ctx context.Context, query *ports.QuoteListQuery) (*ports.QuoteListResult, error) {
	var whereConditions []string
	var args []interface{}
	argCount := 1

	whereConditions = append(whereConditions, fmt.Sprintf("tenant_id = $%d", argCount))
	args = append(args, query.TenantID)
	argCount++

	if query.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, string(*query.Status))
		argCount++
	}

	if query.ClientName != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("client_name ILIKE $%d", argCount))
		args = append(args, "%"+*query.ClientName+"%")
		argCount++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM invoice.quotes WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count quotes: %w", err)
	}

	listQuery := fmt.Sprintf(`
		SELECT id, tenant_id, quote_number, project_id, client_name, client_address_street,
		       client_address_city, client_address_postcode, client_address_country,
		       client_email, sub_total, tax_rate, tax_amount, total, currency,
		       status, valid_until, notes, created_at, updated_at
		FROM invoice.quotes
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, query.Limit, query.Offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list quotes: %w", err)
	}
	defer rows.Close()

	var quotes []*domain.Quote
	for rows.Next() {
		quote := &domain.Quote{}
		err := rows.Scan(
			&quote.ID, &quote.TenantID, &quote.QuoteNumber, &quote.ProjectID, &quote.ClientName,
			&quote.ClientAddress.Street, &quote.ClientAddress.City, &quote.ClientAddress.PostCode,
			&quote.ClientAddress.Country, &quote.ClientEmail, &quote.SubTotal, &quote.TaxRate,
			&quote.TaxAmount, &quote.Total, &quote.Currency, &quote.Status,
			&quote.ValidUntil, &quote.Notes, &quote.CreatedAt, &quote.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quote: %w", err)
		}

		items, err := r.getQuoteItems(ctx, quote.ID)
		if err == nil {
			quote.Items = items
		}

		quotes = append(quotes, quote)
	}

	return &ports.QuoteListResult{
		Items:  quotes,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func (r *QuotePostgres) Update(ctx context.Context, quote *domain.Quote) error {
	query := `
		UPDATE invoice.quotes SET
			client_name = $3, client_address_street = $4, client_address_city = $5,
			client_address_postcode = $6, client_address_country = $7, client_email = $8,
			sub_total = $9, tax_rate = $10, tax_amount = $11, total = $12,
			currency = $13, status = $14, valid_until = $15, notes = $16, updated_at = $17
		WHERE id = $1 AND tenant_id = $2
	`

	_, err := r.db.Exec(ctx, query,
		quote.ID, quote.TenantID, quote.ClientName,
		quote.ClientAddress.Street, quote.ClientAddress.City, quote.ClientAddress.PostCode,
		quote.ClientAddress.Country, quote.ClientEmail, quote.SubTotal, quote.TaxRate,
		quote.TaxAmount, quote.Total, quote.Currency, string(quote.Status),
		quote.ValidUntil, quote.Notes, quote.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update quote: %w", err)
	}

	// Clear and re-insert items
	_, err = r.db.Exec(ctx, "DELETE FROM invoice.quote_items WHERE quote_id = $1", quote.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old items: %w", err)
	}

	for _, item := range quote.Items {
		itemQuery := `
			INSERT INTO invoice.quote_items (
				id, quote_id, description, quantity, unit, unit_price, total_price, tax_rate, equipment_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err := r.db.Exec(ctx, itemQuery,
			item.ID, quote.ID, item.Description, item.Quantity, item.Unit,
			item.UnitPrice, item.TotalPrice, item.TaxRate, item.EquipmentID,
		)
		if err != nil {
			return fmt.Errorf("failed to update quote item: %w", err)
		}
	}

	return nil
}

func (r *QuotePostgres) Delete(ctx context.Context, tenantID, quoteID string) error {
	query := `
		UPDATE invoice.quotes
		SET status = 'rejected', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`

	_, err := r.db.Exec(ctx, query, quoteID, tenantID)
	return err
}

func (r *QuotePostgres) getQuoteItems(ctx context.Context, quoteID string) ([]domain.InvoiceItem, error) {
	query := `
		SELECT id, description, quantity, unit, unit_price, total_price, tax_rate, equipment_id
		FROM invoice.quote_items
		WHERE quote_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []domain.InvoiceItem
	for rows.Next() {
		item := domain.InvoiceItem{}
		err := rows.Scan(
			&item.ID, &item.Description, &item.Quantity, &item.Unit,
			&item.UnitPrice, &item.TotalPrice, &item.TaxRate, &item.EquipmentID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}
