export interface Invoice {
  id: string;
  tenant_id: string;
  project_id: string;
  customer_id: string;
  invoice_number: string;
  status: string;
  issue_date: string;
  due_date: string;
  subtotal: number;
  tax_rate: number;
  tax_amount: number;
  total_amount: number;
  paid_amount: number;
  currency: string;
  notes: string;
  payment_terms: string;
  finalized_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface InvoiceItem {
  id: string;
  invoice_id: string;
  equipment_id: string | null;
  description: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  rental_days: number;
  created_at: string;
  updated_at: string;
}

export interface Payment {
  id: string;
  invoice_id: string;
  amount: number;
  payment_method: string;
  payment_date: string;
  reference: string;
  notes: string;
  created_at: string;
  updated_at: string;
}
