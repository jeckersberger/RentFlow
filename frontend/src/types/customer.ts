export interface Customer {
  id: string;
  tenant_id: string;
  company_name: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  address_street: string;
  address_city: string;
  address_zip: string;
  address_country: string;
  tax_id: string;
  customer_number: string;
  notes: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Contact {
  id: string;
  customer_id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  role: string;
  is_primary: boolean;
  created_at: string;
  updated_at: string;
}
