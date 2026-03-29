export interface Quote {
  id: string;
  tenant_id: string;
  quote_number: string;
  status: string;
  customer_name: string;
  customer_email: string;
  customer_address: string;
  project_id: string;
  subject: string;
  intro_text: string;
  outro_text: string;
  quote_date: string;
  valid_until: string;
  vat_rate: number;
  kleinunternehmer: boolean;
  total_net: number;
  total_vat: number;
  total_gross: number;
  payment_terms_days: number;
  discount_pct: number;
  notes: string;
  converted_invoice_id: string;
  created_at: string;
  updated_at: string;
}

export interface QuoteItem {
  id: string;
  tenant_id: string;
  quote_id: string;
  description: string;
  quantity: number;
  unit: string;
  unit_price: number;
  discount_pct: number;
  position: number;
  created_at: string;
  updated_at: string;
}
