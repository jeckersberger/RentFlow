import { useState } from 'react'
import { Project } from '../../../types/project'
import { PackingListTab } from './PackingListTab'
import styles from '../ProjectDetail.module.scss'

interface EquipmentTabProps {
  project: Project
}

// Placeholder equipment type until real data model exists
interface ProjectEquipment {
  id: string
  name: string
  quantity: number
  unit_price: number
  total: number
  tax_rate: number
  status: 'available' | 'conflict'
}

export function EquipmentTab({ project }: EquipmentTabProps) {
  const [search, setSearch] = useState('')
  const [showPackingList, setShowPackingList] = useState(false)

  // Placeholder: no equipment data from API yet
  const equipment: ProjectEquipment[] = []

  const filtered = equipment.filter((item) =>
    item.name.toLowerCase().includes(search.toLowerCase()),
  )

  const subtotal = filtered.reduce((sum, item) => sum + item.total, 0)

  return (
    <div>
      {/* Mode toggle: Equipment vs Packliste */}
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>
          {showPackingList ? 'Packliste' : 'Equipment'}
        </h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-2)', alignItems: 'center' }}>
          <button
            className={`btn ${showPackingList ? 'btn--primary' : 'btn--secondary'}`}
            onClick={() => setShowPackingList(!showPackingList)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              fontSize: 'var(--font-size-sm)',
              padding: 'var(--spacing-2) var(--spacing-4)',
            }}
          >
            {showPackingList ? (
              <>
                {'\u{1F4CB}'} Zurueck zur Equipment-Ansicht
              </>
            ) : (
              <>
                {'\u{1F4E6}'} Packliste oeffnen
              </>
            )}
          </button>
          {!showPackingList && (
            <button className="btn btn--primary" disabled>
              + Equipment hinzufuegen
            </button>
          )}
        </div>
      </div>

      {showPackingList ? (
        <PackingListTab project={project} />
      ) : (
        <>
          <div className={styles.searchBar}>
            <input
              type="text"
              placeholder="Equipment suchen..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>

          {filtered.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyStateIcon}>{'\u{1F4E6}'}</div>
              <h4 className={styles.emptyStateTitle}>Kein Equipment zugewiesen</h4>
              <p className={styles.emptyStateText}>
                Klicken Sie auf &laquo;Equipment hinzufuegen&raquo;, um Artikel zu diesem Projekt zuzuweisen.
              </p>
            </div>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table className={styles.dataTable}>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th style={{ textAlign: 'right' }}>Menge</th>
                    <th style={{ textAlign: 'right' }}>Einzelpreis</th>
                    <th style={{ textAlign: 'right' }}>Gesamt</th>
                    <th style={{ textAlign: 'right' }}>MwSt.</th>
                    <th style={{ textAlign: 'center' }}>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((item) => (
                    <tr key={item.id}>
                      <td>{item.name}</td>
                      <td style={{ textAlign: 'right' }}>{item.quantity}</td>
                      <td style={{ textAlign: 'right' }}>{'\u20AC'}{item.unit_price.toFixed(2)}</td>
                      <td style={{ textAlign: 'right' }}>{'\u20AC'}{item.total.toFixed(2)}</td>
                      <td style={{ textAlign: 'right' }}>{item.tax_rate}%</td>
                      <td style={{ textAlign: 'center' }}>
                        <span
                          className={`${styles.availabilityDot} ${
                            item.status === 'available'
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
                    <td colSpan={3}>Zwischensumme</td>
                    <td style={{ textAlign: 'right' }}>{'\u20AC'}{subtotal.toFixed(2)}</td>
                    <td colSpan={2} />
                  </tr>
                </tfoot>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  )
}
