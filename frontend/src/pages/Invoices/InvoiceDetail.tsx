import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { invoiceApi, configApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Modal } from '../../components/Modal/Modal'
import { InvoiceStatus, DunningEntry } from '../../types/invoice'
import '../Equipment/Equipment.scss'

const formatCurrency = (value: number | null | undefined) =>
  new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(value ?? 0)

const DUNNING_LEVELS = [
  { level: 1, label: 'Zahlungserinnerung', description: 'Freundliche Erinnerung, keine Gebühr, 7 Tage Frist', fee: 0, days: 7 },
  { level: 2, label: '1. Mahnung', description: 'Formelle Mahnung, 5 € Gebühr, 14 Tage Frist', fee: 5, days: 14 },
  { level: 3, label: '2. Mahnung', description: 'Letzte Mahnung, 10 € Gebühr, 7 Tage Frist, Hinweis auf rechtliche Schritte', fee: 10, days: 7 },
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
  const isKleinunternehmer = invoice?.is_kleinunternehmer || kuConfig?.enabled || kuConfig?.kleinunternehmer || kuConfig?.value || false

  // Fetch dunning history from API
  const { data: dunningHistory = [] } = useQuery({
    queryKey: ['dunning-history', id],
    queryFn: () => invoiceApi.getDunningHistory(id!),
    enabled: !!id,
  })

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
    mutationFn: (level: number) => {
      const levelConfig = DUNNING_LEVELS.find(dl => dl.level === level)
      return invoiceApi.createDunning(id!, level, levelConfig?.fee)
    },
    onSuccess: () => {
      setShowDunningModal(false)
      queryClient.invalidateQueries({ queryKey: ['invoice', id] })
      queryClient.invalidateQueries({ queryKey: ['dunning-history', id] })
      const levelConfig = DUNNING_LEVELS.find(dl => dl.level === dunningLevel)
      addNotification(`${levelConfig?.label || 'Mahnung'} erstellt`, 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Erstellen der Mahnung', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  // PDF download handler — client-side print-to-PDF
  const handleDownloadPDF = () => {
    if (!invoice) return

    const formatDate = (d: string) => new Date(d).toLocaleDateString('de-DE')
    const fmtCur = (v: number) =>
      new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(v)

    const lineItemsRows = (invoice.line_items || [])
      .map((item: { name?: string; description?: string; quantity: number; unit_price: number; total: number; tax_rate?: number; tax_amount?: number }, idx: number) => {
        const taxCol = !isKleinunternehmer
          ? `<td style="text-align:right;padding:8px 12px;border-bottom:1px solid #e5e7eb;">${item.tax_rate != null ? `${item.tax_rate} %` : '\u2014'}</td>`
          : ''
        return `
          <tr>
            <td style="padding:8px 12px;border-bottom:1px solid #e5e7eb;color:#6b7280;">${idx + 1}</td>
            <td style="padding:8px 12px;border-bottom:1px solid #e5e7eb;">
              <strong>${item.name || item.description || ''}</strong>
              ${item.name && item.description && item.name !== item.description ? `<br><span style="font-size:0.85em;color:#6b7280;">${item.description}</span>` : ''}
            </td>
            <td style="text-align:right;padding:8px 12px;border-bottom:1px solid #e5e7eb;">${item.quantity}</td>
            <td style="text-align:right;padding:8px 12px;border-bottom:1px solid #e5e7eb;">${fmtCur(item.unit_price)}</td>
            ${taxCol}
            <td style="text-align:right;padding:8px 12px;border-bottom:1px solid #e5e7eb;font-weight:600;">${fmtCur(item.total)}</td>
          </tr>`
      })
      .join('')

    const taxHeader = !isKleinunternehmer
      ? '<th style="text-align:right;padding:8px 12px;border-bottom:2px solid #d1d5db;font-size:0.85em;color:#6b7280;">MwSt.</th>'
      : ''
    const colCount = isKleinunternehmer ? 4 : 5

    const taxFooterRow =
      !isKleinunternehmer && invoice.tax_total != null && invoice.tax_total > 0
        ? `<tr>
            <td colspan="${colCount}" style="text-align:right;padding:6px 12px;color:#6b7280;">MwSt.:</td>
            <td style="text-align:right;padding:6px 12px;color:#6b7280;">${fmtCur(invoice.tax_total)}</td>
          </tr>`
        : ''

    const kleinunternehmerNote = isKleinunternehmer
      ? `<p style="font-size:0.85em;color:#6b7280;margin-top:12px;">
          ${invoice.kleinunternehmer_text || 'Gem\u00e4\u00df \u00a719 UStG wird keine Umsatzsteuer berechnet.'}
        </p>`
      : ''

    const bankSection =
      invoice.bank_account_holder || invoice.bank_iban
        ? `<div style="margin-top:32px;padding:16px;background:#f9fafb;border:1px solid #e5e7eb;border-radius:6px;">
            <h3 style="margin:0 0 8px;font-size:0.95em;color:#374151;">Bankverbindung</h3>
            ${invoice.bank_account_holder ? `<p style="margin:2px 0;font-size:0.9em;">Kontoinhaber: ${invoice.bank_account_holder}</p>` : ''}
            ${invoice.bank_name ? `<p style="margin:2px 0;font-size:0.9em;">Bank: ${invoice.bank_name}</p>` : ''}
            ${invoice.bank_iban ? `<p style="margin:2px 0;font-size:0.9em;font-family:monospace;">IBAN: ${invoice.bank_iban}</p>` : ''}
            ${invoice.bank_bic ? `<p style="margin:2px 0;font-size:0.9em;font-family:monospace;">BIC: ${invoice.bank_bic}</p>` : ''}
          </div>`
        : ''

    const notesSection = invoice.notes
      ? `<div style="margin-top:24px;">
          <h3 style="font-size:0.95em;color:#374151;margin:0 0 6px;">Notizen</h3>
          <p style="font-size:0.9em;color:#6b7280;white-space:pre-line;margin:0;">${invoice.notes}</p>
        </div>`
      : ''

    const html = `<!DOCTYPE html>
<html lang="de">
<head>
  <meta charset="UTF-8">
  <title>Rechnung ${invoice.number}</title>
  <style>
    @page { size: A4; margin: 20mm 20mm 25mm 20mm; }
    * { box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
           color: #1f2937; margin: 0; padding: 0; font-size: 14px; line-height: 1.5; background: #fff; }
    table { width: 100%; border-collapse: collapse; }
    @media print {
      body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
    }
  </style>
</head>
<body>
  <div style="max-width:210mm;margin:0 auto;padding:20mm;">
    <!-- Header -->
    <div style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:40px;">
      <div>
        <h1 style="margin:0;font-size:1.8em;color:#111827;">RECHNUNG</h1>
        <p style="margin:4px 0 0;font-size:1.1em;color:#6b7280;">${invoice.number}</p>
      </div>
      <div style="text-align:right;font-size:0.9em;color:#6b7280;">
        <div style="width:60px;height:60px;background:#f3f4f6;border:1px solid #e5e7eb;border-radius:8px;display:flex;align-items:center;justify-content:center;margin-left:auto;margin-bottom:8px;font-size:0.75em;color:#9ca3af;">Logo</div>
      </div>
    </div>

    <!-- Invoice meta + Client -->
    <div style="display:flex;justify-content:space-between;margin-bottom:32px;gap:40px;">
      <div style="flex:1;">
        <h3 style="margin:0 0 8px;font-size:0.85em;color:#9ca3af;text-transform:uppercase;letter-spacing:0.05em;">Rechnungsempf\u00e4nger</h3>
        <p style="margin:0;font-weight:600;font-size:1.05em;">${invoice.client_name || '\u2014'}</p>
        ${invoice.client_address ? `<p style="margin:4px 0 0;white-space:pre-line;color:#4b5563;font-size:0.9em;">${invoice.client_address}</p>` : ''}
        ${invoice.client_email ? `<p style="margin:4px 0 0;color:#4b5563;font-size:0.9em;">${invoice.client_email}</p>` : ''}
      </div>
      <div style="text-align:right;">
        <table style="width:auto;margin-left:auto;font-size:0.9em;">
          <tr><td style="padding:3px 16px 3px 0;color:#6b7280;">Rechnungsdatum:</td><td style="padding:3px 0;font-weight:500;">${formatDate(invoice.issue_date)}</td></tr>
          <tr><td style="padding:3px 16px 3px 0;color:#6b7280;">F\u00e4llig am:</td><td style="padding:3px 0;font-weight:500;">${formatDate(invoice.due_date)}</td></tr>
          ${invoice.payment_terms ? `<tr><td style="padding:3px 16px 3px 0;color:#6b7280;">Zahlungsziel:</td><td style="padding:3px 0;">${invoice.payment_terms} Tage</td></tr>` : ''}
          ${invoice.project_name ? `<tr><td style="padding:3px 16px 3px 0;color:#6b7280;">Projekt:</td><td style="padding:3px 0;">${invoice.project_name}</td></tr>` : ''}
        </table>
      </div>
    </div>

    <!-- Line items -->
    <table style="margin-bottom:16px;">
      <thead>
        <tr style="background:#f9fafb;">
          <th style="text-align:left;padding:8px 12px;border-bottom:2px solid #d1d5db;font-size:0.85em;color:#6b7280;width:40px;">Pos.</th>
          <th style="text-align:left;padding:8px 12px;border-bottom:2px solid #d1d5db;font-size:0.85em;color:#6b7280;">Bezeichnung</th>
          <th style="text-align:right;padding:8px 12px;border-bottom:2px solid #d1d5db;font-size:0.85em;color:#6b7280;">Menge</th>
          <th style="text-align:right;padding:8px 12px;border-bottom:2px solid #d1d5db;font-size:0.85em;color:#6b7280;">Einzelpreis</th>
          ${taxHeader}
          <th style="text-align:right;padding:8px 12px;border-bottom:2px solid #d1d5db;font-size:0.85em;color:#6b7280;">Gesamt</th>
        </tr>
      </thead>
      <tbody>
        ${lineItemsRows}
      </tbody>
    </table>

    <!-- Totals -->
    <div style="display:flex;justify-content:flex-end;">
      <table style="width:280px;">
        <tr>
          <td style="padding:6px 12px;font-weight:500;">Netto:</td>
          <td style="text-align:right;padding:6px 12px;">${fmtCur(invoice.subtotal)}</td>
        </tr>
        ${taxFooterRow}
        <tr style="border-top:2px solid #1f2937;">
          <td style="padding:10px 12px;font-weight:700;font-size:1.1em;">Brutto:</td>
          <td style="text-align:right;padding:10px 12px;font-weight:700;font-size:1.1em;">${fmtCur(invoice.total)}</td>
        </tr>
      </table>
    </div>

    ${kleinunternehmerNote}
    ${bankSection}
    ${notesSection}

    <!-- Footer -->
    <div style="margin-top:48px;padding-top:16px;border-top:1px solid #e5e7eb;font-size:0.8em;color:#9ca3af;text-align:center;">
      Rechnung erstellt am ${formatDate(invoice.issue_date)} &mdash; ${invoice.number}
    </div>
  </div>
</body>
</html>`

    const printWindow = window.open('', '_blank')
    if (printWindow) {
      printWindow.document.write(html)
      printWindow.document.close()
      // Wait for content to render, then trigger print
      printWindow.onload = () => {
        printWindow.focus()
        printWindow.print()
      }
      // Fallback if onload doesn't fire (some browsers)
      setTimeout(() => {
        printWindow.focus()
        printWindow.print()
      }, 500)
    } else {
      addNotification('Popup-Blocker verhindert das Öffnen des Druckfensters. Bitte erlauben Sie Popups.', 'warning', { title: 'Hinweis', duration: 5000 })
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

  // Merge dunning entries from invoice payload and API history
  const existingDunning: DunningEntry[] = (dunningHistory && dunningHistory.length > 0)
    ? dunningHistory
    : (invoice.dunning_entries || [])
  const nextDunningLevel = existingDunning.length > 0
    ? Math.min(Math.max(...existingDunning.map((d: DunningEntry) => d.level)) + 1, 3)
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
            Mahnung senden
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
          <div className="detail-card__content" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
            {DUNNING_LEVELS.map((dl) => {
              const entry = existingDunning.find((d: DunningEntry) => d.level === dl.level)
              const isNext = dl.level === nextDunningLevel && !existingDunning.find((d: DunningEntry) => d.level === dl.level)
              const entryDate = entry ? (entry.sent_at || entry.created_at || entry.date) : null
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
                  <div style={{ flex: 1 }}>
                    <div style={{ fontWeight: 600, color: entry ? '#f87171' : isNext ? '#fbbf24' : 'var(--color-text-secondary)' }}>
                      {dl.label}
                    </div>
                    <div style={{ fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>
                      {dl.description}
                    </div>
                    {entry && (
                      <div style={{ display: 'flex', gap: 'var(--spacing-4)', marginTop: '4px', fontSize: '0.8rem', flexWrap: 'wrap' }}>
                        <span style={{ color: '#f87171' }}>
                          Erstellt: {entryDate ? new Date(entryDate).toLocaleDateString('de-DE') : '--'}
                        </span>
                        {entry.due_date && (
                          <span style={{ color: 'var(--color-text-secondary)' }}>
                            Frist bis: {new Date(entry.due_date).toLocaleDateString('de-DE')}
                          </span>
                        )}
                        {(entry.fee != null && entry.fee > 0) && (
                          <span style={{ color: '#fbbf24' }}>
                            Gebühr: {formatCurrency(entry.fee)}
                          </span>
                        )}
                        {entry.fee === 0 && (
                          <span style={{ color: 'var(--color-text-secondary)' }}>
                            Keine Gebühr
                          </span>
                        )}
                      </div>
                    )}
                  </div>
                  <div style={{ marginLeft: 'var(--spacing-3)', flexShrink: 0 }}>
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
                        Mahnung senden
                      </button>
                    ) : (
                      <span style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>Ausstehend</span>
                    )}
                  </div>
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
        title={(() => {
          const cfg = DUNNING_LEVELS.find(dl => dl.level === dunningLevel)
          return cfg ? `${cfg.label} erstellen` : `Mahnung erstellen`
        })()}
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
              disabled={isCreatingDunning || dunningLevel > 3}
            >
              {isCreatingDunning ? 'Wird erstellt...' : 'Mahnung senden'}
            </button>
          </div>
        }
      >
        {(() => {
          const levelConfig = DUNNING_LEVELS.find(dl => dl.level === dunningLevel)
          return (
            <div style={{ color: 'var(--color-text-secondary)' }}>
              {/* Level selector dropdown */}
              <div style={{ marginBottom: 'var(--spacing-4)' }}>
                <label style={{ display: 'block', fontSize: '0.85rem', marginBottom: 'var(--spacing-2)', color: 'var(--color-text-secondary)' }}>
                  Mahnstufe
                </label>
                <select
                  value={dunningLevel}
                  onChange={e => setDunningLevel(Number(e.target.value))}
                  style={{
                    width: '100%',
                    padding: 'var(--spacing-2) var(--spacing-3)',
                    borderRadius: 'var(--radius-md)',
                    border: '1px solid var(--color-border)',
                    background: 'var(--glass-bg)',
                    color: 'var(--color-text)',
                    fontSize: '0.9rem',
                  }}
                >
                  {DUNNING_LEVELS.map(dl => {
                    const alreadySent = existingDunning.find((d: DunningEntry) => d.level === dl.level)
                    return (
                      <option key={dl.level} value={dl.level} disabled={!!alreadySent}>
                        {dl.label}{alreadySent ? ' (bereits gesendet)' : ''}
                      </option>
                    )
                  })}
                </select>
              </div>

              <p>
                Rechnung <strong>{invoice.number}</strong> — Offener Betrag: <strong style={{ color: '#f87171' }}>{formatCurrency(invoice.total)}</strong>
              </p>

              {levelConfig && (
                <div style={{
                  padding: 'var(--spacing-3)',
                  background: 'rgba(245, 158, 11, 0.06)',
                  border: '1px solid rgba(245, 158, 11, 0.15)',
                  borderRadius: 'var(--radius-md)',
                  fontSize: '0.85rem',
                  marginTop: 'var(--spacing-3)',
                }}>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                    <span>{levelConfig.description}</span>
                    <span>Gebühr: <strong>{levelConfig.fee > 0 ? formatCurrency(levelConfig.fee) : 'Keine'}</strong></span>
                    <span>Zahlungsfrist: <strong>{levelConfig.days} Tage</strong></span>
                    {levelConfig.level === 3 && (
                      <span style={{ color: '#f87171', fontWeight: 500, marginTop: '4px' }}>
                        Hinweis: Es wird auf rechtliche Schritte hingewiesen.
                      </span>
                    )}
                  </div>
                </div>
              )}
            </div>
          )
        })()}
      </Modal>
    </div>
  )
}

export default InvoiceDetailPage
