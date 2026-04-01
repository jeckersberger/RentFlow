import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { maintenanceApi, equipmentApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { TextArea } from '../../components/Form/TextArea'
import './Maintenance.scss'

const PRIORITY_OPTIONS = [
  { value: 'low', label: 'Niedrig' },
  { value: 'medium', label: 'Mittel' },
  { value: 'high', label: 'Hoch' },
  { value: 'critical', label: 'Kritisch' },
]

interface TaskFormData {
  equipment_id: string
  title: string
  description: string
  priority: string
  due_date: string
  assigned_to: string
}

function NewMaintenanceTaskPage() {
  const navigate = useNavigate()
  const { addNotification } = useNotificationStore()

  const [formData, setFormData] = useState<TaskFormData>({
    equipment_id: '',
    title: '',
    description: '',
    priority: 'medium',
    due_date: '',
    assigned_to: '',
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

  const { mutate: createTask, isPending } = useMutation({
    mutationFn: () => maintenanceApi.createTask(formData),
    onSuccess: () => {
      addNotification('Wartungsaufgabe erfolgreich erstellt', 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
      navigate('/maintenance')
    },
    onError: (error: unknown) => {
      const msg =
        (error as any)?.response?.data?.error ||
        (error as any)?.response?.data?.message ||
        'Fehler beim Erstellen der Aufgabe'
      setErrors({ submit: msg })
      addNotification(msg, 'error', { title: 'Fehler' })
    },
  })

  const handleChange = (field: keyof TaskFormData, value: string) => {
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
    if (!formData.title.trim()) {
      newErrors.title = 'Titel ist erforderlich'
    }
    if (!formData.equipment_id) {
      newErrors.equipment_id = 'Bitte ein Geraet auswaehlen'
    }
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (validate()) {
      createTask()
    }
  }

  return (
    <div className="maintenance-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Neue Wartungsaufgabe</h1>
          <p className="page-subtitle">
            Erstellen Sie eine neue Wartungs- oder Reparaturaufgabe
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="form-section">
        {errors.submit && (
          <div className="error-message" role="alert">
            {errors.submit}
          </div>
        )}

        <h2 className="form-section__title">Aufgabendetails</h2>
        <div className="form-section__grid">
          <Select
            label="Geraet *"
            options={equipmentOptions}
            value={formData.equipment_id}
            onChange={(e) => handleChange('equipment_id', e.target.value)}
            placeholder="Geraet auswaehlen"
            error={errors.equipment_id}
          />
          <Select
            label="Prioritaet"
            options={PRIORITY_OPTIONS}
            value={formData.priority}
            onChange={(e) => handleChange('priority', e.target.value)}
          />
          <Input
            label="Titel *"
            value={formData.title}
            onChange={(e) => handleChange('title', e.target.value)}
            placeholder="z.B. Jaehrliche Inspektion"
            error={errors.title}
          />
          <Input
            label="Faellig am"
            type="date"
            value={formData.due_date}
            onChange={(e) => handleChange('due_date', e.target.value)}
          />
          <Input
            label="Zugewiesen an"
            value={formData.assigned_to}
            onChange={(e) => handleChange('assigned_to', e.target.value)}
            placeholder="Name des Mitarbeiters"
          />
        </div>

        <TextArea
          label="Beschreibung"
          value={formData.description}
          onChange={(e) => handleChange('description', e.target.value)}
          placeholder="Detaillierte Beschreibung der Wartungsaufgabe..."
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
            {isPending ? 'Wird erstellt...' : 'Aufgabe erstellen'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default NewMaintenanceTaskPage
