import { useQuery } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { crewApi } from '../../../services/api'
import styles from '../ProjectDetail.module.scss'

interface CrewTabProps {
  project: Project
}

interface CrewAssignment {
  id: string
  project_id: string
  crew_member_id: string
  crew_member_name?: string
  first_name?: string
  last_name?: string
  role: string
  hours?: number
  planned_hours?: number
  daily_rate?: number
  hourly_rate?: number
  rate?: number
  status?: string
}

export function CrewTab({ project }: CrewTabProps) {
  const { data: assignmentsData, isLoading, error } = useQuery({
    queryKey: ['project-crew-assignments', project.id],
    queryFn: () => crewApi.listAssignments({ project_id: project.id } as any),
  })

  // Handle different response shapes
  const assignments: CrewAssignment[] = Array.isArray(assignmentsData)
    ? assignmentsData
    : assignmentsData?.data ?? assignmentsData?.items ?? []

  const totalCost = assignments.reduce((sum, p) => {
    const hours = p.planned_hours || p.hours || 0
    const rate = p.rate || p.hourly_rate || p.daily_rate || 0
    return sum + (hours * rate)
  }, 0)

  const getDisplayName = (a: CrewAssignment) => {
    if (a.crew_member_name) return a.crew_member_name
    if (a.first_name || a.last_name) return `${a.first_name || ''} ${a.last_name || ''}`.trim()
    return null
  }

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>
          Crew & Team
          {!isLoading && ` (${assignments.length})`}
        </h3>
        <button className="btn btn--primary" disabled>
          + Position hinzufuegen
        </button>
      </div>

      {isLoading ? (
        <div className={styles.emptyState}>
          <p className={styles.emptyStateText}>Crew-Daten werden geladen...</p>
        </div>
      ) : error ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u26A0\uFE0F'}</div>
          <h4 className={styles.emptyStateTitle}>Fehler beim Laden</h4>
          <p className={styles.emptyStateText}>
            Crew-Daten konnten nicht geladen werden.
          </p>
        </div>
      ) : assignments.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F465}'}</div>
          <h4 className={styles.emptyStateTitle}>Keine Crew-Positionen</h4>
          <p className={styles.emptyStateText}>
            Fuegen Sie Positionen wie Soundtechniker, Lichttechniker oder Stagehand hinzu
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
                  {assignments.some(a => a.rate || a.hourly_rate || a.daily_rate) && (
                    <>
                      <th style={{ textAlign: 'right' }}>Stundensatz</th>
                      <th style={{ textAlign: 'right' }}>Kosten</th>
                    </>
                  )}
                  <th style={{ textAlign: 'center' }}>Status</th>
                </tr>
              </thead>
              <tbody>
                {assignments.map((pos) => {
                  const name = getDisplayName(pos)
                  const hours = pos.planned_hours || pos.hours || 0
                  const rate = pos.rate || pos.hourly_rate || pos.daily_rate || 0
                  const hasRates = assignments.some(a => a.rate || a.hourly_rate || a.daily_rate)
                  return (
                    <tr key={pos.id}>
                      <td>{pos.role || 'Techniker'}</td>
                      <td>
                        {name || (
                          <span style={{ color: 'var(--color-text-muted)' }}>Nicht zugewiesen</span>
                        )}
                      </td>
                      <td style={{ textAlign: 'right' }}>{hours > 0 ? `${hours}h` : '\u2014'}</td>
                      {hasRates && (
                        <>
                          <td style={{ textAlign: 'right' }}>
                            {rate > 0 ? `\u20AC${rate.toFixed(2)}/h` : '\u2014'}
                          </td>
                          <td style={{ textAlign: 'right' }}>
                            {hours > 0 && rate > 0
                              ? `\u20AC${(hours * rate).toFixed(2)}`
                              : '\u2014'}
                          </td>
                        </>
                      )}
                      <td style={{ textAlign: 'center' }}>
                        <span
                          style={{
                            display: 'inline-block',
                            padding: 'var(--spacing-1) var(--spacing-2)',
                            borderRadius: 'var(--radius-full)',
                            fontSize: 'var(--font-size-xs)',
                            fontWeight: 'var(--font-weight-medium)',
                            background: pos.status === 'confirmed'
                              ? 'rgba(16, 185, 129, 0.15)'
                              : 'rgba(0, 212, 255, 0.1)',
                            color: pos.status === 'confirmed'
                              ? 'var(--color-success)'
                              : 'var(--color-primary)',
                          }}
                        >
                          {pos.status || 'zugewiesen'}
                        </span>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
              {totalCost > 0 && (
                <tfoot>
                  <tr>
                    <td colSpan={4}>Gesamtkosten Crew</td>
                    <td style={{ textAlign: 'right' }}>{'\u20AC'}{totalCost.toFixed(2)}</td>
                    <td />
                  </tr>
                </tfoot>
              )}
            </table>
          </div>
        </>
      )}
    </div>
  )
}
