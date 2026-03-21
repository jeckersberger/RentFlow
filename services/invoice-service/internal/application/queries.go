package application

// Query objects for retrieving data

type GetInvoiceQuery struct {
	TenantID  string
	InvoiceID string
}

type ListInvoicesQuery struct {
	TenantID   string
	Status     *string
	ClientName *string
	FromDate   *string
	ToDate     *string
	Limit      int
	Offset     int
}

type GetInvoiceByNumberQuery struct {
	TenantID      string
	InvoiceNumber string
}

type GetQuoteQuery struct {
	TenantID string
	QuoteID  string
}

type ListQuotesQuery struct {
	TenantID   string
	Status     *string
	ClientName *string
	FromDate   *string
	ToDate     *string
	Limit      int
	Offset     int
}

type GetQuoteByNumberQuery struct {
	TenantID    string
	QuoteNumber string
}

type GetDunningQuery struct {
	TenantID  string
	DunningID string
}

type ListDunningQuery struct {
	TenantID  string
	InvoiceID string
}

type GetOverdueInvoicesQuery struct {
	TenantID string
}

type ExportDATEVQuery struct {
	TenantID string
	FromDate string
	ToDate   string
}

type ExportCSVQuery struct {
	TenantID string
	FromDate string
	ToDate   string
	Type     string // "invoices", "quotes", "dunning"
}
