import { useState, useEffect, useRef, useCallback } from 'react'
import { Bell, CheckCircle, XCircle, AlertTriangle, Info, Check } from 'lucide-react'
import { useNotificationStore, type Notification, type NotificationType } from '../../stores/notificationStore'
import styles from './NotificationCenter.module.scss'

// Demo notifications seeded on mount
const DEMO_NOTIFICATIONS: Array<{
  message: string
  type: NotificationType
  title: string
  minutesAgo: number
  read: boolean
}> = [
  {
    title: 'Rechnung überfällig',
    message: 'Rechnung RE-2024-001 ist überfällig (3 Tage)',
    type: 'warning',
    minutesAgo: 25,
    read: false,
  },
  {
    title: 'Projekt bestätigt',
    message: "Projekt 'Konzert im Park' wurde bestätigt",
    type: 'success',
    minutesAgo: 120,
    read: false,
  },
  {
    title: 'Wartung fällig',
    message: 'Wartung fällig: JBL VTX A12',
    type: 'info',
    minutesAgo: 300,
    read: true,
  },
  {
    title: 'Neuer Kontakt',
    message: 'Neuer Kontakt angelegt: Festival GmbH',
    type: 'info',
    minutesAgo: 1440,
    read: true,
  },
]

function formatRelativeTime(timestamp: number): string {
  const now = Date.now()
  const diff = now - timestamp
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (minutes < 1) return 'gerade eben'
  if (minutes < 60) return `vor ${minutes} Min.`
  if (hours < 24) return `vor ${hours} ${hours === 1 ? 'Stunde' : 'Stunden'}`
  if (days < 7) return `vor ${days} ${days === 1 ? 'Tag' : 'Tagen'}`
  return new Date(timestamp).toLocaleDateString('de-DE')
}

function getNotificationIcon(type: NotificationType) {
  const size = 18
  switch (type) {
    case 'success':
      return <CheckCircle size={size} />
    case 'error':
      return <XCircle size={size} />
    case 'warning':
      return <AlertTriangle size={size} />
    case 'info':
      return <Info size={size} />
  }
}

interface NotificationItemProps {
  notification: Notification
  onMarkRead: (id: string) => void
}

function NotificationItem({ notification, onMarkRead }: NotificationItemProps) {
  return (
    <div
      className={`${styles.item} ${!notification.read ? styles['item--unread'] : ''}`}
      onClick={() => !notification.read && onMarkRead(notification.id)}
    >
      <div className={`${styles.item__icon} ${styles[`item__icon--${notification.type}`]}`}>
        {getNotificationIcon(notification.type)}
      </div>
      <div className={styles.item__content}>
        <div className={styles.item__title}>{notification.title}</div>
        <div className={styles.item__message}>{notification.message}</div>
        <div className={styles.item__time}>{formatRelativeTime(notification.timestamp)}</div>
      </div>
      {!notification.read && <div className={styles.item__dot} />}
    </div>
  )
}

// Store ref to track if demo notifications have been seeded
let demoSeeded = false

export function NotificationCenter() {
  const [isOpen, setIsOpen] = useState(false)
  const panelRef = useRef<HTMLDivElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)

  const { notifications, markAsRead, markAllRead, addNotification } = useNotificationStore()
  const unreadCount = notifications.filter((n) => !n.read).length

  // Seed demo notifications once
  useEffect(() => {
    if (demoSeeded) return
    demoSeeded = true

    DEMO_NOTIFICATIONS.forEach((demo) => {
      const id = addNotification(demo.message, demo.type, {
        title: demo.title,
        duration: null, // persistent - these are center notifications, not toasts
      })

      // Set the correct timestamp and read state
      const store = useNotificationStore.getState()
      const notif = store.notifications.find((n: Notification) => n.id === id)
      if (notif) {
        useNotificationStore.setState({
          notifications: store.notifications.map((n: Notification) =>
            n.id === id
              ? {
                  ...n,
                  timestamp: Date.now() - demo.minutesAgo * 60000,
                  read: demo.read,
                }
              : n
          ),
        })
      }
    })
  }, [addNotification])

  // Close on outside click
  useEffect(() => {
    if (!isOpen) return

    const handleClick = (e: MouseEvent) => {
      if (
        panelRef.current &&
        !panelRef.current.contains(e.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false)
      }
    }

    // Use setTimeout so the opening click doesn't immediately close it
    const timer = setTimeout(() => {
      document.addEventListener('click', handleClick)
    }, 0)

    return () => {
      clearTimeout(timer)
      document.removeEventListener('click', handleClick)
    }
  }, [isOpen])

  // Close on Escape
  useEffect(() => {
    if (!isOpen) return

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setIsOpen(false)
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen])

  const handleToggle = useCallback(() => {
    setIsOpen((prev) => !prev)
  }, [])

  // Only show persistent notifications (those in the center), sorted newest first
  const centerNotifications = [...notifications].sort((a, b) => b.timestamp - a.timestamp)

  return (
    <div className={styles.wrapper}>
      <button
        ref={buttonRef}
        className={styles.bellButton}
        onClick={handleToggle}
        aria-label="Benachrichtigungen"
        aria-expanded={isOpen}
      >
        <Bell size={20} />
        {unreadCount > 0 && (
          <span className={styles.badge}>
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div ref={panelRef} className={styles.panel}>
          <div className={styles.panel__header}>
            <h3 className={styles.panel__title}>Benachrichtigungen</h3>
            {unreadCount > 0 && (
              <button className={styles.panel__markAll} onClick={markAllRead}>
                <Check size={14} />
                Alle als gelesen markieren
              </button>
            )}
          </div>

          <div className={styles.panel__list}>
            {centerNotifications.length === 0 ? (
              <div className={styles.panel__empty}>
                <Bell size={32} />
                <p>Keine Benachrichtigungen</p>
              </div>
            ) : (
              centerNotifications.map((n) => (
                <NotificationItem
                  key={n.id}
                  notification={n}
                  onMarkRead={markAsRead}
                />
              ))
            )}
          </div>

          <div className={styles.panel__footer}>
            <a href="/notifications" className={styles.panel__link}>
              Alle Benachrichtigungen
            </a>
          </div>
        </div>
      )}
    </div>
  )
}
