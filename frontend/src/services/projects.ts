import api from './api';
import type { Project } from '../types/project';
import type { PaginatedResponse } from '../types/common';

export async function list(
  params?: Record<string, unknown>
): Promise<PaginatedResponse<Project>> {
  return api.get('/api/v1/projects', { params }) as unknown as PaginatedResponse<Project>;
}

export async function get(id: string): Promise<Project> {
  return api.get(`/api/v1/projects/${id}`) as unknown as Project;
}

export async function create(data: Partial<Project>): Promise<Project> {
  return api.post('/api/v1/projects', data) as unknown as Project;
}

export async function update(id: string, data: Partial<Project>): Promise<Project> {
  return api.put(`/api/v1/projects/${id}`, data) as unknown as Project;
}

export async function remove(id: string): Promise<void> {
  await api.delete(`/api/v1/projects/${id}`);
}

export interface CalendarEvent {
  id: string;
  title: string;
  start_date: string;
  end_date: string;
  status: string;
  color: string;
  customer: string;
  venue_name: string;
}

export async function getCalendar(from: string, to: string): Promise<CalendarEvent[]> {
  const data = await api.get('/api/v1/projects/calendar', {
    params: { from, to },
  });
  return (data as unknown as CalendarEvent[]) ?? [];
}
