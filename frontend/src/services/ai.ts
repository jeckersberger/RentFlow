import api from './api';

export interface AIPrediction {
  id: string;
  type: string;
  input_data: Record<string, unknown>;
  prediction: Record<string, unknown>;
  confidence: number;
  created_at: string;
}

export interface AISuggestion {
  id: string;
  type: string;
  context: Record<string, unknown>;
  suggestion: string;
  accepted: boolean;
  created_at: string;
}

export const aiService = {
  listPredictions: () => api.get('/api/v1/ai-predictions'),
  createPrediction: (data: Partial<AIPrediction>) => api.post('/api/v1/ai-predictions', data),
  listSuggestions: () => api.get('/api/v1/ai-suggestions'),
};
