export function ProjectStatusWidget() {
  const statuses = [
    { label: 'Planung', count: 3, color: '#3b82f6', percentage: 25 },
    { label: 'Aktiv', count: 5, color: '#10b981', percentage: 42 },
    { label: 'Abgeschlossen', count: 3, color: '#6b7280', percentage: 25 },
    { label: 'Storniert', count: 1, color: '#ef4444', percentage: 8 },
  ]
  const total = statuses.reduce((sum, s) => sum + s.count, 0)

  // Build donut chart with SVG
  const radius = 50
  const circumference = 2 * Math.PI * radius
  let offset = 0

  return (
    <div className="project-status-widget">
      <div className="project-status-widget__chart">
        <svg viewBox="0 0 120 120" className="project-status-widget__donut">
          {statuses.map((status, i) => {
            const dashLength = (status.count / total) * circumference
            const dashOffset = -offset
            offset += dashLength
            return (
              <circle
                key={i}
                cx="60"
                cy="60"
                r={radius}
                fill="none"
                stroke={status.color}
                strokeWidth="12"
                strokeDasharray={`${dashLength} ${circumference - dashLength}`}
                strokeDashoffset={dashOffset}
                transform="rotate(-90 60 60)"
                style={{ transition: 'stroke-dasharray 0.8s ease' }}
              />
            )
          })}
          <text x="60" y="56" textAnchor="middle" fill="var(--color-text-primary)" fontSize="20" fontWeight="700">{total}</text>
          <text x="60" y="72" textAnchor="middle" fill="var(--color-text-muted)" fontSize="10">Projekte</text>
        </svg>
      </div>

      <div className="project-status-widget__legend">
        {statuses.map((status, i) => (
          <div key={i} className="project-status-widget__legend-item">
            <span className="project-status-widget__legend-dot" style={{ background: status.color }} />
            <span className="project-status-widget__legend-label">{status.label}</span>
            <span className="project-status-widget__legend-count">{status.count}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
