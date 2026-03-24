import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { useNotificationStore } from '../../../stores/notificationStore'
import { Download, RefreshCw, HardDrive, Clock, Calendar, Shield } from 'lucide-react'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

interface BackupHistoryEntry {
  id: string
  date: string
  size: string
  type: 'Automatisch' | 'Manuell'
  status: 'Erfolgreich' | 'Fehlgeschlagen' | 'Laeuft'
  download_url?: string
}

function BackupsPage() {
  const queryClient = useQueryClient()
  const addNotification = useNotificationStore((s) => s.addNotification)
  const [saveSuccess, setSaveSuccess] = useState(false)

  const [form, setForm] = useState({
    schedule: 'daily',
    time: '02:00',
    retention_days: '30',
  })

  const { data: scheduleConfig, isLoading: isLoadingSchedule } = useQuery({
    queryKey: ['config', 'backup.schedule'],
    queryFn: () => configApi.get('backup.schedule'),
  })

  const { data: timeConfig, isLoading: isLoadingTime } = useQuery({
    queryKey: ['config', 'backup.time'],
    queryFn: () => configApi.get('backup.time'),
  })

  const { data: retentionConfig, isLoading: isLoadingRetention } = useQuery({
    queryKey: ['config', 'backup.retention_days'],
    queryFn: () => configApi.get('backup.retention_days'),
  })

  const { data: statusConfig, isLoading: isLoadingStatus } = useQuery({
    queryKey: ['config', 'backup.status'],
    queryFn: () => configApi.get('backup.status'),
  })

  const { data: historyConfig, isLoading: isLoadingHistory } = useQuery({
    queryKey: ['config', 'backup.history'],
    queryFn: () => configApi.get('backup.history'),
  })

  const isLoading = isLoadingSchedule || isLoadingTime || isLoadingRetention || isLoadingStatus

  useEffect(() => {
    if (scheduleConfig?.value) setForm((f) => ({ ...f, schedule: scheduleConfig.value }))
    if (timeConfig?.value) setForm((f) => ({ ...f, time: timeConfig.value }))
    if (retentionConfig?.value) setForm((f) => ({ ...f, retention_days: retentionConfig.value }))
  }, [scheduleConfig, timeConfig, retentionConfig])

  const saveMutation = useMutation({
    mutationFn: async () => {
      await Promise.all([
        configApi.set('backup.schedule', { value: form.schedule }),
        configApi.set('backup.time', { value: form.time }),
        configApi.set('backup.retention_days', { value: form.retention_days }),
      ])
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'backup.schedule'] })
      queryClient.invalidateQueries({ queryKey: ['config', 'backup.time'] })
      queryClient.invalidateQueries({ queryKey: ['config', 'backup.retention_days'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
    onError: () => {
      addNotification('Fehler beim Speichern der Backup-Einstellungen', 'error')
    },
  })

  const backupNowMutation = useMutation({
    mutationFn: () => configApi.backupNow(),
    onSuccess: () => {
      addNotification('Backup wurde gestartet', 'success')
      queryClient.invalidateQueries({ queryKey: ['config', 'backup.status'] })
      queryClient.invalidateQueries({ queryKey: ['config', 'backup.history'] })
    },
    onError: () => {
      addNotification('Fehler beim Starten des Backups', 'error')
    },
  })

  const update = (field: string, value: string) =>
    setForm((prev) => ({ ...prev, [field]: value }))

  const scheduleLabel = (value: string) => {
    switch (value) {
      case 'daily': return 'Taeglich'
      case 'weekly': return 'Woechentlich'
      case 'monthly': return 'Monatlich'
      default: return value
    }
  }

  const formatNextBackup = () => {
    const now = new Date()
    const [hours, minutes] = (form.time || '02:00').split(':').map(Number)
    const next = new Date(now)
    next.setHours(hours, minutes, 0, 0)

    if (next <= now) {
      if (form.schedule === 'daily') {
        next.setDate(next.getDate() + 1)
      } else if (form.schedule === 'weekly') {
        next.setDate(next.getDate() + 7)
      } else {
        next.setMonth(next.getMonth() + 1)
      }
    }

    return next.toLocaleString('de-DE', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const backupHistory: BackupHistoryEntry[] = Array.isArray(historyConfig?.value)
    ? historyConfig.value
    : []

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Backups</h1></div>
      <SkeletonCard count={3} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Backups</h1>
        <p>Sichern und wiederherstellen Sie Ihre Daten.</p>
      </div>

      {/* Backup Status Card */}
      <div className="sp-card">
        <h3 className="sp-card__title">
          <HardDrive size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
          Backup-Status
        </h3>
        <div className="sp-grid">
          <div className="sp-field">
            <label className="sp-label">Letztes Backup</label>
            <p style={{ color: 'var(--color-text-primary)', margin: 0, fontSize: 'var(--font-size-sm)' }}>
              {statusConfig?.last_backup || 'Noch kein Backup erstellt'}
            </p>
          </div>
          <div className="sp-field">
            <label className="sp-label">Naechstes geplantes Backup</label>
            <p style={{ color: 'var(--color-text-primary)', margin: 0, fontSize: 'var(--font-size-sm)' }}>
              {formatNextBackup()} ({scheduleLabel(form.schedule)})
            </p>
          </div>
          <div className="sp-field">
            <label className="sp-label">Backup-Groesse</label>
            <p style={{ color: 'var(--color-text-primary)', margin: 0, fontSize: 'var(--font-size-sm)' }}>
              {statusConfig?.size || '~ 0 MB'}
            </p>
          </div>
        </div>
      </div>

      {/* Backup Schedule Configuration */}
      <div className="sp-card">
        <h3 className="sp-card__title">
          <Calendar size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
          Zeitplan-Konfiguration
        </h3>
        <div className="sp-grid">
          <div className="sp-field">
            <label className="sp-label">Zeitplan</label>
            <div style={{ display: 'flex', gap: 'var(--spacing-4)', marginTop: 'var(--spacing-2)' }}>
              {(['daily', 'weekly', 'monthly'] as const).map((opt) => (
                <label
                  key={opt}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 'var(--spacing-2)',
                    cursor: 'pointer',
                    fontSize: 'var(--font-size-sm)',
                    color: 'var(--color-text-primary)',
                  }}
                >
                  <input
                    type="radio"
                    name="backup-schedule"
                    value={opt}
                    checked={form.schedule === opt}
                    onChange={(e) => update('schedule', e.target.value)}
                    style={{ accentColor: 'var(--color-primary)' }}
                  />
                  {scheduleLabel(opt)}
                </label>
              ))}
            </div>
          </div>
          <div className="sp-field">
            <label className="sp-label">Uhrzeit</label>
            <input
              className="sp-input"
              type="time"
              value={form.time}
              onChange={(e) => update('time', e.target.value)}
              style={{ width: 140 }}
            />
          </div>
          <div className="sp-field">
            <label className="sp-label">Aufbewahrung (Tage)</label>
            <input
              className="sp-input"
              type="number"
              min="1"
              max="365"
              value={form.retention_days}
              onChange={(e) => update('retention_days', e.target.value)}
              style={{ width: 120 }}
            />
          </div>
        </div>
      </div>

      <div className="sp-footer">
        {saveSuccess && <span className="sp-msg--success">Gespeichert!</span>}
        {saveMutation.isError && <span className="sp-msg--error">Fehler beim Speichern</span>}
        <button
          className="sp-btn sp-btn--primary"
          onClick={() => saveMutation.mutate()}
          disabled={saveMutation.isPending}
        >
          {saveMutation.isPending ? 'Speichern...' : 'Speichern'}
        </button>
      </div>

      {/* Manual Backup */}
      <div className="sp-card">
        <h3 className="sp-card__title">
          <Shield size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
          Manuelles Backup
        </h3>
        <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: '0 0 var(--spacing-4) 0' }}>
          Erstellen Sie ein manuelles Backup aller Daten. Dies kann einige Minuten dauern.
        </p>
        <button
          className="sp-btn sp-btn--primary"
          onClick={() => backupNowMutation.mutate()}
          disabled={backupNowMutation.isPending}
        >
          {backupNowMutation.isPending ? (
            <>
              <RefreshCw size={16} style={{ marginRight: 6, verticalAlign: 'middle', animation: 'spin 1s linear infinite' }} />
              Backup laeuft...
            </>
          ) : (
            <>
              <Download size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
              Jetzt sichern
            </>
          )}
        </button>
      </div>

      {/* Backup History Table */}
      <div className="sp-card">
        <h3 className="sp-card__title">
          <Clock size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
          Backup-Verlauf
        </h3>
        {isLoadingHistory ? (
          <SkeletonCard count={1} />
        ) : backupHistory.length === 0 ? (
          <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: 0 }}>
            Noch keine Backups vorhanden.
          </p>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 'var(--font-size-sm)' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Datum
                  </th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Groesse
                  </th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Typ
                  </th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Status
                  </th>
                  <th style={{ textAlign: 'right', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Aktion
                  </th>
                </tr>
              </thead>
              <tbody>
                {backupHistory.map((entry) => (
                  <tr key={entry.id} style={{ borderBottom: '1px solid var(--color-border)' }}>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-primary)' }}>
                      {entry.date}
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-primary)' }}>
                      {entry.size}
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-primary)' }}>
                      {entry.type}
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)' }}>
                      <span
                        style={{
                          display: 'inline-block',
                          padding: '2px 8px',
                          borderRadius: 'var(--radius-sm, 4px)',
                          fontSize: 'var(--font-size-xs, 12px)',
                          fontWeight: 500,
                          backgroundColor:
                            entry.status === 'Erfolgreich'
                              ? 'var(--color-success-bg, #dcfce7)'
                              : entry.status === 'Fehlgeschlagen'
                              ? 'var(--color-error-bg, #fef2f2)'
                              : 'var(--color-warning-bg, #fefce8)',
                          color:
                            entry.status === 'Erfolgreich'
                              ? 'var(--color-success, #16a34a)'
                              : entry.status === 'Fehlgeschlagen'
                              ? 'var(--color-error, #dc2626)'
                              : 'var(--color-warning, #ca8a04)',
                        }}
                      >
                        {entry.status}
                      </span>
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', textAlign: 'right' }}>
                      {entry.status === 'Erfolgreich' && entry.download_url && (
                        <a
                          href={entry.download_url}
                          className="sp-btn sp-btn--ghost"
                          style={{ fontSize: 'var(--font-size-xs, 12px)', padding: '2px 8px', textDecoration: 'none' }}
                          download
                        >
                          <Download size={14} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                          Herunterladen
                        </a>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

export default BackupsPage
