import { useState, useMemo, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { bankingApi, invoiceApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import '../Equipment/Equipment.scss'

interface BankTransaction {
  id: string
  booking_date: string
  value_date: string
  amount: number
  currency: string
  reference: string
  counterparty_name: string
  counterparty_iban: string
  matched_invoice_id?: string
  match_confidence: string
  import_source: string
  created_at: string
}

const formatCurrency = (cents: number) =>
  new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(cents / 100)

const formatDate = (dateStr: string) => {
  if (!dateStr) return '\u2014'
  const d = new Date(dateStr)
  return isNaN(d.getTime()) ? dateStr : d.toLocaleDateString('de-DE')
}

const confidenceLabel = (c: string) => {
  switch (c) {
    case 'high': return 'Hoch'
    case 'medium': return 'Mittel'
    case 'low': return 'Niedrig'
    case 'confirmed': return 'Bestätigt'
    default: return '\u2014'
  }
}

const confidenceColor = (c: string) => {
  switch (c) {
    case 'high':
    case 'confirmed': return 'var(--color-success, #22c55e)'
    case 'medium': return 'var(--color-warning, #f59e0b)'
    case 'low': return 'var(--color-danger, #ef4444)'
    default: return 'var(--color-text-secondary)'
  }
}

function BankingPage() {
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [filter, setFilter] = useState<'all' | 'matched' | 'unmatched'>('all')
  const [confirmingTx, setConfirmingTx] = useState<string | null>(null)
  const [matchInvoiceId, setMatchInvoiceId] = useState('')

  const { data: transactionsRaw, isLoading } = useQuery({
    queryKey: ['banking-transactions', filter],
    queryFn: () => bankingApi.listTransactions(
      filter === 'matched' ? true : filter === 'unmatched' ? false : undefined
    ),
  })

  const transactions: BankTransaction[] = useMemo(() => {
    if (!transactionsRaw) return []
    const raw = Array.isArray(transactionsRaw) ? transactionsRaw : transactionsRaw?.data || []
    return Array.isArray(raw) ? raw : []
  }, [transactionsRaw])

  const { data: invoicesRaw } = useQuery({
    queryKey: ['invoices-for-matching'],
    queryFn: () => invoiceApi.list(1, 200),
  })

  const openInvoices = useMemo(() => {
    const list = invoicesRaw?.data || []
    return Array.isArray(list) ? list.filter((i: any) =>
      ['finalized', 'sent', 'partial_paid', 'overdue', 'partial'].includes(i.status)
    ) : []
  }, [invoicesRaw])

  const { mutate: importCSV, isPending: isImporting } = useMutation({
    mutationFn: (file: File) => bankingApi.importCSV(file),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['banking-transactions'] })
      addNotification(`${data?.imported || 0} Transaktionen importiert`, 'success', { title: 'Import erfolgreich', duration: 4000 })
    },
    onError: () => {
      addNotification('CSV-Import fehlgeschlagen', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const { mutate: runAutoMatch, isPending: isMatching } = useMutation({
    mutationFn: () => bankingApi.autoMatch(),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['banking-transactions'] })
      addNotification(`${data?.matched || 0} Zuordnungen gefunden`, 'success', { title: 'Auto-Match', duration: 4000 })
    },
    onError: () => {
      addNotification('Auto-Match fehlgeschlagen', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const { mutate: confirmMatch } = useMutation({
    mutationFn: ({ txId, invId }: { txId: string; invId: string }) =>
      bankingApi.confirmMatch(txId, invId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['banking-transactions'] })
      setConfirmingTx(null)
      setMatchInvoiceId('')
      addNotification('Zuordnung bestätigt & Zahlung erfasst', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Zuordnung fehlgeschlagen', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const handleFileSelect = () => {
    const file = fileInputRef.current?.files?.[0]
    if (file) {
      importCSV(file)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  const unmatchedCount = transactions.filter(t => !t.matched_invoice_id && t.match_confidence === 'none').length
  const matchedCount = transactions.filter(t => t.matched_invoice_id || t.match_confidence === 'confirmed').length

  return (
    <div className="equipment-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Bank-Import</h1>
          <p className="page-subtitle">
            {transactions.length} Transaktionen &middot; {matchedCount} zugeordnet &middot; {unmatchedCount} offen
          </p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-2)' }}>
          <input
            ref={fileInputRef}
            type="file"
            accept=".csv"
            style={{ display: 'none' }}
            onChange={handleFileSelect}
          />
          <button
            className="btn btn--secondary"
            onClick={() => fileInputRef.current?.click()}
            disabled={isImporting}
          >
            {isImporting ? 'Importiert...' : 'CSV importieren'}
          </button>
          <button
            className="btn btn--primary"
            onClick={() => runAutoMatch()}
            disabled={isMatching || transactions.length === 0}
          >
            {isMatching ? 'Zuordne...' : 'Auto-Match starten'}
          </button>
        </div>
      </div>

      {/* Filter Tabs */}
      <div style={{ display: 'flex', gap: 'var(--spacing-2)', marginBottom: 'var(--spacing-4)' }}>
        {([['all', 'Alle'], ['unmatched', 'Offen'], ['matched', 'Zugeordnet']] as const).map(([key, label]) => (
          <button
            key={key}
            className={`btn ${filter === key ? 'btn--primary' : 'btn--secondary'}`}
            onClick={() => setFilter(key)}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Transactions Table */}
      <div className="table-container">
        {isLoading ? (
          <p style={{ padding: 'var(--spacing-4)', color: 'var(--color-text-secondary)' }}>Laden...</p>
        ) : transactions.length === 0 ? (
          <div style={{ padding: 'var(--spacing-8)', textAlign: 'center' }}>
            <p style={{ fontSize: '2rem', marginBottom: 'var(--spacing-2)' }}>&#127974;</p>
            <h3>Keine Transaktionen</h3>
            <p style={{ color: 'var(--color-text-secondary)' }}>
              Importieren Sie eine CSV-Datei von Ihrer Bank.
            </p>
          </div>
        ) : (
          <table className="data-table">
            <thead>
              <tr>
                <th>Datum</th>
                <th>Auftraggeber</th>
                <th>Verwendungszweck</th>
                <th style={{ textAlign: 'right' }}>Betrag</th>
                <th>Zuordnung</th>
                <th>Aktion</th>
              </tr>
            </thead>
            <tbody>
              {transactions.map(tx => (
                <tr key={tx.id}>
                  <td>{formatDate(tx.booking_date)}</td>
                  <td>
                    <div>{tx.counterparty_name || '\u2014'}</div>
                    {tx.counterparty_iban && (
                      <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)' }}>
                        {tx.counterparty_iban}
                      </div>
                    )}
                  </td>
                  <td style={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {tx.reference || '\u2014'}
                  </td>
                  <td style={{ textAlign: 'right', fontWeight: 600, color: tx.amount >= 0 ? 'var(--color-success, #22c55e)' : 'var(--color-danger, #ef4444)' }}>
                    {formatCurrency(tx.amount)}
                  </td>
                  <td>
                    {tx.matched_invoice_id ? (
                      <span style={{ color: confidenceColor(tx.match_confidence), fontWeight: 500 }}>
                        {confidenceLabel(tx.match_confidence)}
                      </span>
                    ) : (
                      <span style={{ color: 'var(--color-text-secondary)' }}>\u2014</span>
                    )}
                  </td>
                  <td>
                    {!tx.matched_invoice_id || tx.match_confidence !== 'confirmed' ? (
                      confirmingTx === tx.id ? (
                        <div style={{ display: 'flex', gap: 'var(--spacing-1)', alignItems: 'center' }}>
                          <select
                            value={matchInvoiceId}
                            onChange={e => setMatchInvoiceId(e.target.value)}
                            style={{ fontSize: '0.8rem', padding: '2px 4px' }}
                          >
                            <option value="">Rechnung wählen...</option>
                            {openInvoices.map((inv: any) => (
                              <option key={inv.id} value={inv.id}>
                                {inv.number || inv.invoice_number} — {formatCurrency(inv.total || inv.total_gross || 0)}
                              </option>
                            ))}
                          </select>
                          <button
                            className="btn btn--primary btn--sm"
                            disabled={!matchInvoiceId}
                            onClick={() => confirmMatch({ txId: tx.id, invId: matchInvoiceId })}
                            style={{ fontSize: '0.75rem', padding: '2px 8px' }}
                          >
                            OK
                          </button>
                          <button
                            className="btn btn--secondary btn--sm"
                            onClick={() => { setConfirmingTx(null); setMatchInvoiceId('') }}
                            style={{ fontSize: '0.75rem', padding: '2px 8px' }}
                          >
                            X
                          </button>
                        </div>
                      ) : (
                        <button
                          className="btn btn--secondary btn--sm"
                          onClick={() => setConfirmingTx(tx.id)}
                          style={{ fontSize: '0.75rem' }}
                        >
                          Zuordnen
                        </button>
                      )
                    ) : (
                      <span style={{ color: 'var(--color-success)', fontSize: '0.8rem' }}>&#10003;</span>
                    )}
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

export default BankingPage
