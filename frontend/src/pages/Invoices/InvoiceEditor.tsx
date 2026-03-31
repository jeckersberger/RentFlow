import { useState, useEffect, useMemo, useCallback } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { invoiceApi, configApi, contactApi, projectApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { CreateInvoiceDTO } from '../../types/invoice'
import './InvoiceEditor.scss'

// ── Helpers ─────────────────────────────────────────────────────────────

const formatCurrency = (value: number) =>
  new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(value)

const formatDateDE = (iso: string) => {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleDateString('de-DE')
}

// ── Types ───────────────────────────────────────────────────────────────

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

const PAYMENT_TERMS_OPTIONS = [
  { value: '7', label: '7 Tage' },
  { value: '14', label: '14 Tage' },
  { value: '30', label: '30 Tage' },
  { value: '60', label: '60 Tage' },
]

// ── Component ───────────────────────────────────────────────────────────

function InvoiceEditor() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const prefilledProjectId = searchParams.get('project') || ''
  const isEditing = !!id
  const { addNotification } = useNotificationStore()

  // Form state
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

  // Drag state
  const [dragIndex, setDragIndex] = useState<number | null>(null)
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null)

  // ── Data loading ────────────────────────────────────────────────────

  const { data: kuConfig } = useQuery({
    queryKey: ['config', 'finance.kleinunternehmer'],
    queryFn: () => configApi.get('finance.kleinunternehmer'),
  })
  const isKleinunternehmer = kuConfig?.enabled ?? kuConfig?.kleinunternehmer ?? kuConfig?.value ?? false

  const { data: bankConfig } = useQuery({
    queryKey: ['config', 'finance.bank'],
    queryFn: () => configApi.get('finance.bank'),
  })

  const { data: companyConfig } = useQuery({
    queryKey: ['config', 'company.details'],
    queryFn: () => configApi.get('company.details'),
  })

  const { data: paymentTermsConfig } = useQuery({
    queryKey: ['config', 'finance.payment_terms'],
    queryFn: () => configApi.get('finance.payment_terms'),
  })

  const { data: contactsData } = useQuery({
    queryKey: ['contacts-for-invoice'],
    queryFn: () => contactApi.list({ limit: 200 }),
  })
  const contacts = contactsData?.data || []

  const { data: projectsData } = useQuery({
    queryKey: ['projects-for-invoice'],
    queryFn: () => projectApi.list(1, 200),
  })
  const projects = projectsData?.items || projectsData?.data || []

  const { data: invoice, isLoading: isLoadingInvoice } = useQuery({
    queryKey: ['invoice', id],
    queryFn: () => invoiceApi.getById(id!),
    enabled: isEditing,
  })

  // Company info from config
  const companyName = companyConfig?.company_name || companyConfig?.name || 'Firmenname'
  const companyAddress = companyConfig?.address || companyConfig?.company_address || ''

  // ── Effects ─────────────────────────────────────────────────────────

  useEffect(() => {
    const defaultTerms = paymentTermsConfig?.value || paymentTermsConfig?.days || '30'
    if (!isEditing && defaultTerms) {
      setPaymentTerms(String(defaultTerms))
    }
  }, [paymentTermsConfig, isEditing])

  useEffect(() => {
    if (issueDate && paymentTerms) {
      const issue = new Date(issueDate)
      issue.setDate(issue.getDate() + parseInt(paymentTerms, 10))
      setDueDate(issue.toISOString().split('T')[0])
    }
  }, [issueDate, paymentTerms])

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

  // ── Calculations ────────────────────────────────────────────────────

  const { subtotal, taxTotal, total, taxBreakdown } = useMemo(() => {
    let sub = 0
    let tax = 0
    const breakdown: Record<number, number> = {}

    for (const item of lineItems) {
      const lineTotal = item.quantity * item.unit_price
      const rate = isKleinunternehmer ? 0 : item.tax_rate
      const lineTax = lineTotal * (rate / 100)
      sub += lineTotal
      tax += lineTax
      if (rate > 0) {
        breakdown[rate] = (breakdown[rate] || 0) + lineTax
      }
    }
    return { subtotal: sub, taxTotal: tax, total: sub + tax, taxBreakdown: breakdown }
  }, [lineItems, isKleinunternehmer])

  // ── Line item operations ────────────────────────────────────────────

  const updateLineItem = useCallback((index: number, field: keyof LineItem, value: string | number) => {
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
  }, [errors.lineItems])

  const addLineItem = useCallback(() => {
    setLineItems((prev) => [...prev, { ...EMPTY_LINE_ITEM }])
  }, [])

  const removeLineItem = useCallback((index: number) => {
    if (lineItems.length <= 1) return
    setLineItems((prev) => prev.filter((_, i) => i !== index))
  }, [lineItems.length])

  // ── Drag & drop ─────────────────────────────────────────────────────

  const handleDragStart = useCallback((index: number) => {
    setDragIndex(index)
  }, [])

  const handleDragOver = useCallback((e: React.DragEvent, index: number) => {
    e.preventDefault()
    setDragOverIndex(index)
  }, [])

  const handleDrop = useCallback((index: number) => {
    if (dragIndex === null || dragIndex === index) {
      setDragIndex(null)
      setDragOverIndex(null)
      return
    }
    setLineItems((prev) => {
      const updated = [...prev]
      const [moved] = updated.splice(dragIndex, 1)
      updated.splice(index, 0, moved)
      return updated
    })
    setDragIndex(null)
    setDragOverIndex(null)
  }, [dragIndex])

  const handleDragEnd = useCallback(() => {
    setDragIndex(null)
    setDragOverIndex(null)
  }, [])

  // ── Validation & Submit ─────────────────────────────────────────────

  const validateForm = () => {
    const newErrors: Record<string, string> = {}
    if (!clientName.trim()) newErrors.clientName = 'Kundenname ist erforderlich'
    if (!issueDate) newErrors.issueDate = 'Rechnungsdatum ist erforderlich'
    if (!dueDate) newErrors.dueDate = 'Fälligkeitsdatum ist erforderlich'
    if (lineItems.length === 0) newErrors.lineItems = 'Mindestens eine Position ist erforderlich'
    const hasEmpty = lineItems.some((item) => !item.name.trim() || item.quantity <= 0 || item.unit_price <= 0)
    if (hasEmpty) newErrors.lineItems = 'Alle Positionen müssen Name, Menge und Einzelpreis haben'
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

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

  const handleSave = (send = false) => {
    setSendAfterSave(send)
    if (validateForm()) {
      saveInvoice()
    }
  }

  const handlePrint = () => {
    window.print()
  }

  // ── Loading state ───────────────────────────────────────────────────

  if (isLoadingInvoice) {
    return (
      <div className="invoice-editor">
        <div style={{ textAlign: 'center', color: 'var(--color-text-secondary)', padding: '80px 0' }}>
          <div style={{ fontSize: '2rem', marginBottom: '12px' }}>...</div>
          Rechnung wird geladen...
        </div>
      </div>
    )
  }

  // ── Render ──────────────────────────────────────────────────────────

  const bankHolder = bankConfig?.account_holder || bankConfig?.bank_account_holder
  const bankName = bankConfig?.bank_name || bankConfig?.bank
  const bankIban = bankConfig?.iban || bankConfig?.bank_iban
  const bankBic = bankConfig?.bic || bankConfig?.bank_bic

  return (
    <div className="invoice-editor">
      {/* ─── Toolbar ─────────────────────────────────────────────── */}
      <div className="invoice-editor__toolbar">
        <div className="toolbar-left">
          <button className="btn btn--secondary btn--sm" onClick={() => navigate(-1)}>
            Zurück
          </button>
          <span className="toolbar-title">
            {isEditing ? `Rechnung bearbeiten` : 'Neue Rechnung'}
          </span>
        </div>

        <div className="toolbar-center">
          <select
            className="toolbar-select"
            value={paymentTerms}
            onChange={(e) => setPaymentTerms(e.target.value)}
            title="Zahlungskonditionen"
          >
            {PAYMENT_TERMS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
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
          <button
            className="btn btn--secondary btn--sm"
            onClick={handlePrint}
            title="Drucken / Als PDF"
          >
            Drucken
          </button>
          <button
            className="btn btn--secondary btn--sm"
            onClick={() => handleSave(false)}
            disabled={isPending}
          >
            {isPending && !sendAfterSave ? 'Speichert...' : 'Speichern'}
          </button>
          <button
            className="btn btn--primary btn--sm"
            onClick={() => handleSave(true)}
            disabled={isPending}
          >
            {isPending && sendAfterSave ? 'Sendet...' : 'Speichern & Senden'}
          </button>
        </div>
      </div>

      {/* ─── Errors ──────────────────────────────────────────────── */}
      {Object.keys(errors).length > 0 && (
        <div className="editor-error" role="alert">
          {Object.values(errors).map((msg, i) => (
            <div key={i}>{msg}</div>
          ))}
        </div>
      )}

      {/* ─── A4 Paper ────────────────────────────────────────────── */}
      <div className="invoice-paper">

        {/* Kleinunternehmer notice */}
        {isKleinunternehmer && (
          <div className="ku-banner">
            <strong>Kleinunternehmerregelung aktiv</strong> — Keine MwSt. wird berechnet (§19 UStG).
          </div>
        )}

        {/* ── Header ─────────────────────────────────────────────── */}
        <div className="paper-header">
          <div className="company-block">
            <h1 className="company-name">{companyName}</h1>
            <div className="company-address">
              {companyAddress || 'Adresse wird aus Einstellungen geladen'}
            </div>
          </div>
          <div className="invoice-meta">
            <h2 className="invoice-title">Rechnung</h2>
            <div className="invoice-number">
              {isEditing && invoice?.number ? invoice.number : 'Wird automatisch vergeben'}
            </div>
            <div className="meta-row">
              <span className="meta-label">Datum:</span>
              <input
                type="date"
                className="inline-date"
                value={issueDate}
                onChange={(e) => setIssueDate(e.target.value)}
              />
            </div>
            <div className="meta-row">
              <span className="meta-label">Fällig:</span>
              <input
                type="date"
                className="inline-date"
                value={dueDate}
                onChange={(e) => setDueDate(e.target.value)}
              />
            </div>
            {dueDate && (
              <div className="meta-row" style={{ fontSize: '0.82rem', color: '#9ca3af' }}>
                <span>{formatDateDE(dueDate)}</span>
              </div>
            )}
          </div>
        </div>

        {/* ── Client ─────────────────────────────────────────────── */}
        <div className="paper-client">
          <div className="client-label">Rechnungsempfänger</div>
          <div className="client-selectors">
            <select
              className="inline-select"
              value={selectedContactId}
              onChange={(e) => setSelectedContactId(e.target.value)}
            >
              <option value="">-- Kontakt wählen --</option>
              {contacts.map((c: { id: string; company?: string; name?: string; first_name?: string; last_name?: string }) => (
                <option key={c.id} value={c.id}>
                  {c.company || c.name || `${c.first_name || ''} ${c.last_name || ''}`.trim()}
                </option>
              ))}
            </select>
            <select
              className="inline-select"
              value={selectedProjectId}
              onChange={(e) => setSelectedProjectId(e.target.value)}
            >
              <option value="">-- Projekt (optional) --</option>
              {projects.map((p: { id: string; name?: string; title?: string }) => (
                <option key={p.id} value={p.id}>
                  {p.name || p.title || p.id}
                </option>
              ))}
            </select>
          </div>
          <div className="client-fields">
            <div className="client-field full-width">
              <div className="client-field-label">Name *</div>
              <input
                className="inline-input"
                value={clientName}
                onChange={(e) => setClientName(e.target.value)}
                placeholder="Kundenname / Firma"
                style={{ fontWeight: 600, fontSize: '1.05rem' }}
              />
            </div>
            <div className="client-field full-width">
              <div className="client-field-label">Adresse</div>
              <textarea
                className="inline-textarea"
                value={clientAddress}
                onChange={(e) => setClientAddress(e.target.value)}
                placeholder="Straße, PLZ, Ort"
                rows={2}
              />
            </div>
            <div className="client-field">
              <div className="client-field-label">E-Mail</div>
              <input
                className="inline-input"
                type="email"
                value={clientEmail}
                onChange={(e) => setClientEmail(e.target.value)}
                placeholder="email@beispiel.de"
              />
            </div>
          </div>
        </div>

        {/* ── Line Items Table ────────────────────────────────────── */}
        <div style={{ overflowX: 'auto' }}>
          <table className="paper-items">
            <thead>
              <tr>
                <th className="col-drag"></th>
                <th className="col-pos">Pos</th>
                <th className="col-name">Bezeichnung</th>
                <th className="col-desc">Beschreibung</th>
                <th className="col-qty">Menge</th>
                <th className="col-price">Einzelpreis</th>
                {!isKleinunternehmer && <th className="col-tax">MwSt</th>}
                <th className="col-total">Gesamt</th>
                <th className="col-actions"></th>
              </tr>
            </thead>
            <tbody>
              {lineItems.map((item, index) => {
                const lineTotal = item.quantity * item.unit_price
                return (
                  <tr
                    key={index}
                    className={`${dragIndex === index ? 'dragging' : ''} ${dragOverIndex === index ? 'drag-over' : ''}`}
                    draggable
                    onDragStart={() => handleDragStart(index)}
                    onDragOver={(e) => handleDragOver(e, index)}
                    onDrop={() => handleDrop(index)}
                    onDragEnd={handleDragEnd}
                  >
                    <td>
                      <span className="drag-handle" title="Ziehen zum Umsortieren">&#x2630;</span>
                    </td>
                    <td className="pos-cell">{index + 1}</td>
                    <td>
                      <input
                        className="inline-input"
                        value={item.name}
                        onChange={(e) => updateLineItem(index, 'name', e.target.value)}
                        placeholder="Bezeichnung"
                      />
                    </td>
                    <td>
                      <input
                        className="inline-input"
                        value={item.description}
                        onChange={(e) => updateLineItem(index, 'description', e.target.value)}
                        placeholder="Beschreibung"
                      />
                    </td>
                    <td>
                      <input
                        className="inline-input input-qty"
                        type="number"
                        value={item.quantity}
                        onChange={(e) => updateLineItem(index, 'quantity', parseFloat(e.target.value) || 0)}
                        min="0"
                        step="1"
                      />
                    </td>
                    <td>
                      <input
                        className="inline-input input-price"
                        type="number"
                        value={item.unit_price}
                        onChange={(e) => updateLineItem(index, 'unit_price', parseFloat(e.target.value) || 0)}
                        min="0"
                        step="0.01"
                      />
                    </td>
                    {!isKleinunternehmer && (
                      <td>
                        <select
                          className="inline-select"
                          value={String(item.tax_rate)}
                          onChange={(e) => updateLineItem(index, 'tax_rate', parseFloat(e.target.value))}
                        >
                          {TAX_RATE_OPTIONS.map((opt) => (
                            <option key={opt.value} value={opt.value}>{opt.label}</option>
                          ))}
                        </select>
                      </td>
                    )}
                    <td className="total-cell">{formatCurrency(lineTotal)}</td>
                    <td>
                      <button
                        type="button"
                        className="remove-btn"
                        onClick={() => removeLineItem(index)}
                        disabled={lineItems.length <= 1}
                        title="Position entfernen"
                      >
                        &times;
                      </button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        <button type="button" className="add-line-btn" onClick={addLineItem}>
          + Position hinzufügen
        </button>

        {/* ── Totals ─────────────────────────────────────────────── */}
        <div className="paper-totals">
          <div className="totals-box">
            <div className="totals-row">
              <span className="totals-label">Netto</span>
              <span className="totals-value">{formatCurrency(subtotal)}</span>
            </div>
            {!isKleinunternehmer && Object.entries(taxBreakdown).map(([rate, amount]) => (
              <div className="totals-row" key={rate}>
                <span className="totals-label">MwSt. {rate} %</span>
                <span className="totals-value">{formatCurrency(amount)}</span>
              </div>
            ))}
            {!isKleinunternehmer && Object.keys(taxBreakdown).length === 0 && taxTotal === 0 && (
              <div className="totals-row">
                <span className="totals-label">MwSt.</span>
                <span className="totals-value">{formatCurrency(0)}</span>
              </div>
            )}
            <div className="totals-row totals-final">
              <span className="totals-label">Brutto</span>
              <span className="totals-value">{formatCurrency(total)}</span>
            </div>
            {isKleinunternehmer && (
              <div style={{ fontSize: '0.8rem', color: '#6b7280', marginTop: '6px' }}>
                Kein Ausweis von Umsatzsteuer (§19 UStG)
              </div>
            )}
          </div>
        </div>

        {/* ── Bank details ───────────────────────────────────────── */}
        <div className="paper-bank">
          <div className="bank-title">Bankverbindung</div>
          {bankHolder ? (
            <div className="bank-grid">
              <span className="bank-label">Kontoinhaber:</span>
              <span className="bank-value">{bankHolder}</span>
              <span className="bank-label">Bank:</span>
              <span className="bank-value">{bankName || '—'}</span>
              <span className="bank-label">IBAN:</span>
              <span className="bank-value mono">{bankIban || '—'}</span>
              <span className="bank-label">BIC:</span>
              <span className="bank-value mono">{bankBic || '—'}</span>
            </div>
          ) : (
            <p className="bank-empty">
              Bankverbindung wird aus den Einstellungen geladen. Konfigurieren Sie diese unter Einstellungen &gt; Finanzen.
            </p>
          )}
        </div>

        {/* ── Notes ──────────────────────────────────────────────── */}
        <div className="paper-notes">
          <div className="notes-label">Bemerkungen / Zahlungsbedingungen</div>
          <textarea
            className="inline-textarea"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Optionale Anmerkungen zur Rechnung, z.B. Zahlungshinweise oder besondere Vereinbarungen..."
            rows={3}
          />
        </div>

        {/* ── Legal footer ───────────────────────────────────────── */}
        <div className="paper-legal">
          {isKleinunternehmer
            ? 'Gemäß §19 UStG wird keine Umsatzsteuer berechnet. Der ausgewiesene Betrag ist der Endbetrag.'
            : `Zahlbar innerhalb von ${paymentTerms} Tagen ohne Abzug.`
          }
        </div>
      </div>
    </div>
  )
}

export default InvoiceEditor
