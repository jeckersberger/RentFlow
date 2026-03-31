import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Modal } from '../../components/Modal/Modal'
import { Input } from '../../components/Form/Input'
import { transportApi, projectApi, crewApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { SkeletonKPI, SkeletonTable } from '../../components/Skeleton/SkeletonLoader'
import './Transport.scss'

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

const VEHICLE_TYPES = [
  { value: 'lkw', label: 'LKW' },
  { value: 'transporter', label: 'Transporter' },
  { value: 'pkw', label: 'PKW' },
  { value: 'anhaenger', label: 'Anhänger' },
]

const LICENSE_CLASSES = [
  { value: 'B', label: 'B' },
  { value: 'BE', label: 'BE' },
  { value: 'C', label: 'C' },
  { value: 'CE', label: 'CE' },
  { value: 'C1', label: 'C1' },
  { value: 'C1E', label: 'C1E' },
]

const VEHICLE_STATUS_LABELS: Record<string, string> = {
  available: 'Verfügbar',
  in_use: 'Im Einsatz',
  maintenance: 'Wartung',
}

const TOUR_STATUS_LABELS: Record<string, string> = {
  planned: 'Geplant',
  loading: 'Beladung',
  in_transit: 'Unterwegs',
  delivered: 'Geliefert',
  completed: 'Abgeschlossen',
}

interface VehicleForm {
  name: string
  vehicle_type: string
  license_plate: string
  capacity_kg: string
  capacity_m3: string
  loading_length: string
  loading_width: string
  loading_height: string
  license_class: string
}

const emptyVehicleForm: VehicleForm = {
  name: '',
  vehicle_type: 'transporter',
  license_plate: '',
  capacity_kg: '',
  capacity_m3: '',
  loading_length: '',
  loading_width: '',
  loading_height: '',
  license_class: 'B',
}

interface TourForm {
  vehicle_id: string
  project_id: string
  driver_id: string
  start_date: string
  end_date: string
  notes: string
}

const emptyTourForm: TourForm = {
  vehicle_id: '',
  project_id: '',
  driver_id: '',
  start_date: '',
  end_date: '',
  notes: '',
}

function TransportPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [activeTab, setActiveTab] = useState<'tours' | 'vehicles'>('vehicles')
  const [filterVehicleStatus, setFilterVehicleStatus] = useState<string>('all')
  const [filterTourStatus, setFilterTourStatus] = useState<string>('all')
  const [showVehicleModal, setShowVehicleModal] = useState(false)
  const [vehicleForm, setVehicleForm] = useState<VehicleForm>(emptyVehicleForm)
  const [showTourModal, setShowTourModal] = useState(false)
  const [tourForm, setTourForm] = useState<TourForm>(emptyTourForm)

  const { data: vehiclesData, isLoading: vehiclesLoading, error: vehiclesError } = useQuery({
    queryKey: ['vehicles'],
    queryFn: () => transportApi.listVehicles(),
    staleTime: 1000 * 60 * 5,
    retry: 1,
  })

  const { data: toursData, isLoading: toursLoading, error: toursError } = useQuery({
    queryKey: ['tours'],
    queryFn: () => transportApi.listTours(),
    staleTime: 1000 * 60 * 5,
    retry: 1,
  })

  const { data: projectsData } = useQuery({
    queryKey: ['projects-for-tour'],
    queryFn: () => projectApi.list(1, 200),
    enabled: showTourModal,
    staleTime: 1000 * 60 * 5,
  })

  const { data: driversData } = useQuery({
    queryKey: ['crew-drivers'],
    queryFn: () => crewApi.listMembers({ per_page: 200 }),
    enabled: showTourModal,
    staleTime: 1000 * 60 * 5,
  })

  const projectsList: { id: string; name: string }[] =
    projectsData?.items || projectsData?.data || (Array.isArray(projectsData) ? projectsData : [])

  const driversList: { id: string; first_name: string; last_name: string }[] =
    driversData?.data || driversData?.items || (Array.isArray(driversData) ? driversData : [])

  const createTourMutation = useMutation({
    mutationFn: (data: any) => transportApi.createTour(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tours'] })
      setShowTourModal(false)
      setTourForm(emptyTourForm)
      addNotification('Der Transport wurde erfolgreich angelegt.', 'success', { title: 'Transport erstellt' })
    },
    onError: (err: any) => {
      addNotification(`Transport konnte nicht erstellt werden: ${err.message || 'Unbekannter Fehler'}`, 'error', { title: 'Fehler' })
    },
  })

  const createVehicleMutation = useMutation({
    mutationFn: (data: any) => transportApi.createVehicle(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vehicles'] })
      setShowVehicleModal(false)
      setVehicleForm(emptyVehicleForm)
      addNotification('Das Fahrzeug wurde erfolgreich angelegt.', 'success', { title: 'Fahrzeug erstellt' })
    },
    onError: (err: any) => {
      addNotification(`Fahrzeug konnte nicht erstellt werden: ${err.message || 'Unbekannter Fehler'}`, 'error', { title: 'Fehler' })
    },
  })

  const vehicles: Vehicle[] = vehiclesData?.items || vehiclesData?.data || (Array.isArray(vehiclesData) ? vehiclesData : [])
  const tours: Tour[] = toursData?.items || toursData?.data || (Array.isArray(toursData) ? toursData : [])
  const isLoading = (vehiclesLoading && !vehiclesError) || (toursLoading && !toursError)
  const hasError = vehiclesError || toursError

  const activeTours = tours.filter(t => ['planned', 'loading', 'in_transit'].includes(t.status))
  const availableVehicles = vehicles.filter(v => v.status === 'available')
  const completedTours = tours.filter(t => t.status === 'completed')
  const totalKmThisMonth = completedTours.reduce((sum, t) => {
    const km = (t.km_end || 0) - (t.km_start || 0)
    return sum + (km > 0 ? km : 0)
  }, 0)

  const filteredVehicles = vehicles.filter(v =>
    filterVehicleStatus === 'all' || v.status === filterVehicleStatus
  )

  const filteredTours = tours.filter(t =>
    filterTourStatus === 'all' || t.status === filterTourStatus
  )

  const getCapacityStatus = (used: number, capacity: number) => {
    if (!capacity) return 'ok'
    const percentage = (used / capacity) * 100
    if (percentage <= 70) return 'ok'
    if (percentage <= 90) return 'warning'
    return 'danger'
  }

  const handleTourSubmit = () => {
    if (!tourForm.vehicle_id) {
      addNotification('Bitte waehlen Sie ein Fahrzeug aus.', 'error', { title: 'Fehler' })
      return
    }
    createTourMutation.mutate({
      vehicle_id: tourForm.vehicle_id,
      project_id: tourForm.project_id || undefined,
      driver_id: tourForm.driver_id || undefined,
      departure_at: tourForm.start_date || undefined,
      arrival_at: tourForm.end_date || undefined,
      start_date: tourForm.start_date || undefined,
      end_date: tourForm.end_date || undefined,
      notes: tourForm.notes || undefined,
      status: 'planned',
    })
  }

  const handleVehicleSubmit = () => {
    if (!vehicleForm.name.trim() || !vehicleForm.license_plate.trim()) {
      addNotification('Name und Kennzeichen sind Pflichtfelder.', 'error', { title: 'Fehler' })
      return
    }
    createVehicleMutation.mutate({
      name: vehicleForm.name,
      vehicle_type: vehicleForm.vehicle_type,
      license_plate: vehicleForm.license_plate,
      capacity_kg: vehicleForm.capacity_kg ? parseFloat(vehicleForm.capacity_kg) : 0,
      capacity_m3: vehicleForm.capacity_m3 ? parseFloat(vehicleForm.capacity_m3) : 0,
      loading_length: vehicleForm.loading_length ? parseFloat(vehicleForm.loading_length) : undefined,
      loading_width: vehicleForm.loading_width ? parseFloat(vehicleForm.loading_width) : undefined,
      loading_height: vehicleForm.loading_height ? parseFloat(vehicleForm.loading_height) : undefined,
      license_class: vehicleForm.license_class,
    })
  }

  const getNextTourForVehicle = (vehicleId: string): Tour | undefined => {
    return tours
      .filter(t => t.vehicle_id === vehicleId && ['planned', 'loading', 'in_transit'].includes(t.status))
      .sort((a, b) => {
        const dateA = a.departure_at || a.start_date || ''
        const dateB = b.departure_at || b.start_date || ''
        return dateA.localeCompare(dateB)
      })[0]
  }

  if (isLoading) {
    return (
      <div className="transport-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Transport & Logistik</h1>
          </div>
        </div>
        <SkeletonKPI count={4} />
        <SkeletonTable rows={5} columns={5} />
      </div>
    )
  }

  if (hasError) {
    return (
      <div className="transport-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Transport & Logistik</h1>
            <p className="page-subtitle">Fehler beim Laden der Daten</p>
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
          <p className="page-subtitle">Verwaltung von {vehicles.length} Fahrzeugen und {tours.length} Transporten</p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button
            className="btn btn--primary"
            onClick={() => setShowVehicleModal(true)}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            + Neues Fahrzeug
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => setShowTourModal(true)}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            + Neuer Transport
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-card__label">Aktive Transporte</div>
          <div className="stat-card__value" style={{ color: 'var(--color-primary)' }}>{activeTours.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Verfügbare Fahrzeuge</div>
          <div className="stat-card__value" style={{ color: 'var(--color-success)' }}>{availableVehicles.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Gesamtfahrzeuge</div>
          <div className="stat-card__value">{vehicles.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">KM diesen Monat</div>
          <div className="stat-card__value">{totalKmThisMonth.toLocaleString('de-DE')}</div>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="tabs">
        <button
          className={`tab-button ${activeTab === 'vehicles' ? 'tab-button--active' : ''}`}
          onClick={() => setActiveTab('vehicles')}
        >
          Fahrzeuge ({vehicles.length})
        </button>
        <button
          className={`tab-button ${activeTab === 'tours' ? 'tab-button--active' : ''}`}
          onClick={() => setActiveTab('tours')}
        >
          Transporte ({tours.length})
        </button>
      </div>

      {/* Vehicles Tab */}
      {activeTab === 'vehicles' && (
        <div>
          {/* Vehicle Status Filter */}
          <div className="filters">
            <div className="filter-group">
              <label htmlFor="vehicle-status-filter" className="filter-label">Status:</label>
              <select
                id="vehicle-status-filter"
                className="filter-select"
                value={filterVehicleStatus}
                onChange={(e) => setFilterVehicleStatus(e.target.value)}
              >
                <option value="all">Alle</option>
                <option value="available">Verfügbar</option>
                <option value="in_use">Im Einsatz</option>
                <option value="maintenance">Wartung</option>
              </select>
            </div>
          </div>

          {filteredVehicles.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state__icon">&#x1F699;</div>
              <h3 className="empty-state__title">Keine Fahrzeuge gefunden</h3>
              <p className="empty-state__description">
                {vehicles.length === 0
                  ? 'Legen Sie Ihr erstes Fahrzeug an, um loszulegen.'
                  : 'Ändern Sie den Filter, um Fahrzeuge anzuzeigen.'}
              </p>
              {vehicles.length === 0 && (
                <button className="btn btn--primary" onClick={() => setShowVehicleModal(true)}>
                  + Neues Fahrzeug
                </button>
              )}
            </div>
          ) : (
            <div className="vehicle-table-wrapper">
              <table className="vehicle-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Typ</th>
                    <th>Kennzeichen</th>
                    <th>Kapazität (m³)</th>
                    <th>Status</th>
                    <th>Nächster Einsatz</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredVehicles.map((vehicle) => {
                    const nextTour = getNextTourForVehicle(vehicle.id)
                    const nextDate = nextTour?.departure_at || nextTour?.start_date
                    return (
                      <tr
                        key={vehicle.id}
                        className="vehicle-table__row"
                        onClick={() => navigate(`/transport/vehicles/${vehicle.id}`)}
                      >
                        <td>
                          <span className="vehicle-table__name">{vehicle.name}</span>
                        </td>
                        <td>
                          <span className="vehicle-type-tag">
                            {vehicle.vehicle_type || '—'}
                          </span>
                        </td>
                        <td>
                          <span className="license-plate">{vehicle.license_plate}</span>
                        </td>
                        <td>{vehicle.capacity_m3 ? `${vehicle.capacity_m3} m³` : '—'}</td>
                        <td>
                          <StatusBadge status={vehicle.status} label={VEHICLE_STATUS_LABELS[vehicle.status]} />
                        </td>
                        <td>
                          {nextTour ? (
                            <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                              {nextDate ? new Date(nextDate).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' }) : '—'}
                              {nextTour.notes && (
                                <span style={{ display: 'block', color: 'var(--color-text-secondary)', fontSize: '11px' }}>
                                  {nextTour.notes.substring(0, 30)}{nextTour.notes.length > 30 ? '...' : ''}
                                </span>
                              )}
                            </span>
                          ) : (
                            <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>—</span>
                          )}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Tours Tab */}
      {activeTab === 'tours' && (
        <div>
          {/* Tour Status Filter */}
          <div className="filters">
            <div className="filter-group">
              <label htmlFor="tour-status-filter" className="filter-label">Status:</label>
              <select
                id="tour-status-filter"
                className="filter-select"
                value={filterTourStatus}
                onChange={(e) => setFilterTourStatus(e.target.value)}
              >
                <option value="all">Alle</option>
                <option value="planned">Geplant</option>
                <option value="loading">Beladung</option>
                <option value="in_transit">Unterwegs</option>
                <option value="delivered">Geliefert</option>
                <option value="completed">Abgeschlossen</option>
              </select>
            </div>
          </div>

          {/* Active Tours Section */}
          {activeTours.length > 0 && filterTourStatus === 'all' && (
            <div className="active-tours-section">
              <h3 className="section-title">Aktive Transporte</h3>
              <div className="tour-grid">
                {activeTours.map((tour) => {
                  const vehicle = vehicles.find(v => v.id === tour.vehicle_id)
                  const startDate = tour.departure_at || tour.start_date
                  const endDate = tour.arrival_at || tour.end_date
                  const tourLabel = tour.notes || tour.project_name || `Tour #${tour.id}`

                  return (
                    <div
                      key={tour.id}
                      className="tour-card tour-card--active"
                      onClick={() => navigate(`/transport/tours/${tour.id}`)}
                    >
                      <div className="tour-card__header">
                        <h3 className="tour-card__title">{tourLabel}</h3>
                        <StatusBadge status={tour.status} label={TOUR_STATUS_LABELS[tour.status]} />
                      </div>
                      <div className="tour-card__meta">
                        <p><strong>Fahrzeug:</strong> {vehicle?.name || 'k.A.'}</p>
                        {(tour.driver || tour.driver_id) && (
                          <p><strong>Fahrer:</strong> {tour.driver || tour.driver_id}</p>
                        )}
                        {startDate && (
                          <p><strong>Abfahrt:</strong> {new Date(startDate).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })}</p>
                        )}
                        {endDate && (
                          <p><strong>Ankunft:</strong> {new Date(endDate).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })}</p>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {/* All Tours */}
          {filteredTours.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state__icon">&#x1F69B;</div>
              <h3 className="empty-state__title">Keine Transporte gefunden</h3>
              <p className="empty-state__description">Erstellen Sie einen neuen Transport, um ihn hier zu sehen.</p>
            </div>
          ) : (
            <>
              {(filterTourStatus !== 'all' || activeTours.length === 0) && (
                <div className="tour-grid">
                  {filteredTours.map((tour) => {
                    const vehicle = vehicles.find(v => v.id === tour.vehicle_id)
                    const vCapKg = tour.vehicle_capacity_kg || vehicle?.capacity_kg || 0
                    const vCapM3 = tour.vehicle_capacity_m3 || vehicle?.capacity_m3 || 0
                    const weightKg = tour.weight_kg || 0
                    const volM3 = tour.volume_m3 || 0
                    const weightPercentage = vCapKg ? (weightKg / vCapKg) * 100 : 0
                    const volumePercentage = vCapM3 ? (volM3 / vCapM3) * 100 : 0
                    const weightStatus = getCapacityStatus(weightKg, vCapKg)
                    const volumeStatus = getCapacityStatus(volM3, vCapM3)
                    const startDate = tour.departure_at || tour.start_date
                    const endDate = tour.arrival_at || tour.end_date
                    const tourLabel = tour.notes || tour.project_name || `Tour #${tour.id}`

                    return (
                      <div
                        key={tour.id}
                        className="tour-card"
                        onClick={() => navigate(`/transport/tours/${tour.id}`)}
                      >
                        <div className="tour-card__header">
                          <h3 className="tour-card__title">{tourLabel}</h3>
                          <StatusBadge status={tour.status} label={TOUR_STATUS_LABELS[tour.status]} />
                        </div>

                        <div className="tour-card__meta">
                          <p><strong>Fahrzeug:</strong> {vehicle?.name || 'k.A.'}</p>
                          {(tour.driver || tour.driver_id) && (
                            <p><strong>Fahrer:</strong> {tour.driver || tour.driver_id}</p>
                          )}
                          {startDate && (
                            <p><strong>Abfahrt:</strong> {new Date(startDate).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })}</p>
                          )}
                          {endDate && (
                            <p><strong>Ankunft:</strong> {new Date(endDate).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })}</p>
                          )}
                        </div>

                        {(vCapKg > 0 || vCapM3 > 0) && (
                          <div className="tour-card__progress">
                            {vCapKg > 0 && (
                              <div className="capacity-bar">
                                <div className="capacity-bar__label">
                                  <span>Gewicht</span>
                                  <span>{weightKg}/{vCapKg} kg</span>
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
                                  <span>{volM3}/{vCapM3} m³</span>
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

              {/* Completed Tours History */}
              {filterTourStatus === 'all' && completedTours.length > 0 && (
                <div className="completed-tours-section">
                  <h3 className="section-title">Abgeschlossene Transporte</h3>
                  <div className="tour-grid">
                    {completedTours.map((tour) => {
                      const vehicle = vehicles.find(v => v.id === tour.vehicle_id)
                      const km = (tour.km_end || 0) - (tour.km_start || 0)
                      const tourLabel = tour.notes || tour.project_name || `Tour #${tour.id}`

                      return (
                        <div
                          key={tour.id}
                          className="tour-card tour-card--completed"
                          onClick={() => navigate(`/transport/tours/${tour.id}`)}
                        >
                          <div className="tour-card__header">
                            <h3 className="tour-card__title">{tourLabel}</h3>
                            <StatusBadge status="completed" label="Abgeschlossen" />
                          </div>
                          <div className="tour-card__meta">
                            <p><strong>Fahrzeug:</strong> {vehicle?.name || 'k.A.'}</p>
                            {km > 0 && <p><strong>Strecke:</strong> {km.toLocaleString('de-DE')} km</p>}
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      )}

      {/* New Vehicle Modal */}
      <Modal
        isOpen={showVehicleModal}
        onClose={() => { setShowVehicleModal(false); setVehicleForm(emptyVehicleForm) }}
        title="Neues Fahrzeug"
        size="lg"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
            <button
              className="btn btn--secondary"
              onClick={() => { setShowVehicleModal(false); setVehicleForm(emptyVehicleForm) }}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={handleVehicleSubmit}
              disabled={createVehicleMutation.isPending}
            >
              {createVehicleMutation.isPending ? 'Speichern...' : 'Fahrzeug anlegen'}
            </button>
          </div>
        }
      >
        <div className="form-grid">
          <Input
            label="Fahrzeugname *"
            value={vehicleForm.name}
            onChange={(e) => setVehicleForm(prev => ({ ...prev, name: e.target.value }))}
            placeholder="z.B. MAN TGX 18t"
          />
          <div className="form-group">
            <label className="form-label">Typ</label>
            <select
              className="form-input"
              value={vehicleForm.vehicle_type}
              onChange={(e) => setVehicleForm(prev => ({ ...prev, vehicle_type: e.target.value }))}
            >
              {VEHICLE_TYPES.map(t => (
                <option key={t.value} value={t.value}>{t.label}</option>
              ))}
            </select>
          </div>
          <Input
            label="Kennzeichen *"
            value={vehicleForm.license_plate}
            onChange={(e) => setVehicleForm(prev => ({ ...prev, license_plate: e.target.value.toUpperCase() }))}
            placeholder="M-AB 1234"
          />
          <div className="form-group">
            <label className="form-label">Führerscheinklasse</label>
            <select
              className="form-input"
              value={vehicleForm.license_class}
              onChange={(e) => setVehicleForm(prev => ({ ...prev, license_class: e.target.value }))}
            >
              {LICENSE_CLASSES.map(c => (
                <option key={c.value} value={c.value}>{c.label}</option>
              ))}
            </select>
          </div>
          <Input
            label="Kapazität (kg)"
            type="number"
            value={vehicleForm.capacity_kg}
            onChange={(e) => setVehicleForm(prev => ({ ...prev, capacity_kg: e.target.value }))}
            placeholder="z.B. 18000"
            min="0"
          />
          <Input
            label="Kapazität (m³)"
            type="number"
            value={vehicleForm.capacity_m3}
            onChange={(e) => setVehicleForm(prev => ({ ...prev, capacity_m3: e.target.value }))}
            placeholder="z.B. 52"
            min="0"
          />
          <div className="form-group form-group--full">
            <label className="form-label">Ladeflächenmaße (m)</label>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 'var(--spacing-3)' }}>
              <Input
                placeholder="Länge"
                type="number"
                value={vehicleForm.loading_length}
                onChange={(e) => setVehicleForm(prev => ({ ...prev, loading_length: e.target.value }))}
                min="0"
                step="0.1"
              />
              <Input
                placeholder="Breite"
                type="number"
                value={vehicleForm.loading_width}
                onChange={(e) => setVehicleForm(prev => ({ ...prev, loading_width: e.target.value }))}
                min="0"
                step="0.1"
              />
              <Input
                placeholder="Höhe"
                type="number"
                value={vehicleForm.loading_height}
                onChange={(e) => setVehicleForm(prev => ({ ...prev, loading_height: e.target.value }))}
                min="0"
                step="0.1"
              />
            </div>
          </div>
        </div>
      </Modal>

      {/* New Tour Modal */}
      <Modal
        isOpen={showTourModal}
        onClose={() => { setShowTourModal(false); setTourForm(emptyTourForm) }}
        title="Neuer Transport"
        size="lg"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
            <button
              className="btn btn--secondary"
              onClick={() => { setShowTourModal(false); setTourForm(emptyTourForm) }}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={handleTourSubmit}
              disabled={createTourMutation.isPending}
            >
              {createTourMutation.isPending ? 'Speichern...' : 'Transport anlegen'}
            </button>
          </div>
        }
      >
        <div className="form-grid">
          <div className="form-group">
            <label className="form-label">Fahrzeug *</label>
            <select
              className="form-input"
              value={tourForm.vehicle_id}
              onChange={(e) => setTourForm(prev => ({ ...prev, vehicle_id: e.target.value }))}
            >
              <option value="">-- Fahrzeug waehlen --</option>
              {vehicles.map(v => (
                <option key={v.id} value={v.id}>
                  {v.name} ({v.license_plate}) {v.status !== 'available' ? `[${VEHICLE_STATUS_LABELS[v.status] || v.status}]` : ''}
                </option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label className="form-label">Projekt</label>
            <select
              className="form-input"
              value={tourForm.project_id}
              onChange={(e) => setTourForm(prev => ({ ...prev, project_id: e.target.value }))}
            >
              <option value="">-- Projekt waehlen (optional) --</option>
              {projectsList.map(p => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </div>
          <Input
            label="Abfahrt (Datum/Uhrzeit)"
            type="datetime-local"
            value={tourForm.start_date}
            onChange={(e) => setTourForm(prev => ({ ...prev, start_date: e.target.value }))}
          />
          <Input
            label="Ankunft (Datum/Uhrzeit)"
            type="datetime-local"
            value={tourForm.end_date}
            onChange={(e) => setTourForm(prev => ({ ...prev, end_date: e.target.value }))}
          />
          <div className="form-group">
            <label className="form-label">Fahrer</label>
            <select
              className="form-input"
              value={tourForm.driver_id}
              onChange={(e) => setTourForm(prev => ({ ...prev, driver_id: e.target.value }))}
            >
              <option value="">-- Fahrer waehlen (optional) --</option>
              {driversList.map(d => (
                <option key={d.id} value={d.id}>{d.first_name} {d.last_name}</option>
              ))}
            </select>
          </div>
          <Input
            label="Bemerkungen"
            value={tourForm.notes}
            onChange={(e) => setTourForm(prev => ({ ...prev, notes: e.target.value }))}
            placeholder="z.B. Anlieferung Halle 3"
          />
        </div>
      </Modal>
    </div>
  )
}

export default TransportPage
