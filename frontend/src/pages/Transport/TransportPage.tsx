import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import './Transport.module.scss'

interface Vehicle {
  id: string
  name: string
  license_plate: string
  status: 'available' | 'in_use' | 'maintenance'
  capacity_kg: number
  capacity_m3: number
}

interface Tour {
  id: string
  vehicle_id: string
  project_id: string
  project_name: string
  driver: string
  status: 'planned' | 'loading' | 'in_transit' | 'delivered' | 'completed'
  start_date: string
  end_date: string
  weight_kg: number
  volume_m3: number
  vehicle_capacity_kg: number
  vehicle_capacity_m3: number
}

// Mock data
const mockVehicles: Vehicle[] = [
  { id: '1', name: 'VW Crafter L3H3', license_plate: 'MUC-TR-001', status: 'available', capacity_kg: 1500, capacity_m3: 12 },
  { id: '2', name: 'Mercedes Sprinter', license_plate: 'MUC-TR-002', status: 'in_use', capacity_kg: 1800, capacity_m3: 14 },
  { id: '3', name: 'VW Transporter T5', license_plate: 'MUC-TR-003', status: 'available', capacity_kg: 1200, capacity_m3: 10 },
  { id: '4', name: 'MAN TGX 18 Sattelzug', license_plate: 'MUC-TR-004', status: 'in_use', capacity_kg: 8000, capacity_m3: 85 },
]

const mockTours: Tour[] = [
  {
    id: '1',
    vehicle_id: '1',
    project_id: '1',
    project_name: 'Stadtfest München 2026',
    driver: 'Thomas Müller',
    status: 'in_transit',
    start_date: '2026-03-22T08:00:00Z',
    end_date: '2026-03-22T14:00:00Z',
    weight_kg: 1200,
    volume_m3: 9,
    vehicle_capacity_kg: 1500,
    vehicle_capacity_m3: 12,
  },
  {
    id: '2',
    vehicle_id: '4',
    project_id: '3',
    project_name: 'Open Air Festival Bodensee',
    driver: 'Maria Schmidt',
    status: 'loading',
    start_date: '2026-03-23T06:00:00Z',
    end_date: '2026-03-23T18:00:00Z',
    weight_kg: 6500,
    volume_m3: 72,
    vehicle_capacity_kg: 8000,
    vehicle_capacity_m3: 85,
  },
  {
    id: '3',
    vehicle_id: '2',
    project_id: '2',
    project_name: 'Firmen-Gala TechCorp',
    driver: 'Peter Weber',
    status: 'planned',
    start_date: '2026-04-10T14:00:00Z',
    end_date: '2026-04-10T20:00:00Z',
    weight_kg: 800,
    volume_m3: 6,
    vehicle_capacity_kg: 1800,
    vehicle_capacity_m3: 14,
  },
  {
    id: '4',
    vehicle_id: '3',
    project_id: '1',
    project_name: 'Stadtfest München 2026',
    driver: 'Anna Fischer',
    status: 'completed',
    start_date: '2026-03-21T10:00:00Z',
    end_date: '2026-03-21T16:00:00Z',
    weight_kg: 950,
    volume_m3: 8,
    vehicle_capacity_kg: 1200,
    vehicle_capacity_m3: 10,
  },
]

function TransportPage() {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<'tours' | 'vehicles'>('tours')

  const { data: vehicles = mockVehicles } = useQuery({
    queryKey: ['vehicles'],
    queryFn: async () => mockVehicles,
    staleTime: 1000 * 60 * 5,
  })

  const { data: tours = mockTours } = useQuery({
    queryKey: ['tours'],
    queryFn: async () => mockTours,
    staleTime: 1000 * 60 * 5,
  })

  const activeTours = tours.filter(t => ['planned', 'loading', 'in_transit'].includes(t.status))
  const availableVehicles = vehicles.filter(v => v.status === 'available')
  const totalKmThisMonth = 450 // Mock data

  const getCapacityStatus = (used: number, capacity: number) => {
    const percentage = (used / capacity) * 100
    if (percentage <= 70) return 'ok'
    if (percentage <= 90) return 'warning'
    return 'danger'
  }

  return (
    <div className="transport-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Transport & Logistik</h1>
          <p className="page-subtitle">Verwaltung von Fahrzeugen und Transporten</p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button
            className="btn btn--primary"
            onClick={() => navigate('/transport/new-tour')}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            ➕ Neuer Transport
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => navigate('/transport/new-vehicle')}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            🚐 Fahrzeug hinzufügen
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-card__label">Aktive Transporte</div>
          <div className="stat-card__value">{activeTours.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Verfügbare Fahrzeuge</div>
          <div className="stat-card__value">{availableVehicles.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Gesamtfahrzeuge</div>
          <div className="stat-card__value">{vehicles.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">KM diesen Monat</div>
          <div className="stat-card__value">{totalKmThisMonth}</div>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="tabs">
        <button
          className={`tab-button ${activeTab === 'tours' ? 'tab-button--active' : ''}`}
          onClick={() => setActiveTab('tours')}
        >
          Transporte ({tours.length})
        </button>
        <button
          className={`tab-button ${activeTab === 'vehicles' ? 'tab-button--active' : ''}`}
          onClick={() => setActiveTab('vehicles')}
        >
          Fahrzeuge ({vehicles.length})
        </button>
      </div>

      {/* Tours Section */}
      {activeTab === 'tours' && (
        <div>
          {tours.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state__icon">📭</div>
              <h3 className="empty-state__title">Keine Transporte vorhanden</h3>
              <p className="empty-state__description">Erstellen Sie einen neuen Transport, um ihn hier zu sehen.</p>
            </div>
          ) : (
            <div className="tour-grid">
              {tours.map((tour) => {
                const weightPercentage = (tour.weight_kg / tour.vehicle_capacity_kg) * 100
                const volumePercentage = (tour.volume_m3 / tour.vehicle_capacity_m3) * 100
                const weightStatus = getCapacityStatus(tour.weight_kg, tour.vehicle_capacity_kg)
                const volumeStatus = getCapacityStatus(tour.volume_m3, tour.vehicle_capacity_m3)

                return (
                  <div
                    key={tour.id}
                    className="tour-card"
                    onClick={() => navigate(`/transport/tours/${tour.id}`)}
                  >
                    <div className="tour-card__header">
                      <h3 className="tour-card__title">{tour.project_name}</h3>
                      <StatusBadge status={tour.status} />
                    </div>

                    <div className="tour-card__meta">
                      <p>
                        <strong>Fahrer:</strong> {tour.driver}
                      </p>
                      <p>
                        <strong>Fahrzeug:</strong> {vehicles.find(v => v.id === tour.vehicle_id)?.name || 'N/A'}
                      </p>
                      <p>
                        <strong>Start:</strong> {new Date(tour.start_date).toLocaleDateString('de-DE', { month: 'short', day: 'numeric' })}
                      </p>
                      <p>
                        <strong>Status:</strong> {tour.status.replace(/_/g, ' ')}
                      </p>
                    </div>

                    <div className="tour-card__progress">
                      <div className="capacity-bar">
                        <div className="capacity-bar__label">
                          <span>Gewicht</span>
                          <span>
                            {tour.weight_kg}/{tour.vehicle_capacity_kg} kg
                          </span>
                        </div>
                        <div className="capacity-bar__track">
                          <div
                            className={`capacity-bar__fill capacity-bar__fill--${weightStatus}`}
                            style={{ width: `${Math.min(weightPercentage, 100)}%` }}
                          />
                        </div>
                      </div>

                      <div className="capacity-bar">
                        <div className="capacity-bar__label">
                          <span>Volumen</span>
                          <span>
                            {tour.volume_m3}/{tour.vehicle_capacity_m3} m³
                          </span>
                        </div>
                        <div className="capacity-bar__track">
                          <div
                            className={`capacity-bar__fill capacity-bar__fill--${volumeStatus}`}
                            style={{ width: `${Math.min(volumePercentage, 100)}%` }}
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      )}

      {/* Vehicles Section */}
      {activeTab === 'vehicles' && (
        <div>
          {vehicles.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state__icon">🚐</div>
              <h3 className="empty-state__title">Keine Fahrzeuge vorhanden</h3>
              <p className="empty-state__description">Fügen Sie Fahrzeuge hinzu, um sie hier zu verwalten.</p>
            </div>
          ) : (
            <div className="vehicle-grid">
              {vehicles.map((vehicle) => (
                <div
                  key={vehicle.id}
                  className="vehicle-card"
                  onClick={() => navigate(`/transport/vehicles/${vehicle.id}`)}
                >
                  <div className="vehicle-card__header">
                    <h3 className="vehicle-card__title">{vehicle.name}</h3>
                    <StatusBadge status={vehicle.status} />
                  </div>

                  <div className="vehicle-card__meta">
                    <p>
                      <strong>Kennzeichen:</strong> {vehicle.license_plate}
                    </p>
                    <p>
                      <strong>Max. Gewicht:</strong> {vehicle.capacity_kg} kg
                    </p>
                    <p>
                      <strong>Max. Volumen:</strong> {vehicle.capacity_m3} m³
                    </p>
                  </div>

                  <div className="vehicle-card__footer">
                    <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                      {vehicle.status === 'available' ? '✅ Verfügbar' : vehicle.status === 'in_use' ? '🔄 In Benutzung' : '🔧 Wartung'}
                    </span>
                    <button
                      className="btn btn--sm btn--primary"
                      onClick={(e) => {
                        e.stopPropagation()
                        navigate(`/transport/vehicles/${vehicle.id}`)
                      }}
                      style={{ flex: 'none' }}
                    >
                      Details
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default TransportPage
