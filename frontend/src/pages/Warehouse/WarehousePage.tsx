import { useState, useMemo, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { warehouseApi } from '../../services/api'
import { Input } from '../../components/Form/Input'
import '../Equipment/Equipment.module.scss'
import './Warehouse.module.scss'

interface Zone {
  id: string
  name: string
  type?: string
  capacity_kg: number
  capacity_m3: number
}

interface Rack {
  id: string
  zone_id: string
  name: string
  capacity_kg: number
  capacity_m3: number
  equipment_count: number
}

interface Warehouse {
  id: string
  tenant_id: string
  name: string
  location: string
  total_capacity_kg: number
  total_capacity_m3: number
  zones: Zone[]
  created_at: string
  updated_at: string
}

interface Movement {
  id: string
  equipment_id: string
  from_location: string
  to_location: string
  movement_type: string
  quantity: number
  timestamp: string
  user_id: string
}

function WarehousePage() {
  const queryClient = useQueryClient()
  const [selectedWarehouseId, setSelectedWarehouseId] = useState<string | null>(null)
  const [expandedZones, setExpandedZones] = useState<Set<string>>(new Set())
  const [selectedZone, setSelectedZone] = useState<Zone | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [inventoryCheckInProgress, setInventoryCheckInProgress] = useState(false)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newWarehouse, setNewWarehouse] = useState({ name: '', location: '' })

  // Fetch all warehouses
  const { data: warehousesData, isLoading: isLoadingWarehouses } = useQuery({
    queryKey: ['warehouses'],
    queryFn: () => warehouseApi.listWarehouses(),
  })

  const warehouses: Warehouse[] = warehousesData?.items || []
  const currentWarehouse = warehouses.find(w => w.id === selectedWarehouseId) || warehouses[0] || null

  // Auto-select first warehouse
  if (currentWarehouse && !selectedWarehouseId) {
    // Will be set on next render via effect-like pattern
  }

  // Fetch zones for the selected warehouse
  const { data: zones = [] } = useQuery<Zone[]>({
    queryKey: ['warehouse-zones', currentWarehouse?.id],
    queryFn: () => warehouseApi.listZones(currentWarehouse!.id),
    enabled: !!currentWarehouse?.id,
  })

  // We'll track racks per zone in state loaded on expand
  const [racksMap, setRacksMap] = useState<Record<string, Rack[]>>({})

  const loadRacksForZone = useCallback(async (zoneId: string) => {
    if (!currentWarehouse) return
    if (racksMap[zoneId]) return // already loaded
    try {
      const racks = await warehouseApi.listRacks(currentWarehouse.id, zoneId)
      setRacksMap(prev => ({ ...prev, [zoneId]: Array.isArray(racks) ? racks : [] }))
    } catch {
      setRacksMap(prev => ({ ...prev, [zoneId]: [] }))
    }
  }, [currentWarehouse, racksMap])

  // Fetch recent movements
  const { data: movementsData } = useQuery({
    queryKey: ['warehouse-movements'],
    queryFn: () => warehouseApi.listMovements(1, 10),
  })
  const movements: Movement[] = movementsData?.items || []

  // Create warehouse mutation
  const createWarehouseMutation = useMutation({
    mutationFn: (data: { name: string; location: string }) =>
      warehouseApi.createWarehouse(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['warehouses'] })
      setShowCreateForm(false)
      setNewWarehouse({ name: '', location: '' })
    },
  })

  // Inventory check
  const handleStartInventoryCheck = async () => {
    if (!currentWarehouse) return
    setInventoryCheckInProgress(true)
    try {
      await warehouseApi.startInventoryCheck(currentWarehouse.id, selectedZone?.id)
    } catch {
      // ignore errors, mock or real
    }
    setTimeout(() => setInventoryCheckInProgress(false), 1000)
  }

  // Use zones from API, fall back to warehouse.zones if API returns empty
  const effectiveZones: Zone[] = zones.length > 0 ? zones : (currentWarehouse?.zones || [])

  const filteredZones = useMemo(() => {
    if (!searchQuery) return effectiveZones
    return effectiveZones.filter(
      zone => zone.name.toLowerCase().includes(searchQuery.toLowerCase())
    )
  }, [effectiveZones, searchQuery])

  const toggleZoneExpand = (zoneId: string) => {
    const newExpanded = new Set(expandedZones)
    if (newExpanded.has(zoneId)) {
      newExpanded.delete(zoneId)
    } else {
      newExpanded.add(zoneId)
      loadRacksForZone(zoneId)
    }
    setExpandedZones(newExpanded)
  }

  const handleCreateWarehouse = () => {
    if (!newWarehouse.name.trim()) return
    createWarehouseMutation.mutate(newWarehouse)
  }

  if (isLoadingWarehouses) {
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

  if (!currentWarehouse) {
    return (
      <div className="warehouse-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Lagerbestandsverwaltung</h1>
            <p className="page-subtitle">Kein Lager vorhanden</p>
          </div>
          <button className="btn btn--primary" onClick={() => setShowCreateForm(true)}>
            + Neues Lager erstellen
          </button>
        </div>
        {showCreateForm && renderCreateForm()}
      </div>
    )
  }

  function renderCreateForm() {
    return (
      <div className="detail-card" style={{ marginTop: 'var(--spacing-4)', padding: 'var(--spacing-6)' }}>
        <h3 className="detail-card__title">Neues Lager erstellen</h3>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)', marginTop: 'var(--spacing-4)' }}>
          <Input
            type="text"
            placeholder="Lagername *"
            value={newWarehouse.name}
            onChange={(e) => setNewWarehouse(prev => ({ ...prev, name: e.target.value }))}
          />
          <Input
            type="text"
            placeholder="Standort / Adresse"
            value={newWarehouse.location}
            onChange={(e) => setNewWarehouse(prev => ({ ...prev, location: e.target.value }))}
          />
          <div style={{ display: 'flex', gap: 'var(--spacing-2)' }}>
            <button
              className="btn btn--primary"
              onClick={handleCreateWarehouse}
              disabled={!newWarehouse.name.trim() || createWarehouseMutation.isPending}
            >
              {createWarehouseMutation.isPending ? 'Erstelle...' : 'Erstellen'}
            </button>
            <button className="btn" onClick={() => setShowCreateForm(false)}>
              Abbrechen
            </button>
          </div>
        </div>
      </div>
    )
  }

  // Calculate total equipment from racks data or estimate from zones
  const totalEquipment = Object.values(racksMap).reduce(
    (sum, racks) => sum + racks.reduce((s, r) => s + (r.equipment_count || 0), 0),
    0
  )

  return (
    <div className="warehouse-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Lagerbestandsverwaltung</h1>
          <p className="page-subtitle">
            {currentWarehouse.name} {currentWarehouse.location && `- ${currentWarehouse.location}`}
            {totalEquipment > 0 && ` | ${totalEquipment} Ausruestungen`}
          </p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-2)' }}>
          <button className="btn" onClick={() => setShowCreateForm(!showCreateForm)}>
            + Neues Lager
          </button>
          <button
            className="btn btn--primary"
            onClick={handleStartInventoryCheck}
            disabled={inventoryCheckInProgress}
          >
            {inventoryCheckInProgress ? 'Inventur laeuft...' : 'Inventur starten'}
          </button>
        </div>
      </div>

      {/* Warehouse selector if multiple warehouses */}
      {warehouses.length > 1 && (
        <div style={{ marginBottom: 'var(--spacing-4)', display: 'flex', gap: 'var(--spacing-2)', flexWrap: 'wrap' }}>
          {warehouses.map(wh => (
            <button
              key={wh.id}
              className={`btn btn--sm ${wh.id === currentWarehouse.id ? 'btn--primary' : ''}`}
              onClick={() => {
                setSelectedWarehouseId(wh.id)
                setExpandedZones(new Set())
                setSelectedZone(null)
                setRacksMap({})
              }}
            >
              {wh.name}
            </button>
          ))}
        </div>
      )}

      {showCreateForm && renderCreateForm()}

      <div className="warehouse-stats-grid">
        <div className="stat-card">
          <div className="stat-card__label">Zonen</div>
          <div className="stat-card__value">{effectiveZones.length}</div>
          <div className="stat-card__meta">aktive Lagerzonen</div>
        </div>

        <div className="stat-card">
          <div className="stat-card__label">Gesamtkapazitaet</div>
          <div className="stat-card__value">{(currentWarehouse.total_capacity_kg / 1000).toFixed(0)}t</div>
          <div className="stat-card__meta">{currentWarehouse.total_capacity_m3} m3</div>
        </div>

        <div className="stat-card">
          <div className="stat-card__label">Bewegungen</div>
          <div className="stat-card__value">{movements.length}</div>
          <div className="stat-card__meta">letzte Transaktionen</div>
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
            {filteredZones.map((zone) => {
              const zoneRacks = racksMap[zone.id] || []
              const zoneEquipmentCount = zoneRacks.reduce((s, r) => s + (r.equipment_count || 0), 0)

              return (
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
                        {expandedZones.has(zone.id) ? '\u25BC' : '\u25B6'}
                      </span>
                      <h3 className="zone-card__title">{zone.name}</h3>
                    </div>
                    <div className="zone-card__badges">
                      <span className="badge badge--info">
                        {zone.capacity_kg ? `${(zone.capacity_kg / 1000).toFixed(0)}t` : zone.type || 'Zone'}
                      </span>
                      {zoneEquipmentCount > 0 && (
                        <span className="badge badge--default">{zoneEquipmentCount} Teile</span>
                      )}
                    </div>
                  </div>

                  <div className="zone-card__stats">
                    <div className="zone-stat">
                      <span className="zone-stat__label">Kapazitaet</span>
                      <span className="zone-stat__value">
                        {zone.capacity_kg ? `${zone.capacity_kg} kg` : '-'} / {zone.capacity_m3 ? `${zone.capacity_m3} m3` : '-'}
                      </span>
                    </div>
                  </div>

                  {/* Expanded: show racks tree */}
                  {expandedZones.has(zone.id) && (
                    <div className="zone-card__details">
                      <div className="rack-grid">
                        {zoneRacks.length === 0 ? (
                          <div style={{ padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.875rem' }}>
                            Keine Regale vorhanden
                          </div>
                        ) : (
                          zoneRacks.map((rack) => (
                            <div key={rack.id} className="rack-item">
                              <div className="rack-item__header">{rack.name}</div>
                              <div className="rack-item__bays">
                                <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>
                                  {rack.equipment_count} Teile | {rack.capacity_kg} kg | {rack.capacity_m3} m3
                                </div>
                              </div>
                            </div>
                          ))
                        )}
                      </div>
                    </div>
                  )}
                </div>
              )
            })}

            {filteredZones.length === 0 && (
              <div style={{ textAlign: 'center', padding: 'var(--spacing-6)', color: 'var(--color-text-secondary)' }}>
                {searchQuery ? 'Keine Zonen gefunden' : 'Keine Zonen vorhanden'}
              </div>
            )}
          </div>
        </div>

        {selectedZone && (
          <div className="warehouse-detail-panel">
            <div className="detail-card">
              <h2 className="detail-card__title">{selectedZone.name}</h2>

              <div className="detail-section">
                <h3 className="detail-section__title">Kapazitaet</h3>
                <div className="detail-row">
                  <span className="detail-label">Gewicht</span>
                  <span className="detail-value">{selectedZone.capacity_kg} kg</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Volumen</span>
                  <span className="detail-value">{selectedZone.capacity_m3} m3</span>
                </div>
              </div>

              {racksMap[selectedZone.id] && racksMap[selectedZone.id].length > 0 && (
                <div className="detail-section">
                  <h3 className="detail-section__title">Regale</h3>
                  <div className="rack-list">
                    {racksMap[selectedZone.id].map((rack) => (
                      <div key={rack.id} className="rack-list-item">
                        <span>{rack.name}</span>
                        <span className="rack-list-item__count">
                          {rack.equipment_count} Teile
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Recent movements */}
      {movements.length > 0 && (
        <div className="movement-history">
          <h2 className="movement-history__title">Bewegungsverlauf (letzte Transaktionen)</h2>
          <div className="movement-list">
            {movements.map((movement) => (
              <div key={movement.id} className="movement-item">
                <div className="movement-item__icon">
                  {movement.movement_type === 'checkout' ? '\u2192' : '\u2190'}
                </div>
                <div className="movement-item__content">
                  <p className="movement-item__title">Equipment #{movement.equipment_id}</p>
                  <p className="movement-item__path">
                    {movement.from_location} \u2192 {movement.to_location}
                  </p>
                </div>
                <div className="movement-item__meta">
                  <p className="movement-item__user">{movement.movement_type} ({movement.quantity}x)</p>
                  <p className="movement-item__time">
                    {new Date(movement.timestamp).toLocaleString('de-DE', {
                      day: '2-digit',
                      month: '2-digit',
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
