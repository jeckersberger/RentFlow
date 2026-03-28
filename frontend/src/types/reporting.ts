export interface ReportDefinition {
  id: string;
  tenant_id: string;
  name: string;
  type: string;
  query_config: Record<string, unknown>;
  schedule: string;
  format: string;
  is_active: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface ReportSnapshot {
  id: string;
  definition_id: string;
  tenant_id: string;
  title: string;
  data: Record<string, unknown>;
  file_path: string;
  file_size: number;
  format: string;
  period_start: string;
  period_end: string;
  generated_by: string;
  generated_at: string;
}

export interface DashboardWidget {
  id: string;
  tenant_id: string;
  name: string;
  type: string;
  config: Record<string, unknown>;
  position: number;
  is_active: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}
