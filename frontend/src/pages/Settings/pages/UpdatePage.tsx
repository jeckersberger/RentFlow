import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { systemApi } from '../../../services/api'
import { useNotificationStore } from '../../../stores/notificationStore'
import { Download, RefreshCw, CheckCircle, AlertTriangle, Info } from 'lucide-react'
import '../Settings.scss'

function UpdatePage() {
  const { addNotification } = useNotificationStore()
  const queryClient = useQueryClient()
  const [updateLog, setUpdateLog] = useState<string | null>(null)

  const [checkMessage, setCheckMessage] = useState<string | null>(null)

  const { data: versionInfo, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['system-version'],
    queryFn: () => systemApi.getVersion(),
    staleTime: 0,
  })

  const handleCheckUpdate = async () => {
    setCheckMessage(null)
    try {
      const result = await refetch()
      const data = result.data
      if (data?.update_available) {
        setCheckMessage(`Update auf v${data.latest_version} verfuegbar!`)
      } else {
        setCheckMessage('Ihre Version ist aktuell. Kein Update verfuegbar.')
      }
      // Meldung nach 8 Sekunden ausblenden
      setTimeout(() => setCheckMessage(null), 8000)
    } catch {
      setCheckMessage('Fehler beim Pruefen auf Updates.')
      setTimeout(() => setCheckMessage(null), 5000)
    }
  }

  const updateMutation = useMutation({
    mutationFn: () => systemApi.triggerUpdate(),
    onSuccess: () => {
      addNotification('Update gestartet. Die Anwendung wird in ca. 2 Minuten neu gestartet.', 'success', { title: 'Update laeuft' })
      setUpdateLog('Update wird ausgefuehrt... Bitte warten.')
      setTimeout(() => {
        queryClient.invalidateQueries({ queryKey: ['system-version'] })
        window.location.reload()
      }, 90000)
    },
    onError: (err: any) => {
      addNotification(err?.response?.data?.message || 'Update fehlgeschlagen', 'error', { title: 'Fehler' })
    },
  })

  return (
    <div className="sp-page">
      <div className="sp-header">
        <div>
          <h2 className="sp-title">Software-Update</h2>
          <p className="sp-subtitle">EquipFlow Version und Updates verwalten</p>
        </div>
      </div>

      {/* Aktuelle Version */}
      <div className="sp-card" style={{ marginBottom: '1.5rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1rem' }}>
          <Info size={20} style={{ color: 'var(--color-primary)' }} />
          <h3 className="sp-card-title" style={{ margin: 0 }}>Aktuelle Version</h3>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem' }}>
          <div>
            <div style={{ fontSize: '0.8rem', color: 'var(--color-text-muted)', marginBottom: '0.25rem' }}>Installierte Version</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--color-text-primary)' }}>
              v{versionInfo?.current_version || '...'}
            </div>
          </div>
          <div>
            <div style={{ fontSize: '0.8rem', color: 'var(--color-text-muted)', marginBottom: '0.25rem' }}>Neueste Version</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 700, color: versionInfo?.update_available ? '#f59e0b' : '#10b981' }}>
              {versionInfo?.latest_version ? `v${versionInfo.latest_version}` : (isLoading ? '...' : 'Unbekannt')}
            </div>
          </div>
          <div>
            <div style={{ fontSize: '0.8rem', color: 'var(--color-text-muted)', marginBottom: '0.25rem' }}>Status</div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '1rem', fontWeight: 600 }}>
              {versionInfo?.update_available ? (
                <>
                  <AlertTriangle size={18} style={{ color: '#f59e0b' }} />
                  <span style={{ color: '#f59e0b' }}>Update verfuegbar</span>
                </>
              ) : (
                <>
                  <CheckCircle size={18} style={{ color: '#10b981' }} />
                  <span style={{ color: '#10b981' }}>Aktuell</span>
                </>
              )}
            </div>
          </div>
        </div>
        {versionInfo?.checked_at && (
          <div style={{ marginTop: '0.75rem', fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
            Zuletzt geprueft: {new Date(versionInfo.checked_at).toLocaleString('de-DE')}
          </div>
        )}
      </div>

      {/* Aktionen */}
      <div className="sp-card" style={{ marginBottom: '1.5rem' }}>
        <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
          <button
            className="btn btn--secondary"
            onClick={handleCheckUpdate}
            disabled={isFetching}
            style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}
          >
            <RefreshCw size={16} className={isFetching ? 'spin' : ''} />
            {isFetching ? 'Pruefe...' : 'Auf Updates pruefen'}
          </button>
          {checkMessage && (
            <span style={{
              fontSize: '0.85rem',
              fontWeight: 500,
              color: checkMessage.includes('verfuegbar') ? '#f59e0b' : checkMessage.includes('aktuell') ? '#10b981' : '#ef4444',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
            }}>
              {checkMessage.includes('aktuell') ? <CheckCircle size={16} /> : checkMessage.includes('verfuegbar') ? <AlertTriangle size={16} /> : null}
              {checkMessage}
            </span>
          )}

          {versionInfo?.update_available && (
            <button
              className="btn btn--primary"
              onClick={() => updateMutation.mutate()}
              disabled={updateMutation.isPending}
              style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}
            >
              <Download size={16} />
              {updateMutation.isPending ? 'Update laeuft...' : `Auf v${versionInfo.latest_version} aktualisieren`}
            </button>
          )}

          {versionInfo?.release_url && (
            <a
              href={versionInfo.release_url}
              target="_blank"
              rel="noopener noreferrer"
              className="btn btn--secondary"
              style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', textDecoration: 'none' }}
            >
              Release Notes ansehen
            </a>
          )}
        </div>
      </div>

      {/* Release Notes */}
      {versionInfo?.release_notes && versionInfo.update_available && (
        <div className="sp-card" style={{ marginBottom: '1.5rem' }}>
          <h3 className="sp-card-title">Release Notes (v{versionInfo.latest_version})</h3>
          <div style={{
            padding: '1rem',
            background: 'rgba(255,255,255,0.03)',
            borderRadius: '6px',
            fontSize: '0.85rem',
            whiteSpace: 'pre-wrap',
            lineHeight: 1.6,
            color: 'var(--color-text-secondary)',
          }}>
            {versionInfo.release_notes}
          </div>
        </div>
      )}

      {/* Update Log */}
      {updateLog && (
        <div className="sp-card">
          <h3 className="sp-card-title">Update-Log</h3>
          <div style={{
            padding: '1rem',
            background: '#0a0f1a',
            borderRadius: '6px',
            fontFamily: 'monospace',
            fontSize: '0.8rem',
            color: '#10b981',
            whiteSpace: 'pre-wrap',
          }}>
            {updateLog}
          </div>
        </div>
      )}

      {/* Info */}
      <div style={{
        marginTop: '1.5rem',
        padding: '1rem',
        background: 'rgba(0, 212, 255, 0.06)',
        borderRadius: '8px',
        border: '1px solid rgba(0, 212, 255, 0.15)',
        fontSize: '0.8rem',
        color: 'var(--color-text-secondary)',
      }}>
        <strong style={{ color: 'var(--color-primary)' }}>Hinweis zum Update-Prozess:</strong>
        <ul style={{ margin: '0.5rem 0 0 1rem', padding: 0 }}>
          <li>Vor jedem Update wird automatisch ein Datenbank-Backup erstellt</li>
          <li>Das Update dauert ca. 2-3 Minuten (Download, Build, Neustart)</li>
          <li>Waehrend des Updates ist die Anwendung kurzzeitig nicht erreichbar</li>
          <li>Versionierung: X.Y.Z (X=grosse Aenderungen, Y=neue Features, Z=Bugfixes)</li>
        </ul>
      </div>
    </div>
  )
}

export default UpdatePage
