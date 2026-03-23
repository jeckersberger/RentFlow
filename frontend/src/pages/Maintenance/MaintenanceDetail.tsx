import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import './Maintenance.scss'

interface CheckItem {
  id: string
  description: string
  completed: boolean
}

interface MaintenanceTaskDetail {
  id: string
  equipment_id: string
  equipment_name: string
  plan_id: string
  plan_name: string
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled'
  priority: 'low' | 'medium' | 'high' | 'critical'
  due_date: string
  completed_date?: string
  created_date: string
  notes: string
  checklist: CheckItem[]
  echeck_result?: {
    date: string
    status: 'pass' | 'warning' | 'fail'
    measurements: Array<{ name: string; value: string; status: 'ok' | 'warning' | 'fail' }>
  }
}

// Mock data
const mockTaskDetail: MaintenanceTaskDetail = {
  id: '1',
  equipment_id: '1',
  equipment_name: 'JBL VTX A12',
  plan_id: '1',
  plan_name: 'Vierteljährliche Kontrolle',
  status: 'pending',
  priority: 'critical',
  due_date: '2026-03-15',
  created_date: '2026-03-01T10:00:00Z',
  notes: 'Überprüfung nach dem Einsatz beim Stadtfest München erforderlich. Überprüfen Sie auf Verschleißerscheinungen und führen Sie Kalibrierungstests durch.',
  checklist: [
    { id: '1', description: 'Visuelle Überprüfung auf Verschäden', completed: true },
    { id: '2', description: 'Schalldruckpegel messen (97 dB erwartete', completed: true },
    { id: '3', description: 'Frequenzgang überprüfen', completed: false },
    { id: '4', description: 'Stromversorgung und Kabel testen', completed: false },
    { id: '5', description: 'Dokumentation und Bericht schreiben', completed: false },
  ],
  echeck_result: {
    date: '2026-03-14T14:30:00Z',
    status: 'pass',
    measurements: [
      { name: 'Impedanz', value: '8 Ω', status: 'ok' },
      { name: 'Schalldruckpegel', value: '96 dB', status: 'ok' },
      { name: 'Frequenzgang', value: '50-20000 Hz', status: 'ok' },
      { name: 'Gleichstromwiderstand', value: '5.8 Ω', status: 'ok' },
    ],
  },
}

function MaintenanceDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [checkedItems, setCheckedItems] = useState<Set<string>>(
    new Set(mockTaskDetail.checklist.filter(item => item.completed).map(item => item.id))
  )

  const { data: task = mockTaskDetail, isLoading } = useQuery({
    queryKey: ['maintenance-task', id],
    queryFn: async () => mockTaskDetail,
    enabled: !!id,
  })

  if (isLoading) {
    return <div className="maintenance-detail-page">Lädt...</div>
  }

  const toggleCheckItem = (itemId: string) => {
    const newChecked = new Set(checkedItems)
    if (newChecked.has(itemId)) {
      newChecked.delete(itemId)
    } else {
      newChecked.add(itemId)
    }
    setCheckedItems(newChecked)
  }

  const completedCount = checkedItems.size
  const totalCount = task.checklist.length
  const progressPercentage = (completedCount / totalCount) * 100

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'critical':
        return '#EF4444'
      case 'high':
        return '#F59E0B'
      case 'medium':
        return '#00D4FF'
      case 'low':
        return '#6B7280'
      default:
        return '#6B7280'
    }
  }

  return (
    <div className="maintenance-detail-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{task.equipment_name}</h1>
          <p className="page-subtitle">{task.plan_name}</p>
        </div>
        <button
          className="btn btn--secondary"
          onClick={() => navigate('/maintenance')}
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
        >
          ← Zurück
        </button>
      </div>

      <div className="detail-grid">
        {/* Main Content */}
        <div>
          {/* Task Info */}
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Aufgaben Information</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <div className="detail-card__row-label">Ausrüstung</div>
                <div className="detail-card__row-value">{task.equipment_name}</div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Plan</div>
                <div className="detail-card__row-value">{task.plan_name}</div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Status</div>
                <div className="detail-card__row-value">
                  <StatusBadge status={task.status} />
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Priorität</div>
                <div className="detail-card__row-value">
                  <span
                    style={{
                      padding: '4px 12px',
                      borderRadius: '999px',
                      fontSize: 'var(--font-size-xs)',
                      fontWeight: 'var(--font-weight-semibold)',
                      backgroundColor: `${getPriorityColor(task.priority)}20`,
                      color: getPriorityColor(task.priority),
                    }}
                  >
                    {task.priority.charAt(0).toUpperCase() + task.priority.slice(1)}
                  </span>
                </div>
              </div>
              <div className="detail-card__row">
                <div className="detail-card__row-label">Fällig am</div>
                <div className="detail-card__row-value">
                  {new Date(task.due_date).toLocaleDateString('de-DE')}
                </div>
              </div>
              {task.completed_date && (
                <div className="detail-card__row">
                  <div className="detail-card__row-label">Abgeschlossen</div>
                  <div className="detail-card__row-value">
                    {new Date(task.completed_date).toLocaleDateString('de-DE')}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Notes */}
          {task.notes && (
            <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
              <h2 className="detail-card__title">Notizen</h2>
              <p style={{ margin: 0, color: 'var(--color-text-primary)', fontSize: 'var(--font-size-sm)', lineHeight: '1.6' }}>
                {task.notes}
              </p>
            </div>
          )}

          {/* Checklist */}
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">
              Checkliste ({completedCount}/{totalCount})
            </h2>

            {/* Progress Bar */}
            <div style={{ marginBottom: 'var(--spacing-4)' }}>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: 'var(--font-size-xs)',
                color: 'var(--color-text-secondary)',
                marginBottom: 'var(--spacing-2)',
              }}>
                <span>Fortschritt</span>
                <span>{Math.round(progressPercentage)}%</span>
              </div>
              <div style={{
                height: '8px',
                background: 'var(--color-bg-tertiary)',
                borderRadius: 'var(--radius-full)',
                overflow: 'hidden',
              }}>
                <div
                  style={{
                    height: '100%',
                    width: `${progressPercentage}%`,
                    background: progressPercentage === 100
                      ? 'linear-gradient(90deg, var(--color-success), var(--color-success-light))'
                      : progressPercentage > 70
                        ? 'linear-gradient(90deg, var(--color-warning), var(--color-warning-light))'
                        : 'linear-gradient(90deg, var(--color-primary), var(--color-primary-light))',
                    transition: 'width var(--transition-normal)',
                  }}
                />
              </div>
            </div>

            {/* Checklist Items */}
            <div className="checklist">
              {task.checklist.map((item) => (
                <div key={item.id} className="checklist-item">
                  <input
                    type="checkbox"
                    checked={checkedItems.has(item.id)}
                    onChange={() => toggleCheckItem(item.id)}
                    style={{
                      width: '18px',
                      height: '18px',
                      cursor: 'pointer',
                      margin: 0,
                      marginRight: 'var(--spacing-2)',
                    }}
                  />
                  <label
                    style={{
                      flex: 1,
                      cursor: 'pointer',
                      fontSize: 'var(--font-size-sm)',
                      color: checkedItems.has(item.id) ? 'var(--color-text-secondary)' : 'var(--color-text-primary)',
                      textDecoration: checkedItems.has(item.id) ? 'line-through' : 'none',
                    }}
                    onClick={() => toggleCheckItem(item.id)}
                  >
                    {item.description}
                  </label>
                </div>
              ))}
            </div>
          </div>

          {/* E-Check Results */}
          {task.echeck_result && (
            <div className="detail-card">
              <h2 className="detail-card__title">
                E-Check Ergebnisse{' '}
                <span>
                  <StatusBadge status={task.echeck_result.status} />
                </span>
              </h2>
              <div style={{ marginBottom: 'var(--spacing-3)', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                {new Date(task.echeck_result.date).toLocaleString('de-DE')}
              </div>
              <div className="echeck-results">
                {task.echeck_result.measurements.map((measurement, idx) => (
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
            </div>
          )}
        </div>

        {/* Sidebar */}
        <div>
          {/* Summary */}
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Zusammenfassung</h2>
            <div className="detail-card__content">
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr',
                  gap: 'var(--spacing-3)',
                }}
              >
                <div style={{
                  padding: 'var(--spacing-3)',
                  background: 'var(--color-bg-tertiary)',
                  borderRadius: 'var(--radius-md)',
                  textAlign: 'center',
                }}>
                  <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-1)' }}>
                    Checkliste
                  </div>
                  <div style={{ fontSize: 'var(--font-size-2xl)', fontWeight: 'var(--font-weight-bold)', color: 'var(--color-text-primary)' }}>
                    {completedCount}/{totalCount}
                  </div>
                </div>
                <div style={{
                  padding: 'var(--spacing-3)',
                  background: 'var(--color-bg-tertiary)',
                  borderRadius: 'var(--radius-md)',
                  textAlign: 'center',
                }}>
                  <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-1)' }}>
                    Fortschritt
                  </div>
                  <div style={{
                    fontSize: 'var(--font-size-2xl)',
                    fontWeight: 'var(--font-weight-bold)',
                    color: progressPercentage === 100 ? 'var(--color-success)' : 'var(--color-text-primary)',
                  }}>
                    {Math.round(progressPercentage)}%
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Status Timeline */}
          <div className="detail-card" style={{ marginBottom: 'var(--spacing-6)' }}>
            <h2 className="detail-card__title">Status</h2>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
              {['pending', 'in_progress', 'completed'].map((status, idx) => (
                <div key={status} style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--spacing-2)' }}>
                  <div
                    style={{
                      width: '24px',
                      height: '24px',
                      borderRadius: '50%',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: 'var(--font-size-xs)',
                      fontWeight: 'var(--font-weight-bold)',
                      background: task.status === status ? 'var(--color-primary)' : task.status === 'completed' ? 'var(--color-success)' : 'var(--color-bg-tertiary)',
                      color: (task.status === status || task.status === 'completed') ? 'white' : 'var(--color-text-secondary)',
                      border: `2px solid ${task.status === status ? 'var(--color-primary)' : task.status === 'completed' ? 'var(--color-success)' : 'var(--color-border)'}`,
                      flexShrink: 0,
                      marginTop: '2px',
                    }}
                  >
                    {task.status === 'completed' ? '✓' : idx + 1}
                  </div>
                  <div>
                    <div style={{ fontSize: 'var(--font-size-sm)', fontWeight: 'var(--font-weight-medium)', color: 'var(--color-text-primary)', margin: 0 }}>
                      {status.charAt(0).toUpperCase() + status.slice(1).replace('_', ' ')}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="detail-card">
            <h2 className="detail-card__title">Aktionen</h2>
            <div className="action-buttons">
              {task.status === 'pending' && (
                <>
                  <button className="btn btn--primary">▶ Starten</button>
                  <button className="btn btn--secondary">✏️ Bearbeiten</button>
                  <button className="btn btn--secondary">❌ Abbrechen</button>
                </>
              )}
              {task.status === 'in_progress' && (
                <>
                  <button className="btn btn--primary">✓ Abschließen</button>
                  <button className="btn btn--secondary">⏸️ Pausieren</button>
                </>
              )}
              {task.status === 'completed' && (
                <>
                  <button className="btn btn--secondary">📋 Bericht anzeigen</button>
                  <button className="btn btn--secondary">🖨️ Drucken</button>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default MaintenanceDetailPage
