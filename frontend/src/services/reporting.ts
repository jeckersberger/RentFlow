import api from './api';
import type { ReportDefinition, ReportSnapshot, DashboardWidget } from '@/types/reporting';

export async function listDefinitions(): Promise<ReportDefinition[]> {
  return api.get('/api/v1/report-definitions') as unknown as ReportDefinition[];
}

export async function getDefinition(id: string): Promise<ReportDefinition> {
  return api.get(`/api/v1/report-definitions/${id}`) as unknown as ReportDefinition;
}

export async function createDefinition(body: Partial<ReportDefinition>): Promise<ReportDefinition> {
  return api.post('/api/v1/report-definitions', body) as unknown as ReportDefinition;
}

export async function updateDefinition(id: string, body: Partial<ReportDefinition>): Promise<ReportDefinition> {
  return api.put(`/api/v1/report-definitions/${id}`, body) as unknown as ReportDefinition;
}

export async function generateSnapshot(definitionId: string): Promise<ReportSnapshot> {
  return api.post(`/api/v1/report-definitions/${definitionId}/generate`) as unknown as ReportSnapshot;
}

export async function listSnapshots(params?: { definition_id?: string }): Promise<ReportSnapshot[]> {
  return api.get('/api/v1/report-snapshots', { params }) as unknown as ReportSnapshot[];
}

export async function getSnapshot(id: string): Promise<ReportSnapshot> {
  return api.get(`/api/v1/report-snapshots/${id}`) as unknown as ReportSnapshot;
}

export async function listWidgets(): Promise<DashboardWidget[]> {
  return api.get('/api/v1/dashboard-widgets') as unknown as DashboardWidget[];
}

export async function createWidget(body: Partial<DashboardWidget>): Promise<DashboardWidget> {
  return api.post('/api/v1/dashboard-widgets', body) as unknown as DashboardWidget;
}

export async function updateWidget(id: string, body: Partial<DashboardWidget>): Promise<DashboardWidget> {
  return api.put(`/api/v1/dashboard-widgets/${id}`, body) as unknown as DashboardWidget;
}

export async function deleteWidget(id: string): Promise<void> {
  await api.delete(`/api/v1/dashboard-widgets/${id}`);
}
