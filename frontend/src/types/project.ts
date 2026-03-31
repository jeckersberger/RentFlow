export type ProjectStatus = 'draft' | 'quoted' | 'confirmed' | 'in_progress' | 'completed' | 'cancelled' | 'invoiced'

export interface AddressDTO {
  street: string
  city: string
  state: string
  postal_code: string
  country: string
  coordinates?: string
}

export interface Project {
  id: string
  name: string
  description?: string
  client_name?: string
  client_email?: string
  client_phone?: string
  client_address?: AddressDTO
  venue_address?: AddressDTO
  status: ProjectStatus
  start_date: string
  end_date: string
  setup_date?: string
  teardown_date?: string
  project_manager?: string
  budget?: number
  currency?: string
  notes?: string
  tags?: string[]
  created_at: string
  updated_at: string
  created_by_user_id?: string
  // Legacy fields kept for backward compat
  client_id?: string
  location?: string
}

export interface ProjectPackItem {
  id: string
  project_id: string
  equipment_id: string
  equipment_name?: string
  quantity: number
  packed: boolean
  checked_out: boolean
  returned: boolean
}

export interface ProjectCrew {
  id: string
  project_id: string
  crew_member_id: string
  crew_member_name?: string
  role: string
  hours?: number
}

export interface CreateProjectDTO {
  name: string
  description?: string
  client_name: string
  client_email?: string
  client_phone?: string
  venue_address?: AddressDTO
  status?: ProjectStatus
  start_date: string
  end_date: string
  budget?: number
  currency?: string
  notes?: string
}

export interface UpdateProjectDTO extends Partial<CreateProjectDTO> {}

export interface ProjectListResponse {
  data: Project[]
  total: number
  page: number
  limit: number
}
