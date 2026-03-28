import api from './api';

export interface WorkflowDefinition {
  id: string;
  name: string;
  type: string;
  steps: Array<{ name: string; role: string }>;
  created_at: string;
}

export interface WorkflowInstance {
  id: string;
  definition_id: string;
  entity_type: string;
  entity_id: string;
  status: string;
  current_step: number;
  created_at: string;
}

export const workflowService = {
  listDefinitions: () => api.get('/api/v1/workflow-definitions'),
  createDefinition: (data: Partial<WorkflowDefinition>) => api.post('/api/v1/workflow-definitions', data),
  listInstances: () => api.get('/api/v1/workflow-instances'),
};
