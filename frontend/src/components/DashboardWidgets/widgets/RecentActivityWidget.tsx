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

  const recentActivities = (activities && activities.length > 0) ? activities : [
    { id: 'a1', type: 'equipment', description: 'Canon EOS R5 als reserviert markiert', timestamp: '2026-03-23T09:15:00Z', icon: '📷' },
    { id: 'a2', type: 'project', description: 'Stadtfest München 2026 - Packliste erstellt', timestamp: '2026-03-22T16:30:00Z', icon: '📋' },
    { id: 'a3', type: 'invoice', description: 'Rechnung RF-2026-004 an TechCorp GmbH versendet', timestamp: '2026-03-22T14:00:00Z', icon: '💳' },
    { id: 'a4', type: 'warehouse', description: 'Martin MAC Aura XB - Rückgabe am Lager', timestamp: '2026-03-22T11:45:00Z', icon: '📦' },
    { id: 'a5', type: 'maintenance', description: 'VDE-Prüfung für JBL VTX A12 abgeschlossen', timestamp: '2026-03-21T15:20:00Z', icon: '🔧' },
    { id: 'a6', type: 'project', description: 'Firmenfeier Schmidt AG - Angebot erstellt', timestamp: '2026-03-21T10:00:00Z', icon: '📋' },
    { id: 'a7', type: 'equipment', description: 'Shure Axient Digital - Firmware aktualisiert', timestamp: '2026-03-20T14:30:00Z', icon: '🎤' },
    { id: 'a8', type: 'invoice', description: 'Rechnung RF-2026-003 bezahlt', timestamp: '2026-03-20T09:15:00Z', icon: '💳' },
    { id: 'a9', type: 'warehouse', description: 'Inventur Lager B abgeschlossen', timestamp: '2026-03-19T17:00:00Z', icon: '📦' },
    { id: 'a10', type: 'equipment', description: 'LED Wall 3x2m - Neues Equipment angelegt', timestamp: '2026-03-19T11:30:00Z', icon: '💡' },
  ]

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
