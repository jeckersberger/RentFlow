import { useState, useCallback, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import Papa from 'papaparse'
import { equipmentApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'

// ---------- Types ----------
type Step = 'upload' | 'preview' | 'validation' | 'importing' | 'result'

interface ColumnMapping {
  csvColumn: string
  rentflowField: string
}

interface ValidationError {
  row: number
  field: string
  message: string
  severity: 'error' | 'warning'
}

interface ImportResult {
  imported: number
  skipped: number
  errors: number
  details: string[]
}

// RentFlow fields for equipment
const RENTFLOW_FIELDS = [
  { value: '', label: '-- Nicht zuordnen --' },
  { value: 'name', label: 'Name *' },
  { value: 'sku', label: 'SKU *' },
  { value: 'barcode', label: 'Barcode' },
  { value: 'category', label: 'Kategorie' },
  { value: 'status', label: 'Status' },
  { value: 'condition', label: 'Zustand' },
  { value: 'rental_price_day', label: 'Tagespreis' },
  { value: 'rental_price_week', label: 'Wochenpreis' },
  { value: 'purchase_price', label: 'Einkaufspreis' },
  { value: 'weight', label: 'Gewicht (kg)' },
  { value: 'description', label: 'Beschreibung' },
  { value: 'serial_number', label: 'Seriennummer' },
  { value: 'tags', label: 'Tags' },
]

// Auto-detect mappings from common header names
const AUTO_DETECT_MAP: Record<string, string> = {
  name: 'name',
  bezeichnung: 'name',
  equipment: 'name',
  artikel: 'name',
  sku: 'sku',
  artikelnummer: 'sku',
  barcode: 'barcode',
  ean: 'barcode',
  kategorie: 'category',
  category: 'category',
  status: 'status',
  zustand: 'condition',
  condition: 'condition',
  tagespreis: 'rental_price_day',
  'preis/tag': 'rental_price_day',
  'rental price day': 'rental_price_day',
  mietpreis: 'rental_price_day',
  wochenpreis: 'rental_price_week',
  'preis/woche': 'rental_price_week',
  'rental price week': 'rental_price_week',
  einkaufspreis: 'purchase_price',
  'purchase price': 'purchase_price',
  gewicht: 'weight',
  weight: 'weight',
  beschreibung: 'description',
  description: 'description',
  seriennummer: 'serial_number',
  'serial number': 'serial_number',
  tags: 'tags',
}

const VALID_STATUSES = ['available', 'reserved', 'checked_out', 'in_maintenance', 'retired', 'damaged']
const VALID_CONDITIONS = ['new', 'excellent', 'good', 'fair', 'poor', 'defective']

function EquipmentImportPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const fileInputRef = useRef<HTMLInputElement>(null)

  const [step, setStep] = useState<Step>('upload')
  const [csvData, setCsvData] = useState<string[][]>([])
  const [headers, setHeaders] = useState<string[]>([])
  const [mappings, setMappings] = useState<ColumnMapping[]>([])
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([])
  const [importResult, setImportResult] = useState<ImportResult | null>(null)
  const [importProgress, setImportProgress] = useState(0)
  const [isDragOver, setIsDragOver] = useState(false)
  const [fileName, setFileName] = useState('')

  // Parse CSV file
  const parseFile = useCallback((file: File) => {
    setFileName(file.name)
    Papa.parse(file, {
      encoding: 'UTF-8',
      skipEmptyLines: true,
      complete: (result) => {
        const data = result.data as string[][]
        if (data.length < 2) {
          addNotification('CSV-Datei muss mindestens eine Kopfzeile und eine Datenzeile enthalten.', 'error', { title: 'Fehler' })
          return
        }

        const csvHeaders = data[0].map((h) => h.trim())
        const csvRows = data.slice(1)

        setHeaders(csvHeaders)
        setCsvData(csvRows)

        // Auto-detect column mappings
        const autoMappings: ColumnMapping[] = csvHeaders.map((header) => {
          const normalized = header.toLowerCase().trim()
          const detected = AUTO_DETECT_MAP[normalized] || ''
          return { csvColumn: header, rentflowField: detected }
        })
        setMappings(autoMappings)
        setStep('preview')
      },
      error: () => {
        addNotification('Fehler beim Parsen der CSV-Datei. Bitte Dateiformat prüfen.', 'error', { title: 'Parse-Fehler' })
      },
    })
  }, [addNotification])

  // Handle file drop
  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file && file.name.endsWith('.csv')) {
      parseFile(file)
    } else {
      addNotification('Bitte nur CSV-Dateien hochladen.', 'warning', { title: 'Falsches Format' })
    }
  }, [parseFile, addNotification])

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) parseFile(file)
  }

  // Update a column mapping
  const updateMapping = (index: number, rentflowField: string) => {
    setMappings((prev) => {
      const updated = [...prev]
      updated[index] = { ...updated[index], rentflowField }
      return updated
    })
  }

  // Validate data
  const runValidation = () => {
    const errors: ValidationError[] = []
    const nameIdx = mappings.findIndex((m) => m.rentflowField === 'name')
    const skuIdx = mappings.findIndex((m) => m.rentflowField === 'sku')
    const statusIdx = mappings.findIndex((m) => m.rentflowField === 'status')
    const conditionIdx = mappings.findIndex((m) => m.rentflowField === 'condition')
    const priceFields = ['rental_price_day', 'rental_price_week', 'purchase_price', 'weight']

    if (nameIdx === -1) {
      errors.push({ row: 0, field: 'name', message: 'Spalte "Name" ist nicht zugeordnet (Pflichtfeld).', severity: 'error' })
    }
    if (skuIdx === -1) {
      errors.push({ row: 0, field: 'sku', message: 'Spalte "SKU" ist nicht zugeordnet (Pflichtfeld).', severity: 'error' })
    }

    const seenSkus = new Set<string>()

    csvData.forEach((row, rowIdx) => {
      // Name required
      if (nameIdx >= 0 && !row[nameIdx]?.trim()) {
        errors.push({ row: rowIdx + 2, field: 'name', message: `Zeile ${rowIdx + 2}: Name fehlt.`, severity: 'error' })
      }

      // SKU required + duplicates
      if (skuIdx >= 0) {
        const sku = row[skuIdx]?.trim()
        if (!sku) {
          errors.push({ row: rowIdx + 2, field: 'sku', message: `Zeile ${rowIdx + 2}: SKU fehlt.`, severity: 'error' })
        } else if (seenSkus.has(sku)) {
          errors.push({ row: rowIdx + 2, field: 'sku', message: `Zeile ${rowIdx + 2}: Doppelte SKU "${sku}".`, severity: 'warning' })
        }
        if (sku) seenSkus.add(sku)
      }

      // Validate status
      if (statusIdx >= 0 && row[statusIdx]?.trim()) {
        const val = row[statusIdx].trim().toLowerCase()
        if (!VALID_STATUSES.includes(val)) {
          errors.push({ row: rowIdx + 2, field: 'status', message: `Zeile ${rowIdx + 2}: Ungültiger Status "${row[statusIdx]}".`, severity: 'warning' })
        }
      }

      // Validate condition
      if (conditionIdx >= 0 && row[conditionIdx]?.trim()) {
        const val = row[conditionIdx].trim().toLowerCase()
        if (!VALID_CONDITIONS.includes(val)) {
          errors.push({ row: rowIdx + 2, field: 'condition', message: `Zeile ${rowIdx + 2}: Ungültiger Zustand "${row[conditionIdx]}".`, severity: 'warning' })
        }
      }

      // Validate numeric fields
      priceFields.forEach((field) => {
        const idx = mappings.findIndex((m) => m.rentflowField === field)
        if (idx >= 0 && row[idx]?.trim()) {
          const parsed = parseFloat(row[idx].replace(',', '.'))
          if (isNaN(parsed)) {
            errors.push({ row: rowIdx + 2, field, message: `Zeile ${rowIdx + 2}: "${row[idx]}" ist keine gültige Zahl für ${field}.`, severity: 'error' })
          }
        }
      })
    })

    setValidationErrors(errors)
    setStep('validation')
  }

  // Import mutation
  const { mutate: doImport } = useMutation({
    mutationFn: async () => {
      const hasBlockingErrors = validationErrors.some((e) => e.severity === 'error')
      if (hasBlockingErrors) throw new Error('Es gibt noch Fehler, die behoben werden müssen.')

      const results: ImportResult = { imported: 0, skipped: 0, errors: 0, details: [] }
      const totalRows = csvData.length

      for (let i = 0; i < totalRows; i++) {
        const row = csvData[i]
        try {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const record: Record<string, any> = {}
          mappings.forEach((mapping, idx) => {
            if (mapping.rentflowField && row[idx] !== undefined) {
              const val = row[idx].trim()
              if (!val) return

              if (['rental_price_day', 'rental_price_week', 'purchase_price', 'weight'].includes(mapping.rentflowField)) {
                record[mapping.rentflowField] = parseFloat(val.replace(',', '.'))
              } else if (mapping.rentflowField === 'tags') {
                record[mapping.rentflowField] = val.split(',').map((t: string) => t.trim()).filter(Boolean)
              } else if (mapping.rentflowField === 'status') {
                record[mapping.rentflowField] = val.toLowerCase()
              } else if (mapping.rentflowField === 'condition') {
                record[mapping.rentflowField] = val.toLowerCase()
              } else {
                record[mapping.rentflowField] = val
              }
            }
          })

          if (!record.name || !record.sku) {
            results.skipped++
            results.details.push(`Zeile ${i + 2}: Übersprungen (Name oder SKU fehlt)`)
            setImportProgress(Math.round(((i + 1) / totalRows) * 100))
            continue
          }

          // Set defaults
          if (!record.status) record.status = 'available'
          if (!record.condition) record.condition = 'good'

          await equipmentApi.create(record)
          results.imported++
        } catch (err) {
          results.errors++
          results.details.push(`Zeile ${i + 2}: Fehler beim Import - ${err instanceof Error ? err.message : 'Unbekannter Fehler'}`)
        }
        setImportProgress(Math.round(((i + 1) / totalRows) * 100))
      }
      return results
    },
    onSuccess: (result) => {
      if (result) {
        setImportResult(result)
        setStep('result')
        queryClient.invalidateQueries({ queryKey: ['equipment-list'] })
      }
    },
    onError: (err) => {
      addNotification(err instanceof Error ? err.message : 'Import fehlgeschlagen', 'error', { title: 'Import-Fehler' })
    },
  })

  const startImport = () => {
    setStep('importing')
    setImportProgress(0)
    doImport()
  }

  const errorCount = validationErrors.filter((e) => e.severity === 'error').length
  const warningCount = validationErrors.filter((e) => e.severity === 'warning').length

  return (
    <div style={{ maxWidth: '1000px', margin: '0 auto' }}>
      <div className="page-header" style={{ marginBottom: 'var(--spacing-6)' }}>
        <div>
          <h1 className="page-title">Equipment CSV Import</h1>
          <p className="page-subtitle">Importieren Sie Ausrüstung aus einer CSV-Datei</p>
        </div>
        <button className="btn btn--secondary" onClick={() => navigate('/equipment')}>
          Zurück zur Liste
        </button>
      </div>

      {/* Step Indicator */}
      <div style={stepIndicatorStyle}>
        {(['upload', 'preview', 'validation', 'importing', 'result'] as Step[]).map((s, idx) => {
          const labels = ['Hochladen', 'Vorschau', 'Validierung', 'Import', 'Ergebnis']
          const isActive = s === step
          const stepOrder = ['upload', 'preview', 'validation', 'importing', 'result']
          const isDone = stepOrder.indexOf(step) > idx
          return (
            <div key={s} style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)' }}>
              <div style={{
                width: 32, height: 32, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: '0.85rem', fontWeight: 600,
                backgroundColor: isActive ? 'var(--color-primary)' : isDone ? 'rgba(16, 185, 129, 0.3)' : 'rgba(255,255,255,0.05)',
                color: isActive ? '#fff' : isDone ? '#10b981' : 'var(--color-text-secondary)',
                border: isActive ? '2px solid var(--color-primary)' : '1px solid var(--color-border)',
              }}>
                {isDone ? '\u2713' : idx + 1}
              </div>
              <span style={{ fontSize: '0.85rem', color: isActive ? 'var(--color-primary)' : 'var(--color-text-secondary)' }}>
                {labels[idx]}
              </span>
              {idx < 4 && <div style={{ width: 40, height: 1, backgroundColor: 'var(--color-border)', margin: '0 var(--spacing-2)' }} />}
            </div>
          )
        })}
      </div>

      {/* Content */}
      <div style={glassCardStyle}>
        {/* UPLOAD STEP */}
        {step === 'upload' && (
          <div
            onDrop={handleDrop}
            onDragOver={(e) => { e.preventDefault(); setIsDragOver(true) }}
            onDragLeave={() => setIsDragOver(false)}
            onClick={() => fileInputRef.current?.click()}
            style={{
              ...dropZoneStyle,
              borderColor: isDragOver ? 'var(--color-primary)' : 'var(--color-border)',
              backgroundColor: isDragOver ? 'rgba(0, 212, 255, 0.05)' : 'transparent',
            }}
          >
            <div style={{ fontSize: '3rem', marginBottom: 'var(--spacing-3)' }}>&#128196;</div>
            <h3 style={{ color: 'var(--color-text-primary)', margin: '0 0 var(--spacing-2)' }}>
              CSV-Datei hierher ziehen
            </h3>
            <p style={{ color: 'var(--color-text-secondary)', margin: 0, fontSize: '0.9rem' }}>
              oder klicken zum Durchsuchen
            </p>
            <p style={{ color: 'var(--color-text-tertiary)', margin: 'var(--spacing-3) 0 0', fontSize: '0.8rem' }}>
              Unterstützt: .csv (Semikolon oder Komma getrennt, UTF-8)
            </p>
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv"
              style={{ display: 'none' }}
              onChange={handleFileSelect}
            />
          </div>
        )}

        {/* PREVIEW STEP */}
        {step === 'preview' && (
          <div>
            <div style={{ marginBottom: 'var(--spacing-4)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h3 style={{ margin: 0, color: 'var(--color-text-primary)' }}>Spaltenzuordnung</h3>
                <p style={{ margin: 'var(--spacing-1) 0 0', color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>
                  Datei: {fileName} &mdash; {csvData.length} Zeilen gefunden
                </p>
              </div>
            </div>

            {/* Column Mapping */}
            <div style={{ marginBottom: 'var(--spacing-6)' }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr auto 1fr', gap: 'var(--spacing-2)', alignItems: 'center', marginBottom: 'var(--spacing-3)' }}>
                <span style={labelHeaderStyle}>CSV-Spalte</span>
                <span style={{ color: 'var(--color-text-secondary)', fontSize: '0.8rem' }}>&rarr;</span>
                <span style={labelHeaderStyle}>RentFlow-Feld</span>
              </div>
              {headers.map((header, idx) => (
                <div key={idx} style={{ display: 'grid', gridTemplateColumns: '1fr auto 1fr', gap: 'var(--spacing-2)', alignItems: 'center', marginBottom: 'var(--spacing-2)' }}>
                  <div style={csvColumnLabelStyle}>{header}</div>
                  <span style={{ color: 'var(--color-text-secondary)' }}>&rarr;</span>
                  <select
                    value={mappings[idx]?.rentflowField || ''}
                    onChange={(e) => updateMapping(idx, e.target.value)}
                    style={selectStyle}
                  >
                    {RENTFLOW_FIELDS.map((f) => (
                      <option key={f.value} value={f.value}>{f.label}</option>
                    ))}
                  </select>
                </div>
              ))}
            </div>

            {/* Preview Table */}
            <h4 style={{ color: 'var(--color-text-primary)', marginBottom: 'var(--spacing-3)' }}>
              Vorschau (erste 5 Zeilen)
            </h4>
            <div style={{ overflowX: 'auto' }}>
              <table style={previewTableStyle}>
                <thead>
                  <tr>
                    {headers.map((h, i) => (
                      <th key={i} style={thStyle}>{h}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {csvData.slice(0, 5).map((row, rowIdx) => (
                    <tr key={rowIdx}>
                      {headers.map((_, colIdx) => (
                        <td key={colIdx} style={tdStyle}>{row[colIdx] || ''}</td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div style={{ marginTop: 'var(--spacing-4)', display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
              <button className="btn btn--secondary" onClick={() => { setStep('upload'); setCsvData([]); setHeaders([]); setMappings([]) }}>
                Andere Datei
              </button>
              <button className="btn btn--primary" onClick={runValidation}>
                Weiter zur Validierung
              </button>
            </div>
          </div>
        )}

        {/* VALIDATION STEP */}
        {step === 'validation' && (
          <div>
            <h3 style={{ color: 'var(--color-text-primary)', margin: '0 0 var(--spacing-4)' }}>Validierungsergebnis</h3>

            <div style={{ display: 'flex', gap: 'var(--spacing-4)', marginBottom: 'var(--spacing-4)' }}>
              <div style={statBoxStyle}>
                <div style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--color-text-primary)' }}>{csvData.length}</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>Zeilen gesamt</div>
              </div>
              <div style={{ ...statBoxStyle, borderColor: errorCount > 0 ? 'var(--color-danger)' : 'var(--color-success)' }}>
                <div style={{ fontSize: '1.5rem', fontWeight: 700, color: errorCount > 0 ? 'var(--color-danger)' : 'var(--color-success)' }}>{errorCount}</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>Fehler</div>
              </div>
              <div style={{ ...statBoxStyle, borderColor: warningCount > 0 ? '#f59e0b' : 'var(--color-border)' }}>
                <div style={{ fontSize: '1.5rem', fontWeight: 700, color: warningCount > 0 ? '#f59e0b' : 'var(--color-text-secondary)' }}>{warningCount}</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>Warnungen</div>
              </div>
            </div>

            {validationErrors.length === 0 ? (
              <div style={{ padding: 'var(--spacing-4)', background: 'rgba(16, 185, 129, 0.1)', border: '1px solid rgba(16, 185, 129, 0.3)', borderRadius: 'var(--radius-md)', color: '#10b981' }}>
                Alle Daten sind valide. Der Import kann gestartet werden.
              </div>
            ) : (
              <div style={{ maxHeight: 300, overflowY: 'auto', marginBottom: 'var(--spacing-4)' }}>
                {validationErrors.map((err, idx) => (
                  <div key={idx} style={{
                    padding: 'var(--spacing-2) var(--spacing-3)',
                    marginBottom: 'var(--spacing-1)',
                    borderRadius: 'var(--radius-sm)',
                    fontSize: '0.85rem',
                    background: err.severity === 'error' ? 'rgba(239, 68, 68, 0.1)' : 'rgba(245, 158, 11, 0.1)',
                    color: err.severity === 'error' ? '#ef4444' : '#f59e0b',
                    border: `1px solid ${err.severity === 'error' ? 'rgba(239, 68, 68, 0.2)' : 'rgba(245, 158, 11, 0.2)'}`,
                  }}>
                    {err.severity === 'error' ? 'FEHLER' : 'WARNUNG'}: {err.message}
                  </div>
                ))}
              </div>
            )}

            <div style={{ marginTop: 'var(--spacing-4)', display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
              <button className="btn btn--secondary" onClick={() => setStep('preview')}>
                Zurück
              </button>
              <button
                className="btn btn--primary"
                onClick={startImport}
                disabled={errorCount > 0}
              >
                {errorCount > 0 ? 'Fehler beheben' : `${csvData.length} Einträge importieren`}
              </button>
            </div>
          </div>
        )}

        {/* IMPORTING STEP */}
        {step === 'importing' && (
          <div style={{ textAlign: 'center', padding: 'var(--spacing-8) 0' }}>
            <h3 style={{ color: 'var(--color-text-primary)', marginBottom: 'var(--spacing-4)' }}>Import läuft...</h3>
            <div style={progressBarOuter}>
              <div style={{ ...progressBarInner, width: `${importProgress}%` }} />
            </div>
            <p style={{ color: 'var(--color-text-secondary)', marginTop: 'var(--spacing-3)', fontSize: '0.9rem' }}>
              {importProgress}% abgeschlossen
            </p>
          </div>
        )}

        {/* RESULT STEP */}
        {step === 'result' && importResult && (
          <div>
            <h3 style={{ color: 'var(--color-text-primary)', margin: '0 0 var(--spacing-4)' }}>Import abgeschlossen</h3>

            <div style={{ display: 'flex', gap: 'var(--spacing-4)', marginBottom: 'var(--spacing-4)' }}>
              <div style={{ ...statBoxStyle, borderColor: 'var(--color-success)' }}>
                <div style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--color-success)' }}>{importResult.imported}</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>Importiert</div>
              </div>
              <div style={{ ...statBoxStyle, borderColor: '#f59e0b' }}>
                <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#f59e0b' }}>{importResult.skipped}</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>Übersprungen</div>
              </div>
              <div style={{ ...statBoxStyle, borderColor: 'var(--color-danger)' }}>
                <div style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--color-danger)' }}>{importResult.errors}</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)' }}>Fehler</div>
              </div>
            </div>

            {importResult.details.length > 0 && (
              <div style={{ maxHeight: 200, overflowY: 'auto', marginBottom: 'var(--spacing-4)', padding: 'var(--spacing-3)', background: 'rgba(255,255,255,0.03)', borderRadius: 'var(--radius-md)', border: '1px solid var(--color-border)' }}>
                {importResult.details.map((d, i) => (
                  <div key={i} style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)', padding: 'var(--spacing-1) 0' }}>{d}</div>
                ))}
              </div>
            )}

            <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
              <button className="btn btn--secondary" onClick={() => { setStep('upload'); setCsvData([]); setHeaders([]); setMappings([]); setImportResult(null) }}>
                Weiteren Import starten
              </button>
              <button className="btn btn--primary" onClick={() => navigate('/equipment')}>
                Zur Ausrüstungsliste
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

// ---------- Styles ----------
const stepIndicatorStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  gap: 'var(--spacing-2)',
  marginBottom: 'var(--spacing-6)',
  flexWrap: 'wrap',
}

const glassCardStyle: React.CSSProperties = {
  background: 'rgba(10, 15, 30, 0.6)',
  backdropFilter: 'blur(20px)',
  WebkitBackdropFilter: 'blur(20px)',
  border: '1px solid rgba(255, 255, 255, 0.08)',
  borderRadius: 'var(--radius-card)',
  padding: 'var(--spacing-6)',
  boxShadow: '0 8px 32px rgba(0, 0, 0, 0.3)',
}

const dropZoneStyle: React.CSSProperties = {
  border: '2px dashed var(--color-border)',
  borderRadius: 'var(--radius-card)',
  padding: 'var(--spacing-10, 60px) var(--spacing-6)',
  textAlign: 'center',
  cursor: 'pointer',
  transition: 'all 0.2s ease',
}

const selectStyle: React.CSSProperties = {
  width: '100%',
  padding: 'var(--spacing-2) var(--spacing-3)',
  backgroundColor: 'var(--glass-bg-input-strong)',
  border: '1px solid var(--color-border-strong)',
  borderRadius: 'var(--radius-input)',
  color: 'var(--color-text-primary)',
  fontSize: '0.85rem',
}

const labelHeaderStyle: React.CSSProperties = {
  fontSize: '0.75rem',
  fontWeight: 600,
  color: 'var(--color-text-secondary)',
  textTransform: 'uppercase',
  letterSpacing: '0.05em',
}

const csvColumnLabelStyle: React.CSSProperties = {
  padding: 'var(--spacing-2) var(--spacing-3)',
  backgroundColor: 'rgba(255,255,255,0.03)',
  borderRadius: 'var(--radius-sm)',
  fontSize: '0.85rem',
  color: 'var(--color-text-primary)',
  border: '1px solid var(--color-border)',
}

const previewTableStyle: React.CSSProperties = {
  width: '100%',
  borderCollapse: 'collapse',
  fontSize: '0.8rem',
}

const thStyle: React.CSSProperties = {
  padding: 'var(--spacing-2) var(--spacing-3)',
  textAlign: 'left',
  borderBottom: '1px solid var(--color-border)',
  color: 'var(--color-text-secondary)',
  fontSize: '0.75rem',
  fontWeight: 600,
  textTransform: 'uppercase',
  whiteSpace: 'nowrap',
}

const tdStyle: React.CSSProperties = {
  padding: 'var(--spacing-2) var(--spacing-3)',
  borderBottom: '1px solid rgba(255,255,255,0.03)',
  color: 'var(--color-text-primary)',
  whiteSpace: 'nowrap',
  maxWidth: 200,
  overflow: 'hidden',
  textOverflow: 'ellipsis',
}

const statBoxStyle: React.CSSProperties = {
  flex: 1,
  padding: 'var(--spacing-4)',
  textAlign: 'center',
  borderRadius: 'var(--radius-md)',
  border: '1px solid var(--color-border)',
  background: 'rgba(255,255,255,0.02)',
}

const progressBarOuter: React.CSSProperties = {
  width: '100%',
  height: 12,
  borderRadius: 6,
  backgroundColor: 'rgba(255,255,255,0.05)',
  overflow: 'hidden',
}

const progressBarInner: React.CSSProperties = {
  height: '100%',
  borderRadius: 6,
  background: 'linear-gradient(90deg, var(--color-primary), #10b981)',
  transition: 'width 0.3s ease',
}

export default EquipmentImportPage
