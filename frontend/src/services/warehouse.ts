import api from './api';
import type { Warehouse, Zone, Rack, StockLocation, Movement, InventoryCheck } from '../types/warehouse';
import type { PaginatedResponse } from '../types/common';

// Warehouses
export async function listWarehouses(): Promise<Warehouse[]> {
  return api.get('/api/v1/warehouses') as unknown as Warehouse[];
}

export async function getWarehouse(id: string): Promise<Warehouse> {
  return api.get(`/api/v1/warehouses/${id}`) as unknown as Warehouse;
}

export async function createWarehouse(data: Partial<Warehouse>): Promise<Warehouse> {
  return api.post('/api/v1/warehouses', data) as unknown as Warehouse;
}

// Zones
export async function listZones(params?: Record<string, unknown>): Promise<Zone[]> {
  return api.get('/api/v1/zones', { params }) as unknown as Zone[];
}

export async function createZone(data: Partial<Zone>): Promise<Zone> {
  return api.post('/api/v1/zones', data) as unknown as Zone;
}

export async function updateZone(id: string, data: Partial<Zone>): Promise<Zone> {
  return api.put(`/api/v1/zones/${id}`, data) as unknown as Zone;
}

// Racks
export async function listRacks(params?: Record<string, unknown>): Promise<Rack[]> {
  return api.get('/api/v1/racks', { params }) as unknown as Rack[];
}

export async function createRack(data: Partial<Rack>): Promise<Rack> {
  return api.post('/api/v1/racks', data) as unknown as Rack;
}

// Locations
export async function listLocations(
  params?: Record<string, unknown>
): Promise<PaginatedResponse<StockLocation>> {
  return api.get('/api/v1/locations', { params }) as unknown as PaginatedResponse<StockLocation>;
}

export async function getLocation(id: string): Promise<StockLocation> {
  return api.get(`/api/v1/locations/${id}`) as unknown as StockLocation;
}

export async function createLocation(data: Partial<StockLocation>): Promise<StockLocation> {
  return api.post('/api/v1/locations', data) as unknown as StockLocation;
}

// Movements
export async function listMovements(
  params?: Record<string, unknown>
): Promise<PaginatedResponse<Movement>> {
  return api.get('/api/v1/movements', { params }) as unknown as PaginatedResponse<Movement>;
}

export async function createMovement(data: Partial<Movement>): Promise<Movement> {
  return api.post('/api/v1/movements', data) as unknown as Movement;
}

// Inventory Checks
export async function listInventoryChecks(
  params?: Record<string, unknown>
): Promise<PaginatedResponse<InventoryCheck>> {
  return api.get('/api/v1/inventory-checks', { params }) as unknown as PaginatedResponse<InventoryCheck>;
}

export async function createInventoryCheck(data: {
  zone_id?: string;
  notes?: string;
}): Promise<InventoryCheck> {
  return api.post('/api/v1/inventory-checks', data) as unknown as InventoryCheck;
}

export async function completeInventoryCheck(id: string): Promise<InventoryCheck> {
  return api.patch(`/api/v1/inventory-checks/${id}/complete`) as unknown as InventoryCheck;
}

export async function getInventoryCheckResult(id: string): Promise<unknown> {
  return api.get(`/api/v1/inventory-checks/${id}/result`) as unknown;
}
