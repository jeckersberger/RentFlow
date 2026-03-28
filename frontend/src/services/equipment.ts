import api from './api';
import type { Equipment } from '../types/equipment';
import type { PaginatedResponse } from '../types/common';

export async function list(
  params?: Record<string, unknown>
): Promise<PaginatedResponse<Equipment>> {
  return api.get('/api/v1/equipment', { params }) as unknown as PaginatedResponse<Equipment>;
}

export async function get(id: string): Promise<Equipment> {
  return api.get(`/api/v1/equipment/${id}`) as unknown as Equipment;
}

export async function create(data: Partial<Equipment>): Promise<Equipment> {
  return api.post('/api/v1/equipment', data) as unknown as Equipment;
}

export async function update(id: string, data: Partial<Equipment>): Promise<Equipment> {
  return api.put(`/api/v1/equipment/${id}`, data) as unknown as Equipment;
}

export async function remove(id: string): Promise<void> {
  await api.delete(`/api/v1/equipment/${id}`);
}

export async function search(
  query: string,
  params?: Record<string, unknown>
): Promise<PaginatedResponse<Equipment>> {
  return api.get('/api/v1/equipment/search', {
    params: { q: query, ...params },
  }) as unknown as PaginatedResponse<Equipment>;
}
