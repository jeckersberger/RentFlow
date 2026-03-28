import api from './api';

export interface AuditLog {
  id: string;
  entity_type: string;
  entity_id: string;
  action: string;
  user_id: string;
  description: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export const auditService = {
  list: (page = 1, perPage = 20) => api.get(`/api/v1/audit-logs?page=${page}&per_page=${perPage}`),
  create: (data: Partial<AuditLog>) => api.post('/api/v1/audit-logs', data),
};
