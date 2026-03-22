export type PlanType = 'interval' | 'after_use' | 'hours_based'
export type TaskStatus = 'planned' | 'in_progress' | 'completed' | 'overdue' | 'cancelled'
export type TaskPriority = 'low' | 'medium' | 'high' | 'critical'
export type TestResult = 'passed' | 'failed' | 'conditional'

export interface MaintenancePlan {
  id: string
  tenant_id: string
  equipment_id: string
  equipment_name?: string
  plan_type: PlanType
  interval_days: number | null
  interval_hours: number | null
  name: string
  description: string
  checklist_template_id: string | null
  last_executed_at: string | null
  next_due_at: string | null
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface MaintenanceTask {
  id: string
  tenant_id: string
  plan_id: string | null
  plan_name?: string
  equipment_id: string
  equipment_name?: string
  assigned_to: string | null
  status: TaskStatus
  priority: TaskPriority
  scheduled_at: string
  started_at: string | null
  completed_at: string | null
  notes: string
  checklist_data: ChecklistData | null
  created_at: string
  updated_at: string
}

export interface ChecklistData {
  items: ChecklistItemResult[]
}

export interface ChecklistItemResult {
  name: string
  check_type: 'yes_no' | 'measurement' | 'text' | 'photo'
  value: string | number | boolean
  passed: boolean
}

export interface Checklist {
  id: string
  tenant_id: string
  name: string
  description: string
  items: ChecklistItem[]
  version: number
  created_at: string
}

export interface ChecklistItem {
  name: string
  description: string
  check_type: 'yes_no' | 'measurement' | 'text' | 'photo'
  required: boolean
  unit?: string
  min_value?: number
  max_value?: number
}

export interface ElectricalTest {
  id: string
  tenant_id: string
  task_id: string | null
  equipment_id: string
  equipment_name?: string
  tester_id: string
  test_type: 'vde_0701' | 'vde_0702'
  test_date: string
  next_test_date: string
  result: TestResult
  insulation_resistance_mohm: number | null
  protective_conductor_resistance_ohm: number | null
  leakage_current_ma: number | null
  visual_inspection_ok: boolean
  functional_test_ok: boolean
  certificate_number: string
  notes: string
  created_at: string
}

export interface MaintenanceDashboard {
  due_tasks: MaintenanceTask[]
  overdue_tasks: MaintenanceTask[]
  due_plans: MaintenancePlan[]
  recent_tests: ElectricalTest[]
  stats: {
    total_plans: number
    active_plans: number
    tasks_due: number
    tasks_overdue: number
    tests_this_month: number
  }
}

export interface CreateMaintenancePlanDTO {
  equipment_id: string
  plan_type: PlanType
  interval_days?: number
  interval_hours?: number
  name: string
  description?: string
  checklist_template_id?: string
}

export interface CreateMaintenanceTaskDTO {
  plan_id?: string
  equipment_id: string
  assigned_to?: string
  priority: TaskPriority
  scheduled_at: string
  notes?: string
}
