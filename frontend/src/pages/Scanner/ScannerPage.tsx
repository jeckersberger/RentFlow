import { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import { Html5Qrcode, Html5QrcodeScannerState } from 'html5-qrcode'
import { Select } from '../../components/Form/Select'
import { Input } from '../../components/Form/Input'
import { equipmentApi, projectApi } from '../../services/api'
import '../Equipment/Equipment.module.scss'
import './Scanner.module.scss'

// IndexedDB Offline Queue (max 500 scans)
const OFFLINE_DB_NAME = 'rentflow-scanner-offline'
const OFFLINE_STORE_NAME = 'scan-queue'
const MAX_OFFLINE_SCANS = 500

const openOfflineDB = (): Promise<IDBDatabase> => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(OFFLINE_DB_NAME, 1)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(OFFLINE_STORE_NAME)) {
        db.createObjectStore(OFFLINE_STORE_NAME, { keyPath: 'id', autoIncrement: true })
      }
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

const addToOfflineQueue = async (scanData: { barcode: string; context: string; timestamp: string; projectId: string }) => {
  const db = await openOfflineDB()
  const tx = db.transaction(OFFLINE_STORE_NAME, 'readwrite')
  const store = tx.objectStore(OFFLINE_STORE_NAME)

  // Check count
  const countReq = store.count()
  return new Promise<boolean>((resolve) => {
    countReq.onsuccess = () => {
      if (countReq.result >= MAX_OFFLINE_SCANS) {
        resolve(false) // Queue full
      } else {
        store.add(scanData)
        resolve(true)
      }
    }
  })
}

const getOfflineQueueCount = async (): Promise<number> => {
  const db = await openOfflineDB()
  const tx = db.transaction(OFFLINE_STORE_NAME, 'readonly')
  const store = tx.objectStore(OFFLINE_STORE_NAME)
  return new Promise((resolve) => {
    const req = store.count()
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => resolve(0)
  })
}

const syncOfflineQueue = async (): Promise<number> => {
  const db = await openOfflineDB()
  const tx = db.transaction(OFFLINE_STORE_NAME, 'readwrite')
  const store = tx.objectStore(OFFLINE_STORE_NAME)
  return new Promise((resolve) => {
    const req = store.getAll()
    req.onsuccess = async () => {
      const items = req.result
      let synced = 0
      for (const item of items) {
        try {
          // Would call scannerApi.sync(item) in production
          store.delete(item.id)
          synced++
        } catch { /* retry later */ }
      }
      resolve(synced)
    }
    req.onerror = () => resolve(0)
  })
}

interface EquipmentDetail {
  id: string
  name: string
  status: string
  category_id?: string
  barcode?: string
  rental_price_day?: number
  rental_price_week?: number
  condition?: string
}

interface FeedbackMessage {
  type: 'success' | 'error'
  text: string
  timestamp: number
}

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
  const [offlineMode, setOfflineMode] = useState(false)
  const [offlineQueue, setOfflineQueue] = useState(0)
  const [scannedEquipment, setScannedEquipment] = useState<EquipmentDetail | null>(null)
  const [feedbackMessage, setFeedbackMessage] = useState<FeedbackMessage | null>(null)
  const [checkOutCount, setCheckOutCount] = useState(0)
  const [checkInCount, setCheckInCount] = useState(0)
  const [showConditionReport, setShowConditionReport] = useState(false)
  const [conditionRating, setConditionRating] = useState<string>('')
  const [conditionNotes, setConditionNotes] = useState('')
  const [conditionEquipmentId, setConditionEquipmentId] = useState<string>('')
  const [conditionEquipmentName, setConditionEquipmentName] = useState<string>('')
  const [isSubmittingCondition, setIsSubmittingCondition] = useState(false)

  // Ref to allow useEffects to call processBarcode without circular dependency
  const processBarcodeRef = useRef<((barcode: string) => void) | null>(null)

  // Audio feedback helper
  const playBeep = useCallback((type: 'success' | 'error' | 'info') => {
    try {
      const ctx = new (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)()
      const oscillator = ctx.createOscillator()
      const gain = ctx.createGain()
      oscillator.connect(gain)
      gain.connect(ctx.destination)
      gain.gain.value = 0.3

      if (type === 'success') {
        oscillator.frequency.value = 1200
        oscillator.type = 'sine'
      } else if (type === 'error') {
        oscillator.frequency.value = 300
        oscillator.type = 'square'
      } else {
        oscillator.frequency.value = 800
        oscillator.type = 'sine'
      }

      oscillator.start()
      setTimeout(() => {
        oscillator.stop()
        ctx.close()
      }, type === 'error' ? 300 : 150)
    } catch {
      // Audio not supported, fall back to vibration only
    }
  }, [])

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

  // Zebra DataWedge Integration
  useEffect(() => {
    // DataWedge broadcasts scan results as custom window events
    const handleDataWedgeScan = (event: Event) => {
      const customEvent = event as CustomEvent
      const barcode = customEvent?.detail?.data || ''
      if (barcode && !isProcessing && processBarcodeRef.current) {
        processBarcodeRef.current(barcode)
      }
    }

    // Listen for DataWedge intent broadcast
    window.addEventListener('datawedge:scan', handleDataWedgeScan)

    // Also listen for keyboard wedge mode (DataWedge can inject keystrokes)
    let keyBuffer = ''
    let keyTimer: ReturnType<typeof setTimeout> | null = null

    const handleKeyPress = (event: KeyboardEvent) => {
      // DataWedge keyboard wedge fires rapidly - detect fast input
      if (event.key === 'Enter' && keyBuffer.length > 3) {
        if (processBarcodeRef.current) processBarcodeRef.current(keyBuffer)
        keyBuffer = ''
        return
      }
      if (keyTimer) clearTimeout(keyTimer)
      keyBuffer += event.key
      keyTimer = setTimeout(() => { keyBuffer = '' }, 100) // reset after 100ms pause
    }

    window.addEventListener('keypress', handleKeyPress)

    return () => {
      window.removeEventListener('datawedge:scan', handleDataWedgeScan)
      window.removeEventListener('keypress', handleKeyPress)
    }
  }, [isProcessing])

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

  // Online/offline detection and queue syncing
  useEffect(() => {
    const updateOnlineStatus = () => {
      const isOffline = !navigator.onLine
      setOfflineMode(isOffline)
      if (!isOffline) {
        // Back online - sync queue
        syncOfflineQueue().then(synced => {
          if (synced > 0) {
            getOfflineQueueCount().then(setOfflineQueue)
          }
        })
      }
    }

    const updateQueueCount = () => {
      getOfflineQueueCount().then(setOfflineQueue)
    }

    window.addEventListener('online', updateOnlineStatus)
    window.addEventListener('offline', updateOnlineStatus)
    updateOnlineStatus()
    updateQueueCount()

    return () => {
      window.removeEventListener('online', updateOnlineStatus)
      window.removeEventListener('offline', updateOnlineStatus)
    }
  }, [])

  // WebHID USB Barcode Scanner support
  useEffect(() => {
    if (!('hid' in navigator)) return

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const handleHIDInput = (event: any) => {
      const data = event?.data
      if (!data) return
      const bytes = new Uint8Array(data.buffer)
      const barcode = String.fromCharCode(...bytes.filter((b: number) => b > 0))
      if (barcode.trim() && !isProcessing && processBarcodeRef.current) {
        processBarcodeRef.current(barcode.trim())
      }
    }

    // Auto-connect to previously paired devices
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const hid = (navigator as any).hid
    if (hid?.getDevices) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      hid.getDevices().then((devices: any[]) => {
        devices.forEach((device) => {
          if (!device.opened) {
            device.open().then(() => {
              device.addEventListener('inputreport', handleHIDInput)
            })
          }
        })
      })
    }

    return () => {
      // Cleanup handled by device disconnect
    }
  }, [isProcessing])

  // Auto-focus manual input on page load
  useEffect(() => {
    const timer = setTimeout(() => {
      scanInputRef.current?.focus()
    }, 100)
    return () => clearTimeout(timer)
  }, [])

  // Auto-clear feedback message after 3 seconds
  useEffect(() => {
    if (!feedbackMessage) return
    const timer = setTimeout(() => {
      setFeedbackMessage(null)
    }, 3000)
    return () => clearTimeout(timer)
  }, [feedbackMessage])

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

    // Play info beep on processing start
    playBeep('info')

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
      let equipment: EquipmentDetail | null = null

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

      // Store scanned equipment details for display
      setScannedEquipment(equipment)

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

      // Track check-in/check-out counts and show feedback
      if (scanContext === 'check-out') {
        setCheckOutCount((prev) => prev + 1)
        setFeedbackMessage({ type: 'success', text: `${equipment.name} erfolgreich ausgecheckt`, timestamp: Date.now() })
      } else if (scanContext === 'check-in') {
        setCheckInCount((prev) => prev + 1)
        setFeedbackMessage({ type: 'success', text: `${equipment.name} erfolgreich eingecheckt`, timestamp: Date.now() })
        // Show condition report after successful check-in
        setConditionEquipmentId(equipment.id)
        setConditionEquipmentName(equipment.name)
        setConditionRating('')
        setConditionNotes('')
        setShowConditionReport(true)
      } else {
        setFeedbackMessage({ type: 'success', text: `${equipment.name} erfolgreich gescannt`, timestamp: Date.now() })
      }

      // Vibrieren: Erfolg (kurz-kurz)
      if ('vibrate' in navigator) {
        navigator.vibrate([50, 50, 50])
      }

      // Play success beep
      playBeep('success')
    } catch (err: unknown) {
      const axiosError = (err as { response?: { data?: { error?: string; detail?: string } } })?.response?.data
      const errorMessage = axiosError?.error || axiosError?.detail ||
        (err instanceof Error ? err.message : 'Unbekannter Fehler')

      // Check if offline and add to queue instead
      if (!navigator.onLine) {
        const queued = await addToOfflineQueue({
          barcode: scannedBarcode,
          context: scanContext,
          timestamp: new Date().toISOString(),
          projectId,
        })
        if (queued) {
          getOfflineQueueCount().then(setOfflineQueue)
          const queuedScan: ScanEvent = {
            id: scanId,
            barcode: scannedBarcode,
            scan_type: scanContext,
            timestamp: new Date().toISOString(),
            status: 'pending',
            error_message: 'Offline - in Warteschlange gespeichert',
          }
          setRecentScans((prev) =>
            prev.map((s) => (s.id === scanId ? queuedScan : s))
          )
          playBeep('info')
        } else {
          const errorScan: ScanEvent = {
            ...pendingScan,
            status: 'error',
            error_message: 'Offline-Warteschlange voll',
          }
          setRecentScans((prev) =>
            prev.map((s) => (s.id === scanId ? errorScan : s))
          )
          setErrorCount((prev) => prev + 1)
          playBeep('error')
        }
      } else {
        const errorScan: ScanEvent = {
          ...pendingScan,
          status: 'error',
          error_message: errorMessage,
        }
        setRecentScans((prev) =>
          prev.map((s) => (s.id === scanId ? errorScan : s))
        )
        setErrorCount((prev) => prev + 1)
        setScannedEquipment(null)
        setFeedbackMessage({ type: 'error', text: errorMessage, timestamp: Date.now() })

        // Play error beep
        playBeep('error')
      }

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
  }, [scanContext, projectId, isProcessing, cameraActive, playBeep])

  // Keep ref in sync so useEffects can call processBarcode
  processBarcodeRef.current = processBarcode

  // Manuelle Eingabe per Enter
  const handleManualScan = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key !== 'Enter') return
    const scannedBarcode = barcode.trim()
    if (scannedBarcode) {
      processBarcode(scannedBarcode)
    }
  }

  // Manuelle Eingabe per Button
  const handleManualScanButton = () => {
    const scannedBarcode = barcode.trim()
    if (scannedBarcode) {
      processBarcode(scannedBarcode)
    }
  }

  // Status label/color helpers
  const getStatusLabel = (status: string) => {
    const map: Record<string, string> = {
      available: 'Verfügbar',
      checked_out: 'Ausgecheckt',
      in_maintenance: 'In Wartung',
      reserved: 'Reserviert',
      damaged: 'Beschädigt',
    }
    return map[status] || status
  }

  const getStatusColor = (status: string) => {
    const map: Record<string, string> = {
      available: '#16a34a',
      checked_out: '#dc2626',
      in_maintenance: '#d97706',
      reserved: '#2563eb',
      damaged: '#991b1b',
    }
    return map[status] || '#6b7280'
  }

  const getCategoryLabel = (categoryId?: string) => {
    const map: Record<string, string> = {
      'cat-audio': 'Audio',
      'cat-lighting': 'Licht',
      'cat-video': 'Video',
      'cat-stage': 'Bühne',
    }
    return categoryId ? (map[categoryId] || categoryId) : '—'
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

  const handleSubmitCondition = async () => {
    if (!conditionRating || !conditionEquipmentId) return
    setIsSubmittingCondition(true)
    try {
      await equipmentApi.updateCondition(conditionEquipmentId, conditionRating, conditionNotes || undefined)
      setFeedbackMessage({ type: 'success', text: `Zustandsbericht für ${conditionEquipmentName} gespeichert`, timestamp: Date.now() })
      playBeep('success')
    } catch {
      setFeedbackMessage({ type: 'error', text: 'Zustandsbericht konnte nicht gespeichert werden', timestamp: Date.now() })
      playBeep('error')
    } finally {
      setIsSubmittingCondition(false)
      setShowConditionReport(false)
    }
  }

  const handleSkipCondition = () => {
    setShowConditionReport(false)
    setConditionRating('')
    setConditionNotes('')
  }

  const conditionOptions: { value: string; label: string; color: string }[] = [
    { value: 'excellent', label: 'Ausgezeichnet', color: '#059669' },
    { value: 'good', label: 'Gut', color: '#16a34a' },
    { value: 'fair', label: 'Akzeptabel', color: '#d97706' },
    { value: 'poor', label: 'Schlecht', color: '#dc2626' },
    { value: 'damaged', label: 'Beschädigt', color: '#991b1b' },
  ]

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

      {/* Stats Bar */}
      {sessionActive && (
        <div
          style={{
            display: 'flex',
            gap: 'var(--spacing-4)',
            padding: 'var(--spacing-3) var(--spacing-4)',
            backgroundColor: 'var(--color-bg-secondary)',
            borderRadius: 'var(--radius-card)',
            marginBottom: 'var(--spacing-4)',
            fontWeight: 'var(--font-weight-semibold)',
            fontSize: 'var(--font-size-base)',
            flexWrap: 'wrap',
            alignItems: 'center',
          }}
        >
          <span>Heute:</span>
          <span style={{ color: '#dc2626' }}>{checkOutCount} Check-Outs</span>
          <span style={{ color: '#16a34a' }}>{checkInCount} Check-Ins</span>
          <span style={{ color: errorCount > 0 ? '#dc2626' : 'var(--color-text-secondary)' }}>{errorCount} Fehler</span>
        </div>
      )}

      {/* Feedback Toast */}
      {feedbackMessage && (
        <div
          style={{
            padding: 'var(--spacing-3) var(--spacing-4)',
            borderRadius: 'var(--radius-card)',
            marginBottom: 'var(--spacing-4)',
            fontWeight: 'var(--font-weight-semibold)',
            fontSize: 'var(--font-size-base)',
            backgroundColor: feedbackMessage.type === 'success' ? '#dcfce7' : '#fef2f2',
            color: feedbackMessage.type === 'success' ? '#166534' : '#991b1b',
            border: `2px solid ${feedbackMessage.type === 'success' ? '#16a34a' : '#dc2626'}`,
            animation: 'fadeIn 0.2s ease',
          }}
        >
          {feedbackMessage.type === 'success' ? '✓ ' : '✗ '}
          {feedbackMessage.text}
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

              {/* Manuelle Eingabe - prominent */}
              <div style={{ display: 'flex', gap: 'var(--spacing-2)', alignItems: 'flex-end' }}>
                <div style={{ flex: 1 }}>
                  <Input
                    ref={scanInputRef}
                    type="text"
                    label="Barcode / QR-Code manuell eingeben"
                    placeholder="Code eingeben und Enter drücken..."
                    value={barcode}
                    onChange={(e) => setBarcode(e.target.value)}
                    onKeyPress={handleManualScan}
                    autoFocus
                    disabled={isProcessing}
                    style={{ fontSize: '1.25rem', padding: 'var(--spacing-3) var(--spacing-4)', height: '3.25rem' }}
                  />
                </div>
                <button
                  className="btn btn--primary"
                  onClick={handleManualScanButton}
                  disabled={isProcessing || !barcode.trim()}
                  style={{
                    height: '3.25rem',
                    padding: '0 var(--spacing-6)',
                    fontSize: '1rem',
                    fontWeight: 'var(--font-weight-semibold)',
                    whiteSpace: 'nowrap',
                  }}
                >
                  Scannen
                </button>
              </div>

              {/* Project warning for check-out */}
              {scanContext === 'check-out' && !projectId && (
                <div
                  style={{
                    marginTop: 'var(--spacing-2)',
                    padding: 'var(--spacing-2) var(--spacing-3)',
                    backgroundColor: '#fffbeb',
                    color: '#92400e',
                    borderRadius: 'var(--radius-md)',
                    fontSize: 'var(--font-size-sm)',
                    fontWeight: 'var(--font-weight-semibold)',
                    border: '1px solid #d97706',
                  }}
                >
                  ⚠ Bitte zuerst ein Projekt auswählen, bevor Sie Equipment auschecken.
                </div>
              )}
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

            {/* Scanned Equipment Details */}
            {scannedEquipment && !isProcessing && (
              <div
                style={{
                  marginTop: 'var(--spacing-4)',
                  padding: 'var(--spacing-5)',
                  backgroundColor: 'var(--color-bg-secondary)',
                  borderRadius: 'var(--radius-card)',
                  border: '2px solid var(--color-primary)',
                }}
              >
                <h3 style={{ margin: '0 0 var(--spacing-3) 0', fontSize: '1.5rem', color: 'var(--color-text-primary)' }}>
                  {scannedEquipment.name}
                </h3>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--spacing-2)', marginBottom: 'var(--spacing-4)' }}>
                  <div>
                    <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>Status</span>
                    <div>
                      <span
                        style={{
                          display: 'inline-block',
                          padding: 'var(--spacing-1) var(--spacing-3)',
                          borderRadius: 'var(--radius-base)',
                          backgroundColor: getStatusColor(scannedEquipment.status) + '20',
                          color: getStatusColor(scannedEquipment.status),
                          fontWeight: 'var(--font-weight-semibold)',
                          fontSize: 'var(--font-size-sm)',
                        }}
                      >
                        {getStatusLabel(scannedEquipment.status)}
                      </span>
                    </div>
                  </div>
                  <div>
                    <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>Kategorie</span>
                    <div style={{ fontWeight: 'var(--font-weight-semibold)' }}>{getCategoryLabel(scannedEquipment.category_id)}</div>
                  </div>
                  <div>
                    <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>Barcode</span>
                    <div style={{ fontFamily: 'monospace', fontWeight: 'var(--font-weight-semibold)' }}>{scannedEquipment.barcode || '—'}</div>
                  </div>
                  <div>
                    <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>Mietpreis</span>
                    <div style={{ fontWeight: 'var(--font-weight-semibold)' }}>
                      {scannedEquipment.rental_price_day != null ? `${scannedEquipment.rental_price_day} €/Tag` : '—'}
                      {scannedEquipment.rental_price_week != null && ` · ${scannedEquipment.rental_price_week} €/Woche`}
                    </div>
                  </div>
                </div>

                {/* Action button based on context */}
                {scanContext === 'check-out' && scannedEquipment.status === 'available' && (
                  <button
                    className="btn btn--primary"
                    style={{ width: '100%', padding: 'var(--spacing-3)', fontSize: '1.1rem', fontWeight: 'var(--font-weight-semibold)' }}
                    onClick={() => {
                      if (projectId) {
                        equipmentApi.checkOut(scannedEquipment.id, projectId).then(() => {
                          setFeedbackMessage({ type: 'success', text: `${scannedEquipment.name} ausgecheckt`, timestamp: Date.now() })
                          setScannedEquipment({ ...scannedEquipment, status: 'checked_out' })
                        }).catch(() => {
                          setFeedbackMessage({ type: 'error', text: 'Check-Out fehlgeschlagen', timestamp: Date.now() })
                        })
                      }
                    }}
                    disabled={!projectId}
                  >
                    📤 Check-Out: {scannedEquipment.name}
                  </button>
                )}
                {scanContext === 'check-in' && scannedEquipment.status === 'checked_out' && (
                  <button
                    className="btn btn--primary"
                    style={{ width: '100%', padding: 'var(--spacing-3)', fontSize: '1.1rem', fontWeight: 'var(--font-weight-semibold)', backgroundColor: '#16a34a', borderColor: '#16a34a' }}
                    onClick={() => {
                      equipmentApi.checkIn(scannedEquipment.id).then(() => {
                        setFeedbackMessage({ type: 'success', text: `${scannedEquipment.name} eingecheckt`, timestamp: Date.now() })
                        setScannedEquipment({ ...scannedEquipment, status: 'available' })
                      }).catch(() => {
                        setFeedbackMessage({ type: 'error', text: 'Check-In fehlgeschlagen', timestamp: Date.now() })
                      })
                    }}
                  >
                    📥 Check-In: {scannedEquipment.name}
                  </button>
                )}
              </div>
            )}

            {/* Condition Report after Check-In */}
            {showConditionReport && (
              <div
                style={{
                  marginTop: 'var(--spacing-4)',
                  padding: 'var(--spacing-5)',
                  backgroundColor: '#f0f9ff',
                  borderRadius: 'var(--radius-card)',
                  border: '2px solid #0284c7',
                }}
              >
                <h3 style={{ margin: '0 0 var(--spacing-2) 0', fontSize: '1.25rem', color: 'var(--color-text-primary)' }}>
                  Zustandsbericht
                </h3>
                <p style={{ margin: '0 0 var(--spacing-4) 0', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                  {conditionEquipmentName} — Zustand nach Rückgabe bewerten
                </p>

                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--spacing-2)', marginBottom: 'var(--spacing-4)' }}>
                  {conditionOptions.map((opt) => (
                    <button
                      key={opt.value}
                      onClick={() => setConditionRating(opt.value)}
                      style={{
                        padding: 'var(--spacing-2) var(--spacing-4)',
                        borderRadius: 'var(--radius-base)',
                        border: conditionRating === opt.value ? `2px solid ${opt.color}` : '2px solid var(--color-border)',
                        backgroundColor: conditionRating === opt.value ? opt.color + '20' : 'var(--color-bg-primary)',
                        color: conditionRating === opt.value ? opt.color : 'var(--color-text-primary)',
                        fontWeight: conditionRating === opt.value ? 'var(--font-weight-semibold)' : 'var(--font-weight-normal)',
                        cursor: 'pointer',
                        fontSize: 'var(--font-size-base)',
                        transition: 'all 0.15s ease',
                      }}
                    >
                      {opt.label}
                    </button>
                  ))}
                </div>

                <div style={{ marginBottom: 'var(--spacing-4)' }}>
                  <label
                    style={{
                      display: 'block',
                      fontSize: 'var(--font-size-sm)',
                      fontWeight: 'var(--font-weight-semibold)',
                      color: 'var(--color-text-secondary)',
                      marginBottom: 'var(--spacing-1)',
                    }}
                  >
                    Anmerkungen (optional)
                  </label>
                  <textarea
                    value={conditionNotes}
                    onChange={(e) => setConditionNotes(e.target.value)}
                    placeholder="z.B. Kratzer an der linken Seite..."
                    rows={3}
                    style={{
                      width: '100%',
                      padding: 'var(--spacing-2) var(--spacing-3)',
                      borderRadius: 'var(--radius-md)',
                      border: '1px solid var(--color-border)',
                      fontSize: 'var(--font-size-base)',
                      fontFamily: 'inherit',
                      resize: 'vertical',
                      boxSizing: 'border-box',
                    }}
                  />
                </div>

                <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
                  <button
                    className="btn btn--primary"
                    onClick={handleSubmitCondition}
                    disabled={!conditionRating || isSubmittingCondition}
                    style={{
                      flex: 1,
                      padding: 'var(--spacing-3)',
                      fontSize: '1rem',
                      fontWeight: 'var(--font-weight-semibold)',
                    }}
                  >
                    {isSubmittingCondition ? 'Wird gespeichert...' : 'Speichern'}
                  </button>
                  <button
                    className="btn btn--secondary"
                    onClick={handleSkipCondition}
                    disabled={isSubmittingCondition}
                    style={{
                      flex: 1,
                      padding: 'var(--spacing-3)',
                      fontSize: '1rem',
                      fontWeight: 'var(--font-weight-semibold)',
                    }}
                  >
                    Überspringen
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* Scan-Verlauf */}
          {recentScans.length > 0 && (
            <div className="scanner-history">
              <h2 className="scanner-history__title">
                Scan-Verlauf (letzte {Math.min(recentScans.length, 20)})
              </h2>
              <div className="scanner-history__list">
                {recentScans.slice(0, 20).map((scan, index) => (
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
                        : scan.scan_type === 'check-in'
                        ? '📥'
                        : scan.scan_type === 'check-out'
                        ? '📤'
                        : '📊'}
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <p
                        style={{
                          margin: '0 0 var(--spacing-1) 0',
                          fontWeight: 'var(--font-weight-semibold)',
                          color: 'var(--color-text-primary)',
                          fontSize: scan.equipment_name ? 'var(--font-size-base)' : 'var(--font-size-sm)',
                        }}
                      >
                        {scan.equipment_name || `Code: ${scan.barcode}`}
                      </p>
                      <div
                        style={{
                          display: 'flex',
                          gap: 'var(--spacing-2)',
                          alignItems: 'center',
                          flexWrap: 'wrap',
                          fontSize: 'var(--font-size-sm)',
                          color:
                            scan.status === 'error'
                              ? 'var(--color-error-dark, #991b1b)'
                              : 'var(--color-text-secondary)',
                        }}
                      >
                        {scan.status === 'error'
                          ? <span>{scan.error_message}</span>
                          : scan.status === 'pending'
                          ? <span>Wird verarbeitet...</span>
                          : (
                            <>
                              <span style={{ fontWeight: 'var(--font-weight-semibold)' }}>
                                {scan.scan_type === 'check-in' && 'Check-In'}
                                {scan.scan_type === 'check-out' && 'Check-Out'}
                                {scan.scan_type === 'warehouse-store' && 'Lagerort'}
                                {scan.scan_type === 'inventory' && 'Inventur'}
                              </span>
                              <span>•</span>
                              <span>{new Date(scan.timestamp).toLocaleTimeString('de-DE')}</span>
                              {scan.barcode && (
                                <>
                                  <span>•</span>
                                  <span style={{ fontFamily: 'monospace', fontSize: 'var(--font-size-xs)' }}>{scan.barcode}</span>
                                </>
                              )}
                            </>
                          )}
                      </div>
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
