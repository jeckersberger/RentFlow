import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { transportApi } from '../../services/api'
import './Transport.module.scss'

interface Vehicle {
  id: string
  name: string
  license_plate: string
  status: 'available' | 'in_use' | 'maintenance'
  capacity_kg: number
  capacity_m3: number
  vehicle_type?: string
}

interface Tour {
  id: string
  vehicle_id: string
  project_id: string
  project_name?: string
  driver?: string
  driver_id?: string
  status: 'planned' | 'loading' | 'in_transit' | 'delivered' | 'completed'
  start_date?: string
  end_date?: string
  departure_at?: string
  arrival_at?: string
  weight_kg?: number
  volume_m3?: number
  vehicle_capacity_kg?: number
  vehicle_capacity_m3?: number
  notes?: string
  km_start?: number
  km_end?: number
}

function TransportPage() {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<'tours' | 'vehicles'>('tours')

  const { data: vehiclesData, isLoading: vehiclesLoading, error: vehiclesError } = useQuery({
    queryKey: ['vehicles'],
    queryFn: () => transportApi.listVehicles(),
    staleTime: 1000 * 60 * 5,
  })

  const { data: toursData, isLoading: toursLoading, error: toursError } = useQuery({
    queryKey: ['tours'],
    queryFn: () => transportApi.listTours(),
    staleTime: 1000 * 60 * 5,
  })

  const vehicles: Vehicle[] = vehiclesData?.items || vehiclesData?.data || (Array.isArray(vehiclesData) ? vehiclesData : [])
  const tours: Tour[] = toursData?.items || toursData?.data || (Array.isArray(toursData) ? toursData : [])
  const isLoading = vehiclesLoading || toursLoading
  const hasError = vehiclesError || toursError

  const activeTours = tours.filter(t => ['planned', 'loading', 'in_transit'].includes(t.status))
  const availableVehicles = vehicles.filter(v => v.status === 'available')
  const completedTours = tours.filter(t => t.status === 'completed')
  const totalKmThisMonth = completedTours.reduce((sum, t) => {
    const km = (t.km_end || 0) - (t.km_start || 0)
    return sum + (km > 0 ? km : 0)
  }, 0)

  const getCapacityStatus = (used: number, capacity: number) => {
    if (!capacity) return 'ok'
    const percentage = (used / capacity) * 100
    if (percentage <= 70) return 'ok'
    if (percentage <= 90) return 'warning'
    return 'danger'
  }

  if (isLoading) {
    return (
      <div className="transport-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Transport & Logistik</h1>
            <p className="page-subtitle">Daten werden geladen...</p>
          </div>
        </div>
        <div className="stats-grid">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="stat-card">
              <div className="stat-card__label">Laden...</div>
              <div className="stat-card__value">--</div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (hasError) {
    return (
      <div className="transport-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Transport & Logistik</h1>
            <p className="page-subtitle">Fehler beim Laden der Daten. Bitte versuchen Sie es erneut.</p>
          </div>
        </div>
        <div className="empty-state">
          <div className="empty-state__icon">&#x26A0;</div>
          <h3 className="empty-state__title">Daten konnten nicht geladen werden</h3>
          <p className="empty-state__description">
            {vehiclesError ? String(vehiclesError) : toursError ? String(toursError) : 'Unbekannter Fehler'}
          </p>
        </div>
      </div>
    )
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
                const vehicle = vehicles.find(v => v.id === tour.vehicle_id)
                const vCapKg = tour.vehicle_capacity_kg || vehicle?.capacity_kg || 0
                const vCapM3 = tour.vehicle_capacity_m3 || vehicle?.capacity_m3 || 0
                const weightKg = tour.weight_kg || 0
                const volM3 = tour.volume_m3 || 0
                const weightPercentage = vCapKg ? (weightKg / vCapKg) * 100 : 0
                const volumePercentage = vCapM3 ? (volM3 / vCapM3) * 100 : 0
                const weightStatus = getCapacityStatus(weightKg, vCapKg)
                const volumeStatus = getCapacityStatus(volM3, vCapM3)
                const startDate = tour.start_date || tour.departure_at
                const tourLabel = tour.project_name || tour.notes || `Tour #${tour.id}`

                return (
                  <div
                    key={tour.id}
                    className="tour-card"
                    onClick={() => navigate(`/transport/tours/${tour.id}`)}
                  >
                    <div className="tour-card__header">
                      <h3 className="tour-card__title">{tourLabel}</h3>
                      <StatusBadge status={tour.status} />
                    </div>

                    <div className="tour-card__meta">
                      {(tour.driver || tour.driver_id) && (
                        <p>
                          <strong>Fahrer:</strong> {tour.driver || tour.driver_id}
                        </p>
                      )}
                      <p>
                        <strong>Fahrzeug:</strong> {vehicle?.name || 'N/A'}
                      </p>
                      {startDate && (
                        <p>
                          <strong>Start:</strong> {new Date(startDate).toLocaleDateString('de-DE', { month: 'short', day: 'numeric' })}
                        </p>
                      )}
                      <p>
                        <strong>Status:</strong> {tour.status.replace(/_/g, ' ')}
                      </p>
                    </div>

                    {(vCapKg > 0 || vCapM3 > 0) && (
                      <div className="tour-card__progress">
                        {vCapKg > 0 && (
                          <div className="capacity-bar">
                            <div className="capacity-bar__label">
                              <span>Gewicht</span>
                              <span>
                                {weightKg}/{vCapKg} kg
                              </span>
                            </div>
                            <div className="capacity-bar__track">
                              <div
                                className={`capacity-bar__fill capacity-bar__fill--${weightStatus}`}
                                style={{ width: `${Math.min(weightPercentage, 100)}%` }}
                              />
                            </div>
                          </div>
                        )}

                        {vCapM3 > 0 && (
                          <div className="capacity-bar">
                            <div className="capacity-bar__label">
                              <span>Volumen</span>
                              <span>
                                {volM3}/{vCapM3} m3
                              </span>
                            </div>
                            <div className="capacity-bar__track">
                              <div
                                className={`capacity-bar__fill capacity-bar__fill--${volumeStatus}`}
                                style={{ width: `${Math.min(volumePercentage, 100)}%` }}
                              />
                            </div>
                          </div>
                        )}
                      </div>
                    )}
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
                    {vehicle.vehicle_type && (
                      <p>
                        <strong>Typ:</strong> {vehicle.vehicle_type}
                      </p>
                    )}
                    <p>
                      <strong>Max. Gewicht:</strong> {vehicle.capacity_kg?.toLocaleString('de-DE')} kg
                    </p>
                    <p>
                      <strong>Max. Volumen:</strong> {vehicle.capacity_m3} m3
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
