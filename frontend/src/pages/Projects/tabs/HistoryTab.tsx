import { useState } from 'react'
import { Project } from '../../../types/project'
import styles from '../ProjectDetail.module.scss'

interface HistoryTabProps {
  project: Project
}

interface AuditEntry {
  id: string
  timestamp: string
  user_name: string
  action_type: 'created' | 'status_changed' | 'equipment_added' | 'equipment_removed' | 'crew_assigned' | 'edited' | 'other'
  description: string
}

const ACTION_ICONS: Record<string, string> = {
  created: '\u{2728}',
  status_changed: '\u{1F504}',
  equipment_added: '\u{2795}',
  equipment_removed: '\u{2796}',
  crew_assigned: '\u{1F465}',
  edited: '\u{270F}\u{FE0F}',
  other: '\u{1F4CC}',
}

const ACTION_LABELS: Record<string, string> = {
  created: 'Erstellt',
  status_changed: 'Status',
  equipment_added: 'Equipment',
  equipment_removed: 'Equipment',
  crew_assigned: 'Crew',
  edited: 'Bearbeitet',
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

  // Auto-generate a creation entry from project data
  const auditEntries: AuditEntry[] = [
    {
      id: 'auto-created',
      timestamp: project.created_at,
      user_name: 'System',
      action_type: 'created',
      description: `Projekt "${project.name}" erstellt`,
    },
  ]

  const filtered = filter
    ? auditEntries.filter((e) => e.action_type === filter)
    : auditEntries

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>Verlauf</h3>
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
            Es wurden noch keine Aktivitäten für dieses Projekt erfasst.
          </p>
        </div>
      ) : (
        <div className={styles.timeline}>
          {filtered.map((entry) => (
            <div key={entry.id} className={styles.timelineItem}>
              <div className={styles.timelineIcon}>
                {ACTION_ICONS[entry.action_type] || ACTION_ICONS.other}
              </div>
              <div className={styles.timelineContent}>
                <div className={styles.timelineHeader}>
                  <span className={styles.timelineUser}>{entry.user_name}</span>
                  <span style={{
                    fontSize: 'var(--font-size-xs)',
                    padding: 'var(--spacing-1) var(--spacing-2)',
                    background: 'rgba(0, 212, 255, 0.1)',
                    borderRadius: 'var(--radius-full)',
                    color: 'var(--color-primary)',
                  }}>
                    {ACTION_LABELS[entry.action_type] || 'Sonstig'}
                  </span>
                  <span className={styles.timelineTime}>
                    {new Date(entry.timestamp).toLocaleString('de-DE')}
                  </span>
                </div>
                <div className={styles.timelineDescription}>{entry.description}</div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
