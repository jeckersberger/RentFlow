import { useState, useEffect, useCallback } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { projectApi, contactApi } from '../../services/api'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { TextArea } from '../../components/Form/TextArea'
import { CreateProjectDTO, ProjectStatus } from '../../types/project'
import '../Equipment/Equipment.scss'

const STATUS_OPTIONS: Array<{ value: ProjectStatus; label: string }> = [
  { value: 'draft', label: 'Entwurf' },
  { value: 'quoted', label: 'Angebot' },
  { value: 'confirmed', label: 'Bestätigt' },
  { value: 'in_progress', label: 'In Bearbeitung' },
  { value: 'completed', label: 'Abgeschlossen' },
  { value: 'cancelled', label: 'Storniert' },
]

const PROJECT_TYPE_OPTIONS = [
  { value: 'dryhire', label: 'Dryhire' },
  { value: 'band', label: 'Band' },
  { value: 'production', label: 'Produktion' },
  { value: 'sale', label: 'Verkauf' },
  { value: 'installation', label: 'Festinstallation' },
]

interface Contact {
  id: string
  name?: string
  company_name?: string
  email?: string
}

function ProjectFormPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const isEditing = !!id

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [formData, setFormData] = useState<CreateProjectDTO & { project_type?: string; color?: string; venue_name?: string; contact_id?: string }>({
    name: '',
    description: '',
    client_name: '',
    client_email: '',
    client_phone: '',
    venue_address: {
      street: '',
      city: '',
      state: '',
      postal_code: '',
      country: 'AT',
    },
    status: 'draft',
    start_date: '',
    end_date: '',
    budget: 0,
    currency: 'EUR',
    notes: '',
    project_type: '',
    color: '#00d4ff',
    venue_name: '',
    contact_id: '',
  })

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [showMoreDetails, setShowMoreDetails] = useState(false)

  // Load contacts for customer selector
  const { data: contactsData } = useQuery({
    queryKey: ['contacts-list'],
    queryFn: () => contactApi.list({ limit: 200 }),
    staleTime: 1000 * 60 * 5,
  })
  const contacts: Contact[] = contactsData?.data || []

  const { data: project, isLoading: isLoadingProject } = useQuery({
    queryKey: ['project', id],
    queryFn: () => projectApi.getById(id!),
    enabled: isEditing,
  })

  const { mutate: saveProject, isPending } = useMutation({
    mutationFn: async () => {
      // Convert date strings to ISO 8601 for Go backend
      const payload = {
        ...formData,
        start_date: formData.start_date ? new Date(formData.start_date).toISOString() : undefined,
        end_date: formData.end_date ? new Date(formData.end_date).toISOString() : undefined,
      }
      if (isEditing && id) {
        return projectApi.update(id, payload)
      } else {
        return projectApi.create(payload)
      }
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onSuccess: (data: any) => {
      const newId = data?.id || id
      if (newId) {
        navigate(`/projects/${newId}`)
      } else {
        navigate('/projects')
      }
    },
    onError: (error: unknown) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const errorMessage =
        (error as any)?.response?.data?.error ||
        (error as any)?.response?.data?.message ||
        'Fehler beim Speichern des Projekts'
      setErrors({ submit: errorMessage })
    },
  })

  useEffect(() => {
    if (project && isEditing) {
      setFormData({
        name: project.name || '',
        description: project.description || '',
        client_name: project.client_name || '',
        client_email: project.client_email || '',
        client_phone: project.client_phone || '',
        venue_address: {
          street: project.venue_address?.street || '',
          city: project.venue_address?.city || '',
          state: project.venue_address?.state || '',
          postal_code: project.venue_address?.postal_code || '',
          country: project.venue_address?.country || 'AT',
        },
        status: project.status || 'draft',
        start_date: project.start_date ? project.start_date.slice(0, 10) : '',
        end_date: project.end_date ? project.end_date.slice(0, 10) : '',
        budget: project.budget || 0,
        currency: project.currency || 'EUR',
        notes: project.notes || '',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        project_type: (project as any).project_type || '',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        color: (project as any).color || '#00d4ff',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        venue_name: (project as any).venue_name || '',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        contact_id: (project as any).contact_id || '',
      })
    }
  }, [project, isEditing])

  // Auto-expand "Mehr Details" when editing and optional fields have data
  useEffect(() => {
    if (isEditing && project) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const p = project as any
      const hasOptionalData = p.project_type || p.venue_name ||
        p.venue_address?.street || p.venue_address?.city ||
        p.budget > 0 || p.notes || p.color !== '#00d4ff'
      if (hasOptionalData) {
        setShowMoreDetails(true)
      }
    }
  }, [project, isEditing])

  const toggleMoreDetails = useCallback(() => {
    setShowMoreDetails((prev) => !prev)
  }, [])

  const validateForm = () => {
    const newErrors: Record<string, string> = {}

    if (!formData.name?.trim()) {
      newErrors.name = 'Projektname ist erforderlich'
    }
    if (!formData.start_date) {
      newErrors.start_date = 'Startdatum ist erforderlich'
    }
    if (formData.start_date && formData.end_date && formData.end_date < formData.start_date) {
      newErrors.end_date = 'Enddatum darf nicht vor dem Startdatum liegen'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleInputChange = (field: string, value: unknown) => {
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

  const handleAddressChange = (field: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      venue_address: {
        ...prev.venue_address!,
        [field]: value,
      },
    }))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (validateForm()) {
      saveProject()
    }
  }

  if (isLoadingProject) {
    return (
      <div className="equipment-form-page" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-3)', animation: 'pulse-active 1.5s infinite' }}>
            ...
          </div>
          <p style={{ color: 'var(--color-text-secondary)' }}>Projekt wird geladen...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="equipment-form-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">
            {isEditing ? 'Projekt bearbeiten' : 'Neues Projekt'}
          </h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="form-section">
        {errors.submit && (
          <div className="error-message" role="alert">
            {errors.submit}
          </div>
        )}

        {/* === Essential fields (always visible) === */}
        <div style={{ marginBottom: 'var(--spacing-4)' }}>
          <Input
            label="Projektname *"
            value={formData.name}
            onChange={(e) => handleInputChange('name', e.target.value)}
            placeholder="z.B. Firmenfeier Müller GmbH"
            error={errors.name}
            style={{ fontSize: '1.1rem' }}
          />
        </div>

        <div className="form-section__grid">
          <Input
            label="Startdatum"
            type="date"
            value={formData.start_date}
            onChange={(e) => handleInputChange('start_date', e.target.value)}
            error={errors.start_date}
          />
          <Input
            label="Enddatum"
            type="date"
            value={formData.end_date}
            onChange={(e) => handleInputChange('end_date', e.target.value)}
            error={errors.end_date}
          />
        </div>

        <div style={{ marginTop: 'var(--spacing-4)' }}>
          {contacts.length > 0 ? (
            <Select
              label="Kunde"
              options={contacts.map((c) => ({
                value: c.id,
                label: c.company_name ? `${c.name || ''} (${c.company_name})` : c.name || c.email || c.id,
              }))}
              value={formData.contact_id || ''}
              onChange={(e) => {
                const selectedContact = contacts.find((c) => c.id === e.target.value)
                handleInputChange('contact_id', e.target.value)
                if (selectedContact) {
                  handleInputChange('client_name', selectedContact.company_name || selectedContact.name || '')
                  if (selectedContact.email) {
                    handleInputChange('client_email', selectedContact.email)
                  }
                }
              }}
              placeholder="Kunde auswählen oder unten eingeben..."
            />
          ) : (
            <Input
              label="Kunde"
              value={formData.client_name}
              onChange={(e) => handleInputChange('client_name', e.target.value)}
              placeholder="z.B. Müller GmbH"
            />
          )}
        </div>

        {/* === Expandable "Mehr Details" section === */}
        <div style={{ marginTop: 'var(--spacing-5)' }}>
          <button
            type="button"
            onClick={toggleMoreDetails}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 'var(--spacing-2)',
              background: 'none',
              border: '1px solid var(--color-border)',
              borderRadius: 'var(--radius-md)',
              padding: 'var(--spacing-2) var(--spacing-3)',
              color: 'var(--color-text-secondary)',
              cursor: 'pointer',
              fontSize: 'var(--font-size-sm)',
              fontWeight: 'var(--font-weight-medium)',
              width: '100%',
              justifyContent: 'center',
              transition: 'all 0.15s ease',
            }}
          >
            {showMoreDetails ? '▲ Weniger Details' : '▼ Mehr Details'}
          </button>
        </div>

        {showMoreDetails && (
          <div style={{ marginTop: 'var(--spacing-4)' }}>
            <h2 className="form-section__title">Details</h2>
            <div className="form-section__grid">
              <Select
                label="Projekttyp"
                options={PROJECT_TYPE_OPTIONS}
                value={formData.project_type || ''}
                onChange={(e) => handleInputChange('project_type', e.target.value)}
                placeholder="Typ auswählen..."
              />
              <Select
                label="Status"
                options={STATUS_OPTIONS}
                value={formData.status || 'draft'}
                onChange={(e) => handleInputChange('status', e.target.value)}
                placeholder="Status auswählen"
              />
            </div>

            <div style={{ marginTop: 'var(--spacing-3)' }}>
              <TextArea
                label="Veranstaltungsort"
                value={[
                  formData.venue_name || '',
                  formData.venue_address?.street || '',
                  [formData.venue_address?.postal_code || '', formData.venue_address?.city || ''].filter(Boolean).join(' '),
                ].filter(Boolean).join('\n')}
                onChange={(e) => {
                  const lines = e.target.value.split('\n')
                  handleInputChange('venue_name', lines[0] || '')
                  handleAddressChange('street', lines[1] || '')
                  // Parse "PLZ Stadt" from line 3
                  const line3 = lines[2] || ''
                  const plzMatch = line3.match(/^(\d{4,5})\s*(.*)$/)
                  if (plzMatch) {
                    handleAddressChange('postal_code', plzMatch[1])
                    handleAddressChange('city', plzMatch[2])
                  } else {
                    handleAddressChange('city', line3)
                    handleAddressChange('postal_code', '')
                  }
                }}
                placeholder={"Stadthalle Wien\nHauptstraße 1\n1010 Wien"}
                rows={3}
              />
            </div>

            <div className="form-section__grid" style={{ marginTop: 'var(--spacing-3)' }}>
              <Input
                label="Budget (EUR)"
                type="number"
                value={formData.budget || 0}
                onChange={(e) => handleInputChange('budget', parseFloat(e.target.value) || 0)}
                step="0.01"
                min="0"
              />
              <div className="form-group">
                <label className="form-label">Farbe</label>
                <div style={{ display: 'flex', gap: 'var(--spacing-2)', alignItems: 'center' }}>
                  <input
                    type="color"
                    value={formData.color || '#00d4ff'}
                    onChange={(e) => handleInputChange('color', e.target.value)}
                    style={{
                      width: '48px',
                      height: '40px',
                      padding: '2px',
                      border: '1px solid var(--color-border)',
                      borderRadius: 'var(--radius-md)',
                      backgroundColor: 'var(--color-bg-tertiary)',
                      cursor: 'pointer',
                    }}
                  />
                  <Input
                    value={formData.color || '#00d4ff'}
                    onChange={(e) => handleInputChange('color', e.target.value)}
                    placeholder="#00d4ff"
                    style={{ fontFamily: 'monospace' }}
                  />
                </div>
              </div>
            </div>

            {/* Customer details (when selected from dropdown above) */}
            {contacts.length > 0 && (
              <div className="form-section__grid" style={{ marginTop: 'var(--spacing-3)' }}>
                <Input
                  label="Kundenname"
                  value={formData.client_name}
                  onChange={(e) => handleInputChange('client_name', e.target.value)}
                  placeholder="z.B. Müller GmbH"
                />
                <Input
                  label="E-Mail"
                  type="email"
                  value={formData.client_email || ''}
                  onChange={(e) => handleInputChange('client_email', e.target.value)}
                  placeholder="kunde@beispiel.de"
                />
                <Input
                  label="Telefon"
                  type="tel"
                  value={formData.client_phone || ''}
                  onChange={(e) => handleInputChange('client_phone', e.target.value)}
                  placeholder="+43 ..."
                />
              </div>
            )}

            <div style={{ marginTop: 'var(--spacing-3)' }}>
              <TextArea
                label="Notizen"
                value={formData.notes || ''}
                onChange={(e) => handleInputChange('notes', e.target.value)}
                placeholder="Interne Notizen zum Projekt..."
                rows={4}
              />
            </div>
          </div>
        )}

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
              : 'Projekt erstellen'}
          </button>
        </div>
      </form>
    </div>
  )
}

export default ProjectFormPage
