export type ScanType = 'check_in' | 'check_out' | 'inventory' | 'pack_verify'

export interface ScanEvent {
  id: string
  barcode: string
  equipment_id?: string
  equipment_name?: string
  scan_type: ScanType
  project_id?: string
  user_id: string
  timestamp: string
  location?: string
  notes?: string
}

export interface ScanResult {
  success: boolean
  message: string
  equipment?: {
    id: string
    name: string
    sku: string
    status: string
  }
  event?: ScanEvent
}
