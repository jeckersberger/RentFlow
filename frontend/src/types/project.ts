export type ProjectStatus = 'draft' | 'confirmed' | 'in_progress' | 'completed' | 'cancelled'

export interface Project {
  id: string
  name: string
  description?: string
  client_id: string
  client_name?: string
  status: ProjectStatus
  start_date: string
  end_date: string
  location?: string
  budget?: number
  notes?: string
  created_at: string
  updated_at: string
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
  client_id: string
  status?: ProjectStatus
  start_date: string
  end_date: string
  location?: string
  budget?: number
  notes?: string
}

export interface UpdateProjectDTO extends Partial<CreateProjectDTO> {}

export interface ProjectListResponse {
  data: Project[]
  total: number
  page: number
  limit: number
}
