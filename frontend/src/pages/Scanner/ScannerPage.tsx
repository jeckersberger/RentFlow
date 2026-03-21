import { useState, useRef, useEffect } from 'react'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import '../Equipment/Equipment.module.scss'
import './Scanner.module.scss'

interface ScanEvent {
  id: string
  barcode: string
  equipment_name?: string
  scan_type: string
  timestamp: string
}

function ScannerPage() {
  const scanInputRef = useRef<HTMLInputElement>(null)
  const [barcode, setBarcode] = useState('')
  const [scanType, setScanType] = useState('check_in')
  const [projectId, setProjectId] = useState('')
  const [recentScans, setRecentScans] = useState<ScanEvent[]>([])
  const [batchMode, setBatchMode] = useState(false)

  useEffect(() => {
    scanInputRef.current?.focus()
  }, [])

  const handleScan = async (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key !== 'Enter') return

    const scannedBarcode = barcode.trim()
    if (!scannedBarcode) return

    // Simulate scan event
    const newScan: ScanEvent = {
      id: Date.now().toString(),
      barcode: scannedBarcode,
      equipment_name: 'Beispielausrüstung',
      scan_type: scanType,
      timestamp: new Date().toISOString(),
    }

    setRecentScans((prev) => [newScan, ...prev.slice(0, 9)])
    setBarcode('')
    scanInputRef.current?.focus()
  }

  return (
    <div className="scanner-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Scanner & Bestand</h1>
          <p className="page-subtitle">
            Verwalten Sie Ein- und Ausbuchungen von Ausrüstung
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
                  { value: 'check_in', label: 'Einchecken' },
                  { value: 'check_out', label: 'Auschecken' },
                  { value: 'inventory', label: 'Inventur' },
                ]}
                value={scanType}
                onChange={(e) => setScanType(e.target.value)}
              />
            </div>

            {scanType === 'check_out' && (
              <div className="scanner-panel__section">
                <Select
                  label="Projekt (für Auschecken)"
                  options={[
                    { value: 'proj-1', label: 'Projekt A' },
                    { value: 'proj-2', label: 'Projekt B' },
                  ]}
                  value={projectId}
                  onChange={(e) => setProjectId(e.target.value)}
                  placeholder="Projekt auswählen"
                />
              </div>
            )}

            <div className="scanner-panel__section">
              <div style={{ padding: 'var(--spacing-4)', backgroundColor: 'var(--color-primary-50)', borderRadius: 'var(--radius-md)', textAlign: 'center', marginBottom: 'var(--spacing-4)' }}>
                <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-2)' }}>📷</div>
                <p style={{ margin: '0 0 var(--spacing-2) 0', color: 'var(--color-text-primary)', fontWeight: 'var(--font-weight-semibold)' }}>
                  QR-Code oder Barcode scannen
                </p>
                <p style={{ margin: 0, fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                  Verwenden Sie Ihr Gerät oder geben Sie den Code manuell ein
                </p>
              </div>

              <Input
                ref={scanInputRef}
                type="text"
                label="Barcode / QR-Code"
                placeholder="Hier eintragen oder scannen..."
                value={barcode}
                onChange={(e) => setBarcode(e.target.value)}
                onKeyPress={handleScan}
                autoFocus
              />
            </div>

            <div className="scanner-actions">
              <button className="btn btn--primary" onClick={() => scanInputRef.current?.focus()} style={{ width: '100%' }}>
                📱 Fokus setzen
              </button>
              <label className="scanner-batch-toggle">
                <input
                  type="checkbox"
                  checked={batchMode}
                  onChange={(e) => setBatchMode(e.target.checked)}
                />
                <span>Batch-Modus aktivieren</span>
              </label>
            </div>
          </div>

          {recentScans.length > 0 && (
            <div className="scanner-history">
              <h2 className="scanner-history__title">Scan-Verlauf ({recentScans.length})</h2>
              <div className="scanner-history__list">
                {recentScans.map((scan, index) => (
                  <div
                    key={scan.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 'var(--spacing-3)',
                      padding: 'var(--spacing-3)',
                      backgroundColor: index === 0 ? 'var(--color-primary-50)' : 'var(--color-bg-secondary)',
                      borderRadius: 'var(--radius-md)',
                      borderLeft: index === 0 ? '3px solid var(--color-primary)' : '3px solid var(--color-border)',
                    }}
                  >
                    <div style={{ fontSize: '1.5rem' }}>
                      {scan.scan_type === 'check_in' ? '📥' : scan.scan_type === 'check_out' ? '📤' : '📊'}
                    </div>
                    <div style={{ flex: 1 }}>
                      <p style={{ margin: '0 0 var(--spacing-1) 0', fontWeight: 'var(--font-weight-semibold)', color: 'var(--color-text-primary)' }}>
                        {scan.equipment_name || `Barcode: ${scan.barcode}`}
                      </p>
                      <p style={{ margin: 0, fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                        {scan.scan_type === 'check_in' && 'Einchecken'}
                        {scan.scan_type === 'check_out' && 'Auschecken'}
                        {scan.scan_type === 'inventory' && 'Inventur'}
                        {' • '}
                        {new Date(scan.timestamp).toLocaleTimeString('de-DE')}
                      </p>
                    </div>
                    <span style={{ fontSize: 'var(--font-size-xs)', padding: 'var(--spacing-1) var(--spacing-2)', backgroundColor: 'var(--color-success-light)', color: 'var(--color-success-dark)', borderRadius: 'var(--radius-base)' }}>
                      ✓ OK
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
              <p className="stat-box__value">{recentScans.length}</p>
            </div>

            <div className="stat-box">
              <p className="stat-box__label">Fehlgeschlagen</p>
              <p className="stat-box__value">0</p>
            </div>

            <div className="stat-box">
              <p className="stat-box__label">Im Lager</p>
              <p className="stat-box__value">145</p>
            </div>
          </div>

          <div className="scanner-panel" style={{ marginTop: 'var(--spacing-4)' }}>
            <h2 className="scanner-panel__title">Hilfe</h2>
            <ul style={{ margin: 0, paddingLeft: 'var(--spacing-4)', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)', lineHeight: '1.6' }}>
              <li>Geben Sie den Barcode ein oder scannen Sie ihn</li>
              <li>Das System erkennt automatisch die Ausrüstung</li>
              <li>Wählen Sie bei Auschecken ein Projekt aus</li>
              <li>Im Batch-Modus können Sie mehrere Artikel scannen</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ScannerPage
