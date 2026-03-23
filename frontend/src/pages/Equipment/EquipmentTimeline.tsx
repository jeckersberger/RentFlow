import { useState, useRef, useMemo, useCallback, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { equipmentApi, projectApi, categoryApi } from '../../services/api'
import { Equipment, Category } from '../../types/equipment'
import { Project } from '../../types/project'
import styles from './EquipmentTimeline.module.scss'

// ============================================================================
// Types
// ============================================================================

type ViewMode = 'day' | 'week'

interface Booking {
  id: string
  equipmentId: string
  projectId?: string
  projectName: string
  clientName: string
  type: 'reserved' | 'checked_out' | 'maintenance' | 'damaged'
  startDate: Date
  endDate: Date
}

interface TooltipData {
  booking: Booking
  x: number
  y: number
}

// ============================================================================
// Helpers
// ============================================================================

const DAY_NAMES_DE = ['So', 'Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa']
const MONTH_NAMES_DE = [
  'Januar', 'Februar', 'März', 'April', 'Mai', 'Juni',
  'Juli', 'August', 'September', 'Oktober', 'November', 'Dezember',
]
const MONTH_SHORT_DE = [
  'Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun',
  'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez',
]

function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

function addDays(d: Date, n: number): Date {
  const r = new Date(d)
  r.setDate(r.getDate() + n)
  return r
}

function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

function isWeekend(d: Date): boolean {
  const day = d.getDay()
  return day === 0 || day === 6
}

function getMonday(d: Date): Date {
  const result = new Date(d)
  const day = result.getDay()
  const diff = day === 0 ? -6 : 1 - day
  result.setDate(result.getDate() + diff)
  return startOfDay(result)
}

function formatDateRange(start: Date, end: Date): string {
  const s = `${start.getDate()}. ${MONTH_SHORT_DE[start.getMonth()]}`
  const e = `${end.getDate()}. ${MONTH_SHORT_DE[end.getMonth()]} ${end.getFullYear()}`
  return `${s} - ${e}`
}

// ============================================================================
// Category colors for grouping
// ============================================================================

const CATEGORY_COLORS: Record<string, string> = {
  'cat-audio': '#3b82f6',
  'cat-lighting': '#f59e0b',
  'cat-video': '#8b5cf6',
  'cat-stage': '#10b981',
  'default': '#6b7280',
}

function getCategoryColor(catId: string): string {
  return CATEGORY_COLORS[catId] || CATEGORY_COLORS['default']
}

// ============================================================================
// Generate simulated bookings from project dates + equipment status
// ============================================================================

function generateBookings(equipment: Equipment[], projects: Project[]): Booking[] {
  const bookings: Booking[] = []

  // Assign equipment to projects based on a deterministic pattern
  // This simulates real bookings for the timeline view
  const activeProjects = projects.filter(p =>
    p.status !== 'cancelled' && p.status !== 'completed'
  )
  const completedProjects = projects.filter(p => p.status === 'completed')

  // Simulate: distribute equipment across active projects
  activeProjects.forEach((project, pIdx) => {
    // Each project gets some equipment assigned
    const eqStart = (pIdx * 3) % equipment.length
    const eqCount = 2 + (pIdx % 3) // 2-4 items per project

    for (let i = 0; i < eqCount && i < equipment.length; i++) {
      const eq = equipment[(eqStart + i) % equipment.length]
      const start = new Date(project.start_date)
      const end = new Date(project.end_date)

      // Add setup day before
      const setupStart = project.setup_date ? new Date(project.setup_date) : addDays(start, -1)

      bookings.push({
        id: `booking-${project.id}-${eq.id}`,
        equipmentId: eq.id,
        projectId: project.id,
        projectName: project.name,
        clientName: project.client_name || (project as unknown as Record<string, unknown>).client as string || 'Unbekannt',
        type: 'reserved',
        startDate: startOfDay(setupStart),
        endDate: startOfDay(end),
      })
    }
  })

  // Completed projects - show as checked_out in the past
  completedProjects.forEach((project, pIdx) => {
    const eqStart = (pIdx * 2 + 1) % equipment.length
    for (let i = 0; i < 2 && i < equipment.length; i++) {
      const eq = equipment[(eqStart + i) % equipment.length]
      bookings.push({
        id: `booking-done-${project.id}-${eq.id}`,
        equipmentId: eq.id,
        projectId: project.id,
        projectName: project.name,
        clientName: project.client_name || (project as unknown as Record<string, unknown>).client as string || 'Unbekannt',
        type: 'checked_out',
        startDate: startOfDay(new Date(project.start_date)),
        endDate: startOfDay(new Date(project.end_date)),
      })
    }
  })

  // Add maintenance bookings for equipment currently in maintenance
  equipment.forEach(eq => {
    if (eq.status === 'in_maintenance') {
      const now = new Date()
      bookings.push({
        id: `maint-${eq.id}`,
        equipmentId: eq.id,
        projectName: 'Wartung',
        clientName: '',
        type: 'maintenance',
        startDate: startOfDay(addDays(now, -3)),
        endDate: startOfDay(addDays(now, 4)),
      })
    }
    if (eq.status === 'damaged') {
      const now = new Date()
      bookings.push({
        id: `dmg-${eq.id}`,
        equipmentId: eq.id,
        projectName: 'Beschädigt',
        clientName: '',
        type: 'damaged',
        startDate: startOfDay(addDays(now, -7)),
        endDate: startOfDay(addDays(now, 14)),
      })
    }
  })

  return bookings
}

// ============================================================================
// Component
// ============================================================================

function EquipmentTimeline() {
  const navigate = useNavigate()
  const gridBodyRef = useRef<HTMLDivElement>(null)
  const sidebarBodyRef = useRef<HTMLDivElement>(null)
  const gridHeaderRef = useRef<HTMLDivElement>(null)

  const [viewMode, setViewMode] = useState<ViewMode>('day')
  const [startDate, setStartDate] = useState(() => getMonday(new Date()))
  const [tooltip, setTooltip] = useState<TooltipData | null>(null)

  const today = useMemo(() => startOfDay(new Date()), [])

  // Number of visible days depends on view mode
  const visibleDays = viewMode === 'day' ? 28 : 56 // 4 weeks or 8 weeks
  const cellWidth = viewMode === 'day' ? 48 : 22

  // Generate array of visible dates
  const dates = useMemo(() => {
    const result: Date[] = []
    for (let i = 0; i < visibleDays; i++) {
      result.push(addDays(startDate, i))
    }
    return result
  }, [startDate, visibleDays])

  const endDate = dates[dates.length - 1]

  // ---- Data fetching ----

  const { data: equipmentData, isLoading: eqLoading } = useQuery({
    queryKey: ['equipment-timeline'],
    queryFn: () => equipmentApi.list({ limit: 200, offset: 0 }),
    staleTime: 1000 * 60 * 5,
  })

  const { data: projectData, isLoading: projLoading } = useQuery({
    queryKey: ['projects-timeline'],
    queryFn: () => projectApi.list(1, 100),
    staleTime: 1000 * 60 * 5,
  })

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list() as Promise<Category[]>,
    staleTime: 1000 * 60 * 10,
  })

  const equipment: Equipment[] = equipmentData?.data || []
  const projects: Project[] = projectData?.items || projectData?.data || []
  const isLoading = eqLoading || projLoading

  // ---- Group equipment by category ----

  const categoryMap = useMemo(() => {
    const map = new Map<string, Category>()
    ;(categories || []).forEach((c: Category) => map.set(c.id, c))
    return map
  }, [categories])

  const groupedEquipment = useMemo(() => {
    const groups = new Map<string, { category: Category | null; items: Equipment[] }>()
    equipment.forEach(eq => {
      const catId = eq.category_id || 'uncategorized'
      if (!groups.has(catId)) {
        groups.set(catId, { category: categoryMap.get(catId) || null, items: [] })
      }
      groups.get(catId)!.items.push(eq)
    })
    // Sort categories by name
    return Array.from(groups.entries()).sort((a, b) => {
      const nameA = a[1].category?.name || 'Sonstige'
      const nameB = b[1].category?.name || 'Sonstige'
      return nameA.localeCompare(nameB)
    })
  }, [equipment, categoryMap])

  // Flat list for rendering (category headers + equipment rows)
  const flatRows = useMemo(() => {
    const rows: Array<{ type: 'category'; catId: string; name: string; color: string } | { type: 'equipment'; eq: Equipment }> = []
    groupedEquipment.forEach(([catId, group]) => {
      rows.push({
        type: 'category',
        catId,
        name: group.category?.name || 'Sonstige',
        color: group.category?.color || getCategoryColor(catId),
      })
      group.items.forEach(eq => {
        rows.push({ type: 'equipment', eq })
      })
    })
    return rows
  }, [groupedEquipment])

  // ---- Generate bookings ----

  const bookings = useMemo(() => {
    if (equipment.length === 0 || projects.length === 0) return []
    return generateBookings(equipment, projects)
  }, [equipment, projects])

  // Index bookings by equipment ID for fast lookup
  const bookingsByEquipment = useMemo(() => {
    const map = new Map<string, Booking[]>()
    bookings.forEach(b => {
      if (!map.has(b.equipmentId)) map.set(b.equipmentId, [])
      map.get(b.equipmentId)!.push(b)
    })
    return map
  }, [bookings])

  // ---- Sync sidebar + grid scroll ----

  const handleGridScroll = useCallback(() => {
    const gridBody = gridBodyRef.current
    const sidebarBody = sidebarBodyRef.current
    const gridHeader = gridHeaderRef.current

    if (gridBody && sidebarBody) {
      sidebarBody.scrollTop = gridBody.scrollTop
    }
    if (gridBody && gridHeader) {
      gridHeader.scrollLeft = gridBody.scrollLeft
    }
  }, [])

  useEffect(() => {
    const gridBody = gridBodyRef.current
    if (gridBody) {
      gridBody.addEventListener('scroll', handleGridScroll, { passive: true })
      return () => gridBody.removeEventListener('scroll', handleGridScroll)
    }
  }, [handleGridScroll])

  // Scroll today into view on first load
  useEffect(() => {
    if (!isLoading && gridBodyRef.current) {
      const todayIdx = dates.findIndex(d => isSameDay(d, today))
      if (todayIdx > 3) {
        const scrollTo = (todayIdx - 3) * cellWidth
        gridBodyRef.current.scrollLeft = scrollTo
        if (gridHeaderRef.current) gridHeaderRef.current.scrollLeft = scrollTo
      }
    }
  }, [isLoading, dates, today, cellWidth])

  // ---- Navigation ----

  const goToday = () => {
    setStartDate(getMonday(new Date()))
  }

  const goPrev = () => {
    setStartDate(prev => addDays(prev, viewMode === 'day' ? -7 : -14))
  }

  const goNext = () => {
    setStartDate(prev => addDays(prev, viewMode === 'day' ? 7 : 14))
  }

  // ---- Bar position calculation ----

  const getBarStyle = useCallback((booking: Booking): React.CSSProperties | null => {
    const timelineStart = startDate.getTime()
    const timelineEnd = addDays(startDate, visibleDays).getTime()

    const bStart = Math.max(booking.startDate.getTime(), timelineStart)
    const bEnd = Math.min(booking.endDate.getTime() + 86400000, timelineEnd) // include end day

    if (bEnd <= timelineStart || bStart >= timelineEnd) return null // out of range

    const msPerDay = 86400000
    const leftDays = (bStart - timelineStart) / msPerDay
    const widthDays = (bEnd - bStart) / msPerDay

    return {
      left: `${leftDays * cellWidth}px`,
      width: `${Math.max(widthDays * cellWidth - 2, 4)}px`,
    }
  }, [startDate, visibleDays, cellWidth])

  const getBarClass = (type: Booking['type']): string => {
    switch (type) {
      case 'reserved': return styles.barReserved
      case 'checked_out': return styles.barCheckedOut
      case 'maintenance': return styles.barMaintenance
      case 'damaged': return styles.barDamaged
      default: return styles.barReserved
    }
  }

  const getTooltipTypeClass = (type: Booking['type']): string => {
    switch (type) {
      case 'reserved': return styles.tooltipTypeReserved
      case 'checked_out': return styles.tooltipTypeCheckedOut
      case 'maintenance': return styles.tooltipTypeMaintenance
      case 'damaged': return styles.tooltipTypeDamaged
      default: return ''
    }
  }

  const getTypeLabel = (type: Booking['type']): string => {
    switch (type) {
      case 'reserved': return 'Reserviert'
      case 'checked_out': return 'Vermietet'
      case 'maintenance': return 'Wartung'
      case 'damaged': return 'Beschädigt'
      default: return type
    }
  }

  // ---- Tooltip handlers ----

  const handleBarHover = useCallback((e: React.MouseEvent, booking: Booking) => {
    setTooltip({
      booking,
      x: e.clientX + 12,
      y: e.clientY - 10,
    })
  }, [])

  const handleBarMove = useCallback((e: React.MouseEvent) => {
    setTooltip(prev => prev ? { ...prev, x: e.clientX + 12, y: e.clientY - 10 } : null)
  }, [])

  const handleBarLeave = useCallback(() => {
    setTooltip(null)
  }, [])

  // ---- Status dot color ----

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'available': return '#10b981'
      case 'reserved': return '#3b82f6'
      case 'checked_out': return '#8b5cf6'
      case 'in_maintenance': return '#f59e0b'
      case 'damaged': return '#ef4444'
      case 'retired': return '#6b7280'
      default: return '#6b7280'
    }
  }

  // ---- Today line position ----

  const todayLineLeft = useMemo(() => {
    const todayMs = today.getTime()
    const startMs = startDate.getTime()
    const diff = (todayMs - startMs) / 86400000
    if (diff < 0 || diff >= visibleDays) return null
    return diff * cellWidth + cellWidth / 2
  }, [today, startDate, visibleDays, cellWidth])

  // ============================================================================
  // Render
  // ============================================================================

  return (
    <div className={styles.timelinePage}>
      {/* Header */}
      <div className={styles.header}>
        <div className={styles.headerLeft}>
          <h1 className={styles.title}>Verfügbarkeits-Zeitleiste</h1>
          <p className={styles.subtitle}>Equipment-Auslastung und Projektplanung im Überblick</p>
        </div>
        <div className={styles.headerActions}>
          <button
            className="btn btn--secondary"
            onClick={() => navigate('/equipment')}
          >
            Zurück zur Liste
          </button>
          <button
            className="btn btn--primary"
            onClick={() => navigate('/equipment/new')}
          >
            + Equipment hinzufügen
          </button>
        </div>
      </div>

      {/* Toolbar */}
      <div className={styles.toolbar}>
        <div className={styles.dateNav}>
          <button className={styles.navBtn} onClick={goPrev} title="Zurück">
            &#9664;
          </button>
          <button className={`${styles.navBtn} ${styles.navBtnActive}`} onClick={goToday}>
            Heute
          </button>
          <button className={styles.navBtn} onClick={goNext} title="Vorwärts">
            &#9654;
          </button>
          <span className={styles.dateDisplay}>
            {MONTH_NAMES_DE[startDate.getMonth()]} {startDate.getFullYear()}
            {startDate.getMonth() !== endDate.getMonth() && ` – ${MONTH_SHORT_DE[endDate.getMonth()]} ${endDate.getFullYear()}`}
          </span>
        </div>

        <div className={styles.legend}>
          <div className={styles.legendItem}>
            <span className={`${styles.legendDot} ${styles.legendReserved}`} />
            <span>Reserviert</span>
          </div>
          <div className={styles.legendItem}>
            <span className={`${styles.legendDot} ${styles.legendCheckedOut}`} />
            <span>Vermietet</span>
          </div>
          <div className={styles.legendItem}>
            <span className={`${styles.legendDot} ${styles.legendMaintenance}`} />
            <span>Wartung</span>
          </div>
          <div className={styles.legendItem}>
            <span className={`${styles.legendDot} ${styles.legendDamaged}`} />
            <span>Beschädigt</span>
          </div>
          <div className={styles.legendItem}>
            <span className={`${styles.legendDot} ${styles.legendAvailable}`} />
            <span>Verfügbar</span>
          </div>
        </div>

        <div className={styles.zoomToggle}>
          <button
            className={`${styles.zoomBtn} ${viewMode === 'day' ? styles.zoomBtnActive : ''}`}
            onClick={() => setViewMode('day')}
          >
            Tage
          </button>
          <button
            className={`${styles.zoomBtn} ${viewMode === 'week' ? styles.zoomBtnActive : ''}`}
            onClick={() => setViewMode('week')}
          >
            Wochen
          </button>
        </div>
      </div>

      {/* Timeline */}
      {isLoading ? (
        <div className={styles.loadingState}>Lade Zeitleiste...</div>
      ) : equipment.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>📦</div>
          <div className={styles.emptyStateText}>Noch kein Equipment vorhanden</div>
          <button className="btn btn--primary" onClick={() => navigate('/equipment/new')}>
            Erstes Equipment anlegen
          </button>
        </div>
      ) : (
        <div className={styles.timelineContainer}>
          {/* Left sidebar - equipment names */}
          <div className={styles.sidebarColumn}>
            <div className={styles.sidebarHeader}>Equipment</div>
            <div className={styles.sidebarBody} ref={sidebarBodyRef}>
              {flatRows.map((row) => {
                if (row.type === 'category') {
                  return (
                    <div key={`cat-${row.catId}`} className={styles.categoryGroupHeader}>
                      <span className={styles.categoryGroupHeaderDot} style={{ backgroundColor: row.color }} />
                      {row.name}
                    </div>
                  )
                }
                return (
                  <div
                    key={`eq-${row.eq.id}`}
                    className={styles.equipmentRow}
                    onClick={() => navigate(`/equipment/${row.eq.id}`)}
                    title={row.eq.name}
                  >
                    <span
                      className={styles.equipmentStatus}
                      style={{ backgroundColor: getStatusColor(row.eq.status) }}
                    />
                    <span className={styles.equipmentName}>{row.eq.name}</span>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Right area - grid */}
          <div className={styles.gridArea}>
            {/* Grid header - dates */}
            <div className={styles.gridHeader} ref={gridHeaderRef}>
              <div className={styles.gridHeaderInner}>
                {dates.map((date, idx) => {
                  const isToday = isSameDay(date, today)
                  const isWkend = isWeekend(date)
                  const showMonth = idx === 0 || date.getDate() === 1

                  return (
                    <div
                      key={idx}
                      className={`${styles.dayColumn} ${isWkend ? styles.dayColumnWeekend : ''} ${isToday ? styles.dayColumnToday : ''}`}
                      style={{ width: cellWidth }}
                    >
                      <span className={styles.dayName}>{DAY_NAMES_DE[date.getDay()]}</span>
                      <span className={styles.dayNumber}>{date.getDate()}</span>
                      {showMonth && <span className={styles.dayMonth}>{MONTH_SHORT_DE[date.getMonth()]}</span>}
                    </div>
                  )
                })}
              </div>
            </div>

            {/* Grid body - rows with bars */}
            <div className={styles.gridBody} ref={gridBodyRef}>
              <div className={styles.gridBodyInner} style={{ width: visibleDays * cellWidth }}>
                {/* Today line */}
                {todayLineLeft !== null && (
                  <div className={styles.todayLine} style={{ left: todayLineLeft }} />
                )}

                {flatRows.map((row) => {
                  if (row.type === 'category') {
                    return (
                      <div key={`catrow-${row.catId}`} className={styles.categoryRowHeader}>
                        {dates.map((date, dIdx) => (
                          <div
                            key={dIdx}
                            className={`${styles.gridCell} ${isWeekend(date) ? styles.gridCellWeekend : ''}`}
                            style={{ width: cellWidth, height: 28 }}
                          />
                        ))}
                      </div>
                    )
                  }

                  const eqBookings = bookingsByEquipment.get(row.eq.id) || []

                  return (
                    <div key={`row-${row.eq.id}`} className={styles.gridRow}>
                      {/* Cell backgrounds */}
                      {dates.map((date, dIdx) => (
                        <div
                          key={dIdx}
                          className={`${styles.gridCell} ${isWeekend(date) ? styles.gridCellWeekend : ''} ${isSameDay(date, today) ? styles.gridCellToday : ''}`}
                          style={{ width: cellWidth, height: 40 }}
                        />
                      ))}

                      {/* Booking bars */}
                      {eqBookings.map(booking => {
                        const barStyle = getBarStyle(booking)
                        if (!barStyle) return null

                        return (
                          <div
                            key={booking.id}
                            className={`${styles.bookingBar} ${getBarClass(booking.type)}`}
                            style={barStyle}
                            onClick={(e) => {
                              e.stopPropagation()
                              if (booking.projectId) navigate(`/projects/${booking.projectId}`)
                            }}
                            onMouseEnter={(e) => handleBarHover(e, booking)}
                            onMouseMove={handleBarMove}
                            onMouseLeave={handleBarLeave}
                          >
                            <span className={styles.bookingBarLabel}>
                              {booking.projectName}
                            </span>
                          </div>
                        )
                      })}
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tooltip */}
      {tooltip && (
        <div
          className={styles.tooltip}
          style={{ left: tooltip.x, top: tooltip.y }}
        >
          <div className={`${styles.tooltipType} ${getTooltipTypeClass(tooltip.booking.type)}`}>
            {getTypeLabel(tooltip.booking.type)}
          </div>
          <p className={styles.tooltipTitle}>{tooltip.booking.projectName}</p>
          {tooltip.booking.clientName && (
            <div className={styles.tooltipRow}>
              <span className={styles.tooltipRowLabel}>Kunde</span>
              <span className={styles.tooltipRowValue}>{tooltip.booking.clientName}</span>
            </div>
          )}
          <div className={styles.tooltipRow}>
            <span className={styles.tooltipRowLabel}>Zeitraum</span>
            <span className={styles.tooltipRowValue}>
              {formatDateRange(tooltip.booking.startDate, tooltip.booking.endDate)}
            </span>
          </div>
        </div>
      )}
    </div>
  )
}

export default EquipmentTimeline
