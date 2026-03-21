import { useState } from 'react'
import { Input } from '../../components/Form/Input'
import '../Equipment/Equipment.module.scss'
import './Warehouse.module.scss'

interface WarehouseLocation {
  id: string
  name: string
  type: 'site' | 'room' | 'rack' | 'shelf'
  children?: WarehouseLocation[]
  equipment_count: number
  isExpanded?: boolean
}

const mockWarehouse: WarehouseLocation = {
  id: 'site-1',
  name: 'Hauptlager',
  type: 'site',
  equipment_count: 156,
  children: [
    {
      id: 'room-1',
      name: 'Raum 1 - Beleuchtung',
      type: 'room',
      equipment_count: 48,
      children: [
        { id: 'rack-1', name: 'Rack A', type: 'rack', equipment_count: 24 },
        { id: 'rack-2', name: 'Rack B', type: 'rack', equipment_count: 24 },
      ],
    },
    {
      id: 'room-2',
      name: 'Raum 2 - Ton',
      type: 'room',
      equipment_count: 52,
      children: [
        { id: 'rack-3', name: 'Rack C', type: 'rack', equipment_count: 26 },
        { id: 'rack-4', name: 'Rack D', type: 'rack', equipment_count: 26 },
      ],
    },
    {
      id: 'room-3',
      name: 'Raum 3 - Dekoration',
      type: 'room',
      equipment_count: 56,
      children: [
        { id: 'rack-5', name: 'Rack E', type: 'rack', equipment_count: 28 },
        { id: 'rack-6', name: 'Rack F', type: 'rack', equipment_count: 28 },
      ],
    },
  ],
}

function WarehouseViewPage() {
  const [expandedLocations, setExpandedLocations] = useState<Set<string>>(
    new Set(['site-1', 'room-1', 'room-2', 'room-3'])
  )
  const [selectedLocation, setSelectedLocation] = useState<WarehouseLocation | null>(null)
  const [searchQuery, setSearchQuery] = useState('')

  const toggleLocationExpand = (locationId: string) => {
    const newExpanded = new Set(expandedLocations)
    if (newExpanded.has(locationId)) {
      newExpanded.delete(locationId)
    } else {
      newExpanded.add(locationId)
    }
    setExpandedLocations(newExpanded)
  }

  const getLocationIcon = (type: string) => {
    switch (type) {
      case 'site':
        return '🏢'
      case 'room':
        return '🚪'
      case 'rack':
        return '📦'
      case 'shelf':
        return '📚'
      default:
        return '📍'
    }
  }

  const renderLocation = (location: WarehouseLocation, level = 0) => {
    const isExpanded = expandedLocations.has(location.id)
    const hasChildren = location.children && location.children.length > 0

    return (
      <div key={location.id} style={{ marginLeft: `${level * 1.5}rem` }}>
        <div
          className="warehouse-location"
          onClick={() => {
            setSelectedLocation(location)
            if (hasChildren) toggleLocationExpand(location.id)
          }}
          style={{ cursor: 'pointer' }}
        >
          {hasChildren && (
            <span
              className="warehouse-location__expand"
              style={{
                transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)',
              }}
            >
              ▶
            </span>
          )}
          <span className="warehouse-location__icon">
            {getLocationIcon(location.type)}
          </span>
          <span className="warehouse-location__name">{location.name}</span>
          <span className="warehouse-location__count">
            {location.equipment_count}
          </span>
        </div>

        {hasChildren && isExpanded && (
          <div>
            {location.children!.map((child) =>
              renderLocation(child, level + 1)
            )}
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="warehouse-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Lagerbestandsverwaltung</h1>
          <p className="page-subtitle">
            Verwalten Sie Lagerbereiche und Equipment-Bestände
          </p>
        </div>
      </div>

      <div className="warehouse-container">
        <div className="warehouse-tree">
          <div className="warehouse-tree__header">
            <h2 style={{ margin: 0, fontSize: 'var(--font-size-lg)', fontWeight: 'var(--font-weight-bold)' }}>
              Lagerstruktur
            </h2>
            <Input
              type="text"
              placeholder="Suchen..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{ marginTop: 'var(--spacing-3)' }}
            />
          </div>

          <div className="warehouse-tree__content">
            {renderLocation(mockWarehouse)}
          </div>
        </div>

        {selectedLocation && (
          <div className="warehouse-detail">
            <div className="detail-card">
              <h2 className="detail-card__title">
                {getLocationIcon(selectedLocation.type)} {selectedLocation.name}
              </h2>

              <div className="detail-card__content">
                <div className="detail-card__row">
                  <span className="detail-card__row-label">Typ</span>
                  <span className="detail-card__row-value">
                    {selectedLocation.type === 'site' && 'Lagerstandort'}
                    {selectedLocation.type === 'room' && 'Raum'}
                    {selectedLocation.type === 'rack' && 'Regal'}
                    {selectedLocation.type === 'shelf' && 'Fach'}
                  </span>
                </div>

                <div className="detail-card__row">
                  <span className="detail-card__row-label">Ausrüstung</span>
                  <span className="detail-card__row-value">
                    {selectedLocation.equipment_count} Stück
                  </span>
                </div>

                <div className="detail-card__row">
                  <span className="detail-card__row-label">Kapazität</span>
                  <span className="detail-card__row-value">
                    {Math.round((selectedLocation.equipment_count / 200) * 100)}% genutzt
                  </span>
                </div>
              </div>

              {selectedLocation.children && selectedLocation.children.length > 0 && (
                <div style={{ marginTop: 'var(--spacing-4)' }}>
                  <h3 style={{ fontSize: 'var(--font-size-base)', fontWeight: 'var(--font-weight-medium)', marginBottom: 'var(--spacing-2)' }}>
                    Unterbereiche
                  </h3>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
                    {selectedLocation.children.map((child) => (
                      <button
                        key={child.id}
                        className="list-item"
                        style={{ cursor: 'pointer', textAlign: 'left', width: '100%', padding: 'var(--padding-md)', background: 'var(--color-bg-tertiary)', border: 'none', borderRadius: 'var(--radius-md)' }}
                        onClick={() => setSelectedLocation(child)}
                      >
                        <div className="list-item__content">
                          <p className="list-item__title">
                            {getLocationIcon(child.type)} {child.name}
                          </p>
                          <p className="list-item__meta">
                            {child.equipment_count} Ausrüstung
                          </p>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default WarehouseViewPage
