import { useState, useMemo, useCallback, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { warehouseApi, equipmentApi } from '../../services/api'
import './Inventory.scss'

// ============================================================================
// TYPES
// ============================================================================

interface Warehouse {
  id: string
  name: string
  zones?: Zone[]
}

interface Zone {
  id: string
  name: string
  type?: string
}

interface EquipmentItem {
  id: string
  name: string
  sku?: string
  barcode?: string
  category_id?: string
  status?: string
  location_id?: string
}

interface InventoryRow {
  equipmentId: string
  equipmentName: string
  barcode: string
  category: string
  expectedCount: number
  actualCount: number
  status: 'pending' | 'found' | 'missing' | 'surplus'
  notes: string
}

interface InventoryCheckResult {
  id: string
  name: string
  status: string
  items: Array<{
    equipment_id: string
    expected_count: number
    actual_count: number
    status: string
    notes?: string
    scanned_at?: string
  }>
  started_at?: string
  completed_at?: string
}

type Phase = 'setup' | 'counting' | 'result'

// ============================================================================
// CATEGORY MAP
// ============================================================================

const CATEGORY_LABELS: Record<string, string> = {
  'cat-audio': 'Audio',
  'cat-lighting': 'Lichttechnik',
  'cat-video': 'Video',
  'cat-stage': 'Buehne & Rigging',
}

function getCategoryLabel(catId: string): string {
  return CATEGORY_LABELS[catId] || catId || 'Sonstige'
}

function formatDateTime(iso?: string): string {
  if (!iso) return '-'
  return new Date(iso).toLocaleString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// ============================================================================
// COMPONENT
// ============================================================================

function InventoryPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const printRef = useRef<HTMLDivElement>(null)

  // Phase management
  const [phase, setPhase] = useState<Phase>('setup')

  // Setup phase state
  const [selectedWarehouseId, setSelectedWarehouseId] = useState<string>('')
  const [selectedZoneId, setSelectedZoneId] = useState<string>('')
  const [inventoryName, setInventoryName] = useState('')
  const [checkType, setCheckType] = useState<string>('zone')

  // Counting phase state
  const [activeCheckId, setActiveCheckId] = useState<string | null>(null)
  const [inventoryRows, setInventoryRows] = useState<InventoryRow[]>([])
  const [searchTerm, setSearchTerm] = useState('')

  // Result phase state
  const [completedCheck, setCompletedCheck] = useState<InventoryCheckResult | null>(null)

  // Fetch warehouses
  const { data: warehousesData } = useQuery({
    queryKey: ['warehouses'],
    queryFn: () => warehouseApi.listWarehouses(),
  })
  const warehouses: Warehouse[] = (warehousesData?.items || []).map(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (w: any) => ({ id: w.id, name: w.name, zones: w.zones || [] })
  )

  const selectedWarehouse = warehouses.find(w => w.id === selectedWarehouseId)
  const zones: Zone[] = selectedWarehouse?.zones || []

  // Fetch equipment
  const { data: equipmentData } = useQuery({
    queryKey: ['equipment-all'],
    queryFn: () => equipmentApi.list({ limit: 200, offset: 0 }),
    staleTime: 1000 * 60 * 5,
  })

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const allEquipment: EquipmentItem[] = useMemo(() => {
    const raw = equipmentData?.data || equipmentData?.items || equipmentData || []
    return Array.isArray(raw) ? raw : []
  }, [equipmentData])

  // Fetch past inventory checks
  const { data: pastChecksData } = useQuery({
    queryKey: ['inventory-checks'],
    queryFn: () => warehouseApi.listInventoryChecks(10, 0),
    staleTime: 1000 * 60 * 2,
  })
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const pastChecks = (pastChecksData?.data || []) as any[]

  // Start inventory check mutation
  const startCheckMutation = useMutation({
    mutationFn: () => warehouseApi.startInventoryCheck(
      inventoryName || `Inventur ${new Date().toLocaleDateString('de-DE')}`,
      checkType,
      selectedZoneId || undefined,
    ),
    onSuccess: (data) => {
      setActiveCheckId(data.id)
      // Build inventory rows from equipment list
      const rows: InventoryRow[] = allEquipment.map(eq => ({
        equipmentId: eq.id,
        equipmentName: eq.name,
        barcode: eq.barcode || '',
        category: getCategoryLabel(eq.category_id || ''),
        expectedCount: 1,
        actualCount: 0,
        status: 'pending' as const,
        notes: '',
      }))
      setInventoryRows(rows)
      setPhase('counting')
    },
  })

  // Scan item mutation
  const scanItemMutation = useMutation({
    mutationFn: ({ checkId, equipmentId, notes }: { checkId: string; equipmentId: string; notes?: string }) =>
      warehouseApi.scanInventoryItem(checkId, equipmentId, undefined, notes),
  })

  // Complete check mutation
  const completeCheckMutation = useMutation({
    mutationFn: (checkId: string) =>
      warehouseApi.completeInventoryCheck(checkId, 'current-user'),
    onSuccess: (data) => {
      // Build result from local state
      const result: InventoryCheckResult = {
        id: data.id || activeCheckId || '',
        name: inventoryName || 'Inventur',
        status: 'completed',
        items: inventoryRows.map(row => ({
          equipment_id: row.equipmentId,
          expected_count: row.expectedCount,
          actual_count: row.actualCount,
          status: row.status,
          notes: row.notes,
        })),
        started_at: data.started_at,
        completed_at: data.completed_at || new Date().toISOString(),
      }
      setCompletedCheck(result)
      setPhase('result')
      queryClient.invalidateQueries({ queryKey: ['inventory-checks'] })
    },
  })

  // Toggle item found
  const toggleItemFound = useCallback((equipmentId: string) => {
    setInventoryRows(prev => prev.map(row => {
      if (row.equipmentId !== equipmentId) return row
      const newActual = row.actualCount > 0 ? 0 : row.expectedCount
      const newStatus = newActual >= row.expectedCount ? 'found' : 'pending'
      // Fire scan API in background
      if (newActual > 0 && activeCheckId) {
        scanItemMutation.mutate({ checkId: activeCheckId, equipmentId })
      }
      return { ...row, actualCount: newActual, status: newStatus }
    }))
  }, [activeCheckId, scanItemMutation])

  // Set manual count
  const setManualCount = useCallback((equipmentId: string, count: number) => {
    setInventoryRows(prev => prev.map(row => {
      if (row.equipmentId !== equipmentId) return row
      let status: InventoryRow['status'] = 'pending'
      if (count > row.expectedCount) status = 'surplus'
      else if (count === row.expectedCount) status = 'found'
      else if (count > 0) status = 'found'
      return { ...row, actualCount: count, status }
    }))
  }, [])

  // Set notes
  const setItemNotes = useCallback((equipmentId: string, notes: string) => {
    setInventoryRows(prev => prev.map(row =>
      row.equipmentId === equipmentId ? { ...row, notes } : row
    ))
  }, [])

  // Filtered rows
  const filteredRows = useMemo(() => {
    if (!searchTerm) return inventoryRows
    const term = searchTerm.toLowerCase()
    return inventoryRows.filter(
      r => r.equipmentName.toLowerCase().includes(term) ||
           r.barcode.toLowerCase().includes(term) ||
           r.category.toLowerCase().includes(term)
    )
  }, [inventoryRows, searchTerm])

  // Summary stats
  const summary = useMemo(() => {
    const total = inventoryRows.length
    const found = inventoryRows.filter(r => r.status === 'found' || r.status === 'surplus').length
    const missing = inventoryRows.filter(r => r.actualCount === 0 && r.expectedCount > 0).length
    const surplus = inventoryRows.filter(r => r.status === 'surplus').length
    const pending = total - found - missing
    const accuracy = total > 0 ? Math.round((found / total) * 100) : 0
    return { total, found, missing, surplus, pending, accuracy }
  }, [inventoryRows])

  // Complete the inventory
  const handleComplete = () => {
    // Mark all unscanned as missing
    setInventoryRows(prev => prev.map(row => {
      if (row.actualCount === 0 && row.expectedCount > 0) {
        return { ...row, status: 'missing' as const }
      }
      return row
    }))
    if (activeCheckId) {
      completeCheckMutation.mutate(activeCheckId)
    } else {
      // No backend check — just show result locally
      const result: InventoryCheckResult = {
        id: 'local',
        name: inventoryName || 'Inventur',
        status: 'completed',
        items: inventoryRows.map(row => ({
          equipment_id: row.equipmentId,
          expected_count: row.expectedCount,
          actual_count: row.actualCount,
          status: row.actualCount === 0 ? 'missing' : row.status,
        })),
        completed_at: new Date().toISOString(),
      }
      setCompletedCheck(result)
      setPhase('result')
    }
  }

  // Print report
  const handlePrint = () => {
    window.print()
  }

  // Reset to new inventory
  const handleNewInventory = () => {
    setPhase('setup')
    setActiveCheckId(null)
    setInventoryRows([])
    setCompletedCheck(null)
    setInventoryName('')
    setSelectedZoneId('')
    setSearchTerm('')
  }

  // ============================================================================
  // RENDER: SETUP PHASE
  // ============================================================================

  if (phase === 'setup') {
    return (
      <div className="inv-page">
        <div className="inv-header">
          <div className="inv-header__left">
            <button className="inv-back-btn" onClick={() => navigate('/warehouse')}>
              &larr; Lager
            </button>
            <h1 className="inv-header__title">Inventur starten</h1>
            <p className="inv-header__subtitle">Soll/Ist-Abgleich durchfuehren</p>
          </div>
        </div>

        <div className="inv-setup">
          <div className="inv-setup__form">
            <div className="inv-field">
              <label className="inv-field__label">Bezeichnung</label>
              <input
                className="inv-field__input"
                type="text"
                placeholder={`Inventur ${new Date().toLocaleDateString('de-DE')}`}
                value={inventoryName}
                onChange={e => setInventoryName(e.target.value)}
              />
            </div>

            <div className="inv-field">
              <label className="inv-field__label">Lager</label>
              <select
                className="inv-field__select"
                value={selectedWarehouseId}
                onChange={e => {
                  setSelectedWarehouseId(e.target.value)
                  setSelectedZoneId('')
                }}
              >
                <option value="">Alle Lager</option>
                {warehouses.map(wh => (
                  <option key={wh.id} value={wh.id}>{wh.name}</option>
                ))}
              </select>
            </div>

            {zones.length > 0 && (
              <div className="inv-field">
                <label className="inv-field__label">Zone</label>
                <select
                  className="inv-field__select"
                  value={selectedZoneId}
                  onChange={e => setSelectedZoneId(e.target.value)}
                >
                  <option value="">Alle Zonen</option>
                  {zones.map(z => (
                    <option key={z.id} value={z.id}>{z.name}</option>
                  ))}
                </select>
              </div>
            )}

            <div className="inv-field">
              <label className="inv-field__label">Art der Inventur</label>
              <select
                className="inv-field__select"
                value={checkType}
                onChange={e => setCheckType(e.target.value)}
              >
                <option value="full">Vollinventur</option>
                <option value="zone">Zone</option>
                <option value="cycle">Stichprobe (Cycle Count)</option>
                <option value="spot_check">Spot Check</option>
              </select>
            </div>

            <button
              className="inv-btn inv-btn--primary inv-btn--lg"
              onClick={() => startCheckMutation.mutate()}
              disabled={startCheckMutation.isPending}
            >
              {startCheckMutation.isPending ? 'Wird gestartet...' : 'Inventur starten'}
            </button>

            {startCheckMutation.isError && (
              <div className="inv-error">
                Fehler beim Starten: {(startCheckMutation.error as Error)?.message || 'Unbekannter Fehler'}
              </div>
            )}
          </div>

          {/* Past inventory checks */}
          {pastChecks.length > 0 && (
            <div className="inv-setup__history">
              <h3 className="inv-setup__history-title">Letzte Inventuren</h3>
              <div className="inv-history-list">
                {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                {pastChecks.map((check: any) => (
                  <div key={check.id} className="inv-history-item">
                    <div className="inv-history-item__name">{check.name}</div>
                    <div className="inv-history-item__meta">
                      <span className={`inv-history-item__status inv-history-item__status--${check.status}`}>
                        {check.status === 'completed' ? 'Abgeschlossen' :
                         check.status === 'in_progress' ? 'Laufend' : 'Geplant'}
                      </span>
                      <span className="inv-history-item__date">{formatDateTime(check.created_at)}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    )
  }

  // ============================================================================
  // RENDER: COUNTING PHASE
  // ============================================================================

  if (phase === 'counting') {
    return (
      <div className="inv-page">
        <div className="inv-header">
          <div className="inv-header__left">
            <button className="inv-back-btn" onClick={() => {
              if (confirm('Inventur wirklich abbrechen? Nicht gespeicherte Daten gehen verloren.')) {
                handleNewInventory()
              }
            }}>
              &larr; Abbrechen
            </button>
            <h1 className="inv-header__title">Inventur laeuft</h1>
            <p className="inv-header__subtitle">
              {inventoryName || 'Inventur'} &mdash;{' '}
              {selectedWarehouse?.name || 'Alle Lager'}
              {selectedZoneId && zones.find(z => z.id === selectedZoneId)
                ? ` / ${zones.find(z => z.id === selectedZoneId)!.name}`
                : ''}
            </p>
          </div>
          <div className="inv-header__actions">
            <button
              className="inv-btn inv-btn--success inv-btn--lg"
              onClick={handleComplete}
              disabled={completeCheckMutation.isPending}
            >
              {completeCheckMutation.isPending ? 'Wird abgeschlossen...' : 'Inventur abschliessen'}
            </button>
          </div>
        </div>

        {/* Summary bar */}
        <div className="inv-summary-bar">
          <div className="inv-summary-stat">
            <span className="inv-summary-stat__value">{summary.total}</span>
            <span className="inv-summary-stat__label">Gesamt</span>
          </div>
          <div className="inv-summary-stat inv-summary-stat--found">
            <span className="inv-summary-stat__value">{summary.found}</span>
            <span className="inv-summary-stat__label">Gefunden</span>
          </div>
          <div className="inv-summary-stat inv-summary-stat--pending">
            <span className="inv-summary-stat__value">{summary.pending}</span>
            <span className="inv-summary-stat__label">Offen</span>
          </div>
          <div className="inv-summary-stat inv-summary-stat--surplus">
            <span className="inv-summary-stat__value">{summary.surplus}</span>
            <span className="inv-summary-stat__label">Ueberschuss</span>
          </div>
          <div className="inv-summary-stat">
            <span className="inv-summary-stat__value">{summary.accuracy}%</span>
            <span className="inv-summary-stat__label">Erfasst</span>
          </div>
        </div>

        {/* Progress bar */}
        <div className="inv-progress">
          <div
            className="inv-progress__bar"
            style={{ width: `${summary.total > 0 ? ((summary.found + summary.surplus) / summary.total) * 100 : 0}%` }}
          />
        </div>

        {/* Search */}
        <div className="inv-search">
          <input
            className="inv-search__input"
            type="text"
            placeholder="Equipment suchen (Name, Barcode, Kategorie)..."
            value={searchTerm}
            onChange={e => setSearchTerm(e.target.value)}
            autoFocus
          />
          {searchTerm && (
            <button className="inv-search__clear" onClick={() => setSearchTerm('')}>
              &times;
            </button>
          )}
        </div>

        {/* Equipment table */}
        <div className="inv-table-wrapper">
          <table className="inv-table">
            <thead>
              <tr>
                <th className="inv-table__th inv-table__th--check">Gefunden</th>
                <th className="inv-table__th">Equipment</th>
                <th className="inv-table__th">Barcode</th>
                <th className="inv-table__th">Kategorie</th>
                <th className="inv-table__th inv-table__th--num">Soll</th>
                <th className="inv-table__th inv-table__th--num">Ist</th>
                <th className="inv-table__th">Notizen</th>
              </tr>
            </thead>
            <tbody>
              {filteredRows.map(row => (
                <tr
                  key={row.equipmentId}
                  className={`inv-table__row inv-table__row--${row.status}`}
                >
                  <td className="inv-table__td inv-table__td--check">
                    <label className="inv-checkbox">
                      <input
                        type="checkbox"
                        checked={row.actualCount >= row.expectedCount}
                        onChange={() => toggleItemFound(row.equipmentId)}
                      />
                      <span className="inv-checkbox__mark" />
                    </label>
                  </td>
                  <td className="inv-table__td inv-table__td--name">
                    {row.equipmentName}
                  </td>
                  <td className="inv-table__td inv-table__td--barcode">
                    <code>{row.barcode}</code>
                  </td>
                  <td className="inv-table__td">{row.category}</td>
                  <td className="inv-table__td inv-table__td--num">{row.expectedCount}</td>
                  <td className="inv-table__td inv-table__td--num">
                    <input
                      className="inv-count-input"
                      type="number"
                      min={0}
                      value={row.actualCount}
                      onChange={e => setManualCount(row.equipmentId, Math.max(0, parseInt(e.target.value) || 0))}
                    />
                  </td>
                  <td className="inv-table__td">
                    <input
                      className="inv-notes-input"
                      type="text"
                      placeholder="Notiz..."
                      value={row.notes}
                      onChange={e => setItemNotes(row.equipmentId, e.target.value)}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {filteredRows.length === 0 && (
            <div className="inv-table__empty">
              {searchTerm ? 'Keine Treffer fuer die Suche.' : 'Kein Equipment geladen.'}
            </div>
          )}
        </div>
      </div>
    )
  }

  // ============================================================================
  // RENDER: RESULT PHASE
  // ============================================================================

  const resultRows = inventoryRows.map(row => ({
    ...row,
    variance: row.actualCount - row.expectedCount,
    status: row.actualCount === 0 && row.expectedCount > 0
      ? 'missing' as const
      : row.actualCount > row.expectedCount
        ? 'surplus' as const
        : row.actualCount === row.expectedCount
          ? 'found' as const
          : 'pending' as const,
  }))

  const missingItems = resultRows.filter(r => r.status === 'missing')
  const surplusItems = resultRows.filter(r => r.status === 'surplus')
  const foundItems = resultRows.filter(r => r.status === 'found')
  const partialItems = resultRows.filter(r => r.status === 'pending' && r.actualCount > 0)

  return (
    <div className="inv-page">
      <div className="inv-header inv-header--print" ref={printRef}>
        <div className="inv-header__left">
          <button className="inv-back-btn inv-no-print" onClick={() => navigate('/warehouse')}>
            &larr; Lager
          </button>
          <h1 className="inv-header__title">Inventur-Ergebnis</h1>
          <p className="inv-header__subtitle">
            {completedCheck?.name || inventoryName} &mdash; Abgeschlossen {formatDateTime(completedCheck?.completed_at)}
          </p>
        </div>
        <div className="inv-header__actions inv-no-print">
          <button className="inv-btn inv-btn--secondary" onClick={handlePrint}>
            Drucken / PDF
          </button>
          <button className="inv-btn inv-btn--primary" onClick={handleNewInventory}>
            Neue Inventur
          </button>
        </div>
      </div>

      {/* Result summary */}
      <div className="inv-result-summary">
        <div className="inv-result-card inv-result-card--total">
          <div className="inv-result-card__value">{summary.total}</div>
          <div className="inv-result-card__label">Positionen gesamt</div>
        </div>
        <div className="inv-result-card inv-result-card--found">
          <div className="inv-result-card__value">{foundItems.length}</div>
          <div className="inv-result-card__label">Korrekt</div>
        </div>
        <div className="inv-result-card inv-result-card--missing">
          <div className="inv-result-card__value">{missingItems.length}</div>
          <div className="inv-result-card__label">Fehlend</div>
        </div>
        <div className="inv-result-card inv-result-card--surplus">
          <div className="inv-result-card__value">{surplusItems.length}</div>
          <div className="inv-result-card__label">Ueberschuss</div>
        </div>
        {partialItems.length > 0 && (
          <div className="inv-result-card inv-result-card--partial">
            <div className="inv-result-card__value">{partialItems.length}</div>
            <div className="inv-result-card__label">Teilweise</div>
          </div>
        )}
        <div className="inv-result-card">
          <div className="inv-result-card__value">{summary.accuracy}%</div>
          <div className="inv-result-card__label">Genauigkeit</div>
        </div>
      </div>

      {/* Discrepancies */}
      {(missingItems.length > 0 || surplusItems.length > 0 || partialItems.length > 0) && (
        <div className="inv-discrepancies">
          <h2 className="inv-section-title">Differenzen</h2>
          <table className="inv-table inv-table--result">
            <thead>
              <tr>
                <th className="inv-table__th">Equipment</th>
                <th className="inv-table__th">Barcode</th>
                <th className="inv-table__th">Kategorie</th>
                <th className="inv-table__th inv-table__th--num">Soll</th>
                <th className="inv-table__th inv-table__th--num">Ist</th>
                <th className="inv-table__th inv-table__th--num">Differenz</th>
                <th className="inv-table__th">Status</th>
                <th className="inv-table__th">Notizen</th>
              </tr>
            </thead>
            <tbody>
              {[...missingItems, ...surplusItems, ...partialItems].map(row => (
                <tr key={row.equipmentId} className={`inv-table__row inv-table__row--${row.status}`}>
                  <td className="inv-table__td inv-table__td--name">{row.equipmentName}</td>
                  <td className="inv-table__td inv-table__td--barcode"><code>{row.barcode}</code></td>
                  <td className="inv-table__td">{row.category}</td>
                  <td className="inv-table__td inv-table__td--num">{row.expectedCount}</td>
                  <td className="inv-table__td inv-table__td--num">{row.actualCount}</td>
                  <td className={`inv-table__td inv-table__td--num ${row.variance < 0 ? 'inv-text--danger' : row.variance > 0 ? 'inv-text--warning' : ''}`}>
                    {row.variance > 0 ? `+${row.variance}` : row.variance}
                  </td>
                  <td className="inv-table__td">
                    <span className={`inv-status-badge inv-status-badge--${row.status}`}>
                      {row.status === 'missing' ? 'Fehlend' :
                       row.status === 'surplus' ? 'Ueberschuss' :
                       row.status === 'found' ? 'OK' : 'Teilweise'}
                    </span>
                  </td>
                  <td className="inv-table__td">{row.notes}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Full list */}
      <div className="inv-full-list">
        <h2 className="inv-section-title">Komplette Liste ({resultRows.length} Positionen)</h2>
        <table className="inv-table inv-table--result">
          <thead>
            <tr>
              <th className="inv-table__th">Equipment</th>
              <th className="inv-table__th">Barcode</th>
              <th className="inv-table__th inv-table__th--num">Soll</th>
              <th className="inv-table__th inv-table__th--num">Ist</th>
              <th className="inv-table__th">Status</th>
            </tr>
          </thead>
          <tbody>
            {resultRows.map(row => (
              <tr key={row.equipmentId} className={`inv-table__row inv-table__row--${row.status}`}>
                <td className="inv-table__td inv-table__td--name">{row.equipmentName}</td>
                <td className="inv-table__td inv-table__td--barcode"><code>{row.barcode}</code></td>
                <td className="inv-table__td inv-table__td--num">{row.expectedCount}</td>
                <td className="inv-table__td inv-table__td--num">{row.actualCount}</td>
                <td className="inv-table__td">
                  <span className={`inv-status-badge inv-status-badge--${row.status}`}>
                    {row.status === 'missing' ? 'Fehlend' :
                     row.status === 'surplus' ? 'Ueberschuss' :
                     row.status === 'found' ? 'OK' : 'Offen'}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default InventoryPage
