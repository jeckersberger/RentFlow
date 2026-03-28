import api from './api';
import type { Invoice, InvoiceItem, Payment } from '../types/invoice';
import type { PaginatedResponse } from '../types/common';

export async function list(
  params?: Record<string, unknown>
): Promise<PaginatedResponse<Invoice>> {
  return api.get('/api/v1/invoices', { params }) as unknown as PaginatedResponse<Invoice>;
}

export async function get(id: string): Promise<Invoice> {
  return api.get(`/api/v1/invoices/${id}`) as unknown as Invoice;
}

export async function create(data: Partial<Invoice>): Promise<Invoice> {
  return api.post('/api/v1/invoices', data) as unknown as Invoice;
}

export async function update(id: string, data: Partial<Invoice>): Promise<Invoice> {
  return api.put(`/api/v1/invoices/${id}`, data) as unknown as Invoice;
}

export async function finalize(id: string): Promise<Invoice> {
  return api.post(`/api/v1/invoices/${id}/finalize`) as unknown as Invoice;
}

export async function search(
  query: string,
  params?: Record<string, unknown>
): Promise<PaginatedResponse<Invoice>> {
  return api.get('/api/v1/invoices/search', {
    params: { q: query, ...params },
  }) as unknown as PaginatedResponse<Invoice>;
}

// Invoice Items
export async function addItem(
  invoiceId: string,
  data: Partial<InvoiceItem>
): Promise<InvoiceItem> {
  return api.post(`/api/v1/invoices/${invoiceId}/items`, data) as unknown as InvoiceItem;
}

export async function listItems(invoiceId: string): Promise<InvoiceItem[]> {
  return api.get(`/api/v1/invoices/${invoiceId}/items`) as unknown as InvoiceItem[];
}

export async function removeItem(
  invoiceId: string,
  itemId: string
): Promise<void> {
  await api.delete(`/api/v1/invoices/${invoiceId}/items/${itemId}`);
}

// Payments
export async function addPayment(
  invoiceId: string,
  data: Partial<Payment>
): Promise<Payment> {
  return api.post(`/api/v1/invoices/${invoiceId}/payments`, data) as unknown as Payment;
}

export async function listPayments(invoiceId: string): Promise<Payment[]> {
  return api.get(`/api/v1/invoices/${invoiceId}/payments`) as unknown as Payment[];
}
