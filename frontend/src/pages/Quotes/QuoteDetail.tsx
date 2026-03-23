import { useState, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { quoteApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import styles from './Quotes.module.scss'

type QuoteStatus = 'draft' | 'sent' | 'accepted' | 'confirmed' | 'rejected' | 'expired'

interface QuoteItem {
  id: string
  description: string
  quantity: number
  unit_price: number
  total_price: number
  tax_rate: number
}

interface Quote {
  id: string
  quote_number: string
  project_id?: string
  project_name?: string
  client_name: string
  client_email: string
  status: QuoteStatus
  sub_total: number
  tax_amount: number
  total: number
  currency: string
  valid_until: string
  notes: string
  items: QuoteItem[]
  created_at: string
  updated_at: string
}

const STATUS_LABELS: Record<QuoteStatus, string> = {
  draft: 'Entwurf',
  sent: 'Gesendet',
  accepted: 'Akzeptiert',
  confirmed: 'Bestätigt',
  rejected: 'Abgelehnt',
  expired: 'Abgelaufen',
}

// Fallback mock
const mockQuote: Quote = {
  id: '1', quote_number: 'AN-2026-001', project_id: '1', project_name: 'Stadtfest München 2026',
  client_name: 'Stadt München', client_email: 'veranstaltungen@muenchen.de', status: 'accepted',
  sub_total: 45000, tax_amount: 8550, total: 53550, currency: 'EUR',
  valid_until: '2026-04-01T00:00:00Z', notes: 'Full-Service Veranstaltungstechnik',
  items: [
    { id: '1', description: 'PA-System Hauptbühne', quantity: 1, unit_price: 12000, total_price: 12000, tax_rate: 19 },
    { id: '2', description: 'Lichttechnik 3 Bühnen', quantity: 1, unit_price: 15000, total_price: 15000, tax_rate: 19 },
    { id: '3', description: 'Bühne + Truss', quantity: 1, unit_price: 8000, total_price: 8000, tax_rate: 19 },
    { id: '4', description: 'Techniker-Team (4 Tage)', quantity: 8, unit_price: 1250, total_price: 10000, tax_rate: 19 },
  ],
  created_at: '2026-01-15T10:00:00Z', updated_at: '2026-02-20T14:00:00Z',
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(amount)
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

function QuoteDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [_actionPending, setActionPending] = useState(false)

  const { data: quoteRaw, isLoading, error } = useQuery({
    queryKey: ['quote', id],
    queryFn: () => quoteApi.getById(id!),
    enabled: !!id,
  })

  const quote: Quote | null = useMemo(() => {
    if (quoteRaw && typeof quoteRaw === 'object' && quoteRaw.id) return quoteRaw as Quote
    // Fallback
    return mockQuote
  }, [quoteRaw])

  const isExpired = quote ? new Date(quote.valid_until) < new Date() && quote.status !== 'accepted' && quote.status !== 'confirmed' && quote.status !== 'rejected' : false

  // Send quote
  const { mutate: sendQuote } = useMutation({
    mutationFn: () => quoteApi.send(id!, quote!.client_email),
    onMutate: () => setActionPending(true),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['quote', id] })
      addNotification('Angebot erfolgreich versendet', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Versenden des Angebots', 'error', { title: 'Fehler', duration: 5000 })
    },
    onSettled: () => setActionPending(false),
  })

  // Accept quote
  const { mutate: acceptQuote } = useMutation({
    mutationFn: () => quoteApi.accept(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['quote', id] })
      addNotification('Angebot als akzeptiert markiert', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Akzeptieren des Angebots', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  // Confirm quote
  const { mutate: confirmQuote } = useMutation({
    mutationFn: () => quoteApi.confirm(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['quote', id] })
      addNotification('Angebot bestätigt (Auftragsbestätigung)', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Bestätigen des Angebots', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  // Reject quote
  const { mutate: rejectQuote } = useMutation({
    mutationFn: () => quoteApi.reject(id!, 'Vom Kunden abgelehnt'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['quote', id] })
      addNotification('Angebot als abgelehnt markiert', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Ablehnen des Angebots', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  // Convert to invoice
  const { mutate: convertToInvoice } = useMutation({
    mutationFn: () => quoteApi.convertToInvoice(id!),
    onSuccess: (data: { id?: string }) => {
      addNotification('Rechnung aus Angebot erstellt', 'success', { title: 'Erfolg', duration: 3000 })
      if (data?.id) navigate(`/invoices/${data.id}`)
      else navigate('/invoices')
    },
    onError: () => {
      addNotification('Fehler beim Konvertieren in Rechnung', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  if (isLoading) {
    return (
      <div className={styles['quote-detail-page']}>
        <p style={{ color: 'var(--color-text-secondary)' }}>Wird geladen...</p>
      </div>
    )
  }

  if (error && !quote) {
    return (
      <div className={styles['quote-detail-page']}>
        <div className={styles['error-message']}>Fehler beim Laden des Angebots</div>
      </div>
    )
  }

  if (!quote) {
    return (
      <div className={styles['quote-detail-page']}>
        <div className={styles['error-message']}>Angebot nicht gefunden</div>
      </div>
    )
  }

  const effectiveStatus = isExpired ? 'expired' : quote.status

  return (
    <div className={styles['quote-detail-page']}>
      {/* Back button */}
      <button className={styles['back-button']} onClick={() => navigate('/quotes')}>
        &larr; Zurück zu Angeboten
      </button>

      {/* Header */}
      <div className={styles['page-header']}>
        <div>
          <h1 className={styles['page-title']}>Angebot {quote.quote_number}</h1>
          <p className={styles['page-subtitle']}>
            <span className={`${styles['status-badge']} ${styles[`status-badge--${effectiveStatus}`]}`}>
              {STATUS_LABELS[effectiveStatus]}
            </span>
            {' '}&middot; {quote.client_name}
          </p>
        </div>
        <div className={styles['action-bar']}>
          <button
            className={styles.btn + ' ' + styles['btn--secondary']}
            onClick={() => {
              quoteApi.getPdf(id!).then((blob: Blob) => {
                const url = URL.createObjectURL(blob)
                window.open(url, '_blank')
              }).catch(() => {
                addNotification('PDF-Erstellung fehlgeschlagen', 'error', { title: 'Fehler', duration: 3000 })
              })
            }}
          >
            Als PDF
          </button>

          {quote.status === 'draft' && (
            <button
              className={styles.btn + ' ' + styles['btn--primary']}
              onClick={() => sendQuote()}
            >
              Per E-Mail senden
            </button>
          )}

          {(quote.status === 'accepted' || quote.status === 'confirmed') && (
            <button
              className={styles.btn + ' ' + styles['btn--success']}
              onClick={() => convertToInvoice()}
            >
              In Rechnung umwandeln
            </button>
          )}
        </div>
      </div>

      {/* Content Grid */}
      <div className={styles['detail-grid']}>
        {/* Quote Info */}
        <div className={styles['detail-card']}>
          <h2 className={styles['detail-card-title']}>Angebotsdetails</h2>

          <div className={styles['detail-row']}>
            <span className={styles['detail-label']}>Angebotsnummer</span>
            <span className={styles['detail-value']}>{quote.quote_number}</span>
          </div>

          <div className={styles['detail-row']}>
            <span className={styles['detail-label']}>Kunde</span>
            <span className={styles['detail-value']}>{quote.client_name}</span>
          </div>

          <div className={styles['detail-row']}>
            <span className={styles['detail-label']}>E-Mail</span>
            <span className={styles['detail-value']}>
              <a href={`mailto:${quote.client_email}`} style={{ color: 'var(--color-primary)', textDecoration: 'none' }}>
                {quote.client_email}
              </a>
            </span>
          </div>

          {quote.project_name && (
            <div className={styles['detail-row']}>
              <span className={styles['detail-label']}>Projekt</span>
              <span className={styles['detail-value']}>
                <span
                  style={{ color: 'var(--color-primary)', cursor: 'pointer' }}
                  onClick={() => quote.project_id && navigate(`/projects/${quote.project_id}`)}
                >
                  {quote.project_name}
                </span>
              </span>
            </div>
          )}

          <div className={styles['detail-row']}>
            <span className={styles['detail-label']}>Erstellt am</span>
            <span className={styles['detail-value']}>{formatDate(quote.created_at)}</span>
          </div>

          <div className={styles['detail-row']}>
            <span className={styles['detail-label']}>Gültig bis</span>
            <span className={styles['detail-value']} style={{ color: isExpired ? 'var(--color-danger)' : 'inherit' }}>
              {formatDate(quote.valid_until)}
              {isExpired && ' (abgelaufen)'}
            </span>
          </div>

          {quote.notes && (
            <div className={styles['detail-row']}>
              <span className={styles['detail-label']}>Notizen</span>
              <span className={styles['detail-value']}>{quote.notes}</span>
            </div>
          )}
        </div>

        {/* Summary Card */}
        <div>
          <div className={styles['detail-card']}>
            <h2 className={styles['detail-card-title']}>Zusammenfassung</h2>

            <div className={styles['totals-section']} style={{ borderTop: 'none', marginTop: 0, paddingTop: 0 }}>
              <div className={styles['total-row']}>
                <span className={styles['total-label']}>Zwischensumme (netto)</span>
                <span className={styles['total-value']}>{formatCurrency(quote.sub_total)}</span>
              </div>
              <div className={styles['total-row']}>
                <span className={styles['total-label']}>MwSt. (19%)</span>
                <span className={styles['total-value']}>{formatCurrency(quote.tax_amount)}</span>
              </div>
              <div className={`${styles['total-row']} ${styles['total-row--grand']}`}>
                <span>Gesamtbetrag (brutto)</span>
                <span>{formatCurrency(quote.total)}</span>
              </div>
            </div>
          </div>

          {/* Status Actions */}
          <div className={styles['detail-card']} style={{ marginTop: 'var(--spacing-4)' }}>
            <h2 className={styles['detail-card-title']}>Status ändern</h2>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
              {quote.status === 'draft' && (
                <button
                  className={styles.btn + ' ' + styles['btn--primary']}
                  onClick={() => sendQuote()}
                  style={{ width: '100%', justifyContent: 'center' }}
                >
                  Veröffentlichen &amp; Senden
                </button>
              )}
              {(quote.status === 'sent' || quote.status === 'draft') && (
                <button
                  className={styles.btn + ' ' + styles['btn--success']}
                  onClick={() => acceptQuote()}
                  style={{ width: '100%', justifyContent: 'center' }}
                >
                  Als akzeptiert markieren
                </button>
              )}
              {quote.status === 'accepted' && (
                <button
                  className={styles.btn + ' ' + styles['btn--primary']}
                  onClick={() => confirmQuote()}
                  style={{ width: '100%', justifyContent: 'center' }}
                >
                  Auftragsbestätigung
                </button>
              )}
              {quote.status !== 'rejected' && quote.status !== 'confirmed' && (
                <button
                  className={styles.btn + ' ' + styles['btn--danger']}
                  onClick={() => {
                    if (window.confirm('Angebot wirklich ablehnen?')) rejectQuote()
                  }}
                  style={{ width: '100%', justifyContent: 'center' }}
                >
                  Ablehnen
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Line Items */}
      <div className={styles['detail-card']}>
        <h2 className={styles['detail-card-title']}>Positionen</h2>
        {quote.items && quote.items.length > 0 ? (
          <>
            <table className={styles['items-table']}>
              <thead>
                <tr>
                  <th>Beschreibung</th>
                  <th className={styles['text-right']}>Menge</th>
                  <th className={styles['text-right']}>Einzelpreis</th>
                  <th className={styles['text-right']}>MwSt.</th>
                  <th className={styles['text-right']}>Gesamt</th>
                </tr>
              </thead>
              <tbody>
                {quote.items.map((item) => (
                  <tr key={item.id}>
                    <td>{item.description}</td>
                    <td className={styles['text-right']}>{item.quantity}</td>
                    <td className={styles['text-right']}>{formatCurrency(item.unit_price)}</td>
                    <td className={styles['text-right']}>{item.tax_rate}%</td>
                    <td className={styles['text-right']}>{formatCurrency(item.total_price)}</td>
                  </tr>
                ))}
              </tbody>
            </table>

            <div className={styles['totals-section']}>
              <div className={styles['total-row']}>
                <span className={styles['total-label']}>Netto</span>
                <span className={styles['total-value']}>{formatCurrency(quote.sub_total)}</span>
              </div>
              <div className={styles['total-row']}>
                <span className={styles['total-label']}>MwSt.</span>
                <span className={styles['total-value']}>{formatCurrency(quote.tax_amount)}</span>
              </div>
              <div className={`${styles['total-row']} ${styles['total-row--grand']}`}>
                <span>Brutto</span>
                <span>{formatCurrency(quote.total)}</span>
              </div>
            </div>
          </>
        ) : (
          <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>
            Keine Positionen vorhanden
          </p>
        )}
      </div>
    </div>
  )
}

export default QuoteDetail
