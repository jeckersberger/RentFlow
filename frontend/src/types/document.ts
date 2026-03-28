export interface DocumentTemplate {
  id: string;
  tenant_id: string;
  name: string;
  type: string;
  content: string;
  variables: Record<string, unknown>;
  is_default: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Document {
  id: string;
  tenant_id: string;
  template_id: string;
  type: string;
  title: string;
  reference_id: string;
  reference_type: string;
  content: string;
  file_path: string;
  file_size: number;
  mime_type: string;
  status: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface Attachment {
  id: string;
  tenant_id: string;
  reference_id: string;
  reference_type: string;
  file_name: string;
  file_path: string;
  file_size: number;
  mime_type: string;
  uploaded_by: string;
  created_at: string;
}
