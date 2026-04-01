import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { maintenanceApi } from '../../services/api'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { SkeletonTable } from '../../components/Skeleton/SkeletonLoader'
import './Maintenance.scss'

function formatDate(dateString: string | undefined | null): string {
  if (!dateString) return 'k.A.'
  const date = new Date(dateString)
  return isNaN(date.getTime())
    ? 'k.A.'
    : date.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

function MaintenancePlanDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()

  const {
    data: plan,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['maintenance-plan', id],
    queryFn: () => maintenanceApi.getPlan(id!),
    enabled: !!id,
  })

  const { data: tasksData, isLoading: tasksLoading } = useQuery({
    queryKey: ['maintenance-tasks-for-plan', id],
    queryFn: () => maintenanceApi.listTasks(),
    enabled: !!id,
  })

  // Filter tasks that belong to this plan
  const allTasks: any[] =
    tasksData?.items || tasksData?.data || (Array.isArray(tasksData) ? tasksData : [])
  const planTasks = allTasks.filter((t: any) => t.plan_id === id)

  if (isLoading) {
    return (
      <div className="maintenance-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Wartungsplan wird geladen...</h1>
          </div>
        </div>
        <SkeletonTable rows={4} columns={3} />
      </div>
    )
  }

  if (error || !plan) {
    return (
      <div className="maintenance-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Wartungsplan</h1>
            <p className="page-subtitle">Fehler beim Laden</p>
          </div>
          <button
            className="btn btn--secondary"
            onClick={() => navigate('/maintenance')}
          >
            Zurueck
          </button>
        </div>
        <div className="empty-state">
          <div className="empty-state__icon">!</div>
          <h3 className="empty-state__title">Plan nicht gefunden</h3>
          <p className="empty-state__description">
            {error ? String(error) : 'Der angeforderte Wartungsplan konnte nicht geladen werden.'}
          </p>
        </div>
      </div>
    )
  }

  const intervalLabel =
    plan.frequency ||
    (plan.interval_days ? `Alle ${plan.interval_days} Tage` : '') ||
    (plan.interval_hours ? `Alle ${plan.interval_hours} Stunden` : '') ||
    plan.plan_type ||
    'k.A.'

  const planType = plan.plan_type || plan.type || 'k.A.'
  const isActive = plan.is_active ?? plan.status === 'active'

  return (
    <div className="maintenance-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">
            {plan.equipment_name || plan.name || 'Wartungsplan'}
          </h1>
          <p className="page-subtitle">
            {plan.name && plan.equipment_name ? plan.name : 'Wartungsplan-Details'}
          </p>
        </div>
        <button
          className="btn btn--secondary"
          onClick={() => navigate('/maintenance')}
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
        >
          Zurueck zur Uebersicht
        </button>
      </div>

      {/* Plan Details */}
      <div className="detail-grid">
        <div>
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Plan-Informationen</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <div className="detail-card__row-label">Geraet</div>
                <div className="detail-card__row-value">
                  {plan.equipment_name || plan.equipment_id || 'k.A.'}
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Intervall</div>
                <div className="detail-card__row-value">{intervalLabel}</div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Typ</div>
                <div className="detail-card__row-value">{planType}</div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Zuletzt</div>
                <div className="detail-card__row-value">
                  {formatDate(plan.last_performed || plan.last_maintenance)}
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Naechste</div>
                <div className="detail-card__row-value">
                  {formatDate(plan.next_due_at || plan.next_due || plan.next_maintenance)}
                </div>
              </div>
              {plan.description && (
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Beschreibung</div>
                  <div className="detail-card__row-value">{plan.description}</div>
                </div>
              )}
            </div>
          </div>

          {/* Associated Tasks */}
          <div className="section-card">
            <h2 className="section-card__title">
              Zugehoerige Aufgaben ({planTasks.length})
            </h2>
            <div className="section-card__content">
              {tasksLoading ? (
                <SkeletonTable rows={3} columns={3} />
              ) : planTasks.length === 0 ? (
                <div className="empty-state" style={{ minHeight: '150px' }}>
                  <div className="empty-state__icon">-</div>
                  <h3 className="empty-state__title">Keine Aufgaben</h3>
                  <p className="empty-state__description">
                    Diesem Wartungsplan sind noch keine Aufgaben zugeordnet.
                  </p>
                </div>
              ) : (
                planTasks.map((task: any) => (
                  <div
                    key={task.id}
                    className="task-item"
                    onClick={() => navigate(`/maintenance/tasks/${task.id}`)}
                  >
                    <div className="task-item__header">
                      <div
                        className={`task-item__checkbox ${
                          task.status === 'completed' ? 'task-item__checkbox--checked' : ''
                        }`}
                      >
                        {task.status === 'completed' && '!'}
                      </div>
                      <div className="task-item__info">
                        <p className="task-item__title">
                          {task.equipment_name || task.title || 'Aufgabe'}
                        </p>
                        <p className="task-item__meta">
                          Faellig: {formatDate(task.due_date || task.scheduled_at)}
                        </p>
                      </div>
                    </div>
                    <div className="task-item__badges">
                      <StatusBadge status={task.status} />
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div>
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Status</h2>
            <div className="detail-card__content">
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 'var(--spacing-2)',
                  padding: 'var(--spacing-3)',
                  background: isActive
                    ? 'rgba(16, 185, 129, 0.1)'
                    : 'rgba(107, 114, 128, 0.1)',
                  borderRadius: 'var(--radius-md)',
                  border: `1px solid ${
                    isActive ? 'rgba(16, 185, 129, 0.2)' : 'rgba(107, 114, 128, 0.2)'
                  }`,
                }}
              >
                <span
                  style={{
                    width: '10px',
                    height: '10px',
                    borderRadius: '50%',
                    background: isActive ? 'var(--color-success)' : 'var(--color-text-secondary)',
                  }}
                />
                <span
                  style={{
                    fontWeight: 'var(--font-weight-semibold)' as any,
                    fontSize: 'var(--font-size-sm)',
                    color: isActive ? 'var(--color-success)' : 'var(--color-text-secondary)',
                  }}
                >
                  {isActive ? 'Aktiv' : 'Pausiert'}
                </span>
              </div>
            </div>
          </div>

          <div className="detail-card">
            <h2 className="detail-card__title">Zusammenfassung</h2>
            <div className="detail-card__content">
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr',
                  gap: 'var(--spacing-3)',
                }}
              >
                <div
                  style={{
                    padding: 'var(--spacing-3)',
                    background: 'var(--color-bg-tertiary)',
                    borderRadius: 'var(--radius-md)',
                    textAlign: 'center',
                  }}
                >
                  <div
                    style={{
                      fontSize: 'var(--font-size-xs)',
                      color: 'var(--color-text-secondary)',
                      marginBottom: 'var(--spacing-1)',
                    }}
                  >
                    Aufgaben
                  </div>
                  <div
                    style={{
                      fontSize: 'var(--font-size-2xl)',
                      fontWeight: 'var(--font-weight-bold)' as any,
                      color: 'var(--color-text-primary)',
                    }}
                  >
                    {planTasks.length}
                  </div>
                </div>
                <div
                  style={{
                    padding: 'var(--spacing-3)',
                    background: 'var(--color-bg-tertiary)',
                    borderRadius: 'var(--radius-md)',
                    textAlign: 'center',
                  }}
                >
                  <div
                    style={{
                      fontSize: 'var(--font-size-xs)',
                      color: 'var(--color-text-secondary)',
                      marginBottom: 'var(--spacing-1)',
                    }}
                  >
                    Erledigt
                  </div>
                  <div
                    style={{
                      fontSize: 'var(--font-size-2xl)',
                      fontWeight: 'var(--font-weight-bold)' as any,
                      color:
                        planTasks.filter((t: any) => t.status === 'completed').length ===
                          planTasks.length && planTasks.length > 0
                          ? 'var(--color-success)'
                          : 'var(--color-text-primary)',
                    }}
                  >
                    {planTasks.filter((t: any) => t.status === 'completed').length}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default MaintenancePlanDetailPage
