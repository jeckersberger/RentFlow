import type { ReactNode, ComponentType } from 'react'
import { useNavigate } from 'react-router-dom'
import styles from './EmptyState.module.scss'

interface EmptyStateAction {
  label: string
  href?: string
  onClick?: () => void
}

interface EmptyStateProps {
  title: string
  description?: string
  /** A Lucide icon component (e.g. Package) or any ReactNode */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  icon?: ComponentType<any> | ReactNode
  /** Simple callback action */
  actionLabel?: string
  onAction?: () => void
  /** Richer action with optional href */
  action?: EmptyStateAction
  /** Compact mode — less vertical padding */
  compact?: boolean
}

export default function EmptyState({
  title,
  description,
  icon,
  actionLabel,
  onAction,
  action,
  compact = false,
}: EmptyStateProps) {
  const navigate = useNavigate()

  // Resolve the icon into a ReactNode
  let iconNode: ReactNode = null
  if (icon) {
    if (typeof icon === 'function') {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const IconComponent = icon as ComponentType<any>
      iconNode = <IconComponent size={32} strokeWidth={1.5} />
    } else {
      iconNode = icon
    }
  }

  const handleAction = () => {
    if (action?.onClick) {
      action.onClick()
    } else if (action?.href) {
      navigate(action.href)
    } else if (onAction) {
      onAction()
    }
  }

  const finalLabel = action?.label ?? actionLabel
  const hasAction = !!finalLabel && (!!action?.href || !!action?.onClick || !!onAction)

  return (
    <div className={`${styles.container} ${compact ? styles.compact : ''}`}>
      {iconNode && <div className={styles.icon}>{iconNode}</div>}
      <h3 className={styles.title}>{title}</h3>
      {description && <p className={styles.description}>{description}</p>}
      {hasAction && (
        <button className={styles.action} onClick={handleAction}>
          {finalLabel}
        </button>
      )}
    </div>
  )
}
