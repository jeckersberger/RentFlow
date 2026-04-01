import { useQuery } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { transportApi } from '../../../services/api'
import '../ProjectDetail.scss'

interface TransportTabProps {
  project: Project
}

interface Tour {
  id: string
  vehicle_id: string
  project_id: string
  driver_id?: string
  status: string
  departure_at?: string
  arrival_at?: string
  km_start?: number
  km_end?: number
  total_cost?: number
  notes?: string
  created_at?: string
}

const STATUS_LABELS: Record<string, string> = {
  planned: 'Geplant',
  in_transit: 'Unterwegs',
  completed: 'Abgeschlossen',
  cancelled: 'Storniert',
}

const STATUS_COLORS: Record<string, { bg: string; color: string }> = {
  planned: { bg: 'rgba(59, 130, 246, 0.15)', color: '#3b82f6' },
  in_transit: { bg: 'rgba(245, 158, 11, 0.15)', color: '#f59e0b' },
  completed: { bg: 'rgba(16, 185, 129, 0.15)', color: 'var(--color-success)' },
  cancelled: { bg: 'rgba(107, 114, 128, 0.15)', color: 'var(--color-text-muted)' },
}

export function TransportTab({ project }: TransportTabProps) {
  const { data: toursData, isLoading, error } = useQuery({
    queryKey: ['project-tours', project.id],
    queryFn: () => transportApi.listTours(1, 50),
  })

  // Filter tours for this project
  const allTours: Tour[] = Array.isArray(toursData)
    ? toursData
    : toursData?.data ?? toursData?.items ?? []

  const projectTours = allTours.filter((t) => t.project_id === project.id)

  const formatDateTime = (d?: string) =>
    d ? new Date(d).toLocaleString('de-DE', { dateStyle: 'medium', timeStyle: 'short' }) : '\u2014'

  return (
    <div>
      <div className="sectionHeader">
        <h3 className="sectionTitle">
          Transport
          {!isLoading && ` (${projectTours.length})`}
        </h3>
        <button className="btn btn--primary" disabled>
          + Fahrzeug zuweisen
        </button>
      </div>

      {isLoading ? (
        <div className="emptyState">
          <p className="emptyStateText">Transport-Daten werden geladen...</p>
        </div>
      ) : error ? (
        <div className="emptyState">
          <div className="emptyStateIcon">{'\u26A0\uFE0F'}</div>
          <h4 className="emptyStateTitle">Fehler beim Laden</h4>
          <p className="emptyStateText">
            Transport-Daten konnten nicht geladen werden.
          </p>
        </div>
      ) : projectTours.length === 0 ? (
        <div className="emptyState">
          <div className="emptyStateIcon">{'\u{1F69A}'}</div>
          <h4 className="emptyStateTitle">Keine Fahrzeuge zugewiesen</h4>
          <p className="emptyStateText">
            Weisen Sie Fahrzeuge zu und planen Sie den Transport fuer dieses Projekt.
          </p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-4)' }}>
          {projectTours.map((tour) => {
            const sc = STATUS_COLORS[tour.status] || STATUS_COLORS.planned
            const distanceKm = (tour.km_end && tour.km_start)
              ? tour.km_end - tour.km_start
              : null
            return (
              <div key={tour.id} className="glassCard">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-3)' }}>
                  <div className="glassCardTitle" style={{ margin: 0 }}>
                    Tour #{tour.id.slice(0, 8)}
                  </div>
                  <span
                    style={{
                      display: 'inline-block',
                      padding: 'var(--spacing-1) var(--spacing-3)',
                      borderRadius: 'var(--radius-full)',
                      fontSize: 'var(--font-size-xs)',
                      fontWeight: 'var(--font-weight-medium)',
                      background: sc.bg,
                      color: sc.color,
                    }}
                  >
                    {STATUS_LABELS[tour.status] || tour.status}
                  </span>
                </div>
                <div className="overviewGrid" style={{ gridTemplateColumns: '1fr 1fr 1fr' }}>
                  <div>
                    <div className="infoRow">
                      <span className="infoLabel">Fahrzeug</span>
                      <span className="infoValue">{tour.vehicle_id}</span>
                    </div>
                    {tour.driver_id && (
                      <div className="infoRow">
                        <span className="infoLabel">Fahrer</span>
                        <span className="infoValue">{tour.driver_id}</span>
                      </div>
                    )}
                  </div>
                  <div>
                    <div className="infoRow">
                      <span className="infoLabel">Abfahrt</span>
                      <span className="infoValue">{formatDateTime(tour.departure_at)}</span>
                    </div>
                    <div className="infoRow">
                      <span className="infoLabel">Ankunft</span>
                      <span className="infoValue">{formatDateTime(tour.arrival_at)}</span>
                    </div>
                  </div>
                  <div>
                    {distanceKm !== null && (
                      <div className="infoRow">
                        <span className="infoLabel">Entfernung</span>
                        <span className="infoValue">{distanceKm} km</span>
                      </div>
                    )}
                    {tour.total_cost != null && tour.total_cost > 0 && (
                      <div className="infoRow">
                        <span className="infoLabel">Kosten</span>
                        <span className="infoValue">{'\u20AC'}{tour.total_cost.toFixed(2)}</span>
                      </div>
                    )}
                  </div>
                </div>
                {tour.notes && (
                  <div style={{ marginTop: 'var(--spacing-3)', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                    {tour.notes}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
