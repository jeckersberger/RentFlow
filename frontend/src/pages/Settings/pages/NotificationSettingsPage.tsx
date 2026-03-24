import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { notificationPreferencesApi, type NotificationPreference } from '../../../services/api'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import { Bell, Mail, Smartphone, MonitorSmartphone } from 'lucide-react'
import '../Settings.scss'

// ---------------------------------------------------------------------------
// Event types the user can configure, with German labels
// ---------------------------------------------------------------------------
const EVENT_TYPES: { key: string; label: string; description: string }[] = [
  {
    key: 'invoice_overdue',
    label: 'Rechnung ueberfaellig',
    description: 'Benachrichtigung wenn eine Rechnung ueberfaellig ist',
  },
  {
    key: 'project_created',
    label: 'Neues Projekt erstellt',
    description: 'Wenn ein neues Projekt angelegt wird',
  },
  {
    key: 'equipment_checked_out',
    label: 'Equipment ausgecheckt',
    description: 'Wenn Equipment aus dem Lager entnommen wird',
  },
  {
    key: 'maintenance_due',
    label: 'Wartung faellig',
    description: 'Erinnerung an anstehende Wartungstermine',
  },
  {
    key: 'freelancer_response',
    label: 'Freelancer-Antwort erhalten',
    description: 'Wenn ein Freelancer auf eine Anfrage antwortet',
  },
  {
    key: 'scanner_session_completed',
    label: 'Scanner-Session abgeschlossen',
    description: 'Wenn ein Scan-Vorgang abgeschlossen wurde',
  },
]

const CHANNELS: { key: string; label: string; icon: typeof Mail }[] = [
  { key: 'email', label: 'E-Mail', icon: Mail },
  { key: 'in_app', label: 'In-App', icon: MonitorSmartphone },
  { key: 'push', label: 'Push', icon: Smartphone },
]

// ---------------------------------------------------------------------------
// Helper: build a lookup  map  from the API response
// ---------------------------------------------------------------------------
function buildPrefMap(prefs: NotificationPreference[] | undefined) {
  const map: Record<string, NotificationPreference> = {}
  if (!prefs) return map
  for (const p of prefs) {
    map[p.event_type] = p
  }
  return map
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------
function NotificationSettingsPage() {
  const queryClient = useQueryClient()

  const { data: preferences, isLoading } = useQuery<NotificationPreference[]>({
    queryKey: ['notification-preferences'],
    queryFn: () => notificationPreferencesApi.list(),
  })

  const prefMap = buildPrefMap(preferences)

  const mutation = useMutation({
    mutationFn: (pref: NotificationPreference) => notificationPreferencesApi.update(pref),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-preferences'] })
    },
  })

  // Toggle a single channel for a given event type
  const handleToggle = (eventType: string, channel: string, currentlyEnabled: boolean) => {
    const existing = prefMap[eventType]
    const currentChannels = existing?.channels ?? []
    let newChannels: string[]

    if (currentlyEnabled) {
      // remove channel
      newChannels = currentChannels.filter((c) => c !== channel)
    } else {
      // add channel
      newChannels = [...currentChannels, channel]
    }

    mutation.mutate({
      event_type: eventType,
      channels: newChannels,
      is_enabled: newChannels.length > 0,
    })
  }

  // Toggle ALL channels for an event type on/off
  const handleToggleAll = (eventType: string) => {
    const existing = prefMap[eventType]
    const allEnabled = existing?.is_enabled && (existing?.channels?.length ?? 0) > 0

    if (allEnabled) {
      // disable all
      mutation.mutate({
        event_type: eventType,
        channels: [],
        is_enabled: false,
      })
    } else {
      // enable all channels
      mutation.mutate({
        event_type: eventType,
        channels: CHANNELS.map((c) => c.key),
        is_enabled: true,
      })
    }
  }

  if (isLoading) {
    return (
      <div className="sp-page">
        <div className="sp-header">
          <h1>Benachrichtigungen</h1>
        </div>
        <SkeletonCard count={3} />
      </div>
    )
  }

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Benachrichtigungen</h1>
        <p>
          Legen Sie fest, bei welchen Ereignissen und ueber welche Kanaele Sie
          benachrichtigt werden moechten.
        </p>
      </div>

      {/* Channel legend */}
      <div className="sp-card">
        <h3 className="sp-card__title">Kanaele</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-6)', flexWrap: 'wrap' }}>
          {CHANNELS.map((ch) => {
            const Icon = ch.icon
            return (
              <div
                key={ch.key}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 'var(--spacing-2)',
                  color: 'var(--color-text-secondary)',
                  fontSize: 'var(--font-size-sm)',
                }}
              >
                <Icon size={16} />
                {ch.label}
              </div>
            )
          })}
        </div>
      </div>

      {/* Preference rows */}
      <div className="sp-card">
        <h3 className="sp-card__title">Ereignisse</h3>

        <table className="sp-table" style={{ tableLayout: 'fixed' }}>
          <thead>
            <tr>
              <th style={{ width: '40%' }}>Ereignis</th>
              {CHANNELS.map((ch) => {
                const Icon = ch.icon
                return (
                  <th key={ch.key} style={{ width: '15%', textAlign: 'center' }}>
                    <span
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 4,
                      }}
                    >
                      <Icon size={14} />
                      {ch.label}
                    </span>
                  </th>
                )
              })}
              <th style={{ width: '15%', textAlign: 'center' }}>
                <span
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: 4,
                  }}
                >
                  <Bell size={14} />
                  Alle
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            {EVENT_TYPES.map((evt) => {
              const pref = prefMap[evt.key]
              const channels = pref?.channels ?? []
              const allOn = pref?.is_enabled && channels.length === CHANNELS.length

              return (
                <tr key={evt.key}>
                  <td>
                    <div>
                      <div
                        style={{
                          fontWeight: 'var(--font-weight-medium)' as any,
                          color: 'var(--color-text-primary)',
                        }}
                      >
                        {evt.label}
                      </div>
                      <div
                        style={{
                          fontSize: 'var(--font-size-xs)',
                          color: 'var(--color-text-muted)',
                          marginTop: 2,
                        }}
                      >
                        {evt.description}
                      </div>
                    </div>
                  </td>
                  {CHANNELS.map((ch) => {
                    const isOn = channels.includes(ch.key)
                    return (
                      <td key={ch.key} style={{ textAlign: 'center' }}>
                        <input
                          type="checkbox"
                          checked={isOn}
                          onChange={() => handleToggle(evt.key, ch.key, isOn)}
                          disabled={mutation.isPending}
                          style={{
                            width: '1.25rem',
                            height: '1.25rem',
                            accentColor: 'var(--color-primary)',
                            cursor: mutation.isPending ? 'wait' : 'pointer',
                          }}
                        />
                      </td>
                    )
                  })}
                  <td style={{ textAlign: 'center' }}>
                    <input
                      type="checkbox"
                      checked={!!allOn}
                      onChange={() => handleToggleAll(evt.key)}
                      disabled={mutation.isPending}
                      style={{
                        width: '1.25rem',
                        height: '1.25rem',
                        accentColor: 'var(--color-primary)',
                        cursor: mutation.isPending ? 'wait' : 'pointer',
                      }}
                    />
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {mutation.isError && (
        <div style={{ color: 'var(--color-danger)', fontSize: 'var(--font-size-sm)', marginTop: 'var(--spacing-2)' }}>
          Fehler beim Speichern der Einstellung. Bitte versuchen Sie es erneut.
        </div>
      )}
    </div>
  )
}

export default NotificationSettingsPage
