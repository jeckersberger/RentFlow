import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { equipmentApi } from '../../services/api'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { TextArea } from '../../components/Form/TextArea'
import { FileUpload } from '../../components/Form/FileUpload'
import { CreateEquipmentDTO } from '../../types/equipment'
import './Equipment.module.scss'

const CATEGORIES = [
  { value: 'lighting', label: 'Beleuchtung' },
  { value: 'sound', label: 'Ton' },
  { value: 'staging', label: 'Bühne' },
  { value: 'projection', label: 'Projektion' },
  { value: 'decoration', label: 'Dekoration' },
]

function EquipmentFormPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const isEditing = !!id

  const [formData, setFormData] = useState<CreateEquipmentDTO>({
    name: '',
    description: '',
    sku: '',
    barcode: '',
    category: '',
    location: '',
    price_daily: 0,
    price_weekly: 0,
    price_monthly: 0,
    quantity: 1,
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  const { data: equipment, isLoading: isLoadingEquipment } = useQuery(
    {
      queryKey: ['equipment', id],
      queryFn: () => equipmentApi.getById(id!),
      enabled: isEditing,
    }
  )

  const { mutate: saveEquipment, isPending } = useMutation({
    mutationFn: async () => {
      if (isEditing && id) {
        return equipmentApi.update(id, formData)
      } else {
        return equipmentApi.create(formData)
      }
    },
    onSuccess: (data) => {
      navigate(`/equipment/${data.id}`)
    },
    onError: (error: unknown) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const errorMessage = (error as any)?.response?.data?.message || 'Fehler beim Speichern'
      setErrors({ submit: errorMessage })
    },
  })

  useEffect(() => {
    if (equipment && isEditing) {
      setFormData({
        name: equipment.name,
        description: equipment.description || '',
        sku: equipment.sku,
        barcode: equipment.barcode || '',
        category: equipment.category,
        location: equipment.location || '',
        price_daily: equipment.price_daily || 0,
        price_weekly: equipment.price_weekly || 0,
        price_monthly: equipment.price_monthly || 0,
        quantity: equipment.quantity || 1,
      })
    }
  }, [equipment, isEditing])

  const validateForm = () => {
    const newErrors: typeof errors = {}

    if (!formData.name?.trim()) {
      newErrors.name = 'Name ist erforderlich'
    }
    if (!formData.sku?.trim()) {
      newErrors.sku = 'SKU ist erforderlich'
    }
    if (!formData.category) {
      newErrors.category = 'Kategorie ist erforderlich'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleInputChange = (
    field: keyof CreateEquipmentDTO,
    value: unknown
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }))
    if (errors[field]) {
      setErrors((prev) => ({
        ...prev,
        [field]: undefined,
      }))
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (validateForm()) {
      saveEquipment()
    }
  }

  if (isLoadingEquipment) {
    return <div className="equipment-form-page">Wird geladen...</div>
  }

  return (
    <div className="equipment-form-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">
            {isEditing ? 'Ausrüstung bearbeiten' : 'Neue Ausrüstung'}
          </h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="form-section">
        {errors.submit && (
          <div className="error-message" role="alert">
            {errors.submit}
          </div>
        )}

        <h2 className="form-section__title">Grundinformationen</h2>
        <div className="form-section__grid">
          <Input
            label="Name"
            value={formData.name}
            onChange={(e) => handleInputChange('name', e.target.value)}
            placeholder="z.B. LED-Bühnenbeleuchtung"
            error={errors.name}
          />
          <Input
            label="SKU"
            value={formData.sku}
            onChange={(e) => handleInputChange('sku', e.target.value)}
            placeholder="z.B. LED-001"
            error={errors.sku}
          />
          <Input
            label="Barcode"
            value={formData.barcode || ''}
            onChange={(e) => handleInputChange('barcode', e.target.value)}
            placeholder="Optional"
          />
          <Select
            label="Kategorie"
            options={CATEGORIES}
            value={formData.category}
            onChange={(e) => handleInputChange('category', e.target.value)}
            error={errors.category}
          />
        </div>

        <TextArea
          label="Beschreibung"
          value={formData.description || ''}
          onChange={(e) => handleInputChange('description', e.target.value)}
          placeholder="Detaillierte Beschreibung der Ausrüstung..."
          rows={4}
        />

        <h2 className="form-section__title">Preisgestaltung</h2>
        <div className="form-section__grid">
          <Input
            label="Tagespreis (€)"
            type="number"
            value={formData.price_daily || 0}
            onChange={(e) => handleInputChange('price_daily', parseFloat(e.target.value))}
            step="0.01"
            min="0"
          />
          <Input
            label="Wochenpreis (€)"
            type="number"
            value={formData.price_weekly || 0}
            onChange={(e) => handleInputChange('price_weekly', parseFloat(e.target.value))}
            step="0.01"
            min="0"
          />
          <Input
            label="Monatspreis (€)"
            type="number"
            value={formData.price_monthly || 0}
            onChange={(e) => handleInputChange('price_monthly', parseFloat(e.target.value))}
            step="0.01"
            min="0"
          />
          <Input
            label="Menge"
            type="number"
            value={formData.quantity || 1}
            onChange={(e) => handleInputChange('quantity', parseInt(e.target.value))}
            min="1"
          />
        </div>

        <h2 className="form-section__title">Standort & Bilder</h2>
        <div className="form-section__grid">
          <Input
            label="Standort"
            value={formData.location || ''}
            onChange={(e) => handleInputChange('location', e.target.value)}
            placeholder="z.B. Lagerraum A, Regal 3"
          />
        </div>

        <FileUpload
          label="Bilder"
          accept="image/*"
          multiple={true}
          onChange={() => {}}
        />

        <div className="form-section__footer">
          <button
            type="button"
            className="btn btn--secondary"
            onClick={() => navigate(-1)}
          >
            Abbrechen
          </button>
          <button
            type="submit"
            className="btn btn--primary"
            disabled={isPending}
          >
            {isPending
              ? 'Wird gespeichert...'
              : isEditing
              ? 'Änderungen speichern'
              : 'Ausrüstung erstellen'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default EquipmentFormPage
