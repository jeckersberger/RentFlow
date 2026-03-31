import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { systemApi } from '../../services/api'
import { X, ExternalLink, Download } from 'lucide-react'
import './UpdateBanner.scss'

const DISMISS_KEY_PREFIX = 'rentflow_update_dismissed_'

function UpdateBanner() {
  const [dismissed, setDismissed] = useState(false)
  const [updateStatus, setUpdateStatus] = useState<'idle' | 'updating' | 'success' | 'error'>('idle')
  const [statusMessage, setStatusMessage] = useState('')
  const queryClient = useQueryClient()

  const { data: versionInfo } = useQuery({
    queryKey: ['system-version'],
    queryFn: systemApi.getVersion,
    staleTime: 15 * 60 * 1000, // 15 min cache
    refetchInterval: 30 * 60 * 1000, // Alle 30 min pruefen
    retry: false,
    refetchOnWindowFocus: false,
  })

  const updateMutation = useMutation({
    mutationFn: () => systemApi.triggerUpdate(),
    onSuccess: () => {
      setUpdateStatus('success')
      setStatusMessage('Update wird ausgefuehrt... Seite wird in 90 Sekunden neu geladen.')
      setTimeout(() => {
        queryClient.invalidateQueries({ queryKey: ['system-version'] })
        window.location.reload()
      }, 90000)
    },
    onError: (err: any) => {
      setUpdateStatus('error')
      setStatusMessage(err?.response?.data?.message || 'Update fehlgeschlagen. Bitte manuell aktualisieren.')
    },
  })

  if (!versionInfo?.update_available || dismissed) return null

  const latestVersion = versionInfo.latest_version
  const dismissKey = DISMISS_KEY_PREFIX + latestVersion

  if (latestVersion && localStorage.getItem(dismissKey) === 'true' && updateStatus === 'idle') return null

  const handleDismiss = () => {
    if (latestVersion) {
      localStorage.setItem(dismissKey, 'true')
    }
    setDismissed(true)
  }

  const handleUpdate = () => {
    if (updateStatus === 'updating') return
    setUpdateStatus('updating')
    setStatusMessage('Update wird heruntergeladen und installiert...')
    updateMutation.mutate()
  }

  return (
    <div className="update-banner">
      <div className="update-banner__content">
        {updateStatus === 'idle' ? (
          <>
            <span className="update-banner__text">
              CrateDesk <strong>v{latestVersion}</strong> ist verfuegbar! (aktuell: v{versionInfo.current_version})
            </span>
            <button
              className="update-banner__update-btn"
              onClick={handleUpdate}
            >
              <Download size={14} />
              Jetzt aktualisieren
            </button>
            {versionInfo.release_url && (
              <a
                href={versionInfo.release_url}
                target="_blank"
                rel="noopener noreferrer"
                className="update-banner__link"
              >
                <ExternalLink size={14} />
                Details
              </a>
            )}
          </>
        ) : (
          <span className="update-banner__text" style={{
            color: updateStatus === 'error' ? '#fca5a5' : updateStatus === 'success' ? '#a7f3d0' : undefined
          }}>
            {statusMessage}
          </span>
        )}
      </div>
      {updateStatus === 'idle' && (
        <button
          className="update-banner__dismiss"
          onClick={handleDismiss}
          title="Schliessen"
        >
          <X size={14} />
        </button>
      )}
    </div>
  )
}

export default UpdateBanner
