import { useEffect } from 'react'
import { useNotificationStore, type Notification } from '../../stores/notificationStore'
import styles from './Toast.module.scss'

interface ToastItemProps {
  notification: Notification
  onClose: () => void
}

function ToastItem({ notification, onClose }: ToastItemProps) {
  useEffect(() => {
    if (notification.duration === null) {
      return
    }

    const timer = setTimeout(onClose, notification.duration)
    return () => clearTimeout(timer)
  }, [notification.duration, onClose])

  const getIcon = () => {
    switch (notification.type) {
      case 'success':
        return '✓'
      case 'error':
        return '✕'
      case 'warning':
        return '!'
      case 'info':
        return 'ℹ'
      default:
        return '•'
    }
  }

  return (
    <div className={`${styles.toast} ${styles[`toast--${notification.type}`]}`}>
      <div className={styles.toast__icon}>{getIcon()}</div>
      <div className={styles.toast__content}>
        {notification.title && (
          <div className={styles.toast__title}>{notification.title}</div>
        )}
        <div className={styles.toast__message}>{notification.message}</div>
      </div>
      <button
        className={styles.toast__close}
        onClick={onClose}
        aria-label="Close notification"
      >
        ×
      </button>
    </div>
  )
}

export function ToastContainer() {
  const { notifications, removeNotification } = useNotificationStore()

  return (
    <div className={styles.container}>
      {notifications.map((n) => (
        <ToastItem
          key={n.id}
          notification={n}
          onClose={() => removeNotification(n.id)}
        />
      ))}
    </div>
  )
}
