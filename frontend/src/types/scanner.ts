export interface ScanEvent {
  id: string;
  tenant_id: string;
  user_id: string;
  device_id: string;
  barcode: string;
  rfid_tag: string;
  equipment_id: string;
  action: string;
  project_id: string;
  location_id: string;
  condition_rating: number;
  condition_notes: string;
  gps_lat: number;
  gps_lng: number;
  timestamp: string;
  synced_at: string;
  created_at: string;
}

export interface ScannerDevice {
  id: string;
  tenant_id: string;
  device_id: string;
  device_name: string;
  device_type: string;
  fcm_token: string;
  ring_requested: boolean;
  last_seen: string;
  created_at: string;
}

export interface ScanRequest {
  barcode?: string;
  rfid_tag?: string;
}

export interface CheckoutRequest {
  equipment_ids: string[];
  project_id: string;
  device_id?: string;
}

export interface CheckinRequest {
  equipment_ids: string[];
  condition_rating?: number;
  condition_notes?: string;
  device_id?: string;
}

export interface BulkSyncResponse {
  inserted: number;
  skipped: number;
  total: number;
}
