import { useState, useEffect, useMemo } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { invoiceApi, configApi, contactApi, projectApi } from '../../services/api'
import { Input } from '../../components/Form/Input'
import { TextArea } from '../../components/Form/TextArea'
import { Select } from '../../components/Form/Select'
import { useNotificationStore } from '../../stores/notificationStore'
import { CreateInvoiceDTO } from '../../types/invoice'
import '../Equipment/Equipment.module.scss'

const formatCurrency = (value: number) =>
  new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(value)

interface LineItem {
  name: string
  description: string
  quantity: number
  unit_price: number
  tax_rate: number
}

const EMPTY_LINE_ITEM: LineItem = {
  name: '',
  description: '',
  quantity: 1,
  unit_price: 0,
  tax_rate: 19,
}

const PAYMENT_TERMS_OPTIONS = [
  { value: '7', label: '7 Tage' },
  { value: '14', label: '14 Tage' },
  { value: '30', label: '30 Tage' },
  { value: '60', label: '60 Tage' },
  { value: 'custom', label: 'Benutzerdefiniert' },
]

const TAX_RATE_OPTIONS = [
  { value: '19', label: '19 %' },
  { value: '7', label: '7 %' },
  { value: '0', label: '0 %' },
]

function InvoiceFormPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const prefilledProjectId = searchParams.get('project') || ''
  const isEditing = !!id
  const { addNotification } = useNotificationStore()

  const [clientName, setClientName] = useState('')
  const [clientEmail, setClientEmail] = useState('')
  const [clientAddress, setClientAddress] = useState('')
  const [selectedContactId, setSelectedContactId] = useState('')
  const [selectedProjectId, setSelectedProjectId] = useState(prefilledProjectId)
  const [issueDate, setIssueDate] = useState(new Date().toISOString().split('T')[0])
  const [dueDate, setDueDate] = useState('')
  const [notes, setNotes] = useState('')
  const [paymentTerms, setPaymentTerms] = useState('30')
  const [lineItems, setLineItems] = useState<LineItem[]>([{ ...EMPTY_LINE_ITEM }])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [sendAfterSave, setSendAfterSave] = useState(false)

  // Load Kleinunternehmer config
  const { data: kuConfig } = useQuery({
    queryKey: ['config', 'finance.kleinunternehmer'],
    queryFn: () => configApi.get('finance.kleinunternehmer'),
  })
  const isKleinunternehmer = kuConfig?.kleinunternehmer ?? kuConfig?.value ?? false

  // Load bank details config
  const { data: bankConfig } = useQuery({
    queryKey: ['config', 'finance.bank'],
    queryFn: () => configApi.get('finance.bank'),
  })

  // Load payment terms config
  const { data: paymentTermsConfig } = useQuery({
    queryKey: ['config', 'finance.payment_terms'],
    queryFn: () => configApi.get('finance.payment_terms'),
  })

  // Load contacts for selector
  const { data: contactsData } = useQuery({
    queryKey: ['contacts-for-invoice'],
    queryFn: () => contactApi.list({ limit: 200 }),
  })
  const contacts = contactsData?.data || []

  // Load projects for selector
  const { data: projectsData } = useQuery({
    queryKey: ['projects-for-invoice'],
    queryFn: () => projectApi.list(1, 200),
  })
  const projects = projectsData?.items || projectsData?.data || []

  // Load existing invoice in edit mode
  const { data: invoice, isLoading: isLoadingInvoice } = useQuery({
    queryKey: ['invoice', id],
    queryFn: () => invoiceApi.getById(id!),
    enabled: isEditing,
  })

  // Set default payment terms and due date
  useEffect(() => {
    const defaultTerms = paymentTermsConfig?.value || paymentTermsConfig?.days || '30'
    if (!isEditing && defaultTerms) {
      setPaymentTerms(String(defaultTerms))
    }
  }, [paymentTermsConfig, isEditing])

  // Auto-calculate due date from payment terms
  useEffect(() => {
    if (issueDate && paymentTerms && paymentTerms !== 'custom') {
      const issue = new Date(issueDate)
      issue.setDate(issue.getDate() + parseInt(paymentTerms, 10))
      setDueDate(issue.toISOString().split('T')[0])
    }
  }, [issueDate, paymentTerms])

  // Populate form in edit mode
  useEffect(() => {
    if (invoice && isEditing) {
      setClientName(invoice.client_name || '')
      setClientEmail(invoice.client_email || '')
      setClientAddress(invoice.client_address || '')
      setSelectedContactId(invoice.client_id || '')
      setSelectedProjectId(invoice.project_id || '')
      setIssueDate(invoice.issue_date ? invoice.issue_date.split('T')[0] : '')
      setDueDate(invoice.due_date ? invoice.due_date.split('T')[0] : '')
      setNotes(invoice.notes || '')
      setPaymentTerms(invoice.payment_terms || '30')
      if (invoice.line_items && invoice.line_items.length > 0) {
        setLineItems(
          invoice.line_items.map((item: { name?: string; description?: string; quantity?: number; unit_price?: number; tax_rate?: number }) => ({
            name: item.name || item.description || '',
            description: item.description || '',
            quantity: item.quantity || 1,
            unit_price: item.unit_price || 0,
            tax_rate: item.tax_rate ?? 19,
          }))
        )
      }
    }
  }, [invoice, isEditing])

  // When contact is selected, fill in name/email/address
  useEffect(() => {
    if (selectedContactId && contacts.length > 0) {
      const contact = contacts.find((c: { id: string }) => c.id === selectedContactId)
      if (contact) {
        setClientName(contact.company || contact.name || `${contact.first_name || ''} ${contact.last_name || ''}`.trim())
        setClientEmail(contact.email || '')
        setClientAddress(contact.address || '')
      }
    }
  }, [selectedContactId, contacts])

  const { mutate: saveInvoice, isPending } = useMutation({
    mutationFn: async () => {
      const payload: CreateInvoiceDTO = {
        number: '',
        client_id: selectedContactId || '',
        project_id: selectedProjectId || undefined,
        issue_date: issueDate,
        due_date: dueDate,
        notes,
        line_items: lineItems.map((item) => ({
          name: item.name,
          description: item.description || item.name,
          quantity: item.quantity,
          unit_price: item.unit_price,
          tax_rate: isKleinunternehmer ? 0 : item.tax_rate,
        })),
      }

      const data = {
        ...payload,
        client_name: clientName,
        client_email: clientEmail,
        client_address: clientAddress,
        payment_terms: paymentTerms,
        is_kleinunternehmer: isKleinunternehmer,
        status: sendAfterSave ? 'sent' : 'draft',
      }

      if (isEditing && id) {
        return invoiceApi.update(id, data)
      }
      return invoiceApi.create(data)
    },
    onSuccess: (result) => {
      const msg = sendAfterSave
        ? 'Rechnung gespeichert und gesendet'
        : isEditing
        ? 'Rechnung aktualisiert'
        : 'Rechnung erstellt'
      addNotification(msg, 'success', { title: 'Erfolg', duration: 3000 })

      const invoiceId = result?.id || id
      if (invoiceId) {
        navigate(`/invoices/${invoiceId}`)
      } else {
        navigate('/invoices')
      }
    },
    onError: (error: unknown) => {
      const errorMessage =
        (error as { response?: { data?: { error?: string; message?: string } } })?.response?.data?.error ||
        (error as { response?: { data?: { error?: string; message?: string } } })?.response?.data?.message ||
        'Fehler beim Speichern der Rechnung'
      setErrors({ submit: errorMessage })
    },
  })

  const { subtotal, taxTotal, total } = useMemo(() => {
    let sub = 0
    let tax = 0
    for (const item of lineItems) {
      const lineTotal = item.quantity * item.unit_price
      const lineTax = isKleinunternehmer ? 0 : lineTotal * (item.tax_rate / 100)
      sub += lineTotal
      tax += lineTax
    }
    return { subtotal: sub, taxTotal: tax, total: sub + tax }
  }, [lineItems, isKleinunternehmer])

  const validateForm = () => {
    const newErrors: Record<string, string> = {}

    if (!clientName.trim()) {
      newErrors.clientName = 'Kundenname ist erforderlich'
    }
    if (!issueDate) {
      newErrors.issueDate = 'Rechnungsdatum ist erforderlich'
    }
    if (!dueDate) {
      newErrors.dueDate = 'Fälligkeitsdatum ist erforderlich'
    }
    if (lineItems.length === 0) {
      newErrors.lineItems = 'Mindestens eine Position ist erforderlich'
    }
    const hasEmptyItem = lineItems.some(
      (item) => !item.name.trim() || item.quantity <= 0 || item.unit_price <= 0
    )
    if (hasEmptyItem) {
      newErrors.lineItems = 'Alle Positionen müssen Name, Menge und Einzelpreis haben'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent, send = false) => {
    e.preventDefault()
    setSendAfterSave(send)
    if (validateForm()) {
      saveInvoice()
    }
  }

  const updateLineItem = (index: number, field: keyof LineItem, value: string | number) => {
    setLineItems((prev) => {
      const updated = [...prev]
      updated[index] = { ...updated[index], [field]: value }
      return updated
    })
    if (errors.lineItems) {
      setErrors((prev) => {
        const updated = { ...prev }
        delete updated.lineItems
        return updated
      })
    }
  }

  const addLineItem = () => {
    setLineItems((prev) => [...prev, { ...EMPTY_LINE_ITEM }])
  }

  const removeLineItem = (index: number) => {
    if (lineItems.length <= 1) return
    setLineItems((prev) => prev.filter((_, i) => i !== index))
  }

  const clearFieldError = (field: string) => {
    if (errors[field]) {
      setErrors((prev) => {
        const updated = { ...prev }
        delete updated[field]
        return updated
      })
    }
  }

  if (isLoadingInvoice) {
    return (
      <div className="invoice-form-page" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
        <div style={{ textAlign: 'center', color: 'var(--color-text-secondary)' }}>
          <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-3)' }}>...</div>
          Rechnung wird geladen...
        </div>
      </div>
    )
  }

  return (
    <div className="invoice-form-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">
            {isEditing ? 'Rechnung bearbeiten' : 'Neue Rechnung'}
          </h1>
          <p className="page-subtitle">
            {isEditing ? `Rechnung ${invoice?.number || ''} bearbeiten` : 'Neue Rechnung erstellen'}
          </p>
        </div>
        <button
          className="btn btn--secondary"
          onClick={() => navigate(-1)}
        >
          Zurück
        </button>
      </div>

      {isKleinunternehmer && (
        <div
          role="status"
          style={{
            marginBottom: 'var(--spacing-4)',
            padding: 'var(--spacing-3) var(--spacing-4)',
            background: 'rgba(59, 130, 246, 0.1)',
            border: '1px solid rgba(59, 130, 246, 0.25)',
            borderRadius: 'var(--radius-md)',
            color: '#60a5fa',
            fontSize: '0.9rem',
          }}
        >
          <strong>Kleinunternehmerregelung aktiv</strong> — Keine MwSt wird berechnet (§19 UStG). Alle Rechnungen werden ohne Umsatzsteuer ausgestellt.
        </div>
      )}

      <form onSubmit={(e) => handleSubmit(e, false)} className="form-section">
        {errors.submit && (
          <div className="error-message" role="alert">
            {errors.submit}
          </div>
        )}

        {/* Kundeninformationen */}
        <h2 className="form-section__title">Kundeninformationen</h2>
        <div className="form-section__grid" style={{ marginBottom: 'var(--spacing-4)' }}>
          <div>
            <label style={{ display: 'block', marginBottom: 'var(--spacing-2)', color: 'var(--color-text-secondary)', fontSize: '0.9rem' }}>
              Kontakt auswählen
            </label>
            <select
              className="form-input"
              value={selectedContactId}
              onChange={(e) => setSelectedContactId(e.target.value)}
              style={{
                width: '100%',
                padding: 'var(--spacing-3)',
                backgroundColor: 'var(--glass-bg-input-strong)',
                border: '1px solid var(--color-border-strong)',
                borderRadius: 'var(--radius-input)',
                color: 'var(--color-text-primary)',
              }}
            >
              <option value="">— Kontakt wählen (optional) —</option>
              {contacts.map((c: { id: string; company?: string; name?: string; first_name?: string; last_name?: string }) => (
                <option key={c.id} value={c.id}>
                  {c.company || c.name || `${c.first_name || ''} ${c.last_name || ''}`.trim()}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label style={{ display: 'block', marginBottom: 'var(--spacing-2)', color: 'var(--color-text-secondary)', fontSize: '0.9rem' }}>
              Projekt (optional)
            </label>
            <select
              className="form-input"
              value={selectedProjectId}
              onChange={(e) => setSelectedProjectId(e.target.value)}
              style={{
                width: '100%',
                padding: 'var(--spacing-3)',
                backgroundColor: 'var(--glass-bg-input-strong)',
                border: '1px solid var(--color-border-strong)',
                borderRadius: 'var(--radius-input)',
                color: 'var(--color-text-primary)',
              }}
            >
              <option value="">— Projekt wählen (optional) —</option>
              {projects.map((p: { id: string; name?: string; title?: string }) => (
                <option key={p.id} value={p.id}>
                  {p.name || p.title || p.id}
                </option>
              ))}
            </select>
          </div>
        </div>
        <div className="form-section__grid">
          <Input
            label="Kundenname *"
            value={clientName}
            onChange={(e) => {
              setClientName(e.target.value)
              clearFieldError('clientName')
            }}
            placeholder="z.B. Musterfirma GmbH"
            error={errors.clientName}
          />
          <Input
            label="E-Mail"
            type="email"
            value={clientEmail}
            onChange={(e) => setClientEmail(e.target.value)}
            placeholder="z.B. kontakt@musterfirma.de"
          />
        </div>
        <TextArea
          label="Adresse"
          value={clientAddress}
          onChange={(e) => setClientAddress(e.target.value)}
          placeholder="Straße, PLZ, Ort"
          rows={3}
        />

        {/* Rechnungsdetails */}
        <h2 className="form-section__title" style={{ marginTop: 'var(--spacing-6)' }}>Rechnungsdetails</h2>
        <div className="form-section__grid">
          <Input
            label="Rechnungsdatum *"
            type="date"
            value={issueDate}
            onChange={(e) => {
              setIssueDate(e.target.value)
              clearFieldError('issueDate')
            }}
            error={errors.issueDate}
          />
          <Input
            label="Fälligkeitsdatum *"
            type="date"
            value={dueDate}
            onChange={(e) => {
              setDueDate(e.target.value)
              clearFieldError('dueDate')
            }}
            error={errors.dueDate}
          />
          <Select
            label="Zahlungskonditionen"
            options={PAYMENT_TERMS_OPTIONS}
            value={paymentTerms}
            onChange={(e) => setPaymentTerms(e.target.value)}
          />
        </div>

        {/* Positionen */}
        <h2 className="form-section__title" style={{ marginTop: 'var(--spacing-6)' }}>Positionen</h2>

        {errors.lineItems && (
          <div className="error-message" role="alert" style={{ marginBottom: 'var(--spacing-4)' }}>
            {errors.lineItems}
          </div>
        )}

        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: 'var(--spacing-4)' }}>
            <thead>
              <tr style={{ borderBottom: '2px solid var(--color-border)' }}>
                <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', minWidth: '160px', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                  Name *
                </th>
                <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', minWidth: '140px', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                  Beschreibung
                </th>
                <th style={{ textAlign: 'right', padding: 'var(--spacing-2) var(--spacing-3)', minWidth: '80px', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                  Menge *
                </th>
                <th style={{ textAlign: 'right', padding: 'var(--spacing-2) var(--spacing-3)', minWidth: '120px', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                  Einzelpreis *
                </th>
                {!isKleinunternehmer && (
                  <th style={{ textAlign: 'right', padding: 'var(--spacing-2) var(--spacing-3)', minWidth: '100px', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                    MwSt-Satz
                  </th>
                )}
                <th style={{ textAlign: 'right', padding: 'var(--spacing-2) var(--spacing-3)', minWidth: '120px', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                  Gesamt
                </th>
                <th style={{ padding: 'var(--spacing-2) var(--spacing-3)', width: '50px' }}></th>
              </tr>
            </thead>
            <tbody>
              {lineItems.map((item, index) => {
                const lineTotal = item.quantity * item.unit_price
                return (
                  <tr key={index} style={{ borderBottom: '1px solid var(--color-border)' }}>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)' }}>
                      <input
                        className="form-input"
                        value={item.name}
                        onChange={(e) => updateLineItem(index, 'name', e.target.value)}
                        placeholder="Bezeichnung"
                        style={{ backgroundColor: 'var(--glass-bg-input-strong)', border: '1px solid var(--color-border-strong)', color: 'var(--color-text-primary)', padding: 'var(--spacing-2)', borderRadius: 'var(--radius-input)' }}
                      />
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)' }}>
                      <input
                        className="form-input"
                        value={item.description}
                        onChange={(e) => updateLineItem(index, 'description', e.target.value)}
                        placeholder="Beschreibung"
                        style={{ backgroundColor: 'var(--glass-bg-input-strong)', border: '1px solid var(--color-border-strong)', color: 'var(--color-text-primary)', padding: 'var(--spacing-2)', borderRadius: 'var(--radius-input)' }}
                      />
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)' }}>
                      <input
                        className="form-input"
                        type="number"
                        value={item.quantity}
                        onChange={(e) => updateLineItem(index, 'quantity', parseFloat(e.target.value) || 0)}
                        min="0"
                        step="1"
                        style={{ textAlign: 'right', backgroundColor: 'var(--glass-bg-input-strong)', border: '1px solid var(--color-border-strong)', color: 'var(--color-text-primary)', padding: 'var(--spacing-2)', borderRadius: 'var(--radius-input)' }}
                      />
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)' }}>
                      <input
                        className="form-input"
                        type="number"
                        value={item.unit_price}
                        onChange={(e) => updateLineItem(index, 'unit_price', parseFloat(e.target.value) || 0)}
                        min="0"
                        step="0.01"
                        style={{ textAlign: 'right', backgroundColor: 'var(--glass-bg-input-strong)', border: '1px solid var(--color-border-strong)', color: 'var(--color-text-primary)', padding: 'var(--spacing-2)', borderRadius: 'var(--radius-input)' }}
                      />
                    </td>
                    {!isKleinunternehmer && (
                      <td style={{ padding: 'var(--spacing-2) var(--spacing-3)' }}>
                        <select
                          className="form-input"
                          value={String(item.tax_rate)}
                          onChange={(e) => updateLineItem(index, 'tax_rate', parseFloat(e.target.value))}
                          style={{
                            textAlign: 'right',
                            minWidth: '90px',
                            backgroundColor: 'var(--glass-bg-input-strong)',
                            border: '1px solid var(--color-border-strong)',
                            color: 'var(--color-text-primary)',
                            padding: 'var(--spacing-2)',
                            borderRadius: 'var(--radius-input)',
                          }}
                        >
                          {TAX_RATE_OPTIONS.map((opt) => (
                            <option key={opt.value} value={opt.value}>{opt.label}</option>
                          ))}
                        </select>
                      </td>
                    )}
                    <td style={{ textAlign: 'right', padding: 'var(--spacing-2) var(--spacing-3)', fontWeight: 600, fontVariantNumeric: 'tabular-nums', color: 'var(--color-text-primary)' }}>
                      {formatCurrency(lineTotal)}
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)' }}>
                      <button
                        type="button"
                        className="btn btn--danger btn--sm"
                        onClick={() => removeLineItem(index)}
                        disabled={lineItems.length <= 1}
                        title="Position entfernen"
                        style={{ padding: 'var(--spacing-1) var(--spacing-2)', fontSize: '0.85rem' }}
                      >
                        X
                      </button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        <button
          type="button"
          className="btn btn--secondary"
          onClick={addLineItem}
          style={{ marginBottom: 'var(--spacing-6)' }}
        >
          + Position hinzufügen
        </button>

        {/* Totals */}
        <div style={{
          display: 'flex',
          justifyContent: 'flex-end',
          marginBottom: 'var(--spacing-6)',
        }}>
          <div style={{
            minWidth: '320px',
            background: 'var(--glass-bg-strong)',
            backdropFilter: 'blur(12px)',
            borderRadius: 'var(--radius-card)',
            padding: 'var(--spacing-4)',
            border: '1px solid rgba(0, 212, 255, 0.1)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 'var(--spacing-3)', color: 'var(--color-text-secondary)' }}>
              <span>Netto</span>
              <span style={{ fontVariantNumeric: 'tabular-nums' }}>{formatCurrency(subtotal)}</span>
            </div>
            {!isKleinunternehmer && (
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 'var(--spacing-3)', color: 'var(--color-text-secondary)' }}>
                <span>MwSt.</span>
                <span style={{ fontVariantNumeric: 'tabular-nums' }}>{formatCurrency(taxTotal)}</span>
              </div>
            )}
            <div style={{
              display: 'flex',
              justifyContent: 'space-between',
              fontWeight: 700,
              fontSize: '1.15rem',
              borderTop: '2px solid rgba(0, 212, 255, 0.2)',
              paddingTop: 'var(--spacing-3)',
              color: 'var(--color-text-primary)',
            }}>
              <span>Brutto</span>
              <span style={{ fontVariantNumeric: 'tabular-nums', color: '#00d4ff' }}>{formatCurrency(total)}</span>
            </div>
            {isKleinunternehmer && (
              <div style={{ marginTop: 'var(--spacing-2)', fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>
                Kein Ausweis von Umsatzsteuer (§19 UStG)
              </div>
            )}
          </div>
        </div>

        {/* Bankverbindung */}
        <h2 className="form-section__title">Bankverbindung</h2>
        <div style={{
          background: 'var(--glass-bg)',
          borderRadius: 'var(--radius-md)',
          padding: 'var(--spacing-4)',
          border: '1px solid var(--color-border)',
          marginBottom: 'var(--spacing-6)',
          color: 'var(--color-text-secondary)',
          fontSize: '0.9rem',
        }}>
          {bankConfig?.account_holder || bankConfig?.bank_account_holder ? (
            <div style={{ display: 'grid', gridTemplateColumns: '120px 1fr', gap: 'var(--spacing-2)' }}>
              <span>Kontoinhaber:</span>
              <span style={{ color: 'var(--color-text-primary)' }}>{bankConfig.account_holder || bankConfig.bank_account_holder}</span>
              <span>Bank:</span>
              <span style={{ color: 'var(--color-text-primary)' }}>{bankConfig.bank_name || bankConfig.bank || '—'}</span>
              <span>IBAN:</span>
              <span style={{ color: 'var(--color-text-primary)', fontFamily: 'monospace' }}>{bankConfig.iban || bankConfig.bank_iban || '—'}</span>
              <span>BIC:</span>
              <span style={{ color: 'var(--color-text-primary)', fontFamily: 'monospace' }}>{bankConfig.bic || bankConfig.bank_bic || '—'}</span>
            </div>
          ) : (
            <p style={{ margin: 0 }}>
              Bankverbindung wird aus den Einstellungen geladen. Konfigurieren Sie diese unter Einstellungen &gt; Finanzen.
            </p>
          )}
        </div>

        {/* Notizen */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 className="form-section__title" style={{ margin: 0 }}>Notizen / Bemerkungen</h2>
          <button
            type="button"
            onClick={() => setNotes('Vielen Dank für Ihren Auftrag. Die aufgeführten Geräte stehen Ihnen im vereinbarten Zeitraum zur Verfügung.')}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '6px',
              padding: '4px 12px',
              background: 'linear-gradient(135deg, rgba(139, 92, 246, 0.15), rgba(0, 212, 255, 0.15))',
              border: '1px solid rgba(139, 92, 246, 0.3)',
              borderRadius: 'var(--radius-sm)',
              color: 'var(--color-accent-light)',
              fontSize: 'var(--font-size-xs)',
              fontWeight: 600,
              cursor: 'pointer',
              transition: 'all 150ms ease-in-out',
            }}
            title="KI-generierten Text einfügen"
          >
            KI Text generieren
          </button>
        </div>
        <TextArea
          label=""
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Optionale Anmerkungen zur Rechnung, z.B. Zahlungshinweise oder besondere Vereinbarungen..."
          rows={4}
        />

        {/* Buttons */}
        <div className="form-section__footer">
          <button
            type="button"
            className="btn btn--secondary"
            onClick={() => navigate(-1)}
          >
            Abbrechen
          </button>
          <div style={{ display: 'flex', gap: 'var(--spacing-3)', marginLeft: 'auto' }}>
            <button
              type="submit"
              className="btn btn--secondary"
              disabled={isPending}
            >
              {isPending && !sendAfterSave
                ? 'Wird gespeichert...'
                : 'Speichern'}
            </button>
            <button
              type="button"
              className="btn btn--primary"
              disabled={isPending}
              onClick={(e) => handleSubmit(e as unknown as React.FormEvent, true)}
            >
              {isPending && sendAfterSave
                ? 'Wird gesendet...'
                : 'Speichern & Senden'}
            </button>
          </div>
        </div>
      </form>
    </div>
  )
}

export default InvoiceFormPage
