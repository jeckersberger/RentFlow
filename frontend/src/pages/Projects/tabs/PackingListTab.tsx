import { useState, useCallback, useMemo } from 'react'
import { Project } from '../../../types/project'
import styles from './PackingListTab.module.scss'

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
  serial_number?: string
  weight?: number
  barcode?: string
  status: PackStatus
  replaced_by?: string
}

interface PackingListTabProps {
  project: Project
  /** Optional: called when an item status changes (for API sync) */
  onItemStatusChange?: (itemId: string, status: PackStatus) => void
}

// ============================================================================
// Mock Data (until real API exists)
// ============================================================================

const MOCK_PACKING_ITEMS: PackingItem[] = [
  // Audio
  { id: 'pi-1', equipment_id: 'eq-1', name: 'Shure SM58', category: 'Audio', quantity: 4, serial_number: 'SM58-2024-001', weight: 0.33, barcode: 'RF-SM58-001', status: 'ausstehend' },
  { id: 'pi-2', equipment_id: 'eq-2', name: 'Sennheiser EW 100 G4', category: 'Audio', quantity: 2, serial_number: 'EW100-2023-015', weight: 0.45, barcode: 'RF-EW100-015', status: 'ausstehend' },
  { id: 'pi-3', equipment_id: 'eq-3', name: 'Yamaha TF3 Mischpult', category: 'Audio', quantity: 1, serial_number: 'TF3-2022-003', weight: 16.9, barcode: 'RF-TF3-003', status: 'ausstehend' },
  { id: 'pi-4', equipment_id: 'eq-4', name: 'QSC K12.2 Lautsprecher', category: 'Audio', quantity: 4, serial_number: 'K12-2023-008', weight: 16.8, barcode: 'RF-K12-008', status: 'ausstehend' },
  { id: 'pi-5', equipment_id: 'eq-5', name: 'XLR Kabel 10m', category: 'Audio', quantity: 12, weight: 0.8, barcode: 'RF-XLR10-BULK', status: 'ausstehend' },
  // Licht
  { id: 'pi-6', equipment_id: 'eq-6', name: 'Chauvet Rogue R2 Wash', category: 'Licht', quantity: 6, serial_number: 'R2W-2024-012', weight: 8.7, barcode: 'RF-R2W-012', status: 'ausstehend' },
  { id: 'pi-7', equipment_id: 'eq-7', name: 'ETC Source Four 750W', category: 'Licht', quantity: 8, serial_number: 'S4-2022-044', weight: 7.6, barcode: 'RF-S4-044', status: 'ausstehend' },
  { id: 'pi-8', equipment_id: 'eq-8', name: 'GrandMA3 Light', category: 'Licht', quantity: 1, serial_number: 'GMA3-2024-001', weight: 12.5, barcode: 'RF-GMA3-001', status: 'ausstehend' },
  { id: 'pi-9', equipment_id: 'eq-9', name: 'DMX Kabel 5m', category: 'Licht', quantity: 10, weight: 0.4, barcode: 'RF-DMX5-BULK', status: 'ausstehend' },
  // Video
  { id: 'pi-10', equipment_id: 'eq-10', name: 'Panasonic PT-RZ690', category: 'Video', quantity: 1, serial_number: 'RZ690-2023-002', weight: 18.3, barcode: 'RF-RZ690-002', status: 'ausstehend' },
  { id: 'pi-11', equipment_id: 'eq-11', name: 'Blackmagic ATEM Mini Pro', category: 'Video', quantity: 1, serial_number: 'ATEM-2024-005', weight: 0.55, barcode: 'RF-ATEM-005', status: 'ausstehend' },
  { id: 'pi-12', equipment_id: 'eq-12', name: 'HDMI Kabel 15m', category: 'Video', quantity: 4, weight: 0.9, barcode: 'RF-HDMI15-BULK', status: 'ausstehend' },
  // Buhnentechnik
  { id: 'pi-13', equipment_id: 'eq-13', name: 'Eurotruss FD34 3m', category: 'Buhnentechnik', quantity: 8, serial_number: 'FD34-2021-020', weight: 9.2, barcode: 'RF-FD34-020', status: 'ausstehend' },
  { id: 'pi-14', equipment_id: 'eq-14', name: 'Chain Motor 0.5t', category: 'Buhnentechnik', quantity: 4, serial_number: 'CM05-2023-011', weight: 22.0, barcode: 'RF-CM05-011', status: 'ausstehend' },
  { id: 'pi-15', equipment_id: 'eq-15', name: 'Stageblock 2x1m', category: 'Buhnentechnik', quantity: 6, weight: 35.0, barcode: 'RF-SB21-BULK', status: 'ausstehend' },
]

// Category config: icon, label, sort order
const CATEGORY_CONFIG: Record<string, { icon: string; label: string; order: number }> = {
  Audio: { icon: '\u{1F3B5}', label: 'Audio', order: 1 },
  Licht: { icon: '\u{1F4A1}', label: 'Licht', order: 2 },
  Video: { icon: '\u{1F4F9}', label: 'Video', order: 3 },
  Buhnentechnik: { icon: '\u{1F3AD}', label: 'Buhnentechnik', order: 4 },
  Sonstiges: { icon: '\u{1F4E6}', label: 'Sonstiges', order: 99 },
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
  const [items, setItems] = useState<PackingItem[]>(MOCK_PACKING_ITEMS)
  const [filter, setFilter] = useState<'alle' | PackStatus>('alle')
  const [animatingIds, setAnimatingIds] = useState<Set<string>>(new Set())
  const [isPrintMode, setIsPrintMode] = useState(false)

  // ---- Computed values ----
  const totalItems = items.length
  const packedCount = items.filter((i) => i.status === 'gepackt').length
  const fehltCount = items.filter((i) => i.status === 'fehlt').length
  const ersetztCount = items.filter((i) => i.status === 'ersetzt').length
  const progressPercent = totalItems > 0 ? Math.round((packedCount / totalItems) * 100) : 0
  const totalWeight = items.reduce((sum, i) => sum + (i.weight || 0) * i.quantity, 0)

  const filteredItems = useMemo(() => {
    if (filter === 'alle') return items
    return items.filter((i) => i.status === filter)
  }, [items, filter])

  // Group items by category
  const groupedItems = useMemo(() => {
    const groups: Record<string, PackingItem[]> = {}
    for (const item of filteredItems) {
      const cat = item.category || 'Sonstiges'
      if (!groups[cat]) groups[cat] = []
      groups[cat].push(item)
    }
    // Sort groups by category order
    const sortedEntries = Object.entries(groups).sort((a, b) => {
      const orderA = CATEGORY_CONFIG[a[0]]?.order ?? 99
      const orderB = CATEGORY_CONFIG[b[0]]?.order ?? 99
      return orderA - orderB
    })
    return sortedEntries
  }, [filteredItems])

  // ---- Handlers ----
  const toggleItemStatus = useCallback((itemId: string, newStatus: PackStatus) => {
    setItems((prev) =>
      prev.map((item) => {
        if (item.id !== itemId) return item
        // If clicking the same status, toggle back to ausstehend
        const finalStatus = item.status === newStatus ? 'ausstehend' : newStatus
        onItemStatusChange?.(itemId, finalStatus)
        return { ...item, status: finalStatus }
      })
    )
    // Trigger animation
    setAnimatingIds((prev) => new Set(prev).add(itemId))
    setTimeout(() => {
      setAnimatingIds((prev) => {
        const next = new Set(prev)
        next.delete(itemId)
        return next
      })
    }, 600)
  }, [onItemStatusChange])

  const markAllPacked = useCallback(() => {
    setItems((prev) =>
      prev.map((item) => {
        if (item.status === 'ausstehend') {
          onItemStatusChange?.(item.id, 'gepackt')
          return { ...item, status: 'gepackt' as PackStatus }
        }
        return item
      })
    )
    // Flash all items
    const allIds = items.filter((i) => i.status === 'ausstehend').map((i) => i.id)
    setAnimatingIds(new Set(allIds))
    setTimeout(() => setAnimatingIds(new Set()), 600)
  }, [items, onItemStatusChange])

  const resetAll = useCallback(() => {
    setItems((prev) =>
      prev.map((item) => {
        onItemStatusChange?.(item.id, 'ausstehend')
        return { ...item, status: 'ausstehend' as PackStatus }
      })
    )
  }, [onItemStatusChange])

  /** Called from scanner: verify a barcode against this packing list */
  const verifyBarcode = useCallback((barcode: string): { found: boolean; item?: PackingItem } => {
    const item = items.find((i) => i.barcode === barcode)
    if (!item) return { found: false }
    if (item.status !== 'gepackt') {
      toggleItemStatus(item.id, 'gepackt')
    }
    return { found: true, item }
  }, [items, toggleItemStatus])

  // Expose verifyBarcode via window for scanner integration
  // In a real app, this would use React context or a store
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

  // ---- Render ----
  if (isPrintMode) {
    return (
      <div className={styles.printView}>
        <div className={styles.printHeader}>
          <h1>Packliste</h1>
          <div className={styles.printMeta}>
            <p><strong>Projekt:</strong> {project.name}</p>
            <p><strong>Datum:</strong> {new Date().toLocaleDateString('de-DE')}</p>
            <p><strong>Zeitraum:</strong> {new Date(project.start_date).toLocaleDateString('de-DE')} &ndash; {new Date(project.end_date).toLocaleDateString('de-DE')}</p>
            {project.client_name && <p><strong>Kunde:</strong> {project.client_name}</p>}
          </div>
          <div className={styles.printProgress}>
            {packedCount} von {totalItems} Positionen gepackt ({progressPercent}%) &bull; Gesamtgewicht: {totalWeight.toFixed(1)} kg
          </div>
        </div>
        {groupedItems.map(([category, catItems]) => (
          <div key={category} className={styles.printCategory}>
            <h2>{CATEGORY_CONFIG[category]?.label || category}</h2>
            <table className={styles.printTable}>
              <thead>
                <tr>
                  <th style={{ width: '30px' }}></th>
                  <th>Artikel</th>
                  <th>Menge</th>
                  <th>Seriennr.</th>
                  <th>Gewicht</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {catItems.map((item) => (
                  <tr key={item.id}>
                    <td className={styles.printCheckbox}>
                      {item.status === 'gepackt' ? '\u2611' : '\u2610'}
                    </td>
                    <td>{item.name}</td>
                    <td>{item.quantity}x</td>
                    <td>{item.serial_number || '\u2014'}</td>
                    <td>{item.weight ? `${(item.weight * item.quantity).toFixed(1)} kg` : '\u2014'}</td>
                    <td>{STATUS_CONFIG[item.status].label}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ))}
        <div className={styles.printFooter}>
          <p>Gepackt von: _________________________ &nbsp;&nbsp; Datum: _________________________</p>
          <p>Kontrolliert von: _________________________ &nbsp;&nbsp; Unterschrift: _________________________</p>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.packingList}>
      {/* Header with progress */}
      <div className={styles.header}>
        <div className={styles.headerLeft}>
          <h3 className={styles.title}>Packliste</h3>
          <span className={styles.projectLabel}>{project.name}</span>
        </div>
        <div className={styles.headerActions}>
          <button className={`btn btn--secondary ${styles.actionBtn}`} onClick={handlePrint}>
            Packliste drucken
          </button>
          <button className={`btn btn--secondary ${styles.actionBtn}`} onClick={resetAll}>
            Zuruecksetzen
          </button>
          <button
            className={`btn btn--primary ${styles.actionBtn}`}
            onClick={markAllPacked}
            disabled={packedCount === totalItems}
          >
            Alles als gepackt markieren
          </button>
        </div>
      </div>

      {/* Progress bar */}
      <div className={styles.progressSection}>
        <div className={styles.progressInfo}>
          <span className={styles.progressText}>
            {packedCount} von {totalItems} Positionen gepackt ({progressPercent}%)
          </span>
          <div className={styles.progressStats}>
            {fehltCount > 0 && (
              <span className={styles.statBadgeDanger}>{fehltCount} fehlt</span>
            )}
            {ersetztCount > 0 && (
              <span className={styles.statBadgeWarning}>{ersetztCount} ersetzt</span>
            )}
            <span className={styles.statBadgeInfo}>
              {totalWeight.toFixed(1)} kg Gesamtgewicht
            </span>
          </div>
        </div>
        <div className={styles.progressBar}>
          <div
            className={styles.progressFill}
            style={{ width: `${progressPercent}%` }}
          />
        </div>
      </div>

      {/* Filter bar */}
      <div className={styles.filterBar}>
        {(['alle', 'ausstehend', 'gepackt', 'fehlt', 'ersetzt'] as const).map((f) => (
          <button
            key={f}
            className={`${styles.filterBtn} ${filter === f ? styles.filterBtnActive : ''}`}
            onClick={() => setFilter(f)}
          >
            {f === 'alle' ? 'Alle' : STATUS_CONFIG[f].icon + ' ' + STATUS_CONFIG[f].label}
            <span className={styles.filterCount}>
              {f === 'alle'
                ? totalItems
                : items.filter((i) => i.status === f).length}
            </span>
          </button>
        ))}
      </div>

      {/* Empty state */}
      {totalItems === 0 && (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F4E6}'}</div>
          <h4 className={styles.emptyStateTitle}>Keine Artikel in der Packliste</h4>
          <p className={styles.emptyStateText}>
            Weisen Sie zuerst Equipment im Equipment-Tab zu, um die Packliste zu erstellen.
          </p>
        </div>
      )}

      {/* Category groups */}
      {groupedItems.map(([category, catItems]) => {
        const catConfig = CATEGORY_CONFIG[category] || CATEGORY_CONFIG.Sonstiges
        const catPacked = catItems.filter((i) => i.status === 'gepackt').length
        return (
          <div key={category} className={styles.categoryGroup}>
            <div className={styles.categoryHeader}>
              <span className={styles.categoryIcon}>{catConfig.icon}</span>
              <span className={styles.categoryLabel}>{catConfig.label}</span>
              <span className={styles.categoryCount}>
                {catPacked}/{catItems.length} gepackt
              </span>
            </div>
            <div className={styles.itemList}>
              {catItems.map((item) => {
                const statusConf = STATUS_CONFIG[item.status]
                const isAnimating = animatingIds.has(item.id)
                return (
                  <div
                    key={item.id}
                    className={`${styles.packItem} ${
                      item.status === 'gepackt' ? styles.packItemPacked : ''
                    } ${isAnimating ? styles.packItemAnimating : ''}`}
                  >
                    {/* Status checkbox area */}
                    <div className={styles.packItemCheck}>
                      <button
                        className={`${styles.checkBtn} ${
                          item.status === 'gepackt' ? styles.checkBtnPacked : ''
                        }`}
                        onClick={() => toggleItemStatus(item.id, 'gepackt')}
                        title="Als gepackt markieren"
                      >
                        {item.status === 'gepackt' ? (
                          <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                            <path d="M4 10l4 4 8-8" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
                          </svg>
                        ) : (
                          <span className={styles.checkBtnEmpty} />
                        )}
                      </button>
                    </div>

                    {/* Item info */}
                    <div className={styles.packItemInfo}>
                      <div className={styles.packItemName}>{item.name}</div>
                      <div className={styles.packItemMeta}>
                        <span>{item.quantity}x</span>
                        {item.serial_number && (
                          <span className={styles.serialNumber}>SN: {item.serial_number}</span>
                        )}
                        {item.weight && (
                          <span>{(item.weight * item.quantity).toFixed(1)} kg</span>
                        )}
                      </div>
                    </div>

                    {/* Status badge */}
                    <div className={styles.packItemStatus}>
                      <span
                        className={styles.statusBadge}
                        style={{
                          color: statusConf.color,
                          backgroundColor: statusConf.bgColor,
                        }}
                      >
                        {statusConf.icon} {statusConf.label}
                      </span>
                    </div>

                    {/* Status action buttons */}
                    <div className={styles.packItemActions}>
                      <button
                        className={`${styles.miniBtn} ${styles.miniBtnDanger}`}
                        onClick={() => toggleItemStatus(item.id, 'fehlt')}
                        title="Als fehlend markieren"
                      >
                        {'\u274C'}
                      </button>
                      <button
                        className={`${styles.miniBtn} ${styles.miniBtnWarning}`}
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
