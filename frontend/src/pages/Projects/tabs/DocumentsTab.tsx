import { Project } from '../../../types/project'
import styles from '../ProjectDetail.module.scss'

interface DocumentsTabProps {
  project: Project
}

interface ProjectDocument {
  id: string
  type: 'angebot' | 'rechnung' | 'packliste' | 'lieferschein' | 'carnet'
  name: string
  created_at: string
  status: string
}

const DOC_TYPE_LABELS: Record<string, string> = {
  angebot: 'Angebot',
  rechnung: 'Rechnung',
  packliste: 'Packliste',
  lieferschein: 'Lieferschein',
  carnet: 'Carnet',
}

const DOC_TYPE_ICONS: Record<string, string> = {
  angebot: '\u{1F4CB}',
  rechnung: '\u{1F4B0}',
  packliste: '\u{1F4E6}',
  lieferschein: '\u{1F69A}',
  carnet: '\u{1F4C4}',
}

export function DocumentsTab({ project: _project }: DocumentsTabProps) {
  // Placeholder: no document data from API yet
  const documents: ProjectDocument[] = []

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>Dokumente</h3>
        <button className="btn btn--primary" disabled>
          + Dokument erstellen
        </button>
      </div>

      {documents.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F4C4}'}</div>
          <h4 className={styles.emptyStateTitle}>Noch keine Dokumente</h4>
          <p className={styles.emptyStateText}>
            Erstellen Sie ein Angebot oder eine Packliste für dieses Projekt.
          </p>
        </div>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <table className={styles.dataTable}>
            <thead>
              <tr>
                <th>Typ</th>
                <th>Name</th>
                <th>Erstellt</th>
                <th>Status</th>
                <th style={{ textAlign: 'right' }}>Aktionen</th>
              </tr>
            </thead>
            <tbody>
              {documents.map((doc) => (
                <tr key={doc.id}>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)' }}>
                      <span className={styles.docTypeIcon}>
                        {DOC_TYPE_ICONS[doc.type] || '\u{1F4C4}'}
                      </span>
                      {DOC_TYPE_LABELS[doc.type] || doc.type}
                    </div>
                  </td>
                  <td>{doc.name}</td>
                  <td>{new Date(doc.created_at).toLocaleDateString('de-DE')}</td>
                  <td>{doc.status}</td>
                  <td style={{ textAlign: 'right' }}>
                    <button className="btn btn--secondary" style={{ padding: 'var(--spacing-1) var(--spacing-3)', fontSize: 'var(--font-size-xs)' }}>
                      Herunterladen
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
