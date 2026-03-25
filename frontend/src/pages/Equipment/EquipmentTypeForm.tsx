import { useState, useEffect } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { equipmentTypeApi, categoryApi } from '../../services/api'
import { Category } from '../../types/equipment'
import { Modal } from '../../components/Modal/Modal'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { useNotificationStore } from '../../stores/notificationStore'
import './EquipmentTypes.scss'

interface EquipmentType {
  id: string
  name: string
  manufacturer?: string
  model?: string
  category_id?: string
  sku_prefix?: string
  rental_price_day?: number
  rental_price_week?: number
  replacement_value?: number
  weight?: number
  dimensions?: string
  image_url?: string
  tags?: string[]
}

interface Props {
  type: EquipmentType | null
  onClose: () => void
  onSuccess: () => void
}

function EquipmentTypeForm({ type, onClose, onSuccess }: Props) {
  const { addNotification } = useNotificationStore()
  const isEdit = !!type

  const [form, setForm] = useState({
    name: '',
    manufacturer: '',
    model: '',
    category_id: '',
    sku_prefix: '',
    rental_price_day: '',
    rental_price_week: '',
    replacement_value: '',
    weight: '',
    dimensions: '',
    image_url: '',
    tags: '',
  })
  const [itemCount, setItemCount] = useState(0)
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

  useEffect(() => {
    if (type) {
      setForm({
        name: type.name || '',
        manufacturer: type.manufacturer || '',
        model: type.model || '',
        category_id: type.category_id || '',
        sku_prefix: type.sku_prefix || '',
        rental_price_day: type.rental_price_day != null ? String(type.rental_price_day) : '',
        rental_price_week: type.rental_price_week != null ? String(type.rental_price_week) : '',
        replacement_value: type.replacement_value != null ? String(type.replacement_value) : '',
        weight: type.weight != null ? String(type.weight) : '',
        dimensions: type.dimensions || '',
        image_url: type.image_url || '',
        tags: (type.tags || []).join(', '),
      })
    }
  }, [type])

  const { mutate: saveType, isPending: isSaving } = useMutation({
    mutationFn: async () => {
      const payload: Record<string, unknown> = {
        name: form.name.trim(),
      }
      if (form.manufacturer.trim()) payload.manufacturer = form.manufacturer.trim()
      if (form.model.trim()) payload.model = form.model.trim()
      if (form.category_id) payload.category_id = form.category_id
      if (form.sku_prefix.trim()) payload.sku_prefix = form.sku_prefix.trim()
      if (form.rental_price_day) payload.rental_price_day = parseFloat(form.rental_price_day)
      if (form.rental_price_week) payload.rental_price_week = parseFloat(form.rental_price_week)
      if (form.replacement_value) payload.replacement_value = parseFloat(form.replacement_value)
      if (form.weight) payload.weight = parseFloat(form.weight)
      if (form.dimensions.trim()) payload.dimensions = form.dimensions.trim()
      if (form.image_url.trim()) payload.image_url = form.image_url.trim()
      if (form.tags.trim()) {
        payload.tags = form.tags.split(',').map(t => t.trim()).filter(Boolean)
      }

      let result
      if (isEdit) {
        result = await equipmentTypeApi.update(type.id, payload)
      } else {
        result = await equipmentTypeApi.create(payload)
      }

      // Create items if requested (only on create)
      if (!isEdit && itemCount > 0 && result?.id) {
        await equipmentTypeApi.createItems(result.id, itemCount)
      }

      return result
    },
    onSuccess: () => {
      const message = isEdit
        ? 'Equipment-Typ erfolgreich aktualisiert'
        : `Equipment-Typ erstellt${itemCount > 0 ? ` mit ${itemCount} Einzelartikel(n)` : ''}`
      addNotification(message, 'success', { title: 'Erfolg', duration: 3000 })
      onSuccess()
    },
    onError: (err: unknown) => {
      const message = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
        || 'Fehler beim Speichern des Equipment-Typs'
      addNotification(message, 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const validate = () => {
    const newErrors: Record<string, string> = {}
    if (!form.name.trim()) newErrors.name = 'Name ist erforderlich'
    if (form.rental_price_day && isNaN(parseFloat(form.rental_price_day))) {
      newErrors.rental_price_day = 'Ungueltige Zahl'
    }
    if (form.rental_price_week && isNaN(parseFloat(form.rental_price_week))) {
      newErrors.rental_price_week = 'Ungueltige Zahl'
    }
    if (form.replacement_value && isNaN(parseFloat(form.replacement_value))) {
      newErrors.replacement_value = 'Ungueltige Zahl'
    }
    if (form.weight && isNaN(parseFloat(form.weight))) {
      newErrors.weight = 'Ungueltige Zahl'
    }
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = () => {
    if (validate()) {
      saveType()
    }
  }

  const updateField = (field: string, value: string) => {
    setForm(prev => ({ ...prev, [field]: value }))
    if (errors[field]) {
      setErrors(prev => {
        const next = { ...prev }
        delete next[field]
        return next
      })
    }
  }

  return (
    <Modal
      isOpen={true}
      onClose={onClose}
      title={isEdit ? 'Equipment-Typ bearbeiten' : 'Neuen Equipment-Typ erstellen'}
      size="lg"
      footer={
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
          <button className="btn btn--secondary" onClick={onClose}>
            Abbrechen
          </button>
          <button
            className="btn btn--primary"
            onClick={handleSubmit}
            disabled={isSaving}
          >
            {isSaving ? 'Wird gespeichert...' : isEdit ? 'Aktualisieren' : 'Erstellen'}
          </button>
        </div>
      }
    >
      <div className="et-form">
        {/* Basic info */}
        <Input
          label="Name *"
          value={form.name}
          onChange={(e) => updateField('name', e.target.value)}
          placeholder="z.B. Shure SM58"
          error={errors.name}
        />

        <div className="et-form__row">
          <Input
            label="Hersteller"
            value={form.manufacturer}
            onChange={(e) => updateField('manufacturer', e.target.value)}
            placeholder="z.B. Shure"
          />
          <Input
            label="Modell"
            value={form.model}
            onChange={(e) => updateField('model', e.target.value)}
            placeholder="z.B. SM58"
          />
        </div>

        <div className="et-form__row">
          <Select
            label="Kategorie"
            options={categoryOptions}
            value={form.category_id}
            onChange={(e) => updateField('category_id', e.target.value)}
            placeholder="Kategorie waehlen"
          />
          <Input
            label="SKU Praefix"
            value={form.sku_prefix}
            onChange={(e) => updateField('sku_prefix', e.target.value)}
            placeholder="z.B. SM58"
            helperText="Wird fuer automatische SKU-Generierung genutzt"
          />
        </div>

        {/* Pricing */}
        <h4 className="et-form__section-title">Preise</h4>
        <div className="et-form__row">
          <Input
            label="Tagespreis (EUR)"
            type="number"
            step="0.01"
            min="0"
            value={form.rental_price_day}
            onChange={(e) => updateField('rental_price_day', e.target.value)}
            placeholder="0.00"
            error={errors.rental_price_day}
          />
          <Input
            label="Wochenpreis (EUR)"
            type="number"
            step="0.01"
            min="0"
            value={form.rental_price_week}
            onChange={(e) => updateField('rental_price_week', e.target.value)}
            placeholder="0.00"
            error={errors.rental_price_week}
          />
        </div>
        <Input
          label="Wiederbeschaffungswert (EUR)"
          type="number"
          step="0.01"
          min="0"
          value={form.replacement_value}
          onChange={(e) => updateField('replacement_value', e.target.value)}
          placeholder="0.00"
          error={errors.replacement_value}
        />

        {/* Details */}
        <h4 className="et-form__section-title">Details</h4>
        <div className="et-form__row">
          <Input
            label="Gewicht (kg)"
            type="number"
            step="0.01"
            min="0"
            value={form.weight}
            onChange={(e) => updateField('weight', e.target.value)}
            placeholder="0.00"
            error={errors.weight}
          />
          <Input
            label="Abmessungen"
            value={form.dimensions}
            onChange={(e) => updateField('dimensions', e.target.value)}
            placeholder="z.B. 30x20x15 cm"
          />
        </div>

        <Input
          label="Bild-URL"
          type="url"
          value={form.image_url}
          onChange={(e) => updateField('image_url', e.target.value)}
          placeholder="https://..."
        />

        <Input
          label="Tags"
          value={form.tags}
          onChange={(e) => updateField('tags', e.target.value)}
          placeholder="Kommagetrennt, z.B. Audio, Mikrofon, Live"
          helperText="Mehrere Tags mit Komma trennen"
        />

        {/* Item creation (only on create) */}
        {!isEdit && (
          <>
            <h4 className="et-form__section-title">Einzelartikel erstellen</h4>
            <Input
              label="Anzahl Einzelartikel"
              type="number"
              min="0"
              max="100"
              value={String(itemCount)}
              onChange={(e) => setItemCount(Math.max(0, parseInt(e.target.value) || 0))}
              helperText="Optional: Sofort N Einzelartikel aus diesem Typ generieren"
            />
          </>
        )}
      </div>
    </Modal>
  )
}

export default EquipmentTypeForm
