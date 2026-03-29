export interface Invoice {
  id: string;
  tenant_id: string;
  invoice_number: string;
  invoice_type: string;
  status: string;
  customer_name: string;
  customer_email: string;
  customer_address: string;
  invoice_date: string;
  due_date: string;
  vat_rate: number;
  kleinunternehmer: boolean;
  total_net: number;
  total_vat: number;
  total_gross: number;
  amount_paid: number;
  notes: string;
  hash: string;
  previous_hash: string;
  finalized_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface InvoiceItem {
  id: string;
  tenant_id: string;
  invoice_id: string;
  description: string;
  quantity: number;
  unit: string;
  unit_price: number;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface Payment {
  id: string;
  tenant_id: string;
  invoice_id: string;
  amount: number;
  payment_date: string;
  payment_method: string;
  reference: string;
  created_at: string;
}
