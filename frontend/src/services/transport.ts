import api from './api';

export interface Vehicle {
  id: string;
  name: string;
  license_plate: string;
  type: string;
  status: string;
  created_at: string;
}

export interface TransportOrder {
  id: string;
  vehicle_id?: string;
  project_id?: string;
  type: string;
  status: string;
  pickup_address: string;
  delivery_address: string;
  scheduled_at: string;
  completed_at?: string;
  notes: string;
  created_at: string;
}

export const transportService = {
  listVehicles: () => api.get('/api/v1/vehicles'),
  createVehicle: (data: Partial<Vehicle>) => api.post('/api/v1/vehicles', data),
  listOrders: () => api.get('/api/v1/transport-orders'),
  createOrder: (data: Partial<TransportOrder>) => api.post('/api/v1/transport-orders', data),
};
