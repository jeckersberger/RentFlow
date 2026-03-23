import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { quoteApi } from '../../services/api'
import { Input } from '../../components/Form/Input'
import styles from './Quotes.module.scss'

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

// Mock data
const mockQuotes: Quote[] = [
  {
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
  },
  {
    id: '2', quote_number: 'AN-2026-002', project_id: '2', project_name: 'Firmen-Gala TechCorp',
    client_name: 'TechCorp GmbH', client_email: 'events@techcorp.de', status: 'sent',
    sub_total: 18500, tax_amount: 3515, total: 22015, currency: 'EUR',
    valid_until: '2026-04-20T00:00:00Z', notes: 'Ton/Licht/Video für Gala',
    items: [
      { id: '1', description: 'Audio-Paket Gala', quantity: 1, unit_price: 6500, total_price: 6500, tax_rate: 19 },
      { id: '2', description: 'Licht-Design Gala', quantity: 1, unit_price: 7000, total_price: 7000, tax_rate: 19 },
      { id: '3', description: 'Video/Streaming', quantity: 1, unit_price: 5000, total_price: 5000, tax_rate: 19 },
    ],
    created_at: '2026-02-18T14:00:00Z', updated_at: '2026-03-18T09:00:00Z',
  },
  {
    id: '3', quote_number: 'AN-2026-003', project_id: '3', project_name: 'Open Air Festival Bodensee',
    client_name: 'Festival GmbH', client_email: 'info@festivalgmbh.de', status: 'draft',
    sub_total: 85000, tax_amount: 16150, total: 101150, currency: 'EUR',
    valid_until: '2026-05-15T00:00:00Z', notes: '3-Tages Festival, 2 Stages',
    items: [
      { id: '1', description: 'Main Stage PA + Monitoring', quantity: 1, unit_price: 28000, total_price: 28000, tax_rate: 19 },
      { id: '2', description: 'Second Stage PA', quantity: 1, unit_price: 12000, total_price: 12000, tax_rate: 19 },
      { id: '3', description: 'Lichttechnik beide Stages', quantity: 1, unit_price: 22000, total_price: 22000, tax_rate: 19 },
      { id: '4', description: 'Bühnen + Rigging', quantity: 1, unit_price: 15000, total_price: 15000, tax_rate: 19 },
      { id: '5', description: 'Crew (3 Tage)', quantity: 10, unit_price: 800, total_price: 8000, tax_rate: 19 },
    ],
    created_at: '2026-02-25T11:00:00Z', updated_at: '2026-03-22T15:00:00Z',
  },
  {
    id: '4', quote_number: 'AN-2026-004', project_id: '4', project_name: 'Produktlaunch AutoBrand',
    client_name: 'AutoBrand AG', client_email: 'events@autobrand.de', status: 'confirmed',
    sub_total: 32000, tax_amount: 6080, total: 38080, currency: 'EUR',
    valid_until: '2026-02-15T00:00:00Z', notes: 'Produktpräsentation neues E-Auto',
    items: [
      { id: '1', description: 'LED-Wall 6x3m (3 Tage)', quantity: 1, unit_price: 4500, total_price: 4500, tax_rate: 19 },
      { id: '2', description: 'Line Array System (3 Tage)', quantity: 2, unit_price: 3600, total_price: 7200, tax_rate: 19 },
      { id: '3', description: 'Lichttechnik Paket', quantity: 1, unit_price: 8500, total_price: 8500, tax_rate: 19 },
      { id: '4', description: 'Techniker (3 Tage)', quantity: 4, unit_price: 2950, total_price: 11800, tax_rate: 19 },
    ],
    created_at: '2025-12-10T10:00:00Z', updated_at: '2026-02-10T16:00:00Z',
  },
  {
    id: '5', quote_number: 'AN-2026-005', project_name: 'Sommerfest Weber & Partner',
    client_name: 'Weber & Partner Anwaltskanzlei', client_email: 'sekretariat@weberpartner.de', status: 'rejected',
    sub_total: 6500, tax_amount: 1235, total: 7735, currency: 'EUR',
    valid_until: '2026-03-01T00:00:00Z', notes: 'Rejected: Budget zu hoch',
    items: [
      { id: '1', description: 'DJ-Equipment + PA', quantity: 1, unit_price: 3500, total_price: 3500, tax_rate: 19 },
      { id: '2', description: 'Beleuchtung Garten', quantity: 1, unit_price: 2000, total_price: 2000, tax_rate: 19 },
      { id: '3', description: 'Techniker', quantity: 1, unit_price: 1000, total_price: 1000, tax_rate: 19 },
    ],
    created_at: '2026-01-28T09:00:00Z', updated_at: '2026-02-25T11:00:00Z',
  },
]

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

  const { data: quotesRaw, isLoading: _isLoading, error } = useQuery({
    queryKey: ['quotes'],
    queryFn: () => quoteApi.list(),
    staleTime: 1000 * 60 * 5,
  })

  const quotes: Quote[] = useMemo(() => {
    const raw = quotesRaw?.data || quotesRaw
    if (Array.isArray(raw) && raw.length > 0) return raw
    return mockQuotes
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
    <div className={styles['quotes-page']}>
      {/* Header */}
      <div className={styles['page-header']}>
        <div>
          <h1 className={styles['page-title']}>Angebote</h1>
          <p className={styles['page-subtitle']}>
            {filteredQuotes.length} Angebot{filteredQuotes.length !== 1 ? 'e' : ''} gesamt
          </p>
        </div>
        <button
          className={styles.btn + ' ' + styles['btn--primary']}
          onClick={() => navigate('/invoices/new')}
        >
          + Angebot erstellen
        </button>
      </div>

      {/* Status Tabs */}
      <div className={styles['status-tabs']}>
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            className={`${styles.btn} ${selectedStatus === tab.value ? styles['btn--primary'] : styles['btn--secondary']}`}
            onClick={() => setSelectedStatus(tab.value as QuoteStatus | '')}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Search */}
      <div className={styles['filters-bar']}>
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
        <div className={styles['error-message']} role="alert">
          Fehler beim Laden der Angebote. Verwende Demo-Daten.
        </div>
      )}

      {/* Quotes Table */}
      <div className={styles['quotes-table-container']}>
        {filteredQuotes.length === 0 ? (
          <div className={styles['empty-state']}>
            <div className={styles['empty-state-icon']}>&#128221;</div>
            <h3 className={styles['empty-state-title']}>Keine Angebote gefunden</h3>
            <p className={styles['empty-state-text']}>
              {searchQuery || selectedStatus
                ? 'Versuchen Sie andere Suchkriterien.'
                : 'Erstellen Sie Ihr erstes Angebot, um loszulegen.'}
            </p>
          </div>
        ) : (
          <table className={styles['quotes-table']}>
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
                    <span className={`${styles['status-badge']} ${styles[`status-badge--${isExpired(quote) ? 'expired' : quote.status}`]}`}>
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
                        className={styles.btn + ' ' + styles['btn--secondary'] + ' ' + styles['btn--sm']}
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
