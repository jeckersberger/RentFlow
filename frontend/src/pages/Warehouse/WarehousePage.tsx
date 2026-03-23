import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { warehouseApi, projectApi, api } from '../../services/api'
import './Warehouse.scss'

// ============================================================================
// TYPES
// ============================================================================

type WarehouseStatus = 'confirmed' | 'packed' | 'in_transit' | 'on_site' | 'return_expected'

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

// ============================================================================
// STATUS CONFIG
// ============================================================================

const STATUS_COLUMNS: { key: WarehouseStatus; label: string; color: string; icon: string }[] = [
  { key: 'confirmed', label: 'Bestätigt', color: '#3b82f6', icon: '✓' },
  { key: 'packed', label: 'Gepackt', color: '#8b5cf6', icon: '📦' },
  { key: 'in_transit', label: 'Unterwegs', color: '#f59e0b', icon: '🚛' },
  { key: 'on_site', label: 'Vor Ort', color: '#10b981', icon: '📍' },
  { key: 'return_expected', label: 'Rückgabe erwartet', color: '#ef4444', icon: '↩' },
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
// COMPONENT
// ============================================================================

function WarehousePage() {
  const navigate = useNavigate()
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
            onClick={() => navigate('/scanner')}
          >
            <span className="wh-btn__icon">↩</span>
            Retour scannen
          </button>
        </div>
      </div>

      {/* ====== DATE NAVIGATION ====== */}
      <div className="wh-date-nav">
        <div className="wh-date-nav__buttons">
          <button className="wh-date-btn" onClick={goPrev}>‹</button>
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
          <button className="wh-date-btn" onClick={goNext}>›</button>
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
                            📦 {project.equipment_count}
                          </span>
                          {project.notes_count > 0 && (
                            <span className="wh-card__stat" title="Notizen">
                              💬 {project.notes_count}
                            </span>
                          )}
                        </div>

                        <div className="wh-card__actions" onClick={e => e.stopPropagation()}>
                          {nextStatus && (
                            <button
                              className="wh-card__action-btn"
                              onClick={() => changeStatus(project.id, nextStatus)}
                              title={`Status ändern zu: ${nextLabel}`}
                            >
                              → {nextLabel}
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
    </div>
  )
}

export default WarehousePage
