export type WorkflowStatus = 'draft' | 'active' | 'paused' | 'archived'
export type WorkflowStepType = 'trigger' | 'condition' | 'action' | 'notification' | 'decision'
export type WorkflowTrigger = 'manual' | 'scheduled' | 'event' | 'webhook'

export interface WorkflowStep {
  id: string
  type: WorkflowStepType
  name: string
  description: string
  config: Record<string, unknown>
}

export interface WorkflowDefinition {
  id: string
  name: string
  description: string
  status: WorkflowStatus
  trigger: WorkflowTrigger
  steps: WorkflowStep[]
  created_at: string
  updated_at: string
  created_by: string
  instances_count: number
}

export interface WorkflowInstance {
  id: string
  workflow_id: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  started_at: string
  completed_at?: string
  current_step: number
  context: Record<string, unknown>
}

export interface WorkflowTemplate {
  id: string
  name: string
  description: string
  category: string
  steps: WorkflowStep[]
  icon: string
}
