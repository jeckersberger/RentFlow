import { useEffect, useState, useRef } from 'react'
import { WifiOff, Wifi } from 'lucide-react'
import styles from './OfflineIndicator.module.scss'

type Status = 'online' | 'offline' | 'reconnected' | null

export default function OfflineIndicator() {
  const [status, setStatus] = useState<Status>(null)
  const [exiting, setExiting] = useState(false)
  const wasOffline = useRef(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    // Initialize — only show if already offline at mount
    if (!navigator.onLine) {
      setStatus('offline')
      wasOffline.current = true
    }

    const handleOffline = () => {
      wasOffline.current = true
      setExiting(false)
      setStatus('offline')
      if (timerRef.current) clearTimeout(timerRef.current)
    }

    const handleOnline = () => {
      if (wasOffline.current) {
        wasOffline.current = false
        setExiting(false)
        setStatus('reconnected')

        if (timerRef.current) clearTimeout(timerRef.current)
        timerRef.current = setTimeout(() => {
          setExiting(true)
          setTimeout(() => {
            setStatus(null)
            setExiting(false)
          }, 300)
        }, 3000)
      }
    }

    window.addEventListener('offline', handleOffline)
    window.addEventListener('online', handleOnline)

    return () => {
      window.removeEventListener('offline', handleOffline)
      window.removeEventListener('online', handleOnline)
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [])

  if (!status) return null

  const isOffline = status === 'offline'

  return (
    <div
      className={`${styles.banner} ${isOffline ? styles.offline : styles.online} ${exiting ? styles.exiting : ''}`}
    >
      <span className={styles.icon}>
        {isOffline ? <WifiOff size={16} /> : <Wifi size={16} />}
      </span>
      {isOffline
        ? 'Sie sind offline. Einige Funktionen sind eingeschr\u00e4nkt.'
        : 'Verbindung wiederhergestellt'}
    </div>
  )
}
