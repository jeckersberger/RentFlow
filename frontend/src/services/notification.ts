import api from './api';

export interface Notification {
  id: string;
  type: string;
  title: string;
  message: string;
  read: boolean;
  entity_type?: string;
  entity_id?: string;
  created_at: string;
}

export const notificationService = {
  list: () => api.get('/api/v1/notifications'),
  getUnreadCount: () => api.get('/api/v1/notifications/unread-count'),
  markRead: (id: string) => api.put(`/api/v1/notifications/${id}/read`, {}),
};
