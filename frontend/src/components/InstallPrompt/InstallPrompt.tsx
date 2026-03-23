import { useEffect, useState, useCallback } from 'react'
import { Download } from 'lucide-react'
import styles from './InstallPrompt.module.scss'

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

const DISMISSED_KEY = 'rentflow-install-dismissed'

function isStandalone(): boolean {
  return (
    window.matchMedia('(display-mode: standalone)').matches ||
    (navigator as unknown as { standalone?: boolean }).standalone === true
  )
}

export default function InstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    // Don't show if already installed or previously dismissed
    if (isStandalone()) return
    if (localStorage.getItem(DISMISSED_KEY)) return

    const handler = (e: Event) => {
      e.preventDefault()
      setDeferredPrompt(e as BeforeInstallPromptEvent)
      setVisible(true)
    }

    window.addEventListener('beforeinstallprompt', handler)

    return () => {
      window.removeEventListener('beforeinstallprompt', handler)
    }
  }, [])

  const handleInstall = useCallback(async () => {
    if (!deferredPrompt) return

    await deferredPrompt.prompt()
    const { outcome } = await deferredPrompt.userChoice

    if (outcome === 'accepted') {
      setVisible(false)
    }
    setDeferredPrompt(null)
  }, [deferredPrompt])

  const handleDismiss = useCallback(() => {
    setVisible(false)
    setDeferredPrompt(null)
    localStorage.setItem(DISMISSED_KEY, 'true')
  }, [])

  if (!visible) return null

  return (
    <div className={styles.banner}>
      <div className={styles.icon}>
        <Download size={20} />
      </div>
      <div className={styles.text}>
        <div className={styles.title}>RentFlow als App installieren</div>
        <div className={styles.subtitle}>Schnellzugriff direkt vom Homescreen</div>
      </div>
      <div className={styles.actions}>
        <button className={styles.laterBtn} onClick={handleDismiss}>
          Sp\u00e4ter
        </button>
        <button className={styles.installBtn} onClick={handleInstall}>
          <Download size={16} />
          Installieren
        </button>
      </div>
    </div>
  )
}
