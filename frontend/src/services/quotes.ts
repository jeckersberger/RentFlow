import api from './api';
import type { Quote, QuoteItem } from '../types/quote';
import type { PaginatedResponse } from '../types/common';

export async function list(
  params?: Record<string, unknown>
): Promise<PaginatedResponse<Quote>> {
  return api.get('/api/v1/quotes', { params }) as unknown as PaginatedResponse<Quote>;
}

export async function get(id: string): Promise<Quote> {
  return api.get(`/api/v1/quotes/${id}`) as unknown as Quote;
}

export async function create(data: Partial<Quote>): Promise<Quote> {
  return api.post('/api/v1/quotes', data) as unknown as Quote;
}

export async function update(id: string, data: Partial<Quote>): Promise<Quote> {
  return api.put(`/api/v1/quotes/${id}`, data) as unknown as Quote;
}

export async function addItem(
  quoteId: string,
  data: Partial<QuoteItem>
): Promise<QuoteItem> {
  return api.post(`/api/v1/quotes/${quoteId}/items`, data) as unknown as QuoteItem;
}

export async function listItems(quoteId: string): Promise<QuoteItem[]> {
  return api.get(`/api/v1/quotes/${quoteId}/items`) as unknown as QuoteItem[];
}

export async function removeItem(quoteId: string, itemId: string): Promise<void> {
  await api.delete(`/api/v1/quotes/${quoteId}/items/${itemId}`);
}

export async function convertToInvoice(quoteId: string): Promise<{ invoice_id: string }> {
  return api.post(`/api/v1/quotes/${quoteId}/convert-to-invoice`) as unknown as { invoice_id: string };
}
