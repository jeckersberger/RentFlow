import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { equipmentApi, categoryApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { TextArea } from '../../components/Form/TextArea'
import { Modal } from '../../components/Modal/Modal'
import { CreateEquipmentDTO, Category } from '../../types/equipment'
import './Equipment.scss'

const STATUS_OPTIONS: Array<{ value: string; label: string }> = [
  { value: 'available', label: 'Verfügbar' },
  { value: 'reserved', label: 'Reserviert' },
  { value: 'checked_out', label: 'Vermietet' },
  { value: 'in_maintenance', label: 'In Wartung' },
  { value: 'damaged', label: 'Beschädigt' },
  { value: 'retired', label: 'Ausgemustert' },
]

const CONDITION_OPTIONS: Array<{ value: string; label: string }> = [
  { value: 'new', label: 'Neu' },
  { value: 'excellent', label: 'Sehr gut' },
  { value: 'good', label: 'Gut' },
  { value: 'fair', label: 'Befriedigend' },
  { value: 'poor', label: 'Mangelhaft' },
  { value: 'defective', label: 'Defekt' },
]

interface FormDataExtended extends CreateEquipmentDTO {
  status?: string
  condition?: string
}

function EquipmentFormPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const { addNotification } = useNotificationStore()
  const isEditing = !!id

  const [formData, setFormData] = useState<FormDataExtended>({
    name: '',
    description: '',
    sku: '',
    barcode: '',
    category_id: '',
    location_id: '',
    rental_price_day: 0,
    rental_price_week: 0,
    purchase_price: 0,
    weight: 0,
    dimensions: { length: 0, width: 0, height: 0, unit: 'cm' },
    tags: [],
    status: 'available',
    condition: 'good',
  })

  const [tagInput, setTagInput] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [imagePreview, setImagePreview] = useState<string | null>(null)
  const [isDragging, setIsDragging] = useState(false)
  const imageInputRef = useRef<HTMLInputElement>(null)
  const [showAiModal, setShowAiModal] = useState(false)
  const [aiImage, setAiImage] = useState<File | null>(null)
  const [aiImagePreview, setAiImagePreview] = useState<string | null>(null)
  const [isAiAnalyzing, setIsAiAnalyzing] = useState(false)
  const aiFileInputRef = useRef<HTMLInputElement>(null)

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
      // Build payload - include status/condition only for create
      const payload: Record<string, unknown> = { ...formData }
      let result
      if (isEditing && id) {
        result = await equipmentApi.update(id, payload)
      } else {
        result = await equipmentApi.create(payload)
      }
      // Upload image if one was selected
      if (imageFile && result?.id) {
        try {
          await equipmentApi.uploadImage(result.id, imageFile)
        } catch (imgErr) {
          console.warn('Image upload failed, equipment was saved:', imgErr)
          addNotification(
            'Ausrüstung gespeichert, aber Bild-Upload fehlgeschlagen',
            'warning',
            { title: 'Hinweis', duration: 5000 }
          )
        }
      }
      return result
    },
    onSuccess: (data) => {
      addNotification(
        isEditing ? 'Ausrüstung erfolgreich aktualisiert' : 'Ausrüstung erfolgreich erstellt',
        'success',
        { title: 'Erfolg', duration: 3000 }
      )
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
        purchase_price: equipment.purchase_price || 0,
        weight: equipment.weight || 0,
        dimensions: equipment.dimensions || { length: 0, width: 0, height: 0, unit: 'cm' },
        serial_number: equipment.serial_number || '',
        tags: equipment.tags || [],
        rfid_tag: equipment.rfid_tag || '',
        status: equipment.status || 'available',
        condition: equipment.condition || 'good',
      })
    }
    // Show existing image if available
    if (equipment.image_url) {
      setImagePreview(equipment.image_url)
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
    field: keyof FormDataExtended,
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

  const handleDimensionChange = (dim: 'length' | 'width' | 'height', value: number) => {
    setFormData((prev) => ({
      ...prev,
      dimensions: {
        ...(prev.dimensions || {}),
        [dim]: value,
        unit: prev.dimensions?.unit || 'cm',
      },
    }))
  }

  const handleAddTag = () => {
    const tag = tagInput.trim().toLowerCase()
    if (tag && !(formData.tags || []).includes(tag)) {
      handleInputChange('tags', [...(formData.tags || []), tag])
    }
    setTagInput('')
  }

  const handleRemoveTag = (tagToRemove: string) => {
    handleInputChange('tags', (formData.tags || []).filter((t) => t !== tagToRemove))
  }

  const handleTagKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddTag()
    }
  }

  const handleImageSelect = useCallback((file: File) => {
    if (!file.type.startsWith('image/')) return
    if (file.size > 10 * 1024 * 1024) {
      addNotification('Bild darf maximal 10 MB groß sein', 'error', { title: 'Fehler', duration: 4000 })
      return
    }
    setImageFile(file)
    const reader = new FileReader()
    reader.onload = (ev) => {
      setImagePreview(ev.target?.result as string)
    }
    reader.readAsDataURL(file)
  }, [addNotification])

  const handleImageInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) handleImageSelect(file)
  }

  const handleImageDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    const file = e.dataTransfer.files?.[0]
    if (file) handleImageSelect(file)
  }, [handleImageSelect])

  const handleImageDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleImageDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
  }, [])

  const handleRemoveImage = () => {
    setImageFile(null)
    setImagePreview(null)
    if (imageInputRef.current) imageInputRef.current.value = ''
  }

  const handleAiImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setAiImage(file)
      const reader = new FileReader()
      reader.onload = (ev) => {
        setAiImagePreview(ev.target?.result as string)
      }
      reader.readAsDataURL(file)
    }
  }

  const handleAiAnalyze = () => {
    if (!aiImage) return
    setIsAiAnalyzing(true)
    // Placeholder: simulate AI analysis
    setTimeout(() => {
      setIsAiAnalyzing(false)
      setShowAiModal(false)
      setAiImage(null)
      setAiImagePreview(null)
      addNotification(
        'KI-Erkennung wird in einer zukünftigen Version verfügbar',
        'info',
        { title: 'KI-Erkennung', duration: 5000 }
      )
    }, 2000)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (validateForm()) {
      saveEquipment()
    }
  }

  if (isLoadingEquipment) {
    return (
      <div className="equipment-form-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Wird geladen...</h1>
          </div>
        </div>
        <div className="form-section">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} style={{ height: '2.5rem', background: 'var(--color-bg-tertiary)', borderRadius: 'var(--radius-md)', marginBottom: 'var(--spacing-4)' }} />
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="equipment-form-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">
            {isEditing ? 'Ausrüstung bearbeiten' : 'Neue Ausrüstung'}
          </h1>
          <p className="page-subtitle">
            {isEditing ? `${equipment?.name} bearbeiten` : 'Fügen Sie ein neues Gerät hinzu'}
          </p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button
            type="button"
            className="btn btn--secondary"
            onClick={() => setShowAiModal(true)}
            style={{
              background: 'linear-gradient(135deg, rgba(139, 92, 246, 0.15), rgba(0, 212, 255, 0.15))',
              border: '1px solid rgba(139, 92, 246, 0.3)',
              color: 'var(--color-accent-light)',
            }}
          >
            KI erkennen
          </button>
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
            label="Name *"
            value={formData.name}
            onChange={(e) => handleInputChange('name', e.target.value)}
            placeholder="z.B. LED-Bühnenbeleuchtung"
            error={errors.name}
          />
          <Input
            label="SKU *"
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
            label="Kategorie *"
            options={categoryOptions}
            value={formData.category_id}
            onChange={(e) => handleInputChange('category_id', e.target.value)}
            error={errors.category_id}
            placeholder="Kategorie auswählen"
          />
          <Select
            label="Status"
            options={STATUS_OPTIONS}
            value={formData.status || 'available'}
            onChange={(e) => handleInputChange('status', e.target.value)}
          />
          <Select
            label="Zustand"
            options={CONDITION_OPTIONS}
            value={formData.condition || 'good'}
            onChange={(e) => handleInputChange('condition', e.target.value)}
          />
        </div>

        <TextArea
          label="Beschreibung"
          value={formData.description || ''}
          onChange={(e) => handleInputChange('description', e.target.value)}
          placeholder="Detaillierte Beschreibung der Ausrüstung..."
          rows={4}
        />

        <h2 className="form-section__title">Preisgestaltung & Gewicht</h2>
        <div className="form-section__grid">
          <Input
            label="Tagespreis (€)"
            type="number"
            value={formData.rental_price_day || 0}
            onChange={(e) => handleInputChange('rental_price_day', parseFloat(e.target.value) || 0)}
            step="0.01"
            min="0"
          />
          <Input
            label="Wochenpreis (€)"
            type="number"
            value={formData.rental_price_week || 0}
            onChange={(e) => handleInputChange('rental_price_week', parseFloat(e.target.value) || 0)}
            step="0.01"
            min="0"
          />
          <Input
            label="Einkaufspreis (€)"
            type="number"
            value={formData.purchase_price || 0}
            onChange={(e) => handleInputChange('purchase_price', parseFloat(e.target.value) || 0)}
            step="0.01"
            min="0"
          />
          <Input
            label="Gewicht (kg)"
            type="number"
            value={formData.weight || 0}
            onChange={(e) => handleInputChange('weight', parseFloat(e.target.value) || 0)}
            step="0.1"
            min="0"
          />
        </div>

        <h2 className="form-section__title">Abmessungen</h2>
        <div className="form-section__grid">
          <Input
            label="Länge (cm)"
            type="number"
            value={formData.dimensions?.length || 0}
            onChange={(e) => handleDimensionChange('length', parseFloat(e.target.value) || 0)}
            step="0.1"
            min="0"
          />
          <Input
            label="Breite (cm)"
            type="number"
            value={formData.dimensions?.width || 0}
            onChange={(e) => handleDimensionChange('width', parseFloat(e.target.value) || 0)}
            step="0.1"
            min="0"
          />
          <Input
            label="Höhe (cm)"
            type="number"
            value={formData.dimensions?.height || 0}
            onChange={(e) => handleDimensionChange('height', parseFloat(e.target.value) || 0)}
            step="0.1"
            min="0"
          />
        </div>

        <h2 className="form-section__title">Standort, Tags & Bilder</h2>
        <div className="form-section__grid">
          <Input
            label="Standort-ID"
            value={formData.location_id || ''}
            onChange={(e) => handleInputChange('location_id', e.target.value)}
            placeholder="z.B. Lager-A"
          />
          <div className="form-group">
            <label className="form-label">Tags</label>
            <div style={{ display: 'flex', gap: 'var(--spacing-2)', alignItems: 'center' }}>
              <input
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={handleTagKeyDown}
                placeholder="Tag eingeben + Enter"
                className="form-input"
                style={{
                  flex: 1,
                  height: 'var(--input-height)',
                  padding: '0 var(--input-padding-x)',
                  backgroundColor: 'var(--glass-bg-input-strong)',
                  border: '1px solid var(--color-border-strong)',
                  borderRadius: 'var(--radius-input)',
                  color: 'var(--color-text-primary)',
                  fontSize: 'var(--font-size-sm)',
                }}
              />
              <button
                type="button"
                className="btn btn--secondary btn--sm"
                onClick={handleAddTag}
              >
                +
              </button>
            </div>
            {(formData.tags || []).length > 0 && (
              <div style={{ display: 'flex', gap: 'var(--spacing-2)', flexWrap: 'wrap', marginTop: 'var(--spacing-2)' }}>
                {(formData.tags || []).map((tag) => (
                  <span
                    key={tag}
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 'var(--spacing-1)',
                      padding: '0.2rem 0.6rem',
                      backgroundColor: 'rgba(0, 212, 255, 0.1)',
                      border: '1px solid rgba(0, 212, 255, 0.2)',
                      borderRadius: 'var(--radius-full)',
                      fontSize: 'var(--font-size-xs)',
                      color: 'var(--color-primary)',
                    }}
                  >
                    {tag}
                    <button
                      type="button"
                      onClick={() => handleRemoveTag(tag)}
                      style={{
                        background: 'none',
                        border: 'none',
                        color: 'var(--color-primary)',
                        cursor: 'pointer',
                        padding: '0 2px',
                        fontSize: 'var(--font-size-xs)',
                        lineHeight: 1,
                      }}
                    >
                      x
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="form-group">
          <label className="form-label">Bild</label>
          {imagePreview ? (
            <div style={{ position: 'relative', width: '100%', maxWidth: '400px' }}>
              <img
                src={imagePreview}
                alt="Vorschau"
                style={{
                  width: '100%',
                  maxHeight: '300px',
                  objectFit: 'cover',
                  borderRadius: 'var(--radius-lg)',
                  border: '1px solid var(--color-border)',
                }}
              />
              <button
                type="button"
                onClick={handleRemoveImage}
                style={{
                  position: 'absolute',
                  top: '8px',
                  right: '8px',
                  background: 'rgba(0, 0, 0, 0.6)',
                  border: 'none',
                  color: 'white',
                  borderRadius: 'var(--radius-full)',
                  width: '28px',
                  height: '28px',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 'var(--font-size-sm)',
                }}
              >
                x
              </button>
              {imageFile && (
                <p style={{ margin: 'var(--spacing-2) 0 0 0', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                  {imageFile.name} ({(imageFile.size / 1024).toFixed(0)} KB)
                </p>
              )}
            </div>
          ) : (
            <div
              onClick={() => imageInputRef.current?.click()}
              onDrop={handleImageDrop}
              onDragOver={handleImageDragOver}
              onDragLeave={handleImageDragLeave}
              style={{
                width: '100%',
                maxWidth: '400px',
                height: '160px',
                border: `2px dashed ${isDragging ? 'var(--color-primary)' : 'rgba(0, 212, 255, 0.3)'}`,
                borderRadius: 'var(--radius-lg)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                backgroundColor: isDragging ? 'rgba(0, 212, 255, 0.08)' : 'rgba(0, 212, 255, 0.03)',
                transition: 'all 0.2s ease',
              }}
            >
              <span style={{ fontSize: '2rem', marginBottom: 'var(--spacing-2)', opacity: 0.5 }}>+</span>
              <span style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>
                Bild hierher ziehen oder klicken
              </span>
              <span style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-xs)', marginTop: 'var(--spacing-1)' }}>
                Max. 10 MB (JPG, PNG, WebP)
              </span>
            </div>
          )}
          <input
            ref={imageInputRef}
            type="file"
            accept="image/*"
            style={{ display: 'none' }}
            onChange={handleImageInputChange}
          />
        </div>

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

      {/* AI Recognition Modal */}
      <Modal
        isOpen={showAiModal}
        onClose={() => {
          setShowAiModal(false)
          setAiImage(null)
          setAiImagePreview(null)
          setIsAiAnalyzing(false)
        }}
        title="KI-Erkennung"
        size="md"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => {
                setShowAiModal(false)
                setAiImage(null)
                setAiImagePreview(null)
                setIsAiAnalyzing(false)
              }}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={handleAiAnalyze}
              disabled={!aiImage || isAiAnalyzing}
            >
              {isAiAnalyzing ? 'KI analysiert Bild...' : 'Analysieren'}
            </button>
          </div>
        }
      >
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 'var(--spacing-4)' }}>
          <p style={{ margin: 0, color: 'var(--color-text-secondary)', textAlign: 'center', fontSize: 'var(--font-size-sm)' }}>
            Laden Sie ein Foto der Ausrüstung hoch. Die KI wird versuchen, das Gerät zu erkennen und die Formularfelder automatisch auszufüllen.
          </p>

          {aiImagePreview ? (
            <div style={{ position: 'relative', width: '100%', maxWidth: '400px' }}>
              <img
                src={aiImagePreview}
                alt="Hochgeladenes Bild"
                style={{
                  width: '100%',
                  borderRadius: 'var(--radius-lg)',
                  border: '1px solid var(--color-border)',
                }}
              />
              <button
                type="button"
                onClick={() => {
                  setAiImage(null)
                  setAiImagePreview(null)
                  if (aiFileInputRef.current) aiFileInputRef.current.value = ''
                }}
                style={{
                  position: 'absolute',
                  top: '8px',
                  right: '8px',
                  background: 'rgba(0, 0, 0, 0.6)',
                  border: 'none',
                  color: 'white',
                  borderRadius: 'var(--radius-full)',
                  width: '28px',
                  height: '28px',
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 'var(--font-size-sm)',
                }}
              >
                x
              </button>
            </div>
          ) : (
            <div
              onClick={() => aiFileInputRef.current?.click()}
              style={{
                width: '100%',
                maxWidth: '400px',
                height: '200px',
                border: '2px dashed rgba(0, 212, 255, 0.3)',
                borderRadius: 'var(--radius-lg)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                backgroundColor: 'rgba(0, 212, 255, 0.03)',
                transition: 'all 0.2s ease',
              }}
            >
              <span style={{ fontSize: '2rem', marginBottom: 'var(--spacing-2)', opacity: 0.5 }}>
                +
              </span>
              <span style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>
                Bild hochladen
              </span>
            </div>
          )}

          <input
            ref={aiFileInputRef}
            type="file"
            accept="image/*"
            style={{ display: 'none' }}
            onChange={handleAiImageSelect}
          />

          {isAiAnalyzing && (
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: 'var(--spacing-3)',
              padding: 'var(--spacing-3)',
              backgroundColor: 'rgba(139, 92, 246, 0.1)',
              border: '1px solid rgba(139, 92, 246, 0.2)',
              borderRadius: 'var(--radius-md)',
              width: '100%',
            }}>
              <div style={{
                width: '20px',
                height: '20px',
                border: '2px solid rgba(139, 92, 246, 0.3)',
                borderTopColor: 'var(--color-accent)',
                borderRadius: '50%',
                animation: 'spin 1s linear infinite',
              }} />
              <span style={{ color: 'var(--color-accent-light)', fontSize: 'var(--font-size-sm)' }}>
                KI analysiert Bild...
              </span>
            </div>
          )}
        </div>
      </Modal>

      {/* Inline CSS for spinner animation */}
      <style>{`
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  )
}

export default EquipmentFormPage
