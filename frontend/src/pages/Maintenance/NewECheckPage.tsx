import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { maintenanceApi, equipmentApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { TextArea } from '../../components/Form/TextArea'
import './Maintenance.scss'

const RESULT_OPTIONS = [
  { value: 'passed', label: 'Bestanden' },
  { value: 'failed', label: 'Nicht bestanden' },
  { value: 'conditional', label: 'Bedingt bestanden' },
]

function todayISO(): string {
  return new Date().toISOString().slice(0, 10)
}

interface ECheckFormData {
  equipment_id: string
  check_date: string
  result: string
  next_check_date: string
  inspector_name: string
  notes: string
}

function NewECheckPage() {
  const navigate = useNavigate()
  const { addNotification } = useNotificationStore()

  const [formData, setFormData] = useState<ECheckFormData>({
    equipment_id: '',
    check_date: todayISO(),
    result: 'passed',
    next_check_date: '',
    inspector_name: '',
    notes: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})

  const { data: equipmentData } = useQuery({
    queryKey: ['equipment-list'],
    queryFn: () => equipmentApi.list({ limit: 200 }),
    staleTime: 1000 * 60 * 10,
  })

  const equipmentItems: Array<{ id: string; name: string }> =
    equipmentData?.data || equipmentData?.items || (Array.isArray(equipmentData) ? equipmentData : [])

  const equipmentOptions = equipmentItems.map((e) => ({
    value: e.id,
    label: e.name,
  }))

  const { mutate: createECheck, isPending } = useMutation({
    mutationFn: () => maintenanceApi.createElectricalTest(formData),
    onSuccess: () => {
      addNotification('E-Check erfolgreich gespeichert', 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
      navigate('/maintenance')
    },
    onError: (error: unknown) => {
      const msg =
        (error as any)?.response?.data?.error ||
        (error as any)?.response?.data?.message ||
        'Fehler beim Speichern des E-Checks'
      setErrors({ submit: msg })
      addNotification(msg, 'error', { title: 'Fehler' })
    },
  })

  const handleChange = (field: keyof ECheckFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }))
    if (errors[field]) {
      setErrors((prev) => {
        const updated = { ...prev }
        delete updated[field]
        return updated
      })
    }
  }

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {}
    if (!formData.equipment_id) {
      newErrors.equipment_id = 'Bitte ein Geraet auswaehlen'
    }
    if (!formData.check_date) {
      newErrors.check_date = 'Pruefdatum ist erforderlich'
    }
    if (!formData.result) {
      newErrors.result = 'Ergebnis ist erforderlich'
    }
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (validate()) {
      createECheck()
    }
  }

  const resultColor = (result: string): string => {
    switch (result) {
      case 'passed':
        return 'var(--color-success)'
      case 'failed':
        return 'var(--color-danger)'
      case 'conditional':
        return 'var(--color-warning)'
      default:
        return 'var(--color-text-secondary)'
    }
  }

  return (
    <div className="maintenance-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">E-Check durchfuehren</h1>
          <p className="page-subtitle">
            Elektrische Sicherheitspruefung nach DGUV Vorschrift 3
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="form-section">
        {errors.submit && (
          <div className="error-message" role="alert">
            {errors.submit}
          </div>
        )}

        <h2 className="form-section__title">Pruefungsdaten</h2>
        <div className="form-section__grid">
          <Select
            label="Geraet *"
            options={equipmentOptions}
            value={formData.equipment_id}
            onChange={(e) => handleChange('equipment_id', e.target.value)}
            placeholder="Geraet auswaehlen"
            error={errors.equipment_id}
          />
          <Input
            label="Pruefdatum *"
            type="date"
            value={formData.check_date}
            onChange={(e) => handleChange('check_date', e.target.value)}
            error={errors.check_date}
          />
          <Select
            label="Ergebnis *"
            options={RESULT_OPTIONS}
            value={formData.result}
            onChange={(e) => handleChange('result', e.target.value)}
            error={errors.result}
          />
          <Input
            label="Naechster Prueftermin"
            type="date"
            value={formData.next_check_date}
            onChange={(e) => handleChange('next_check_date', e.target.value)}
          />
          <Input
            label="Pruefer"
            value={formData.inspector_name}
            onChange={(e) => handleChange('inspector_name', e.target.value)}
            placeholder="Name des Pruefers"
          />
        </div>

        {formData.result && (
          <div
            style={{
              padding: 'var(--spacing-3) var(--spacing-4)',
              borderRadius: 'var(--radius-md)',
              border: `1px solid ${resultColor(formData.result)}`,
              backgroundColor: `${resultColor(formData.result)}10`,
              marginBottom: 'var(--spacing-4)',
            }}
          >
            <span
              style={{
                fontWeight: 'var(--font-weight-semibold)' as any,
                color: resultColor(formData.result),
                fontSize: 'var(--font-size-sm)',
              }}
            >
              {formData.result === 'passed' && 'Bestanden -- Geraet ist sicher fuer den Betrieb'}
              {formData.result === 'failed' && 'Nicht bestanden -- Geraet darf NICHT eingesetzt werden'}
              {formData.result === 'conditional' && 'Bedingt bestanden -- Geraet unter Auflagen einsetzbar'}
            </span>
          </div>
        )}

        <TextArea
          label="Bemerkungen"
          value={formData.notes}
          onChange={(e) => handleChange('notes', e.target.value)}
          placeholder="Zusaetzliche Bemerkungen zur Pruefung..."
          rows={4}
        />

        <div className="form-section__footer">
          <button
            type="button"
            className="btn btn--secondary"
            onClick={() => navigate('/maintenance')}
          >
            Abbrechen
          </button>
          <button
            type="submit"
            className="btn btn--primary"
            disabled={isPending}
          >
            {isPending ? 'Wird gespeichert...' : 'E-Check speichern'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default NewECheckPage
