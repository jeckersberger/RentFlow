import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { api } from '../services/api'
import { KPICard } from '../components/KPICard/KPICard'
import './Dashboard.scss'

interface DashboardStats {
  total_equipment: number
  available_equipment: number
  active_projects: number
  pending_invoices: number
  total_revenue?: number
}

interface ActivityEvent {
  id: string
  type: string
  description: string
  timestamp: string
  icon: string
}

function DashboardPage() {
  const navigate = useNavigate()

  const { data: stats, isLoading, error } = useQuery<DashboardStats>({
    queryKey: ['dashboard-stats'],
    queryFn: () => api.get('/api/v1/dashboard/stats').then(res => res.data),
    staleTime: 1000 * 60 * 5,
  })

  const { data: activities } = useQuery<ActivityEvent[]>({
    queryKey: ['dashboard-activity'],
    queryFn: () =>
      api.get('/api/v1/dashboard/activity').then(res => res.data || []).catch(() => []),
    staleTime: 1000 * 60 * 2,
  })

  const equipmentDistribution = [
    { status: 'Verfügbar', count: stats?.available_equipment || 0, percentage: stats?.total_equipment ? Math.round((stats.available_equipment / stats.total_equipment) * 100) : 0 },
    { status: 'Vermietet', count: 12, percentage: 35 },
    { status: 'Reserviert', count: 8, percentage: 25 },
    { status: 'Wartung', count: 3, percentage: 10 },
  ]

  const handleQuickAction = (action: string) => {
    switch (action) {
      case 'project':
        navigate('/projects/new')
        break
      case 'equipment':
        navigate('/equipment/new')
        break
      case 'invoice':
        navigate('/invoices/new')
        break
      case 'scanner':
        navigate('/scanner')
        break
    }
  }

  return (
    <div className="dashboard">
      <div className="dashboard__header">
        <h1 className="dashboard__title">Dashboard</h1>
        <p className="dashboard__subtitle">Willkommen zurück! Hier ist ein Überblick über Ihr Geschäft.</p>
      </div>

      {error && (
        <div className="dashboard__error" role="alert">
          Fehler beim Laden des Dashboards. Bitte versuchen Sie es später erneut.
        </div>
      )}

      <div className="dashboard__kpi-grid">
        <KPICard
          title="Ausrüstung gesamt"
          value={stats?.total_equipment || 0}
          icon="📦"
          isLoading={isLoading}
        />
        <KPICard
          title="Verfügbar"
          value={stats?.available_equipment || 0}
          icon="✅"
          isLoading={isLoading}
        />
        <KPICard
          title="Aktive Projekte"
          value={stats?.active_projects || 0}
          icon="📋"
          isLoading={isLoading}
        />
        <KPICard
          title="Ausstehende Rechnungen"
          value={stats?.pending_invoices || 0}
          icon="💰"
          isLoading={isLoading}
        />
      </div>

      <div className="dashboard__grid">
        <div className="dashboard__section">
          <h2 className="dashboard__section-title">Ausrüstungsverteilung</h2>
          <div className="distribution-chart">
            {equipmentDistribution.map((item) => (
              <div key={item.status} className="distribution-item">
                <div className="distribution-header">
                  <span className="distribution-label">{item.status}</span>
                  <span className="distribution-count">{item.count}</span>
                </div>
                <div className="distribution-bar">
                  <div
                    className={`distribution-fill distribution-fill--${item.status.toLowerCase()}`}
                    style={{ width: `${item.percentage}%` }}
                  />
                </div>
                <span className="distribution-percentage">{item.percentage}%</span>
              </div>
            ))}
          </div>
        </div>

        <div className="dashboard__section">
          <h2 className="dashboard__section-title">Schnellzugriff</h2>
          <div className="quick-actions">
            <button
              className="quick-action-btn"
              onClick={() => handleQuickAction('equipment')}
            >
              <span className="quick-action-icon">➕</span>
              <span className="quick-action-text">Ausrüstung hinzufügen</span>
            </button>
            <button
              className="quick-action-btn"
              onClick={() => handleQuickAction('project')}
            >
              <span className="quick-action-icon">📋</span>
              <span className="quick-action-text">Neues Projekt</span>
            </button>
            <button
              className="quick-action-btn"
              onClick={() => handleQuickAction('invoice')}
            >
              <span className="quick-action-icon">💳</span>
              <span className="quick-action-text">Rechnung erstellen</span>
            </button>
            <button
              className="quick-action-btn"
              onClick={() => handleQuickAction('scanner')}
            >
              <span className="quick-action-icon">📱</span>
              <span className="quick-action-text">Scannen</span>
            </button>
          </div>
        </div>
      </div>

      {activities && activities.length > 0 && (
        <div className="dashboard__section dashboard__section--full">
          <h2 className="dashboard__section-title">Aktuelle Aktivität</h2>
          <div className="activity-feed">
            {activities.slice(0, 5).map((activity) => (
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
        </div>
      )}
    </div>
  )
}

export default DashboardPage
