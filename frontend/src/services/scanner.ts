import api from './api';
import type { ScanEvent, ScannerDevice, BulkSyncResponse } from '../types/scanner';
import type { PaginatedResponse } from '../types/common';

export async function scan(data: { barcode?: string; rfid_tag?: string }): Promise<unknown> {
  return api.post('/api/v1/scanner/scan', data) as unknown;
}

export async function checkout(data: {
  equipment_ids: string[];
  project_id: string;
  device_id?: string;
}): Promise<unknown> {
  return api.post('/api/v1/scanner/checkout', data) as unknown;
}

export async function checkin(data: {
  equipment_ids: string[];
  condition_rating?: number;
  condition_notes?: string;
  device_id?: string;
}): Promise<unknown> {
  return api.post('/api/v1/scanner/checkin', data) as unknown;
}

export async function bulkSync(data: { events: unknown[] }): Promise<BulkSyncResponse> {
  return api.post('/api/v1/scanner/bulk', data) as unknown as BulkSyncResponse;
}

export async function listEvents(
  params?: Record<string, unknown>
): Promise<PaginatedResponse<ScanEvent>> {
  return api.get('/api/v1/scanner/events', { params }) as unknown as PaginatedResponse<ScanEvent>;
}

export async function registerDevice(data: {
  device_id: string;
  device_name?: string;
  device_type?: string;
}): Promise<ScannerDevice> {
  return api.post('/api/v1/scanner/devices/register', data) as unknown as ScannerDevice;
}

export async function listDevices(): Promise<ScannerDevice[]> {
  return api.get('/api/v1/scanner/devices') as unknown as ScannerDevice[];
}

export async function triggerRing(deviceId: string): Promise<void> {
  await api.post(`/api/v1/scanner/devices/${deviceId}/ring`);
}
