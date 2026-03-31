import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { scannerDeviceApi } from '../../../services/api'
import { Bell, RefreshCw, QrCode, CheckCircle2, Wifi, WifiOff } from 'lucide-react'
import '../Settings.scss'

interface ScannerDevice {
  id: string
  name: string
  device_id: string
  type: string
  is_online: boolean
  last_seen: string
}

function ScannerDevicesPage() {
  const queryClient = useQueryClient()

  // QR token state
  const [qrToken, setQrToken] = useState<string | null>(null)
  const [qrExpiry, setQrExpiry] = useState<number>(0) // seconds remaining
  const [ringingDeviceId, setRingingDeviceId] = useState<string | null>(null)
  const [ringSuccess, setRingSuccess] = useState<string | null>(null)

  // Countdown timer
  useEffect(() => {
    if (qrExpiry <= 0) {
      if (qrToken) setQrToken(null)
      return
    }
    const timer = setInterval(() => {
      setQrExpiry((prev) => {
        if (prev <= 1) {
          setQrToken(null)
          return 0
        }
        return prev - 1
      })
    }, 1000)
    return () => clearInterval(timer)
  }, [qrExpiry, qrToken])

  // Fetch devices
  const { data: devicesData, isLoading } = useQuery<ScannerDevice[] | Record<string, unknown>>({
    queryKey: ['scanner-devices'],
    queryFn: () => scannerDeviceApi.list(),
    refetchInterval: 15000, // refresh every 15s for online status
    retry: 1,
  })

  // Unwrap response: API may return plain array or { data: [...] } or { devices: [...] }
  const devices: ScannerDevice[] = Array.isArray(devicesData)
    ? devicesData
    : ((devicesData as any)?.data || (devicesData as any)?.devices || [])

  // Generate QR token
  const qrMutation = useMutation({
    mutationFn: () => scannerDeviceApi.generateQRToken(),
    onSuccess: (data) => {
      setQrToken(data.token)
      setQrExpiry(300) // 5 minutes
    },
  })

  // Ring device
  const ringMutation = useMutation({
    mutationFn: (id: string) => scannerDeviceApi.ring(id),
    onSuccess: (_data, id) => {
      setRingingDeviceId(null)
      setRingSuccess(id)
      setTimeout(() => setRingSuccess(null), 4000)
      queryClient.invalidateQueries({ queryKey: ['scanner-devices'] })
    },
    onError: () => {
      setRingingDeviceId(null)
    },
  })

  const handleRing = useCallback(
    (id: string) => {
      setRingingDeviceId(id)
      setRingSuccess(null)
      ringMutation.mutate(id)
    },
    [ringMutation]
  )

  const formatTime = (seconds: number): string => {
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return `${m}:${s.toString().padStart(2, '0')}`
  }

  const formatLastSeen = (iso: string): string => {
    if (!iso) return '-'
    const d = new Date(iso)
    return d.toLocaleString('de-DE', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const isOnline = (lastSeen: string): boolean => {
    if (!lastSeen) return false
    const diff = Date.now() - new Date(lastSeen).getTime()
    return diff < 60_000 // 60 seconds
  }

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Scanner-Geraete</h1>
        <p>Verwalten Sie Ihre registrierten Scanner-Geraete, generieren Sie QR-Login-Codes und finden Sie Geraete per Klingelfunktion.</p>
      </div>

      {/* QR Login Code Section */}
      <div className="sp-card">
        <h3 className="sp-card__title">QR-Login-Code generieren</h3>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 'var(--spacing-4)' }}>
          {!qrToken ? (
            <button
              className="sp-btn sp-btn--primary"
              onClick={() => qrMutation.mutate()}
              disabled={qrMutation.isPending}
              style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}
            >
              <QrCode size={16} />
              {qrMutation.isPending ? 'Wird generiert...' : 'QR-Code generieren'}
            </button>
          ) : (
            <>
              <img
                src={`https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=${encodeURIComponent(`rentflow://qr-login/${qrToken}`)}`}
                alt="Scanner QR-Login-Code"
                style={{ width: 300, height: 300, borderRadius: 'var(--radius-md)', border: '1px solid var(--color-border)', background: '#fff' }}
              />
              <div style={{ textAlign: 'center' }}>
                <div style={{
                  fontSize: 'var(--font-size-2xl)',
                  fontWeight: 'var(--font-weight-bold)',
                  color: qrExpiry < 60 ? 'var(--color-danger)' : 'var(--color-text-primary)',
                  fontVariantNumeric: 'tabular-nums',
                }}>
                  {formatTime(qrExpiry)}
                </div>
                <p style={{ color: 'var(--color-text-muted)', fontSize: 'var(--font-size-sm)', margin: 'var(--spacing-2) 0 0 0' }}>
                  Scannen Sie diesen Code mit der CrateDesk Scanner App
                </p>
              </div>

              {/* Deep Link for scanner app */}
              <a
                href={`rentflow://qr-login/${qrToken}`}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '0.4rem',
                  padding: 'var(--spacing-2) var(--spacing-4)',
                  backgroundColor: 'rgba(0, 212, 255, 0.1)',
                  color: 'var(--color-primary)',
                  borderRadius: 'var(--radius-md)',
                  fontSize: 'var(--font-size-sm)',
                  fontWeight: 'var(--font-weight-medium)',
                  textDecoration: 'none',
                  border: '1px solid rgba(0, 212, 255, 0.2)',
                }}
              >
                Scanner-App oeffnen
              </a>

              {/* Manual token entry fallback */}
              <div style={{
                width: '100%',
                padding: 'var(--spacing-3) var(--spacing-4)',
                backgroundColor: 'var(--color-bg-secondary)',
                borderRadius: 'var(--radius-md)',
                border: '1px solid var(--color-border)',
              }}>
                <p style={{ margin: '0 0 var(--spacing-2) 0', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', textAlign: 'center' }}>
                  Alternativ: Token manuell eingeben
                </p>
                <div
                  style={{
                    fontFamily: 'monospace',
                    fontSize: 'var(--font-size-lg)',
                    fontWeight: 'var(--font-weight-bold)',
                    textAlign: 'center',
                    letterSpacing: '0.15em',
                    color: 'var(--color-text-primary)',
                    wordBreak: 'break-all',
                    userSelect: 'all',
                    cursor: 'text',
                    padding: 'var(--spacing-2) 0',
                  }}
                  title="Klicken zum Markieren"
                >
                  {qrToken}
                </div>
              </div>
              <button
                className="sp-btn sp-btn--secondary"
                onClick={() => {
                  setQrToken(null)
                  setQrExpiry(0)
                  qrMutation.mutate()
                }}
                style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}
              >
                <RefreshCw size={14} />
                Neuen Code generieren
              </button>
            </>
          )}
          {qrMutation.isError && (
            <p style={{ color: 'var(--color-danger)', fontSize: 'var(--font-size-sm)' }}>
              Fehler beim Generieren des QR-Codes. Bitte versuchen Sie es erneut.
            </p>
          )}
        </div>
      </div>

      {/* Registered Devices Table */}
      <div className="sp-card">
        <h3 className="sp-card__title">Registrierte Scanner</h3>

        {isLoading ? (
          <div className="sp-loading">Lade Geraete...</div>
        ) : !devices || devices.length === 0 ? (
          <div className="sp-empty">
            Keine Scanner-Geraete registriert. Generieren Sie einen QR-Code und scannen Sie ihn mit der CrateDesk Scanner App.
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className="sp-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Geraete-ID</th>
                  <th>Typ</th>
                  <th>Online-Status</th>
                  <th>Letzter Kontakt</th>
                  <th>Aktionen</th>
                </tr>
              </thead>
              <tbody>
                {devices.map((device) => {
                  const online = device.is_online ?? isOnline(device.last_seen)
                  const isRinging = ringingDeviceId === device.id
                  const didRespond = ringSuccess === device.id

                  return (
                    <tr key={device.id}>
                      <td style={{ fontWeight: 'var(--font-weight-medium)' }}>{device.name}</td>
                      <td>
                        <code style={{ fontSize: 'var(--font-size-xs)', background: 'var(--color-bg-secondary)', padding: '2px 6px', borderRadius: 'var(--radius-sm)' }}>
                          {device.device_id}
                        </code>
                      </td>
                      <td>{device.type}</td>
                      <td>
                        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
                          {online ? (
                            <>
                              <span style={{
                                display: 'inline-block',
                                width: 8,
                                height: 8,
                                borderRadius: '50%',
                                background: '#10b981',
                                boxShadow: '0 0 6px rgba(16, 185, 129, 0.5)',
                              }} />
                              <Wifi size={14} style={{ color: '#10b981' }} />
                              Online
                            </>
                          ) : (
                            <>
                              <span style={{
                                display: 'inline-block',
                                width: 8,
                                height: 8,
                                borderRadius: '50%',
                                background: 'var(--color-text-muted)',
                              }} />
                              <WifiOff size={14} style={{ color: 'var(--color-text-muted)' }} />
                              Offline
                            </>
                          )}
                        </span>
                      </td>
                      <td>{formatLastSeen(device.last_seen)}</td>
                      <td>
                        {didRespond ? (
                          <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', color: 'var(--color-success)', fontSize: 'var(--font-size-sm)' }}>
                            <CheckCircle2 size={14} />
                            Geraet hat geantwortet
                          </span>
                        ) : (
                          <button
                            className="sp-btn sp-btn--secondary"
                            onClick={() => handleRing(device.id)}
                            disabled={isRinging || !online}
                            title={!online ? 'Geraet ist offline' : 'Geraet klingeln lassen'}
                            style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', fontSize: 'var(--font-size-xs)', padding: 'var(--spacing-1) var(--spacing-3)' }}
                          >
                            <Bell size={14} className={isRinging ? 'scanner-ring-anim' : ''} />
                            {isRinging ? 'Klingelt...' : 'Klingeln'}
                          </button>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Inline keyframe for ring animation */}
      <style>{`
        .scanner-ring-anim {
          animation: scannerRing 0.4s ease-in-out infinite alternate;
        }
        @keyframes scannerRing {
          0% { transform: rotate(-12deg); }
          100% { transform: rotate(12deg); }
        }
      `}</style>
    </div>
  )
}

export default ScannerDevicesPage
