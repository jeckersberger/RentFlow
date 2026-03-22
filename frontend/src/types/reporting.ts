export type ReportPeriod = 'week' | 'month' | 'quarter' | 'year'
export type TrendDirection = 'up' | 'down' | 'stable'

export interface KPIMetric {
  id: string
  label: string
  value: number | string
  unit?: string
  trend: TrendDirection
  trend_percentage: number
  target?: number
}

export interface ReportDefinition {
  id: string
  name: string
  description: string
  type: string
  created_date: string
  last_run?: string
  run_count: number
}

export interface ReportRun {
  id: string
  report_id: string
  report_name: string
  period: ReportPeriod
  generated_date: string
  generated_by: string
  file_url?: string
}

export interface ReportingStats {
  total_revenue: number
  equipment_utilization: number
  active_projects: number
  overdue_invoices: number
  avg_rental_days: number
  maintenance_compliance: number
}
