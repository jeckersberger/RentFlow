export interface CrewMember {
  id: string;
  tenant_id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  role: string;
  hourly_rate: number;
  is_active: boolean;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface CrewAssignment {
  id: string;
  crew_member_id: string;
  project_id: string;
  tenant_id: string;
  role: string;
  start_date: string;
  end_date: string;
  hours_planned: number;
  hours_actual: number;
  status: string;
  notes: string;
  created_at: string;
}

export interface CrewQualification {
  id: string;
  crew_member_id: string;
  tenant_id: string;
  name: string;
  issued_at: string;
  expires_at: string;
  certificate_number: string;
  notes: string;
  created_at: string;
}
