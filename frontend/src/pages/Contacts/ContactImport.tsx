import { useState, useCallback, useRef } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import Papa from 'papaparse'
import { contactApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { Modal } from '../../components/Modal/Modal'

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

const CONTACT_FIELDS = [
  { value: '', label: '-- Nicht zuordnen --' },
  { value: 'type', label: 'Typ (company/person)' },
  { value: 'company_name', label: 'Firmenname' },
  { value: 'first_name', label: 'Vorname' },
  { value: 'last_name', label: 'Nachname' },
  { value: 'email', label: 'E-Mail' },
  { value: 'phone', label: 'Telefon' },
  { value: 'mobile', label: 'Mobil' },
  { value: 'website', label: 'Website' },
  { value: 'street', label: 'Straße' },
  { value: 'house_number', label: 'Hausnummer' },
  { value: 'zip', label: 'PLZ' },
  { value: 'city', label: 'Stadt' },
  { value: 'country', label: 'Land' },
  { value: 'vat_id', label: 'USt-ID' },
  { value: 'notes', label: 'Notizen' },
  { value: 'tags', label: 'Tags' },
]

const AUTO_DETECT: Record<string, string> = {
  typ: 'type', type: 'type',
  firmenname: 'company_name', firma: 'company_name', company: 'company_name', 'company name': 'company_name',
  vorname: 'first_name', 'first name': 'first_name',
  nachname: 'last_name', 'last name': 'last_name',
  email: 'email', 'e-mail': 'email',
  telefon: 'phone', phone: 'phone',
  mobil: 'mobile', mobile: 'mobile', handy: 'mobile',
  website: 'website', webseite: 'website',
  'straße': 'street', strasse: 'street', street: 'street',
  hausnummer: 'house_number', 'house number': 'house_number',
  plz: 'zip', postleitzahl: 'zip', zip: 'zip',
  stadt: 'city', ort: 'city', city: 'city',
  land: 'country', country: 'country',
  'ust-id': 'vat_id', 'vat id': 'vat_id', ustid: 'vat_id',
  notizen: 'notes', notes: 'notes',
  tags: 'tags',
}

interface ContactImportProps {
  isOpen: boolean
  onClose: () => void
}

function ContactImport({ isOpen, onClose }: ContactImportProps) {
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

  const reset = () => {
    setStep('upload')
    setCsvData([])
    setHeaders([])
    setMappings([])
    setValidationErrors([])
    setImportResult(null)
    setImportProgress(0)
    setFileName('')
  }

  const handleClose = () => {
    reset()
    onClose()
  }

  const parseFile = useCallback((file: File) => {
    setFileName(file.name)
    Papa.parse(file, {
      encoding: 'UTF-8',
      skipEmptyLines: true,
      complete: (result) => {
        const data = result.data as string[][]
        if (data.length < 2) {
          addNotification('CSV muss mindestens Kopfzeile + eine Datenzeile enthalten.', 'error', { title: 'Fehler' })
          return
        }
        const csvHeaders = data[0].map((h) => h.trim())
        setCsvData(data.slice(1))
        setHeaders(csvHeaders)
        setMappings(csvHeaders.map((h) => ({
          csvColumn: h,
          rentflowField: AUTO_DETECT[h.toLowerCase().trim()] || '',
        })))
        setStep('preview')
      },
      error: () => {
        addNotification('Fehler beim Parsen der CSV-Datei.', 'error', { title: 'Parse-Fehler' })
      },
    })
  }, [addNotification])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file?.name.endsWith('.csv')) parseFile(file)
    else addNotification('Bitte nur CSV-Dateien.', 'warning', { title: 'Format' })
  }, [parseFile, addNotification])

  const updateMapping = (idx: number, val: string) => {
    setMappings((prev) => {
      const u = [...prev]
      u[idx] = { ...u[idx], rentflowField: val }
      return u
    })
  }

  const validate = () => {
    const errors: ValidationError[] = []
    const hasName = mappings.some((m) => m.rentflowField === 'company_name' || m.rentflowField === 'last_name')
    if (!hasName) {
      errors.push({ row: 0, field: 'name', message: 'Mindestens "Firmenname" oder "Nachname" muss zugeordnet sein.', severity: 'error' })
    }

    const emailIdx = mappings.findIndex((m) => m.rentflowField === 'email')
    csvData.forEach((row, i) => {
      if (emailIdx >= 0 && row[emailIdx]?.trim()) {
        const email = row[emailIdx].trim()
        if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
          errors.push({ row: i + 2, field: 'email', message: `Zeile ${i + 2}: Ungültige E-Mail "${email}".`, severity: 'warning' })
        }
      }
    })
    setValidationErrors(errors)
    setStep('validation')
  }

  const { mutate: doImport } = useMutation({
    mutationFn: async () => {
      if (validationErrors.some((e) => e.severity === 'error')) throw new Error('Fehler vorhanden.')
      const results: ImportResult = { imported: 0, skipped: 0, errors: 0, details: [] }
      for (let i = 0; i < csvData.length; i++) {
        const row = csvData[i]
        try {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const record: Record<string, any> = {}
          mappings.forEach((m, idx) => {
            if (m.rentflowField && row[idx]?.trim()) {
              const val = row[idx].trim()
              if (m.rentflowField === 'tags') {
                record.tags = val.split(',').map((t: string) => t.trim()).filter(Boolean)
              } else {
                record[m.rentflowField] = val
              }
            }
          })
          if (!record.type) record.type = record.company_name ? 'company' : 'person'
          if (!record.company_name && !record.last_name) {
            results.skipped++
            results.details.push(`Zeile ${i + 2}: Übersprungen (kein Name)`)
            setImportProgress(Math.round(((i + 1) / csvData.length) * 100))
            continue
          }
          await contactApi.create(record)
          results.imported++
        } catch (err) {
          results.errors++
          results.details.push(`Zeile ${i + 2}: ${err instanceof Error ? err.message : 'Fehler'}`)
        }
        setImportProgress(Math.round(((i + 1) / csvData.length) * 100))
      }
      return results
    },
    onSuccess: (result) => {
      if (result) {
        setImportResult(result)
        setStep('result')
        queryClient.invalidateQueries({ queryKey: ['contacts'] })
      }
    },
    onError: (err) => {
      addNotification(err instanceof Error ? err.message : 'Import fehlgeschlagen', 'error', { title: 'Fehler' })
    },
  })

  const startImport = () => { setStep('importing'); setImportProgress(0); doImport() }
  const errCount = validationErrors.filter((e) => e.severity === 'error').length
  const warnCount = validationErrors.filter((e) => e.severity === 'warning').length

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Kontakte CSV Import" size="lg">
      <div style={{ minHeight: 300 }}>
        {/* Step Indicator */}
        <div style={{ display: 'flex', justifyContent: 'center', gap: 'var(--spacing-2)', marginBottom: 'var(--spacing-4)', flexWrap: 'wrap' }}>
          {(['upload', 'preview', 'validation', 'importing', 'result'] as Step[]).map((s, idx) => {
            const labels = ['Hochladen', 'Vorschau', 'Validierung', 'Import', 'Ergebnis']
            const stepOrder = ['upload', 'preview', 'validation', 'importing', 'result']
            const isActive = s === step
            const isDone = stepOrder.indexOf(step) > idx
            return (
              <div key={s} style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-1)' }}>
                <div style={{
                  width: 24, height: 24, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: '0.7rem', fontWeight: 600,
                  backgroundColor: isActive ? 'var(--color-primary)' : isDone ? 'rgba(16,185,129,0.3)' : 'rgba(255,255,255,0.05)',
                  color: isActive ? '#fff' : isDone ? '#10b981' : 'var(--color-text-secondary)',
                  border: isActive ? '2px solid var(--color-primary)' : '1px solid var(--color-border)',
                }}>
                  {isDone ? '\u2713' : idx + 1}
                </div>
                <span style={{ fontSize: '0.75rem', color: isActive ? 'var(--color-primary)' : 'var(--color-text-secondary)' }}>{labels[idx]}</span>
                {idx < 4 && <div style={{ width: 20, height: 1, backgroundColor: 'var(--color-border)', margin: '0 var(--spacing-1)' }} />}
              </div>
            )
          })}
        </div>

        {/* UPLOAD */}
        {step === 'upload' && (
          <div
            onDrop={handleDrop}
            onDragOver={(e) => { e.preventDefault(); setIsDragOver(true) }}
            onDragLeave={() => setIsDragOver(false)}
            onClick={() => fileInputRef.current?.click()}
            style={{
              border: `2px dashed ${isDragOver ? 'var(--color-primary)' : 'var(--color-border)'}`,
              borderRadius: 'var(--radius-card)',
              padding: '40px var(--spacing-6)',
              textAlign: 'center',
              cursor: 'pointer',
              backgroundColor: isDragOver ? 'rgba(0,212,255,0.05)' : 'transparent',
              transition: 'all 0.2s',
            }}
          >
            <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-2)' }}>&#128196;</div>
            <h3 style={{ color: 'var(--color-text-primary)', margin: '0 0 var(--spacing-1)' }}>CSV-Datei hierher ziehen</h3>
            <p style={{ color: 'var(--color-text-secondary)', margin: 0, fontSize: '0.85rem' }}>oder klicken zum Durchsuchen</p>
            <input ref={fileInputRef} type="file" accept=".csv" style={{ display: 'none' }} onChange={(e) => { const f = e.target.files?.[0]; if (f) parseFile(f) }} />
          </div>
        )}

        {/* PREVIEW */}
        {step === 'preview' && (
          <div>
            <p style={{ color: 'var(--color-text-secondary)', fontSize: '0.85rem', marginBottom: 'var(--spacing-3)' }}>
              Datei: {fileName} &mdash; {csvData.length} Zeilen
            </p>
            <div style={{ marginBottom: 'var(--spacing-4)' }}>
              {headers.map((h, idx) => (
                <div key={idx} style={{ display: 'grid', gridTemplateColumns: '1fr auto 1fr', gap: 'var(--spacing-2)', alignItems: 'center', marginBottom: 'var(--spacing-2)' }}>
                  <div style={{ padding: 'var(--spacing-2)', background: 'rgba(255,255,255,0.03)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', color: 'var(--color-text-primary)', border: '1px solid var(--color-border)' }}>{h}</div>
                  <span style={{ color: 'var(--color-text-secondary)' }}>&rarr;</span>
                  <select value={mappings[idx]?.rentflowField || ''} onChange={(e) => updateMapping(idx, e.target.value)} style={{ width: '100%', padding: 'var(--spacing-2)', backgroundColor: 'var(--glass-bg-input-strong)', border: '1px solid var(--color-border-strong)', borderRadius: 'var(--radius-input)', color: 'var(--color-text-primary)', fontSize: '0.85rem' }}>
                    {CONTACT_FIELDS.map((f) => <option key={f.value} value={f.value}>{f.label}</option>)}
                  </select>
                </div>
              ))}
            </div>
            <div style={{ overflowX: 'auto', marginBottom: 'var(--spacing-3)' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem' }}>
                <thead>
                  <tr>{headers.map((h, i) => <th key={i} style={{ padding: 'var(--spacing-2)', textAlign: 'left', borderBottom: '1px solid var(--color-border)', color: 'var(--color-text-secondary)', fontSize: '0.75rem', whiteSpace: 'nowrap' }}>{h}</th>)}</tr>
                </thead>
                <tbody>
                  {csvData.slice(0, 5).map((row, ri) => (
                    <tr key={ri}>{headers.map((_, ci) => <td key={ci} style={{ padding: 'var(--spacing-2)', borderBottom: '1px solid rgba(255,255,255,0.03)', color: 'var(--color-text-primary)', whiteSpace: 'nowrap', maxWidth: 150, overflow: 'hidden', textOverflow: 'ellipsis' }}>{row[ci] || ''}</td>)}</tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
              <button className="btn btn--secondary" onClick={reset}>Andere Datei</button>
              <button className="btn btn--primary" onClick={validate}>Validieren</button>
            </div>
          </div>
        )}

        {/* VALIDATION */}
        {step === 'validation' && (
          <div>
            <div style={{ display: 'flex', gap: 'var(--spacing-3)', marginBottom: 'var(--spacing-3)' }}>
              <div style={{ flex: 1, padding: 'var(--spacing-3)', textAlign: 'center', borderRadius: 'var(--radius-md)', border: '1px solid var(--color-border)', background: 'rgba(255,255,255,0.02)' }}>
                <div style={{ fontSize: '1.3rem', fontWeight: 700, color: 'var(--color-text-primary)' }}>{csvData.length}</div>
                <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)' }}>Zeilen</div>
              </div>
              <div style={{ flex: 1, padding: 'var(--spacing-3)', textAlign: 'center', borderRadius: 'var(--radius-md)', border: `1px solid ${errCount > 0 ? 'var(--color-danger)' : 'var(--color-success)'}`, background: 'rgba(255,255,255,0.02)' }}>
                <div style={{ fontSize: '1.3rem', fontWeight: 700, color: errCount > 0 ? 'var(--color-danger)' : 'var(--color-success)' }}>{errCount}</div>
                <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)' }}>Fehler</div>
              </div>
              <div style={{ flex: 1, padding: 'var(--spacing-3)', textAlign: 'center', borderRadius: 'var(--radius-md)', border: `1px solid ${warnCount > 0 ? '#f59e0b' : 'var(--color-border)'}`, background: 'rgba(255,255,255,0.02)' }}>
                <div style={{ fontSize: '1.3rem', fontWeight: 700, color: warnCount > 0 ? '#f59e0b' : 'var(--color-text-secondary)' }}>{warnCount}</div>
                <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)' }}>Warnungen</div>
              </div>
            </div>
            {validationErrors.length === 0 ? (
              <div style={{ padding: 'var(--spacing-3)', background: 'rgba(16,185,129,0.1)', border: '1px solid rgba(16,185,129,0.3)', borderRadius: 'var(--radius-md)', color: '#10b981', fontSize: '0.9rem' }}>
                Alle Daten valide. Import kann starten.
              </div>
            ) : (
              <div style={{ maxHeight: 200, overflowY: 'auto', marginBottom: 'var(--spacing-3)' }}>
                {validationErrors.map((e, i) => (
                  <div key={i} style={{ padding: 'var(--spacing-2)', marginBottom: 'var(--spacing-1)', borderRadius: 'var(--radius-sm)', fontSize: '0.8rem', background: e.severity === 'error' ? 'rgba(239,68,68,0.1)' : 'rgba(245,158,11,0.1)', color: e.severity === 'error' ? '#ef4444' : '#f59e0b', border: `1px solid ${e.severity === 'error' ? 'rgba(239,68,68,0.2)' : 'rgba(245,158,11,0.2)'}` }}>
                    {e.severity === 'error' ? 'FEHLER' : 'WARNUNG'}: {e.message}
                  </div>
                ))}
              </div>
            )}
            <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
              <button className="btn btn--secondary" onClick={() => setStep('preview')}>Zurück</button>
              <button className="btn btn--primary" onClick={startImport} disabled={errCount > 0}>
                {errCount > 0 ? 'Fehler beheben' : `${csvData.length} Kontakte importieren`}
              </button>
            </div>
          </div>
        )}

        {/* IMPORTING */}
        {step === 'importing' && (
          <div style={{ textAlign: 'center', padding: 'var(--spacing-6) 0' }}>
            <h3 style={{ color: 'var(--color-text-primary)', marginBottom: 'var(--spacing-3)' }}>Import läuft...</h3>
            <div style={{ width: '100%', height: 10, borderRadius: 5, backgroundColor: 'rgba(255,255,255,0.05)', overflow: 'hidden' }}>
              <div style={{ height: '100%', borderRadius: 5, background: 'linear-gradient(90deg, var(--color-primary), #10b981)', transition: 'width 0.3s', width: `${importProgress}%` }} />
            </div>
            <p style={{ color: 'var(--color-text-secondary)', marginTop: 'var(--spacing-2)', fontSize: '0.85rem' }}>{importProgress}%</p>
          </div>
        )}

        {/* RESULT */}
        {step === 'result' && importResult && (
          <div>
            <div style={{ display: 'flex', gap: 'var(--spacing-3)', marginBottom: 'var(--spacing-3)' }}>
              {[
                { label: 'Importiert', value: importResult.imported, color: 'var(--color-success)' },
                { label: 'Übersprungen', value: importResult.skipped, color: '#f59e0b' },
                { label: 'Fehler', value: importResult.errors, color: 'var(--color-danger)' },
              ].map((s) => (
                <div key={s.label} style={{ flex: 1, padding: 'var(--spacing-3)', textAlign: 'center', borderRadius: 'var(--radius-md)', border: `1px solid ${s.color}`, background: 'rgba(255,255,255,0.02)' }}>
                  <div style={{ fontSize: '1.3rem', fontWeight: 700, color: s.color }}>{s.value}</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)' }}>{s.label}</div>
                </div>
              ))}
            </div>
            {importResult.details.length > 0 && (
              <div style={{ maxHeight: 150, overflowY: 'auto', padding: 'var(--spacing-2)', background: 'rgba(255,255,255,0.03)', borderRadius: 'var(--radius-md)', border: '1px solid var(--color-border)', marginBottom: 'var(--spacing-3)' }}>
                {importResult.details.map((d, i) => <div key={i} style={{ fontSize: '0.75rem', color: 'var(--color-text-secondary)', padding: '2px 0' }}>{d}</div>)}
              </div>
            )}
            <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
              <button className="btn btn--secondary" onClick={reset}>Weiteren Import</button>
              <button className="btn btn--primary" onClick={handleClose}>Schließen</button>
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}

export default ContactImport
