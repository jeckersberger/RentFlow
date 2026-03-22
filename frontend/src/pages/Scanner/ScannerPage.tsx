import { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import { Html5Qrcode, Html5QrcodeScannerState } from 'html5-qrcode'
import { Select } from '../../components/Form/Select'
import { Input } from '../../components/Form/Input'
import { equipmentApi, projectApi } from '../../services/api'
import '../Equipment/Equipment.module.scss'
import './Scanner.module.scss'

interface ScanEvent {
  id: string
  barcode: string
  equipment_name?: string
  equipment_id?: string
  scan_type: string
  timestamp: string
  status: 'success' | 'error' | 'pending'
  error_message?: string
}

interface Project {
  id: string
  name: string
  status: string
}

interface ScanSession {
  id: string
  context: 'check-out' | 'check-in' | 'warehouse-store' | 'inventory'
  project_id?: string
  started_at: string
  total_scans: number
  successful_scans: number
  failed_scans: number
}

function ScannerPage() {
  const scanInputRef = useRef<HTMLInputElement>(null)
  const html5QrcodeRef = useRef<Html5Qrcode | null>(null)
  const [barcode, setBarcode] = useState('')
  const [scanContext, setScanContext] = useState<'check-out' | 'check-in' | 'warehouse-store' | 'inventory'>('check-in')
  const [projectId, setProjectId] = useState('')
  const [projects, setProjects] = useState<Project[]>([])
  const [recentScans, setRecentScans] = useState<ScanEvent[]>([])
  const [batchMode, setBatchMode] = useState(false)
  const [cameraActive, setCameraActive] = useState(false)
  const [cameraError, setCameraError] = useState<string | null>(null)
  const [isProcessing, setIsProcessing] = useState(false)
  const [scanCount, setScanCount] = useState(0)
  const [successCount, setSuccessCount] = useState(0)
  const [errorCount, setErrorCount] = useState(0)
  const [sessionActive, setSessionActive] = useState(false)
  const [offlineMode] = useState(false)
  const [offlineQueue] = useState(0)

  // Mock session data
  const currentSession: ScanSession = {
    id: 'session-' + Date.now(),
    context: scanContext,
    project_id: projectId,
    started_at: new Date().toISOString(),
    total_scans: scanCount,
    successful_scans: successCount,
    failed_scans: errorCount,
  }

  // Compute session stats
  const sessionStats = useMemo(() => ({
    total: scanCount,
    successful: successCount,
    failed: errorCount,
    successRate: scanCount > 0 ? Math.round((successCount / scanCount) * 100) : 0,
  }), [scanCount, successCount, errorCount])

  // Projekte für Check-Out laden
  useEffect(() => {
    projectApi.list(1, 100)
      .then((data) => {
        const projectList = (data?.items || []).filter(
          (p: Project) => p.status === 'confirmed' || p.status === 'in_progress' || p.status === 'active' || p.status === 'planning'
        )
        setProjects(projectList)
      })
      .catch(() => {
        // Fallback: leere Liste, Projekt-Service evtl. nicht erreichbar
        setProjects([])
      })
  }, [])

  // Kamera-Scanner aufräumen beim Unmount
  useEffect(() => {
    return () => {
      if (html5QrcodeRef.current) {
        const state = html5QrcodeRef.current.getState()
        if (state === Html5QrcodeScannerState.SCANNING) {
          html5QrcodeRef.current.stop().catch(() => {})
        }
      }
    }
  }, [])

  // Equipment per Barcode/QR-Code suchen und Aktion ausführen
  const processBarcode = useCallback(async (scannedBarcode: string) => {
    if (!scannedBarcode || isProcessing) return

    setIsProcessing(true)
    const scanId = Date.now().toString()

    // Vibrieren als haptisches Feedback (falls verfügbar)
    if ('vibrate' in navigator) {
      navigator.vibrate(100)
    }

    // Pending Scan anzeigen
    const pendingScan: ScanEvent = {
      id: scanId,
      barcode: scannedBarcode,
      scan_type: scanContext,
      timestamp: new Date().toISOString(),
      status: 'pending',
    }
    setRecentScans((prev) => [pendingScan, ...prev.slice(0, 49)])
    setScanCount((prev) => prev + 1)

    try {
      // Schritt 1: Equipment per Barcode/UUID suchen
      let equipment: { id: string; name: string } | null = null

      // QR-Code-Format prüfen: "rentflow://equipment/{uuid}"
      const qrMatch = scannedBarcode.match(/^rentflow:\/\/equipment\/(.+)$/)
      if (qrMatch) {
        const equipmentId = qrMatch[1]
        equipment = await equipmentApi.getById(equipmentId)
      } else {
        // Als Barcode interpretieren
        equipment = await equipmentApi.getByBarcode(scannedBarcode)
      }

      if (!equipment) {
        throw new Error('Ausrüstung nicht gefunden')
      }

      // Schritt 2: Aktion basierend auf Scan-Kontext ausführen
      if (scanContext === 'check-out') {
        if (!projectId) {
          throw new Error('Bitte wählen Sie ein Projekt für das Auschecken')
        }
        await equipmentApi.checkOut(equipment.id, projectId)
      } else if (scanContext === 'check-in') {
        await equipmentApi.checkIn(equipment.id)
      } else if (scanContext === 'warehouse-store') {
        // Warehouse location update (mock)
        await new Promise(resolve => setTimeout(resolve, 200))
      } else if (scanContext === 'inventory') {
        // Inventory check (mock, no action needed)
        await new Promise(resolve => setTimeout(resolve, 100))
      }

      // Erfolg
      const successScan: ScanEvent = {
        ...pendingScan,
        equipment_name: equipment.name,
        equipment_id: equipment.id,
        status: 'success',
      }
      setRecentScans((prev) =>
        prev.map((s) => (s.id === scanId ? successScan : s))
      )
      setSuccessCount((prev) => prev + 1)

      // Vibrieren: Erfolg (kurz-kurz)
      if ('vibrate' in navigator) {
        navigator.vibrate([50, 50, 50])
      }
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message :
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Unbekannter Fehler'

      const errorScan: ScanEvent = {
        ...pendingScan,
        status: 'error',
        error_message: errorMessage,
      }
      setRecentScans((prev) =>
        prev.map((s) => (s.id === scanId ? errorScan : s))
      )
      setErrorCount((prev) => prev + 1)

      // Vibrieren: Fehler (lang)
      if ('vibrate' in navigator) {
        navigator.vibrate(300)
      }
    } finally {
      setIsProcessing(false)
      setBarcode('')
      if (!cameraActive) {
        scanInputRef.current?.focus()
      }
    }
  }, [scanContext, projectId, isProcessing, cameraActive])

  // Manuelle Eingabe per Enter
  const handleManualScan = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key !== 'Enter') return
    const scannedBarcode = barcode.trim()
    if (scannedBarcode) {
      processBarcode(scannedBarcode)
    }
  }

  // Kamera starten/stoppen
  const toggleCamera = async () => {
    if (cameraActive) {
      // Kamera stoppen
      if (html5QrcodeRef.current) {
        try {
          await html5QrcodeRef.current.stop()
        } catch { /* Ignorieren */ }
      }
      setCameraActive(false)
      setCameraError(null)
      return
    }

    // Kamera starten
    setCameraError(null)
    try {
      if (!html5QrcodeRef.current) {
        html5QrcodeRef.current = new Html5Qrcode('scanner-camera-view')
      }

      await html5QrcodeRef.current.start(
        { facingMode: 'environment' }, // Rückkamera bevorzugen
        {
          fps: 10,
          qrbox: { width: 250, height: 250 },
          aspectRatio: 1.0,
        },
        (decodedText) => {
          // QR-Code/Barcode erkannt
          processBarcode(decodedText)

          // Im Nicht-Batch-Modus Kamera nach erfolgreichem Scan stoppen
          if (!batchMode && html5QrcodeRef.current) {
            html5QrcodeRef.current.stop().catch(() => {})
            setCameraActive(false)
          }
        },
        () => {
          // Scan-Fehler (normaler Zustand während des Scannens) – ignorieren
        }
      )
      setCameraActive(true)
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Kamera konnte nicht gestartet werden'
      setCameraError(msg)
      setCameraActive(false)
    }
  }

  const handleStartSession = () => {
    setSessionActive(true)
    setScanCount(0)
    setSuccessCount(0)
    setErrorCount(0)
    setRecentScans([])
  }

  const handleEndSession = () => {
    setSessionActive(false)
  }

  return (
    <div className="scanner-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Scanner & Bestandsverwaltung</h1>
          <p className="page-subtitle">
            Scannen Sie Equipment per Kamera oder geben Sie den Code manuell ein
          </p>
        </div>
        <div className="header-actions">
          {!sessionActive ? (
            <button className="btn btn--primary" onClick={handleStartSession}>
              ▶ Sitzung starten
            </button>
          ) : (
            <button className="btn btn--secondary" onClick={handleEndSession}>
              ⏹ Sitzung beenden
            </button>
          )}
        </div>
      </div>

      {offlineMode && (
        <div className="offline-banner">
          <span>📡 Offline-Modus aktiv</span>
          <span className="offline-queue">Queue: {offlineQueue} Scans</span>
        </div>
      )}

      <div className="context-selector">
        <h3 className="context-selector__title">Scan-Kontext</h3>
        <div className="context-cards">
          {(['check-in', 'check-out', 'warehouse-store', 'inventory'] as const).map((ctx) => (
            <button
              key={ctx}
              className={`context-card ${scanContext === ctx ? 'context-card--active' : ''}`}
              onClick={() => setScanContext(ctx)}
            >
              <div className="context-card__icon">
                {ctx === 'check-in' && '📥'}
                {ctx === 'check-out' && '📤'}
                {ctx === 'warehouse-store' && '🏢'}
                {ctx === 'inventory' && '📊'}
              </div>
              <div className="context-card__label">
                {ctx === 'check-in' && 'Einchecken'}
                {ctx === 'check-out' && 'Auschecken'}
                {ctx === 'warehouse-store' && 'Lagerort'}
                {ctx === 'inventory' && 'Inventur'}
              </div>
            </button>
          ))}
        </div>
      </div>

      <div className="scanner-container">
        <div className="scanner-main">
          <div className="scanner-panel">
            <h2 className="scanner-panel__title">Scan-Eingabe</h2>

            {scanContext === 'check-out' && (
              <div className="scanner-panel__section">
                <Select
                  label="Projekt auswählen"
                  options={projects.map((p) => ({
                    value: p.id,
                    label: p.name,
                  }))}
                  value={projectId}
                  onChange={(e) => setProjectId(e.target.value)}
                  placeholder="Projekt für Auschecken..."
                />
              </div>
            )}

            {/* Kamera-Scan-Bereich */}
            <div className="scanner-panel__section">
              <div
                id="scanner-camera-view"
                style={{
                  width: '100%',
                  minHeight: cameraActive ? '300px' : '0px',
                  borderRadius: 'var(--radius-md)',
                  overflow: 'hidden',
                  marginBottom: cameraActive ? 'var(--spacing-4)' : '0',
                  transition: 'min-height 0.3s ease',
                }}
              />

              {!cameraActive && (
                <div
                  style={{
                    padding: 'var(--spacing-4)',
                    backgroundColor: 'var(--color-primary-50)',
                    borderRadius: 'var(--radius-md)',
                    textAlign: 'center',
                    marginBottom: 'var(--spacing-4)',
                    cursor: 'pointer',
                  }}
                  onClick={toggleCamera}
                >
                  <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-2)' }}>
                    📷
                  </div>
                  <p
                    style={{
                      margin: '0 0 var(--spacing-2) 0',
                      color: 'var(--color-text-primary)',
                      fontWeight: 'var(--font-weight-semibold)',
                    }}
                  >
                    Kamera starten zum Scannen
                  </p>
                  <p
                    style={{
                      margin: 0,
                      fontSize: 'var(--font-size-sm)',
                      color: 'var(--color-text-secondary)',
                    }}
                  >
                    Tippen Sie hier oder nutzen Sie die manuelle Eingabe unten
                  </p>
                </div>
              )}

              {cameraError && (
                <div
                  style={{
                    padding: 'var(--spacing-3)',
                    backgroundColor: 'var(--color-error-light, #fef2f2)',
                    color: 'var(--color-error-dark, #991b1b)',
                    borderRadius: 'var(--radius-md)',
                    marginBottom: 'var(--spacing-3)',
                    fontSize: 'var(--font-size-sm)',
                  }}
                >
                  Kamera-Fehler: {cameraError}
                </div>
              )}

              {/* Manuelle Eingabe */}
              <Input
                ref={scanInputRef}
                type="text"
                label="Barcode / QR-Code manuell eingeben"
                placeholder="Code eingeben und Enter drücken..."
                value={barcode}
                onChange={(e) => setBarcode(e.target.value)}
                onKeyPress={handleManualScan}
                autoFocus={!cameraActive}
                disabled={isProcessing}
              />
            </div>

            <div className="scanner-actions">
              <button
                className={`btn ${cameraActive ? 'btn--secondary' : 'btn--primary'}`}
                onClick={toggleCamera}
                style={{ width: '100%' }}
              >
                {cameraActive ? '⏹ Kamera stoppen' : '📷 Kamera starten'}
              </button>
              <label className="scanner-batch-toggle">
                <input
                  type="checkbox"
                  checked={batchMode}
                  onChange={(e) => setBatchMode(e.target.checked)}
                />
                <span>Batch-Modus (Kamera bleibt aktiv)</span>
              </label>
            </div>

            {isProcessing && (
              <div
                style={{
                  marginTop: 'var(--spacing-3)',
                  padding: 'var(--spacing-3)',
                  backgroundColor: 'var(--color-primary-50)',
                  borderRadius: 'var(--radius-md)',
                  textAlign: 'center',
                  fontSize: 'var(--font-size-sm)',
                }}
              >
                Verarbeite Scan...
              </div>
            )}
          </div>

          {/* Scan-Verlauf */}
          {recentScans.length > 0 && (
            <div className="scanner-history">
              <h2 className="scanner-history__title">
                Scan-Verlauf ({recentScans.length})
              </h2>
              <div className="scanner-history__list">
                {recentScans.map((scan, index) => (
                  <div
                    key={scan.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 'var(--spacing-3)',
                      padding: 'var(--spacing-3)',
                      backgroundColor:
                        scan.status === 'error'
                          ? 'var(--color-error-light, #fef2f2)'
                          : scan.status === 'pending'
                          ? 'var(--color-warning-light, #fffbeb)'
                          : index === 0
                          ? 'var(--color-primary-50)'
                          : 'var(--color-bg-secondary)',
                      borderRadius: 'var(--radius-md)',
                      borderLeft:
                        scan.status === 'error'
                          ? '3px solid var(--color-error, #dc2626)'
                          : scan.status === 'pending'
                          ? '3px solid var(--color-warning, #d97706)'
                          : index === 0
                          ? '3px solid var(--color-primary)'
                          : '3px solid var(--color-border)',
                      transition: 'all 0.2s ease',
                    }}
                  >
                    <div style={{ fontSize: '1.5rem' }}>
                      {scan.status === 'pending'
                        ? '⏳'
                        : scan.status === 'error'
                        ? '❌'
                        : scan.scan_type === 'check_in'
                        ? '📥'
                        : scan.scan_type === 'check_out'
                        ? '📤'
                        : '📊'}
                    </div>
                    <div style={{ flex: 1 }}>
                      <p
                        style={{
                          margin: '0 0 var(--spacing-1) 0',
                          fontWeight: 'var(--font-weight-semibold)',
                          color: 'var(--color-text-primary)',
                        }}
                      >
                        {scan.equipment_name || `Code: ${scan.barcode}`}
                      </p>
                      <p
                        style={{
                          margin: 0,
                          fontSize: 'var(--font-size-sm)',
                          color:
                            scan.status === 'error'
                              ? 'var(--color-error-dark, #991b1b)'
                              : 'var(--color-text-secondary)',
                        }}
                      >
                        {scan.status === 'error'
                          ? scan.error_message
                          : scan.status === 'pending'
                          ? 'Wird verarbeitet...'
                          : (
                            <>
                              {scan.scan_type === 'check_in' && 'Eingecheckt'}
                              {scan.scan_type === 'check_out' && 'Ausgecheckt'}
                              {scan.scan_type === 'inventory' && 'Inventur'}
                              {' • '}
                              {new Date(scan.timestamp).toLocaleTimeString('de-DE')}
                            </>
                          )}
                      </p>
                    </div>
                    <span
                      style={{
                        fontSize: 'var(--font-size-xs)',
                        padding: 'var(--spacing-1) var(--spacing-2)',
                        backgroundColor:
                          scan.status === 'error'
                            ? 'var(--color-error-light, #fef2f2)'
                            : scan.status === 'pending'
                            ? 'var(--color-warning-light, #fffbeb)'
                            : 'var(--color-success-light)',
                        color:
                          scan.status === 'error'
                            ? 'var(--color-error-dark, #991b1b)'
                            : scan.status === 'pending'
                            ? 'var(--color-warning-dark, #92400e)'
                            : 'var(--color-success-dark)',
                        borderRadius: 'var(--radius-base)',
                      }}
                    >
                      {scan.status === 'error'
                        ? '✗ Fehler'
                        : scan.status === 'pending'
                        ? '...'
                        : '✓ OK'}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        <div className="scanner-sidebar">
          <div className="session-info">
            <div className="session-info__status">
              {sessionActive ? (
                <>
                  <span className="status-badge status-badge--active">Sitzung aktiv</span>
                  <span className="session-timer">
                    {Math.floor(Date.now() / 1000 % 3600 / 60)}m aktiv
                  </span>
                </>
              ) : (
                <span className="status-badge">Keine Sitzung</span>
              )}
            </div>
          </div>

          <div className="scanner-stats">
            <div className="stat-box">
              <p className="stat-box__label">Gesamt gescannt</p>
              <p className="stat-box__value">{sessionStats.total}</p>
            </div>

            <div className="stat-box">
              <p className="stat-box__label">Erfolgreich</p>
              <p className="stat-box__value" style={{ color: 'var(--color-success)' }}>
                {sessionStats.successful}
              </p>
            </div>

            <div className="stat-box">
              <p className="stat-box__label">Fehler</p>
              <p className="stat-box__value" style={{ color: errorCount > 0 ? 'var(--color-danger)' : undefined }}>
                {sessionStats.failed}
              </p>
            </div>

            {sessionStats.total > 0 && (
              <div className="stat-box">
                <p className="stat-box__label">Erfolgsquote</p>
                <p className="stat-box__value" style={{ color: 'var(--color-primary)' }}>
                  {sessionStats.successRate}%
                </p>
              </div>
            )}
          </div>

          <div className="session-protocol">
            <h3 className="session-protocol__title">Sitzungsprotokoll</h3>
            <div className="protocol-info">
              <div className="info-row">
                <span className="info-label">Kontext</span>
                <span className="info-value">
                  {scanContext === 'check-in' && 'Einchecken'}
                  {scanContext === 'check-out' && 'Auschecken'}
                  {scanContext === 'warehouse-store' && 'Lagerort'}
                  {scanContext === 'inventory' && 'Inventur'}
                </span>
              </div>
              {projectId && projects.find(p => p.id === projectId) && (
                <div className="info-row">
                  <span className="info-label">Projekt</span>
                  <span className="info-value">{projects.find(p => p.id === projectId)?.name}</span>
                </div>
              )}
              <div className="info-row">
                <span className="info-label">Gestartet</span>
                <span className="info-value">
                  {new Date(currentSession.started_at).toLocaleTimeString('de-DE', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ScannerPage
