export interface Customer {
  id: string;
  tenant_id: string;
  company_name: string;
  customer_number: string;
  email: string;
  phone: string;
  mobile: string;
  website: string;
  billing_address_street: string;
  billing_address_city: string;
  billing_address_zip: string;
  billing_address_country: string;
  shipping_address_street: string;
  shipping_address_city: string;
  shipping_address_zip: string;
  shipping_address_country: string;
  tax_id: string;
  notes: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Contact {
  id: string;
  tenant_id: string;
  customer_id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  mobile: string;
  position: string;
  is_primary: boolean;
  notes: string;
  created_at: string;
  updated_at: string;
}
