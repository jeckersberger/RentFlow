export type InvoiceStatus = 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled' | 'partial'

export interface InvoiceLineItem {
  id: string
  invoice_id: string
  name?: string
  description: string
  quantity: number
  unit_price: number
  total: number
  tax_rate?: number
  tax_amount?: number
}

export interface InvoicePayment {
  id: string
  invoice_id: string
  amount: number
  date: string
  method?: string
  notes?: string
}

export interface DunningEntry {
  level: number
  date: string
  sent: boolean
}

export interface Invoice {
  id: string
  number: string
  client_id: string
  client_name?: string
  client_email?: string
  client_address?: string
  project_id?: string
  project_name?: string
  status: InvoiceStatus
  issue_date: string
  due_date: string
  subtotal: number
  tax_total?: number
  total: number
  notes?: string
  paid_date?: string
  payment_terms?: string
  is_kleinunternehmer?: boolean
  kleinunternehmer_text?: string
  bank_name?: string
  bank_iban?: string
  bank_bic?: string
  bank_account_holder?: string
  payments?: InvoicePayment[]
  dunning_entries?: DunningEntry[]
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
    name?: string
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
