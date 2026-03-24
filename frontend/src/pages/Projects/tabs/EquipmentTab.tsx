import { useState, useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { api, reservationApi } from '../../../services/api'
import { getStatusLabel } from '../../../utils/statusLabels'
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

interface ConflictInfo {
  reservation_id: string
  project_id: string
  project_name?: string
  equipment_id: string
  start_date: string
  end_date: string
  status: string
}

export function EquipmentTab({ project }: EquipmentTabProps) {
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [showPackingList, setShowPackingList] = useState(false)
  const [showAddModal, setShowAddModal] = useState(false)
  const [addEquipmentId, setAddEquipmentId] = useState('')
  const [isCheckingConflicts, setIsCheckingConflicts] = useState(false)
  const [conflicts, setConflicts] = useState<ConflictInfo[]>([])
  const [showConflictWarning, setShowConflictWarning] = useState(false)
  const [isAdding, setIsAdding] = useState(false)

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
      confirmed: 'Bestaetigt',
      pending: 'Ausstehend',
      cancelled: 'Storniert',
      checked_out: 'Ausgegeben',
      returned: 'Zurueckgegeben',
    }
    return map[s] || s
  }

  // Check for double-booking conflicts before adding equipment
  const handleCheckAndAdd = useCallback(async () => {
    if (!addEquipmentId) return

    setIsCheckingConflicts(true)
    setConflicts([])
    setShowConflictWarning(false)

    try {
      const conflictResults = await reservationApi.checkConflicts(
        addEquipmentId,
        project.start_date,
        project.end_date,
        project.id, // exclude current project
      )

      if (conflictResults && conflictResults.length > 0) {
        setConflicts(conflictResults)
        setShowConflictWarning(true)
        setIsCheckingConflicts(false)
        return // Don't add yet, show warning first
      }

      // No conflicts, proceed with adding
      await addEquipmentToProject()
    } catch {
      // If conflict check fails, proceed anyway (don't block the user)
      await addEquipmentToProject()
    } finally {
      setIsCheckingConflicts(false)
    }
  }, [addEquipmentId, project.start_date, project.end_date, project.id])

  const addEquipmentToProject = useCallback(async (force = false) => {
    setIsAdding(true)
    try {
      await reservationApi.create({
        project_id: project.id,
        equipment_id: addEquipmentId,
        start_date: project.start_date,
        end_date: project.end_date,
        ...(force ? { force: true } : {}),
      })
      queryClient.invalidateQueries({ queryKey: ['project-reservations', project.id] })
      // Reset modal state
      setShowAddModal(false)
      setAddEquipmentId('')

      setConflicts([])
      setShowConflictWarning(false)
    } catch {
      // silent - user sees no change
    } finally {
      setIsAdding(false)
    }
  }, [addEquipmentId, project.id, project.start_date, project.end_date, queryClient])

  const handleForceAdd = useCallback(async () => {
    // User confirmed they want to add despite conflicts
    setShowConflictWarning(false)
    await addEquipmentToProject(true)
  }, [addEquipmentToProject])

  const handleCloseModal = () => {
    setShowAddModal(false)
    setAddEquipmentId('')
    setAddEquipmentName('')
    setConflicts([])
    setShowConflictWarning(false)
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
            <button
              className="btn btn--primary"
              onClick={() => setShowAddModal(true)}
            >
              + Equipment hinzufuegen
            </button>
          )}
        </div>
      </div>

      {/* Add Equipment Modal with double-booking detection */}
      {showAddModal && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0,0,0,0.6)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
          }}
          onClick={(e) => { if (e.target === e.currentTarget) handleCloseModal() }}
        >
          <div
            style={{
              background: 'var(--color-bg-secondary, #1a1a2e)',
              borderRadius: 'var(--radius-lg, 12px)',
              padding: 'var(--spacing-6, 1.5rem)',
              minWidth: '400px',
              maxWidth: '500px',
              border: '1px solid rgba(255,255,255,0.1)',
            }}
          >
            <h3 style={{ marginTop: 0, marginBottom: '1rem' }}>Equipment hinzufuegen</h3>

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>
                Equipment-ID
              </label>
              <input
                type="text"
                value={addEquipmentId}
                onChange={(e) => setAddEquipmentId(e.target.value)}
                placeholder="z.B. eq-123 oder Equipment-ID eingeben"
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  background: 'rgba(255,255,255,0.05)',
                  border: '1px solid rgba(255,255,255,0.15)',
                  borderRadius: '6px',
                  color: 'inherit',
                  boxSizing: 'border-box',
                }}
              />
            </div>

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.25rem', fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>
                Zeitraum
              </label>
              <div style={{ fontSize: '0.9rem', color: 'var(--color-text-muted)' }}>
                {formatDate(project.start_date)} &ndash; {formatDate(project.end_date)}
              </div>
            </div>

            {/* Conflict Warning */}
            {showConflictWarning && conflicts.length > 0 && (
              <div
                style={{
                  background: 'rgba(245, 158, 11, 0.15)',
                  border: '1px solid rgba(245, 158, 11, 0.3)',
                  borderRadius: '8px',
                  padding: '0.75rem',
                  marginBottom: '1rem',
                }}
              >
                <div style={{ fontWeight: 600, marginBottom: '0.5rem', color: '#f59e0b' }}>
                  {'\u26A0\uFE0F'} Doppelbuchung erkannt!
                </div>
                <div style={{ fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>
                  Dieses Equipment ist bereits fuer folgende Projekte reserviert:
                </div>
                <ul style={{ margin: '0.5rem 0', paddingLeft: '1.25rem', fontSize: '0.85rem' }}>
                  {conflicts.map((c) => (
                    <li key={c.reservation_id}>
                      <strong>{c.project_name || c.project_id}</strong>
                      {' '}({formatDate(c.start_date)} &ndash; {formatDate(c.end_date)})
                      {' '}<span style={{ color: 'var(--color-text-muted)' }}>- {getStatusLabel(c.status)}</span>
                    </li>
                  ))}
                </ul>
                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.75rem' }}>
                  <button
                    className="btn btn--primary"
                    onClick={handleForceAdd}
                    disabled={isAdding}
                    style={{ fontSize: '0.8rem', padding: '0.35rem 0.75rem' }}
                  >
                    {isAdding ? 'Wird hinzugefuegt...' : 'Trotzdem hinzufuegen'}
                  </button>
                  <button
                    className="btn btn--secondary"
                    onClick={handleCloseModal}
                    style={{ fontSize: '0.8rem', padding: '0.35rem 0.75rem' }}
                  >
                    Abbrechen
                  </button>
                </div>
              </div>
            )}

            {/* Action buttons (hidden when conflict warning is shown) */}
            {!showConflictWarning && (
              <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
                <button
                  className="btn btn--secondary"
                  onClick={handleCloseModal}
                >
                  Abbrechen
                </button>
                <button
                  className="btn btn--primary"
                  onClick={handleCheckAndAdd}
                  disabled={!addEquipmentId || isCheckingConflicts || isAdding}
                >
                  {isCheckingConflicts ? 'Pruefe Verfuegbarkeit...' : isAdding ? 'Wird hinzugefuegt...' : 'Hinzufuegen'}
                </button>
              </div>
            )}
          </div>
        </div>
      )}

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
