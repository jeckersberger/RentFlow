import { useState, useEffect } from 'react'
import { Modal } from '../../components/Modal/Modal'
import { ChevronDown, ChevronUp } from 'lucide-react'
import './Contacts.scss'

export interface ContactFormData {
  type: 'company' | 'person'
  company_name: string
  salutation: string
  first_name: string
  last_name: string
  email: string
  phone: string
  mobile: string
  website: string
  street: string
  house_number: string
  zip: string
  city: string
  country: string
  vat_id: string
  trade_register: string
  contact_person: string
  notes: string
  tags: string[]
}

const emptyForm: ContactFormData = {
  type: 'company',
  company_name: '',
  salutation: '',
  first_name: '',
  last_name: '',
  email: '',
  phone: '',
  mobile: '',
  website: '',
  street: '',
  house_number: '',
  zip: '',
  city: '',
  country: 'Deutschland',
  vat_id: '',
  trade_register: '',
  contact_person: '',
  notes: '',
  tags: [],
}

interface ContactFormProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: ContactFormData) => void
  initialData?: Partial<ContactFormData>
  isLoading?: boolean
  title?: string
}

export default function ContactForm({
  isOpen,
  onClose,
  onSubmit,
  initialData,
  isLoading = false,
  title = 'Kontakt hinzufügen',
}: ContactFormProps) {
  const [form, setForm] = useState<ContactFormData>({ ...emptyForm })
  const [showMore, setShowMore] = useState(false)

  useEffect(() => {
    if (initialData) {
      setForm({ ...emptyForm, ...initialData })
      // Show advanced if any advanced field has data
      if (initialData.street || initialData.vat_id || initialData.website || initialData.mobile || initialData.notes) {
        setShowMore(true)
      }
    } else {
      setForm({ ...emptyForm })
      setShowMore(false)
    }
  }, [initialData, isOpen])

  const update = (field: keyof ContactFormData, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }))
    // Auto-detect type from company_name
    if (field === 'company_name' && value.trim()) {
      setForm((prev) => ({ ...prev, [field]: value, type: 'company' }))
    }
  }

  const handleSubmit = () => onSubmit(form)

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '10px 14px',
    borderRadius: '8px',
    border: '1px solid var(--color-border-strong, #1e293b)',
    background: 'var(--glass-bg-input-strong, #0f172a)',
    color: 'var(--color-text-primary, #e2e8f0)',
    fontSize: '14px',
    fontFamily: 'inherit',
  }

  const labelStyle: React.CSSProperties = {
    display: 'block',
    fontSize: '13px',
    fontWeight: 500,
    color: 'var(--color-text-secondary, #94a3b8)',
    marginBottom: '4px',
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={title}
      size="md"
      footer={
        <div style={{ display: 'flex', gap: '8px' }}>
          <button className="btn btn--secondary" onClick={onClose}>
            Abbrechen
          </button>
          <button
            className="btn btn--primary"
            onClick={handleSubmit}
            disabled={isLoading || (!form.company_name && !form.last_name)}
          >
            {isLoading ? 'Wird gespeichert...' : 'Speichern'}
          </button>
        </div>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
        {/* ESSENTIAL FIELDS — always visible */}
        <div>
          <label style={labelStyle}>Firma oder Name *</label>
          <input
            style={inputStyle}
            value={form.company_name || `${form.first_name} ${form.last_name}`.trim()}
            onChange={(e) => {
              const val = e.target.value
              // If it contains a space and no company_name, treat as person name
              if (!form.company_name && val.includes(' ')) {
                const parts = val.split(' ')
                update('first_name', parts.slice(0, -1).join(' '))
                update('last_name', parts[parts.length - 1])
              } else {
                update('company_name', val)
              }
            }}
            placeholder='z.B. "TechCorp GmbH" oder "Max Mustermann"'
            autoFocus
          />
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
          <div>
            <label style={labelStyle}>E-Mail</label>
            <input style={inputStyle} type="email" value={form.email} onChange={e => update('email', e.target.value)} placeholder="email@beispiel.de" />
          </div>
          <div>
            <label style={labelStyle}>Telefon</label>
            <input style={inputStyle} value={form.phone} onChange={e => update('phone', e.target.value)} placeholder="+49 89 123456" />
          </div>
        </div>

        {/* EXPANDABLE — more details */}
        <button
          onClick={() => setShowMore(!showMore)}
          type="button"
          style={{
            display: 'flex', alignItems: 'center', gap: '6px',
            background: 'none', border: 'none', color: 'var(--color-text-muted, #64748b)',
            fontSize: '13px', cursor: 'pointer', padding: '4px 0',
          }}
        >
          {showMore ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
          {showMore ? 'Weniger Details' : 'Mehr Details (Adresse, USt-ID, Notizen...)'}
        </button>

        {showMore && (
          <div style={{
            display: 'flex', flexDirection: 'column', gap: '12px',
            padding: '16px', borderRadius: '8px',
            background: 'rgba(15, 23, 42, 0.5)',
            border: '1px solid var(--color-border, #1e293b)',
          }}>
            {/* Person details if company */}
            {form.company_name && (
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                <div>
                  <label style={labelStyle}>Ansprechpartner Vorname</label>
                  <input style={inputStyle} value={form.first_name} onChange={e => update('first_name', e.target.value)} placeholder="Max" />
                </div>
                <div>
                  <label style={labelStyle}>Nachname</label>
                  <input style={inputStyle} value={form.last_name} onChange={e => update('last_name', e.target.value)} placeholder="Mustermann" />
                </div>
              </div>
            )}

            <div>
              <label style={labelStyle}>Adresse</label>
              <textarea
                style={{ ...inputStyle, resize: 'vertical' }}
                rows={2}
                value={[form.street, form.house_number, form.zip, form.city].filter(Boolean).join(', ') || ''}
                onChange={e => {
                  // Parse address parts from single field
                  const val = e.target.value
                  const parts = val.split(',').map(p => p.trim())
                  update('street', parts[0] || '')
                  update('city', parts[parts.length - 1] || '')
                  if (parts.length >= 3) update('zip', parts[parts.length - 2] || '')
                }}
                placeholder="Musterstraße 42, 80331, München"
              />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
              <div>
                <label style={labelStyle}>Mobil</label>
                <input style={inputStyle} value={form.mobile} onChange={e => update('mobile', e.target.value)} placeholder="+49 170 1234567" />
              </div>
              <div>
                <label style={labelStyle}>Website</label>
                <input style={inputStyle} value={form.website} onChange={e => update('website', e.target.value)} placeholder="https://beispiel.de" />
              </div>
            </div>

            {form.company_name && (
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                <div>
                  <label style={labelStyle}>USt-IdNr.</label>
                  <input style={inputStyle} value={form.vat_id} onChange={e => update('vat_id', e.target.value)} placeholder="DE123456789" />
                </div>
                <div>
                  <label style={labelStyle}>Handelsregister</label>
                  <input style={inputStyle} value={form.trade_register} onChange={e => update('trade_register', e.target.value)} placeholder="HRB 12345" />
                </div>
              </div>
            )}

            <div>
              <label style={labelStyle}>Notizen</label>
              <textarea
                style={{ ...inputStyle, resize: 'vertical' }}
                rows={2}
                value={form.notes}
                onChange={e => update('notes', e.target.value)}
                placeholder="Interne Notizen..."
              />
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}
