import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import axios from 'axios'
import './PublicQuoteView.scss'

interface QuoteItem {
  id: string
  description: string
  quantity: number
  unit_price: number
  vat_rate: number
  position: number
}

interface Quote {
  id: string
  quote_number: string
  status: string
  customer_name: string
  subject: string
  intro_text: string
  outro_text: string
  quote_date: string
  valid_until: string
  total_net: number
  total_vat: number
  total_gross: number
  kleinunternehmer: boolean
  customer_response: string
  customer_response_at: string
}

const formatCurrency = (cents: number) =>
  (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' })

const formatDate = (d: string) =>
  d ? new Date(d).toLocaleDateString('de-DE') : '—'

export default function PublicQuoteView() {
  const { token } = useParams<{ token: string }>()
  const [quote, setQuote] = useState<Quote | null>(null)
  const [items, setItems] = useState<QuoteItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [responding, setResponding] = useState(false)
  const [message, setMessage] = useState('')
  const [submitted, setSubmitted] = useState(false)

  useEffect(() => {
    if (!token) return
    axios
      .get(`/api/v1/public/quotes/${token}`)
      .then((res) => {
        setQuote(res.data?.data?.quote || res.data?.quote)
        setItems(res.data?.data?.items || res.data?.items || [])
      })
      .catch(() => setError('Angebot nicht gefunden oder abgelaufen.'))
      .finally(() => setLoading(false))
  }, [token])

  const handleRespond = async (response: 'accepted' | 'declined') => {
    if (!token) return
    setResponding(true)
    try {
      await axios.post(`/api/v1/public/quotes/${token}/respond`, {
        response,
        message,
      })
      setSubmitted(true)
      if (quote) {
        setQuote({ ...quote, customer_response: response, status: response })
      }
    } catch {
      setError('Fehler beim Senden der Antwort.')
    } finally {
      setResponding(false)
    }
  }

  if (loading) {
    return (
      <div className="portal-container">
        <div className="portal-loading">Angebot wird geladen...</div>
      </div>
    )
  }

  if (error || !quote) {
    return (
      <div className="portal-container">
        <div className="portal-error">{error || 'Angebot nicht gefunden.'}</div>
      </div>
    )
  }

  const alreadyResponded = !!quote.customer_response || submitted
  const isExpired = quote.valid_until && new Date(quote.valid_until) < new Date()

  return (
    <div className="portal-container">
      <div className="portal-card">
        <div className="portal-header">
          <h1>Angebot {quote.quote_number}</h1>
          <span className={`portal-status portal-status--${quote.status}`}>
            {quote.status === 'accepted' && 'Angenommen'}
            {quote.status === 'declined' && 'Abgelehnt'}
            {quote.status === 'sent' && 'Offen'}
            {quote.status === 'expired' && 'Abgelaufen'}
            {quote.status === 'draft' && 'Entwurf'}
          </span>
        </div>

        <div className="portal-meta">
          <div><strong>Kunde:</strong> {quote.customer_name}</div>
          <div><strong>Betreff:</strong> {quote.subject}</div>
          <div><strong>Datum:</strong> {formatDate(quote.quote_date)}</div>
          <div><strong>Gültig bis:</strong> {formatDate(quote.valid_until)}</div>
        </div>

        {quote.intro_text && <p className="portal-intro">{quote.intro_text}</p>}

        <table className="portal-table">
          <thead>
            <tr>
              <th>#</th>
              <th>Beschreibung</th>
              <th>Menge</th>
              <th>Einzelpreis</th>
              <th>Gesamt</th>
            </tr>
          </thead>
          <tbody>
            {items
              .sort((a, b) => a.position - b.position)
              .map((item, i) => (
                <tr key={item.id}>
                  <td>{i + 1}</td>
                  <td>{item.description}</td>
                  <td>{item.quantity}</td>
                  <td>{formatCurrency(item.unit_price)}</td>
                  <td>{formatCurrency(item.unit_price * item.quantity)}</td>
                </tr>
              ))}
          </tbody>
        </table>

        <div className="portal-totals">
          <div>Netto: {formatCurrency(quote.total_net)}</div>
          {!quote.kleinunternehmer && (
            <div>MwSt: {formatCurrency(quote.total_vat)}</div>
          )}
          {quote.kleinunternehmer && (
            <div className="portal-note">Gem. §19 UStG wird keine USt. berechnet.</div>
          )}
          <div className="portal-total-gross">
            Gesamt: {formatCurrency(quote.total_gross)}
          </div>
        </div>

        {quote.outro_text && <p className="portal-outro">{quote.outro_text}</p>}

        {alreadyResponded ? (
          <div className="portal-responded">
            {quote.customer_response === 'accepted'
              ? 'Vielen Dank! Sie haben dieses Angebot angenommen.'
              : 'Sie haben dieses Angebot abgelehnt.'}
          </div>
        ) : isExpired ? (
          <div className="portal-expired">
            Dieses Angebot ist leider abgelaufen.
          </div>
        ) : (
          <div className="portal-actions">
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="Optionale Nachricht..."
              rows={3}
            />
            <div className="portal-buttons">
              <button
                className="portal-btn portal-btn--accept"
                onClick={() => handleRespond('accepted')}
                disabled={responding}
              >
                {responding ? 'Wird gesendet...' : 'Angebot annehmen'}
              </button>
              <button
                className="portal-btn portal-btn--decline"
                onClick={() => handleRespond('declined')}
                disabled={responding}
              >
                Ablehnen
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
