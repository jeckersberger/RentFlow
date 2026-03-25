import { useQuery } from '@tanstack/react-query'
import { api } from '../../../services/api'

interface ActivityEvent {
  id: string
  type: string
  description: string
  timestamp: string
  icon: string
}

export function RecentActivityWidget() {
  const { data: activities } = useQuery<ActivityEvent[]>({
    queryKey: ['dashboard-activity'],
    queryFn: () =>
      api.get('/api/v1/dashboard/activity').then(res => Array.isArray(res.data) ? res.data : []).catch(() => []),
    staleTime: 1000 * 60 * 2,
  })

  const recentActivities = activities || []

  if (recentActivities.length === 0) {
    return (
      <div className="activity-feed">
        <p style={{ color: 'var(--color-text-muted)', textAlign: 'center', padding: '1rem' }}>
          Noch keine Aktivitäten vorhanden.
        </p>
      </div>
    )
  }

  return (
    <div className="activity-feed">
      {recentActivities.slice(0, 10).map((activity) => (
        <div key={activity.id} className="activity-item">
          <div className="activity-icon">{activity.icon}</div>
          <div className="activity-content">
            <p className="activity-description">{activity.description}</p>
            <p className="activity-time">
              {new Date(activity.timestamp).toLocaleString('de-DE')}
            </p>
          </div>
        </div>
      ))}
    </div>
  )
}
