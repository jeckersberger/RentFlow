import { useQuery } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { documentApi } from '../../../services/api'
import styles from '../ProjectDetail.module.scss'

interface DocumentsTabProps {
  project: Project
}

interface ProjectDocument {
  id: string
  document_type?: string
  type?: string
  reference_id?: string
  project_id?: string
  document_number?: string
  title?: string
  name?: string
  status?: string
  created_at: string
  updated_at?: string
}

const DOC_TYPE_LABELS: Record<string, string> = {
  angebot: 'Angebot',
  rechnung: 'Rechnung',
  packliste: 'Packliste',
  lieferschein: 'Lieferschein',
  carnet: 'Carnet',
  quote: 'Angebot',
  invoice: 'Rechnung',
  packing_list: 'Packliste',
  delivery_note: 'Lieferschein',
}

const DOC_TYPE_ICONS: Record<string, string> = {
  angebot: '\u{1F4CB}',
  rechnung: '\u{1F4B0}',
  packliste: '\u{1F4E6}',
  lieferschein: '\u{1F69A}',
  carnet: '\u{1F4C4}',
  quote: '\u{1F4CB}',
  invoice: '\u{1F4B0}',
  packing_list: '\u{1F4E6}',
  delivery_note: '\u{1F69A}',
}

export function DocumentsTab({ project }: DocumentsTabProps) {
  const { data: docsData, isLoading, error } = useQuery({
    queryKey: ['project-documents', project.id],
    queryFn: () => documentApi.list(),
  })

  // Filter documents for this project
  const allDocs: ProjectDocument[] = Array.isArray(docsData)
    ? docsData
    : docsData?.data ?? docsData?.items ?? []

  const documents = allDocs.filter(
    (doc) =>
      doc.reference_id === project.id ||
      doc.project_id === project.id,
  )

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>
          Dokumente
          {!isLoading && ` (${documents.length})`}
        </h3>
        <button className="btn btn--primary" disabled>
          + Dokument erstellen
        </button>
      </div>

      {isLoading ? (
        <div className={styles.emptyState}>
          <p className={styles.emptyStateText}>Dokumente werden geladen...</p>
        </div>
      ) : error ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u26A0\uFE0F'}</div>
          <h4 className={styles.emptyStateTitle}>Fehler beim Laden</h4>
          <p className={styles.emptyStateText}>
            Dokumente konnten nicht geladen werden.
          </p>
        </div>
      ) : documents.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F4C4}'}</div>
          <h4 className={styles.emptyStateTitle}>Noch keine Dokumente</h4>
          <p className={styles.emptyStateText}>
            Erstellen Sie ein Angebot oder eine Packliste fuer dieses Projekt.
          </p>
        </div>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <table className={styles.dataTable}>
            <thead>
              <tr>
                <th>Typ</th>
                <th>Name</th>
                <th>Nummer</th>
                <th>Erstellt</th>
                <th style={{ textAlign: 'center' }}>Status</th>
                <th style={{ textAlign: 'right' }}>Aktionen</th>
              </tr>
            </thead>
            <tbody>
              {documents.map((doc) => {
                const docType = doc.document_type || doc.type || 'other'
                return (
                  <tr key={doc.id}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)' }}>
                        <span className={styles.docTypeIcon}>
                          {DOC_TYPE_ICONS[docType] || '\u{1F4C4}'}
                        </span>
                        {DOC_TYPE_LABELS[docType] || docType}
                      </div>
                    </td>
                    <td>{doc.title || doc.name || '\u2014'}</td>
                    <td style={{ color: 'var(--color-text-muted)', fontSize: 'var(--font-size-sm)' }}>
                      {doc.document_number || '\u2014'}
                    </td>
                    <td>{new Date(doc.created_at).toLocaleDateString('de-DE')}</td>
                    <td style={{ textAlign: 'center' }}>
                      {doc.status && (
                        <span
                          style={{
                            display: 'inline-block',
                            padding: 'var(--spacing-1) var(--spacing-2)',
                            borderRadius: 'var(--radius-full)',
                            fontSize: 'var(--font-size-xs)',
                            fontWeight: 'var(--font-weight-medium)',
                            background: 'rgba(0, 212, 255, 0.1)',
                            color: 'var(--color-primary)',
                          }}
                        >
                          {doc.status}
                        </span>
                      )}
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      <button className="btn btn--secondary" style={{ padding: 'var(--spacing-1) var(--spacing-3)', fontSize: 'var(--font-size-xs)' }}>
                        Herunterladen
                      </button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
