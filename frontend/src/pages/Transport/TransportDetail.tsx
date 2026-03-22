import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
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

// Mock data
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
    {
      id: '1',
      timestamp: '2026-03-22T08:15:00Z',
      action: 'Tour gestartet',
      notes: 'Fahrzeug verlässt Depot mit vollständiger Ausrüstung',
    },
    {
      id: '2',
      timestamp: '2026-03-22T09:30:00Z',
      action: 'Beladung abgeschlossen',
      notes: 'Gesamtgewicht: 1200 kg, Volumen: 9 m³',
    },
    {
      id: '3',
      timestamp: '2026-03-22T11:45:00Z',
      action: 'An Einsatzort angekommen',
      notes: 'Marienplatz München - Aufbau wird begonnen',
    },
  ],
}

const tourStatusSteps = ['planned', 'loading', 'in_transit', 'delivered', 'completed']

function TransportDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [expandedEquipment, setExpandedEquipment] = useState<string | null>(null)

  const { data: tour = mockTourDetail, isLoading } = useQuery({
    queryKey: ['tour', id],
    queryFn: async () => mockTourDetail,
    enabled: !!id,
  })

  if (isLoading) {
    return <div className="transport-detail-page">Lädt...</div>
  }

  const weightPercentage = (tour.weight_kg / tour.vehicle_capacity_kg) * 100
  const volumePercentage = (tour.volume_m3 / tour.vehicle_capacity_m3) * 100

  const getCapacityStatus = (used: number, capacity: number) => {
    const percentage = (used / capacity) * 100
    if (percentage <= 70) return 'ok'
    if (percentage <= 90) return 'warning'
    return 'danger'
  }

  const currentStatusIndex = tourStatusSteps.indexOf(tour.status)

  const statusLabels = {
    planned: 'Geplant',
    loading: 'Beladung',
    in_transit: 'Unterwegs',
    delivered: 'Geliefert',
    completed: 'Abgeschlossen',
  }

  return (
    <div className="transport-detail-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{tour.project_name}</h1>
          <p className="page-subtitle">Transport Details & Status</p>
        </div>
        <button
          className="btn btn--secondary"
          onClick={() => navigate('/transport')}
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
        >
          ← Zurück
        </button>
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
                  <StatusBadge status={tour.status} />
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Startzeit</div>
                <div className="detail-card__row-value">
                  {new Date(tour.start_date).toLocaleString('de-DE')}
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Endzeit (erwartet)</div>
                <div className="detail-card__row-value">
                  {new Date(tour.end_date).toLocaleString('de-DE')}
                </div>
              </div>
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
                      {index < currentStatusIndex ? '✓' : index === currentStatusIndex ? '●' : index + 1}
                    </div>
                    <div className="tour-timeline__step-label">{statusLabels[step as keyof typeof statusLabels]}</div>
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
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Auslastung</h2>
            <div className="detail-card__content" style={{ gap: 'var(--spacing-4)' }}>
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
            </div>
          </div>

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
                        {item.weight_kg} kg • {item.volume_m3} m³
                      </p>
                    </div>
                    <div className="equipment-item__quantity">
                      <strong>x{item.quantity}</strong>
                      <span style={{ marginLeft: 'var(--spacing-2)', fontSize: '0.8em', opacity: 0.6 }}>
                        {expandedEquipment === item.id ? '▼' : '▶'}
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
                        <strong>Gesamtgewicht:</strong> {item.weight_kg} kg • <strong>Gesamtvolumen:</strong> {item.volume_m3} m³
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
                  <button className="btn btn--primary">🚚 Transport starten</button>
                  <button className="btn btn--secondary">✏️ Bearbeiten</button>
                </>
              )}
              {tour.status === 'loading' && (
                <>
                  <button className="btn btn--primary">🚗 Fahrt starten</button>
                  <button className="btn btn--secondary">➕ Ausrüstung hinzufügen</button>
                </>
              )}
              {tour.status === 'in_transit' && (
                <>
                  <button className="btn btn--primary">✅ Geliefert markieren</button>
                  <button className="btn btn--secondary">📝 Notiz hinzufügen</button>
                </>
              )}
              {tour.status === 'delivered' && (
                <>
                  <button className="btn btn--primary">🏁 Abschließen</button>
                  <button className="btn btn--secondary">📋 Bericht anzeigen</button>
                </>
              )}
              {tour.status === 'completed' && (
                <>
                  <button className="btn btn--secondary">📋 Bericht anzeigen</button>
                  <button className="btn btn--secondary">🖨️ Drucken</button>
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
