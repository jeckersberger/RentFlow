import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { api, projectApi } from '../services/api'
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

interface UpcomingProject {
  id: string
  name: string
  start_date: string
  end_date: string
  status: string
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

  const { data: upcomingProjects } = useQuery<UpcomingProject[]>({
    queryKey: ['dashboard-upcoming-projects'],
    queryFn: async () => {
      const now = new Date()
      const end = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000) // 30 days from now
      return projectApi.getCalendar(now.toISOString().split('T')[0], end.toISOString().split('T')[0])
        .then(res => (res.data || []).slice(0, 5))
        .catch(() => [])
    },
    staleTime: 1000 * 60 * 10,
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

        <div className="dashboard__section dashboard__section--full">
          <h2 className="dashboard__section-title">Nächste Projekte (30 Tage)</h2>
          {!upcomingProjects || upcomingProjects.length === 0 ? (
            <p style={{ margin: 0, color: 'var(--color-text-secondary)' }}>
              Keine bevorstehenden Projekte geplant.
            </p>
          ) : (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 'var(--spacing-3)' }}>
              {upcomingProjects.map((project) => (
                <div
                  key={project.id}
                  style={{
                    padding: 'var(--spacing-3)',
                    borderLeft: '4px solid var(--color-primary)',
                    backgroundColor: 'var(--color-bg-secondary)',
                    borderRadius: 'var(--radius-md)',
                    cursor: 'pointer',
                    transition: 'all var(--transition-fast)',
                  }}
                  onClick={() => navigate(`/projects/${project.id}`)}
                  onMouseEnter={(e) => (e.currentTarget.style.transform = 'translateY(-2px)')}
                  onMouseLeave={(e) => (e.currentTarget.style.transform = 'translateY(0)')}
                >
                  <p style={{ margin: '0 0 var(--spacing-1) 0', fontWeight: 'var(--font-weight-semibold)', color: 'var(--color-text-primary)' }}>
                    {project.name}
                  </p>
                  <p style={{ margin: '0 0 var(--spacing-2) 0', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                    {new Date(project.start_date).toLocaleDateString('de-DE')}
                  </p>
                  <span style={{ fontSize: 'var(--font-size-xs)', padding: 'var(--spacing-1) var(--spacing-2)', backgroundColor: 'var(--color-primary-100)', color: 'var(--color-primary-900)', borderRadius: 'var(--radius-base)' }}>
                    {project.status}
                  </span>
                </div>
              ))}
            </div>
          )}
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
