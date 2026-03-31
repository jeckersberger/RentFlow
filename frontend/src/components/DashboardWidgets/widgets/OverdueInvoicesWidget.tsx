import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { invoiceApi } from '../../../services/api'

export function OverdueInvoicesWidget() {
  const navigate = useNavigate()

  const { data: overdueInvoiceData } = useQuery<{ count: number; total: number }>({
    queryKey: ['dashboard-overdue-invoices'],
    queryFn: async () => {
      try {
        const res = await invoiceApi.list(1, 100)
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const invoices: any[] = res?.data || res?.items || []
        if (!Array.isArray(invoices)) return { count: 0, total: 0 }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const overdue = invoices.filter((inv: any) => inv.status === 'overdue')
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const total = overdue.reduce((sum: number, inv: any) => sum + (inv.total || 0), 0)
        return { count: overdue.length, total }
      } catch {
        return { count: 0, total: 0 }
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  if (!overdueInvoiceData || overdueInvoiceData.count === 0) {
    return (
      <div className="widget-empty">
        <span className="widget-empty__icon">✓</span>
        <p className="widget-empty__text">Keine überfälligen Rechnungen.</p>
      </div>
    )
  }

  return (
    <div className="overdue-widget">
      <div className="overdue-widget__stats">
        <div className="overdue-widget__stat">
          <span className="overdue-widget__number overdue-widget__number--danger">{overdueInvoiceData.count}</span>
          <span className="overdue-widget__label">Rechnungen</span>
        </div>
        <div className="overdue-widget__stat">
          <span className="overdue-widget__number">{overdueInvoiceData.total.toLocaleString('de-DE', { style: 'currency', currency: 'EUR' })}</span>
          <span className="overdue-widget__label">Gesamtbetrag</span>
        </div>
      </div>
      <button className="widget-link" onClick={() => navigate('/invoices')}>
        Rechnungen anzeigen →
      </button>
    </div>
  )
}
