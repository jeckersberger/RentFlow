import { useQuery } from '@tanstack/react-query'
import { invoiceApi } from '../../../services/api'

export function RevenueChartWidget() {
  const { data: months, isLoading } = useQuery({
    queryKey: ['dashboard-revenue-chart'],
    queryFn: async () => {
      try {
        const res = await invoiceApi.list(1, 500)
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const invoices: any[] = res?.data || res?.items || []
        if (!Array.isArray(invoices) || invoices.length === 0) return []

        const monthNames = ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez']
        const now = new Date()
        const result: { month: string; revenue: number }[] = []

        for (let i = 5; i >= 0; i--) {
          const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
          const y = d.getFullYear()
          const m = d.getMonth()
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const monthRevenue = invoices.filter((inv: any) => {
            if (inv.status === 'cancelled' || inv.status === 'draft') return false
            const invDate = new Date(inv.invoice_date || inv.created_at)
            return invDate.getFullYear() === y && invDate.getMonth() === m
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          }).reduce((sum: number, inv: any) => sum + (inv.total || 0), 0)
          result.push({ month: monthNames[m], revenue: monthRevenue })
        }
        return result
      } catch {
        return []
      }
    },
    staleTime: 1000 * 60 * 10,
  })

  if (isLoading) {
    return (
      <div className="revenue-chart-widget">
        <div className="revenue-chart-widget__summary">
          <div className="revenue-chart-widget__total">
            <span className="revenue-chart-widget__total-value">—</span>
            <span className="revenue-chart-widget__total-label">Umsatz (6 Monate)</span>
          </div>
        </div>
      </div>
    )
  }

  if (!months || months.length === 0) {
    return (
      <div className="revenue-chart-widget">
        <div className="revenue-chart-widget__summary">
          <div className="revenue-chart-widget__total">
            <span className="revenue-chart-widget__total-value">
              {(0).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' })}
            </span>
            <span className="revenue-chart-widget__total-label">Umsatz (6 Monate)</span>
          </div>
        </div>
        <p style={{ color: 'var(--color-text-muted)', textAlign: 'center', padding: '1rem' }}>
          Noch keine Rechnungsdaten vorhanden.
        </p>
      </div>
    )
  }

  const maxRevenue = Math.max(...months.map(m => m.revenue), 1)
  const totalRevenue = months.reduce((sum, m) => sum + m.revenue, 0)

  return (
    <div className="revenue-chart-widget">
      <div className="revenue-chart-widget__summary">
        <div className="revenue-chart-widget__total">
          <span className="revenue-chart-widget__total-value">
            {totalRevenue.toLocaleString('de-DE', { style: 'currency', currency: 'EUR' })}
          </span>
          <span className="revenue-chart-widget__total-label">Umsatz (6 Monate)</span>
        </div>
      </div>

      <div className="revenue-chart-widget__bars">
        {months.map((m, i) => (
          <div key={i} className="revenue-chart-widget__bar-col">
            <div className="revenue-chart-widget__bar-wrapper">
              <div
                className="revenue-chart-widget__bar"
                style={{ height: `${(m.revenue / maxRevenue) * 100}%` }}
                title={m.revenue.toLocaleString('de-DE', { style: 'currency', currency: 'EUR' })}
              />
            </div>
            <span className="revenue-chart-widget__bar-label">{m.month}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
