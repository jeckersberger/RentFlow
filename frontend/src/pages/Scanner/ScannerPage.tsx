import { useState, useRef, useEffect, useCallback } from 'react'
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

function ScannerPage() {
  const scanInputRef = useRef<HTMLInputElement>(null)
  const html5QrcodeRef = useRef<Html5Qrcode | null>(null)
  const [barcode, setBarcode] = useState('')
  const [scanType, setScanType] = useState('check_in')
  const [projectId, setProjectId] = useState('')
  const [projects, setProjects] = useState<Project[]>([])
  const [recentScans, setRecentScans] = useState<ScanEvent[]>([])
  const [batchMode, setBatchMode] = useState(false)
  const [cameraActive, setCameraActive] = useState(false)
  const [cameraError, setCameraError] = useState<string | null>(null)
  const [isProcessing, setIsProcessing] = useState(false)
  const [scanCount, setScanCount] = useState(0)
  const [errorCount, setErrorCount] = useState(0)

  // Projekte für Check-Out laden
  useEffect(() => {
    projectApi.list(1, 100)
      .then((data) => {
        const projectList = (data?.data || []).filter(
          (p: Project) => p.status === 'confirmed' || p.status === 'in_progress'
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
      scan_type: scanType,
      timestamp: new Date().toISOString(),
      status: 'pending',
    }
    setRecentScans((prev) => [pendingScan, ...prev.slice(0, 49)])

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

      // Schritt 2: Aktion basierend auf Scan-Typ ausführen
      if (scanType === 'check_out') {
        if (!projectId) {
          throw new Error('Bitte wählen Sie ein Projekt für das Auschecken')
        }
        await equipmentApi.checkOut(equipment.id, projectId)
      } else if (scanType === 'check_in') {
        await equipmentApi.checkIn(equipment.id)
      }
      // Bei 'inventory' nur Equipment-Lookup, keine Aktion

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
      setScanCount((prev) => prev + 1)

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
  }, [scanType, projectId, isProcessing, cameraActive])

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

  return (
    <div className="scanner-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Scanner & Bestand</h1>
          <p className="page-subtitle">
            Scannen Sie Equipment per Kamera oder geben Sie den Code manuell ein
          </p>
        </div>
      </div>

      <div className="scanner-container">
        <div className="scanner-main">
          <div className="scanner-panel">
            <h2 className="scanner-panel__title">Ausrüstung scannen</h2>

            <div className="scanner-panel__section">
              <Select
                label="Scantyp"
                options={[
                  { value: 'check_in', label: 'Einchecken (Rückgabe)' },
                  { value: 'check_out', label: 'Auschecken (Projekt zuweisen)' },
                  { value: 'inventory', label: 'Inventur (nur prüfen)' },
                ]}
                value={scanType}
                onChange={(e) => setScanType(e.target.value)}
              />
            </div>

            {scanType === 'check_out' && (
              <div className="scanner-panel__section">
                <Select
                  label="Projekt (für Auschecken)"
                  options={projects.map((p) => ({
                    value: p.id,
                    label: p.name,
                  }))}
                  value={projectId}
                  onChange={(e) => setProjectId(e.target.value)}
                  placeholder="Projekt auswählen"
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
          <div className="scanner-stats">
            <div className="stat-box">
              <p className="stat-box__label">Heute gescannt</p>
              <p className="stat-box__value">{scanCount}</p>
            </div>

            <div className="stat-box">
              <p className="stat-box__label">Fehlgeschlagen</p>
              <p className="stat-box__value" style={{ color: errorCount > 0 ? 'var(--color-error)' : undefined }}>
                {errorCount}
              </p>
            </div>

            <div className="stat-box">
              <p className="stat-box__label">Im Verlauf</p>
              <p className="stat-box__value">{recentScans.length}</p>
            </div>
          </div>

          <div
            className="scanner-panel"
            style={{ marginTop: 'var(--spacing-4)' }}
          >
            <h2 className="scanner-panel__title">Hilfe</h2>
            <ul
              style={{
                margin: 0,
                paddingLeft: 'var(--spacing-4)',
                fontSize: 'var(--font-size-sm)',
                color: 'var(--color-text-secondary)',
                lineHeight: '1.6',
              }}
            >
              <li>Tippen Sie auf den Kamera-Bereich zum Scannen</li>
              <li>QR-Codes und Barcodes werden automatisch erkannt</li>
              <li>Oder geben Sie den Code manuell ein</li>
              <li>Bei Auschecken muss ein Projekt ausgewählt sein</li>
              <li>Im Batch-Modus bleibt die Kamera nach dem Scan aktiv</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ScannerPage
