import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { documentApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { Input } from '../../components/Form/Input'
import { TextArea } from '../../components/Form/TextArea'
import '../Equipment/Equipment.scss'

interface DocumentFormData {
  title: string
  document_type: string
  document_number: string
  reference_id: string
  template_id: string
  content: string
}

const DOCTYPE_OPTIONS = [
  { value: 'delivery_note', label: 'Lieferschein' },
  { value: 'return_note', label: 'Rueckgabeschein' },
  { value: 'invoice', label: 'Rechnung' },
  { value: 'quote', label: 'Angebot' },
  { value: 'contract', label: 'Vertrag' },
  { value: 'other', label: 'Sonstiges' },
]

function NewDocumentPage() {
  const navigate = useNavigate()
  const { addNotification } = useNotificationStore()

  const [formData, setFormData] = useState<DocumentFormData>({
    title: '',
    document_type: 'other',
    document_number: '',
    reference_id: '',
    template_id: '',
    content: '',
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  const { mutate: saveDocument, isPending } = useMutation({
    mutationFn: () =>
      documentApi.create({
        title: formData.title,
        document_type: formData.document_type,
        document_number: formData.document_number,
        reference_id: formData.reference_id,
        template_id: formData.template_id || undefined,
        metadata: formData.content ? { content: formData.content } : undefined,
      }),
    onSuccess: () => {
      addNotification('Dokument erfolgreich erstellt', 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
      navigate('/documents')
    },
    onError: (error: unknown) => {
      const msg =
        (error as any)?.response?.data?.error ||
        (error as any)?.response?.data?.message ||
        'Fehler beim Erstellen des Dokuments'
      setErrors({ submit: msg })
    },
  })

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}
    if (!formData.title.trim()) {
      newErrors.title = 'Name ist erforderlich'
    }
    if (!formData.document_number.trim()) {
      newErrors.document_number = 'Dokumentnummer ist erforderlich'
    }
    if (!formData.reference_id.trim()) {
      newErrors.reference_id = 'Referenz-ID ist erforderlich'
    }
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleChange = (field: keyof DocumentFormData, value: string) => {
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
      saveDocument()
    }
  }

  return (
    <div className="equipment-form-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Neues Dokument</h1>
          <p className="page-subtitle">
            Erstellen Sie ein neues Dokument oder eine Vorlage
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="form-section">
        {errors.submit && (
          <div className="error-message" role="alert">
            {errors.submit}
          </div>
        )}

        <h2 className="form-section__title">Dokumentdaten</h2>
        <div className="form-section__grid">
          <Input
            label="Name *"
            value={formData.title}
            onChange={(e) => handleChange('title', e.target.value)}
            placeholder="z.B. Lieferschein Projekt X"
            error={errors.title}
          />
          <Input
            label="Dokumentnummer *"
            value={formData.document_number}
            onChange={(e) => handleChange('document_number', e.target.value)}
            placeholder="z.B. DOC-2026-001"
            error={errors.document_number}
          />
          <div className="form-group">
            <label className="form-label">Dokumenttyp</label>
            <select
              className="form-select"
              value={formData.document_type}
              onChange={(e) => handleChange('document_type', e.target.value)}
            >
              {DOCTYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
          <Input
            label="Referenz-ID *"
            value={formData.reference_id}
            onChange={(e) => handleChange('reference_id', e.target.value)}
            placeholder="z.B. Projekt- oder Rechnungs-ID"
            error={errors.reference_id}
          />
          <Input
            label="Vorlage-ID"
            value={formData.template_id}
            onChange={(e) => handleChange('template_id', e.target.value)}
            placeholder="Optional"
          />
        </div>

        <TextArea
          label="Inhalt"
          value={formData.content}
          onChange={(e) => handleChange('content', e.target.value)}
          placeholder="Dokumentinhalt oder Beschreibung..."
          rows={6}
        />

        <div className="form-section__footer">
          <button
            type="button"
            className="btn btn--secondary"
            onClick={() => navigate('/documents')}
          >
            Abbrechen
          </button>
          <button
            type="submit"
            className="btn btn--primary"
            disabled={isPending}
          >
            {isPending ? 'Wird erstellt...' : 'Dokument erstellen'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default NewDocumentPage
