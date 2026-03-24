import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { systemApi } from '../../services/api'
import { X, ExternalLink } from 'lucide-react'
import './UpdateBanner.scss'

const DISMISS_KEY_PREFIX = 'rentflow_update_dismissed_'

function UpdateBanner() {
  const [dismissed, setDismissed] = useState(false)

  const { data: versionInfo } = useQuery({
    queryKey: ['system-version'],
    queryFn: systemApi.getVersion,
    staleTime: 6 * 60 * 60 * 1000, // 6 hours
    retry: false,
    refetchOnWindowFocus: false,
  })

  if (!versionInfo?.update_available || dismissed) return null

  const latestVersion = versionInfo.latest_version
  const dismissKey = DISMISS_KEY_PREFIX + latestVersion

  // Check localStorage for previously dismissed version
  if (latestVersion && localStorage.getItem(dismissKey) === 'true') return null

  const handleDismiss = () => {
    if (latestVersion) {
      localStorage.setItem(dismissKey, 'true')
    }
    setDismissed(true)
  }

  return (
    <div className="update-banner">
      <div className="update-banner__content">
        <span className="update-banner__text">
          RentFlow v{latestVersion} ist verfügbar! Sie nutzen v{versionInfo.current_version}.
        </span>
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
      </div>
      <button
        className="update-banner__dismiss"
        onClick={handleDismiss}
        title="Schließen"
      >
        <X size={14} />
      </button>
    </div>
  )
}

export default UpdateBanner
