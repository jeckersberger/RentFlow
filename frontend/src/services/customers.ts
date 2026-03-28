import api from './api';
import type { Customer } from '../types/customer';
import type { PaginatedResponse } from '../types/common';

export async function list(
  params?: Record<string, unknown>
): Promise<PaginatedResponse<Customer>> {
  return api.get('/api/v1/customers', { params }) as unknown as PaginatedResponse<Customer>;
}

export async function get(id: string): Promise<Customer> {
  return api.get(`/api/v1/customers/${id}`) as unknown as Customer;
}

export async function create(data: Partial<Customer>): Promise<Customer> {
  return api.post('/api/v1/customers', data) as unknown as Customer;
}

export async function update(id: string, data: Partial<Customer>): Promise<Customer> {
  return api.put(`/api/v1/customers/${id}`, data) as unknown as Customer;
}

export async function remove(id: string): Promise<void> {
  await api.delete(`/api/v1/customers/${id}`);
}

export async function search(
  query: string,
  params?: Record<string, unknown>
): Promise<PaginatedResponse<Customer>> {
  return api.get('/api/v1/customers/search', {
    params: { q: query, ...params },
  }) as unknown as PaginatedResponse<Customer>;
}
