import { Project } from '../../../types/project'
import styles from '../ProjectDetail.module.scss'

interface CommunicationTabProps {
  project: Project
}

export function CommunicationTab({ project: _project }: CommunicationTabProps) {
  // Communication/notes feature is not yet backed by an API endpoint.
  // Show a clean empty state for now.

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>Kommunikation</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-2)' }}>
          <button className="btn btn--secondary" disabled>Notiz hinzufuegen</button>
          <button className="btn btn--primary" disabled>E-Mail senden</button>
        </div>
      </div>

      <div className={styles.emptyState}>
        <div className={styles.emptyStateIcon}>{'\u{1F4AC}'}</div>
        <h4 className={styles.emptyStateTitle}>Keine Eintraege</h4>
        <p className={styles.emptyStateText}>
          Noch keine Kommunikation zu diesem Projekt. Fuegen Sie eine Notiz hinzu oder senden Sie eine E-Mail.
        </p>
      </div>
    </div>
  )
}
