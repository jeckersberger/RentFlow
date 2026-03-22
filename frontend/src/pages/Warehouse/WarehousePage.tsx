import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Input } from '../../components/Form/Input'
import '../Equipment/Equipment.module.scss'
import './Warehouse.module.scss'

// Phase 2: Extended data structures for hierarchical warehouse
interface Zone {
  id: string
  name: string
  rack_count: number
  bay_count: number
  total_bays: number
  occupancy_percentage: number
  equipment_count: number
  capacity_max: number
  isExpanded?: boolean
}

// Rack interface reserved for detailed rack view expansion
// interface Rack { id: string; zone_id: string; name: string; bay_count: number; ... }

interface MovementEvent {
  id: string
  equipment_name: string
  from_location: string
  to_location: string
  timestamp: string
  user: string
  status: 'completed' | 'pending'
}

interface WarehouseData {
  warehouse_id: string
  warehouse_name: string
  zones: Zone[]
  recent_movements: MovementEvent[]
  total_equipment: number
  total_occupancy: number
}

// Mock data for Phase 2
const mockWarehouseData: WarehouseData = {
  warehouse_id: 'wh-1',
  warehouse_name: 'Hauptlager München',
  total_equipment: 156,
  total_occupancy: 68,
  zones: [
    {
      id: 'zone-1',
      name: 'Zone A - Beleuchtung',
      rack_count: 3,
      bay_count: 24,
      total_bays: 36,
      occupancy_percentage: 67,
      equipment_count: 48,
      capacity_max: 72,
      isExpanded: true,
    },
    {
      id: 'zone-2',
      name: 'Zone B - Ton',
      rack_count: 2,
      bay_count: 18,
      total_bays: 30,
      occupancy_percentage: 60,
      equipment_count: 52,
      capacity_max: 90,
      isExpanded: true,
    },
    {
      id: 'zone-3',
      name: 'Zone C - Video & Dekoration',
      rack_count: 4,
      bay_count: 28,
      total_bays: 44,
      occupancy_percentage: 64,
      equipment_count: 56,
      capacity_max: 88,
      isExpanded: false,
    },
  ],
  recent_movements: [
    {
      id: 'mv-1',
      equipment_name: 'JBL VTX A12',
      from_location: 'Zone A - Rack 1',
      to_location: 'Zone A - Rack 2',
      timestamp: '2026-03-22T10:45:00Z',
      user: 'Marco Berger',
      status: 'completed',
    },
    {
      id: 'mv-2',
      equipment_name: 'Shure SM58',
      from_location: 'Zone B - Rack 3',
      to_location: 'Projekt: Stadtfest München',
      timestamp: '2026-03-22T09:30:00Z',
      user: 'Anna Schmidt',
      status: 'completed',
    },
    {
      id: 'mv-3',
      equipment_name: 'MA Lighting grandMA3',
      from_location: 'Projekt: Stadtfest München',
      to_location: 'Zone A - Rack 1',
      timestamp: '2026-03-21T17:15:00Z',
      user: 'Marco Berger',
      status: 'pending',
    },
  ],
}

function WarehousePage() {
  const [expandedZones, setExpandedZones] = useState<Set<string>>(
    new Set(mockWarehouseData.zones.filter(z => z.isExpanded).map(z => z.id))
  )
  const [selectedZone, setSelectedZone] = useState<Zone | null>(mockWarehouseData.zones[0] || null)
  const [searchQuery, setSearchQuery] = useState('')
  const [inventoryCheckInProgress, setInventoryCheckInProgress] = useState(false)

  // Mock query for warehouse data (would use React Query in production)
  const { data: warehouseData = mockWarehouseData, isLoading } = useQuery<WarehouseData>({
    queryKey: ['warehouse'],
    queryFn: async () => {
      // Simulate API call
      return new Promise<WarehouseData>(resolve => {
        setTimeout(() => resolve(mockWarehouseData), 300)
      })
    },
  })

  const filteredZones = useMemo(() => {
    if (!searchQuery) return warehouseData.zones
    return warehouseData.zones.filter(
      zone => zone.name.toLowerCase().includes(searchQuery.toLowerCase())
    )
  }, [warehouseData.zones, searchQuery])

  const toggleZoneExpand = (zoneId: string) => {
    const newExpanded = new Set(expandedZones)
    if (newExpanded.has(zoneId)) {
      newExpanded.delete(zoneId)
    } else {
      newExpanded.add(zoneId)
    }
    setExpandedZones(newExpanded)
  }

  const getOccupancyColor = (percentage: number) => {
    if (percentage < 50) return 'var(--color-success)'
    if (percentage < 80) return 'var(--color-warning)'
    return 'var(--color-danger)'
  }

  const handleStartInventoryCheck = () => {
    setInventoryCheckInProgress(true)
    setTimeout(() => {
      setInventoryCheckInProgress(false)
      // Show success notification
    }, 1000)
  }

  if (isLoading) {
    return (
      <div className="warehouse-page">
        <div className="page-header">
          <h1 className="page-title">Lagerbestandsverwaltung</h1>
        </div>
        <div style={{ textAlign: 'center', padding: 'var(--spacing-8)', color: 'var(--color-text-secondary)' }}>
          Laden...
        </div>
      </div>
    )
  }

  return (
    <div className="warehouse-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Lagerbestandsverwaltung</h1>
          <p className="page-subtitle">
            {warehouseData.warehouse_name} • {warehouseData.total_equipment} Ausrüstungen
          </p>
        </div>
        <button
          className="btn btn--primary"
          onClick={handleStartInventoryCheck}
          disabled={inventoryCheckInProgress}
        >
          {inventoryCheckInProgress ? '⏳ Inventur läuft...' : '📊 Inventur starten'}
        </button>
      </div>

      <div className="warehouse-stats-grid">
        <div className="stat-card">
          <div className="stat-card__label">Gesamtauslastung</div>
          <div className="stat-card__value">{warehouseData.total_occupancy}%</div>
          <div className="stat-card__bar">
            <div
              className="stat-card__bar-fill"
              style={{
                width: `${warehouseData.total_occupancy}%`,
                backgroundColor: getOccupancyColor(warehouseData.total_occupancy),
              }}
            />
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-card__label">Lagerbestände</div>
          <div className="stat-card__value">{warehouseData.zones.length}</div>
          <div className="stat-card__meta">Zonen aktiv</div>
        </div>

        <div className="stat-card">
          <div className="stat-card__label">Verfügbar</div>
          <div className="stat-card__value">{Math.round(warehouseData.total_equipment * 0.65)}</div>
          <div className="stat-card__meta">von {warehouseData.total_equipment}</div>
        </div>
      </div>

      <div className="warehouse-container">
        <div className="warehouse-zones">
          <div className="warehouse-zones__header">
            <h2 className="warehouse-zones__title">Lagerzonen</h2>
            <Input
              type="text"
              placeholder="Suche nach Zone..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{ maxWidth: '250px' }}
            />
          </div>

          <div className="warehouse-zones__list">
            {filteredZones.map((zone) => (
              <div key={zone.id} className="zone-card">
                <div
                  className="zone-card__header"
                  onClick={() => {
                    setSelectedZone(zone)
                    toggleZoneExpand(zone.id)
                  }}
                  style={{ cursor: 'pointer' }}
                >
                  <div className="zone-card__title-section">
                    <span className="zone-card__expand">
                      {expandedZones.has(zone.id) ? '▼' : '▶'}
                    </span>
                    <h3 className="zone-card__title">{zone.name}</h3>
                  </div>
                  <div className="zone-card__badges">
                    <span className="badge badge--info">{zone.rack_count} Regale</span>
                    <span className="badge badge--default">{zone.bay_count}/{zone.total_bays} Fächer</span>
                  </div>
                </div>

                <div className="zone-card__stats">
                  <div className="zone-stat">
                    <span className="zone-stat__label">Auslastung</span>
                    <div className="zone-stat__bar">
                      <div
                        className="zone-stat__bar-fill"
                        style={{
                          width: `${zone.occupancy_percentage}%`,
                          backgroundColor: getOccupancyColor(zone.occupancy_percentage),
                        }}
                      />
                    </div>
                    <span className="zone-stat__value">{zone.occupancy_percentage}%</span>
                  </div>

                  <div className="zone-stat">
                    <span className="zone-stat__label">Ausrüstung</span>
                    <span className="zone-stat__value">{zone.equipment_count} Stück</span>
                  </div>
                </div>

                {expandedZones.has(zone.id) && (
                  <div className="zone-card__details">
                    <div className="rack-grid">
                      {Array.from({ length: zone.rack_count }, (_, i) => (
                        <div key={i} className="rack-item">
                          <div className="rack-item__header">Regal {String.fromCharCode(65 + i)}</div>
                          <div className="rack-item__bays">
                            {Array.from({ length: Math.min(4, zone.total_bays / zone.rack_count) }, (_, j) => (
                              <div
                                key={j}
                                className="bay"
                                style={{
                                  backgroundColor:
                                    Math.random() > 0.3
                                      ? 'var(--color-primary-800)'
                                      : 'var(--color-bg-tertiary)',
                                }}
                              >
                                {Math.random() > 0.3 && '📦'}
                              </div>
                            ))}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {selectedZone && (
          <div className="warehouse-detail-panel">
            <div className="detail-card">
              <h2 className="detail-card__title">📍 {selectedZone.name}</h2>

              <div className="detail-section">
                <h3 className="detail-section__title">Kapazität</h3>
                <div className="detail-row">
                  <span className="detail-label">Fächer</span>
                  <span className="detail-value">
                    {selectedZone.bay_count}/{selectedZone.total_bays}
                  </span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Ausrüstung</span>
                  <span className="detail-value">{selectedZone.equipment_count}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Auslastung</span>
                  <span className="detail-value">{selectedZone.occupancy_percentage}%</span>
                </div>
              </div>

              <div className="detail-section">
                <h3 className="detail-section__title">Regale</h3>
                <div className="rack-list">
                  {Array.from({ length: selectedZone.rack_count }, (_, i) => (
                    <div key={i} className="rack-list-item">
                      <span>Regal {String.fromCharCode(65 + i)}</span>
                      <span className="rack-list-item__count">
                        {Math.floor(selectedZone.equipment_count / selectedZone.rack_count)} Stück
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {warehouseData.recent_movements.length > 0 && (
        <div className="movement-history">
          <h2 className="movement-history__title">Bewegungsverlauf (letzte Transaktionen)</h2>
          <div className="movement-list">
            {warehouseData.recent_movements.map((movement) => (
              <div key={movement.id} className="movement-item">
                <div className="movement-item__icon">
                  {movement.status === 'pending' ? '⏳' : '✓'}
                </div>
                <div className="movement-item__content">
                  <p className="movement-item__title">{movement.equipment_name}</p>
                  <p className="movement-item__path">
                    {movement.from_location} → {movement.to_location}
                  </p>
                </div>
                <div className="movement-item__meta">
                  <p className="movement-item__user">{movement.user}</p>
                  <p className="movement-item__time">
                    {new Date(movement.timestamp).toLocaleTimeString('de-DE', {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export default WarehousePage
