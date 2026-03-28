import api from './api';

export interface MaintenanceSchedule {
  id: string;
  equipment_id: string;
  name: string;
  interval_days: number;
  last_performed: string;
  next_due: string;
  created_at: string;
}

export interface MaintenanceTask {
  id: string;
  equipment_id: string;
  schedule_id?: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  assigned_to?: string;
  due_date?: string;
  completed_at?: string;
  cost: number;
  notes: string;
  created_at: string;
}

export const maintenanceService = {
  listSchedules: () => api.get('/api/v1/maintenance-schedules'),
  listTasks: () => api.get('/api/v1/maintenance-tasks'),
  getTask: (id: string) => api.get(`/api/v1/maintenance-tasks/${id}`),
  createTask: (data: Partial<MaintenanceTask>) => api.post('/api/v1/maintenance-tasks', data),
};
