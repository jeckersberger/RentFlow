import { useState, useRef, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Upload, Search, CheckCircle, AlertTriangle, XCircle } from 'lucide-react'
import { bankingApi, invoiceApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import './BankingPage.scss'

interface BankTransaction {
  amount: number
  reference: string
  date: string
  payer: string
}

interface MatchedInvoice {
  id: string
  invoice_number: string
  client_name: string
  total: number
  remaining_amount: number
  status: string
}

interface MatchResult {
  transaction: BankTransaction
  confidence: 'exact' | 'probable' | 'no_match'
  invoice?: MatchedInvoice
  match_reason?: string
}

interface TransactionRow {
  transaction: BankTransaction
  matchResult?: MatchResult
  loading: boolean
  confirmed: boolean
  manualInvoiceId: string
}

const formatCurrency = (value: number) =>
  new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(value)

const confidenceLabel: Record<string, string> = {
  exact: 'Exakt',
  probable: 'Wahrscheinlich',
  no_match: 'Kein Treffer',
}

const confidenceIcon: Record<string, React.ReactNode> = {
  exact: <CheckCircle size={12} />,
  probable: <AlertTriangle size={12} />,
  no_match: <XCircle size={12} />,
}

function BankingPage() {
  const queryClient = useQueryClient()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const addNotification = useNotificationStore(s => s.addNotification)
  const [rows, setRows] = useState<TransactionRow[]>([])

  // Load open invoices for manual selection dropdown
  const { data: openInvoices } = useQuery({
    queryKey: ['open-invoices'],
    queryFn: () => invoiceApi.getOpen(),
  })
  const invoiceOptions: MatchedInvoice[] = openInvoices?.invoices || []

  // CSV import mutation
  const importMutation = useMutation({
    mutationFn: (file: File) => bankingApi.importCSV(file),
    onSuccess: (data) => {
      const newRows: TransactionRow[] = (data.transactions || []).map((tx: BankTransaction) => ({
        transaction: tx,
        matchResult: undefined,
        loading: false,
        confirmed: false,
        manualInvoiceId: '',
      }))
      setRows(prev => [...prev, ...newRows])
      addNotification(`${data.imported} Transaktionen importiert`, 'success')
    },
    onError: () => {
      addNotification('CSV-Import fehlgeschlagen', 'error')
    },
  })

  // Match mutation
  const matchMutation = useMutation({
    mutationFn: (tx: BankTransaction) => bankingApi.matchPayment(tx),
  })

  // Mark paid mutation
  const markPaidMutation = useMutation({
    mutationFn: ({ invoiceId, ref }: { invoiceId: string; ref: string }) =>
      bankingApi.markInvoicePaid(invoiceId, 'bank_transfer', ref),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['open-invoices'] })
      queryClient.invalidateQueries({ queryKey: ['invoices'] })
    },
  })

  const handleFileUpload = useCallback(() => {
    fileInputRef.current?.click()
  }, [])

  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      importMutation.mutate(file)
      e.target.value = ''
    }
  }, [importMutation])

  const handleMatch = useCallback(async (index: number) => {
    setRows(prev => prev.map((r, i) => i === index ? { ...r, loading: true } : r))

    try {
      const row = rows[index]
      const result = await matchMutation.mutateAsync(row.transaction)
      setRows(prev => prev.map((r, i) =>
        i === index ? { ...r, matchResult: result, loading: false } : r
      ))
    } catch {
      setRows(prev => prev.map((r, i) =>
        i === index ? { ...r, loading: false } : r
      ))
      addNotification('Zuordnung fehlgeschlagen', 'error')
    }
  }, [rows, matchMutation, addNotification])

  const handleConfirm = useCallback(async (index: number, invoiceId: string) => {
    const row = rows[index]
    const ref = `${row.transaction.date} ${row.transaction.payer} ${row.transaction.reference}`.trim()

    try {
      await markPaidMutation.mutateAsync({ invoiceId, ref })
      setRows(prev => prev.map((r, i) =>
        i === index ? { ...r, confirmed: true } : r
      ))
      addNotification('Zahlung bestätigt und Rechnung als bezahlt markiert', 'success')
    } catch {
      addNotification('Fehler beim Markieren als bezahlt', 'error')
    }
  }, [rows, markPaidMutation, addNotification])

  const handleManualSelect = useCallback((index: number, invoiceId: string) => {
    setRows(prev => prev.map((r, i) =>
      i === index ? { ...r, manualInvoiceId: invoiceId } : r
    ))
  }, [])

  const handleManualConfirm = useCallback((index: number) => {
    const row = rows[index]
    if (row.manualInvoiceId) {
      handleConfirm(index, row.manualInvoiceId)
    }
  }, [rows, handleConfirm])

  const unmatchedCount = rows.filter(r => !r.confirmed).length

  return (
    <div className="banking-page">
      <h1>Zahlungsabgleich</h1>

      <div className="banking-actions">
        <input
          ref={fileInputRef}
          type="file"
          accept=".csv,.txt"
          onChange={handleFileChange}
          style={{ display: 'none' }}
        />
        <button
          className="upload-btn"
          onClick={handleFileUpload}
          disabled={importMutation.isPending}
        >
          <Upload size={16} />
          {importMutation.isPending ? 'Importiere...' : 'CSV importieren'}
        </button>

        {rows.length > 0 && (
          <div className="transaction-count">
            {unmatchedCount} offene Transaktionen / {rows.length} gesamt
          </div>
        )}
      </div>

      {rows.length === 0 ? (
        <div className="empty-state">
          <Upload size={48} strokeWidth={1} />
          <p>Keine Transaktionen vorhanden</p>
          <p>Importiere einen Kontoauszug als CSV-Datei (Datum;Betrag;Verwendungszweck;Auftraggeber)</p>
        </div>
      ) : (
        <table className="transactions-table">
          <thead>
            <tr>
              <th>Datum</th>
              <th>Betrag</th>
              <th>Verwendungszweck</th>
              <th>Auftraggeber</th>
              <th>Zuordnung</th>
              <th>Aktion</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row, index) => (
              <tr key={index} style={row.confirmed ? { opacity: 0.5 } : undefined}>
                <td>{row.transaction.date}</td>
                <td className="amount-cell">{formatCurrency(row.transaction.amount)}</td>
                <td className="reference-cell" title={row.transaction.reference}>
                  {row.transaction.reference}
                </td>
                <td>{row.transaction.payer}</td>
                <td>
                  {row.confirmed ? (
                    <span className="status-paid">Bezahlt</span>
                  ) : row.matchResult ? (
                    <div className="match-result">
                      <span className={`confidence-badge ${row.matchResult.confidence}`}>
                        {confidenceIcon[row.matchResult.confidence]}
                        {' '}{confidenceLabel[row.matchResult.confidence]}
                      </span>
                      {row.matchResult.invoice && (
                        <span className="match-invoice">
                          {row.matchResult.invoice.invoice_number} - {row.matchResult.invoice.client_name}
                        </span>
                      )}
                      {row.matchResult.match_reason && (
                        <span className="match-reason">{row.matchResult.match_reason}</span>
                      )}
                      {row.matchResult.confidence === 'no_match' && (
                        <div className="manual-select">
                          <select
                            value={row.manualInvoiceId}
                            onChange={(e) => handleManualSelect(index, e.target.value)}
                          >
                            <option value="">Rechnung waehlen...</option>
                            {invoiceOptions.map(inv => (
                              <option key={inv.id} value={inv.id}>
                                {inv.invoice_number} - {inv.client_name} ({formatCurrency(inv.remaining_amount || inv.total)})
                              </option>
                            ))}
                          </select>
                        </div>
                      )}
                    </div>
                  ) : null}
                </td>
                <td>
                  {row.confirmed ? (
                    <CheckCircle size={18} color="#16a34a" />
                  ) : !row.matchResult ? (
                    <button
                      className="match-btn"
                      onClick={() => handleMatch(index)}
                      disabled={row.loading}
                    >
                      <Search size={14} />
                      {row.loading ? 'Suche...' : 'Zuordnen'}
                    </button>
                  ) : row.matchResult.invoice ? (
                    <button
                      className="confirm-btn"
                      onClick={() => handleConfirm(index, row.matchResult!.invoice!.id)}
                      disabled={markPaidMutation.isPending}
                    >
                      <CheckCircle size={14} />
                      Bestaetigen
                    </button>
                  ) : row.manualInvoiceId ? (
                    <button
                      className="confirm-btn"
                      onClick={() => handleManualConfirm(index)}
                      disabled={markPaidMutation.isPending}
                    >
                      <CheckCircle size={14} />
                      Bestaetigen
                    </button>
                  ) : null}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}

export default BankingPage
