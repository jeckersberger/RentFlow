import { create } from 'zustand'

export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export interface NotificationAction {
  label: string
  onClick: () => void
}

export interface Notification {
  id: string
  type: NotificationType
  message: string
  title?: string
  timestamp: number
  read: boolean
  duration?: number | null // in milliseconds, null for persistent
  action?: NotificationAction
}

interface NotificationStore {
  notifications: Notification[]
  addNotification: (
    message: string,
    type: NotificationType,
    options?: {
      title?: string
      duration?: number | null
      id?: string
      action?: NotificationAction
    }
  ) => string
  removeNotification: (id: string) => void
  markAsRead: (id: string) => void
  markAllRead: () => void
  clearAll: () => void
  unreadCount: () => number
}

const generateId = () => `notif_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`

export const useNotificationStore = create<NotificationStore>((set) => ({
  notifications: [],

  addNotification: (message: string, type: NotificationType, options = {}) => {
    const id = options.id || generateId()
    const duration = options.duration === undefined ? 4000 : options.duration

    const notification: Notification = {
      id,
      type,
      message,
      title: options.title,
      timestamp: Date.now(),
      read: false,
      duration,
      action: options.action,
    }

    set((state) => ({
      notifications: [...state.notifications, notification],
    }))

    // Auto-remove notification after duration
    if (duration !== null) {
      setTimeout(() => {
        set((state) => ({
          notifications: state.notifications.filter((n) => n.id !== id),
        }))
      }, duration)
    }

    return id
  },

  removeNotification: (id: string) =>
    set((state) => ({
      notifications: state.notifications.filter((n) => n.id !== id),
    })),

  markAsRead: (id: string) =>
    set((state) => ({
      notifications: state.notifications.map((n) =>
        n.id === id ? { ...n, read: true } : n
      ),
    })),

  markAllRead: () =>
    set((state) => ({
      notifications: state.notifications.map((n) => ({ ...n, read: true })),
    })),

  clearAll: () =>
    set({
      notifications: [],
    }),

  unreadCount: (): number => {
    return (useNotificationStore as { getState: () => NotificationStore }).getState().notifications.filter((n: Notification) => !n.read).length
  },
}))
