import { useState, useEffect, useMemo, useCallback } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { quoteApi, configApi, contactApi, projectApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import '../Invoices/InvoiceEditor.scss'

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

const TAX_RATE_OPTIONS = [
  { value: '19', label: '19 %' },
  { value: '7', label: '7 %' },
  { value: '0', label: '0 %' },
]

function QuoteNew() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const prefilledProjectId = searchParams.get('project') || ''
  const { addNotification } = useNotificationStore()

  const [clientName, setClientName] = useState('')
  const [clientEmail, setClientEmail] = useState('')
  const [clientAddress, setClientAddress] = useState('')
  const [selectedContactId, setSelectedContactId] = useState('')
  const [selectedProjectId, setSelectedProjectId] = useState(prefilledProjectId)
  const [validUntil, setValidUntil] = useState('')
  const [notes, setNotes] = useState('')
  const [lineItems, setLineItems] = useState<LineItem[]>([{ ...EMPTY_LINE_ITEM }])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [sendAfterSave, setSendAfterSave] = useState(false)

  const { data: kuConfig } = useQuery({
    queryKey: ['config', 'finance.kleinunternehmer'],
    queryFn: () => configApi.get('finance.kleinunternehmer'),
  })
  const isKleinunternehmer = kuConfig?.kleinunternehmer ?? kuConfig?.value ?? false

  const { data: companyConfig } = useQuery({
    queryKey: ['config', 'company.details'],
    queryFn: () => configApi.get('company.details'),
  })

  const { data: contactsData } = useQuery({
    queryKey: ['contacts-for-quote'],
    queryFn: () => contactApi.list({ limit: 200 }),
  })
  const contacts = contactsData?.data || []

  const { data: projectsData } = useQuery({
    queryKey: ['projects-for-quote'],
    queryFn: () => projectApi.list(1, 200),
  })
  const projects = projectsData?.items || projectsData?.data || []

  const companyName = companyConfig?.company_name || companyConfig?.name || 'Firmenname'
  const companyAddress = companyConfig?.address || companyConfig?.company_address || ''

  // Default valid_until to 30 days from now
  useEffect(() => {
    if (!validUntil) {
      const d = new Date()
      d.setDate(d.getDate() + 30)
      setValidUntil(d.toISOString().split('T')[0])
    }
  }, [validUntil])

  // Auto-fill from selected contact
  useEffect(() => {
    if (selectedContactId && contacts.length > 0) {
      const c = contacts.find((ct: any) => ct.id === selectedContactId)
      if (c) {
        setClientName(c.company_name || `${c.first_name || ''} ${c.last_name || ''}`.trim() || '')
        setClientEmail(c.email || '')
        const addr = [c.street, c.house_number, c.zip, c.city].filter(Boolean).join(', ')
        if (addr) setClientAddress(addr)
      }
    }
  }, [selectedContactId, contacts])

  const totals = useMemo(() => {
    let subtotal = 0
    let taxTotal = 0
    lineItems.forEach((item) => {
      const lineTotal = item.quantity * item.unit_price
      subtotal += lineTotal
      if (!isKleinunternehmer) {
        taxTotal += lineTotal * (item.tax_rate / 100)
      }
    })
    return { subtotal, taxTotal, total: subtotal + taxTotal }
  }, [lineItems, isKleinunternehmer])

  const addLineItem = useCallback(() => {
    setLineItems((prev) => [...prev, { ...EMPTY_LINE_ITEM }])
  }, [])

  const removeLineItem = useCallback((index: number) => {
    setLineItems((prev) => prev.filter((_, i) => i !== index))
  }, [])

  const updateLineItem = useCallback((index: number, field: keyof LineItem, value: string | number) => {
    setLineItems((prev) => prev.map((item, i) => (i === index ? { ...item, [field]: value } : item)))
  }, [])

  const validateForm = () => {
    const newErrors: Record<string, string> = {}
    if (!clientName.trim()) newErrors.clientName = 'Kundenname ist erforderlich'
    if (!validUntil) newErrors.validUntil = 'Gueltig-bis-Datum ist erforderlich'
    if (lineItems.length === 0) newErrors.lineItems = 'Mindestens eine Position ist erforderlich'
    const hasEmpty = lineItems.some((item) => !item.name.trim() || item.quantity <= 0 || item.unit_price <= 0)
    if (hasEmpty) newErrors.lineItems = 'Alle Positionen muessen Name, Menge und Einzelpreis haben'
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const { mutate: saveQuote, isPending } = useMutation({
    mutationFn: async () => {
      const data = {
        project_id: selectedProjectId || undefined,
        client_name: clientName,
        client_email: clientEmail,
        client_address: clientAddress,
        valid_until: validUntil ? new Date(validUntil).toISOString() : undefined,
        notes,
        items: lineItems.map((item) => ({
          description: item.name,
          quantity: item.quantity,
          unit_price: item.unit_price,
          tax_rate: isKleinunternehmer ? 0 : item.tax_rate,
        })),
        status: sendAfterSave ? 'sent' : 'draft',
      }
      return quoteApi.create(data)
    },
    onSuccess: (result) => {
      const msg = sendAfterSave ? 'Angebot erstellt und gesendet' : 'Angebot erstellt'
      addNotification(msg, 'success', { title: 'Erfolg', duration: 3000 })
      const quoteId = result?.id
      if (quoteId) {
        navigate(`/quotes/${quoteId}`)
      } else {
        navigate('/quotes')
      }
    },
    onError: (error: unknown) => {
      const errorMessage =
        (error as any)?.response?.data?.error ||
        (error as any)?.response?.data?.message ||
        'Fehler beim Erstellen des Angebots'
      setErrors({ submit: errorMessage })
    },
  })

  const handleSave = (send = false) => {
    setSendAfterSave(send)
    if (validateForm()) {
      saveQuote()
    }
  }

  const handlePrint = () => window.print()

  return (
    <div className="invoice-editor">
      {/* Toolbar */}
      <div className="invoice-editor__toolbar">
        <div className="toolbar-left">
          <button className="btn btn--secondary btn--sm" onClick={() => navigate(-1)}>
            Zurueck
          </button>
          <span className="toolbar-title">Neues Angebot</span>
        </div>
        <div className="toolbar-center">
          {isKleinunternehmer && (
            <span style={{
              padding: '4px 10px',
              background: 'rgba(59, 130, 246, 0.15)',
              border: '1px solid rgba(59, 130, 246, 0.3)',
              borderRadius: '6px',
              fontSize: '0.78rem',
              color: '#60a5fa',
              fontWeight: 600,
            }}>
              Kleinunternehmer
            </span>
          )}
        </div>
        <div className="toolbar-right">
          <button className="btn btn--secondary btn--sm" onClick={handlePrint} title="Drucken / Als PDF">
            Drucken
          </button>
          <button className="btn btn--secondary btn--sm" onClick={() => handleSave(false)} disabled={isPending}>
            {isPending && !sendAfterSave ? 'Speichert...' : 'Speichern'}
          </button>
          <button className="btn btn--primary btn--sm" onClick={() => handleSave(true)} disabled={isPending}>
            {isPending && sendAfterSave ? 'Sendet...' : 'Speichern & Senden'}
          </button>
        </div>
      </div>

      {/* Errors */}
      {Object.keys(errors).length > 0 && (
        <div style={{
          background: 'rgba(239, 68, 68, 0.1)',
          border: '1px solid rgba(239, 68, 68, 0.3)',
          borderRadius: '8px',
          padding: '12px 16px',
          margin: '0 0 16px',
          color: '#ef4444',
          fontSize: '0.85rem',
        }}>
          {Object.values(errors).map((e, i) => <div key={i}>{e}</div>)}
        </div>
      )}

      {/* Document */}
      <div className="invoice-editor__document">
        {/* Header */}
        <div className="document-header">
          <div className="document-header__company">
            <div style={{ fontWeight: 700, fontSize: '1.1rem' }}>{companyName}</div>
            {companyAddress && <div style={{ fontSize: '0.8rem', color: '#888', whiteSpace: 'pre-line' }}>{companyAddress}</div>}
          </div>
          <div className="document-header__meta" style={{ textAlign: 'right' }}>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--color-primary, #00d4ff)', marginBottom: '8px' }}>
              ANGEBOT
            </div>
            <div style={{ fontSize: '0.85rem', color: '#888' }}>
              Gueltig bis:
              <input
                type="date"
                value={validUntil}
                onChange={(e) => setValidUntil(e.target.value)}
                style={{
                  marginLeft: '8px',
                  background: 'transparent',
                  border: '1px solid var(--color-border, #333)',
                  borderRadius: '4px',
                  padding: '2px 6px',
                  color: 'inherit',
                  fontSize: 'inherit',
                }}
              />
            </div>
          </div>
        </div>

        {/* Client Info */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', marginBottom: '24px' }}>
          <div>
            <label style={{ display: 'block', fontSize: '0.75rem', color: '#888', marginBottom: '4px' }}>Kunde</label>
            {contacts.length > 0 && (
              <select
                value={selectedContactId}
                onChange={(e) => setSelectedContactId(e.target.value)}
                style={{
                  width: '100%',
                  padding: '6px 8px',
                  background: 'var(--color-bg-tertiary, #1a1a2e)',
                  border: '1px solid var(--color-border, #333)',
                  borderRadius: '6px',
                  color: 'inherit',
                  marginBottom: '8px',
                  fontSize: '0.85rem',
                }}
              >
                <option value="">Kontakt auswaehlen...</option>
                {contacts.map((c: any) => (
                  <option key={c.id} value={c.id}>
                    {c.company_name || `${c.first_name || ''} ${c.last_name || ''}`.trim() || c.email}
                  </option>
                ))}
              </select>
            )}
            <input
              type="text"
              placeholder="Kundenname"
              value={clientName}
              onChange={(e) => setClientName(e.target.value)}
              className="document-input"
              style={{ width: '100%', marginBottom: '6px' }}
            />
            <input
              type="email"
              placeholder="E-Mail"
              value={clientEmail}
              onChange={(e) => setClientEmail(e.target.value)}
              className="document-input"
              style={{ width: '100%', marginBottom: '6px' }}
            />
            <textarea
              placeholder="Adresse"
              value={clientAddress}
              onChange={(e) => setClientAddress(e.target.value)}
              className="document-input"
              rows={2}
              style={{ width: '100%', resize: 'vertical' }}
            />
          </div>
          <div>
            <label style={{ display: 'block', fontSize: '0.75rem', color: '#888', marginBottom: '4px' }}>Projekt</label>
            <select
              value={selectedProjectId}
              onChange={(e) => setSelectedProjectId(e.target.value)}
              style={{
                width: '100%',
                padding: '6px 8px',
                background: 'var(--color-bg-tertiary, #1a1a2e)',
                border: '1px solid var(--color-border, #333)',
                borderRadius: '6px',
                color: 'inherit',
                fontSize: '0.85rem',
              }}
            >
              <option value="">Kein Projekt</option>
              {(Array.isArray(projects) ? projects : []).map((p: any) => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </div>
        </div>

        {/* Line Items */}
        <table className="line-items-table" style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '16px' }}>
          <thead>
            <tr style={{ borderBottom: '2px solid var(--color-border, #333)', fontSize: '0.8rem', color: '#888', textAlign: 'left' }}>
              <th style={{ padding: '8px 4px', width: '40%' }}>Beschreibung</th>
              <th style={{ padding: '8px 4px', width: '10%', textAlign: 'right' }}>Menge</th>
              <th style={{ padding: '8px 4px', width: '15%', textAlign: 'right' }}>Einzelpreis</th>
              {!isKleinunternehmer && <th style={{ padding: '8px 4px', width: '12%', textAlign: 'right' }}>MwSt</th>}
              <th style={{ padding: '8px 4px', width: '15%', textAlign: 'right' }}>Gesamt</th>
              <th style={{ padding: '8px 4px', width: '8%' }}></th>
            </tr>
          </thead>
          <tbody>
            {lineItems.map((item, idx) => (
              <tr key={idx} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                <td style={{ padding: '6px 4px' }}>
                  <input
                    type="text"
                    value={item.name}
                    onChange={(e) => updateLineItem(idx, 'name', e.target.value)}
                    placeholder="Position..."
                    className="document-input"
                    style={{ width: '100%' }}
                  />
                </td>
                <td style={{ padding: '6px 4px' }}>
                  <input
                    type="number"
                    value={item.quantity}
                    onChange={(e) => updateLineItem(idx, 'quantity', parseInt(e.target.value) || 0)}
                    min="1"
                    className="document-input"
                    style={{ width: '100%', textAlign: 'right' }}
                  />
                </td>
                <td style={{ padding: '6px 4px' }}>
                  <input
                    type="number"
                    value={item.unit_price}
                    onChange={(e) => updateLineItem(idx, 'unit_price', parseFloat(e.target.value) || 0)}
                    min="0"
                    step="0.01"
                    className="document-input"
                    style={{ width: '100%', textAlign: 'right' }}
                  />
                </td>
                {!isKleinunternehmer && (
                  <td style={{ padding: '6px 4px' }}>
                    <select
                      value={String(item.tax_rate)}
                      onChange={(e) => updateLineItem(idx, 'tax_rate', parseFloat(e.target.value))}
                      className="document-input"
                      style={{ width: '100%', textAlign: 'right' }}
                    >
                      {TAX_RATE_OPTIONS.map((opt) => (
                        <option key={opt.value} value={opt.value}>{opt.label}</option>
                      ))}
                    </select>
                  </td>
                )}
                <td style={{ padding: '6px 4px', textAlign: 'right', fontWeight: 600 }}>
                  {formatCurrency(item.quantity * item.unit_price)}
                </td>
                <td style={{ padding: '6px 4px', textAlign: 'center' }}>
                  {lineItems.length > 1 && (
                    <button
                      type="button"
                      onClick={() => removeLineItem(idx)}
                      style={{ background: 'none', border: 'none', color: '#ef4444', cursor: 'pointer', fontSize: '1.1rem' }}
                      title="Position entfernen"
                    >
                      x
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <button
          type="button"
          onClick={addLineItem}
          className="btn btn--secondary btn--sm"
          style={{ marginBottom: '24px' }}
        >
          + Position hinzufuegen
        </button>

        {/* Totals */}
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '24px' }}>
          <div style={{ width: '280px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', fontSize: '0.9rem' }}>
              <span>Zwischensumme (netto):</span>
              <span>{formatCurrency(totals.subtotal)}</span>
            </div>
            {!isKleinunternehmer && (
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', fontSize: '0.9rem' }}>
                <span>MwSt:</span>
                <span>{formatCurrency(totals.taxTotal)}</span>
              </div>
            )}
            <div style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '10px 0',
              fontSize: '1.1rem',
              fontWeight: 700,
              borderTop: '2px solid var(--color-border, #333)',
              marginTop: '4px',
            }}>
              <span>Gesamt:</span>
              <span>{formatCurrency(totals.total)}</span>
            </div>
            {isKleinunternehmer && (
              <div style={{ fontSize: '0.75rem', color: '#888', marginTop: '4px' }}>
                Kein Ausweis von Umsatzsteuer (Kleinunternehmerregelung gem. &sect; 19 UStG)
              </div>
            )}
          </div>
        </div>

        {/* Notes */}
        <div style={{ marginBottom: '16px' }}>
          <label style={{ display: 'block', fontSize: '0.75rem', color: '#888', marginBottom: '4px' }}>Notizen / Bemerkungen</label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Zahlungsbedingungen, Anmerkungen..."
            className="document-input"
            rows={3}
            style={{ width: '100%', resize: 'vertical' }}
          />
        </div>
      </div>
    </div>
  )
}

export default QuoteNew
