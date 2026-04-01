import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { transportApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import './Transport.scss'

const VEHICLE_STATUS_LABELS: Record<string, string> = {
  available: 'Verfuegbar',
  in_use: 'Im Einsatz',
  maintenance: 'Wartung',
}

const VEHICLE_TYPE_LABELS: Record<string, string> = {
  lkw: 'LKW',
  transporter: 'Transporter',
  anhaenger: 'Anhaenger',
  pkw: 'PKW',
}

function VehicleDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()

  const { data: vehicle, isLoading, error } = useQuery({
    queryKey: ['vehicle', id],
    queryFn: () => transportApi.getVehicle(id!),
    enabled: !!id,
    staleTime: 1000 * 60 * 5,
    retry: 1,
  })

  const deleteMutation = useMutation({
    mutationFn: () => transportApi.deleteVehicle(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vehicles'] })
      addNotification('Das Fahrzeug wurde erfolgreich geloescht.', 'success', { title: 'Fahrzeug geloescht' })
      navigate('/transport')
    },
    onError: (err: any) => {
      addNotification(
        `Fahrzeug konnte nicht geloescht werden: ${err.message || 'Unbekannter Fehler'}`,
        'error',
        { title: 'Fehler' }
      )
    },
  })

  const handleDelete = () => {
    if (window.confirm('Moechten Sie dieses Fahrzeug wirklich loeschen? Diese Aktion kann nicht rueckgaengig gemacht werden.')) {
      deleteMutation.mutate()
    }
  }

  if (isLoading) {
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

  if (error || !vehicle) {
    return (
      <div className="transport-detail-page">
        <div className="page-header">
          <div>
            <button
              className="btn btn--sm btn--secondary"
              onClick={() => navigate('/transport')}
              style={{ marginBottom: 'var(--spacing-3)' }}
            >
              Zurueck
            </button>
            <h1 className="page-title">Fahrzeug nicht gefunden</h1>
            <p className="page-subtitle">
              {error ? String(error) : 'Das angeforderte Fahrzeug existiert nicht.'}
            </p>
          </div>
        </div>
      </div>
    )
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
            Zurueck
          </button>
          <h1 className="page-title">{vehicle.name}</h1>
          <p className="page-subtitle">Fahrzeugdetails und Informationen</p>
        </div>
        <StatusBadge
          status={vehicle.status}
          label={VEHICLE_STATUS_LABELS[vehicle.status] || vehicle.status}
          size="lg"
        />
      </div>

      <div className="detail-grid">
        <div>
          <div className="detail-card">
            <h2 className="detail-card__title">Fahrzeugdaten</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <div className="detail-card__row-label">Name</div>
                <div className="detail-card__row-value">{vehicle.name}</div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Kennzeichen</div>
                <div className="detail-card__row-value">
                  {vehicle.license_plate ? (
                    <span className="license-plate">{vehicle.license_plate}</span>
                  ) : (
                    <span style={{ color: 'var(--color-text-secondary)' }}>---</span>
                  )}
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Typ</div>
                <div className="detail-card__row-value">
                  <span className="vehicle-type-tag">
                    {VEHICLE_TYPE_LABELS[vehicle.vehicle_type] || vehicle.vehicle_type || '---'}
                  </span>
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Max. Gewicht</div>
                <div className="detail-card__row-value">
                  {vehicle.capacity_kg
                    ? `${vehicle.capacity_kg.toLocaleString('de-DE')} kg`
                    : '---'}
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Max. Volumen</div>
                <div className="detail-card__row-value">
                  {vehicle.capacity_m3 ? `${vehicle.capacity_m3} m3` : '---'}
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Status</div>
                <div className="detail-card__row-value">
                  <StatusBadge
                    status={vehicle.status}
                    label={VEHICLE_STATUS_LABELS[vehicle.status] || vehicle.status}
                  />
                </div>
              </div>
              {vehicle.loading_length && (
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Ladeflaeche</div>
                  <div className="detail-card__row-value">
                    {vehicle.loading_length}m x {vehicle.loading_width || '---'}m x {vehicle.loading_height || '---'}m
                  </div>
                </div>
              )}
              {vehicle.license_class && (
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Fuehrerschein</div>
                  <div className="detail-card__row-value">Klasse {vehicle.license_class}</div>
                </div>
              )}
              {vehicle.notes && (
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Bemerkungen</div>
                  <div className="detail-card__row-value">{vehicle.notes}</div>
                </div>
              )}
            </div>
          </div>
        </div>

        <div>
          <div className="detail-card">
            <h2 className="detail-card__title">Aktionen</h2>
            <div className="action-buttons" style={{ flexDirection: 'column' }}>
              <button
                className="btn btn--secondary"
                onClick={() => navigate('/transport')}
              >
                Bearbeiten
              </button>
              <button
                className="btn btn--danger"
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
              >
                {deleteMutation.isPending ? 'Loeschen...' : 'Fahrzeug loeschen'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default VehicleDetailPage
