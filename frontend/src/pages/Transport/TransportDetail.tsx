import { useState } from 'react'
import { useNavigate, useParams, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { transportApi } from '../../services/api'
import './Transport.module.scss'

interface Equipment {
  id: string
  name: string
  quantity: number
  weight_kg: number
  volume_m3: number
}

interface LogEntry {
  id: string
  timestamp: string
  action: string
  notes: string
}

interface TourDetail {
  id: string
  project_name: string
  vehicle_name: string
  driver: string
  status: 'planned' | 'loading' | 'in_transit' | 'delivered' | 'completed'
  start_date: string
  end_date: string
  weight_kg: number
  volume_m3: number
  vehicle_capacity_kg: number
  vehicle_capacity_m3: number
  equipment: Equipment[]
  logs: LogEntry[]
}

// Mock data for fallback
const mockTourDetail: TourDetail = {
  id: '1',
  project_name: 'Stadtfest München 2026',
  vehicle_name: 'VW Crafter L3H3',
  driver: 'Thomas Müller',
  status: 'in_transit',
  start_date: '2026-03-22T08:00:00Z',
  end_date: '2026-03-22T14:00:00Z',
  weight_kg: 1200,
  volume_m3: 9,
  vehicle_capacity_kg: 1500,
  vehicle_capacity_m3: 12,
  equipment: [
    { id: '1', name: 'JBL VTX A12 Lautsprecher', quantity: 4, weight_kg: 200, volume_m3: 1.2 },
    { id: '2', name: 'Shure SM58 Mikrofone', quantity: 12, weight_kg: 180, volume_m3: 0.8 },
    { id: '3', name: 'MA Lighting grandMA3 Pult', quantity: 1, weight_kg: 150, volume_m3: 1.5 },
    { id: '4', name: 'Martin MAC Aura XB Moving Head', quantity: 8, weight_kg: 480, volume_m3: 2.4 },
    { id: '5', name: 'Prolyte X30V Truss 3m', quantity: 12, weight_kg: 192, volume_m3: 3.1 },
  ],
  logs: [
    { id: '1', timestamp: '2026-03-22T08:15:00Z', action: 'Tour gestartet', notes: 'Fahrzeug verlässt Depot mit vollständiger Ausrüstung' },
    { id: '2', timestamp: '2026-03-22T09:30:00Z', action: 'Beladung abgeschlossen', notes: 'Gesamtgewicht: 1200 kg, Volumen: 9 m³' },
    { id: '3', timestamp: '2026-03-22T11:45:00Z', action: 'An Einsatzort angekommen', notes: 'Marienplatz München - Aufbau wird begonnen' },
  ],
}

const TOUR_STATUS_LABELS: Record<string, string> = {
  planned: 'Geplant',
  loading: 'Beladung',
  in_transit: 'Unterwegs',
  delivered: 'Geliefert',
  completed: 'Abgeschlossen',
}

const tourStatusSteps = ['planned', 'loading', 'in_transit', 'delivered', 'completed']

function TransportDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const location = useLocation()
  const [expandedEquipment, setExpandedEquipment] = useState<string | null>(null)

  const isVehicleDetail = location.pathname.includes('/vehicles/')

  // Vehicle Detail Query
  const { data: vehicleData, isLoading: vehicleLoading } = useQuery({
    queryKey: ['vehicle', id],
    queryFn: () => transportApi.getVehicle(id!),
    enabled: isVehicleDetail && !!id,
    staleTime: 1000 * 60 * 5,
  })

  // Vehicle Tours Query
  const { data: toursData } = useQuery({
    queryKey: ['tours'],
    queryFn: () => transportApi.listTours(),
    enabled: isVehicleDetail && !!id,
    staleTime: 1000 * 60 * 5,
  })

  // Tour Detail Query
  const { data: tourRaw, isLoading: tourLoading } = useQuery({
    queryKey: ['tour', id],
    queryFn: async () => {
      try {
        const result = await transportApi.getTour(id!)
        return result
      } catch {
        return null
      }
    },
    enabled: !isVehicleDetail && !!id,
  })

  // Vehicle Detail View
  if (isVehicleDetail) {
    const vehicle = vehicleData
    const allTours = toursData?.items || toursData?.data || (Array.isArray(toursData) ? toursData : [])
    const vehicleTours = allTours.filter((t: any) => t.vehicle_id === id)
    const upcomingTours = vehicleTours.filter((t: any) => ['planned', 'loading', 'in_transit'].includes(t.status))
    const completedTours = vehicleTours.filter((t: any) => t.status === 'completed')

    if (vehicleLoading) {
      return (
        <div className="transport-detail-page">
          <div className="page-header">
            <div>
              <h1 className="page-title">Fahrzeug wird geladen...</h1>
            </div>
          </div>
        </div>
      )
    }

    if (!vehicle) {
      return (
        <div className="transport-detail-page">
          <div className="page-header">
            <div>
              <button className="btn btn--secondary" onClick={() => navigate('/transport')} style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}>
                Zurück
              </button>
              <h1 className="page-title" style={{ marginTop: 'var(--spacing-3)' }}>Fahrzeug nicht gefunden</h1>
            </div>
          </div>
        </div>
      )
    }

    const VEHICLE_STATUS_LABELS: Record<string, string> = {
      available: 'Verfügbar',
      in_use: 'Im Einsatz',
      maintenance: 'Wartung',
    }

    return (
      <div className="transport-detail-page">
        <div className="page-header">
          <div>
            <button
              className="btn btn--sm btn--secondary"
              onClick={() => navigate('/transport')}
              style={{ marginBottom: 'var(--spacing-3)' }}
            >
              Zurück
            </button>
            <h1 className="page-title">{vehicle.name}</h1>
            <p className="page-subtitle">Fahrzeugdetails und Einsatzhistorie</p>
          </div>
          <StatusBadge status={vehicle.status} label={VEHICLE_STATUS_LABELS[vehicle.status]} size="lg" />
        </div>

        <div className="detail-grid">
          <div>
            {/* Vehicle Info */}
            <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
              <h2 className="detail-card__title">Fahrzeugdaten</h2>
              <div className="detail-card__content">
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Name</div>
                  <div className="detail-card__row-value">{vehicle.name}</div>
                </div>
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Typ</div>
                  <div className="detail-card__row-value">{vehicle.vehicle_type || '—'}</div>
                </div>
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Kennzeichen</div>
                  <div className="detail-card__row-value">
                    <span className="license-plate">{vehicle.license_plate}</span>
                  </div>
                </div>
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Max. Gewicht</div>
                  <div className="detail-card__row-value">{vehicle.capacity_kg?.toLocaleString('de-DE')} kg</div>
                </div>
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Max. Volumen</div>
                  <div className="detail-card__row-value">{vehicle.capacity_m3} m³</div>
                </div>
                {vehicle.loading_length && (
                  <div className="detail-card__row">
                    <div className="detail-card__row-label">Ladefläche</div>
                    <div className="detail-card__row-value">
                      {vehicle.loading_length}m x {vehicle.loading_width || '—'}m x {vehicle.loading_height || '—'}m
                    </div>
                  </div>
                )}
                {vehicle.license_class && (
                  <div className="detail-card__row">
                    <div className="detail-card__row-label">Führerschein</div>
                    <div className="detail-card__row-value">Klasse {vehicle.license_class}</div>
                  </div>
                )}
                {vehicle.dguv_next_check && (
                  <div className="detail-card__row">
                    <div className="detail-card__row-label">DGUV-Prüfung</div>
                    <div className="detail-card__row-value">
                      Nächste: {new Date(vehicle.dguv_next_check).toLocaleDateString('de-DE')}
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Upcoming Tours */}
            <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
              <h2 className="detail-card__title">Geplante Einsätze ({upcomingTours.length})</h2>
              {upcomingTours.length === 0 ? (
                <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', textAlign: 'center', padding: 'var(--spacing-4)' }}>
                  Keine anstehenden Einsätze
                </p>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
                  {upcomingTours.map((tour: any) => {
                    const startDate = tour.departure_at || tour.start_date
                    return (
                      <div
                        key={tour.id}
                        className="tour-list-item"
                        onClick={() => navigate(`/transport/tours/${tour.id}`)}
                      >
                        <div>
                          <p style={{ margin: '0 0 var(--spacing-1) 0', fontWeight: 'var(--font-weight-semibold)', color: 'var(--color-text-primary)', fontSize: 'var(--font-size-sm)' }}>
                            {tour.notes || `Tour #${tour.id}`}
                          </p>
                          {startDate && (
                            <p style={{ margin: 0, fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                              {new Date(startDate).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' })}
                            </p>
                          )}
                        </div>
                        <StatusBadge status={tour.status} label={TOUR_STATUS_LABELS[tour.status]} size="sm" />
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          </div>

          {/* Sidebar */}
          <div>
            {/* KM Stand */}
            <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
              <h2 className="detail-card__title">Kilometerstand</h2>
              <div style={{ textAlign: 'center', padding: 'var(--spacing-4)' }}>
                <div style={{
                  fontSize: 'var(--font-size-3xl)',
                  fontWeight: 'var(--font-weight-bold)',
                  color: 'var(--color-primary)',
                  fontVariantNumeric: 'tabular-nums',
                }}>
                  {completedTours.length > 0
                    ? Math.max(...completedTours.map((t: any) => t.km_end || 0)).toLocaleString('de-DE')
                    : '—'}
                </div>
                <div style={{
                  fontSize: 'var(--font-size-xs)',
                  color: 'var(--color-text-secondary)',
                  textTransform: 'uppercase',
                  marginTop: 'var(--spacing-1)',
                }}>km</div>
              </div>
            </div>

            {/* Maintenance History */}
            <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
              <h2 className="detail-card__title">Wartungshistorie</h2>
              {vehicle.dguv_last_check ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
                  <div className="log-entry">
                    <div className="log-entry__time">
                      {new Date(vehicle.dguv_last_check).toLocaleDateString('de-DE')}
                    </div>
                    <div className="log-entry__content">
                      <p className="log-entry__text">DGUV Prüfung</p>
                      <p className="log-entry__note">Letzte Prüfung bestanden</p>
                    </div>
                  </div>
                </div>
              ) : (
                <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', textAlign: 'center', padding: 'var(--spacing-4)' }}>
                  Keine Wartungseinträge vorhanden
                </p>
              )}
            </div>

            {/* Completed Tours History */}
            <div className="detail-card">
              <h2 className="detail-card__title">Abgeschlossene Touren ({completedTours.length})</h2>
              {completedTours.length === 0 ? (
                <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', textAlign: 'center', padding: 'var(--spacing-4)' }}>
                  Noch keine abgeschlossenen Touren
                </p>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
                  {completedTours.slice(0, 5).map((tour: any) => {
                    const km = (tour.km_end || 0) - (tour.km_start || 0)
                    return (
                      <div key={tour.id} className="log-entry">
                        <div className="log-entry__time">
                          {tour.departure_at
                            ? new Date(tour.departure_at).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' })
                            : '—'}
                        </div>
                        <div className="log-entry__content">
                          <p className="log-entry__text">{tour.notes || `Tour #${tour.id}`}</p>
                          {km > 0 && <p className="log-entry__note">{km} km</p>}
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    )
  }

  // ===== TOUR DETAIL VIEW =====
  const tour: TourDetail = tourRaw
    ? {
        id: tourRaw.id,
        project_name: tourRaw.notes || tourRaw.project_name || `Tour #${tourRaw.id}`,
        vehicle_name: tourRaw.vehicle_name || 'Unbekanntes Fahrzeug',
        driver: tourRaw.driver || tourRaw.driver_id || 'Nicht zugewiesen',
        status: tourRaw.status,
        start_date: tourRaw.departure_at || tourRaw.start_date || '',
        end_date: tourRaw.arrival_at || tourRaw.end_date || '',
        weight_kg: tourRaw.weight_kg || 0,
        volume_m3: tourRaw.volume_m3 || 0,
        vehicle_capacity_kg: tourRaw.vehicle_capacity_kg || 0,
        vehicle_capacity_m3: tourRaw.vehicle_capacity_m3 || 0,
        equipment: tourRaw.equipment || mockTourDetail.equipment,
        logs: tourRaw.logs || mockTourDetail.logs,
      }
    : mockTourDetail

  if (tourLoading) {
    return <div className="transport-detail-page">Lädt...</div>
  }

  const weightPercentage = tour.vehicle_capacity_kg ? (tour.weight_kg / tour.vehicle_capacity_kg) * 100 : 0
  const volumePercentage = tour.vehicle_capacity_m3 ? (tour.volume_m3 / tour.vehicle_capacity_m3) * 100 : 0

  const getCapacityStatus = (used: number, capacity: number) => {
    if (!capacity) return 'ok'
    const percentage = (used / capacity) * 100
    if (percentage <= 70) return 'ok'
    if (percentage <= 90) return 'warning'
    return 'danger'
  }

  const currentStatusIndex = tourStatusSteps.indexOf(tour.status)

  return (
    <div className="transport-detail-page">
      <div className="page-header">
        <div>
          <button
            className="btn btn--sm btn--secondary"
            onClick={() => navigate('/transport')}
            style={{ marginBottom: 'var(--spacing-3)' }}
          >
            Zurück
          </button>
          <h1 className="page-title">{tour.project_name}</h1>
          <p className="page-subtitle">Transport Details & Status</p>
        </div>
      </div>

      <div className="detail-grid">
        {/* Main Content */}
        <div>
          {/* Tour Info Card */}
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Transport Information</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <div className="detail-card__row-label">Projekt</div>
                <div className="detail-card__row-value">{tour.project_name}</div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Fahrzeug</div>
                <div className="detail-card__row-value">{tour.vehicle_name}</div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Fahrer</div>
                <div className="detail-card__row-value">{tour.driver}</div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Status</div>
                <div className="detail-card__row-value">
                  <StatusBadge status={tour.status} label={TOUR_STATUS_LABELS[tour.status]} />
                </div>
              </div>
              {tour.start_date && (
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Startzeit</div>
                  <div className="detail-card__row-value">
                    {new Date(tour.start_date).toLocaleString('de-DE')}
                  </div>
                </div>
              )}
              {tour.end_date && (
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Ankunft (erw.)</div>
                  <div className="detail-card__row-value">
                    {new Date(tour.end_date).toLocaleString('de-DE')}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Status Timeline */}
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Status Ablauf</h2>
            <div className="tour-timeline">
              {tourStatusSteps.map((step, index) => (
                <div key={step}>
                  <div className="tour-timeline__step">
                    <div
                      className={`tour-timeline__step-dot ${
                        index < currentStatusIndex
                          ? 'tour-timeline__step-dot--completed'
                          : index === currentStatusIndex
                            ? 'tour-timeline__step-dot--active'
                            : ''
                      }`}
                    >
                      {index < currentStatusIndex ? '?' : index === currentStatusIndex ? '?' : index + 1}
                    </div>
                    <div className="tour-timeline__step-label">{TOUR_STATUS_LABELS[step]}</div>
                  </div>
                  {index < tourStatusSteps.length - 1 && (
                    <div
                      className={`tour-timeline__connector ${
                        index < currentStatusIndex
                          ? 'tour-timeline__connector--completed'
                          : index === currentStatusIndex
                            ? 'tour-timeline__connector--active'
                            : ''
                      }`}
                    />
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Capacity Visualization */}
          {(tour.vehicle_capacity_kg > 0 || tour.vehicle_capacity_m3 > 0) && (
            <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
              <h2 className="detail-card__title">Auslastung</h2>
              <div className="detail-card__content" style={{ gap: 'var(--spacing-4)' }}>
                {tour.vehicle_capacity_kg > 0 && (
                  <div className="capacity-bar">
                    <div className="capacity-bar__label">
                      <span>Gewicht</span>
                      <span>
                        {tour.weight_kg}/{tour.vehicle_capacity_kg} kg ({weightPercentage.toFixed(0)}%)
                      </span>
                    </div>
                    <div className="capacity-bar__track">
                      <div
                        className={`capacity-bar__fill capacity-bar__fill--${getCapacityStatus(tour.weight_kg, tour.vehicle_capacity_kg)}`}
                        style={{ width: `${Math.min(weightPercentage, 100)}%` }}
                      />
                    </div>
                  </div>
                )}
                {tour.vehicle_capacity_m3 > 0 && (
                  <div className="capacity-bar">
                    <div className="capacity-bar__label">
                      <span>Volumen</span>
                      <span>
                        {tour.volume_m3}/{tour.vehicle_capacity_m3} m³ ({volumePercentage.toFixed(0)}%)
                      </span>
                    </div>
                    <div className="capacity-bar__track">
                      <div
                        className={`capacity-bar__fill capacity-bar__fill--${getCapacityStatus(tour.volume_m3, tour.vehicle_capacity_m3)}`}
                        style={{ width: `${Math.min(volumePercentage, 100)}%` }}
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Equipment List */}
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Ausrüstung ({tour.equipment.length} Posten)</h2>
            <div className="equipment-list">
              {tour.equipment.map((item) => (
                <div key={item.id}>
                  <div
                    className="equipment-item"
                    onClick={() => setExpandedEquipment(expandedEquipment === item.id ? null : item.id)}
                    style={{ cursor: 'pointer' }}
                  >
                    <div className="equipment-item__info">
                      <p className="equipment-item__name">{item.name}</p>
                      <p className="equipment-item__meta">
                        {item.weight_kg} kg | {item.volume_m3} m³
                      </p>
                    </div>
                    <div className="equipment-item__quantity">
                      <strong>x{item.quantity}</strong>
                      <span style={{ marginLeft: 'var(--spacing-2)', fontSize: '0.8em', opacity: 0.6 }}>
                        {expandedEquipment === item.id ? 'v' : '>'}
                      </span>
                    </div>
                  </div>
                  {expandedEquipment === item.id && (
                    <div style={{ padding: 'var(--spacing-3)', background: 'var(--color-bg-tertiary)', borderRadius: '0 0 var(--radius-md) var(--radius-md)', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                      <p style={{ margin: '0 0 var(--spacing-1) 0' }}>
                        <strong>Gewicht pro Einheit:</strong> {(item.weight_kg / item.quantity).toFixed(1)} kg
                      </p>
                      <p style={{ margin: '0 0 var(--spacing-1) 0' }}>
                        <strong>Volumen pro Einheit:</strong> {(item.volume_m3 / item.quantity).toFixed(2)} m³
                      </p>
                      <p style={{ margin: 0 }}>
                        <strong>Gesamtgewicht:</strong> {item.weight_kg} kg | <strong>Gesamtvolumen:</strong> {item.volume_m3} m³
                      </p>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div>
          {/* Driver Log */}
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Fahrtprotokoll</h2>
            <div className="driver-log">
              {tour.logs.map((entry) => (
                <div key={entry.id} className="log-entry">
                  <div className="log-entry__time">
                    {new Date(entry.timestamp).toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' })}
                  </div>
                  <div className="log-entry__content">
                    <p className="log-entry__text">{entry.action}</p>
                    <p className="log-entry__note">{entry.notes}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="detail-card">
            <h2 className="detail-card__title">Aktionen</h2>
            <div className="action-buttons">
              {tour.status === 'planned' && (
                <>
                  <button className="btn btn--primary">Transport starten</button>
                  <button className="btn btn--secondary">Bearbeiten</button>
                </>
              )}
              {tour.status === 'loading' && (
                <>
                  <button className="btn btn--primary">Fahrt starten</button>
                  <button className="btn btn--secondary">Ausrüstung hinzufügen</button>
                </>
              )}
              {tour.status === 'in_transit' && (
                <>
                  <button className="btn btn--primary">Geliefert markieren</button>
                  <button className="btn btn--secondary">Notiz hinzufügen</button>
                </>
              )}
              {tour.status === 'delivered' && (
                <>
                  <button className="btn btn--primary">Abschließen</button>
                  <button className="btn btn--secondary">Bericht anzeigen</button>
                </>
              )}
              {tour.status === 'completed' && (
                <>
                  <button className="btn btn--secondary">Bericht anzeigen</button>
                  <button className="btn btn--secondary">Drucken</button>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default TransportDetailPage
