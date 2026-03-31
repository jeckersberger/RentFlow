export type AuditOperation = 'CREATE' | 'READ' | 'UPDATE' | 'DELETE' | 'EXPORT' | 'LOGIN' | 'LOGOUT'
export type AuditEntityType = 'equipment' | 'project' | 'crew' | 'invoice' | 'user' | 'system'

export interface AuditEntry {
  id: string
  timestamp: string
  service: string
  operation: AuditOperation
  entity_type: AuditEntityType
  entity_id: string
  user_id: string
  user_name: string
  changes: Record<string, unknown>
  status: 'success' | 'failure'
  checksum: string
  ip_address: string
}

export interface AuditFilter {
  start_date?: string
  end_date?: string
  entity_type?: AuditEntityType
  operation?: AuditOperation
  user_id?: string
  search_query?: string
}

export interface ChainValidationResult {
  valid: boolean
  entries_checked: number
  invalid_entries: number
  timestamp: string
  verification_hash: string
}

export interface GoBDExportData {
  format: string
  entries_count: number
  date_range: string
  checksum: string
  signature: string
}
