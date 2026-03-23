export type CrewRole = 'technician' | 'rigger' | 'driver' | 'supervisor' | 'assistant' | 'stagehand' | 'light_tech' | 'sound_tech'
export type AvailabilityStatus = 'available' | 'busy' | 'on_leave' | 'sick'
export type QualificationType = 'IPAF' | 'electrical_cert' | 'first_aid' | 'forklift' | 'rope_access'

export interface Qualification {
  id: string
  type: QualificationType
  name: string
  expiry_date: string
  is_expired: boolean
}

export interface CrewMember {
  id: string
  name: string
  email: string
  phone: string
  role: CrewRole
  availability: AvailabilityStatus
  qualifications: Qualification[]
  hours_this_week: number
  current_assignment?: string
  joined_date: string
}

export interface TimeRecord {
  id: string
  crew_member_id: string
  start_time: string
  end_time?: string
  project_id: string
  project_name: string
  duration_hours: number
  status: 'active' | 'completed'
}

export interface Assignment {
  id: string
  crew_member_id: string
  project_id: string
  project_name: string
  start_date: string
  end_date: string
  status: 'scheduled' | 'in_progress' | 'completed'
}
