export function RevenueChartWidget() {
  // Demo monthly revenue data
  const months = [
    { month: 'Okt', revenue: 12400 },
    { month: 'Nov', revenue: 18200 },
    { month: 'Dez', revenue: 15800 },
    { month: 'Jan', revenue: 22100 },
    { month: 'Feb', revenue: 19500 },
    { month: 'Mär', revenue: 24800 },
  ]

  const maxRevenue = Math.max(...months.map(m => m.revenue))

  return (
    <div className="revenue-chart-widget">
      <div className="revenue-chart-widget__summary">
        <div className="revenue-chart-widget__total">
          <span className="revenue-chart-widget__total-value">
            {(months.reduce((sum, m) => sum + m.revenue, 0)).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' })}
          </span>
          <span className="revenue-chart-widget__total-label">Umsatz (6 Monate)</span>
        </div>
        <div className="revenue-chart-widget__trend revenue-chart-widget__trend--positive">
          ↑ 12.3% zum Vorjahr
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
