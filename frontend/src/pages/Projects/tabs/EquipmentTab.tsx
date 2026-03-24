import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { api } from '../../../services/api'
import { PackingListTab } from './PackingListTab'
import styles from '../ProjectDetail.module.scss'

interface EquipmentTabProps {
  project: Project
}

interface Reservation {
  id: string
  project_id: string
  equipment_id: string
  equipment_name?: string
  quantity?: number
  start_date: string
  end_date: string
  status: string
  unit_price?: number
  total?: number
  tax_rate?: number
}

export function EquipmentTab({ project }: EquipmentTabProps) {
  const [search, setSearch] = useState('')
  const [showPackingList, setShowPackingList] = useState(false)

  const { data: reservations = [], isLoading, error } = useQuery({
    queryKey: ['project-reservations', project.id],
    queryFn: async () => {
      const res = await api.get('/api/v1/reservations', {
        params: { project_id: project.id },
      })
      // API may return array directly or wrapped in {data: [...]}
      const raw = res.data
      return Array.isArray(raw) ? raw : (raw?.data ?? raw?.items ?? [])
    },
  })

  const filtered = (reservations as Reservation[]).filter((item) =>
    (item.equipment_name || item.equipment_id || '')
      .toLowerCase()
      .includes(search.toLowerCase()),
  )

  const subtotal = filtered.reduce((sum, item) => sum + (item.total || item.unit_price || 0), 0)

  const formatDate = (d: string) =>
    new Date(d).toLocaleDateString('de-DE')

  const statusLabel = (s: string) => {
    const map: Record<string, string> = {
      confirmed: 'Bestätigt',
      pending: 'Ausstehend',
      cancelled: 'Storniert',
      checked_out: 'Ausgegeben',
      returned: 'Zurückgegeben',
    }
    return map[s] || s
  }

  return (
    <div>
      {/* Mode toggle: Equipment vs Packliste */}
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>
          {showPackingList ? 'Packliste' : 'Equipment'}
          {!showPackingList && !isLoading && ` (${filtered.length})`}
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

          {isLoading ? (
            <div className={styles.emptyState}>
              <p className={styles.emptyStateText}>Equipment wird geladen...</p>
            </div>
          ) : error ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyStateIcon}>{'\u26A0\uFE0F'}</div>
              <h4 className={styles.emptyStateTitle}>Fehler beim Laden</h4>
              <p className={styles.emptyStateText}>
                Equipment-Daten konnten nicht geladen werden.
              </p>
            </div>
          ) : filtered.length === 0 ? (
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
                    <th>Equipment</th>
                    <th>Zeitraum</th>
                    <th style={{ textAlign: 'center' }}>Status</th>
                    {filtered.some(i => i.unit_price) && (
                      <th style={{ textAlign: 'right' }}>Preis</th>
                    )}
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((item) => (
                    <tr key={item.id}>
                      <td>
                        <div style={{ fontWeight: 'var(--font-weight-medium)' }}>
                          {item.equipment_name || item.equipment_id}
                        </div>
                        {item.quantity && item.quantity > 1 && (
                          <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)' }}>
                            Menge: {item.quantity}
                          </div>
                        )}
                      </td>
                      <td>
                        <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                          {formatDate(item.start_date)} &ndash; {formatDate(item.end_date)}
                        </span>
                      </td>
                      <td style={{ textAlign: 'center' }}>
                        <span
                          style={{
                            display: 'inline-block',
                            padding: 'var(--spacing-1) var(--spacing-2)',
                            borderRadius: 'var(--radius-full)',
                            fontSize: 'var(--font-size-xs)',
                            fontWeight: 'var(--font-weight-medium)',
                            background: item.status === 'confirmed'
                              ? 'rgba(16, 185, 129, 0.15)'
                              : item.status === 'cancelled'
                              ? 'rgba(239, 68, 68, 0.15)'
                              : 'rgba(0, 212, 255, 0.1)',
                            color: item.status === 'confirmed'
                              ? 'var(--color-success)'
                              : item.status === 'cancelled'
                              ? 'var(--color-danger)'
                              : 'var(--color-primary)',
                          }}
                        >
                          {statusLabel(item.status)}
                        </span>
                      </td>
                      {filtered.some(i => i.unit_price) && (
                        <td style={{ textAlign: 'right' }}>
                          {item.total
                            ? `\u20AC${item.total.toFixed(2)}`
                            : item.unit_price
                            ? `\u20AC${item.unit_price.toFixed(2)}`
                            : '\u2014'}
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
                {subtotal > 0 && (
                  <tfoot>
                    <tr>
                      <td colSpan={2}>Zwischensumme</td>
                      <td />
                      <td style={{ textAlign: 'right' }}>{'\u20AC'}{subtotal.toFixed(2)}</td>
                    </tr>
                  </tfoot>
                )}
              </table>
            </div>
          )}
        </>
      )}
    </div>
  )
}
