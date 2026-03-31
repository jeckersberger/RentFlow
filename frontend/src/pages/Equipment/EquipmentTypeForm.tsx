import { useState, useEffect } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { equipmentTypeApi, categoryApi, api } from '../../services/api'
import { Category } from '../../types/equipment'
import { Modal } from '../../components/Modal/Modal'
import { useNotificationStore } from '../../stores/notificationStore'
import { ChevronDown, ChevronUp, Sparkles, Loader2 } from 'lucide-react'
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

  // Minimal fields (always visible)
  const [name, setName] = useState('')
  const [itemCount, setItemCount] = useState(1)

  // Advanced fields (expandable)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [manufacturer, setManufacturer] = useState('')
  const [model, setModel] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [rentalPriceDay, setRentalPriceDay] = useState('')
  const [rentalPriceWeek, setRentalPriceWeek] = useState('')
  const [weight, setWeight] = useState('')
  const [imageUrl, setImageUrl] = useState('')

  // KI lookup state
  const [aiLoading, setAiLoading] = useState(false)

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list() as Promise<Category[]>,
    staleTime: 1000 * 60 * 10,
  })

  useEffect(() => {
    if (type) {
      setName(type.name || '')
      setManufacturer(type.manufacturer || '')
      setModel(type.model || '')
      setCategoryId(type.category_id || '')
      setRentalPriceDay(type.rental_price_day ? String(type.rental_price_day) : '')
      setRentalPriceWeek(type.rental_price_week ? String(type.rental_price_week) : '')
      setWeight(type.weight ? String(type.weight) : '')
      setImageUrl(type.image_url || '')
      setShowAdvanced(true)
      setItemCount(0)
    }
  }, [type])

  const { mutate: save, isPending } = useMutation({
    mutationFn: async () => {
      const payload: Record<string, unknown> = { name: name.trim() }
      if (manufacturer.trim()) payload.manufacturer = manufacturer.trim()
      if (model.trim()) payload.model = model.trim()
      if (categoryId) payload.category_id = categoryId
      if (rentalPriceDay) payload.rental_price_day = parseFloat(rentalPriceDay)
      if (rentalPriceWeek) payload.rental_price_week = parseFloat(rentalPriceWeek)
      if (weight) payload.weight = parseFloat(weight)
      if (imageUrl.trim()) payload.image_url = imageUrl.trim()

      let result
      if (isEdit) {
        result = await equipmentTypeApi.update(type.id, payload)
      } else {
        result = await equipmentTypeApi.create(payload)
      }

      if (!isEdit && itemCount > 0 && result?.id) {
        await equipmentTypeApi.createItems(result.id, itemCount)
      }

      // Trigger KI auto-fill in background (non-blocking)
      if (!isEdit && result?.id && !manufacturer && !weight) {
        triggerAIAutoFill(result.id, name.trim())
      }

      return result
    },
    onSuccess: () => {
      addNotification(
        isEdit
          ? 'Gespeichert'
          : `${name} erstellt${itemCount > 0 ? ` — ${itemCount} Artikel angelegt` : ''}`,
        'success',
        { duration: 3000 }
      )
      onSuccess()
    },
    onError: () => {
      addNotification('Fehler beim Speichern', 'error', { duration: 5000 })
    },
  })

  const triggerAIAutoFill = async (typeId: string, equipmentName: string) => {
    try {
      const res = await api.post('/api/v1/ai/asset-lookup', { name: equipmentName })
      const data = res.data
      if (data && (data.manufacturer || data.weight || data.image_url)) {
        const updates: Record<string, unknown> = {}
        if (data.manufacturer) updates.manufacturer = data.manufacturer
        if (data.model) updates.model = data.model
        if (data.weight) updates.weight = data.weight
        if (data.image_url) updates.image_url = data.image_url
        if (data.rental_price_day) updates.rental_price_day = data.rental_price_day
        if (data.category) updates.category_id = data.category
        await equipmentTypeApi.update(typeId, updates)
        addNotification(`KI hat Details zu "${equipmentName}" ergänzt`, 'success', { duration: 4000 })
      }
    } catch {
      // KI nicht verfügbar — kein Problem, User kann manuell ausfüllen
    }
  }

  const handleAILookup = async () => {
    if (!name.trim()) return
    setAiLoading(true)
    try {
      const res = await api.post('/api/v1/ai/asset-lookup', { name: name.trim() })
      const data = res.data
      if (data) {
        if (data.manufacturer) setManufacturer(data.manufacturer)
        if (data.model) setModel(data.model)
        if (data.weight) setWeight(String(data.weight))
        if (data.image_url) setImageUrl(data.image_url)
        if (data.rental_price_day) setRentalPriceDay(String(data.rental_price_day))
        if (data.rental_price_week) setRentalPriceWeek(String(data.rental_price_week))
        setShowAdvanced(true)
        addNotification('KI-Daten geladen', 'success', { duration: 2000 })
      }
    } catch {
      addNotification('KI-Suche nicht verfügbar — bitte API-Key eintragen', 'warning', { duration: 4000 })
    } finally {
      setAiLoading(false)
    }
  }

  return (
    <Modal isOpen onClose={onClose} title={isEdit ? 'Equipment bearbeiten' : 'Neues Equipment'} size="md">
      <div className="et-quick-form">
        {/* HAUPTFELDER — immer sichtbar */}
        <div className="et-quick-form__main">
          <div className="et-quick-form__name-row">
            <div style={{ flex: 1 }}>
              <label className="et-quick-form__label">Was möchten Sie anlegen?</label>
              <input
                className="et-quick-form__input"
                type="text"
                placeholder='z.B. "QSC K12.2" oder "XLR Kabel 10m"'
                value={name}
                onChange={e => setName(e.target.value)}
                autoFocus
              />
            </div>
            <button
              className="et-quick-form__ai-btn"
              onClick={handleAILookup}
              disabled={!name.trim() || aiLoading}
              title="KI sucht Details automatisch"
            >
              {aiLoading ? <Loader2 size={18} className="spin" /> : <Sparkles size={18} />}
            </button>
          </div>

          {!isEdit && (
            <div>
              <label className="et-quick-form__label">Anzahl</label>
              <input
                className="et-quick-form__input et-quick-form__input--small"
                type="number"
                min={0}
                max={100}
                value={itemCount}
                onChange={e => setItemCount(parseInt(e.target.value) || 0)}
              />
              <span className="et-quick-form__hint">
                {itemCount > 0
                  ? `${itemCount} Einzelartikel werden mit automatischer Seriennummer erstellt`
                  : 'Nur den Typ anlegen, Artikel später erstellen'}
              </span>
            </div>
          )}
        </div>

        {/* ERWEITERTE FELDER — aufklappbar */}
        <button
          className="et-quick-form__expand"
          onClick={() => setShowAdvanced(!showAdvanced)}
          type="button"
        >
          {showAdvanced ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
          {showAdvanced ? 'Weniger Details' : 'Mehr Details (optional)'}
        </button>

        {showAdvanced && (
          <div className="et-quick-form__advanced">
            <div className="et-quick-form__grid">
              <div>
                <label className="et-quick-form__label">Hersteller</label>
                <input className="et-quick-form__input" value={manufacturer} onChange={e => setManufacturer(e.target.value)} placeholder="z.B. QSC" />
              </div>
              <div>
                <label className="et-quick-form__label">Modell</label>
                <input className="et-quick-form__input" value={model} onChange={e => setModel(e.target.value)} placeholder="z.B. K12.2" />
              </div>
              <div>
                <label className="et-quick-form__label">Kategorie</label>
                <select className="et-quick-form__input" value={categoryId} onChange={e => setCategoryId(e.target.value)}>
                  <option value="">Keine Kategorie</option>
                  {(categories || []).map((cat: Category) => (
                    <option key={cat.id} value={cat.id}>{cat.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="et-quick-form__label">Gewicht (kg)</label>
                <input className="et-quick-form__input" type="number" step="0.1" value={weight} onChange={e => setWeight(e.target.value)} placeholder="0.0" />
              </div>
              <div>
                <label className="et-quick-form__label">Tagespreis (€)</label>
                <input className="et-quick-form__input" type="number" step="0.01" value={rentalPriceDay} onChange={e => setRentalPriceDay(e.target.value)} placeholder="0.00" />
              </div>
              <div>
                <label className="et-quick-form__label">Wochenpreis (€)</label>
                <input className="et-quick-form__input" type="number" step="0.01" value={rentalPriceWeek} onChange={e => setRentalPriceWeek(e.target.value)} placeholder="0.00" />
              </div>
            </div>
            <div>
              <label className="et-quick-form__label">Bild-URL</label>
              <input className="et-quick-form__input" value={imageUrl} onChange={e => setImageUrl(e.target.value)} placeholder="https://..." />
            </div>
          </div>
        )}

        {/* AKTIONEN */}
        <div className="et-quick-form__actions">
          <button className="btn btn--secondary" onClick={onClose}>Abbrechen</button>
          <button
            className="btn btn--primary"
            disabled={!name.trim() || isPending}
            onClick={() => save()}
          >
            {isPending ? 'Wird erstellt...' : isEdit ? 'Speichern' : `Erstellen${itemCount > 0 ? ` (${itemCount} Artikel)` : ''}`}
          </button>
        </div>
      </div>
    </Modal>
  )
}

export default EquipmentTypeForm
