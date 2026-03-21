import { create } from 'zustand'

export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export interface Notification {
  id: string
  type: NotificationType
  message: string
  title?: string
  timestamp: number
  read: boolean
  duration?: number // in milliseconds, null for persistent
}

interface NotificationStore {
  notifications: Notification[]
  addNotification: (
    message: string,
    type: NotificationType,
    options?: {
      title?: string
      duration?: number
      id?: string
    }
  ) => string
  removeNotification: (id: string) => void
  markAsRead: (id: string) => void
  markAllRead: () => void
  clearAll: () => void
}

const generateId = () => `notif_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`

export const useNotificationStore = create<NotificationStore>((set, get) => ({
  notifications: [],

  addNotification: (message: string, type: NotificationType, options = {}) => {
    const id = options.id || generateId()
    const duration = options.duration ?? 5000 // Default 5 seconds

    const notification: Notification = {
      id,
      type,
      message,
      title: options.title,
      timestamp: Date.now(),
      read: false,
      duration,
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
}))
