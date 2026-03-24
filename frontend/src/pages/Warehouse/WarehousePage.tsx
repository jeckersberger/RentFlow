import { useState, useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { warehouseApi, projectApi, api } from '../../services/api'
import './Warehouse.scss'

// ============================================================================
// TYPES
// ============================================================================

type WarehouseStatus = 'confirmed' | 'packed' | 'in_transit' | 'on_site' | 'return_expected'
type ViewMode = 'kanban' | 'lagerorte'

interface WarehouseProject {
  id: string
  name: string
  client: string
  status: WarehouseStatus
  start_date: string
  end_date: string
  equipment_count: number
  notes_count: number
  warehouse_location: string
}

interface Warehouse {
  id: string
  tenant_id: string
  name: string
  location: string
  total_capacity_kg: number
  total_capacity_m3: number
  zones: { id: string; name: string; type?: string; capacity_kg: number; capacity_m3: number }[]
  created_at: string
  updated_at: string
}

// Location tree types
interface TreeItem {
  id: string
  name: string
  quantity: number
  equipment_id: string
}

interface TreeBay {
  id: string
  name: string
  item_count: number
  items: TreeItem[]
}

interface TreeRack {
  id: string
  name: string
  item_count: number
  bays: TreeBay[]
}

interface TreeZone {
  id: string
  name: string
  zone_type: string
  item_count: number
  racks: TreeRack[]
}

interface TreeWarehouse {
  id: string
  name: string
  code: string
  item_count: number
  zones: TreeZone[]
}

// ============================================================================
// STATUS CONFIG
// ============================================================================

const STATUS_COLUMNS: { key: WarehouseStatus; label: string; color: string; icon: string }[] = [
  { key: 'confirmed', label: 'Bestaetigt', color: '#3b82f6', icon: '\u2713' },
  { key: 'packed', label: 'Gepackt', color: '#8b5cf6', icon: '\uD83D\uDCE6' },
  { key: 'in_transit', label: 'Unterwegs', color: '#f59e0b', icon: '\uD83D\uDE9B' },
  { key: 'on_site', label: 'Vor Ort', color: '#10b981', icon: '\uD83D\uDCCD' },
  { key: 'return_expected', label: 'Rueckgabe erwartet', color: '#ef4444', icon: '\u21A9' },
]

// ============================================================================
// HELPERS
// ============================================================================

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

function formatDateShort(date: Date): string {
  return date.toLocaleDateString('de-DE', {
    weekday: 'short',
    day: '2-digit',
    month: 'long',
    year: 'numeric',
  })
}

function isSameDay(d1: Date, d2: Date): boolean {
  return d1.getFullYear() === d2.getFullYear() &&
    d1.getMonth() === d2.getMonth() &&
    d1.getDate() === d2.getDate()
}

// ============================================================================
// LOCATION TREE COMPONENT
// ============================================================================

function LocationTreeView({ warehouses }: { warehouses: TreeWarehouse[] }) {
  const navigate = useNavigate()
  const [expanded, setExpanded] = useState<Record<string, boolean>>({})

  const toggle = useCallback((id: string) => {
    setExpanded(prev => ({ ...prev, [id]: !prev[id] }))
  }, [])

  const isExpanded = (id: string) => !!expanded[id]

  if (!warehouses || warehouses.length === 0) {
    return (
      <div className="wh-tree__empty">
        Keine Lagerorte vorhanden
      </div>
    )
  }

  return (
    <div className="wh-tree">
      {warehouses.map(wh => (
        <div key={wh.id} className="wh-tree__warehouse">
          <button
            className={`wh-tree__node wh-tree__node--warehouse ${isExpanded(wh.id) ? 'wh-tree__node--open' : ''}`}
            onClick={() => toggle(wh.id)}
          >
            <span className="wh-tree__toggle">{isExpanded(wh.id) ? '\u25BC' : '\u25B6'}</span>
            <span className="wh-tree__icon wh-tree__icon--warehouse">{'\uD83C\uDFED'}</span>
            <span className="wh-tree__label">{wh.name}</span>
            <span className="wh-tree__count">{wh.item_count} Items</span>
          </button>

          {isExpanded(wh.id) && (
            <div className="wh-tree__children">
              {wh.zones.map(zone => (
                <div key={zone.id} className="wh-tree__zone">
                  <button
                    className={`wh-tree__node wh-tree__node--zone ${isExpanded(zone.id) ? 'wh-tree__node--open' : ''}`}
                    onClick={() => toggle(zone.id)}
                  >
                    <span className="wh-tree__toggle">{isExpanded(zone.id) ? '\u25BC' : '\u25B6'}</span>
                    <span className="wh-tree__icon wh-tree__icon--zone">{'\uD83C\uDFF7\uFE0F'}</span>
                    <span className="wh-tree__label">{zone.name}</span>
                    <span className="wh-tree__count">{zone.item_count} Items</span>
                  </button>

                  {isExpanded(zone.id) && (
                    <div className="wh-tree__children">
                      {zone.racks.map(rack => (
                        <div key={rack.id} className="wh-tree__rack">
                          <button
                            className={`wh-tree__node wh-tree__node--rack ${isExpanded(rack.id) ? 'wh-tree__node--open' : ''}`}
                            onClick={() => toggle(rack.id)}
                          >
                            <span className="wh-tree__toggle">{isExpanded(rack.id) ? '\u25BC' : '\u25B6'}</span>
                            <span className="wh-tree__icon wh-tree__icon--rack">{'\uD83D\uDDC4\uFE0F'}</span>
                            <span className="wh-tree__label">{rack.name}</span>
                            <span className="wh-tree__count">{rack.item_count} Items</span>
                          </button>

                          {isExpanded(rack.id) && (
                            <div className="wh-tree__children">
                              {rack.bays.map(bay => (
                                <div key={bay.id} className="wh-tree__bay">
                                  <button
                                    className={`wh-tree__node wh-tree__node--bay ${isExpanded(bay.id) ? 'wh-tree__node--open' : ''}`}
                                    onClick={() => toggle(bay.id)}
                                  >
                                    <span className="wh-tree__toggle">
                                      {bay.items.length > 0 ? (isExpanded(bay.id) ? '\u25BC' : '\u25B6') : '\u2022'}
                                    </span>
                                    <span className="wh-tree__icon wh-tree__icon--bay">{'\uD83D\uDCE5'}</span>
                                    <span className="wh-tree__label">{bay.name}</span>
                                    <span className="wh-tree__count">{bay.item_count} Items</span>
                                  </button>

                                  {isExpanded(bay.id) && bay.items.length > 0 && (
                                    <div className="wh-tree__children">
                                      {bay.items.map(item => (
                                        <div
                                          key={item.id}
                                          className="wh-tree__node wh-tree__node--item"
                                          onClick={() => navigate(`/equipment/${item.equipment_id}`)}
                                          role="button"
                                          tabIndex={0}
                                          onKeyDown={e => { if (e.key === 'Enter') navigate(`/equipment/${item.equipment_id}`) }}
                                        >
                                          <span className="wh-tree__toggle">{'\u2022'}</span>
                                          <span className="wh-tree__label wh-tree__label--item">
                                            {item.name}
                                          </span>
                                          <span className="wh-tree__quantity">x{item.quantity}</span>
                                        </div>
                                      ))}
                                    </div>
                                  )}
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

// ============================================================================
// MAIN COMPONENT
// ============================================================================

function WarehousePage() {
  const navigate = useNavigate()
  const [viewMode, setViewMode] = useState<ViewMode>('kanban')
  const [selectedDate, setSelectedDate] = useState(new Date())
  const [selectedWarehouseId, setSelectedWarehouseId] = useState<string | null>(null)
  const [draggedCard, setDraggedCard] = useState<string | null>(null)
  const [projectStatuses, setProjectStatuses] = useState<Record<string, WarehouseStatus>>({})

  const today = new Date()
  const tomorrow = new Date(today)
  tomorrow.setDate(tomorrow.getDate() + 1)

  // Fetch warehouses
  const { data: warehousesData, isLoading: isLoadingWarehouses } = useQuery({
    queryKey: ['warehouses'],
    queryFn: () => warehouseApi.listWarehouses(),
  })
  const warehouses: Warehouse[] = warehousesData?.items || []
  const currentWarehouse = warehouses.find(w => w.id === selectedWarehouseId) || warehouses[0] || null

  // Fetch location tree (only when lagerorte view is active)
  const { data: locationTree, isLoading: isLoadingTree } = useQuery({
    queryKey: ['warehouse-location-tree'],
    queryFn: () => warehouseApi.getLocationTree(),
    enabled: viewMode === 'lagerorte',
    staleTime: 1000 * 60 * 5,
  })

  // Fetch projects for the selected date range
  const { data: projectsData, isLoading: isLoadingProjects } = useQuery({
    queryKey: ['warehouse-projects'],
    queryFn: async () => {
      try {
        const res = await projectApi.list(1, 50)
        return res?.items || []
      } catch {
        return []
      }
    },
    staleTime: 1000 * 60 * 2,
  })

  // Fetch equipment count
  const { data: equipmentData } = useQuery({
    queryKey: ['warehouse-equipment-count'],
    queryFn: async () => {
      try {
        const res = await api.get('/api/v1/equipment')
        const equipment = res.data?.data || res.data?.items || res.data || []
        return Array.isArray(equipment) ? equipment : []
      } catch {
        return []
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  // Map projects to warehouse view with statuses
  const warehouseProjects: WarehouseProject[] = useMemo(() => {
    const projects = projectsData || []
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return projects.map((p: any) => {
      // Assign warehouse status based on project status or override
      let warehouseStatus: WarehouseStatus = 'confirmed'
      if (projectStatuses[p.id]) {
        warehouseStatus = projectStatuses[p.id]
      } else if (p.status === 'completed') {
        warehouseStatus = 'return_expected'
      } else if (p.status === 'active') {
        warehouseStatus = 'on_site'
      } else if (p.status === 'planning') {
        warehouseStatus = 'confirmed'
      }

      const equipmentCount = equipmentData
        ? Math.floor(Math.random() * 15) + 3 // demo: random equipment count per project
        : 0

      return {
        id: p.id,
        name: p.name,
        client: p.client || 'Unbekannt',
        status: warehouseStatus,
        start_date: p.start_date,
        end_date: p.end_date,
        equipment_count: equipmentCount,
        notes_count: Math.floor(Math.random() * 5),
        warehouse_location: currentWarehouse?.name || 'Hauptlager',
      }
    })
  }, [projectsData, projectStatuses, equipmentData, currentWarehouse])

  // Filter projects relevant to the selected date
  const filteredProjects = useMemo(() => {
    return warehouseProjects.filter(p => {
      const start = new Date(p.start_date)
      const end = new Date(p.end_date)
      // Show projects that overlap with selected date (within 7 days window)
      const windowStart = new Date(selectedDate)
      windowStart.setDate(windowStart.getDate() - 3)
      const windowEnd = new Date(selectedDate)
      windowEnd.setDate(windowEnd.getDate() + 7)
      return start <= windowEnd && end >= windowStart
    })
  }, [warehouseProjects, selectedDate])

  // Group by status
  const projectsByStatus = useMemo(() => {
    const grouped: Record<WarehouseStatus, WarehouseProject[]> = {
      confirmed: [],
      packed: [],
      in_transit: [],
      on_site: [],
      return_expected: [],
    }
    filteredProjects.forEach(p => {
      grouped[p.status].push(p)
    })
    return grouped
  }, [filteredProjects])

  // Date navigation
  const goToDate = (date: Date) => setSelectedDate(new Date(date))
  const goToToday = () => goToDate(new Date())
  const goToTomorrow = () => {
    const t = new Date()
    t.setDate(t.getDate() + 1)
    goToDate(t)
  }
  const goPrev = () => {
    const d = new Date(selectedDate)
    d.setDate(d.getDate() - 1)
    goToDate(d)
  }
  const goNext = () => {
    const d = new Date(selectedDate)
    d.setDate(d.getDate() + 1)
    goToDate(d)
  }

  // Status change
  const changeStatus = (projectId: string, newStatus: WarehouseStatus) => {
    setProjectStatuses(prev => ({ ...prev, [projectId]: newStatus }))
  }

  // Drag and drop
  const handleDragStart = (projectId: string) => {
    setDraggedCard(projectId)
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
  }

  const handleDrop = (status: WarehouseStatus) => {
    if (draggedCard) {
      changeStatus(draggedCard, status)
      setDraggedCard(null)
    }
  }

  // Get next status for a project
  const getNextStatus = (current: WarehouseStatus): WarehouseStatus | null => {
    const order: WarehouseStatus[] = ['confirmed', 'packed', 'in_transit', 'on_site', 'return_expected']
    const idx = order.indexOf(current)
    return idx < order.length - 1 ? order[idx + 1] : null
  }

  // Loading
  if (isLoadingWarehouses) {
    return (
      <div className="wh-page">
        <div className="wh-page__loading">
          <div className="wh-page__spinner" />
          <span>Lager wird geladen...</span>
        </div>
      </div>
    )
  }

  return (
    <div className="wh-page">
      {/* ====== HEADER ====== */}
      <div className="wh-header">
        <div className="wh-header__left">
          <h1 className="wh-header__title">Lager</h1>
          <p className="wh-header__subtitle">
            {currentWarehouse ? currentWarehouse.name : 'Lagerverwaltung'}
          </p>
        </div>
        <div className="wh-header__actions">
          {/* View Mode Toggle */}
          <div className="wh-view-toggle">
            <button
              className={`wh-view-toggle__btn ${viewMode === 'kanban' ? 'wh-view-toggle__btn--active' : ''}`}
              onClick={() => setViewMode('kanban')}
            >
              Kanban
            </button>
            <button
              className={`wh-view-toggle__btn ${viewMode === 'lagerorte' ? 'wh-view-toggle__btn--active' : ''}`}
              onClick={() => setViewMode('lagerorte')}
            >
              Lagerorte
            </button>
          </div>

          {warehouses.length > 1 && (
            <select
              className="wh-select"
              value={currentWarehouse?.id || ''}
              onChange={e => setSelectedWarehouseId(e.target.value)}
            >
              {warehouses.map(wh => (
                <option key={wh.id} value={wh.id}>{wh.name}</option>
              ))}
            </select>
          )}
          <button
            className="wh-btn wh-btn--accent"
            onClick={() => navigate('/inventory')}
          >
            Inventur starten
          </button>
          <button
            className="wh-btn wh-btn--accent"
            onClick={() => navigate('/scanner')}
          >
            <span className="wh-btn__icon">{'\u21A9'}</span>
            Retour scannen
          </button>
        </div>
      </div>

      {viewMode === 'kanban' ? (
        <>
          {/* ====== DATE NAVIGATION ====== */}
          <div className="wh-date-nav">
            <div className="wh-date-nav__buttons">
              <button className="wh-date-btn" onClick={goPrev}>{'\u2039'}</button>
              <button
                className={`wh-date-btn ${isSameDay(selectedDate, today) ? 'wh-date-btn--active' : ''}`}
                onClick={goToToday}
              >
                Heute
              </button>
              <button
                className={`wh-date-btn ${isSameDay(selectedDate, tomorrow) ? 'wh-date-btn--active' : ''}`}
                onClick={goToTomorrow}
              >
                Morgen
              </button>
              <button className="wh-date-btn" onClick={goNext}>{'\u203A'}</button>
            </div>
            <div className="wh-date-nav__current">
              {formatDateShort(selectedDate)}
            </div>
            <div className="wh-date-nav__stats">
              <span className="wh-date-stat">
                {filteredProjects.length} Projekte
              </span>
              <span className="wh-date-stat">
                {filteredProjects.reduce((sum, p) => sum + p.equipment_count, 0)} Teile
              </span>
            </div>
          </div>

          {/* ====== KANBAN BOARD ====== */}
          <div className="wh-kanban">
            {STATUS_COLUMNS.map(col => (
              <div
                key={col.key}
                className={`wh-kanban__column ${draggedCard ? 'wh-kanban__column--droppable' : ''}`}
                onDragOver={handleDragOver}
                onDrop={() => handleDrop(col.key)}
              >
                <div className="wh-kanban__column-header" style={{ '--col-color': col.color } as React.CSSProperties}>
                  <div className="wh-kanban__column-title">
                    <span className="wh-kanban__column-icon">{col.icon}</span>
                    <span>{col.label}</span>
                  </div>
                  <span className="wh-kanban__column-count">
                    {projectsByStatus[col.key].length}
                  </span>
                </div>

                <div className="wh-kanban__cards">
                  {isLoadingProjects ? (
                    <div className="wh-kanban__loading">Laden...</div>
                  ) : projectsByStatus[col.key].length === 0 ? (
                    <div className="wh-kanban__empty">
                      Keine Projekte
                    </div>
                  ) : (
                    projectsByStatus[col.key].map(project => {
                      const nextStatus = getNextStatus(project.status)
                      const nextLabel = nextStatus
                        ? STATUS_COLUMNS.find(c => c.key === nextStatus)?.label
                        : null

                      return (
                        <div
                          key={project.id}
                          className="wh-card"
                          draggable
                          onDragStart={() => handleDragStart(project.id)}
                          onClick={() => navigate(`/projects/${project.id}`)}
                          style={{ '--card-accent': col.color } as React.CSSProperties}
                        >
                          <div className="wh-card__header">
                            <h3 className="wh-card__name">{project.name}</h3>
                            <span className="wh-card__badge" style={{ backgroundColor: `${col.color}22`, color: col.color }}>
                              {col.label}
                            </span>
                          </div>

                          <div className="wh-card__client">{project.client}</div>

                          <div className="wh-card__meta">
                            <span className="wh-card__date">
                              {formatDate(project.start_date)} - {formatDate(project.end_date)}
                            </span>
                          </div>

                          <div className="wh-card__footer">
                            <div className="wh-card__stats">
                              <span className="wh-card__stat" title="Equipment">
                                {'\uD83D\uDCE6'} {project.equipment_count}
                              </span>
                              {project.notes_count > 0 && (
                                <span className="wh-card__stat" title="Notizen">
                                  {'\uD83D\uDCAC'} {project.notes_count}
                                </span>
                              )}
                            </div>

                            <div className="wh-card__actions" onClick={e => e.stopPropagation()}>
                              {nextStatus && (
                                <button
                                  className="wh-card__action-btn"
                                  onClick={() => changeStatus(project.id, nextStatus)}
                                  title={`Status aendern zu: ${nextLabel}`}
                                >
                                  {'\u2192'} {nextLabel}
                                </button>
                              )}
                            </div>
                          </div>
                        </div>
                      )
                    })
                  )}
                </div>
              </div>
            ))}
          </div>
        </>
      ) : (
        /* ====== LAGERORTE TREE VIEW ====== */
        <div className="wh-tree-container">
          {isLoadingTree ? (
            <div className="wh-page__loading">
              <div className="wh-page__spinner" />
              <span>Lagerorte werden geladen...</span>
            </div>
          ) : (
            <LocationTreeView warehouses={(locationTree || []) as TreeWarehouse[]} />
          )}
        </div>
      )}
    </div>
  )
}

export default WarehousePage
