import api from './api';

export interface InsurancePolicy {
  id: string;
  name: string;
  provider: string;
  policy_number: string;
  premium: number;
  coverage_amount: number;
  start_date: string;
  end_date: string;
  status: string;
  created_at: string;
}

export interface InsuranceClaim {
  id: string;
  policy_id: string;
  equipment_id: string;
  description: string;
  amount: number;
  status: string;
  filed_at: string;
  resolved_at?: string;
  created_at: string;
}

export const insuranceService = {
  listPolicies: () => api.get('/api/v1/insurance-policies'),
  getPolicy: (id: string) => api.get(`/api/v1/insurance-policies/${id}`),
  createPolicy: (data: Partial<InsurancePolicy>) => api.post('/api/v1/insurance-policies', data),
  listClaims: () => api.get('/api/v1/insurance-claims'),
};
