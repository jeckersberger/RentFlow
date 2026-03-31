export type VehicleStatus = 'available' | 'in_use' | 'maintenance'
export type TourStatus = 'planned' | 'loading' | 'in_transit' | 'delivered' | 'completed'

export interface Vehicle {
  id: string
  tenant_id: string
  license_plate: string
  name: string
  capacity_kg: number
  capacity_m3: number
  vehicle_type: string
  status: VehicleStatus
  dguv_last_check: string
  dguv_next_check: string
  created_at: string
  updated_at: string
}

export interface Tour {
  id: string
  tenant_id: string
  vehicle_id: string
  vehicle?: Vehicle
  project_id: string
  driver_id: string
  status: TourStatus
  departure_at: string
  arrival_at: string
  km_start: number
  km_end: number
  total_cost: number
  notes: string
  equipment_items?: TourEquipment[]
  created_at: string
  updated_at: string
}

export interface TourEquipment {
  id: string
  tour_id: string
  equipment_id: string
  equipment_name?: string
  weight_kg: number
  volume_m3: number
  loaded_at: string | null
  unloaded_at: string | null
}

export interface DriverLog {
  id: string
  tour_id: string
  driver_id: string
  start_time: string
  end_time: string
  break_minutes: number
  km_driven: number
  notes: string
}

export interface TourCapacity {
  vehicle_capacity_kg: number
  vehicle_capacity_m3: number
  used_kg: number
  used_m3: number
  remaining_kg: number
  remaining_m3: number
}

export interface CreateVehicleDTO {
  license_plate: string
  name: string
  capacity_kg: number
  capacity_m3: number
  vehicle_type: string
  dguv_last_check?: string
  dguv_next_check?: string
}

export interface CreateTourDTO {
  vehicle_id: string
  project_id: string
  driver_id?: string
  departure_at: string
  notes?: string
}
