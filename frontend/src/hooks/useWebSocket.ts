import { useEffect, useRef, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';

const RECONNECT_DELAY_MS = 5000;
const WS_PATH = '/api/v1/notifications/ws';

/**
 * Connects to the notification WebSocket when the user is authenticated.
 * On each incoming message the notifications query cache is invalidated
 * so that any mounted component re-fetches automatically.
 *
 * Auto-reconnects with a 5-second delay when the connection drops.
 * Cleans up on unmount.
 */
export function useWebSocket(): void {
  const queryClient = useQueryClient();
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const unmountedRef = useRef(false);

  const clearReconnectTimer = useCallback(() => {
    if (reconnectTimerRef.current !== null) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
  }, []);

  const connect = useCallback(() => {
    const token = localStorage.getItem('cd_access_token');
    if (!token || unmountedRef.current) {
      return;
    }

    // Build ws(s) URL based on current page protocol
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${protocol}//${window.location.host}${WS_PATH}?token=${encodeURIComponent(token)}`;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onmessage = (event: MessageEvent) => {
      try {
        // Parse to verify it is valid JSON, but we only need to invalidate
        JSON.parse(event.data as string);
      } catch {
        // Not JSON — ignore malformed messages
        return;
      }
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    };

    ws.onclose = () => {
      wsRef.current = null;
      if (!unmountedRef.current) {
        reconnectTimerRef.current = setTimeout(connect, RECONNECT_DELAY_MS);
      }
    };

    ws.onerror = () => {
      // onclose will fire after onerror, triggering reconnect
      ws.close();
    };
  }, [queryClient, clearReconnectTimer]);

  useEffect(() => {
    unmountedRef.current = false;
    connect();

    return () => {
      unmountedRef.current = true;
      clearReconnectTimer();
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect, clearReconnectTimer]);
}
