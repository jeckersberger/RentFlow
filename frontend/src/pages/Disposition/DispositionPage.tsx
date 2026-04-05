import { useState, useRef, useMemo, useCallback, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { equipmentApi, projectApi, categoryApi } from '../../services/api'
import { Equipment, Category } from '../../types/equipment'
import { Project } from '../../types/project'
import {
  ViewMode, Allocation, TooltipData,
  DAY_NAMES_DE, MONTH_NAMES_DE, MONTH_SHORT_DE,
  startOfDay, addDays, isSameDay, isWeekend, getMonday, getColorForProject,
  buildAllocations, detectConflicts,
} from './disposition-utils'
import './Disposition.scss'

function DispositionPage() {
  const navigate = useNavigate()
  const gridBodyRef = useRef<HTMLDivElement>(null)
  const sidebarBodyRef = useRef<HTMLDivElement>(null)
  const gridHeaderRef = useRef<HTMLDivElement>(null)

  const [viewMode, setViewMode] = useState<ViewMode>('week')
  const [startDate, setStartDate] = useState(() => getMonday(new Date()))
  const [tooltip, setTooltip] = useState<TooltipData | null>(null)
  const [searchTerm, setSearchTerm] = useState('')

  const today = useMemo(() => startOfDay(new Date()), [])
  const visibleDays = viewMode === 'week' ? 28 : 56
  const cellWidth = viewMode === 'week' ? 42 : 20

  const dates = useMemo(() => {
    const result: Date[] = []
    for (let i = 0; i < visibleDays; i++) result.push(addDays(startDate, i))
    return result
  }, [startDate, visibleDays])

  // ---- Data fetching ----

  const { data: equipmentData, isLoading: eqLoading } = useQuery({
    queryKey: ['disposition-equipment'],
    queryFn: () => equipmentApi.list({ limit: 500, offset: 0 }),
    staleTime: 1000 * 60 * 5,
  })

  const { data: projectData, isLoading: projLoading } = useQuery({
    queryKey: ['disposition-projects'],
    queryFn: () => projectApi.list(1, 200),
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

  // ---- Group equipment by category, with search filter ----

  const categoryMap = useMemo(() => {
    const map = new Map<string, Category>()
    ;(categories || []).forEach((c: Category) => map.set(c.id, c))
    return map
  }, [categories])

  const filteredEquipment = useMemo(() => {
    if (!searchTerm.trim()) return equipment
    const term = searchTerm.toLowerCase()
    return equipment.filter(eq => eq.name.toLowerCase().includes(term))
  }, [equipment, searchTerm])

  const flatRows = useMemo(() => {
    const groups = new Map<string, { category: Category | null; items: Equipment[] }>()
    filteredEquipment.forEach(eq => {
      const catId = eq.category_id || 'uncategorized'
      if (!groups.has(catId)) groups.set(catId, { category: categoryMap.get(catId) || null, items: [] })
      groups.get(catId)!.items.push(eq)
    })
    const sorted = Array.from(groups.entries()).sort((a, b) =>
      (a[1].category?.name || 'Sonstige').localeCompare(b[1].category?.name || 'Sonstige')
    )
    const rows: Array<
      | { type: 'category'; catId: string; name: string; count: number }
      | { type: 'equipment'; eq: Equipment }
    > = []
    sorted.forEach(([catId, group]) => {
      rows.push({ type: 'category', catId, name: group.category?.name || 'Sonstige', count: group.items.length })
      group.items.forEach(eq => rows.push({ type: 'equipment', eq }))
    })
    return rows
  }, [filteredEquipment, categoryMap])

  // ---- Allocations and conflicts ----

  const allocations = useMemo(() => {
    if (equipment.length === 0 || projects.length === 0) return []
    return buildAllocations(equipment, projects)
  }, [equipment, projects])

  const allocationsByEquipment = useMemo(() => {
    const map = new Map<string, Allocation[]>()
    allocations.forEach(a => {
      if (!map.has(a.equipmentId)) map.set(a.equipmentId, [])
      map.get(a.equipmentId)!.push(a)
    })
    return map
  }, [allocations])

  const conflicts = useMemo(() => detectConflicts(allocations), [allocations])

  const conflictAllocIds = useMemo(() => {
    const ids = new Set<string>()
    conflicts.forEach(c => c.allocations.forEach(a => ids.add(a.id)))
    return ids
  }, [conflicts])

  const stats = useMemo(() => {
    const allocatedIds = new Set(allocations.map(a => a.equipmentId))
    return {
      totalTypes: equipment.length,
      allocated: allocatedIds.size,
      available: equipment.length - allocatedIds.size,
      conflictCount: conflicts.size,
    }
  }, [equipment, allocations, conflicts])

  // ---- Sync scroll between sidebar + grid ----

  const handleGridScroll = useCallback(() => {
    const g = gridBodyRef.current, s = sidebarBodyRef.current, h = gridHeaderRef.current
    if (g && s) s.scrollTop = g.scrollTop
    if (g && h) h.scrollLeft = g.scrollLeft
  }, [])

  useEffect(() => {
    const g = gridBodyRef.current
    if (g) {
      g.addEventListener('scroll', handleGridScroll, { passive: true })
      return () => g.removeEventListener('scroll', handleGridScroll)
    }
  }, [handleGridScroll])

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
  const goToday = () => setStartDate(getMonday(new Date()))
  const goPrev = () => setStartDate(prev => addDays(prev, viewMode === 'week' ? -7 : -14))
  const goNext = () => setStartDate(prev => addDays(prev, viewMode === 'week' ? 7 : 14))

  // ---- Bar position calculation ----

  const getBarStyle = useCallback((alloc: Allocation): React.CSSProperties | null => {
    const tStart = startDate.getTime()
    const tEnd = addDays(startDate, visibleDays).getTime()
    const bStart = Math.max(alloc.startDate.getTime(), tStart)
    const bEnd = Math.min(alloc.endDate.getTime() + 86400000, tEnd)
    if (bEnd <= tStart || bStart >= tEnd) return null
    const ms = 86400000
    return {
      left: `${((bStart - tStart) / ms) * cellWidth}px`,
      width: `${Math.max(((bEnd - bStart) / ms) * cellWidth - 2, 4)}px`,
      backgroundColor: alloc.color,
      borderColor: alloc.color,
    }
  }, [startDate, visibleDays, cellWidth])

  // ---- Tooltip handlers ----

  const handleBarHover = useCallback((e: React.MouseEvent, alloc: Allocation) => {
    const c = conflicts.get(alloc.equipmentId)
    setTooltip({
      allocation: alloc,
      conflict: c?.allocations.filter(a => a.id !== alloc.id),
      x: e.clientX + 12, y: e.clientY - 10,
    })
  }, [conflicts])

  const handleBarMove = useCallback((e: React.MouseEvent) => {
    setTooltip(prev => prev ? { ...prev, x: e.clientX + 12, y: e.clientY - 10 } : null)
  }, [])

  const handleBarLeave = useCallback(() => setTooltip(null), [])

  // ---- Today line ----
  const todayLineLeft = useMemo(() => {
    const diff = (today.getTime() - startDate.getTime()) / 86400000
    if (diff < 0 || diff >= visibleDays) return null
    return diff * cellWidth + cellWidth / 2
  }, [today, startDate, visibleDays, cellWidth])

  // ---- Date label ----
  const endDate = dates[dates.length - 1]
  const dateLabel = useMemo(() => {
    const s = `${MONTH_NAMES_DE[startDate.getMonth()]} ${startDate.getFullYear()}`
    if (startDate.getMonth() !== endDate.getMonth()) {
      return `${s} - ${MONTH_SHORT_DE[endDate.getMonth()]} ${endDate.getFullYear()}`
    }
    return s
  }, [startDate, endDate])

  // ============================================================================
  // Render
  // ============================================================================

  return (
    <div className="dispositionPage">
      <div className="dispoHeader">
        <div className="dispoHeaderLeft">
          <h1 className="dispoTitle">Disposition</h1>
          <p className="dispoSubtitle">Equipment-Zuordnung und Projektzeitplanung im Gantt-Format</p>
        </div>
        <div className="dispoHeaderActions">
          <button className="dispoBtn dispoBtnSecondary" onClick={() => navigate('/equipment/timeline')}>Equipment-Zeitleiste</button>
          <button className="dispoBtn dispoBtnPrimary" onClick={() => navigate('/projects/new')}>+ Neues Projekt</button>
        </div>
      </div>

      <div className="dispoStats">
        <div className="dispoStatCard"><span className="dispoStatValue">{stats.totalTypes}</span><span className="dispoStatLabel">Equipment</span></div>
        <div className="dispoStatCard"><span className="dispoStatValue dispoStatAllocated">{stats.allocated}</span><span className="dispoStatLabel">Zugeordnet</span></div>
        <div className="dispoStatCard"><span className="dispoStatValue dispoStatAvailable">{stats.available}</span><span className="dispoStatLabel">Verfuegbar</span></div>
        <div className={`dispoStatCard ${stats.conflictCount > 0 ? 'dispoStatCardConflict' : ''}`}><span className="dispoStatValue dispoStatConflict">{stats.conflictCount}</span><span className="dispoStatLabel">Konflikte</span></div>
      </div>

      <div className="dispoToolbar">
        <div className="dispoDateNav">
          <button className="dispoNavBtn" onClick={goPrev} title="Zurueck">&#9664;</button>
          <button className="dispoNavBtn dispoNavBtnActive" onClick={goToday}>Heute</button>
          <button className="dispoNavBtn" onClick={goNext} title="Vorwaerts">&#9654;</button>
          <span className="dispoDateDisplay">{dateLabel}</span>
        </div>
        <div className="dispoSearchWrap">
          <input type="text" className="dispoSearch" placeholder="Equipment filtern..." value={searchTerm} onChange={e => setSearchTerm(e.target.value)} />
          {searchTerm && <button className="dispoSearchClear" onClick={() => setSearchTerm('')}>&#10005;</button>}
        </div>
        <div className="dispoViewToggle">
          <button className={`dispoViewBtn ${viewMode === 'week' ? 'dispoViewBtnActive' : ''}`} onClick={() => setViewMode('week')}>Wochen</button>
          <button className={`dispoViewBtn ${viewMode === 'month' ? 'dispoViewBtnActive' : ''}`} onClick={() => setViewMode('month')}>Monate</button>
        </div>
      </div>

      {isLoading ? (
        <div className="dispoLoading"><div className="dispoSpinner" /><p>Disposition wird geladen...</p></div>
      ) : equipment.length === 0 ? (
        <div className="dispoEmpty">
          <div className="dispoEmptyIcon">&#128230;</div>
          <div className="dispoEmptyText">Noch kein Equipment vorhanden</div>
          <button className="dispoBtn dispoBtnPrimary" onClick={() => navigate('/equipment/new')}>Erstes Equipment anlegen</button>
        </div>
      ) : (
        <div className="dispoTimeline">
          <div className="dispoSidebar">
            <div className="dispoSidebarHeader">Equipment</div>
            <div className="dispoSidebarBody" ref={sidebarBodyRef}>
              {flatRows.map(row => {
                if (row.type === 'category') {
                  return (<div key={`cat-${row.catId}`} className="dispoCatHeader"><span className="dispoCatName">{row.name}</span><span className="dispoCatCount">{row.count}</span></div>)
                }
                const hasConflict = conflicts.has(row.eq.id)
                return (
                  <div key={`eq-${row.eq.id}`} className={`dispoEqRow ${hasConflict ? 'dispoEqRowConflict' : ''}`} onClick={() => navigate(`/equipment/${row.eq.id}`)} title={row.eq.name}>
                    <span className="dispoEqName">{row.eq.name}</span>
                    {hasConflict && <span className="dispoEqConflictDot" />}
                  </div>
                )
              })}
            </div>
          </div>

          <div className="dispoGrid">
            <div className="dispoGridHeader" ref={gridHeaderRef}>
              <div className="dispoGridHeaderInner">
                {dates.map((date, idx) => (
                  <div key={idx} className={`dispoDateCol ${isWeekend(date) ? 'dispoDateColWeekend' : ''} ${isSameDay(date, today) ? 'dispoDateColToday' : ''}`} style={{ width: cellWidth }}>
                    <span className="dispoDayName">{DAY_NAMES_DE[date.getDay()]}</span>
                    <span className="dispoDayNum">{date.getDate()}</span>
                    {(idx === 0 || date.getDate() === 1) && <span className="dispoDayMonth">{MONTH_SHORT_DE[date.getMonth()]}</span>}
                  </div>
                ))}
              </div>
            </div>

            <div className="dispoGridBody" ref={gridBodyRef}>
              <div className="dispoGridBodyInner" style={{ width: visibleDays * cellWidth }}>
                {todayLineLeft !== null && <div className="dispoTodayLine" style={{ left: todayLineLeft }} />}
                {flatRows.map(row => {
                  if (row.type === 'category') {
                    return (<div key={`catrow-${row.catId}`} className="dispoCatRow">{dates.map((date, dIdx) => (<div key={dIdx} className={`dispoCell ${isWeekend(date) ? 'dispoCellWeekend' : ''}`} style={{ width: cellWidth, height: 28 }} />))}</div>)
                  }
                  const eqAllocs = allocationsByEquipment.get(row.eq.id) || []
                  return (
                    <div key={`row-${row.eq.id}`} className={`dispoRow ${conflicts.has(row.eq.id) ? 'dispoRowConflict' : ''}`}>
                      {dates.map((date, dIdx) => (<div key={dIdx} className={`dispoCell ${isWeekend(date) ? 'dispoCellWeekend' : ''} ${isSameDay(date, today) ? 'dispoCellToday' : ''}`} style={{ width: cellWidth, height: 40 }} />))}
                      {eqAllocs.map(alloc => {
                        const s = getBarStyle(alloc)
                        if (!s) return null
                        return (
                          <div key={alloc.id} className={`dispoBar ${conflictAllocIds.has(alloc.id) ? 'dispoBarConflict' : ''}`} style={s}
                            onClick={e => { e.stopPropagation(); navigate(`/projects/${alloc.projectId}`) }}
                            onMouseEnter={e => handleBarHover(e, alloc)} onMouseMove={handleBarMove} onMouseLeave={handleBarLeave}>
                            <span className="dispoBarLabel">{alloc.projectName}</span>
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

      {projects.length > 0 && (
        <div className="dispoLegend">
          <span className="dispoLegendTitle">Projekte:</span>
          {projects.filter(p => p.status !== 'cancelled' && p.start_date && p.end_date).slice(0, 12).map(p => (
            <div key={p.id} className="dispoLegendItem" onClick={() => navigate(`/projects/${p.id}`)}>
              <span className="dispoLegendDot" style={{ backgroundColor: getColorForProject(p.id) }} />
              <span className="dispoLegendName">{p.name}</span>
            </div>
          ))}
        </div>
      )}

      {tooltip && (
        <div className="dispoTooltip" style={{ left: tooltip.x, top: tooltip.y }}>
          <div className="dispoTooltipProject" style={{ borderLeftColor: tooltip.allocation.color }}>
            <strong>{tooltip.allocation.projectName}</strong>
            {tooltip.allocation.clientName && <span className="dispoTooltipClient">{tooltip.allocation.clientName}</span>}
          </div>
          <div className="dispoTooltipDates">
            {tooltip.allocation.startDate.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })}
            {' - '}
            {tooltip.allocation.endDate.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })}
          </div>
          {tooltip.conflict && tooltip.conflict.length > 0 && (
            <div className="dispoTooltipConflict">
              <span className="dispoTooltipConflictLabel">Konflikt mit:</span>
              {tooltip.conflict.map(c => <span key={c.id} className="dispoTooltipConflictName">{c.projectName}</span>)}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default DispositionPage
