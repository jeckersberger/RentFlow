import { useState, useRef, useEffect, useCallback } from 'react'
import { Html5Qrcode, Html5QrcodeScannerState } from 'html5-qrcode'
import { equipmentApi, projectApi } from '../../services/api'
import { useIndustry } from '../../hooks/useIndustry'
import './Scanner.scss'

// ─── Offline Queue (IndexedDB) ───────────────────────────────────────────────
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
  const countReq = store.count()
  return new Promise<boolean>((resolve) => {
    countReq.onsuccess = () => {
      if (countReq.result >= MAX_OFFLINE_SCANS) {
        resolve(false)
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
          store.delete(item.id)
          synced++
        } catch { /* retry later */ }
      }
      resolve(synced)
    }
    req.onerror = () => resolve(0)
  })
}

// ─── Types ───────────────────────────────────────────────────────────────────

interface ScanResult {
  id: string
  barcode: string
  equipmentName?: string
  status: 'success' | 'error' | 'pending'
  message?: string
  timestamp: string
}

interface Project {
  id: string
  name: string
  status: string
}

type ScanContext = 'check-in' | 'check-out' | 'warehouse-store' | 'inventory' | 'pack-verify'

// ─── Component ───────────────────────────────────────────────────────────────

function ScannerPage() {
  const { label } = useIndustry()
  const scanInputRef = useRef<HTMLInputElement>(null)
  const html5QrcodeRef = useRef<Html5Qrcode | null>(null)
  const processBarcodeRef = useRef<((barcode: string) => void) | null>(null)

  // Industry-aware context labels (fallback to defaults)
  const contextLabels: Record<ScanContext, string> = {
    'check-in': label('checkin') || 'Check-In',
    'check-out': label('checkout') || 'Check-Out',
    'warehouse-store': 'Einlagern',
    'inventory': 'Inventur',
    'pack-verify': 'Packliste',
  }

  // Core state
  const [barcode, setBarcode] = useState('')
  const [scanContext, setScanContext] = useState<ScanContext>('check-in')
  const [projectId, setProjectId] = useState('')
  const [projects, setProjects] = useState<Project[]>([])
  const [recentScans, setRecentScans] = useState<ScanResult[]>([])
  const [isProcessing, setIsProcessing] = useState(false)
  const [scanCount, setScanCount] = useState(0)
  const [flashType, setFlashType] = useState<'success' | 'error' | null>(null)

  // Advanced options (collapsed by default)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [cameraActive, setCameraActive] = useState(false)
  const [cameraError, setCameraError] = useState<string | null>(null)
  const [batchMode, setBatchMode] = useState(false)
  const [conditionEnabled, setConditionEnabled] = useState(false)
  const [conditionRating, setConditionRating] = useState(0)
  const [conditionEquipmentId, setConditionEquipmentId] = useState('')
  const [showConditionInline, setShowConditionInline] = useState(false)

  // Offline
  const [offlineMode, setOfflineMode] = useState(false)
  const [offlineQueue, setOfflineQueue] = useState(0)

  // ─── Audio feedback ──────────────────────────────────────────────────────
  const playBeep = useCallback((type: 'success' | 'error' | 'info') => {
    try {
      const ctx = new (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)()
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()
      osc.connect(gain)
      gain.connect(ctx.destination)
      gain.gain.value = 0.3
      if (type === 'success') { osc.frequency.value = 1200; osc.type = 'sine' }
      else if (type === 'error') { osc.frequency.value = 300; osc.type = 'square' }
      else { osc.frequency.value = 800; osc.type = 'sine' }
      osc.start()
      setTimeout(() => { osc.stop(); ctx.close() }, type === 'error' ? 300 : 150)
    } catch { /* Audio not supported */ }
  }, [])

  // ─── Load projects ──────────────────────────────────────────────────────
  useEffect(() => {
    projectApi.list(1, 100)
      .then((data) => {
        const all = data?.items || data?.data || []
        setProjects(all.filter(
          (p: Project) => ['confirmed', 'in_progress', 'active', 'planning', 'draft'].includes(p.status)
        ))
      })
      .catch(() => setProjects([]))
  }, [])

  // ─── Online/offline detection ───────────────────────────────────────────
  useEffect(() => {
    const update = () => {
      const isOff = !navigator.onLine
      setOfflineMode(isOff)
      if (!isOff) {
        syncOfflineQueue().then(synced => {
          if (synced > 0) getOfflineQueueCount().then(setOfflineQueue)
        })
      }
    }
    window.addEventListener('online', update)
    window.addEventListener('offline', update)
    update()
    getOfflineQueueCount().then(setOfflineQueue)
    return () => { window.removeEventListener('online', update); window.removeEventListener('offline', update) }
  }, [])

  // ─── DataWedge / keyboard wedge ─────────────────────────────────────────
  useEffect(() => {
    const handleDataWedgeScan = (event: Event) => {
      const barcode = (event as CustomEvent)?.detail?.data || ''
      if (barcode && !isProcessing && processBarcodeRef.current) {
        processBarcodeRef.current(barcode)
      }
    }
    window.addEventListener('datawedge:scan', handleDataWedgeScan)

    let keyBuffer = ''
    let keyTimer: ReturnType<typeof setTimeout> | null = null
    const handleKeyPress = (event: KeyboardEvent) => {
      // Ignore if user is typing in the input field
      if (document.activeElement === scanInputRef.current) return
      if (event.key === 'Enter' && keyBuffer.length > 3) {
        if (processBarcodeRef.current) processBarcodeRef.current(keyBuffer)
        keyBuffer = ''
        return
      }
      if (keyTimer) clearTimeout(keyTimer)
      keyBuffer += event.key
      keyTimer = setTimeout(() => { keyBuffer = '' }, 100)
    }
    window.addEventListener('keypress', handleKeyPress)

    return () => {
      window.removeEventListener('datawedge:scan', handleDataWedgeScan)
      window.removeEventListener('keypress', handleKeyPress)
    }
  }, [isProcessing])

  // ─── WebHID USB scanner ─────────────────────────────────────────────────
  useEffect(() => {
    if (!('hid' in navigator)) return
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const handleHIDInput = (event: any) => {
      const data = event?.data
      if (!data) return
      const bytes = new Uint8Array(data.buffer)
      const bc = String.fromCharCode(...bytes.filter((b: number) => b > 0))
      if (bc.trim() && !isProcessing && processBarcodeRef.current) {
        processBarcodeRef.current(bc.trim())
      }
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const hid = (navigator as any).hid
    if (hid?.getDevices) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      hid.getDevices().then((devices: any[]) => {
        devices.forEach((device) => {
          if (!device.opened) {
            device.open().then(() => { device.addEventListener('inputreport', handleHIDInput) })
          }
        })
      })
    }
  }, [isProcessing])

  // ─── Auto-focus input ───────────────────────────────────────────────────
  useEffect(() => {
    const t = setTimeout(() => scanInputRef.current?.focus(), 100)
    return () => clearTimeout(t)
  }, [])

  // ─── Cleanup camera on unmount ──────────────────────────────────────────
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

  // ─── Process barcode ────────────────────────────────────────────────────
  const processBarcode = useCallback(async (scannedBarcode: string) => {
    if (!scannedBarcode || isProcessing) return
    setIsProcessing(true)
    const scanId = Date.now().toString()

    if ('vibrate' in navigator) navigator.vibrate(100)
    playBeep('info')

    // Add pending entry
    const pending: ScanResult = {
      id: scanId,
      barcode: scannedBarcode,
      status: 'pending',
      timestamp: new Date().toISOString(),
    }
    setRecentScans(prev => [pending, ...prev.slice(0, 49)])

    try {
      // Step 1: Find equipment
      let equipment: { id: string; name: string; status: string } | null = null
      const qrMatch = scannedBarcode.match(/^rentflow:\/\/equipment\/(.+)$/)
      const isRfidTag = /^[0-9A-Fa-f]{24,}$/.test(scannedBarcode)

      if (qrMatch) {
        equipment = await equipmentApi.getById(qrMatch[1])
      } else if (isRfidTag) {
        equipment = await equipmentApi.getByRfidTag(scannedBarcode)
      } else {
        equipment = await equipmentApi.getByBarcode(scannedBarcode)
      }

      if (!equipment) throw new Error('Equipment nicht gefunden')

      // Step 2: Execute action based on context
      if (scanContext === 'check-out') {
        if (!projectId) throw new Error(`Bitte ${label('project') || 'Projekt'} auswaehlen`)
        await equipmentApi.checkOut(equipment.id, projectId)
      } else if (scanContext === 'check-in') {
        await equipmentApi.checkIn(equipment.id)
      } else if (scanContext === 'warehouse-store') {
        await new Promise(r => setTimeout(r, 200))
      } else if (scanContext === 'inventory') {
        await new Promise(r => setTimeout(r, 100))
      } else if (scanContext === 'pack-verify') {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const verifyFn = (window as any).__packingListVerify as ((barcode: string) => { found: boolean; item?: { name: string } }) | undefined
        if (verifyFn) {
          const result = verifyFn(scannedBarcode)
          if (!result.found) throw new Error('Nicht in dieser Packliste')
        } else {
          await new Promise(r => setTimeout(r, 100))
        }
      }

      // Success
      const success: ScanResult = {
        id: scanId,
        barcode: scannedBarcode,
        equipmentName: equipment.name,
        status: 'success',
        message: contextLabels[scanContext],
        timestamp: new Date().toISOString(),
      }
      setRecentScans(prev => prev.map(s => s.id === scanId ? success : s))
      setScanCount(prev => prev + 1)

      setFlashType('success')
      setTimeout(() => setFlashType(null), 600)
      playBeep('success')
      if ('vibrate' in navigator) navigator.vibrate([50, 50, 50])

      // Condition report inline (if enabled)
      if (conditionEnabled && scanContext === 'check-in') {
        setConditionEquipmentId(equipment.id)
        setConditionRating(0)
        setShowConditionInline(true)
      }

    } catch (err: unknown) {
      const axiosError = (err as { response?: { data?: { error?: string; detail?: string } } })?.response?.data
      const errorMessage = axiosError?.error || axiosError?.detail ||
        (err instanceof Error ? err.message : 'Unbekannter Fehler')

      if (!navigator.onLine) {
        const queued = await addToOfflineQueue({
          barcode: scannedBarcode,
          context: scanContext,
          timestamp: new Date().toISOString(),
          projectId,
        })
        if (queued) {
          getOfflineQueueCount().then(setOfflineQueue)
          setRecentScans(prev => prev.map(s => s.id === scanId ? {
            ...s, status: 'pending' as const, message: 'Offline - in Warteschlange',
          } : s))
          playBeep('info')
        } else {
          setRecentScans(prev => prev.map(s => s.id === scanId ? {
            ...s, status: 'error' as const, message: 'Offline-Queue voll',
          } : s))
          playBeep('error')
        }
      } else {
        setRecentScans(prev => prev.map(s => s.id === scanId ? {
          ...s, status: 'error' as const, message: errorMessage,
        } : s))
        setFlashType('error')
        setTimeout(() => setFlashType(null), 600)
        playBeep('error')
      }
      if ('vibrate' in navigator) navigator.vibrate(300)
    } finally {
      setIsProcessing(false)
      setBarcode('')
      if (!cameraActive) scanInputRef.current?.focus()
    }
  }, [scanContext, projectId, isProcessing, cameraActive, playBeep, conditionEnabled])

  processBarcodeRef.current = processBarcode

  // ─── Input handlers ─────────────────────────────────────────────────────
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && barcode.trim()) {
      processBarcode(barcode.trim())
    }
  }

  // ─── Camera toggle ──────────────────────────────────────────────────────
  const toggleCamera = async () => {
    if (cameraActive) {
      if (html5QrcodeRef.current) {
        try { await html5QrcodeRef.current.stop() } catch { /* ignore */ }
      }
      setCameraActive(false)
      setCameraError(null)
      return
    }
    setCameraError(null)
    try {
      if (!html5QrcodeRef.current) {
        html5QrcodeRef.current = new Html5Qrcode('scanner-camera-view')
      }
      await html5QrcodeRef.current.start(
        { facingMode: 'environment' },
        { fps: 10, qrbox: { width: 250, height: 250 }, aspectRatio: 1.0 },
        (decodedText) => {
          processBarcode(decodedText)
          if (!batchMode && html5QrcodeRef.current) {
            html5QrcodeRef.current.stop().catch(() => {})
            setCameraActive(false)
          }
        },
        () => {}
      )
      setCameraActive(true)
    } catch (err) {
      setCameraError(err instanceof Error ? err.message : 'Kamera konnte nicht gestartet werden')
      setCameraActive(false)
    }
  }

  // ─── Condition report submit ────────────────────────────────────────────
  const submitCondition = async () => {
    if (!conditionRating || !conditionEquipmentId) return
    const ratingMap: Record<number, string> = { 1: 'damaged', 2: 'poor', 3: 'fair', 4: 'good', 5: 'excellent' }
    try {
      await equipmentApi.updateCondition(conditionEquipmentId, ratingMap[conditionRating])
      playBeep('success')
    } catch {
      playBeep('error')
    }
    setShowConditionInline(false)
    setConditionRating(0)
  }

  const showProjectSelector = scanContext === 'check-out' || scanContext === 'pack-verify'

  // ─── Render ─────────────────────────────────────────────────────────────
  return (
    <div className="scanner-page-simple">
      {/* Flash overlay */}
      {flashType && <div className={`scan-flash scan-flash--${flashType}`} />}

      {/* Offline banner */}
      {offlineMode && (
        <div className="scanner-offline-banner">
          Offline-Modus — {offlineQueue} Scans in Warteschlange
        </div>
      )}

      {/* Top bar: context + project dropdowns */}
      <div className="scanner-top-bar">
        <div className="scanner-dropdown-group">
          <select
            className="scanner-select"
            value={scanContext}
            onChange={(e) => setScanContext(e.target.value as ScanContext)}
          >
            {(Object.keys(contextLabels) as ScanContext[]).map(ctx => (
              <option key={ctx} value={ctx}>{contextLabels[ctx]}</option>
            ))}
          </select>

          {showProjectSelector && (
            <select
              className="scanner-select"
              value={projectId}
              onChange={(e) => setProjectId(e.target.value)}
            >
              <option value="">{label('project') || 'Projekt'}...</option>
              {projects.map(p => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          )}
        </div>

        {offlineQueue > 0 && !offlineMode && (
          <span className="scanner-queue-badge">{offlineQueue} in Queue</span>
        )}
      </div>

      {/* Warning: no project selected */}
      {showProjectSelector && !projectId && (
        <div className="scanner-warning">
          Bitte {label('project') || 'Projekt'} auswaehlen bevor du scannst.
        </div>
      )}

      {/* Scan area */}
      <div className="scanner-scan-area">
        {/* Camera view (hidden when not active) */}
        <div
          id="scanner-camera-view"
          className="scanner-camera"
          style={{ display: cameraActive ? 'block' : 'none' }}
        />

        {!cameraActive && (
          <div className="scanner-scan-prompt">
            <span className="scanner-scan-icon">{'{ }'}</span>
            <span className="scanner-scan-label">Barcode scannen oder eingeben</span>
          </div>
        )}

        {cameraError && (
          <div className="scanner-camera-error">Kamera-Fehler: {cameraError}</div>
        )}

        <div className="scanner-input-row">
          <input
            ref={scanInputRef}
            type="text"
            className="scanner-barcode-input"
            placeholder="Barcode / QR-Code hier eingeben..."
            value={barcode}
            onChange={(e) => setBarcode(e.target.value)}
            onKeyDown={handleKeyDown}
            autoFocus
            disabled={isProcessing}
          />
          <button
            className="scanner-submit-btn"
            onClick={() => barcode.trim() && processBarcode(barcode.trim())}
            disabled={isProcessing || !barcode.trim()}
          >
            {isProcessing ? '...' : 'Scan'}
          </button>
        </div>
      </div>

      {/* Inline condition report (star rating) */}
      {showConditionInline && (
        <div className="scanner-condition-inline">
          <span className="scanner-condition-label">Zustand bewerten:</span>
          <div className="scanner-stars">
            {[1, 2, 3, 4, 5].map(star => (
              <button
                key={star}
                className={`scanner-star ${conditionRating >= star ? 'scanner-star--active' : ''}`}
                onClick={() => setConditionRating(star)}
              >
                {conditionRating >= star ? '\u2605' : '\u2606'}
              </button>
            ))}
          </div>
          <button
            className="scanner-condition-save"
            onClick={submitCondition}
            disabled={conditionRating === 0}
          >
            OK
          </button>
          <button
            className="scanner-condition-skip"
            onClick={() => { setShowConditionInline(false); setConditionRating(0) }}
          >
            Skip
          </button>
        </div>
      )}

      {/* Recent scan results */}
      {recentScans.length > 0 && (
        <div className="scanner-results">
          {recentScans.slice(0, 20).map((scan) => (
            <div key={scan.id} className={`scanner-result-item scanner-result-item--${scan.status}`}>
              <span className="scanner-result-icon">
                {scan.status === 'success' ? '\u2705' : scan.status === 'error' ? '\u274C' : '\u23F3'}
              </span>
              <span className="scanner-result-text">
                {scan.equipmentName || scan.barcode}
                {scan.message && <span className="scanner-result-detail"> — {scan.message}</span>}
              </span>
              <span className="scanner-result-time">
                {new Date(scan.timestamp).toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
              </span>
            </div>
          ))}
        </div>
      )}

      {/* Simple counter */}
      <div className="scanner-counter">
        Gescannt: <strong>{scanCount}</strong>
      </div>

      {/* Advanced options (collapsed) */}
      <button
        className="scanner-advanced-toggle"
        onClick={() => setShowAdvanced(!showAdvanced)}
      >
        {showAdvanced ? '\u25B2' : '\u25BC'} Erweiterte Optionen
      </button>

      {showAdvanced && (
        <div className="scanner-advanced">
          <div className="scanner-advanced-row">
            <button
              className={`scanner-adv-btn ${cameraActive ? 'scanner-adv-btn--active' : ''}`}
              onClick={toggleCamera}
            >
              {cameraActive ? 'Kamera stoppen' : 'Kamera starten'}
            </button>
          </div>

          <label className="scanner-advanced-check">
            <input
              type="checkbox"
              checked={batchMode}
              onChange={(e) => setBatchMode(e.target.checked)}
            />
            Batch-Modus (Kamera bleibt aktiv)
          </label>

          <label className="scanner-advanced-check">
            <input
              type="checkbox"
              checked={conditionEnabled}
              onChange={(e) => setConditionEnabled(e.target.checked)}
            />
            Zustandsbericht bei {label('checkin') || 'Check-In'} (1-5 Sterne)
          </label>

          {recentScans.length > 5 && (
            <div className="scanner-history-section">
              <h4 className="scanner-history-title">Scan-Verlauf ({recentScans.length})</h4>
              <div className="scanner-history-list">
                {recentScans.map((scan) => (
                  <div key={scan.id} className="scanner-history-item">
                    <span>{scan.status === 'success' ? '\u2713' : scan.status === 'error' ? '\u2717' : '...'}</span>
                    <span>{scan.equipmentName || scan.barcode}</span>
                    <span className="scanner-history-time">
                      {new Date(scan.timestamp).toLocaleTimeString('de-DE')}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default ScannerPage
