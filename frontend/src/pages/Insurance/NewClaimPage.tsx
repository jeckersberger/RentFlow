import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { insuranceApi, equipmentApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { TextArea } from '../../components/Form/TextArea'
import '../Equipment/Equipment.scss'

interface ClaimFormData {
  policy_id: string
  equipment_id: string
  incident_date: string
  description: string
  damage_amount: number
}

function todayISO(): string {
  return new Date().toISOString().split('T')[0]
}

function NewClaimPage() {
  const navigate = useNavigate()
  const { addNotification } = useNotificationStore()

  const [formData, setFormData] = useState<ClaimFormData>({
    policy_id: '',
    equipment_id: '',
    incident_date: todayISO(),
    description: '',
    damage_amount: 0,
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  const { data: policiesRaw } = useQuery({
    queryKey: ['insurance-policies'],
    queryFn: () => insuranceApi.listPolicies(),
    staleTime: 1000 * 60 * 5,
  })

  const { data: equipmentRaw } = useQuery({
    queryKey: ['equipment-list-for-claim'],
    queryFn: () => equipmentApi.list({ limit: 200 }),
    staleTime: 1000 * 60 * 5,
  })

  const policies = Array.isArray(policiesRaw)
    ? policiesRaw
    : policiesRaw?.data ?? policiesRaw?.items ?? []

  const equipmentItems = Array.isArray(equipmentRaw)
    ? equipmentRaw
    : equipmentRaw?.data ?? equipmentRaw?.items ?? []

  const policyOptions = policies.map((p: any) => ({
    value: p.id,
    label: p.name || p.policy_number || p.id,
  }))

  const equipmentOptions = equipmentItems.map((e: any) => ({
    value: e.id,
    label: e.name || e.sku || e.id,
  }))

  const { mutate: saveClaim, isPending } = useMutation({
    mutationFn: () =>
      insuranceApi.createClaim({
        policy_id: formData.policy_id,
        equipment_id: formData.equipment_id || undefined,
        incident_date: formData.incident_date,
        description: formData.description,
        damage_amount: Math.round(formData.damage_amount * 100),
      }),
    onSuccess: () => {
      addNotification('Schadensfall erfolgreich gemeldet', 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
      navigate('/insurance')
    },
    onError: (error: unknown) => {
      const msg =
        (error as any)?.response?.data?.error ||
        (error as any)?.response?.data?.message ||
        'Fehler beim Melden des Schadensfalls'
      setErrors({ submit: msg })
    },
  })

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}
    if (!formData.policy_id) {
      newErrors.policy_id = 'Police ist erforderlich'
    }
    if (!formData.description.trim()) {
      newErrors.description = 'Beschreibung ist erforderlich'
    }
    if (!formData.incident_date) {
      newErrors.incident_date = 'Schadensdatum ist erforderlich'
    }
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleChange = (
    field: keyof ClaimFormData,
    value: string | number
  ) => {
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
      saveClaim()
    }
  }

  return (
    <div className="equipment-form-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Neuen Schadensfall melden</h1>
          <p className="page-subtitle">
            Erfassen Sie einen neuen Versicherungsfall mit Schadensbeschreibung
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="form-section">
        {errors.submit && (
          <div className="error-message" role="alert">
            {errors.submit}
          </div>
        )}

        <h2 className="form-section__title">Schadensdaten</h2>
        <div className="form-section__grid">
          <Select
            label="Versicherungspolice *"
            options={policyOptions}
            value={formData.policy_id}
            onChange={(e) => handleChange('policy_id', e.target.value)}
            placeholder="Police auswaehlen"
            error={errors.policy_id}
          />
          <Select
            label="Betroffenes Equipment"
            options={equipmentOptions}
            value={formData.equipment_id}
            onChange={(e) => handleChange('equipment_id', e.target.value)}
            placeholder="Equipment auswaehlen (optional)"
          />
          <Input
            label="Schadensdatum *"
            type="date"
            value={formData.incident_date}
            onChange={(e) => handleChange('incident_date', e.target.value)}
            error={errors.incident_date}
          />
          <Input
            label="Schadenshoehe (EUR)"
            type="number"
            value={formData.damage_amount || ''}
            onChange={(e) =>
              handleChange('damage_amount', parseFloat(e.target.value) || 0)
            }
            step="0.01"
            min="0"
            placeholder="0.00"
          />
        </div>

        <TextArea
          label="Schadensbeschreibung *"
          value={formData.description}
          onChange={(e) => handleChange('description', e.target.value)}
          placeholder="Beschreiben Sie den Schadenshergang moeglichst genau..."
          rows={5}
          error={errors.description}
        />

        <div className="form-section__footer">
          <button
            type="button"
            className="btn btn--secondary"
            onClick={() => navigate('/insurance')}
          >
            Abbrechen
          </button>
          <button
            type="submit"
            className="btn btn--primary"
            disabled={isPending}
          >
            {isPending ? 'Wird gemeldet...' : 'Schadensfall melden'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default NewClaimPage
