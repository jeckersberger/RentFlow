import { useEffect, useState, useCallback } from 'react'
import { CheckCircle, XCircle, AlertTriangle, Info, X } from 'lucide-react'
import { useNotificationStore, type Notification } from '../../stores/notificationStore'
import styles from './Toast.module.scss'

interface ToastItemProps {
  notification: Notification
  onClose: () => void
}

function ToastItem({ notification, onClose }: ToastItemProps) {
  const [exiting, setExiting] = useState(false)
  const [paused, setPaused] = useState(false)
  const [progress, setProgress] = useState(100)

  const duration = notification.duration

  const handleClose = useCallback(() => {
    setExiting(true)
    setTimeout(onClose, 300)
  }, [onClose])

  useEffect(() => {
    if (duration === null || duration === undefined || duration === 0) return

    const interval = 50
    let elapsed = 0

    const timer = setInterval(() => {
      if (paused) return
      elapsed += interval
      const remaining = Math.max(0, 100 - (elapsed / duration) * 100)
      setProgress(remaining)

      if (elapsed >= duration) {
        clearInterval(timer)
        handleClose()
      }
    }, interval)

    return () => clearInterval(timer)
  }, [duration, paused, handleClose])

  const getIcon = () => {
    const size = 20
    switch (notification.type) {
      case 'success':
        return <CheckCircle size={size} />
      case 'error':
        return <XCircle size={size} />
      case 'warning':
        return <AlertTriangle size={size} />
      case 'info':
        return <Info size={size} />
      default:
        return <Info size={size} />
    }
  }

  return (
    <div
      className={`${styles.toast} ${styles[`toast--${notification.type}`]} ${exiting ? styles['toast--exiting'] : ''}`}
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
    >
      <div className={styles.toast__icon}>{getIcon()}</div>
      <div className={styles.toast__content}>
        {notification.title && (
          <div className={styles.toast__title}>{notification.title}</div>
        )}
        <div className={styles.toast__message}>{notification.message}</div>
        {notification.action && (
          <button
            className={styles.toast__action}
            onClick={() => {
              notification.action?.onClick()
              handleClose()
            }}
          >
            {notification.action.label}
          </button>
        )}
      </div>
      <button
        className={styles.toast__close}
        onClick={handleClose}
        aria-label="Benachrichtigung schließen"
      >
        <X size={16} />
      </button>
      {duration !== null && duration !== undefined && duration > 0 && (
        <div className={styles.toast__progress}>
          <div
            className={styles.toast__progressBar}
            style={{ width: `${progress}%` }}
          />
        </div>
      )}
    </div>
  )
}

export function ToastContainer() {
  const { notifications, removeNotification } = useNotificationStore()

  return (
    <div className={styles.container}>
      {notifications.map((n: Notification) => (
        <ToastItem
          key={n.id}
          notification={n}
          onClose={() => removeNotification(n.id)}
        />
      ))}
    </div>
  )
}
