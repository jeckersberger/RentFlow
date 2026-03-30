import { useState, useRef, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion, AnimatePresence } from 'framer-motion';
import { Bell, CheckCheck } from 'lucide-react';
import toast from 'react-hot-toast';
import api from '@/services/api';
import './NotificationBell.scss';

interface Notification {
  id: string;
  title: string;
  body?: string;
  created_at: string;
  read: boolean;
}

interface UnreadCountResponse {
  count: number;
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'gerade eben';
  if (mins < 60) return `vor ${mins} Min.`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `vor ${hours} Std.`;
  const days = Math.floor(hours / 24);
  return `vor ${days} Tag${days > 1 ? 'en' : ''}`;
}

export function NotificationBell() {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  // Close on outside click
  const handleClickOutside = useCallback((e: MouseEvent) => {
    if (ref.current && !ref.current.contains(e.target as Node)) {
      setOpen(false);
    }
  }, []);

  useEffect(() => {
    if (open) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [open, handleClickOutside]);

  // Unread count
  const { data: unreadData } = useQuery({
    queryKey: ['notifications-unread-count'],
    queryFn: () =>
      api.get('/api/v1/notifications/unread-count') as unknown as UnreadCountResponse,
    refetchInterval: 30000,
  });
  const unreadCount =
    typeof unreadData === 'number'
      ? unreadData
      : (unreadData as UnreadCountResponse)?.count ?? 0;

  // Recent notifications
  const { data: notificationsRaw } = useQuery({
    queryKey: ['notifications-recent'],
    queryFn: () =>
      api.get('/api/v1/notifications', {
        params: { page: 1, per_page: 5 },
      }) as unknown as Notification[],
    enabled: open,
  });
  const notifications: Notification[] = Array.isArray(notificationsRaw)
    ? notificationsRaw
    : [];

  // Mark all read
  const markAllRead = useMutation({
    mutationFn: async () => {
      await api.post('/api/v1/notifications/mark-all-read');
    },
    onSuccess: () => {
      toast.success('Alle Benachrichtigungen als gelesen markiert');
      queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] });
      queryClient.invalidateQueries({ queryKey: ['notifications-recent'] });
    },
    onError: () => toast.error('Fehler beim Markieren'),
  });

  return (
    <div className="notification-bell" ref={ref}>
      <button
        className="notification-bell__trigger"
        onClick={() => setOpen((prev) => !prev)}
        aria-label="Benachrichtigungen"
      >
        <Bell size={20} />
        {unreadCount > 0 && (
          <span className="notification-bell__badge">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>

      <AnimatePresence>
        {open && (
          <motion.div
            className="notification-bell__dropdown"
            initial={{ opacity: 0, y: -6, scale: 0.97 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -6, scale: 0.97 }}
            transition={{ duration: 0.15 }}
          >
            <div className="notification-bell__header">
              <span className="notification-bell__title">
                Benachrichtigungen
              </span>
              {unreadCount > 0 && (
                <button
                  className="notification-bell__mark-read"
                  onClick={() => markAllRead.mutate()}
                  disabled={markAllRead.isPending}
                >
                  <CheckCheck size={14} />
                  Alle als gelesen
                </button>
              )}
            </div>

            <div className="notification-bell__list">
              {notifications.length === 0 ? (
                <div className="notification-bell__empty">
                  Keine Benachrichtigungen
                </div>
              ) : (
                notifications.map((n) => (
                  <div
                    key={n.id}
                    className={`notification-bell__item ${!n.read ? 'notification-bell__item--unread' : ''}`}
                  >
                    <span className="notification-bell__item-title">
                      {n.title}
                    </span>
                    <span className="notification-bell__item-time">
                      {timeAgo(n.created_at)}
                    </span>
                  </div>
                ))
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
