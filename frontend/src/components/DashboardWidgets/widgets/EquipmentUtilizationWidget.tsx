import { useQuery } from '@tanstack/react-query'
import { api } from '../../../services/api'

export function EquipmentUtilizationWidget() {
  const { data: stats } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      try {
        const res = await api.get('/api/v1/equipment')
        const equipment = res.data?.data || res.data?.items || res.data || []
        if (Array.isArray(equipment)) {
          const total = equipment.length
          const available = equipment.filter((e: Record<string, string>) => e.status === 'available').length
          return { total, available }
        }
        return { total: 0, available: 0 }
      } catch {
        return { total: 0, available: 0 }
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  const total = stats?.total || 0
  const available = stats?.available || 0
  const inUse = total - available
  const utilization = total > 0 ? Math.round((inUse / total) * 100) : 0

  const categories = [
    { name: 'Audio', utilization: 78, total: 24 },
    { name: 'Licht', utilization: 65, total: 18 },
    { name: 'Video', utilization: 82, total: 12 },
    { name: 'Rigging', utilization: 45, total: 8 },
    { name: 'Strom', utilization: 55, total: 15 },
  ]

  return (
    <div className="utilization-widget">
      <div className="utilization-widget__overview">
        <div className="utilization-widget__gauge">
          <svg viewBox="0 0 100 60" className="utilization-widget__gauge-svg">
            <path
              d="M 10 55 A 40 40 0 0 1 90 55"
              fill="none"
              stroke="var(--color-bg-tertiary)"
              strokeWidth="8"
              strokeLinecap="round"
            />
            <path
              d="M 10 55 A 40 40 0 0 1 90 55"
              fill="none"
              stroke="url(#utilGradient)"
              strokeWidth="8"
              strokeLinecap="round"
              strokeDasharray={`${(utilization / 100) * 126} 126`}
              style={{ transition: 'stroke-dasharray 1s ease' }}
            />
            <defs>
              <linearGradient id="utilGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#00d4ff" />
                <stop offset="100%" stopColor="#8b5cf6" />
              </linearGradient>
            </defs>
            <text x="50" y="48" textAnchor="middle" fill="var(--color-text-primary)" fontSize="18" fontWeight="700">{utilization}%</text>
            <text x="50" y="58" textAnchor="middle" fill="var(--color-text-muted)" fontSize="7">Auslastung</text>
          </svg>
        </div>
        <div className="utilization-widget__stats">
          <div className="utilization-widget__stat">
            <span className="utilization-widget__stat-value">{inUse}</span>
            <span className="utilization-widget__stat-label">In Verwendung</span>
          </div>
          <div className="utilization-widget__stat">
            <span className="utilization-widget__stat-value">{available}</span>
            <span className="utilization-widget__stat-label">Verfügbar</span>
          </div>
        </div>
      </div>

      <div className="utilization-widget__categories">
        {categories.map((cat, i) => (
          <div key={i} className="utilization-widget__cat">
            <div className="utilization-widget__cat-header">
              <span className="utilization-widget__cat-name">{cat.name}</span>
              <span className="utilization-widget__cat-pct">{cat.utilization}%</span>
            </div>
            <div className="utilization-widget__cat-bar">
              <div
                className="utilization-widget__cat-fill"
                style={{ width: `${cat.utilization}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
