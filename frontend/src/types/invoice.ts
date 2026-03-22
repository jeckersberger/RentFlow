export type InvoiceStatus = 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled'

export interface InvoiceLineItem {
  id: string
  invoice_id: string
  description: string
  quantity: number
  unit_price: number
  total: number
  tax_rate?: number
  tax_amount?: number
}

export interface Invoice {
  id: string
  number: string
  client_id: string
  client_name?: string
  project_id?: string
  status: InvoiceStatus
  issue_date: string
  due_date: string
  subtotal: number
  tax_total?: number
  total: number
  notes?: string
  paid_date?: string
  created_at: string
  updated_at: string
  line_items?: InvoiceLineItem[]
}

export interface CreateInvoiceDTO {
  number: string
  client_id: string
  project_id?: string
  issue_date: string
  due_date: string
  notes?: string
  line_items: Array<{
    description: string
    quantity: number
    unit_price: number
    tax_rate?: number
  }>
}

export interface UpdateInvoiceDTO extends Partial<CreateInvoiceDTO> {
  status?: InvoiceStatus
}

export interface InvoiceListResponse {
  data: Invoice[]
  total: number
  page: number
  limit: number
}
