import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { transportApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { Input } from '../../components/Form/Input'
import './Transport.scss'

const VEHICLE_TYPES = [
  { value: 'lkw', label: 'LKW' },
  { value: 'transporter', label: 'Transporter' },
  { value: 'anhaenger', label: 'Anhaenger' },
  { value: 'pkw', label: 'PKW' },
]

interface VehicleFormData {
  name: string
  license_plate: string
  vehicle_type: string
  max_weight_kg: string
  max_volume_m3: string
  notes: string
}

const emptyForm: VehicleFormData = {
  name: '',
  license_plate: '',
  vehicle_type: 'transporter',
  max_weight_kg: '',
  max_volume_m3: '',
  notes: '',
}

function NewVehiclePage() {
  const navigate = useNavigate()
  const { addNotification } = useNotificationStore()
  const [form, setForm] = useState<VehicleFormData>(emptyForm)

  const createMutation = useMutation({
    mutationFn: (data: object) => transportApi.createVehicle(data),
    onSuccess: () => {
      addNotification('Das Fahrzeug wurde erfolgreich angelegt.', 'success', { title: 'Fahrzeug erstellt' })
      navigate('/transport')
    },
    onError: (err: any) => {
      addNotification(
        `Fahrzeug konnte nicht erstellt werden: ${err.message || 'Unbekannter Fehler'}`,
        'error',
        { title: 'Fehler' }
      )
    },
  })

  const handleChange = (field: keyof VehicleFormData, value: string) => {
    setForm(prev => ({ ...prev, [field]: value }))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!form.name.trim()) {
      addNotification('Bitte geben Sie einen Fahrzeugnamen ein.', 'error', { title: 'Validierung' })
      return
    }

    createMutation.mutate({
      name: form.name.trim(),
      license_plate: form.license_plate.trim().toUpperCase(),
      vehicle_type: form.vehicle_type,
      capacity_kg: form.max_weight_kg ? parseFloat(form.max_weight_kg) : 0,
      capacity_m3: form.max_volume_m3 ? parseFloat(form.max_volume_m3) : 0,
      notes: form.notes || undefined,
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
          <h1 className="page-title">Neues Fahrzeug anlegen</h1>
          <p className="page-subtitle">Erfassen Sie ein neues Fahrzeug mit Typ, Kennzeichen und technischen Daten.</p>
        </div>
      </div>

      <div className="detail-card">
        <h2 className="detail-card__title">Fahrzeugdaten</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-grid">
            <Input
              label="Fahrzeugname *"
              value={form.name}
              onChange={(e) => handleChange('name', e.target.value)}
              placeholder="z.B. MAN TGX 18t"
            />

            <Input
              label="Kennzeichen"
              value={form.license_plate}
              onChange={(e) => handleChange('license_plate', e.target.value.toUpperCase())}
              placeholder="z.B. M-AB 1234"
            />

            <div className="form-group">
              <label className="form-label">Fahrzeugtyp</label>
              <select
                className="form-input"
                value={form.vehicle_type}
                onChange={(e) => handleChange('vehicle_type', e.target.value)}
              >
                {VEHICLE_TYPES.map(t => (
                  <option key={t.value} value={t.value}>{t.label}</option>
                ))}
              </select>
            </div>

            <Input
              label="Max. Gewicht (kg)"
              type="number"
              value={form.max_weight_kg}
              onChange={(e) => handleChange('max_weight_kg', e.target.value)}
              placeholder="z.B. 18000"
              min="0"
            />

            <Input
              label="Max. Volumen (m3)"
              type="number"
              value={form.max_volume_m3}
              onChange={(e) => handleChange('max_volume_m3', e.target.value)}
              placeholder="z.B. 52"
              min="0"
              step="0.1"
            />

            <div className="form-group form-group--full">
              <label className="form-label">Bemerkungen</label>
              <textarea
                className="form-input"
                value={form.notes}
                onChange={(e) => handleChange('notes', e.target.value)}
                placeholder="z.B. Ladebordwand vorhanden, TUeV bis 03/2027"
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
              {createMutation.isPending ? 'Speichern...' : 'Fahrzeug anlegen'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default NewVehiclePage
