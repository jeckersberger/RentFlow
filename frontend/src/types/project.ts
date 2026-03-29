export interface Project {
  id: string;
  tenant_id: string;
  name: string;
  project_number: string;
  description: string;
  status: string;
  customer_id: string;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  venue_name: string;
  venue_address: string;
  start_date: string;
  end_date: string;
  setup_date: string;
  teardown_date: string;
  color: string;
  budget: number;
  currency: string;
  manager_id: string;
  notes: string;
  created_at: string;
  updated_at: string;
}
