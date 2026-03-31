import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { auditApi } from '../../../services/api'
import styles from '../ProjectDetail.module.scss'

interface HistoryTabProps {
  project: Project
}

interface AuditEntry {
  id: string
  timestamp: string
  user_name?: string
  user_id?: string
  action_type?: string
  action?: string
  entity_type?: string
  entity_id?: string
  description?: string
  details?: string
  changes?: any
}

const ACTION_ICONS: Record<string, string> = {
  created: '\u{2728}',
  status_changed: '\u{1F504}',
  equipment_added: '\u{2795}',
  equipment_removed: '\u{2796}',
  crew_assigned: '\u{1F465}',
  edited: '\u{270F}\u{FE0F}',
  updated: '\u{270F}\u{FE0F}',
  deleted: '\u{1F5D1}\u{FE0F}',
  other: '\u{1F4CC}',
}

const ACTION_LABELS: Record<string, string> = {
  created: 'Erstellt',
  status_changed: 'Status',
  equipment_added: 'Equipment',
  equipment_removed: 'Equipment',
  crew_assigned: 'Crew',
  edited: 'Bearbeitet',
  updated: 'Bearbeitet',
  deleted: 'Geloescht',
  other: 'Sonstig',
}

const FILTER_OPTIONS = [
  { value: '', label: 'Alle' },
  { value: 'created', label: 'Erstellt' },
  { value: 'status_changed', label: 'Status' },
  { value: 'equipment_added', label: 'Equipment' },
  { value: 'crew_assigned', label: 'Crew' },
  { value: 'edited', label: 'Bearbeitet' },
]

export function HistoryTab({ project }: HistoryTabProps) {
  const [filter, setFilter] = useState('')

  // Try to fetch audit logs for this project
  const { data: auditData } = useQuery({
    queryKey: ['project-audit', project.id],
    queryFn: () => auditApi.logs({ entity_type: 'project', entity_id: project.id }),
    retry: false,
  })

  // Parse audit entries from API or fall back to auto-generated
  const apiEntries: AuditEntry[] = Array.isArray(auditData)
    ? auditData
    : auditData?.data ?? auditData?.items ?? []

  // Always include auto-generated creation entry
  const creationEntry: AuditEntry = {
    id: 'auto-created',
    timestamp: project.created_at,
    user_name: 'System',
    action_type: 'created',
    description: `Projekt "${project.name}" erstellt`,
  }

  // Merge: API entries + creation entry (avoid duplicate)
  const hasCreationFromApi = apiEntries.some(
    (e) => (e.action_type === 'created' || e.action === 'created') && e.entity_id === project.id,
  )
  const auditEntries = hasCreationFromApi
    ? apiEntries
    : [...apiEntries, creationEntry]

  // Normalize action_type
  const normalizedEntries = auditEntries.map((e) => ({
    ...e,
    action_type: e.action_type || e.action || 'other',
  }))

  // Sort newest first
  normalizedEntries.sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
  )

  const filtered = filter
    ? normalizedEntries.filter((e) => e.action_type === filter)
    : normalizedEntries

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>Verlauf ({filtered.length})</h3>
      </div>

      <div className={styles.filterBar}>
        {FILTER_OPTIONS.map((opt) => (
          <button
            key={opt.value}
            className={`${styles.filterBtn} ${filter === opt.value ? styles.filterBtnActive : ''}`}
            onClick={() => setFilter(opt.value)}
          >
            {opt.label}
          </button>
        ))}
      </div>

      {filtered.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F4DC}'}</div>
          <h4 className={styles.emptyStateTitle}>Kein Verlauf</h4>
          <p className={styles.emptyStateText}>
            Es wurden noch keine Aktivitaeten fuer dieses Projekt erfasst.
          </p>
        </div>
      ) : (
        <div className={styles.timeline}>
          {filtered.map((entry) => (
            <div key={entry.id} className={styles.timelineItem}>
              <div className={styles.timelineIcon}>
                {ACTION_ICONS[entry.action_type!] || ACTION_ICONS.other}
              </div>
              <div className={styles.timelineContent}>
                <div className={styles.timelineHeader}>
                  <span className={styles.timelineUser}>
                    {entry.user_name || entry.user_id || 'System'}
                  </span>
                  <span style={{
                    fontSize: 'var(--font-size-xs)',
                    padding: 'var(--spacing-1) var(--spacing-2)',
                    background: 'rgba(0, 212, 255, 0.1)',
                    borderRadius: 'var(--radius-full)',
                    color: 'var(--color-primary)',
                  }}>
                    {ACTION_LABELS[entry.action_type!] || 'Sonstig'}
                  </span>
                  <span className={styles.timelineTime}>
                    {new Date(entry.timestamp).toLocaleString('de-DE')}
                  </span>
                </div>
                <div className={styles.timelineDescription}>
                  {entry.description || entry.details || `${entry.action_type} auf ${entry.entity_type || 'Projekt'}`}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
