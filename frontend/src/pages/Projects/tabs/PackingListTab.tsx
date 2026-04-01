import { useState, useCallback, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { projectApi, packlistApi } from '../../../services/api'
import './PackingListTab.scss'

// ============================================================================
// Types
// ============================================================================

export type PackStatus = 'ausstehend' | 'gepackt' | 'fehlt' | 'ersetzt'

export interface PackingItem {
  id: string
  equipment_id: string
  name: string
  category: string
  quantity: number
  quantity_packed: number
  serial_number?: string
  weight?: number
  barcode?: string
  status: PackStatus
  storage_location?: string
  packlist_id?: string
  replaced_by?: string
}

interface PackingListTabProps {
  project: Project
  /** Optional: called when an item status changes (for API sync) */
  onItemStatusChange?: (itemId: string, status: PackStatus) => void
}

// Map backend status to our frontend PackStatus
function mapBackendStatus(backendStatus: string): PackStatus {
  switch (backendStatus) {
    case 'packed':
      return 'gepackt'
    case 'missing':
      return 'fehlt'
    case 'damaged':
      return 'ersetzt'
    default:
      return 'ausstehend'
  }
}


const STATUS_CONFIG: Record<PackStatus, { icon: string; label: string; color: string; bgColor: string }> = {
  ausstehend: { icon: '\u2B1C', label: 'Ausstehend', color: 'var(--color-text-muted)', bgColor: 'rgba(107, 114, 128, 0.15)' },
  gepackt: { icon: '\u2705', label: 'Gepackt', color: 'var(--color-success)', bgColor: 'rgba(16, 185, 129, 0.15)' },
  fehlt: { icon: '\u274C', label: 'Fehlt', color: 'var(--color-danger)', bgColor: 'rgba(239, 68, 68, 0.15)' },
  ersetzt: { icon: '\u26A0\uFE0F', label: 'Ersetzt', color: 'var(--color-warning)', bgColor: 'rgba(245, 158, 11, 0.15)' },
}

// ============================================================================
// Component
// ============================================================================

export function PackingListTab({ project, onItemStatusChange }: PackingListTabProps) {
  const [localStatusOverrides, setLocalStatusOverrides] = useState<Record<string, PackStatus>>({})
  const [filter, setFilter] = useState<'alle' | PackStatus>('alle')
  const [animatingIds, setAnimatingIds] = useState<Set<string>>(new Set())
  const [isPrintMode, setIsPrintMode] = useState(false)

  // Fetch packing list data from JSON endpoint (grouped by location)
  const { data: packingListData, isLoading: isLoadingJSON } = useQuery({
    queryKey: ['project-packing-list-json', project.id],
    queryFn: () => projectApi.getPackingListJSON(project.id),
  })

  // Also fetch raw packlists for item-level operations (pack/return)
  const { data: packlistsRaw, isLoading: isLoadingPacklists } = useQuery({
    queryKey: ['project-packlists', project.id],
    queryFn: () => packlistApi.list(project.id),
  })

  const isLoading = isLoadingJSON || isLoadingPacklists

  // Extract packlists array
  const packlists = useMemo(() => {
    if (!packlistsRaw) return []
    const raw = packlistsRaw
    const arr = Array.isArray(raw) ? raw : (raw?.data ?? raw?.items ?? [])
    return arr
  }, [packlistsRaw])

  // Build flat list of PackingItems from the packlists data
  const items: PackingItem[] = useMemo(() => {
    const result: PackingItem[] = []
    for (const pl of packlists) {
      const plItems = pl.items || []
      for (const item of plItems) {
        const itemId = item.id || `${pl.id}-${item.equipment_id}`
        result.push({
          id: itemId,
          equipment_id: item.equipment_id,
          name: item.equipment_name || item.equipment_id,
          category: item.storage_location || 'Nicht zugeordnet',
          quantity: item.quantity || 1,
          quantity_packed: item.quantity_packed || 0,
          barcode: item.barcode,
          status: localStatusOverrides[itemId] ?? mapBackendStatus(item.status),
          storage_location: item.storage_location,
          packlist_id: pl.id,
        })
      }
    }
    return result
  }, [packlists, localStatusOverrides])

  // If no packlist items, fall back to the JSON packing list locations data
  const locationGroups: Array<{ location: string; items: PackingItem[] }> = useMemo(() => {
    if (items.length > 0) {
      // Group items by storage_location
      const groups: Record<string, PackingItem[]> = {}
      for (const item of items) {
        const loc = item.storage_location || item.category || 'Nicht zugeordnet'
        if (!groups[loc]) groups[loc] = []
        groups[loc].push(item)
      }
      return Object.entries(groups)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([location, locationItems]) => ({ location, items: locationItems }))
    }

    // Fall back to JSON packing list data
    if (packingListData?.locations) {
      return packingListData.locations.map((loc: any) => ({
        location: loc.location,
        items: (loc.items || []).map((item: any, idx: number) => ({
          id: `json-${loc.location}-${idx}`,
          equipment_id: item.sku || `eq-${idx}`,
          name: item.name,
          category: loc.location,
          quantity: item.quantity || 1,
          quantity_packed: item.quantity_packed || 0,
          barcode: item.barcode,
          status: (localStatusOverrides[`json-${loc.location}-${idx}`] ?? mapBackendStatus(item.status || 'pending')) as PackStatus,
          storage_location: loc.location,
        })),
      }))
    }

    return []
  }, [items, packingListData, localStatusOverrides])

  // Flat items list for stats
  const allItems = useMemo(() => locationGroups.flatMap(g => g.items), [locationGroups])
  const filteredLocationGroups = useMemo(() => {
    if (filter === 'alle') return locationGroups
    return locationGroups
      .map(g => ({ ...g, items: g.items.filter(i => i.status === filter) }))
      .filter(g => g.items.length > 0)
  }, [locationGroups, filter])

  // ---- Computed values ----
  const totalItems = allItems.length
  const packedCount = allItems.filter((i) => i.status === 'gepackt').length
  const fehltCount = allItems.filter((i) => i.status === 'fehlt').length
  const ersetztCount = allItems.filter((i) => i.status === 'ersetzt').length
  const progressPercent = totalItems > 0 ? Math.round((packedCount / totalItems) * 100) : 0

  // ---- Handlers ----
  const toggleItemStatus = useCallback((itemId: string, newStatus: PackStatus) => {
    setLocalStatusOverrides((prev) => {
      const current = prev[itemId]
      const finalStatus = current === newStatus ? 'ausstehend' : newStatus
      onItemStatusChange?.(itemId, finalStatus)

      // Try to sync with API
      const item = allItems.find(i => i.id === itemId)
      if (item?.packlist_id && item.equipment_id) {
        if (finalStatus === 'gepackt') {
          packlistApi.markItemPacked(item.packlist_id, {
            equipment_id: item.equipment_id,
            quantity_packed: item.quantity,
          }).catch(() => { /* silent */ })
        }
      }

      return { ...prev, [itemId]: finalStatus }
    })
    // Trigger animation
    setAnimatingIds((prev) => new Set(prev).add(itemId))
    setTimeout(() => {
      setAnimatingIds((prev) => {
        const next = new Set(prev)
        next.delete(itemId)
        return next
      })
    }, 600)
  }, [onItemStatusChange, allItems])

  const markAllPacked = useCallback(() => {
    const updates: Record<string, PackStatus> = {}
    for (const item of allItems) {
      if (item.status === 'ausstehend') {
        updates[item.id] = 'gepackt'
        onItemStatusChange?.(item.id, 'gepackt')
        if (item.packlist_id && item.equipment_id) {
          packlistApi.markItemPacked(item.packlist_id, {
            equipment_id: item.equipment_id,
            quantity_packed: item.quantity,
          }).catch(() => { /* silent */ })
        }
      }
    }
    setLocalStatusOverrides(prev => ({ ...prev, ...updates }))
    setAnimatingIds(new Set(Object.keys(updates)))
    setTimeout(() => setAnimatingIds(new Set()), 600)
  }, [allItems, onItemStatusChange])

  const resetAll = useCallback(() => {
    const updates: Record<string, PackStatus> = {}
    for (const item of allItems) {
      updates[item.id] = 'ausstehend'
      onItemStatusChange?.(item.id, 'ausstehend')
    }
    setLocalStatusOverrides(prev => ({ ...prev, ...updates }))
  }, [allItems, onItemStatusChange])

  /** Called from scanner: verify a barcode against this packing list */
  const verifyBarcode = useCallback((barcode: string): { found: boolean; item?: PackingItem } => {
    const item = allItems.find((i) => i.barcode === barcode)
    if (!item) return { found: false }
    if (item.status !== 'gepackt') {
      toggleItemStatus(item.id, 'gepackt')
    }
    return { found: true, item }
  }, [allItems, toggleItemStatus])

  // Expose verifyBarcode via window for scanner integration
  if (typeof window !== 'undefined') {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).__packingListVerify = verifyBarcode;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).__packingListProjectId = project.id
  }

  const handlePrint = () => {
    setIsPrintMode(true)
    setTimeout(() => {
      window.print()
      setIsPrintMode(false)
    }, 100)
  }

  // ---- Render: Loading ----
  if (isLoading) {
    return (
      <div className="packingList">
        <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--color-text-muted)' }}>
          Packliste wird geladen...
        </div>
      </div>
    )
  }

  // ---- Render: Print mode ----
  if (isPrintMode) {
    return (
      <div className="printView">
        <div className="printHeader">
          <h1>Packliste</h1>
          <div className="printMeta">
            <p><strong>Projekt:</strong> {project.name}</p>
            <p><strong>Datum:</strong> {new Date().toLocaleDateString('de-DE')}</p>
            <p><strong>Zeitraum:</strong> {new Date(project.start_date).toLocaleDateString('de-DE')} &ndash; {new Date(project.end_date).toLocaleDateString('de-DE')}</p>
            {project.client_name && <p><strong>Kunde:</strong> {project.client_name}</p>}
          </div>
          <div className="printProgress">
            {packedCount} von {totalItems} Positionen gepackt ({progressPercent}%)
          </div>
        </div>
        {filteredLocationGroups.map(({ location, items: locItems }) => (
          <div key={location} className="printCategory">
            <h2>{location}</h2>
            <table className="printTable">
              <thead>
                <tr>
                  <th style={{ width: '30px' }}></th>
                  <th>Artikel</th>
                  <th>Menge</th>
                  <th>Gepackt</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {locItems.map((item) => (
                  <tr key={item.id}>
                    <td className="printCheckbox">
                      {item.status === 'gepackt' ? '\u2611' : '\u2610'}
                    </td>
                    <td>{item.name}</td>
                    <td>{item.quantity}x</td>
                    <td>{item.quantity_packed}/{item.quantity}</td>
                    <td>{STATUS_CONFIG[item.status].label}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ))}
        <div className="printFooter">
          <p>Gepackt von: _________________________ &nbsp;&nbsp; Datum: _________________________</p>
          <p>Kontrolliert von: _________________________ &nbsp;&nbsp; Unterschrift: _________________________</p>
        </div>
      </div>
    )
  }

  return (
    <div className="packingList">
      {/* Header with progress */}
      <div className="header">
        <div className="headerLeft">
          <h3 className="title">Packliste</h3>
          <span className="projectLabel">{project.name}</span>
        </div>
        <div className="headerActions">
          <button className="btn btn--secondary actionBtn" onClick={handlePrint}>
            Drucken
          </button>
          <button className="btn btn--secondary actionBtn" onClick={resetAll}>
            Zuruecksetzen
          </button>
          <button
            className="btn btn--primary actionBtn"
            onClick={markAllPacked}
            disabled={packedCount === totalItems}
          >
            Alles als gepackt markieren
          </button>
        </div>
      </div>

      {/* Progress bar */}
      <div className="progressSection">
        <div className="progressInfo">
          <span className="progressText">
            {packedCount} von {totalItems} Positionen gepackt ({progressPercent}%)
          </span>
          <div className="progressStats">
            {fehltCount > 0 && (
              <span className="statBadgeDanger">{fehltCount} fehlt</span>
            )}
            {ersetztCount > 0 && (
              <span className="statBadgeWarning">{ersetztCount} ersetzt</span>
            )}
            <span className="statBadgeInfo">
              {totalItems} Positionen gesamt
            </span>
          </div>
        </div>
        <div className="progressBar">
          <div
            className="progressFill"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
      </div>

      {/* Filter bar */}
      <div className="filterBar">
        {(['alle', 'ausstehend', 'gepackt', 'fehlt', 'ersetzt'] as const).map((f) => (
          <button
            key={f}
            className={`filterBtn ${filter === f ? 'filterBtnActive' : ''}`}
            onClick={() => setFilter(f)}
          >
            {f === 'alle' ? 'Alle' : STATUS_CONFIG[f].icon + ' ' + STATUS_CONFIG[f].label}
            <span className="filterCount">
              {f === 'alle'
                ? totalItems
                : allItems.filter((i) => i.status === f).length}
            </span>
          </button>
        ))}
      </div>

      {/* Empty state */}
      {totalItems === 0 && (
        <div className="emptyState">
          <div className="emptyStateIcon">{'\u{1F4E6}'}</div>
          <h4 className="emptyStateTitle">Keine Artikel in der Packliste</h4>
          <p className="emptyStateText">
            Erstellen Sie zuerst eine Packliste und fuegen Sie Equipment hinzu, um die Packliste zu verwenden.
          </p>
        </div>
      )}

      {/* Location groups (sorted by warehouse location) */}
      {filteredLocationGroups.map(({ location, items: locItems }) => {
        const locPacked = locItems.filter((i) => i.status === 'gepackt').length
        return (
          <div key={location} className="categoryGroup">
            <div className="categoryHeader">
              <span className="categoryIcon">{'\u{1F4CD}'}</span>
              <span className="categoryLabel">{location}</span>
              <span className="categoryCount">
                {locPacked}/{locItems.length} gepackt
              </span>
            </div>
            <div className="itemList">
              {locItems.map((item) => {
                const statusConf = STATUS_CONFIG[item.status]
                const isAnimating = animatingIds.has(item.id)
                return (
                  <div
                    key={item.id}
                    className={`packItem ${
                      item.status === 'gepackt' ? 'packItemPacked' : ''
                    } ${isAnimating ? 'packItemAnimating' : ''}`}
                  >
                    {/* Status checkbox area */}
                    <div className="packItemCheck">
                      <button
                        className={`checkBtn ${
                          item.status === 'gepackt' ? 'checkBtnPacked' : ''
                        }`}
                        onClick={() => toggleItemStatus(item.id, 'gepackt')}
                        title="Als gepackt markieren"
                      >
                        {item.status === 'gepackt' ? (
                          <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                            <path d="M4 10l4 4 8-8" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
                          </svg>
                        ) : (
                          <span className="checkBtnEmpty" />
                        )}
                      </button>
                    </div>

                    {/* Item info */}
                    <div className="packItemInfo">
                      <div className="packItemName">{item.name}</div>
                      <div className="packItemMeta">
                        <span>{item.quantity}x</span>
                        {item.barcode && (
                          <span className="serialNumber">{item.barcode}</span>
                        )}
                        <span>{item.quantity_packed}/{item.quantity} gepackt</span>
                      </div>
                    </div>

                    {/* Status badge */}
                    <div className="packItemStatus">
                      <span
                        className="statusBadge"
                        style={{
                          color: statusConf.color,
                          backgroundColor: statusConf.bgColor,
                        }}
                      >
                        {statusConf.icon} {statusConf.label}
                      </span>
                    </div>

                    {/* Status action buttons */}
                    <div className="packItemActions">
                      <button
                        className="miniBtn miniBtnDanger"
                        onClick={() => toggleItemStatus(item.id, 'fehlt')}
                        title="Als fehlend markieren"
                      >
                        {'\u274C'}
                      </button>
                      <button
                        className="miniBtn miniBtnWarning"
                        onClick={() => toggleItemStatus(item.id, 'ersetzt')}
                        title="Als ersetzt markieren"
                      >
                        {'\u26A0\uFE0F'}
                      </button>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )
      })}
    </div>
  )
}

export default PackingListTab
