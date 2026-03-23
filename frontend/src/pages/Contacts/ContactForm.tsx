import { useState, useEffect } from 'react'
import { Input } from '../../components/Form/Input'
import { Modal } from '../../components/Modal/Modal'
import styles from './Contacts.module.scss'

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

  useEffect(() => {
    if (initialData) {
      setForm({ ...emptyForm, ...initialData })
    } else {
      setForm({ ...emptyForm })
    }
  }, [initialData, isOpen])

  const handleChange = (field: keyof ContactFormData, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  const handleSubmit = () => {
    onSubmit(form)
  }

  const displayName = form.type === 'company'
    ? form.company_name || 'Neuer Kontakt'
    : `${form.first_name} ${form.last_name}`.trim() || 'Neuer Kontakt'

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={title}
      size="lg"
      footer={
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button className={styles.btn + ' ' + styles['btn--secondary']} onClick={onClose}>
            Abbrechen
          </button>
          <button
            className={styles.btn + ' ' + styles['btn--primary']}
            onClick={handleSubmit}
            disabled={isLoading || (!form.company_name && !form.last_name)}
          >
            {isLoading ? 'Wird gespeichert...' : 'Speichern'}
          </button>
        </div>
      }
    >
      <div>
        {/* Type Selection */}
        <div className={styles['radio-group']}>
          <label
            className={`${styles['radio-option']} ${form.type === 'company' ? styles['radio-option--active'] : ''}`}
          >
            <input
              type="radio"
              name="contact-type"
              value="company"
              checked={form.type === 'company'}
              onChange={() => handleChange('type', 'company')}
            />
            Firma
          </label>
          <label
            className={`${styles['radio-option']} ${form.type === 'person' ? styles['radio-option--active'] : ''}`}
          >
            <input
              type="radio"
              name="contact-type"
              value="person"
              checked={form.type === 'person'}
              onChange={() => handleChange('type', 'person')}
            />
            Person
          </label>
        </div>

        <div className={styles['form-grid']}>
          {form.type === 'company' && (
            <div className={styles['form-grid-full']}>
              <Input
                label="Firmenname *"
                value={form.company_name}
                onChange={(e) => handleChange('company_name', e.target.value)}
                placeholder="z.B. TechCorp GmbH"
              />
            </div>
          )}

          <Input
            label="Anrede"
            value={form.salutation}
            onChange={(e) => handleChange('salutation', e.target.value)}
            placeholder="Herr / Frau / Divers"
          />

          <Input
            label="Vorname"
            value={form.first_name}
            onChange={(e) => handleChange('first_name', e.target.value)}
            placeholder="Max"
          />

          <Input
            label={form.type === 'person' ? 'Nachname *' : 'Nachname'}
            value={form.last_name}
            onChange={(e) => handleChange('last_name', e.target.value)}
            placeholder="Mustermann"
          />

          <Input
            label="E-Mail"
            type="email"
            value={form.email}
            onChange={(e) => handleChange('email', e.target.value)}
            placeholder="email@beispiel.de"
          />

          <Input
            label="Telefon"
            value={form.phone}
            onChange={(e) => handleChange('phone', e.target.value)}
            placeholder="+49 89 123456"
          />

          <Input
            label="Mobil"
            value={form.mobile}
            onChange={(e) => handleChange('mobile', e.target.value)}
            placeholder="+49 170 1234567"
          />

          <Input
            label="Website"
            value={form.website}
            onChange={(e) => handleChange('website', e.target.value)}
            placeholder="https://beispiel.de"
          />

          {/* Address */}
          <Input
            label="Straße"
            value={form.street}
            onChange={(e) => handleChange('street', e.target.value)}
            placeholder="Musterstraße"
          />

          <Input
            label="Hausnummer"
            value={form.house_number}
            onChange={(e) => handleChange('house_number', e.target.value)}
            placeholder="42"
          />

          <Input
            label="PLZ"
            value={form.zip}
            onChange={(e) => handleChange('zip', e.target.value)}
            placeholder="80331"
          />

          <Input
            label="Stadt"
            value={form.city}
            onChange={(e) => handleChange('city', e.target.value)}
            placeholder="München"
          />

          <Input
            label="Land"
            value={form.country}
            onChange={(e) => handleChange('country', e.target.value)}
            placeholder="Deutschland"
          />

          {/* Company-specific fields */}
          {form.type === 'company' && (
            <>
              <Input
                label="USt-ID"
                value={form.vat_id}
                onChange={(e) => handleChange('vat_id', e.target.value)}
                placeholder="DE123456789"
              />

              <Input
                label="Handelsregister"
                value={form.trade_register}
                onChange={(e) => handleChange('trade_register', e.target.value)}
                placeholder="HRB 12345"
              />

              <Input
                label="Ansprechpartner"
                value={form.contact_person}
                onChange={(e) => handleChange('contact_person', e.target.value)}
                placeholder="Name des Ansprechpartners"
              />
            </>
          )}

          {/* Notes - full width */}
          <div className={styles['form-grid-full']}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
              <label style={{
                fontSize: 'var(--font-size-sm)',
                fontWeight: 'var(--font-weight-medium)' as unknown as number,
                color: 'var(--color-text-secondary)',
              }}>
                Notizen
              </label>
              <textarea
                value={form.notes}
                onChange={(e) => handleChange('notes', e.target.value)}
                placeholder="Interne Notizen zum Kontakt..."
                rows={3}
                style={{
                  width: '100%',
                  background: 'var(--glass-bg-input-strong)',
                  border: '1px solid var(--color-border-strong)',
                  borderRadius: 'var(--radius-input)',
                  padding: 'var(--padding-md)',
                  color: 'var(--color-text-primary)',
                  fontSize: 'var(--font-size-sm)',
                  resize: 'vertical',
                  fontFamily: 'inherit',
                }}
              />
            </div>
          </div>
        </div>

        {/* Preview */}
        <div style={{
          marginTop: 'var(--spacing-4)',
          padding: 'var(--spacing-3)',
          background: 'rgba(0, 212, 255, 0.04)',
          borderRadius: 'var(--radius-md)',
          border: '1px solid rgba(0, 212, 255, 0.1)',
        }}>
          <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
            Vorschau: {form.type === 'company' ? 'Firma' : 'Person'} - {displayName}
          </span>
        </div>
      </div>
    </Modal>
  )
}
