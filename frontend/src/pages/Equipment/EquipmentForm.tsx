import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { equipmentApi, categoryApi } from '../../services/api'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { TextArea } from '../../components/Form/TextArea'
import { FileUpload } from '../../components/Form/FileUpload'
import { CreateEquipmentDTO, Category } from '../../types/equipment'
import './Equipment.module.scss'

function EquipmentFormPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const isEditing = !!id

  const [formData, setFormData] = useState<CreateEquipmentDTO>({
    name: '',
    description: '',
    sku: '',
    barcode: '',
    category_id: '',
    location_id: '',
    rental_price_day: 0,
    rental_price_week: 0,
    weight: 0,
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list() as Promise<Category[]>,
    staleTime: 1000 * 60 * 10,
  })

  const categoryOptions = (categories || []).map((cat: Category) => ({
    value: cat.id,
    label: cat.name,
  }))

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
      const errorMessage = (error as any)?.response?.data?.error || (error as any)?.response?.data?.message || 'Fehler beim Speichern'
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
        category_id: equipment.category_id,
        location_id: equipment.location_id || '',
        rental_price_day: equipment.rental_price_day || 0,
        rental_price_week: equipment.rental_price_week || 0,
        weight: equipment.weight || 0,
        serial_number: equipment.serial_number || '',
        tags: equipment.tags || [],
        rfid_tag: equipment.rfid_tag || '',
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
    if (!formData.category_id) {
      newErrors.category_id = 'Kategorie ist erforderlich'
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
          <Input
            label="Seriennummer"
            value={formData.serial_number || ''}
            onChange={(e) => handleInputChange('serial_number', e.target.value)}
            placeholder="Optional"
          />
          <Input
            label="RFID Tag"
            value={formData.rfid_tag || ''}
            onChange={(e) => handleInputChange('rfid_tag', e.target.value)}
            placeholder="Optional (z.B. E28011606000020...)"
          />
          <Select
            label="Kategorie"
            options={categoryOptions}
            value={formData.category_id}
            onChange={(e) => handleInputChange('category_id', e.target.value)}
            error={errors.category_id}
            placeholder="Kategorie auswählen"
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
            value={formData.rental_price_day || 0}
            onChange={(e) => handleInputChange('rental_price_day', parseFloat(e.target.value))}
            step="0.01"
            min="0"
          />
          <Input
            label="Wochenpreis (€)"
            type="number"
            value={formData.rental_price_week || 0}
            onChange={(e) => handleInputChange('rental_price_week', parseFloat(e.target.value))}
            step="0.01"
            min="0"
          />
          <Input
            label="Einkaufspreis (€)"
            type="number"
            value={formData.purchase_price || 0}
            onChange={(e) => handleInputChange('purchase_price', parseFloat(e.target.value))}
            step="0.01"
            min="0"
          />
          <Input
            label="Gewicht (kg)"
            type="number"
            value={formData.weight || 0}
            onChange={(e) => handleInputChange('weight', parseFloat(e.target.value))}
            step="0.1"
            min="0"
          />
        </div>

        <h2 className="form-section__title">Standort & Bilder</h2>
        <div className="form-section__grid">
          <Input
            label="Standort-ID"
            value={formData.location_id || ''}
            onChange={(e) => handleInputChange('location_id', e.target.value)}
            placeholder="z.B. Lager-A"
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
