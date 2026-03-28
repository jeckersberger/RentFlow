import api from './api';

export interface FederationPartner {
  id: string;
  name: string;
  api_url: string;
  status: string;
  created_at: string;
}

export interface SharedListing {
  id: string;
  equipment_id: string;
  partner_id?: string;
  status: string;
  daily_rate: number;
  available_from: string;
  available_until: string;
  created_at: string;
}

export const federationService = {
  listPartners: () => api.get('/api/v1/federation-partners'),
  getPartner: (id: string) => api.get(`/api/v1/federation-partners/${id}`),
  createPartner: (data: Partial<FederationPartner>) => api.post('/api/v1/federation-partners', data),
  listListings: () => api.get('/api/v1/shared-listings'),
};
