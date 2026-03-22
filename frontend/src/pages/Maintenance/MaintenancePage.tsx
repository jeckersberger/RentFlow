import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { maintenanceApi } from '../../services/api'
import './Maintenance.module.scss'

interface MaintenanceTask {
  id: string
  equipment_id: string
  equipment_name: string
  plan_id: string
  plan_name: string
  status: 'pending' | 'planned' | 'overdue' | 'in_progress' | 'completed' | 'cancelled'
  priority: 'low' | 'medium' | 'high' | 'critical'
  due_date?: string
  scheduled_at?: string
  completed_date?: string
  completed_at?: string
}

interface ECheckResult {
  id: string
  equipment_id: string
  equipment_name: string
  result_date?: string
  test_date?: string
  status?: 'pass' | 'warning' | 'fail'
  result?: 'passed' | 'conditional' | 'failed'
  measurements?: Array<{ name: string; value: string; status: 'ok' | 'warning' | 'fail' }>
  insulation_resistance_mohm?: number
  protective_conductor_resistance_ohm?: number
  leakage_current_ma?: number
  certificate_number?: string
}

interface MaintenancePlan {
  id: string
  equipment_id: string
  equipment_name: string
  frequency?: string
  plan_type?: string
  interval_days?: number | null
  interval_hours?: number | null
  name?: string
  next_maintenance?: string
  next_due_at?: string
  status?: 'active' | 'completed' | 'paused'
  is_active?: boolean
}

function MaintenancePage() {
  const navigate = useNavigate()
  const [expandedTask, setExpandedTask] = useState<string | null>(null)

  const { data: tasksData, isLoading: tasksLoading, error: tasksError } = useQuery({
    queryKey: ['maintenance-tasks'],
    queryFn: () => maintenanceApi.listTasks(),
    staleTime: 1000 * 60 * 5,
  })

  const { data: eChecksData, isLoading: eChecksLoading } = useQuery({
    queryKey: ['echeck-results'],
    queryFn: () => maintenanceApi.listElectricalTests(),
    staleTime: 1000 * 60 * 5,
  })

  const { data: plansData, isLoading: plansLoading } = useQuery({
    queryKey: ['maintenance-plans'],
    queryFn: () => maintenanceApi.listPlans(),
    staleTime: 1000 * 60 * 5,
  })

  const tasks: MaintenanceTask[] = (tasksData?.items || tasksData?.data || (Array.isArray(tasksData) ? tasksData : [])).map((t: any) => ({
    ...t,
    due_date: t.due_date || t.scheduled_at,
    status: t.status === 'planned' ? 'pending' : t.status,
  }))
  const eChecks: ECheckResult[] = eChecksData?.items || eChecksData?.data || (Array.isArray(eChecksData) ? eChecksData : [])
  const plans: MaintenancePlan[] = (plansData?.items || plansData?.data || (Array.isArray(plansData) ? plansData : [])).map((p: any) => ({
    ...p,
    next_maintenance: p.next_maintenance || p.next_due_at,
    status: p.status || (p.is_active ? 'active' : 'paused'),
    frequency: p.frequency || (p.interval_days ? `Alle ${p.interval_days} Tage` : p.interval_hours ? `Alle ${p.interval_hours}h` : p.plan_type || ''),
  }))
  const isLoading = tasksLoading || eChecksLoading || plansLoading

  const now = new Date()
  const overdueTasks = tasks.filter(t => t.due_date && new Date(t.due_date) < now && (t.status === 'pending' || t.status === 'overdue'))
  const dueTasks = tasks.filter(t => {
    if (!t.due_date) return false
    const dueDate = new Date(t.due_date)
    return dueDate >= now && dueDate <= new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000) && t.status === 'pending'
  })
  const activePlans = plans.filter(p => p.status === 'active' || p.is_active)

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'critical':
        return 'critical'
      case 'high':
        return 'high'
      case 'medium':
        return 'medium'
      case 'low':
        return 'low'
      default:
        return 'medium'
    }
  }

  const formatDate = (dateString: string | undefined) => {
    if (!dateString) return 'N/A'
    const date = new Date(dateString)
    return isNaN(date.getTime()) ? 'N/A' : date.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: '2-digit' })
  }

  if (isLoading) {
    return (
      <div className="maintenance-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Wartung & Inspektion</h1>
            <p className="page-subtitle">Daten werden geladen...</p>
          </div>
        </div>
        <div className="stats-grid">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="stat-card">
              <div className="stat-card__label">Laden...</div>
              <div className="stat-card__value">--</div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (tasksError) {
    return (
      <div className="maintenance-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Wartung & Inspektion</h1>
            <p className="page-subtitle">Fehler beim Laden der Daten</p>
          </div>
        </div>
        <div className="empty-state">
          <div className="empty-state__icon">&#x26A0;</div>
          <h3 className="empty-state__title">Daten konnten nicht geladen werden</h3>
          <p className="empty-state__description">{String(tasksError)}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="maintenance-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Wartung & Inspektion</h1>
          <p className="page-subtitle">Verwaltung von Wartungsplänen und Inspektionen</p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button
            className="btn btn--primary"
            onClick={() => navigate('/maintenance/new-task')}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            ➕ Neue Aufgabe
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => navigate('/maintenance/new-echeck')}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            ✓ E-Check durchführen
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="stats-grid">
        <div className={`stat-card ${activePlans.length > 0 ? '' : 'stat-card--danger'}`}>
          <div className="stat-card__label">Aktive Pläne</div>
          <div className="stat-card__value">{activePlans.length}</div>
        </div>
        <div className={`stat-card ${tasks.filter(t => t.status === 'pending').length === 0 ? '' : 'stat-card--warning'}`}>
          <div className="stat-card__label">Offene Aufgaben</div>
          <div className="stat-card__value">{tasks.filter(t => t.status === 'pending').length}</div>
        </div>
        <div className={`stat-card ${overdueTasks.length === 0 ? '' : 'stat-card--danger'}`}>
          <div className="stat-card__label">Überfällige Aufgaben</div>
          <div className="stat-card__value">{overdueTasks.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">E-Checks diesen Monat</div>
          <div className="stat-card__value">{eChecks.length}</div>
        </div>
      </div>

      {/* Overdue Tasks Alert */}
      {overdueTasks.length > 0 && (
        <div className="alert-section">
          <h3 className="alert-section__title">⚠️ Überfällige Aufgaben</h3>
          <div className="alert-section__items">
            {overdueTasks.map(task => (
              <div key={task.id} className="alert-item">
                <div className="alert-item__content">
                  <p className="alert-item__name">{task.equipment_name}</p>
                  <p className="alert-item__meta">
                    Fällig seit: {formatDate(task.due_date)} • {task.plan_name}
                  </p>
                </div>
                <button
                  className="btn btn--sm btn--primary"
                  onClick={() => navigate(`/maintenance/tasks/${task.id}`)}
                >
                  Öffnen
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Due Tasks Section */}
      {dueTasks.length > 0 && (
        <div className="section-card">
          <h2 className="section-card__title">📋 Nächste Aufgaben (nächste 7 Tage)</h2>
          <div className="section-card__content">
            {dueTasks.map(task => (
              <div
                key={task.id}
                className="task-item"
                onClick={() => navigate(`/maintenance/tasks/${task.id}`)}
              >
                <div className="task-item__header">
                  <div className="task-item__checkbox">
                    {task.status === 'completed' && '✓'}
                  </div>
                  <div className="task-item__info">
                    <p className="task-item__title">{task.equipment_name}</p>
                    <p className="task-item__meta">
                      {task.plan_name} • Fällig: {formatDate(task.due_date)}
                    </p>
                  </div>
                </div>
                <div className="task-item__badges">
                  <span className={`priority-badge priority-badge--${getPriorityColor(task.priority)}`}>
                    {task.priority.toUpperCase()}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* E-Check Results */}
      <div className="section-card">
        <h2 className="section-card__title">✓ Aktuelle E-Check Ergebnisse</h2>
        <div className="section-card__content">
          {eChecks.length === 0 ? (
            <div className="empty-state" style={{ minHeight: '200px' }}>
              <div className="empty-state__icon">📊</div>
              <h3 className="empty-state__title">Keine E-Check Ergebnisse</h3>
              <p className="empty-state__description">Führen Sie einen E-Check durch, um die Ergebnisse hier zu sehen.</p>
            </div>
          ) : (
            eChecks.slice(0, 3).map(eCheck => {
              const displayDate = eCheck.result_date || eCheck.test_date || ''
              const displayStatus = eCheck.status || (eCheck.result === 'passed' ? 'pass' : eCheck.result === 'conditional' ? 'warning' : eCheck.result === 'failed' ? 'fail' : 'pass')
              // Build measurements from API fields if not present
              const measurements = eCheck.measurements || [
                ...(eCheck.insulation_resistance_mohm != null ? [{ name: 'Isolationswiderstand', value: `${eCheck.insulation_resistance_mohm} MOhm`, status: (eCheck.insulation_resistance_mohm >= 1 ? 'ok' : 'warning') as 'ok' | 'warning' | 'fail' }] : []),
                ...(eCheck.protective_conductor_resistance_ohm != null ? [{ name: 'Schutzleiterwiderstand', value: `${eCheck.protective_conductor_resistance_ohm} Ohm`, status: (eCheck.protective_conductor_resistance_ohm <= 0.3 ? 'ok' : 'warning') as 'ok' | 'warning' | 'fail' }] : []),
                ...(eCheck.leakage_current_ma != null ? [{ name: 'Ableitstrom', value: `${eCheck.leakage_current_ma} mA`, status: (eCheck.leakage_current_ma <= 3.5 ? 'ok' : 'fail') as 'ok' | 'warning' | 'fail' }] : []),
              ]

              return (
                <div key={eCheck.id} style={{ paddingBottom: 'var(--spacing-3)', borderBottom: 'var(--card-border-width) solid var(--color-border)', marginBottom: 'var(--spacing-3)' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-2)' }}>
                    <div>
                      <p style={{ margin: 0, fontWeight: 'var(--font-weight-semibold)', color: 'var(--color-text-primary)', fontSize: 'var(--font-size-sm)' }}>
                        {eCheck.equipment_name}
                      </p>
                      {displayDate && (
                        <p style={{ margin: '0', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                          {new Date(displayDate).toLocaleString('de-DE')}
                        </p>
                      )}
                      {eCheck.certificate_number && (
                        <p style={{ margin: '0', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                          Zertifikat: {eCheck.certificate_number}
                        </p>
                      )}
                    </div>
                    <StatusBadge status={displayStatus} />
                  </div>
                  {measurements.length > 0 && (
                    <div className="echeck-results">
                      {measurements.map((measurement, idx) => (
                        <div key={idx} className="echeck-item">
                          <div className="echeck-item__label">{measurement.name}</div>
                          <div className="echeck-item__value">
                            <div className="echeck-item__measurement">{measurement.value}</div>
                            <div className={`echeck-item__status echeck-item__status--${measurement.status}`}>
                              {measurement.status === 'ok' ? '✓ OK' : measurement.status === 'warning' ? '⚠ Warnung' : '✗ Fehler'}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )
            })
          )}
        </div>
      </div>

      {/* Maintenance Plans Overview */}
      <div className="section-card">
        <h2 className="section-card__title">📅 Aktive Wartungspläne</h2>
        <div className="section-card__content">
          {activePlans.length === 0 ? (
            <div className="empty-state" style={{ minHeight: '200px' }}>
              <div className="empty-state__icon">📋</div>
              <h3 className="empty-state__title">Keine Wartungspläne</h3>
              <p className="empty-state__description">Erstellen Sie einen Wartungsplan, um Inspektionen zu planen.</p>
            </div>
          ) : (
            <div className="calendar-list">
              {activePlans
                .filter(p => p.next_maintenance)
                .sort((a, b) => new Date(a.next_maintenance!).getTime() - new Date(b.next_maintenance!).getTime())
                .map(plan => {
                  const nextDate = new Date(plan.next_maintenance!)
                  const daysUntil = Math.ceil((nextDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))

                  return (
                    <div
                      key={plan.id}
                      className="calendar-item"
                      onClick={() => navigate(`/maintenance/plans/${plan.id}`)}
                    >
                      <div className="calendar-item__date">
                        <div className="calendar-item__date-day">{nextDate.getDate()}</div>
                        <div className="calendar-item__date-month">
                          {nextDate.toLocaleDateString('de-DE', { month: 'short' })}
                        </div>
                      </div>
                      <div className="calendar-item__content">
                        <p className="calendar-item__title" style={{ margin: 0, fontWeight: 'var(--font-weight-semibold)', color: 'var(--color-text-primary)' }}>
                          {plan.equipment_name || plan.name}
                        </p>
                        <p className="calendar-item__equipment" style={{ margin: '0 0 var(--spacing-1) 0' }}>
                          {plan.frequency}{plan.name && plan.equipment_name ? ` - ${plan.name}` : ''}
                        </p>
                        <p style={{ margin: 0, fontSize: 'var(--font-size-xs)', color: daysUntil <= 3 ? 'var(--color-danger)' : 'var(--color-text-secondary)' }}>
                          {daysUntil <= 0 ? 'Heute fällig' : daysUntil === 1 ? 'Morgen' : `In ${daysUntil} Tagen`}
                        </p>
                      </div>
                    </div>
                  )
                })}
            </div>
          )}
        </div>
      </div>

      {/* All Tasks */}
      <div className="section-card">
        <h2 className="section-card__title">📊 Alle Aufgaben ({tasks.length})</h2>
        <div className="section-card__content">
          {tasks.length === 0 ? (
            <div className="empty-state" style={{ minHeight: '200px' }}>
              <div className="empty-state__icon">✓</div>
              <h3 className="empty-state__title">Keine Aufgaben</h3>
              <p className="empty-state__description">Alle Wartungsaufgaben sind abgeschlossen.</p>
            </div>
          ) : (
            tasks.map(task => (
              <div
                key={task.id}
                className="task-item"
                onClick={() => setExpandedTask(expandedTask === task.id ? null : task.id)}
              >
                <div className="task-item__header">
                  <div className={`task-item__checkbox ${task.status === 'completed' ? 'task-item__checkbox--checked' : ''}`}>
                    {task.status === 'completed' && '✓'}
                  </div>
                  <div className="task-item__info">
                    <p className="task-item__title">{task.equipment_name}</p>
                    <p className="task-item__meta">
                      {task.plan_name} • Fällig: {formatDate(task.due_date)}
                    </p>
                  </div>
                </div>
                <div className="task-item__badges">
                  <span className={`priority-badge priority-badge--${getPriorityColor(task.priority)}`}>
                    {task.priority.charAt(0).toUpperCase() + task.priority.slice(1)}
                  </span>
                  <StatusBadge status={task.status} />
                  <span style={{ marginLeft: 'var(--spacing-2)' }}>
                    {expandedTask === task.id ? '▼' : '▶'}
                  </span>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}

export default MaintenancePage
