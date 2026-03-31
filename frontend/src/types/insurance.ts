export type ClaimStatus = 'reported' | 'documented' | 'submitted' | 'approved' | 'settled' | 'rejected'
export type PolicyType = 'liability' | 'equipment' | 'vehicle' | 'workers_comp'

export interface InsurancePolicy {
  id: string
  policy_number: string
  type: PolicyType
  provider: string
  coverage_amount: number
  premium_annual: number
  start_date: string
  end_date: string
  is_active: boolean
  deductible: number
}

export interface ClaimItem {
  id: string
  description: string
  claimed_amount: number
  approved_amount?: number
  cost_category: string
  receipt_url?: string
}

export interface InsuranceClaim {
  id: string
  claim_number: string
  policy_id: string
  status: ClaimStatus
  reported_date: string
  incident_date: string
  description: string
  total_claimed: number
  total_approved?: number
  total_settled?: number
  items: ClaimItem[]
  photos: string[]
  assigned_adjuster?: string
}

export interface ClaimStatusTimeline {
  status: ClaimStatus
  timestamp: string
  notes?: string
}
