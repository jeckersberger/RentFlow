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
              <Input
                ref={scanInputRef}
                type="text"
                label="Barcode"
                placeholder="Barcode scannen oder manuell eingeben..."
                value={barcode}
                onChange={(e) => setBarcode(e.target.value)}
                onKeyPress={handleScan}
                autoFocus
              />
            </div>

            <div className="scanner-actions">
              <button className="btn btn--primary" onClick={() => scanInputRef.current?.focus()}>
                📱 Kamerascan
              </button>
              <label className="scanner-batch-toggle">
                <input
                  type="checkbox"
                  checked={batchMode}
                  onChange={(e) => setBatchMode(e.target.checked)}
                />
                Batch-Modus
              </label>
            </div>
          </div>

          {recentScans.length > 0 && (
            <div className="scanner-history">
              <h2 className="scanner-history__title">Letzte Scans</h2>
              <div className="scanner-history__list">
                {recentScans.map((scan) => (
                  <div key={scan.id} className="scan-item">
                    <div className="scan-item__icon">
                      {scan.scan_type === 'check_in' ? '📥' : '📤'}
                    </div>
                    <div className="scan-item__content">
                      <p className="scan-item__name">
                        {scan.equipment_name || scan.barcode}
                      </p>
                      <p className="scan-item__meta">
                        {scan.scan_type === 'check_in' ? 'Einchecken' : 'Auschecken'} •{' '}
                        {new Date(scan.timestamp).toLocaleTimeString('de-DE')}
                      </p>
                    </div>
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
