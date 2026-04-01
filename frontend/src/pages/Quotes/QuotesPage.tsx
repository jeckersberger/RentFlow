import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { quoteApi } from '../../services/api'
import { Input } from '../../components/Form/Input'
import './Quotes.scss'

type QuoteStatus = 'draft' | 'sent' | 'accepted' | 'confirmed' | 'rejected' | 'expired'

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
  items: Array<{
    id: string
    description: string
    quantity: number
    unit_price: number
    total_price: number
    tax_rate: number
  }>
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

const STATUS_TABS: Array<{ value: QuoteStatus | ''; label: string }> = [
  { value: '', label: 'Alle' },
  { value: 'draft', label: 'Entwurf' },
  { value: 'sent', label: 'Gesendet' },
  { value: 'accepted', label: 'Akzeptiert' },
  { value: 'rejected', label: 'Abgelehnt' },
  { value: 'expired', label: 'Abgelaufen' },
]

// No mock data - use real API only

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(amount)
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

function QuotesPage() {
  const navigate = useNavigate()
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState<QuoteStatus | ''>('')

  const { data: quotesRaw, isLoading, error } = useQuery({
    queryKey: ['quotes'],
    queryFn: () => quoteApi.list(),
    staleTime: 1000 * 60 * 5,
  })

  const quotes: Quote[] = useMemo(() => {
    if (!quotesRaw) return []
    const raw = quotesRaw?.data || quotesRaw?.items || quotesRaw
    if (Array.isArray(raw)) return raw
    return []
  }, [quotesRaw])

  const filteredQuotes = useMemo(() => {
    return quotes.filter((q) => {
      const matchesSearch = !searchQuery ||
        q.quote_number.toLowerCase().includes(searchQuery.toLowerCase()) ||
        q.client_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (q.project_name || '').toLowerCase().includes(searchQuery.toLowerCase())
      const matchesStatus = !selectedStatus || q.status === selectedStatus
      return matchesSearch && matchesStatus
    })
  }, [quotes, searchQuery, selectedStatus])

  const isExpired = (quote: Quote) => {
    return new Date(quote.valid_until) < new Date() && quote.status !== 'accepted' && quote.status !== 'confirmed' && quote.status !== 'rejected'
  }

  return (
    <div className="quotes-page">
      {/* Header */}
      <div className="page-header">
        <div>
          <h1 className="page-title">Angebote</h1>
          <p className="page-subtitle">
            {filteredQuotes.length} Angebot{filteredQuotes.length !== 1 ? 'e' : ''} gesamt
          </p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => navigate('/quotes/new')}
        >
          + Angebot erstellen
        </button>
      </div>

      {/* Status Tabs */}
      <div className="status-tabs">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            className={`btn ${selectedStatus === tab.value ? 'btn--primary' : 'btn--secondary'}`}
            onClick={() => setSelectedStatus(tab.value as QuoteStatus | '')}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Search */}
      <div className="filters-bar">
        <div style={{ flex: 1 }}>
          <Input
            type="text"
            placeholder="Nach Angebotsnummer, Kunde oder Projekt suchen..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {error && (
        <div className="error-message" role="alert">
          Fehler beim Laden der Angebote: {(error as any)?.message || 'Unbekannter Fehler'}
        </div>
      )}

      {/* Quotes Table */}
      <div className="quotes-table-container">
        {isLoading ? (
          <div className="empty-state">
            <h3 className="empty-state-title">Daten werden geladen...</h3>
          </div>
        ) : filteredQuotes.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">&#128221;</div>
            <h3 className="empty-state-title">Keine Angebote gefunden</h3>
            <p className="empty-state-text">
              {searchQuery || selectedStatus
                ? 'Versuchen Sie andere Suchkriterien.'
                : 'Erstellen Sie Ihr erstes Angebot, um loszulegen.'}
            </p>
            {!searchQuery && !selectedStatus && (
              <button
                className="btn btn--primary"
                onClick={() => navigate('/quotes/new')}
                style={{ marginTop: 'var(--spacing-3)' }}
              >
                + Angebot erstellen
              </button>
            )}
          </div>
        ) : (
          <table className="quotes-table">
            <thead>
              <tr>
                <th>Nummer</th>
                <th>Projekt</th>
                <th>Kunde</th>
                <th>Status</th>
                <th>Betrag (netto)</th>
                <th>Erstellt am</th>
                <th>Gültig bis</th>
                <th>Aktionen</th>
              </tr>
            </thead>
            <tbody>
              {filteredQuotes.map((quote) => (
                <tr
                  key={quote.id}
                  onClick={() => navigate(`/quotes/${quote.id}`)}
                >
                  <td style={{ fontWeight: 600 }}>{quote.quote_number}</td>
                  <td>
                    {quote.project_name ? (
                      <span
                        style={{ color: 'var(--color-primary)', cursor: 'pointer' }}
                        onClick={(e) => {
                          e.stopPropagation()
                          if (quote.project_id) navigate(`/projects/${quote.project_id}`)
                        }}
                      >
                        {quote.project_name}
                      </span>
                    ) : '\u2014'}
                  </td>
                  <td>{quote.client_name}</td>
                  <td>
                    <span className={`status-badge ${`status-badge--${isExpired(quote) ? 'expired' : quote.status}`}`}>
                      {isExpired(quote) ? STATUS_LABELS.expired : STATUS_LABELS[quote.status]}
                    </span>
                  </td>
                  <td>{formatCurrency(quote.sub_total)}</td>
                  <td>{formatDate(quote.created_at)}</td>
                  <td style={{ color: isExpired(quote) ? 'var(--color-danger)' : 'inherit' }}>
                    {formatDate(quote.valid_until)}
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: 'var(--spacing-1)', justifyContent: 'flex-end' }}>
                      <button
                        className="btn btn--secondary btn--sm"
                        onClick={(e) => { e.stopPropagation(); navigate(`/quotes/${quote.id}`) }}
                        title="Anzeigen"
                      >
                        Anzeigen
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

export default QuotesPage
