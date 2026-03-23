import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { invoiceApi, configApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Modal } from '../../components/Modal/Modal'
import { InvoiceStatus, DunningEntry } from '../../types/invoice'
import '../Equipment/Equipment.module.scss'

const formatCurrency = (value: number) =>
  new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(value)

const DUNNING_LEVELS = [
  { level: 1, label: '1. Mahnung', description: 'Freundliche Zahlungserinnerung' },
  { level: 2, label: '2. Mahnung', description: 'Nachdrückliche Zahlungsaufforderung' },
  { level: 3, label: '3. Mahnung', description: 'Letzte Mahnung vor rechtlichen Schritten' },
]

function InvoiceDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [showCancelModal, setShowCancelModal] = useState(false)
  const [showDunningModal, setShowDunningModal] = useState(false)
  const [dunningLevel, setDunningLevel] = useState(1)

  const {
    data: invoice,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['invoice', id],
    queryFn: () => invoiceApi.getById(id!),
    enabled: !!id,
  })

  // Load Kleinunternehmer config
  const { data: kuConfig } = useQuery({
    queryKey: ['config', 'finance.kleinunternehmer'],
    queryFn: () => configApi.get('finance.kleinunternehmer'),
  })
  const isKleinunternehmer = invoice?.is_kleinunternehmer || kuConfig?.kleinunternehmer || kuConfig?.value || false

  const isOverdue = invoice
    ? (invoice.status === 'overdue' || (
        new Date(invoice.due_date) < new Date() &&
        invoice.status !== 'paid' &&
        invoice.status !== 'cancelled' &&
        invoice.status !== 'draft'
      ))
    : false

  const effectiveStatus = isOverdue && invoice?.status !== 'overdue' ? 'overdue' : invoice?.status

  // Mark as paid mutation
  const { mutate: markAsPaid, isPending: isMarkingPaid } = useMutation({
    mutationFn: () => invoiceApi.update(id!, { status: 'paid' as InvoiceStatus }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invoice', id] })
      queryClient.invalidateQueries({ queryKey: ['invoice-list'] })
      addNotification('Rechnung als bezahlt markiert', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Aktualisieren', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  // Send email mutation
  const { mutate: sendEmail, isPending: isSending } = useMutation({
    mutationFn: () => invoiceApi.sendEmail(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invoice', id] })
      addNotification('Rechnung per E-Mail gesendet', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('E-Mail-Versand fehlgeschlagen. Bitte prüfen Sie die SMTP-Einstellungen.', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  // Cancel mutation
  const { mutate: cancelInvoice, isPending: isCancelling } = useMutation({
    mutationFn: () => invoiceApi.update(id!, { status: 'cancelled' as InvoiceStatus }),
    onSuccess: () => {
      setShowCancelModal(false)
      queryClient.invalidateQueries({ queryKey: ['invoice', id] })
      queryClient.invalidateQueries({ queryKey: ['invoice-list'] })
      addNotification('Rechnung storniert', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Stornieren', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  // Create dunning mutation
  const { mutate: createDunning, isPending: isCreatingDunning } = useMutation({
    mutationFn: (level: number) => invoiceApi.createDunning(id!, level),
    onSuccess: () => {
      setShowDunningModal(false)
      queryClient.invalidateQueries({ queryKey: ['invoice', id] })
      addNotification(`Mahnung (Stufe ${dunningLevel}) erstellt`, 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Erstellen der Mahnung', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  // PDF download handler
  const handleDownloadPDF = async () => {
    try {
      const blob = await invoiceApi.getPDF(id!)
      const url = window.URL.createObjectURL(new Blob([blob]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `Rechnung_${invoice?.number || id}.pdf`)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    } catch {
      addNotification('PDF-Download nicht verfügbar. PDF-Generierung wird noch implementiert.', 'info', { title: 'Info', duration: 4000 })
    }
  }

  if (isLoading) {
    return (
      <div className="invoice-detail-page" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
        <div style={{ textAlign: 'center', color: 'var(--color-text-secondary)' }}>
          <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-3)' }}>...</div>
          Rechnung wird geladen...
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="invoice-detail-page">
        <div className="error-message" role="alert">
          Fehler beim Laden der Rechnung. Bitte versuchen Sie es später erneut.
        </div>
        <button className="btn btn--secondary" onClick={() => navigate('/invoices')}>
          Zurück zur Übersicht
        </button>
      </div>
    )
  }

  if (!invoice) {
    return (
      <div className="invoice-detail-page">
        <div className="error-message">Rechnung nicht gefunden</div>
        <button className="btn btn--secondary" onClick={() => navigate('/invoices')}>
          Zurück zur Übersicht
        </button>
      </div>
    )
  }

  const existingDunning: DunningEntry[] = invoice.dunning_entries || []
  const nextDunningLevel = existingDunning.length > 0
    ? Math.max(...existingDunning.map((d: DunningEntry) => d.level)) + 1
    : 1

  const paidAmount = (invoice.payments || []).reduce((sum: number, p: { amount: number }) => sum + p.amount, 0)
  const remainingAmount = invoice.total - paidAmount

  return (
    <div className="invoice-detail-page">
      {/* Header */}
      <div className="page-header">
        <div>
          <h1 className="page-title">Rechnung {invoice.number}</h1>
          <p className="page-subtitle" style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-3)' }}>
            <StatusBadge status={effectiveStatus || invoice.status} />
            <span style={{ color: 'var(--color-text-secondary)' }}>
              vom {new Date(invoice.issue_date).toLocaleDateString('de-DE')}
            </span>
          </p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          <button className="btn btn--secondary" onClick={() => navigate('/invoices')}>
            Zurück
          </button>
          <button className="btn btn--secondary" onClick={() => navigate(`/invoices/${id}/edit`)}>
            Bearbeiten
          </button>
        </div>
      </div>

      {/* Action buttons */}
      <div style={{
        display: 'flex',
        gap: 'var(--spacing-3)',
        flexWrap: 'wrap',
        padding: 'var(--spacing-4)',
        background: 'var(--glass-bg)',
        backdropFilter: 'blur(12px)',
        borderRadius: 'var(--radius-card)',
        border: '1px solid var(--color-border)',
        marginBottom: 'var(--spacing-6)',
      }}>
        <button className="btn btn--secondary" onClick={handleDownloadPDF}>
          Als PDF herunterladen
        </button>
        <button
          className="btn btn--secondary"
          onClick={() => sendEmail()}
          disabled={isSending}
        >
          {isSending ? 'Wird gesendet...' : 'Per E-Mail senden'}
        </button>
        {invoice.status !== 'paid' && invoice.status !== 'cancelled' && (
          <button
            className="btn btn--primary"
            onClick={() => markAsPaid()}
            disabled={isMarkingPaid}
          >
            {isMarkingPaid ? 'Wird aktualisiert...' : 'Als bezahlt markieren'}
          </button>
        )}
        {(isOverdue || invoice.status === 'overdue') && invoice.status !== 'paid' && invoice.status !== 'cancelled' && (
          <button
            className="btn btn--secondary"
            onClick={() => {
              setDunningLevel(nextDunningLevel)
              setShowDunningModal(true)
            }}
            style={{ borderColor: 'rgba(239, 68, 68, 0.4)', color: '#f87171' }}
          >
            Mahnung erstellen
          </button>
        )}
        {invoice.status !== 'cancelled' && invoice.status !== 'paid' && (
          <button
            className="btn btn--danger"
            onClick={() => setShowCancelModal(true)}
          >
            Stornieren
          </button>
        )}
      </div>

      {/* Main grid: Details + Summary */}
      <div className="detail-grid">
        {/* Left: Invoice Details + Customer */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-6)' }}>
          {/* Customer info */}
          <div className="detail-card">
            <h2 className="detail-card__title">Kunde</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <span className="detail-card__row-label">Name</span>
                <span className="detail-card__row-value">{invoice.client_name || '—'}</span>
              </div>
              {invoice.client_email && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">E-Mail</span>
                  <span className="detail-card__row-value">{invoice.client_email}</span>
                </div>
              )}
              {invoice.client_address && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">Adresse</span>
                  <span className="detail-card__row-value" style={{ whiteSpace: 'pre-line' }}>{invoice.client_address}</span>
                </div>
              )}
              {invoice.project_name && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">Projekt</span>
                  <span className="detail-card__row-value">{invoice.project_name}</span>
                </div>
              )}
            </div>
          </div>

          {/* Invoice details */}
          <div className="detail-card">
            <h2 className="detail-card__title">Rechnungsdetails</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <span className="detail-card__row-label">Rechnungsnr.</span>
                <span className="detail-card__row-value" style={{ fontFamily: 'monospace', color: 'var(--color-primary)' }}>{invoice.number}</span>
              </div>
              <div className="detail-card__row">
                <span className="detail-card__row-label">Ausstellungsdatum</span>
                <span className="detail-card__row-value">{new Date(invoice.issue_date).toLocaleDateString('de-DE')}</span>
              </div>
              <div className="detail-card__row">
                <span className="detail-card__row-label">Fälligkeitsdatum</span>
                <span className="detail-card__row-value" style={{ color: isOverdue ? 'var(--color-danger)' : 'inherit', fontWeight: isOverdue ? 600 : 400 }}>
                  {new Date(invoice.due_date).toLocaleDateString('de-DE')}
                  {isOverdue && ' (überfällig)'}
                </span>
              </div>
              {invoice.payment_terms && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">Zahlungsziel</span>
                  <span className="detail-card__row-value">{invoice.payment_terms} Tage</span>
                </div>
              )}
              {invoice.paid_date && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">Zahlungsdatum</span>
                  <span className="detail-card__row-value" style={{ color: '#34d399' }}>
                    {new Date(invoice.paid_date).toLocaleDateString('de-DE')}
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right: Summary */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-6)' }}>
          <div className="detail-card">
            <h2 className="detail-card__title">Zusammenfassung</h2>
            <div className="detail-card__content">
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: 'var(--spacing-2) 0', color: 'var(--color-text-secondary)' }}>
                <span>Netto</span>
                <span style={{ fontVariantNumeric: 'tabular-nums' }}>{formatCurrency(invoice.subtotal)}</span>
              </div>

              {!isKleinunternehmer && invoice.tax_total != null && invoice.tax_total > 0 && (
                <div style={{ display: 'flex', justifyContent: 'space-between', padding: 'var(--spacing-2) 0', color: 'var(--color-text-secondary)' }}>
                  <span>MwSt.</span>
                  <span style={{ fontVariantNumeric: 'tabular-nums' }}>{formatCurrency(invoice.tax_total)}</span>
                </div>
              )}

              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                padding: 'var(--spacing-3) 0',
                borderTop: '2px solid rgba(0, 212, 255, 0.2)',
                marginTop: 'var(--spacing-2)',
                fontWeight: 700,
                fontSize: '1.2rem',
              }}>
                <span>Brutto</span>
                <span style={{ fontVariantNumeric: 'tabular-nums', color: '#00d4ff' }}>{formatCurrency(invoice.total)}</span>
              </div>

              {isKleinunternehmer && (
                <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)', marginTop: 'var(--spacing-2)' }}>
                  {invoice.kleinunternehmer_text || 'Gemäß §19 UStG wird keine Umsatzsteuer berechnet.'}
                </div>
              )}

              {invoice.status === 'partial' && (
                <>
                  <div style={{ display: 'flex', justifyContent: 'space-between', padding: 'var(--spacing-2) 0', marginTop: 'var(--spacing-3)', borderTop: '1px solid var(--color-border)', color: '#34d399' }}>
                    <span>Bezahlt</span>
                    <span style={{ fontVariantNumeric: 'tabular-nums' }}>{formatCurrency(paidAmount)}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', padding: 'var(--spacing-2) 0', fontWeight: 600, color: '#fbbf24' }}>
                    <span>Offen</span>
                    <span style={{ fontVariantNumeric: 'tabular-nums' }}>{formatCurrency(remainingAmount)}</span>
                  </div>
                </>
              )}
            </div>
          </div>

          {/* Status indicator */}
          <div className="detail-card" style={{
            background: effectiveStatus === 'paid'
              ? 'rgba(16, 185, 129, 0.08)'
              : effectiveStatus === 'overdue'
              ? 'rgba(239, 68, 68, 0.08)'
              : effectiveStatus === 'cancelled'
              ? 'rgba(107, 114, 128, 0.08)'
              : 'rgba(59, 130, 246, 0.08)',
            borderColor: effectiveStatus === 'paid'
              ? 'rgba(16, 185, 129, 0.2)'
              : effectiveStatus === 'overdue'
              ? 'rgba(239, 68, 68, 0.2)'
              : effectiveStatus === 'cancelled'
              ? 'rgba(107, 114, 128, 0.2)'
              : 'rgba(59, 130, 246, 0.2)',
          }}>
            <div style={{ textAlign: 'center' }}>
              <StatusBadge status={effectiveStatus || invoice.status} size="lg" />
              {isOverdue && (
                <p style={{ margin: 'var(--spacing-2) 0 0 0', fontSize: '0.85rem', color: '#f87171' }}>
                  Seit {Math.floor((Date.now() - new Date(invoice.due_date).getTime()) / (1000 * 60 * 60 * 24))} Tagen überfällig
                </p>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Line items table */}
      {invoice.line_items && invoice.line_items.length > 0 && (
        <div className="detail-card" style={{ marginTop: 'var(--spacing-6)' }}>
          <h2 className="detail-card__title">Positionen</h2>
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '2px solid var(--color-border)' }}>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                    Pos.
                  </th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                    Bezeichnung
                  </th>
                  <th style={{ textAlign: 'right', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                    Menge
                  </th>
                  <th style={{ textAlign: 'right', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                    Einzelpreis
                  </th>
                  {!isKleinunternehmer && (
                    <th style={{ textAlign: 'right', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                      MwSt.
                    </th>
                  )}
                  <th style={{ textAlign: 'right', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', fontWeight: 600 }}>
                    Gesamt
                  </th>
                </tr>
              </thead>
              <tbody>
                {invoice.line_items.map((item: { name?: string; description?: string; quantity: number; unit_price: number; total: number; tax_rate?: number; tax_amount?: number }, idx: number) => (
                  <tr
                    key={idx}
                    style={{ borderBottom: '1px solid var(--color-border)' }}
                  >
                    <td style={{ padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)' }}>
                      {idx + 1}
                    </td>
                    <td style={{ padding: 'var(--spacing-3)' }}>
                      <div style={{ fontWeight: 500 }}>{item.name || item.description}</div>
                      {item.name && item.description && item.name !== item.description && (
                        <div style={{ fontSize: '0.85rem', color: 'var(--color-text-secondary)', marginTop: '2px' }}>
                          {item.description}
                        </div>
                      )}
                    </td>
                    <td style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontVariantNumeric: 'tabular-nums' }}>
                      {item.quantity}
                    </td>
                    <td style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontVariantNumeric: 'tabular-nums' }}>
                      {formatCurrency(item.unit_price)}
                    </td>
                    {!isKleinunternehmer && (
                      <td style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontVariantNumeric: 'tabular-nums', color: 'var(--color-text-secondary)' }}>
                        {item.tax_rate != null ? `${item.tax_rate} %` : '—'}
                      </td>
                    )}
                    <td style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
                      {formatCurrency(item.total)}
                    </td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr style={{ borderTop: '2px solid var(--color-border)' }}>
                  <td colSpan={isKleinunternehmer ? 4 : 5} style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontWeight: 600 }}>
                    Netto:
                  </td>
                  <td style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
                    {formatCurrency(invoice.subtotal)}
                  </td>
                </tr>
                {!isKleinunternehmer && invoice.tax_total != null && invoice.tax_total > 0 && (
                  <tr>
                    <td colSpan={5} style={{ textAlign: 'right', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)' }}>
                      MwSt.:
                    </td>
                    <td style={{ textAlign: 'right', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontVariantNumeric: 'tabular-nums' }}>
                      {formatCurrency(invoice.tax_total)}
                    </td>
                  </tr>
                )}
                <tr>
                  <td colSpan={isKleinunternehmer ? 4 : 5} style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontWeight: 700, fontSize: '1.1rem' }}>
                    Brutto:
                  </td>
                  <td style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontWeight: 700, fontSize: '1.1rem', fontVariantNumeric: 'tabular-nums', color: '#00d4ff' }}>
                    {formatCurrency(invoice.total)}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      )}

      {/* Payment history (for partial payments) */}
      {invoice.payments && invoice.payments.length > 0 && (
        <div className="detail-card" style={{ marginTop: 'var(--spacing-6)' }}>
          <h2 className="detail-card__title">Zahlungshistorie</h2>
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '2px solid var(--color-border)' }}>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>Datum</th>
                  <th style={{ textAlign: 'right', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>Betrag</th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>Zahlungsart</th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>Notiz</th>
                </tr>
              </thead>
              <tbody>
                {invoice.payments.map((payment: { id: string; date: string; amount: number; method?: string; notes?: string }, idx: number) => (
                  <tr key={payment.id || idx} style={{ borderBottom: '1px solid var(--color-border)' }}>
                    <td style={{ padding: 'var(--spacing-3)' }}>{new Date(payment.date).toLocaleDateString('de-DE')}</td>
                    <td style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontWeight: 600, color: '#34d399', fontVariantNumeric: 'tabular-nums' }}>
                      {formatCurrency(payment.amount)}
                    </td>
                    <td style={{ padding: 'var(--spacing-3)' }}>{payment.method || '—'}</td>
                    <td style={{ padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)' }}>{payment.notes || '—'}</td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr style={{ borderTop: '2px solid var(--color-border)' }}>
                  <td style={{ padding: 'var(--spacing-3)', fontWeight: 600 }}>Gesamt bezahlt</td>
                  <td style={{ textAlign: 'right', padding: 'var(--spacing-3)', fontWeight: 700, color: '#34d399', fontVariantNumeric: 'tabular-nums' }}>
                    {formatCurrency(paidAmount)}
                  </td>
                  <td colSpan={2} style={{ padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)' }}>
                    Offen: <span style={{ fontWeight: 600, color: '#fbbf24' }}>{formatCurrency(remainingAmount)}</span>
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      )}

      {/* Dunning / Mahnwesen section */}
      {(isOverdue || invoice.status === 'overdue' || existingDunning.length > 0) && invoice.status !== 'paid' && invoice.status !== 'cancelled' && (
        <div className="detail-card" style={{
          marginTop: 'var(--spacing-6)',
          borderColor: 'rgba(239, 68, 68, 0.2)',
          background: 'rgba(239, 68, 68, 0.04)',
        }}>
          <h2 className="detail-card__title" style={{ color: '#f87171' }}>Mahnwesen</h2>
          <div className="detail-card__content">
            {DUNNING_LEVELS.map((dl) => {
              const entry = existingDunning.find((d: DunningEntry) => d.level === dl.level)
              const isNext = dl.level === nextDunningLevel
              return (
                <div
                  key={dl.level}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: 'var(--spacing-3) var(--spacing-4)',
                    borderRadius: 'var(--radius-md)',
                    border: `1px solid ${entry ? 'rgba(239, 68, 68, 0.3)' : isNext ? 'rgba(245, 158, 11, 0.3)' : 'var(--color-border)'}`,
                    background: entry ? 'rgba(239, 68, 68, 0.06)' : isNext ? 'rgba(245, 158, 11, 0.04)' : 'transparent',
                  }}
                >
                  <div>
                    <div style={{ fontWeight: 600, color: entry ? '#f87171' : isNext ? '#fbbf24' : 'var(--color-text-secondary)' }}>
                      {dl.label}
                    </div>
                    <div style={{ fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>
                      {dl.description}
                    </div>
                    {entry && (
                      <div style={{ fontSize: '0.8rem', color: '#f87171', marginTop: '2px' }}>
                        Gesendet am {new Date(entry.date).toLocaleDateString('de-DE')}
                      </div>
                    )}
                  </div>
                  {entry ? (
                    <StatusBadge status="overdue" label="Gesendet" size="sm" />
                  ) : isNext ? (
                    <button
                      className="btn btn--secondary btn--sm"
                      onClick={() => {
                        setDunningLevel(dl.level)
                        setShowDunningModal(true)
                      }}
                      style={{ borderColor: 'rgba(245, 158, 11, 0.4)', color: '#fbbf24' }}
                    >
                      Mahnung erstellen
                    </button>
                  ) : (
                    <span style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>Ausstehend</span>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Notes */}
      {invoice.notes && (
        <div className="detail-card" style={{ marginTop: 'var(--spacing-6)' }}>
          <h2 className="detail-card__title">Notizen</h2>
          <p style={{ color: 'var(--color-text-secondary)', whiteSpace: 'pre-line', margin: 0 }}>
            {invoice.notes}
          </p>
        </div>
      )}

      {/* Bank details */}
      <div className="detail-card" style={{ marginTop: 'var(--spacing-6)' }}>
        <h2 className="detail-card__title">Bankverbindung</h2>
        <div className="detail-card__content">
          {(invoice.bank_account_holder || invoice.bank_iban) ? (
            <>
              {invoice.bank_account_holder && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">Kontoinhaber</span>
                  <span className="detail-card__row-value">{invoice.bank_account_holder}</span>
                </div>
              )}
              {invoice.bank_name && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">Bank</span>
                  <span className="detail-card__row-value">{invoice.bank_name}</span>
                </div>
              )}
              {invoice.bank_iban && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">IBAN</span>
                  <span className="detail-card__row-value" style={{ fontFamily: 'monospace' }}>{invoice.bank_iban}</span>
                </div>
              )}
              {invoice.bank_bic && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">BIC</span>
                  <span className="detail-card__row-value" style={{ fontFamily: 'monospace' }}>{invoice.bank_bic}</span>
                </div>
              )}
            </>
          ) : (
            <p style={{ color: 'var(--color-text-secondary)', margin: 0 }}>
              Bankverbindung nicht hinterlegt. Bitte in den Einstellungen konfigurieren.
            </p>
          )}
        </div>
      </div>

      {/* Kleinunternehmer notice */}
      {isKleinunternehmer && (
        <div style={{
          marginTop: 'var(--spacing-4)',
          padding: 'var(--spacing-3) var(--spacing-4)',
          background: 'rgba(59, 130, 246, 0.06)',
          border: '1px solid rgba(59, 130, 246, 0.15)',
          borderRadius: 'var(--radius-md)',
          color: 'var(--color-text-secondary)',
          fontSize: '0.85rem',
        }}>
          {invoice.kleinunternehmer_text || 'Gemäß §19 UStG wird keine Umsatzsteuer berechnet. Der ausgewiesene Betrag ist der Endbetrag.'}
        </div>
      )}

      {/* Cancel Modal */}
      <Modal
        isOpen={showCancelModal}
        onClose={() => setShowCancelModal(false)}
        title="Rechnung stornieren"
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => setShowCancelModal(false)}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--danger"
              onClick={() => cancelInvoice()}
              disabled={isCancelling}
            >
              {isCancelling ? 'Wird storniert...' : 'Rechnung stornieren'}
            </button>
          </div>
        }
      >
        <p style={{ color: 'var(--color-text-secondary)' }}>
          Möchten Sie die Rechnung <strong>{invoice.number}</strong> wirklich stornieren?
          Diese Aktion kann nicht rückgängig gemacht werden.
        </p>
      </Modal>

      {/* Dunning Modal */}
      <Modal
        isOpen={showDunningModal}
        onClose={() => setShowDunningModal(false)}
        title={`${dunningLevel}. Mahnung erstellen`}
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => setShowDunningModal(false)}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={() => createDunning(dunningLevel)}
              disabled={isCreatingDunning}
            >
              {isCreatingDunning ? 'Wird erstellt...' : 'Mahnung erstellen'}
            </button>
          </div>
        }
      >
        <div style={{ color: 'var(--color-text-secondary)' }}>
          <p>
            Es wird eine <strong>{dunningLevel}. Mahnung</strong> für Rechnung <strong>{invoice.number}</strong> erstellt.
          </p>
          <p style={{ fontSize: '0.9rem' }}>
            Offener Betrag: <strong style={{ color: '#f87171' }}>{formatCurrency(invoice.total)}</strong>
          </p>
          <p style={{ fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>
            {dunningLevel === 1 && 'Eine freundliche Zahlungserinnerung wird generiert.'}
            {dunningLevel === 2 && 'Eine nachdrückliche Zahlungsaufforderung wird generiert.'}
            {dunningLevel === 3 && 'Eine letzte Mahnung mit Hinweis auf rechtliche Schritte wird generiert.'}
          </p>
          <div style={{
            marginTop: 'var(--spacing-3)',
            padding: 'var(--spacing-3)',
            background: 'rgba(245, 158, 11, 0.08)',
            border: '1px solid rgba(245, 158, 11, 0.2)',
            borderRadius: 'var(--radius-md)',
            fontSize: '0.85rem',
          }}>
            Die PDF-Generierung für Mahnungen wird in einer zukünftigen Version implementiert.
            Die Mahnung wird als gesendet markiert.
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default InvoiceDetailPage
