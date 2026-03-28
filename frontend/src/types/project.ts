export interface Project {
  id: string;
  tenant_id: string;
  customer_id: string;
  name: string;
  description: string;
  status: string;
  start_date: string;
  end_date: string;
  venue: string;
  venue_address: string;
  contact_person: string;
  contact_phone: string;
  contact_email: string;
  notes: string;
  total_amount: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}
