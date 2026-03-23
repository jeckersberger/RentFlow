import { Project } from '../../../types/project'
import styles from '../ProjectDetail.module.scss'

interface TransportTabProps {
  project: Project
}

interface VehicleAssignment {
  id: string
  name: string
  type: string
  capacity: string
  driver: string
  departure_time: string
  return_time: string
  distance_km: number
  travel_time_min: number
  equipment_items: string[]
}

export function TransportTab({ project: _project }: TransportTabProps) {
  // Placeholder: no transport data from API yet
  const vehicles: VehicleAssignment[] = []

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>Transport</h3>
        <button className="btn btn--primary" disabled>
          + Fahrzeug zuweisen
        </button>
      </div>

      {vehicles.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F69A}'}</div>
          <h4 className={styles.emptyStateTitle}>Keine Fahrzeuge zugewiesen</h4>
          <p className={styles.emptyStateText}>
            Weisen Sie Fahrzeuge zu und planen Sie den Transport für dieses Projekt.
          </p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-4)' }}>
          {vehicles.map((vehicle) => (
            <div key={vehicle.id} className={styles.glassCard}>
              <div className={styles.glassCardTitle}>
                {vehicle.name} &mdash; {vehicle.type}
              </div>
              <div className={styles.overviewGrid} style={{ gridTemplateColumns: '1fr 1fr 1fr' }}>
                <div>
                  <div className={styles.infoRow}>
                    <span className={styles.infoLabel}>Kapazität</span>
                    <span className={styles.infoValue}>{vehicle.capacity}</span>
                  </div>
                  <div className={styles.infoRow}>
                    <span className={styles.infoLabel}>Fahrer</span>
                    <span className={styles.infoValue}>{vehicle.driver}</span>
                  </div>
                </div>
                <div>
                  <div className={styles.infoRow}>
                    <span className={styles.infoLabel}>Abfahrt</span>
                    <span className={styles.infoValue}>{vehicle.departure_time}</span>
                  </div>
                  <div className={styles.infoRow}>
                    <span className={styles.infoLabel}>Rückkehr</span>
                    <span className={styles.infoValue}>{vehicle.return_time}</span>
                  </div>
                </div>
                <div>
                  <div className={styles.infoRow}>
                    <span className={styles.infoLabel}>Entfernung</span>
                    <span className={styles.infoValue}>{vehicle.distance_km} km</span>
                  </div>
                  <div className={styles.infoRow}>
                    <span className={styles.infoLabel}>Fahrzeit</span>
                    <span className={styles.infoValue}>{vehicle.travel_time_min} min</span>
                  </div>
                </div>
              </div>
              {vehicle.equipment_items.length > 0 && (
                <div style={{ marginTop: 'var(--spacing-3)' }}>
                  <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', marginBottom: 'var(--spacing-2)', textTransform: 'uppercase', letterSpacing: 'var(--letter-spacing-wide)' }}>
                    Ladeliste
                  </div>
                  <ul style={{ margin: 0, paddingLeft: 'var(--spacing-4)', color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>
                    {vehicle.equipment_items.map((item, i) => (
                      <li key={i}>{item}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
