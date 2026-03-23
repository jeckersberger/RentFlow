import { Project } from '../../../types/project'
import styles from '../ProjectDetail.module.scss'

interface CrewTabProps {
  project: Project
}

interface CrewPosition {
  id: string
  role: string
  assigned_person: string | null
  planned_hours: number
  rate: number
  available: boolean
}

export function CrewTab({ project: _project }: CrewTabProps) {
  // Placeholder: no crew data from API yet
  const positions: CrewPosition[] = []

  const totalCost = positions.reduce(
    (sum, p) => sum + (p.planned_hours * p.rate),
    0,
  )

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>Crew & Team</h3>
        <button className="btn btn--primary" disabled>
          + Position hinzufügen
        </button>
      </div>

      {positions.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F465}'}</div>
          <h4 className={styles.emptyStateTitle}>Keine Crew-Positionen</h4>
          <p className={styles.emptyStateText}>
            Fügen Sie Positionen wie Soundtechniker, Lichttechniker oder Stagehand hinzu
            und weisen Sie Crew-Mitglieder zu.
          </p>
        </div>
      ) : (
        <>
          <div style={{ overflowX: 'auto' }}>
            <table className={styles.dataTable}>
              <thead>
                <tr>
                  <th>Rolle</th>
                  <th>Zugewiesene Person</th>
                  <th style={{ textAlign: 'right' }}>Stunden</th>
                  <th style={{ textAlign: 'right' }}>Stundensatz</th>
                  <th style={{ textAlign: 'right' }}>Kosten</th>
                  <th style={{ textAlign: 'center' }}>Status</th>
                </tr>
              </thead>
              <tbody>
                {positions.map((pos) => (
                  <tr key={pos.id}>
                    <td>{pos.role}</td>
                    <td>{pos.assigned_person || <span style={{ color: 'var(--color-text-muted)' }}>Nicht zugewiesen</span>}</td>
                    <td style={{ textAlign: 'right' }}>{pos.planned_hours}h</td>
                    <td style={{ textAlign: 'right' }}>{'\u20AC'}{pos.rate.toFixed(2)}/h</td>
                    <td style={{ textAlign: 'right' }}>{'\u20AC'}{(pos.planned_hours * pos.rate).toFixed(2)}</td>
                    <td style={{ textAlign: 'center' }}>
                      <span
                        className={`${styles.availabilityDot} ${
                          pos.available
                            ? styles.availabilityDotAvailable
                            : styles.availabilityDotConflict
                        }`}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr>
                  <td colSpan={4}>Gesamtkosten Crew</td>
                  <td style={{ textAlign: 'right' }}>{'\u20AC'}{totalCost.toFixed(2)}</td>
                  <td />
                </tr>
              </tfoot>
            </table>
          </div>
        </>
      )}
    </div>
  )
}
