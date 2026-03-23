import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { Download, RefreshCw } from 'lucide-react'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.module.scss'

function BackupsPage() {
  const [triggerSuccess, setTriggerSuccess] = useState(false)

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'backups'],
    queryFn: () => configApi.get('backups'),
  })

  const backupMutation = useMutation({
    mutationFn: () => configApi.set('backups', { trigger: true }),
    onSuccess: () => {
      setTriggerSuccess(true)
      setTimeout(() => setTriggerSuccess(false), 3000)
    },
  })

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Backups</h1>
        <p>Sichern und wiederherstellen Sie Ihre Daten.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Letztes Backup</h3>
        {isLoading ? (
          <SkeletonCard count={1} />
        ) : (
          <div className="sp-grid">
            <div className="sp-field">
              <label className="sp-label">Datum</label>
              <p style={{ color: 'var(--color-text-primary)', margin: 0, fontSize: 'var(--font-size-sm)' }}>
                {config?.last_backup || 'Noch kein Backup erstellt'}
              </p>
            </div>
            <div className="sp-field">
              <label className="sp-label">Zeitplan</label>
              <p style={{ color: 'var(--color-text-primary)', margin: 0, fontSize: 'var(--font-size-sm)' }}>
                {config?.schedule || 'Täglich um 03:00 Uhr'}
              </p>
            </div>
          </div>
        )}
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Manuelles Backup</h3>
        <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: '0 0 var(--spacing-4) 0' }}>
          Erstellen Sie ein manuelles Backup aller Daten. Dies kann einige Minuten dauern.
        </p>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button className="sp-btn sp-btn--primary" onClick={() => backupMutation.mutate()} disabled={backupMutation.isPending}>
            <Download size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
            {backupMutation.isPending ? 'Backup läuft...' : 'Backup jetzt erstellen'}
          </button>
          {triggerSuccess && (
            <span style={{ display: 'flex', alignItems: 'center', gap: 4, color: 'var(--color-success)', fontSize: 'var(--font-size-sm)' }}>
              <RefreshCw size={14} /> Backup gestartet!
            </span>
          )}
        </div>
      </div>
    </div>
  )
}

export default BackupsPage
