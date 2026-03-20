import { useQuery } from '@tanstack/react-query'
import { api } from '../services/api'
import './Dashboard.scss'

interface DashboardStats {
  total_equipment: number
  active_projects: number
  pending_invoices: number
  crew_members: number
}

function DashboardPage() {
  const { data: stats, isLoading, error } = useQuery<DashboardStats>({
    queryKey: ['dashboard-stats'],
    queryFn: () => api.get('/api/v1/dashboard/stats').then(res => res.data),
  })

  return (
    <div className="dashboard">
      <h1 className="dashboard__title">Dashboard</h1>

      {error && (
        <div className="dashboard__error">
          <p>Failed to load dashboard data</p>
        </div>
      )}

      <div className="dashboard__grid">
        <div className="stat-card">
          <div className="stat-card__content">
            <p className="stat-card__label">Total Equipment</p>
            <p className="stat-card__value">
              {isLoading ? '—' : stats?.total_equipment || 0}
            </p>
          </div>
          <div className="stat-card__icon">📦</div>
        </div>

        <div className="stat-card">
          <div className="stat-card__content">
            <p className="stat-card__label">Active Projects</p>
            <p className="stat-card__value">
              {isLoading ? '—' : stats?.active_projects || 0}
            </p>
          </div>
          <div className="stat-card__icon">📋</div>
        </div>

        <div className="stat-card">
          <div className="stat-card__content">
            <p className="stat-card__label">Pending Invoices</p>
            <p className="stat-card__value">
              {isLoading ? '—' : stats?.pending_invoices || 0}
            </p>
          </div>
          <div className="stat-card__icon">💰</div>
        </div>

        <div className="stat-card">
          <div className="stat-card__content">
            <p className="stat-card__label">Crew Members</p>
            <p className="stat-card__value">
              {isLoading ? '—' : stats?.crew_members || 0}
            </p>
          </div>
          <div className="stat-card__icon">👥</div>
        </div>
      </div>

      <div className="dashboard__section">
        <h2 className="dashboard__section-title">Recent Activity</h2>
        <p className="dashboard__section-text">Activity feed coming soon...</p>
      </div>

      <div className="dashboard__section">
        <h2 className="dashboard__section-title">Quick Actions</h2>
        <div className="dashboard__actions">
          <button className="action-button">+ New Project</button>
          <button className="action-button">+ Add Equipment</button>
          <button className="action-button">+ Create Invoice</button>
          <button className="action-button">+ Schedule Crew</button>
        </div>
      </div>
    </div>
  )
}

export default DashboardPage
