import { useState, useRef, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { expenseApi, projectApi, invoiceApi, aiApi } from '../../services/api'
import { Modal } from '../../components/Modal/Modal'
import { Input } from '../../components/Form/Input'
import { useNotificationStore } from '../../stores/notificationStore'

interface Expense {
  id: string
  date: string
  description: string
  amount: number
  vat_rate: number
  category: string
  project_id?: string
  project_name?: string
  receipt_url?: string
  notes?: string
  created_at: string
}

interface ProjectOption {
  id: string
  name: string
}

const CATEGORIES = [
  { value: 'miete', label: 'Miete' },
  { value: 'versicherung', label: 'Versicherung' },
  { value: 'fahrzeug', label: 'Fahrzeug' },
  { value: 'material', label: 'Material' },
  { value: 'freelancer', label: 'Freelancer' },
  { value: 'buero', label: 'Buero' },
  { value: 'reise', label: 'Reisekosten' },
  { value: 'sonstig', label: 'Sonstiges' },
]

const VAT_RATES = [
  { value: 19, label: '19% MwSt' },
  { value: 7, label: '7% MwSt' },
  { value: 0, label: '0% (steuerfrei)' },
]

const CATEGORY_LABELS: Record<string, string> = Object.fromEntries(
  CATEGORIES.map(c => [c.value, c.label])
)

interface ExpenseForm {
  date: string
  description: string
  amount: string
  vat_rate: number
  category: string
  project_id: string
  notes: string
  // Neue Felder
  type: 'invoice' | 'receipt' | 'entertainment'
  vendor: string
  vendor_address: string
  vendor_vat_id: string
  vendor_iban: string
  invoice_number: string
  booking_account: string
  due_date: string
  discount_percent: string
  discount_days: string
  payment_method: string
  // Bewirtungsbeleg
  entertainment_location: string
  entertainment_reason: string
  entertainment_guests: string
  entertainment_tip: string
}

const emptyForm: ExpenseForm = {
  date: new Date().toISOString().split('T')[0],
  description: '',
  amount: '',
  vat_rate: 19,
  category: 'sonstig',
  project_id: '',
  notes: '',
  type: 'invoice',
  vendor: '',
  vendor_address: '',
  vendor_vat_id: '',
  vendor_iban: '',
  invoice_number: '',
  booking_account: '',
  due_date: '',
  discount_percent: '',
  discount_days: '',
  payment_method: 'bank_transfer',
  entertainment_location: '',
  entertainment_reason: '',
  entertainment_guests: '',
  entertainment_tip: '',
}

function ExpensesPage() {
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState<ExpenseForm>(emptyForm)
  const [activeView, setActiveView] = useState<'list' | 'euer'>('list')
  const [filterCategory, setFilterCategory] = useState('all')
  const [selectedYear, setSelectedYear] = useState(new Date().getFullYear())
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [receiptFile, setReceiptFile] = useState<File | null>(null)
  const [isAnalyzing, setIsAnalyzing] = useState(false)
  const [aiConfidence, setAiConfidence] = useState<number | null>(null)

  // KI-Beleganalyse: Datei an Claude Vision senden und Formular auto-befuellen
  const analyzeReceipt = useCallback(async (file: File) => {
    setIsAnalyzing(true)
    setAiConfidence(null)
    try {
      const result = await aiApi.analyzeReceipt(file)
      if (!result.is_invoice && result.document_type === 'other') {
        addNotification('Das Dokument wurde nicht als Rechnung/Beleg erkannt.', 'error', { title: 'Nicht erkannt' })
        setIsAnalyzing(false)
        return
      }

      // Formular auto-befuellen
      setForm(prev => ({
        ...prev,
        type: result.document_type === 'entertainment' ? 'entertainment' : (result.document_type === 'receipt' ? 'receipt' : 'invoice'),
        vendor: result.vendor || prev.vendor,
        vendor_address: result.vendor_address || prev.vendor_address,
        vendor_vat_id: result.vendor_vat_id || prev.vendor_vat_id,
        vendor_iban: result.vendor_iban || prev.vendor_iban,
        description: result.description || prev.description,
        amount: result.gross_amount ? String(result.gross_amount) : prev.amount,
        vat_rate: result.tax_rate || prev.vat_rate,
        invoice_number: result.invoice_number || prev.invoice_number,
        date: result.invoice_date || prev.date,
        due_date: result.due_date || prev.due_date,
        discount_percent: result.discount_percent ? String(result.discount_percent) : prev.discount_percent,
        discount_days: result.discount_days ? String(result.discount_days) : prev.discount_days,
        payment_method: result.payment_method || prev.payment_method,
        booking_account: result.suggested_skr03 || prev.booking_account,
        category: result.suggested_category || prev.category,
        entertainment_location: result.entertainment_location || prev.entertainment_location,
        entertainment_tip: result.tip ? String(result.tip) : prev.entertainment_tip,
      }))

      setAiConfidence(result.confidence)
      addNotification(
        `Beleg erkannt: ${result.vendor || 'Unbekannt'} - ${result.gross_amount?.toFixed(2) || '?'} EUR (${Math.round((result.confidence || 0) * 100)}% Konfidenz)`,
        'success',
        { title: 'KI-Analyse abgeschlossen' }
      )
    } catch (err: any) {
      addNotification(
        `KI-Analyse fehlgeschlagen: ${err?.message || 'Unbekannter Fehler'}`,
        'error',
        { title: 'Analysefehler' }
      )
    } finally {
      setIsAnalyzing(false)
    }
  }, [addNotification])

  const { data: expensesData, isLoading } = useQuery({
    queryKey: ['expenses'],
    queryFn: () => expenseApi.list({ limit: 500 }),
    staleTime: 1000 * 60 * 2,
  })

  const { data: projectsData } = useQuery({
    queryKey: ['projects-for-expenses'],
    queryFn: () => projectApi.list(1, 200),
    enabled: showModal,
  })

  const { data: invoicesData } = useQuery({
    queryKey: ['invoices-for-euer'],
    queryFn: () => invoiceApi.list(1, 500),
    enabled: activeView === 'euer',
  })

  const projects: ProjectOption[] =
    projectsData?.items || projectsData?.data || (Array.isArray(projectsData) ? projectsData : [])

  const expenses: Expense[] =
    expensesData?.items || expensesData?.data || (Array.isArray(expensesData) ? expensesData : [])

  const invoices: any[] =
    invoicesData?.items || invoicesData?.data || (Array.isArray(invoicesData) ? invoicesData : [])

  const createMutation = useMutation({
    mutationFn: async (data: any) => {
      const result = await expenseApi.create(data)
      // Upload receipt if selected
      if (receiptFile && result?.id) {
        const formData = new FormData()
        formData.append('file', receiptFile)
        try {
          await expenseApi.uploadReceipt(result.id, formData)
        } catch {
          // receipt upload failed but expense was created
        }
      }
      return result
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] })
      setShowModal(false)
      setForm(emptyForm)
      setReceiptFile(null)
      addNotification('Beleg wurde erfolgreich erfasst.', 'success', { title: 'Beleg gespeichert' })
    },
    onError: (err: any) => {
      addNotification(`Beleg konnte nicht gespeichert werden: ${err.message || 'Unbekannter Fehler'}`, 'error', { title: 'Fehler' })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => expenseApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] })
      addNotification('Beleg wurde geloescht.', 'success', { title: 'Geloescht' })
    },
  })

  const handleSubmit = () => {
    if (!form.description.trim() || !form.amount) {
      addNotification('Beschreibung und Betrag sind Pflichtfelder.', 'error', { title: 'Fehler' })
      return
    }
    createMutation.mutate({
      date: form.date,
      description: form.description,
      amount: parseFloat(form.amount),
      vat_rate: form.vat_rate,
      category: form.category,
      project_id: form.project_id || undefined,
      notes: form.notes || undefined,
    })
  }

  const filteredExpenses = expenses.filter(e =>
    filterCategory === 'all' || e.category === filterCategory
  )

  const totalExpenses = filteredExpenses.reduce((sum, e) => sum + (e.amount || 0), 0)

  // EUeR calculation
  const yearExpenses = expenses.filter(e => {
    const d = new Date(e.date || e.created_at)
    return d.getFullYear() === selectedYear
  })

  const yearInvoices = invoices.filter(i => {
    const d = new Date(i.paid_at || i.invoice_date || i.created_at)
    return d.getFullYear() === selectedYear && (i.status === 'paid' || i.status === 'bezahlt')
  })

  const monthlyData: { month: number; label: string; income: number; expenses: number; profit: number }[] = []
  for (let m = 0; m < 12; m++) {
    const monthExpenses = yearExpenses
      .filter(e => new Date(e.date || e.created_at).getMonth() === m)
      .reduce((sum, e) => sum + (e.amount || 0), 0)
    const monthIncome = yearInvoices
      .filter(i => new Date(i.paid_at || i.invoice_date || i.created_at).getMonth() === m)
      .reduce((sum, i) => sum + (i.total_amount || i.amount || i.total || 0), 0)
    monthlyData.push({
      month: m + 1,
      label: new Date(selectedYear, m, 1).toLocaleDateString('de-DE', { month: 'long' }),
      income: monthIncome,
      expenses: monthExpenses,
      profit: monthIncome - monthExpenses,
    })
  }

  const yearTotalIncome = monthlyData.reduce((s, m) => s + m.income, 0)
  const yearTotalExpenses = monthlyData.reduce((s, m) => s + m.expenses, 0)
  const yearProfit = yearTotalIncome - yearTotalExpenses

  const exportCSV = () => {
    const header = 'Monat;Einnahmen;Ausgaben;Gewinn/Verlust\n'
    const rows = monthlyData.map(m =>
      `${m.label};${m.income.toFixed(2).replace('.', ',')};${m.expenses.toFixed(2).replace('.', ',')};${m.profit.toFixed(2).replace('.', ',')}`
    ).join('\n')
    const totalRow = `\nGESAMT;${yearTotalIncome.toFixed(2).replace('.', ',')};${yearTotalExpenses.toFixed(2).replace('.', ',')};${yearProfit.toFixed(2).replace('.', ',')}`
    const csv = header + rows + totalRow
    const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `EUeR_${selectedYear}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  const formatEUR = (n: number) => n.toLocaleString('de-DE', { style: 'currency', currency: 'EUR' })

  return (
    <div style={{ padding: '0' }}>
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1.5rem' }}>
        <div>
          <h1 className="page-title">Belege & EUeR</h1>
          <p className="page-subtitle">
            {expenses.length} Belege erfasst | Gesamtausgaben: {formatEUR(totalExpenses)}
          </p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => setShowModal(true)}
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
        >
          + Beleg erfassen
        </button>
      </div>

      {/* Stats */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '1rem', marginBottom: '1.5rem' }}>
        <div className="stat-card">
          <div className="stat-card__label">Belege gesamt</div>
          <div className="stat-card__value">{expenses.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Ausgaben gesamt</div>
          <div className="stat-card__value" style={{ color: '#ef4444' }}>{formatEUR(totalExpenses)}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Diesen Monat</div>
          <div className="stat-card__value">
            {formatEUR(expenses.filter(e => {
              const d = new Date(e.date || e.created_at)
              const now = new Date()
              return d.getMonth() === now.getMonth() && d.getFullYear() === now.getFullYear()
            }).reduce((s, e) => s + (e.amount || 0), 0))}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Kategorien</div>
          <div className="stat-card__value">{new Set(expenses.map(e => e.category)).size}</div>
        </div>
      </div>

      {/* Tab Navigation */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem', borderBottom: '1px solid rgba(255,255,255,0.08)', paddingBottom: '0.5rem' }}>
        <button
          style={{
            padding: '0.5rem 1rem',
            background: activeView === 'list' ? 'var(--color-primary)' : 'transparent',
            color: activeView === 'list' ? '#fff' : 'var(--color-text-secondary)',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontWeight: 500,
          }}
          onClick={() => setActiveView('list')}
        >
          Belege
        </button>
        <button
          style={{
            padding: '0.5rem 1rem',
            background: activeView === 'euer' ? 'var(--color-primary)' : 'transparent',
            color: activeView === 'euer' ? '#fff' : 'var(--color-text-secondary)',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontWeight: 500,
          }}
          onClick={() => setActiveView('euer')}
        >
          EUeR-Ansicht
        </button>
      </div>

      {/* Belege List View */}
      {activeView === 'list' && (
        <div>
          <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem', alignItems: 'center' }}>
            <label style={{ fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>Kategorie:</label>
            <select
              style={{
                padding: '0.4rem 0.75rem',
                background: 'rgba(255,255,255,0.05)',
                border: '1px solid rgba(255,255,255,0.1)',
                borderRadius: '6px',
                color: 'inherit',
                fontSize: '0.85rem',
              }}
              value={filterCategory}
              onChange={(e) => setFilterCategory(e.target.value)}
            >
              <option value="all">Alle</option>
              {CATEGORIES.map(c => (
                <option key={c.value} value={c.value}>{c.label}</option>
              ))}
            </select>
          </div>

          {isLoading ? (
            <div style={{ textAlign: 'center', padding: '3rem', color: 'var(--color-text-muted)' }}>
              Belege werden geladen...
            </div>
          ) : filteredExpenses.length === 0 ? (
            <div style={{
              textAlign: 'center',
              padding: '3rem 2rem',
              background: 'var(--color-surface)',
              borderRadius: '12px',
              border: '1px solid rgba(255,255,255,0.06)',
            }}>
              <div style={{ fontSize: '2.5rem', marginBottom: '0.75rem' }}>{'\u{1F4B3}'}</div>
              <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '0.5rem' }}>
                {expenses.length === 0 ? 'Noch keine Belege erfasst' : 'Keine Belege in dieser Kategorie'}
              </h3>
              <p style={{ color: 'var(--color-text-muted)', marginBottom: '1rem' }}>
                Erfassen Sie Ihre Ausgaben, um eine automatische EUeR zu erstellen.
              </p>
              {expenses.length === 0 && (
                <button className="btn btn--primary" onClick={() => setShowModal(true)}>
                  Ersten Beleg erfassen
                </button>
              )}
            </div>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{
                width: '100%',
                borderCollapse: 'collapse',
                fontSize: '0.875rem',
              }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.08)' }}>
                    <th style={{ textAlign: 'left', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>Datum</th>
                    <th style={{ textAlign: 'left', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>Beschreibung</th>
                    <th style={{ textAlign: 'left', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>Kategorie</th>
                    <th style={{ textAlign: 'right', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>Netto</th>
                    <th style={{ textAlign: 'right', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>MwSt</th>
                    <th style={{ textAlign: 'right', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>Brutto</th>
                    <th style={{ textAlign: 'center', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>Beleg</th>
                    <th style={{ textAlign: 'right', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>Aktionen</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredExpenses
                    .sort((a, b) => new Date(b.date || b.created_at).getTime() - new Date(a.date || a.created_at).getTime())
                    .map((exp) => {
                      const netto = exp.amount || 0
                      const mwst = netto * (exp.vat_rate / 100)
                      const brutto = netto + mwst
                      return (
                        <tr key={exp.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
                          <td style={{ padding: '0.75rem 0.5rem' }}>
                            {new Date(exp.date || exp.created_at).toLocaleDateString('de-DE')}
                          </td>
                          <td style={{ padding: '0.75rem 0.5rem' }}>
                            <div>{exp.description}</div>
                            {exp.project_name && (
                              <div style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
                                Projekt: {exp.project_name}
                              </div>
                            )}
                          </td>
                          <td style={{ padding: '0.75rem 0.5rem' }}>
                            <span style={{
                              display: 'inline-block',
                              padding: '2px 8px',
                              borderRadius: '9999px',
                              fontSize: '0.75rem',
                              fontWeight: 500,
                              background: 'rgba(0, 212, 255, 0.1)',
                              color: 'var(--color-primary)',
                            }}>
                              {CATEGORY_LABELS[exp.category] || exp.category}
                            </span>
                          </td>
                          <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right' }}>
                            {formatEUR(netto)}
                          </td>
                          <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', color: 'var(--color-text-muted)' }}>
                            {exp.vat_rate}%
                          </td>
                          <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', fontWeight: 600 }}>
                            {formatEUR(brutto)}
                          </td>
                          <td style={{ padding: '0.75rem 0.5rem', textAlign: 'center' }}>
                            {exp.receipt_url ? (
                              <a
                                href={exp.receipt_url}
                                target="_blank"
                                rel="noopener noreferrer"
                                style={{ color: 'var(--color-primary)', textDecoration: 'none', fontSize: '0.8rem' }}
                              >
                                Ansehen
                              </a>
                            ) : (
                              <span style={{ color: 'var(--color-text-muted)', fontSize: '0.75rem' }}>{'\u2014'}</span>
                            )}
                          </td>
                          <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right' }}>
                            <button
                              onClick={() => {
                                if (confirm('Beleg wirklich loeschen?')) {
                                  deleteMutation.mutate(exp.id)
                                }
                              }}
                              style={{
                                padding: '2px 8px',
                                fontSize: '0.75rem',
                                background: 'rgba(239, 68, 68, 0.1)',
                                color: '#ef4444',
                                border: '1px solid rgba(239, 68, 68, 0.2)',
                                borderRadius: '4px',
                                cursor: 'pointer',
                              }}
                            >
                              Loeschen
                            </button>
                          </td>
                        </tr>
                      )
                    })}
                </tbody>
                <tfoot>
                  <tr style={{ borderTop: '2px solid rgba(255,255,255,0.1)' }}>
                    <td colSpan={3} style={{ padding: '0.75rem 0.5rem', fontWeight: 600 }}>
                      Summe ({filteredExpenses.length} Belege)
                    </td>
                    <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', fontWeight: 600 }}>
                      {formatEUR(filteredExpenses.reduce((s, e) => s + (e.amount || 0), 0))}
                    </td>
                    <td />
                    <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', fontWeight: 600 }}>
                      {formatEUR(filteredExpenses.reduce((s, e) => s + (e.amount || 0) * (1 + (e.vat_rate || 0) / 100), 0))}
                    </td>
                    <td colSpan={2} />
                  </tr>
                </tfoot>
              </table>
            </div>
          )}
        </div>
      )}

      {/* EUeR View */}
      {activeView === 'euer' && (
        <div>
          <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
              <label style={{ fontSize: '0.85rem', color: 'var(--color-text-secondary)' }}>Jahr:</label>
              <select
                style={{
                  padding: '0.4rem 0.75rem',
                  background: 'rgba(255,255,255,0.05)',
                  border: '1px solid rgba(255,255,255,0.1)',
                  borderRadius: '6px',
                  color: 'inherit',
                  fontSize: '0.85rem',
                }}
                value={selectedYear}
                onChange={(e) => setSelectedYear(Number(e.target.value))}
              >
                {[2024, 2025, 2026, 2027].map(y => (
                  <option key={y} value={y}>{y}</option>
                ))}
              </select>
            </div>
            <button
              className="btn btn--secondary"
              onClick={exportCSV}
              style={{ padding: '0.4rem 1rem', fontSize: '0.85rem' }}
            >
              CSV exportieren
            </button>
          </div>

          {/* Summary Cards */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1rem', marginBottom: '1.5rem' }}>
            <div style={{
              padding: '1.25rem',
              background: 'var(--color-surface)',
              border: '1px solid rgba(255,255,255,0.06)',
              borderRadius: '8px',
              textAlign: 'center',
            }}>
              <div style={{ fontSize: '0.8rem', color: 'var(--color-text-muted)', marginBottom: '0.5rem' }}>Einnahmen {selectedYear}</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#10b981' }}>{formatEUR(yearTotalIncome)}</div>
            </div>
            <div style={{
              padding: '1.25rem',
              background: 'var(--color-surface)',
              border: '1px solid rgba(255,255,255,0.06)',
              borderRadius: '8px',
              textAlign: 'center',
            }}>
              <div style={{ fontSize: '0.8rem', color: 'var(--color-text-muted)', marginBottom: '0.5rem' }}>Ausgaben {selectedYear}</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#ef4444' }}>{formatEUR(yearTotalExpenses)}</div>
            </div>
            <div style={{
              padding: '1.25rem',
              background: 'var(--color-surface)',
              border: `1px solid ${yearProfit >= 0 ? 'rgba(16, 185, 129, 0.2)' : 'rgba(239, 68, 68, 0.2)'}`,
              borderRadius: '8px',
              textAlign: 'center',
            }}>
              <div style={{ fontSize: '0.8rem', color: 'var(--color-text-muted)', marginBottom: '0.5rem' }}>
                {yearProfit >= 0 ? 'Gewinn' : 'Verlust'} {selectedYear}
              </div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: yearProfit >= 0 ? '#10b981' : '#ef4444' }}>
                {formatEUR(yearProfit)}
              </div>
            </div>
          </div>

          {/* Monthly Table */}
          <div style={{ overflowX: 'auto' }}>
            <table style={{
              width: '100%',
              borderCollapse: 'collapse',
              fontSize: '0.875rem',
            }}>
              <thead>
                <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.08)' }}>
                  <th style={{ textAlign: 'left', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>Monat</th>
                  <th style={{ textAlign: 'right', padding: '0.75rem 0.5rem', color: '#10b981', fontWeight: 500 }}>Einnahmen</th>
                  <th style={{ textAlign: 'right', padding: '0.75rem 0.5rem', color: '#ef4444', fontWeight: 500 }}>Ausgaben</th>
                  <th style={{ textAlign: 'right', padding: '0.75rem 0.5rem', color: 'var(--color-text-muted)', fontWeight: 500 }}>Gewinn/Verlust</th>
                </tr>
              </thead>
              <tbody>
                {monthlyData.map((m) => (
                  <tr key={m.month} style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
                    <td style={{ padding: '0.75rem 0.5rem', fontWeight: 500 }}>{m.label}</td>
                    <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', color: m.income > 0 ? '#10b981' : 'var(--color-text-muted)' }}>
                      {formatEUR(m.income)}
                    </td>
                    <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', color: m.expenses > 0 ? '#ef4444' : 'var(--color-text-muted)' }}>
                      {formatEUR(m.expenses)}
                    </td>
                    <td style={{
                      padding: '0.75rem 0.5rem',
                      textAlign: 'right',
                      fontWeight: 600,
                      color: m.profit >= 0 ? '#10b981' : '#ef4444',
                    }}>
                      {formatEUR(m.profit)}
                    </td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr style={{ borderTop: '2px solid rgba(255,255,255,0.1)' }}>
                  <td style={{ padding: '0.75rem 0.5rem', fontWeight: 700 }}>GESAMT {selectedYear}</td>
                  <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', fontWeight: 700, color: '#10b981' }}>
                    {formatEUR(yearTotalIncome)}
                  </td>
                  <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', fontWeight: 700, color: '#ef4444' }}>
                    {formatEUR(yearTotalExpenses)}
                  </td>
                  <td style={{
                    padding: '0.75rem 0.5rem',
                    textAlign: 'right',
                    fontWeight: 700,
                    color: yearProfit >= 0 ? '#10b981' : '#ef4444',
                  }}>
                    {formatEUR(yearProfit)}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      )}

      {/* New Expense Modal */}
      <Modal
        isOpen={showModal}
        onClose={() => { setShowModal(false); setForm(emptyForm); setReceiptFile(null) }}
        title="Beleg erfassen"
        size="lg"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
            <button
              className="btn btn--secondary"
              onClick={() => { setShowModal(false); setForm(emptyForm); setReceiptFile(null) }}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={handleSubmit}
              disabled={createMutation.isPending}
            >
              {createMutation.isPending ? 'Speichern...' : 'Beleg speichern'}
            </button>
          </div>
        }
      >
        <div className="form-grid">
          {/* KI-Scan Bereich */}
          <div className="form-group form-group--full" style={{
            padding: '1rem',
            background: 'rgba(0, 212, 255, 0.06)',
            border: '1px dashed rgba(0, 212, 255, 0.3)',
            borderRadius: '8px',
            marginBottom: '0.5rem',
          }}>
            <label className="form-label" style={{ color: 'var(--color-primary)', fontWeight: 600 }}>
              KI-Beleganalyse (Foto/PDF hochladen)
            </label>
            <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
              <input
                ref={fileInputRef}
                type="file"
                accept=".pdf,.jpg,.jpeg,.png,.webp,.gif"
                onChange={(e) => {
                  const file = e.target.files?.[0] || null
                  setReceiptFile(file)
                  if (file) analyzeReceipt(file)
                }}
                style={{
                  flex: 1,
                  padding: '0.5rem',
                  background: 'rgba(255,255,255,0.05)',
                  border: '1px solid rgba(255,255,255,0.1)',
                  borderRadius: '6px',
                  color: 'inherit',
                }}
              />
              {receiptFile && !isAnalyzing && (
                <button
                  type="button"
                  className="btn btn--primary"
                  onClick={() => analyzeReceipt(receiptFile)}
                  style={{ padding: '0.5rem 1rem', fontSize: '0.85rem', whiteSpace: 'nowrap' }}
                >
                  Erneut analysieren
                </button>
              )}
            </div>
            {isAnalyzing && (
              <div style={{
                marginTop: '0.75rem',
                padding: '0.75rem',
                background: 'rgba(0, 212, 255, 0.1)',
                borderRadius: '6px',
                fontSize: '0.85rem',
                color: 'var(--color-primary)',
                fontWeight: 500,
                animation: 'pulse 1.5s ease-in-out infinite',
              }}>
                KI analysiert Beleg... (Claude Vision)
              </div>
            )}
            {aiConfidence !== null && !isAnalyzing && (
              <div style={{
                marginTop: '0.5rem',
                fontSize: '0.8rem',
                color: aiConfidence > 0.8 ? '#10b981' : aiConfidence > 0.5 ? '#f59e0b' : '#ef4444',
              }}>
                KI-Konfidenz: {Math.round(aiConfidence * 100)}% — Bitte Daten pruefen und ggf. korrigieren
              </div>
            )}
          </div>

          {/* Belegart */}
          <div className="form-group">
            <label className="form-label">Belegart</label>
            <select
              className="form-input"
              value={form.type}
              onChange={(e) => setForm(prev => ({ ...prev, type: e.target.value as any }))}
            >
              <option value="invoice">Rechnung</option>
              <option value="receipt">Quittung</option>
              <option value="entertainment">Bewirtungsbeleg</option>
            </select>
          </div>

          <Input
            label="Datum *"
            type="date"
            value={form.date}
            onChange={(e) => setForm(prev => ({ ...prev, date: e.target.value }))}
          />

          <Input
            label="Lieferant / Vendor *"
            value={form.vendor}
            onChange={(e) => setForm(prev => ({ ...prev, vendor: e.target.value }))}
            placeholder="z.B. Thomann GmbH"
          />

          <Input
            label="Beschreibung *"
            value={form.description}
            onChange={(e) => setForm(prev => ({ ...prev, description: e.target.value }))}
            placeholder="z.B. Mikrofonkabel 10m"
          />

          <Input
            label="Betrag (Brutto, EUR) *"
            type="number"
            value={form.amount}
            onChange={(e) => setForm(prev => ({ ...prev, amount: e.target.value }))}
            placeholder="0.00"
            min="0"
            step="0.01"
          />

          <div className="form-group">
            <label className="form-label">MwSt-Satz</label>
            <select
              className="form-input"
              value={form.vat_rate}
              onChange={(e) => setForm(prev => ({ ...prev, vat_rate: Number(e.target.value) }))}
            >
              {VAT_RATES.map(v => (
                <option key={v.value} value={v.value}>{v.label}</option>
              ))}
            </select>
          </div>

          <Input
            label="Rechnungsnummer"
            value={form.invoice_number}
            onChange={(e) => setForm(prev => ({ ...prev, invoice_number: e.target.value }))}
            placeholder="RE-2026-001"
          />

          <Input
            label="USt-IdNr Lieferant"
            value={form.vendor_vat_id}
            onChange={(e) => setForm(prev => ({ ...prev, vendor_vat_id: e.target.value }))}
            placeholder="DE123456789"
          />

          <Input
            label="IBAN Lieferant"
            value={form.vendor_iban}
            onChange={(e) => setForm(prev => ({ ...prev, vendor_iban: e.target.value }))}
            placeholder="DE89..."
          />

          <Input
            label="Buchungskonto (SKR03)"
            value={form.booking_account}
            onChange={(e) => setForm(prev => ({ ...prev, booking_account: e.target.value }))}
            placeholder="z.B. 4930"
          />

          <div className="form-group">
            <label className="form-label">Kategorie</label>
            <select
              className="form-input"
              value={form.category}
              onChange={(e) => setForm(prev => ({ ...prev, category: e.target.value }))}
            >
              {CATEGORIES.map(c => (
                <option key={c.value} value={c.value}>{c.label}</option>
              ))}
            </select>
          </div>

          <Input
            label="Zahlungsziel"
            type="date"
            value={form.due_date}
            onChange={(e) => setForm(prev => ({ ...prev, due_date: e.target.value }))}
          />

          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <div style={{ flex: 1 }}>
              <Input
                label="Skonto %"
                type="number"
                value={form.discount_percent}
                onChange={(e) => setForm(prev => ({ ...prev, discount_percent: e.target.value }))}
                placeholder="2"
                min="0"
                step="0.5"
              />
            </div>
            <div style={{ flex: 1 }}>
              <Input
                label="Skonto Tage"
                type="number"
                value={form.discount_days}
                onChange={(e) => setForm(prev => ({ ...prev, discount_days: e.target.value }))}
                placeholder="10"
                min="0"
              />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Zahlungsart</label>
            <select
              className="form-input"
              value={form.payment_method}
              onChange={(e) => setForm(prev => ({ ...prev, payment_method: e.target.value }))}
            >
              <option value="bank_transfer">Ueberweisung</option>
              <option value="cash">Bar</option>
              <option value="card">Karte</option>
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Projekt (optional)</label>
            <select
              className="form-input"
              value={form.project_id}
              onChange={(e) => setForm(prev => ({ ...prev, project_id: e.target.value }))}
            >
              <option value="">-- Kein Projekt --</option>
              {projects.map(p => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </div>

          <Input
            label="Lieferant-Adresse"
            value={form.vendor_address}
            onChange={(e) => setForm(prev => ({ ...prev, vendor_address: e.target.value }))}
            placeholder="Strasse, PLZ Ort"
          />

          {/* Bewirtungsbeleg-Felder */}
          {form.type === 'entertainment' && (
            <>
              <div className="form-group form-group--full" style={{
                padding: '0.75rem',
                background: 'rgba(245, 158, 11, 0.08)',
                border: '1px solid rgba(245, 158, 11, 0.2)',
                borderRadius: '8px',
              }}>
                <label style={{ fontWeight: 600, color: '#f59e0b', fontSize: '0.85rem', display: 'block', marginBottom: '0.5rem' }}>
                  Bewirtungsbeleg-Angaben (steuerlich erforderlich)
                </label>
                <div className="form-grid" style={{ gap: '0.75rem' }}>
                  <Input
                    label="Ort der Bewirtung"
                    value={form.entertainment_location}
                    onChange={(e) => setForm(prev => ({ ...prev, entertainment_location: e.target.value }))}
                    placeholder="z.B. Restaurant Maximilians, Muenchen"
                  />
                  <Input
                    label="Trinkgeld (EUR)"
                    type="number"
                    value={form.entertainment_tip}
                    onChange={(e) => setForm(prev => ({ ...prev, entertainment_tip: e.target.value }))}
                    placeholder="0.00"
                    min="0"
                    step="0.01"
                  />
                  <div className="form-group form-group--full">
                    <Input
                      label="Anlass der Bewirtung *"
                      value={form.entertainment_reason}
                      onChange={(e) => setForm(prev => ({ ...prev, entertainment_reason: e.target.value }))}
                      placeholder="z.B. Projektbesprechung Stadtfest 2026"
                    />
                  </div>
                  <div className="form-group form-group--full">
                    <Input
                      label="Teilnehmer (komma-separiert) *"
                      value={form.entertainment_guests}
                      onChange={(e) => setForm(prev => ({ ...prev, entertainment_guests: e.target.value }))}
                      placeholder="z.B. Max Mueller, Lisa Schmidt, Janis Eckersberger"
                    />
                  </div>
                </div>
                {form.amount && (
                  <div style={{ marginTop: '0.5rem', fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>
                    70% absetzbar: {(parseFloat(form.amount || '0') * 0.70).toFixed(2)} EUR |
                    30% nicht absetzbar: {(parseFloat(form.amount || '0') * 0.30).toFixed(2)} EUR
                  </div>
                )}
              </div>
            </>
          )}

          <div className="form-group form-group--full">
            <label className="form-label">Bemerkungen</label>
            <textarea
              className="form-input"
              value={form.notes}
              onChange={(e) => setForm(prev => ({ ...prev, notes: e.target.value }))}
              placeholder="Optionale Bemerkungen..."
              rows={2}
              style={{ resize: 'vertical' }}
            />
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default ExpensesPage
