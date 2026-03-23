import { AlertTriangle, WifiOff, ShieldX, FileQuestion } from 'lucide-react'
import styles from './ErrorState.module.scss'

export type ErrorVariant = 'generic' | 'network' | 'not-found' | 'permission'

interface ErrorStateProps {
  variant?: ErrorVariant
  title?: string
  description?: string
  onRetry?: () => void
  retryLabel?: string
  compact?: boolean
}

const DEFAULTS: Record<ErrorVariant, { icon: typeof AlertTriangle; title: string; description: string }> = {
  generic: {
    icon: AlertTriangle,
    title: 'Ein Fehler ist aufgetreten',
    description: 'Etwas ist schiefgelaufen. Bitte versuchen Sie es erneut.',
  },
  network: {
    icon: WifiOff,
    title: 'Keine Verbindung',
    description: 'Der Server konnte nicht erreicht werden. Bitte pruefen Sie Ihre Internetverbindung.',
  },
  'not-found': {
    icon: FileQuestion,
    title: 'Nicht gefunden',
    description: 'Die angeforderte Ressource konnte nicht gefunden werden.',
  },
  permission: {
    icon: ShieldX,
    title: 'Zugriff verweigert',
    description: 'Sie haben keine Berechtigung, diese Seite anzuzeigen.',
  },
}

export default function ErrorState({
  variant = 'generic',
  title,
  description,
  onRetry,
  retryLabel = 'Erneut versuchen',
  compact = false,
}: ErrorStateProps) {
  const defaults = DEFAULTS[variant]
  const Icon = defaults.icon

  return (
    <div className={`${styles.container} ${compact ? styles.compact : ''}`}>
      <div className={styles.card}>
        <div className={styles.iconWrap}>
          <Icon size={28} strokeWidth={1.5} />
        </div>
        <h3 className={styles.title}>{title ?? defaults.title}</h3>
        <p className={styles.description}>{description ?? defaults.description}</p>
        {onRetry && (
          <button className={styles.retryBtn} onClick={onRetry}>
            {retryLabel}
          </button>
        )}
      </div>
    </div>
  )
}
