import { useState, useCallback, useMemo } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { api, equipmentApi, reservationApi } from '../../../services/api'
import { getStatusLabel } from '../../../utils/statusLabels'
import { PackingListTab } from './PackingListTab'
import { SignaturePad } from '../../../components/SignaturePad/SignaturePad'
import { useNotificationStore } from '../../../stores/notificationStore'
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

interface EquipmentItem {
  id: string
  name: string
  sku?: string
  category_name?: string
  category_id?: string
  status?: string
  daily_rate?: number
  replacement_cost?: number
}

export function EquipmentTab({ project }: EquipmentTabProps) {
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [search, setSearch] = useState('')
  const [showSignaturePad, setShowSignaturePad] = useState(false)
  const [signatureDataUrl, setSignatureDataUrl] = useState<string | null>(null)
  const [showPackingList, setShowPackingList] = useState(false)
  const [showAddModal, setShowAddModal] = useState(false)
  const [addEquipmentId, setAddEquipmentId] = useState('')
  const [addQuantity, setAddQuantity] = useState(1)
  const [equipmentSearch, setEquipmentSearch] = useState('')
  const [isCheckingConflicts, setIsCheckingConflicts] = useState(false)
  const [conflicts, setConflicts] = useState<ConflictInfo[]>([])
  const [showConflictWarning, setShowConflictWarning] = useState(false)
  const [isAdding, setIsAdding] = useState(false)

  // Fetch all equipment for the picker
  const { data: allEquipmentRaw } = useQuery({
    queryKey: ['equipment-for-picker'],
    queryFn: () => equipmentApi.list({ limit: 200 }),
    enabled: showAddModal,
    staleTime: 1000 * 60 * 5,
  })

  const allEquipment: EquipmentItem[] = useMemo(() => {
    if (!allEquipmentRaw) return []
    const raw = allEquipmentRaw?.data || allEquipmentRaw?.items || allEquipmentRaw
    return Array.isArray(raw) ? raw : []
  }, [allEquipmentRaw])

  const filteredEquipment = useMemo(() => {
    if (!equipmentSearch.trim()) return allEquipment
    const q = equipmentSearch.toLowerCase()
    return allEquipment.filter((e) =>
      (e.name || '').toLowerCase().includes(q) ||
      (e.sku || '').toLowerCase().includes(q) ||
      (e.category_name || '').toLowerCase().includes(q)
    )
  }, [allEquipment, equipmentSearch])

  const selectedEquipment = useMemo(() => {
    return allEquipment.find((e) => e.id === addEquipmentId)
  }, [allEquipment, addEquipmentId])

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
    setAddQuantity(1)
    setEquipmentSearch('')
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
              minWidth: '500px',
              maxWidth: '650px',
              maxHeight: '80vh',
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
              border: '1px solid rgba(255,255,255,0.1)',
            }}
          >
            <h3 style={{ marginTop: 0, marginBottom: '1rem' }}>Equipment hinzufuegen</h3>

            {/* Search input */}
            <div style={{ marginBottom: '0.75rem' }}>
              <input
                type="text"
                value={equipmentSearch}
                onChange={(e) => setEquipmentSearch(e.target.value)}
                placeholder="Equipment suchen (Name, SKU, Kategorie)..."
                autoFocus
                style={{
                  width: '100%',
                  padding: '0.6rem 0.75rem',
                  background: 'rgba(255,255,255,0.05)',
                  border: '1px solid rgba(255,255,255,0.15)',
                  borderRadius: '6px',
                  color: 'inherit',
                  boxSizing: 'border-box',
                  fontSize: '0.9rem',
                }}
              />
            </div>

            {/* Equipment list */}
            <div style={{
              flex: 1,
              overflowY: 'auto',
              maxHeight: '300px',
              marginBottom: '0.75rem',
              border: '1px solid rgba(255,255,255,0.08)',
              borderRadius: '6px',
            }}>
              {filteredEquipment.length === 0 ? (
                <div style={{ padding: '1.5rem', textAlign: 'center', color: 'var(--color-text-muted)', fontSize: '0.85rem' }}>
                  {allEquipment.length === 0 ? 'Equipment wird geladen...' : 'Kein Equipment gefunden'}
                </div>
              ) : (
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.83rem' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.1)', position: 'sticky', top: 0, background: 'var(--color-bg-secondary, #1a1a2e)' }}>
                      <th style={{ padding: '0.5rem 0.5rem', textAlign: 'left', fontWeight: 600 }}>Name</th>
                      <th style={{ padding: '0.5rem 0.5rem', textAlign: 'left', fontWeight: 600 }}>SKU</th>
                      <th style={{ padding: '0.5rem 0.5rem', textAlign: 'left', fontWeight: 600 }}>Kategorie</th>
                      <th style={{ padding: '0.5rem 0.5rem', textAlign: 'center', fontWeight: 600 }}>Status</th>
                      <th style={{ padding: '0.5rem 0.5rem', textAlign: 'right', fontWeight: 600 }}>Tagespreis</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredEquipment.map((eq) => (
                      <tr
                        key={eq.id}
                        onClick={() => setAddEquipmentId(eq.id)}
                        style={{
                          cursor: 'pointer',
                          borderBottom: '1px solid rgba(255,255,255,0.04)',
                          background: addEquipmentId === eq.id ? 'rgba(0, 212, 255, 0.12)' : 'transparent',
                          transition: 'background 0.15s',
                        }}
                        onMouseEnter={(e) => { if (addEquipmentId !== eq.id) e.currentTarget.style.background = 'rgba(255,255,255,0.03)' }}
                        onMouseLeave={(e) => { if (addEquipmentId !== eq.id) e.currentTarget.style.background = 'transparent' }}
                      >
                        <td style={{ padding: '0.45rem 0.5rem', fontWeight: addEquipmentId === eq.id ? 600 : 400 }}>{eq.name}</td>
                        <td style={{ padding: '0.45rem 0.5rem', color: 'var(--color-text-muted)', fontFamily: 'monospace', fontSize: '0.78rem' }}>{eq.sku || '\u2014'}</td>
                        <td style={{ padding: '0.45rem 0.5rem', color: 'var(--color-text-secondary)' }}>{eq.category_name || '\u2014'}</td>
                        <td style={{ padding: '0.45rem 0.5rem', textAlign: 'center' }}>
                          <span style={{
                            display: 'inline-block',
                            padding: '2px 8px',
                            borderRadius: '999px',
                            fontSize: '0.75rem',
                            fontWeight: 500,
                            background: eq.status === 'available' ? 'rgba(16,185,129,0.15)' : eq.status === 'rented' ? 'rgba(245,158,11,0.15)' : 'rgba(0,212,255,0.1)',
                            color: eq.status === 'available' ? 'var(--color-success, #10b981)' : eq.status === 'rented' ? '#f59e0b' : 'var(--color-primary, #00d4ff)',
                          }}>
                            {getStatusLabel(eq.status || 'available')}
                          </span>
                        </td>
                        <td style={{ padding: '0.45rem 0.5rem', textAlign: 'right' }}>
                          {eq.daily_rate ? `\u20AC${eq.daily_rate.toFixed(2)}` : '\u2014'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>

            {/* Selected equipment info & quantity */}
            {selectedEquipment && (
              <div style={{
                marginBottom: '0.75rem',
                padding: '0.6rem 0.75rem',
                background: 'rgba(0, 212, 255, 0.08)',
                borderRadius: '6px',
                border: '1px solid rgba(0, 212, 255, 0.2)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: '1rem',
              }}>
                <div style={{ fontSize: '0.85rem' }}>
                  <strong>{selectedEquipment.name}</strong>
                  {selectedEquipment.sku && <span style={{ marginLeft: '0.5rem', color: 'var(--color-text-muted)', fontFamily: 'monospace', fontSize: '0.78rem' }}>{selectedEquipment.sku}</span>}
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexShrink: 0 }}>
                  <label style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>Menge:</label>
                  <input
                    type="number"
                    min="1"
                    value={addQuantity}
                    onChange={(e) => setAddQuantity(Math.max(1, parseInt(e.target.value) || 1))}
                    style={{
                      width: '60px',
                      padding: '0.3rem 0.4rem',
                      background: 'rgba(255,255,255,0.05)',
                      border: '1px solid rgba(255,255,255,0.15)',
                      borderRadius: '4px',
                      color: 'inherit',
                      textAlign: 'center',
                      fontSize: '0.85rem',
                    }}
                  />
                </div>
              </div>
            )}

            <div style={{ marginBottom: '0.75rem' }}>
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

      {/* Uebergabeprotokoll — Signature section for check-out */}
      {(project.status === 'confirmed' || project.status === 'in_progress') && (
        <div style={{ marginTop: 'var(--spacing-6, 1.5rem)' }}>
          <div className={styles.sectionHeader}>
            <h3 className={styles.sectionTitle}>Uebergabeprotokoll</h3>
          </div>
          {signatureDataUrl ? (
            <div style={{
              background: 'var(--color-bg-card, #111827)',
              borderRadius: 'var(--radius-lg, 12px)',
              padding: 'var(--spacing-4, 1rem)',
              border: '1px solid var(--color-border, #1e293b)',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)', marginBottom: 'var(--spacing-3, 0.75rem)' }}>
                <span style={{ color: 'var(--color-success, #10b981)', fontWeight: 600, fontSize: 'var(--font-size-sm)' }}>
                  Unterschrift erfasst
                </span>
              </div>
              <img
                src={signatureDataUrl}
                alt="Unterschrift"
                style={{
                  maxWidth: '300px',
                  borderRadius: '8px',
                  border: '1px solid var(--color-border, #1e293b)',
                }}
              />
              <div style={{ marginTop: 'var(--spacing-3, 0.75rem)' }}>
                <button
                  className="btn btn--secondary"
                  onClick={() => {
                    setSignatureDataUrl(null)
                    setShowSignaturePad(true)
                  }}
                  style={{ fontSize: 'var(--font-size-sm)' }}
                >
                  Unterschrift aendern
                </button>
              </div>
            </div>
          ) : showSignaturePad ? (
            <SignaturePad
              onSave={(dataUrl) => {
                setSignatureDataUrl(dataUrl)
                setShowSignaturePad(false)
                addNotification('Unterschrift wurde erfolgreich gespeichert.', 'success')
              }}
              onCancel={() => setShowSignaturePad(false)}
              title="Unterschrift des Empfaengers"
            />
          ) : (
            <div style={{
              background: 'var(--color-bg-card, #111827)',
              borderRadius: 'var(--radius-lg, 12px)',
              padding: 'var(--spacing-6, 1.5rem)',
              border: '1px solid var(--color-border, #1e293b)',
              textAlign: 'center',
            }}>
              <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-4, 1rem)', fontSize: 'var(--font-size-sm)' }}>
                Bei der Uebergabe kann hier eine Unterschrift erfasst werden.
              </p>
              <button
                className="btn btn--primary"
                onClick={() => setShowSignaturePad(true)}
              >
                Unterschrift erfassen
              </button>
            </div>
          )}
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

      {/* Übergabeprotokoll mit Unterschrift */}
      {(reservations as Reservation[]).length > 0 && (project.status === 'confirmed' || project.status === 'completed') && (
        <div style={{ marginTop: 'var(--spacing-6)' }}>
          {signatureDataUrl ? (
            <div className={styles.glassCard}>
              <div className={styles.glassCardTitle}>Übergabeprotokoll</div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                <img src={signatureDataUrl} alt="Unterschrift" style={{ maxWidth: '200px', borderRadius: '8px', border: '1px solid var(--color-border)' }} />
                <div>
                  <div style={{ color: 'var(--color-success)', fontWeight: 600 }}>Unterschrieben</div>
                  <button onClick={() => setSignatureDataUrl(null)} style={{ background: 'none', border: 'none', color: 'var(--color-text-muted)', cursor: 'pointer', fontSize: '12px', padding: 0 }}>
                    Unterschrift löschen
                  </button>
                </div>
              </div>
            </div>
          ) : (
            <div>
              {showSignaturePad ? (
                <SignaturePad
                  title="Übergabe-Unterschrift"
                  onSave={(dataUrl) => {
                    setSignatureDataUrl(dataUrl)
                    setShowSignaturePad(false)
                    addNotification('Unterschrift gespeichert', 'success', { duration: 3000 })
                  }}
                  onCancel={() => setShowSignaturePad(false)}
                />
              ) : (
                <button className="btn btn--secondary" onClick={() => setShowSignaturePad(true)}>
                  Übergabeprotokoll unterschreiben
                </button>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
