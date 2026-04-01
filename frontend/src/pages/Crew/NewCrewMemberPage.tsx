import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { crewApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { TextArea } from '../../components/Form/TextArea'
import '../Equipment/Equipment.scss'

const ROLE_OPTIONS = [
  { value: 'technician', label: 'Techniker' },
  { value: 'driver', label: 'Fahrer' },
  { value: 'projectmanager', label: 'Projektleiter' },
  { value: 'intern', label: 'Praktikant' },
  { value: 'freelancer', label: 'Freelancer' },
]

interface CrewFormData {
  first_name: string
  last_name: string
  email: string
  phone: string
  role: string
  hourly_rate: number
  notes: string
}

function NewCrewMemberPage() {
  const navigate = useNavigate()
  const { addNotification } = useNotificationStore()

  const [formData, setFormData] = useState<CrewFormData>({
    first_name: '',
    last_name: '',
    email: '',
    phone: '',
    role: 'technician',
    hourly_rate: 0,
    notes: '',
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  const { mutate: saveMember, isPending } = useMutation({
    mutationFn: () =>
      crewApi.createMember({
        first_name: formData.first_name,
        last_name: formData.last_name,
        email: formData.email,
        phone: formData.phone || undefined,
        role: formData.role,
        hourly_rate: formData.hourly_rate || undefined,
        notes: formData.notes || undefined,
      }),
    onSuccess: () => {
      addNotification('Mitarbeiter erfolgreich angelegt', 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
      navigate('/crew')
    },
    onError: (error: unknown) => {
      const msg =
        (error as any)?.response?.data?.error ||
        (error as any)?.response?.data?.message ||
        'Fehler beim Anlegen des Mitarbeiters'
      setErrors({ submit: msg })
    },
  })

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}
    if (!formData.first_name.trim()) {
      newErrors.first_name = 'Vorname ist erforderlich'
    }
    if (!formData.last_name.trim()) {
      newErrors.last_name = 'Nachname ist erforderlich'
    }
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleChange = (field: keyof CrewFormData, value: string | number) => {
    setFormData((prev) => ({ ...prev, [field]: value }))
    if (errors[field]) {
      setErrors((prev) => {
        const updated = { ...prev }
        delete updated[field]
        return updated
      })
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (validateForm()) {
      saveMember()
    }
  }

  return (
    <div className="equipment-form-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Neuer Mitarbeiter</h1>
          <p className="page-subtitle">
            Legen Sie einen neuen Mitarbeiter mit Kontaktdaten und Rolle an
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="form-section">
        {errors.submit && (
          <div className="error-message" role="alert">
            {errors.submit}
          </div>
        )}

        <h2 className="form-section__title">Persoenliche Daten</h2>
        <div className="form-section__grid">
          <Input
            label="Vorname *"
            value={formData.first_name}
            onChange={(e) => handleChange('first_name', e.target.value)}
            placeholder="z.B. Max"
            error={errors.first_name}
          />
          <Input
            label="Nachname *"
            value={formData.last_name}
            onChange={(e) => handleChange('last_name', e.target.value)}
            placeholder="z.B. Mustermann"
            error={errors.last_name}
          />
          <Input
            label="E-Mail"
            type="email"
            value={formData.email}
            onChange={(e) => handleChange('email', e.target.value)}
            placeholder="max@beispiel.de"
          />
          <Input
            label="Telefon"
            value={formData.phone}
            onChange={(e) => handleChange('phone', e.target.value)}
            placeholder="+49 170 1234567"
          />
        </div>

        <h2 className="form-section__title">Rolle & Konditionen</h2>
        <div className="form-section__grid">
          <Select
            label="Rolle"
            options={ROLE_OPTIONS}
            value={formData.role}
            onChange={(e) => handleChange('role', e.target.value)}
          />
          <Input
            label="Stundensatz (EUR)"
            type="number"
            value={formData.hourly_rate || ''}
            onChange={(e) =>
              handleChange('hourly_rate', parseFloat(e.target.value) || 0)
            }
            step="0.01"
            min="0"
            placeholder="0.00"
          />
        </div>

        <TextArea
          label="Notizen"
          value={formData.notes}
          onChange={(e) => handleChange('notes', e.target.value)}
          placeholder="Interne Notizen zum Mitarbeiter..."
          rows={4}
        />

        <div className="form-section__footer">
          <button
            type="button"
            className="btn btn--secondary"
            onClick={() => navigate('/crew')}
          >
            Abbrechen
          </button>
          <button
            type="submit"
            className="btn btn--primary"
            disabled={isPending}
          >
            {isPending ? 'Wird gespeichert...' : 'Mitarbeiter anlegen'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default NewCrewMemberPage
