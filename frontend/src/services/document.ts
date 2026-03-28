import api from './api';
import type { DocumentTemplate, Document, Attachment } from '@/types/document';

export async function listTemplates(): Promise<DocumentTemplate[]> {
  return api.get('/api/v1/document-templates') as unknown as DocumentTemplate[];
}

export async function getTemplate(id: string): Promise<DocumentTemplate> {
  return api.get(`/api/v1/document-templates/${id}`) as unknown as DocumentTemplate;
}

export async function createTemplate(body: Partial<DocumentTemplate>): Promise<DocumentTemplate> {
  return api.post('/api/v1/document-templates', body) as unknown as DocumentTemplate;
}

export async function updateTemplate(id: string, body: Partial<DocumentTemplate>): Promise<DocumentTemplate> {
  return api.put(`/api/v1/document-templates/${id}`, body) as unknown as DocumentTemplate;
}

export async function listDocuments(): Promise<Document[]> {
  return api.get('/api/v1/documents') as unknown as Document[];
}

export async function getDocument(id: string): Promise<Document> {
  return api.get(`/api/v1/documents/${id}`) as unknown as Document;
}

export async function createDocument(body: Partial<Document>): Promise<Document> {
  return api.post('/api/v1/documents', body) as unknown as Document;
}

export async function listAttachments(params?: { reference_id?: string; reference_type?: string }): Promise<Attachment[]> {
  return api.get('/api/v1/attachments', { params }) as unknown as Attachment[];
}

export async function getAttachment(id: string): Promise<Attachment> {
  return api.get(`/api/v1/attachments/${id}`) as unknown as Attachment;
}

export async function deleteAttachment(id: string): Promise<void> {
  await api.delete(`/api/v1/attachments/${id}`);
}
