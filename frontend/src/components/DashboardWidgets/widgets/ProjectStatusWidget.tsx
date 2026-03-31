import { useQuery } from '@tanstack/react-query'
import { projectApi } from '../../../services/api'

const STATUS_CONFIG: Record<string, { label: string; color: string }> = {
  planning: { label: 'Planung', color: '#3b82f6' },
  confirmed: { label: 'Bestätigt', color: '#8b5cf6' },
  active: { label: 'Aktiv', color: '#10b981' },
  in_progress: { label: 'Aktiv', color: '#10b981' },
  loading: { label: 'Laden', color: '#f59e0b' },
  on_site: { label: 'Vor Ort', color: '#06b6d4' },
  completed: { label: 'Abgeschlossen', color: '#6b7280' },
  cancelled: { label: 'Storniert', color: '#ef4444' },
  draft: { label: 'Entwurf', color: '#9ca3af' },
}

export function ProjectStatusWidget() {
  const { data: projectData, isLoading } = useQuery({
    queryKey: ['dashboard-project-status'],
    queryFn: async () => {
      try {
        const res = await projectApi.list(1, 200)
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const projects: any[] = res?.items || []
        if (!Array.isArray(projects)) return []
        const counts: Record<string, number> = {}
        for (const p of projects) {
          const status = p.status || 'draft'
          counts[status] = (counts[status] || 0) + 1
        }
        return Object.entries(counts).map(([status, count]) => {
          const config = STATUS_CONFIG[status] || { label: status, color: '#9ca3af' }
          return { label: config.label, count, color: config.color }
        })
      } catch {
        return []
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  const statuses = projectData || []
  const total = statuses.reduce((sum, s) => sum + s.count, 0)

  if (isLoading) {
    return (
      <div className="project-status-widget">
        <p style={{ color: 'var(--color-text-muted)', textAlign: 'center', padding: '2rem' }}>Laden...</p>
      </div>
    )
  }

  if (total === 0) {
    return (
      <div className="project-status-widget">
        <p style={{ color: 'var(--color-text-muted)', textAlign: 'center', padding: '2rem' }}>
          Noch keine Projekte vorhanden.
        </p>
      </div>
    )
  }

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
