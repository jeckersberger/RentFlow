import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { transportApi, projectApi, crewApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { Input } from '../../components/Form/Input'
import './Transport.scss'

interface TourFormData {
  project_id: string
  pickup_date: string
  delivery_date: string
  pickup_address: string
  delivery_address: string
  driver_id: string
  vehicle_id: string
  notes: string
}

const emptyForm: TourFormData = {
  project_id: '',
  pickup_date: '',
  delivery_date: '',
  pickup_address: '',
  delivery_address: '',
  driver_id: '',
  vehicle_id: '',
  notes: '',
}

function NewTourPage() {
  const navigate = useNavigate()
  const { addNotification } = useNotificationStore()
  const [form, setForm] = useState<TourFormData>(emptyForm)

  const { data: vehiclesData } = useQuery({
    queryKey: ['vehicles-for-tour'],
    queryFn: () => transportApi.listVehicles(1, 200),
    staleTime: 1000 * 60 * 5,
  })

  const { data: projectsData } = useQuery({
    queryKey: ['projects-for-tour'],
    queryFn: () => projectApi.list(1, 200),
    staleTime: 1000 * 60 * 5,
  })

  const { data: driversData } = useQuery({
    queryKey: ['crew-drivers'],
    queryFn: () => crewApi.listMembers({ per_page: 200 }),
    staleTime: 1000 * 60 * 5,
  })

  const vehiclesList: { id: string; name: string; license_plate: string; status: string }[] =
    vehiclesData?.items || vehiclesData?.data || (Array.isArray(vehiclesData) ? vehiclesData : [])

  const projectsList: { id: string; name: string }[] =
    projectsData?.items || projectsData?.data || (Array.isArray(projectsData) ? projectsData : [])

  const driversList: { id: string; first_name: string; last_name: string }[] =
    driversData?.data || driversData?.items || (Array.isArray(driversData) ? driversData : [])

  const createMutation = useMutation({
    mutationFn: (data: object) => transportApi.createTour(data),
    onSuccess: () => {
      addNotification('Der Transport wurde erfolgreich angelegt.', 'success', { title: 'Transport erstellt' })
      navigate('/transport')
    },
    onError: (err: any) => {
      addNotification(
        `Transport konnte nicht erstellt werden: ${err.message || 'Unbekannter Fehler'}`,
        'error',
        { title: 'Fehler' }
      )
    },
  })

  const handleChange = (field: keyof TourFormData, value: string) => {
    setForm(prev => ({ ...prev, [field]: value }))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!form.vehicle_id) {
      addNotification('Bitte waehlen Sie ein Fahrzeug aus.', 'error', { title: 'Validierung' })
      return
    }

    createMutation.mutate({
      vehicle_id: form.vehicle_id,
      project_id: form.project_id || undefined,
      driver_id: form.driver_id || undefined,
      departure_at: form.pickup_date || undefined,
      arrival_at: form.delivery_date || undefined,
      start_date: form.pickup_date || undefined,
      end_date: form.delivery_date || undefined,
      pickup_address: form.pickup_address || undefined,
      delivery_address: form.delivery_address || undefined,
      notes: form.notes || undefined,
      status: 'planned',
    })
  }

  return (
    <div className="transport-page">
      <div className="page-header">
        <div>
          <button
            className="btn btn--sm btn--secondary"
            onClick={() => navigate('/transport')}
            style={{ marginBottom: 'var(--spacing-3)' }}
          >
            Zurueck
          </button>
          <h1 className="page-title">Neue Tour erstellen</h1>
          <p className="page-subtitle">Planen Sie einen neuen Transport mit Fahrzeug, Route und Zeitfenster.</p>
        </div>
      </div>

      <div className="detail-card">
        <h2 className="detail-card__title">Transportdaten</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-grid">
            <div className="form-group">
              <label className="form-label">Fahrzeug *</label>
              <select
                className="form-input"
                value={form.vehicle_id}
                onChange={(e) => handleChange('vehicle_id', e.target.value)}
              >
                <option value="">-- Fahrzeug waehlen --</option>
                {vehiclesList.map(v => (
                  <option key={v.id} value={v.id}>
                    {v.name} ({v.license_plate})
                  </option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Projekt (optional)</label>
              <select
                className="form-input"
                value={form.project_id}
                onChange={(e) => handleChange('project_id', e.target.value)}
              >
                <option value="">-- Projekt waehlen --</option>
                {projectsList.map(p => (
                  <option key={p.id} value={p.id}>{p.name}</option>
                ))}
              </select>
            </div>

            <Input
              label="Abholdatum"
              type="datetime-local"
              value={form.pickup_date}
              onChange={(e) => handleChange('pickup_date', e.target.value)}
            />

            <Input
              label="Lieferdatum"
              type="datetime-local"
              value={form.delivery_date}
              onChange={(e) => handleChange('delivery_date', e.target.value)}
            />

            <Input
              label="Abholadresse"
              value={form.pickup_address}
              onChange={(e) => handleChange('pickup_address', e.target.value)}
              placeholder="z.B. Lager Muenchen, Industriestr. 5"
            />

            <Input
              label="Lieferadresse"
              value={form.delivery_address}
              onChange={(e) => handleChange('delivery_address', e.target.value)}
              placeholder="z.B. Olympiahalle, Spiridon-Louis-Ring 21"
            />

            <div className="form-group">
              <label className="form-label">Fahrer (optional)</label>
              <select
                className="form-input"
                value={form.driver_id}
                onChange={(e) => handleChange('driver_id', e.target.value)}
              >
                <option value="">-- Fahrer waehlen --</option>
                {driversList.map(d => (
                  <option key={d.id} value={d.id}>
                    {d.first_name} {d.last_name}
                  </option>
                ))}
              </select>
            </div>

            <div className="form-group form-group--full">
              <label className="form-label">Bemerkungen</label>
              <textarea
                className="form-input"
                value={form.notes}
                onChange={(e) => handleChange('notes', e.target.value)}
                placeholder="z.B. Anlieferung bis 08:00 Uhr, Rampe 3"
                rows={3}
                style={{ resize: 'vertical' }}
              />
            </div>
          </div>

          <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end', marginTop: 'var(--spacing-6)' }}>
            <button
              type="button"
              className="btn btn--secondary"
              onClick={() => navigate('/transport')}
            >
              Abbrechen
            </button>
            <button
              type="submit"
              className="btn btn--primary"
              disabled={createMutation.isPending}
            >
              {createMutation.isPending ? 'Speichern...' : 'Transport anlegen'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default NewTourPage
