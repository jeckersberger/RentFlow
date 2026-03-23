import { Project } from '../../../types/project'
import styles from '../ProjectDetail.module.scss'

interface CommunicationTabProps {
  project: Project
}

interface CommunicationEntry {
  id: string
  type: 'email' | 'note' | 'status_change'
  timestamp: string
  from: string
  to?: string
  subject?: string
  content: string
}

const TYPE_ICONS: Record<string, string> = {
  email: '\u{1F4E7}',
  note: '\u{1F4DD}',
  status_change: '\u{1F504}',
}

export function CommunicationTab({ project: _project }: CommunicationTabProps) {
  // Placeholder: no communication data from API yet
  const entries: CommunicationEntry[] = []

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>Kommunikation</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-2)' }}>
          <button className="btn btn--secondary" disabled>Notiz hinzufügen</button>
          <button className="btn btn--primary" disabled>E-Mail senden</button>
        </div>
      </div>

      {entries.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F4AC}'}</div>
          <h4 className={styles.emptyStateTitle}>Noch keine Kommunikation</h4>
          <p className={styles.emptyStateText}>
            Noch keine Kommunikation zu diesem Projekt. Fügen Sie eine Notiz hinzu oder senden Sie eine E-Mail.
          </p>
        </div>
      ) : (
        <div className={styles.timeline}>
          {entries.map((entry) => (
            <div key={entry.id} className={styles.timelineItem}>
              <div className={styles.timelineIcon}>
                {TYPE_ICONS[entry.type] || '\u{1F4AC}'}
              </div>
              <div className={styles.timelineContent}>
                <div className={styles.timelineHeader}>
                  <span className={styles.timelineUser}>{entry.from}</span>
                  {entry.to && (
                    <span style={{ color: 'var(--color-text-muted)', fontSize: 'var(--font-size-xs)' }}>
                      an {entry.to}
                    </span>
                  )}
                  <span className={styles.timelineTime}>
                    {new Date(entry.timestamp).toLocaleString('de-DE')}
                  </span>
                </div>
                {entry.subject && (
                  <div style={{ fontWeight: 'var(--font-weight-medium)', color: 'var(--color-text-primary)', fontSize: 'var(--font-size-sm)', marginBottom: 'var(--spacing-1)' }}>
                    {entry.subject}
                  </div>
                )}
                <div className={styles.timelineDescription}>{entry.content}</div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
