import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Project } from '../../../types/project'
import { invoiceApi, reservationApi } from '../../../services/api'
import styles from '../ProjectDetail.module.scss'

interface FinanceTabProps {
  project: Project
}

interface Invoice {
  id: string
  invoice_number?: string
  number?: string
  project_id: string
  client_name: string
  status: string
  sub_total?: number
  subtotal?: number
  tax_amount?: number
  tax_total?: number
  total: number
  currency?: string
  issue_date: string
  due_date: string
  paid_date?: string | null
  paid_amount?: number
  items?: any[]
}

const STATUS_LABELS: Record<string, string> = {
  draft: 'Entwurf',
  sent: 'Versendet',
  paid: 'Bezahlt',
  overdue: 'Ueberfaellig',
  cancelled: 'Storniert',
  partial: 'Teilbezahlt',
}

const STATUS_COLORS: Record<string, { bg: string; color: string }> = {
  draft: { bg: 'rgba(107, 114, 128, 0.15)', color: 'var(--color-text-muted)' },
  sent: { bg: 'rgba(59, 130, 246, 0.15)', color: '#3b82f6' },
  paid: { bg: 'rgba(16, 185, 129, 0.15)', color: 'var(--color-success)' },
  overdue: { bg: 'rgba(239, 68, 68, 0.15)', color: 'var(--color-danger)' },
  cancelled: { bg: 'rgba(107, 114, 128, 0.15)', color: 'var(--color-text-muted)' },
  partial: { bg: 'rgba(245, 158, 11, 0.15)', color: '#f59e0b' },
}

export function FinanceTab({ project }: FinanceTabProps) {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const [autoInvoiceError, setAutoInvoiceError] = useState<string | null>(null)

  const { data: invoicesData, isLoading, error } = useQuery({
    queryKey: ['project-invoices', project.id],
    queryFn: () => invoiceApi.list(1, 50),
  })

  // Fetch reservations for auto-invoice
  const { data: reservations = [] } = useQuery({
    queryKey: ['project-reservations', project.id],
    queryFn: () => reservationApi.list(project.id),
  })

  // Auto-generate invoice from equipment reservations
  const autoInvoiceMutation = useMutation({
    mutationFn: async () => {
      setAutoInvoiceError(null)
      const reservationList = Array.isArray(reservations)
        ? reservations
        : (reservations as any)?.data ?? (reservations as any)?.items ?? []

      if (reservationList.length === 0) {
        throw new Error('Keine Equipment-Reservierungen vorhanden. Fuegen Sie zuerst Equipment im Equipment-Tab hinzu.')
      }

      // Build invoice items from reservations
      const items = reservationList.map((res: any) => {
        const startDate = new Date(res.start_date)
        const endDate = new Date(res.end_date)
        const rentalDays = Math.max(1, Math.ceil((endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24)))
        const quantity = res.quantity || 1
        const dailyRate = res.unit_price || res.daily_rate || 0
        const totalPrice = quantity * dailyRate * rentalDays

        return {
          description: res.equipment_name || `Equipment ${res.equipment_id?.slice(0, 8) || ''}`,
          quantity: quantity,
          unit: 'Tage',
          unit_price: dailyRate,
          days: rentalDays,
          total_price: totalPrice,
        }
      })

      const subtotal = items.reduce((sum: number, item: any) => sum + item.total_price, 0)
      const taxRate = 0.19
      const taxAmount = subtotal * taxRate
      const total = subtotal + taxAmount

      // Create invoice via API
      const invoiceData = {
        project_id: project.id,
        client_name: project.client_name || 'Unbekannt',
        client_email: project.client_email || '',
        status: 'draft',
        issue_date: new Date().toISOString().split('T')[0],
        due_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        sub_total: subtotal,
        tax_rate: taxRate,
        tax_amount: taxAmount,
        total: total,
        currency: project.currency || 'EUR',
        items: items,
        notes: `Auto-generiert aus Projekt: ${project.name}`,
      }

      return invoiceApi.create(invoiceData)
    },
    onSuccess: (result: any) => {
      queryClient.invalidateQueries({ queryKey: ['project-invoices', project.id] })
      // Navigate to the new invoice
      const invoiceId = result?.id || result?.data?.id
      if (invoiceId) {
        navigate(`/invoices/${invoiceId}`)
      }
    },
    onError: (err: any) => {
      setAutoInvoiceError(err?.message || 'Fehler beim Erstellen der Auto-Rechnung')
    },
  })

  // Filter invoices for this project client-side (API might not support project_id filter)
  const allInvoices: Invoice[] = Array.isArray(invoicesData)
    ? invoicesData
    : invoicesData?.data ?? invoicesData?.items ?? []

  const projectInvoices = allInvoices.filter(
    (inv) => inv.project_id === project.id,
  )

  // Compute financials
  const budget = project.budget || 0
  const totalInvoiced = projectInvoices.reduce((s, inv) => s + (inv.total || 0), 0)
  const totalPaid = projectInvoices
    .filter((inv) => inv.status === 'paid')
    .reduce((s, inv) => s + (inv.total || 0), 0)
  const totalOpen = totalInvoiced - totalPaid

  const invoicedPercent = budget > 0 ? Math.min(100, Math.round((totalInvoiced / budget) * 100)) : 0
  const paidPercent = totalInvoiced > 0 ? Math.round((totalPaid / totalInvoiced) * 100) : 0

  const summaryCards = [
    { label: 'Budget', value: budget, cls: styles.financeCardValueRevenue },
    { label: 'In Rechnung gestellt', value: totalInvoiced, cls: styles.financeCardValueCost },
    { label: 'Bezahlt', value: totalPaid, cls: styles.financeCardValueProfit },
    { label: 'Offen', value: totalOpen, cls: styles.financeCardValueDiscount },
  ]

  const formatDate = (d: string) => new Date(d).toLocaleDateString('de-DE')

  return (
    <div>
      {/* Summary Cards */}
      <div className={styles.financeSummary}>
        {summaryCards.map((card) => (
          <div key={card.label} className={styles.financeCard}>
            <div className={styles.financeCardLabel}>{card.label}</div>
            <div className={`${styles.financeCardValue} ${card.cls}`}>
              {'\u20AC'}{card.value.toFixed(2)}
            </div>
          </div>
        ))}
      </div>

      {/* Progress bars */}
      <div className={styles.glassCard} style={{ marginBottom: 'var(--spacing-6)' }}>
        <div style={{ display: 'flex', gap: 'var(--spacing-8)', justifyContent: 'center' }}>
          <div style={{ textAlign: 'center', flex: 1 }}>
            <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', marginBottom: 'var(--spacing-2)', textTransform: 'uppercase', letterSpacing: 'var(--letter-spacing-wide)' }}>
              In Rechnung gestellt
            </div>
            <div style={{
              fontSize: 'var(--font-size-3xl)',
              fontWeight: 'var(--font-weight-bold)',
              color: invoicedPercent > 0 ? 'var(--color-primary)' : 'var(--color-text-muted)',
            }}>
              {invoicedPercent}%
            </div>
          </div>
          <div style={{ textAlign: 'center', flex: 1 }}>
            <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', marginBottom: 'var(--spacing-2)', textTransform: 'uppercase', letterSpacing: 'var(--letter-spacing-wide)' }}>
              Bezahlt
            </div>
            <div style={{
              fontSize: 'var(--font-size-3xl)',
              fontWeight: 'var(--font-weight-bold)',
              color: paidPercent >= 100 ? 'var(--color-success)' : paidPercent > 0 ? '#f59e0b' : 'var(--color-text-muted)',
            }}>
              {paidPercent}%
            </div>
          </div>
        </div>
      </div>

      {/* Invoices Section */}
      <div className={styles.glassCard} style={{ marginBottom: 'var(--spacing-5)' }}>
        <div className={styles.sectionHeader}>
          <div className={styles.glassCardTitle} style={{ margin: 0 }}>
            Rechnungen ({projectInvoices.length})
          </div>
          <div style={{ display: 'flex', gap: 'var(--spacing-2)', alignItems: 'center' }}>
            <button
              className="btn btn--primary"
              onClick={() => autoInvoiceMutation.mutate()}
              disabled={autoInvoiceMutation.isPending}
              title="Erstellt automatisch eine Rechnung aus allen Equipment-Reservierungen dieses Projekts"
              style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)' }}
            >
              {autoInvoiceMutation.isPending ? 'Erstelle...' : '\u26A1 Auto-Rechnung'}
            </button>
            <button className="btn btn--secondary" disabled>Rechnung erstellen</button>
          </div>
        </div>
        {autoInvoiceError && (
          <div style={{
            background: 'rgba(239, 68, 68, 0.1)',
            border: '1px solid rgba(239, 68, 68, 0.3)',
            borderRadius: 'var(--radius-sm)',
            padding: 'var(--spacing-3) var(--spacing-4)',
            marginBottom: 'var(--spacing-3)',
            color: 'var(--color-danger)',
            fontSize: 'var(--font-size-sm)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}>
            <span>{autoInvoiceError}</span>
            <button
              onClick={() => setAutoInvoiceError(null)}
              style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer', fontSize: 'var(--font-size-base)' }}
            >
              {'\u2715'}
            </button>
          </div>
        )}

        {isLoading ? (
          <div className={styles.emptyState} style={{ padding: 'var(--spacing-6)' }}>
            <p className={styles.emptyStateText}>Rechnungen werden geladen...</p>
          </div>
        ) : error ? (
          <div className={styles.emptyState} style={{ padding: 'var(--spacing-6)' }}>
            <p className={styles.emptyStateText}>Fehler beim Laden der Rechnungen.</p>
          </div>
        ) : projectInvoices.length === 0 ? (
          <div className={styles.emptyState} style={{ padding: 'var(--spacing-6)' }}>
            <p className={styles.emptyStateText}>Noch keine Rechnungen fuer dieses Projekt.</p>
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className={styles.dataTable}>
              <thead>
                <tr>
                  <th>Nummer</th>
                  <th>Datum</th>
                  <th>Faellig</th>
                  <th style={{ textAlign: 'right' }}>Betrag</th>
                  <th style={{ textAlign: 'center' }}>Status</th>
                </tr>
              </thead>
              <tbody>
                {projectInvoices.map((inv) => {
                  const sc = STATUS_COLORS[inv.status] || STATUS_COLORS.draft
                  return (
                    <tr key={inv.id}>
                      <td style={{ fontWeight: 'var(--font-weight-medium)' }}>
                        {inv.invoice_number || inv.number || inv.id.slice(0, 12)}
                      </td>
                      <td>{formatDate(inv.issue_date)}</td>
                      <td>
                        {formatDate(inv.due_date)}
                        {inv.paid_date && (
                          <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-success)' }}>
                            Bezahlt: {formatDate(inv.paid_date)}
                          </div>
                        )}
                      </td>
                      <td style={{ textAlign: 'right', fontWeight: 'var(--font-weight-medium)' }}>
                        {'\u20AC'}{inv.total.toFixed(2)}
                      </td>
                      <td style={{ textAlign: 'center' }}>
                        <span
                          style={{
                            display: 'inline-block',
                            padding: 'var(--spacing-1) var(--spacing-2)',
                            borderRadius: 'var(--radius-full)',
                            fontSize: 'var(--font-size-xs)',
                            fontWeight: 'var(--font-weight-medium)',
                            background: sc.bg,
                            color: sc.color,
                          }}
                        >
                          {STATUS_LABELS[inv.status] || inv.status}
                        </span>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
              <tfoot>
                <tr>
                  <td colSpan={3}>Gesamt</td>
                  <td style={{ textAlign: 'right', fontWeight: 'var(--font-weight-bold)' }}>
                    {'\u20AC'}{totalInvoiced.toFixed(2)}
                  </td>
                  <td />
                </tr>
              </tfoot>
            </table>
          </div>
        )}
      </div>

      {/* Financial Breakdown (static categories based on invoices) */}
      {projectInvoices.length > 0 && (
        <div className={styles.glassCard}>
          <div className={styles.glassCardTitle}>Rechnungspositionen</div>
          <div style={{ overflowX: 'auto' }}>
            <table className={styles.dataTable}>
              <thead>
                <tr>
                  <th>Position</th>
                  <th style={{ textAlign: 'right' }}>Menge</th>
                  <th style={{ textAlign: 'right' }}>Einzelpreis</th>
                  <th style={{ textAlign: 'right' }}>Gesamt</th>
                </tr>
              </thead>
              <tbody>
                {projectInvoices.flatMap((inv) =>
                  (inv.items || []).map((item: any) => (
                    <tr key={item.id}>
                      <td>{item.description || item.name}</td>
                      <td style={{ textAlign: 'right' }}>
                        {item.quantity} {item.unit || ''}
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        {'\u20AC'}{(item.unit_price || 0).toFixed(2)}
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        {'\u20AC'}{(item.total_price || item.total || 0).toFixed(2)}
                      </td>
                    </tr>
                  )),
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
