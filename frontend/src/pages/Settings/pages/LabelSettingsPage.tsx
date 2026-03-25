import { useState, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import '../Settings.scss'

interface LabelConfig {
  label_size: '50x25' | '70x35' | '100x50' | 'custom'
  custom_width_mm?: number
  custom_height_mm?: number
  show_qr_code: boolean
  show_name: boolean
  show_sku: boolean
  show_barcode: boolean
  show_category: boolean
  show_location: boolean
  show_company_name: boolean
  show_logo: boolean
  font_size: 'small' | 'medium' | 'large'
  qr_size: 'small' | 'medium' | 'large'
  printer_type: 'browser' | 'zebra'
  printer_profile?: string
}

const DEFAULT_CONFIG: LabelConfig = {
  label_size: '50x25',
  show_qr_code: true,
  show_name: true,
  show_sku: true,
  show_barcode: true,
  show_category: false,
  show_location: false,
  show_company_name: true,
  show_logo: false,
  font_size: 'medium',
  qr_size: 'medium',
  printer_type: 'browser',
}

const LABEL_DIMENSIONS: Record<string, { w: number; h: number }> = {
  '50x25': { w: 200, h: 100 },
  '70x35': { w: 280, h: 140 },
  '100x50': { w: 400, h: 200 },
  'custom': { w: 300, h: 150 },
}

const FONT_SIZES: Record<string, { title: number; body: number; small: number }> = {
  small: { title: 8, body: 6.5, small: 5.5 },
  medium: { title: 10, body: 8, small: 6.5 },
  large: { title: 13, body: 10, small: 8 },
}

const QR_SIZES: Record<string, number> = {
  small: 0.35,
  medium: 0.5,
  large: 0.65,
}

function LabelPreview({ config }: { config: LabelConfig }) {
  const dim = LABEL_DIMENSIONS[config.label_size]
  const fonts = FONT_SIZES[config.font_size]
  const qrRatio = QR_SIZES[config.qr_size]
  const qrPx = Math.round(dim.h * qrRatio)
  const padding = 8

  const hasLeftColumn = config.show_qr_code || config.show_logo

  const textItems: { text: string; size: number; bold?: boolean; color?: string }[] = []

  if (config.show_company_name) {
    textItems.push({ text: 'JE Sound & Light', size: fonts.small, color: '#666' })
  }
  if (config.show_name) {
    textItems.push({ text: 'Shure SM58', size: fonts.title, bold: true })
  }
  if (config.show_sku) {
    textItems.push({ text: 'SKU-2024-0042', size: fonts.body, color: '#555' })
  }
  if (config.show_barcode) {
    textItems.push({ text: '|||||||||||||||||||', size: fonts.body, color: '#000' })
  }
  if (config.show_category) {
    textItems.push({ text: 'Mikrofone', size: fonts.small, color: '#888' })
  }
  if (config.show_location) {
    textItems.push({ text: 'Lager A / Regal 3', size: fonts.small, color: '#888' })
  }

  return (
    <div
      style={{
        width: dim.w,
        height: dim.h,
        background: '#fff',
        border: '1.5px solid #bbb',
        borderRadius: 4,
        padding,
        display: 'flex',
        gap: padding,
        overflow: 'hidden',
        boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
        fontFamily: 'Arial, Helvetica, sans-serif',
        position: 'relative',
      }}
    >
      {/* Left column: QR / Logo */}
      {hasLeftColumn && (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 3, flexShrink: 0, width: qrPx }}>
          {config.show_logo && (
            <div
              style={{
                width: Math.min(qrPx, 30),
                height: Math.min(qrPx * 0.35, 14),
                background: '#e2e2e2',
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: 5,
                color: '#999',
              }}
            >
              LOGO
            </div>
          )}
          {config.show_qr_code && (
            <div
              style={{
                width: qrPx,
                height: qrPx,
                border: '1px solid #ccc',
                borderRadius: 2,
                display: 'grid',
                gridTemplateColumns: 'repeat(7, 1fr)',
                gridTemplateRows: 'repeat(7, 1fr)',
                gap: 0,
                overflow: 'hidden',
              }}
            >
              {/* Simple QR-code-like pattern */}
              {Array.from({ length: 49 }).map((_, i) => {
                const row = Math.floor(i / 7)
                const col = i % 7
                const isCorner =
                  (row < 3 && col < 3) ||
                  (row < 3 && col > 3) ||
                  (row > 3 && col < 3)
                const isFill = isCorner || (i % 3 === 0)
                return (
                  <div
                    key={i}
                    style={{
                      background: isFill ? '#222' : '#fff',
                    }}
                  />
                )
              })}
            </div>
          )}
        </div>
      )}

      {/* Right column: text content */}
      <div
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          gap: 2,
          overflow: 'hidden',
          minWidth: 0,
        }}
      >
        {textItems.map((item, idx) => (
          <div
            key={idx}
            style={{
              fontSize: item.size,
              fontWeight: item.bold ? 700 : 400,
              color: item.color || '#000',
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              lineHeight: 1.3,
            }}
          >
            {item.text}
          </div>
        ))}
        {textItems.length === 0 && (
          <div style={{ fontSize: fonts.small, color: '#bbb', fontStyle: 'italic' }}>Kein Inhalt ausgewaehlt</div>
        )}
      </div>
    </div>
  )
}

function LabelSettingsPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [saveError, setSaveError] = useState(false)
  const printRef = useRef<HTMLDivElement>(null)

  const [form, setForm] = useState<LabelConfig>(DEFAULT_CONFIG)

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'labels.settings'],
    queryFn: () => configApi.get('labels.settings'),
  })

  useEffect(() => {
    if (config && typeof config === 'object' && Object.keys(config).length > 0) {
      setForm((prev) => ({
        ...prev,
        ...config,
      }))
    }
  }, [config])

  const saveMutation = useMutation({
    mutationFn: (data: LabelConfig) => configApi.set('labels.settings', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'labels.settings'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
    onError: () => {
      setSaveError(true)
      setTimeout(() => setSaveError(false), 5000)
    },
  })

  const handleSave = () => {
    saveMutation.mutate(form)
  }

  const toggleField = (field: keyof LabelConfig) => {
    setForm((prev) => ({
      ...prev,
      [field]: !prev[field],
    }))
  }

  const handlePrintPreview = () => {
    const printWindow = window.open('', '_blank', 'width=600,height=400')
    if (!printWindow) return

    const dim = LABEL_DIMENSIONS[form.label_size]
    const labels = Array.from({ length: 4 })

    printWindow.document.write(`
      <!DOCTYPE html>
      <html>
      <head>
        <title>Label-Vorschau</title>
        <style>
          body { font-family: Arial, Helvetica, sans-serif; padding: 20px; }
          .label-grid { display: flex; flex-wrap: wrap; gap: 12px; }
          .label {
            width: ${dim.w}px;
            height: ${dim.h}px;
            border: 1.5px solid #bbb;
            border-radius: 4px;
            padding: 8px;
            display: flex;
            gap: 8px;
            box-sizing: border-box;
            page-break-inside: avoid;
          }
          .qr-box {
            width: ${Math.round(dim.h * QR_SIZES[form.qr_size])}px;
            height: ${Math.round(dim.h * QR_SIZES[form.qr_size])}px;
            border: 1px solid #ccc;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 8px;
            color: #999;
            flex-shrink: 0;
          }
          .text-col { flex: 1; display: flex; flex-direction: column; justify-content: center; gap: 2px; overflow: hidden; }
          .text-col div { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; line-height: 1.3; }
          @media print { body { padding: 0; } }
        </style>
      </head>
      <body>
        <div class="label-grid">
          ${labels.map((_, i) => {
            const names = ['Shure SM58', 'Sennheiser e945', 'XLR-Kabel 10m', 'LED PAR 64']
            const skus = ['SKU-2024-0042', 'SKU-2024-0108', 'SKU-2024-0203', 'SKU-2024-0055']
            const fonts = FONT_SIZES[form.font_size]
            const hasLeft = form.show_qr_code || form.show_logo
            return `
              <div class="label">
                ${hasLeft ? `<div style="display:flex;flex-direction:column;align-items:center;justify-content:center;gap:3px;flex-shrink:0;">
                  ${form.show_logo ? '<div style="width:30px;height:12px;background:#e2e2e2;border-radius:2px;display:flex;align-items:center;justify-content:center;font-size:5px;color:#999;">LOGO</div>' : ''}
                  ${form.show_qr_code ? '<div class="qr-box">QR</div>' : ''}
                </div>` : ''}
                <div class="text-col">
                  ${form.show_company_name ? `<div style="font-size:${fonts.small}px;color:#666;">JE Sound & Light</div>` : ''}
                  ${form.show_name ? `<div style="font-size:${fonts.title}px;font-weight:700;">${names[i]}</div>` : ''}
                  ${form.show_sku ? `<div style="font-size:${fonts.body}px;color:#555;">${skus[i]}</div>` : ''}
                  ${form.show_barcode ? `<div style="font-size:${fonts.body}px;">|||||||||||||||||||</div>` : ''}
                  ${form.show_category ? `<div style="font-size:${fonts.small}px;color:#888;">Mikrofone</div>` : ''}
                  ${form.show_location ? `<div style="font-size:${fonts.small}px;color:#888;">Lager A / Regal 3</div>` : ''}
                </div>
              </div>
            `
          }).join('')}
        </div>
        <script>window.onload = function() { window.print(); }</script>
      </body>
      </html>
    `)
    printWindow.document.close()
  }

  if (isLoading) {
    return <div className="sp-loading">Label-Einstellungen werden geladen...</div>
  }

  const contentCheckboxes: { field: keyof LabelConfig; label: string; desc: string }[] = [
    { field: 'show_qr_code', label: 'QR-Code', desc: 'QR-Code mit Equipment-ID' },
    { field: 'show_name', label: 'Equipment-Name', desc: 'Name des Geraets' },
    { field: 'show_sku', label: 'SKU', desc: 'Artikelnummer / SKU-Code' },
    { field: 'show_barcode', label: 'Barcode-Text', desc: 'Barcode als Strichcode-Darstellung' },
    { field: 'show_category', label: 'Kategorie', desc: 'Kategorie des Geraets' },
    { field: 'show_location', label: 'Standort', desc: 'Aktueller Standort / Lagerplatz' },
    { field: 'show_company_name', label: 'Firmenname', desc: 'JE Sound & Light' },
    { field: 'show_logo', label: 'Logo', desc: 'Firmenlogo auf dem Label' },
  ]

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Label / QR-Code Editor</h1>
        <p>Gestalten Sie das Layout Ihrer Equipment-Labels. Die Vorschau aktualisiert sich automatisch.</p>
      </div>

      {/* Live Preview */}
      <div className="sp-card">
        <h3 className="sp-card__title">Live-Vorschau</h3>
        <div
          ref={printRef}
          style={{
            display: 'flex',
            justifyContent: 'center',
            padding: 'var(--spacing-6)',
            background: 'var(--color-bg-secondary, #f5f5f5)',
            borderRadius: 'var(--radius-md)',
            minHeight: 140,
            alignItems: 'center',
          }}
        >
          <LabelPreview config={form} />
        </div>
        <div style={{ marginTop: 'var(--spacing-3)', textAlign: 'center' }}>
          <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)' }}>
            Label-Groesse: {form.label_size}mm | Schrift: {form.font_size === 'small' ? 'Klein' : form.font_size === 'medium' ? 'Mittel' : 'Gross'} | QR: {form.qr_size === 'small' ? 'Klein' : form.qr_size === 'medium' ? 'Mittel' : 'Gross'}
          </span>
        </div>
      </div>

      {/* Label Size */}
      <div className="sp-card">
        <h3 className="sp-card__title">Label-Groesse</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          {([
            { value: '50x25', label: '50 x 25 mm', desc: 'Standard klein' },
            { value: '70x35', label: '70 x 35 mm', desc: 'Standard mittel' },
            { value: '100x50', label: '100 x 50 mm', desc: 'Gross' },
          ] as const).map((opt) => (
            <label
              key={opt.value}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 'var(--spacing-2)',
                padding: 'var(--spacing-3) var(--spacing-4)',
                border: form.label_size === opt.value ? '2px solid var(--color-primary, #6366f1)' : '2px solid var(--color-border, #ddd)',
                borderRadius: 'var(--radius-md)',
                cursor: 'pointer',
                background: form.label_size === opt.value ? 'var(--color-primary-light, #eef2ff)' : 'transparent',
                transition: 'all 0.15s ease',
              }}
            >
              <input
                type="radio"
                name="label_size"
                value={opt.value}
                checked={form.label_size === opt.value}
                onChange={() => setForm((prev) => ({ ...prev, label_size: opt.value }))}
                style={{ accentColor: 'var(--color-primary, #6366f1)' }}
              />
              <div>
                <div style={{ fontWeight: 600, fontSize: 'var(--font-size-sm)' }}>{opt.label}</div>
                <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)' }}>{opt.desc}</div>
              </div>
            </label>
          ))}
        </div>
      </div>

      {/* Content Checkboxes */}
      <div className="sp-card">
        <h3 className="sp-card__title">Inhalt auswaehlen</h3>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
          {contentCheckboxes.map((item) => (
            <label
              key={item.field}
              className="sp-toggle"
              onClick={() => toggleField(item.field)}
              style={{ cursor: 'pointer' }}
            >
              <div className="sp-toggle__label">
                <strong>{item.label}</strong>
                <span>{item.desc}</span>
              </div>
              <input type="checkbox" checked={!!form[item.field]} readOnly />
            </label>
          ))}
        </div>
      </div>

      {/* Font Size */}
      <div className="sp-card">
        <h3 className="sp-card__title">Schriftgroesse</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          {([
            { value: 'small', label: 'Klein' },
            { value: 'medium', label: 'Mittel' },
            { value: 'large', label: 'Gross' },
          ] as const).map((opt) => (
            <label
              key={opt.value}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 'var(--spacing-2)',
                padding: 'var(--spacing-2) var(--spacing-4)',
                border: form.font_size === opt.value ? '2px solid var(--color-primary, #6366f1)' : '2px solid var(--color-border, #ddd)',
                borderRadius: 'var(--radius-md)',
                cursor: 'pointer',
                background: form.font_size === opt.value ? 'var(--color-primary-light, #eef2ff)' : 'transparent',
                transition: 'all 0.15s ease',
              }}
            >
              <input
                type="radio"
                name="font_size"
                value={opt.value}
                checked={form.font_size === opt.value}
                onChange={() => setForm((prev) => ({ ...prev, font_size: opt.value }))}
                style={{ accentColor: 'var(--color-primary, #6366f1)' }}
              />
              <span style={{ fontWeight: 500, fontSize: 'var(--font-size-sm)' }}>{opt.label}</span>
            </label>
          ))}
        </div>
      </div>

      {/* QR Size */}
      <div className="sp-card">
        <h3 className="sp-card__title">QR-Code Groesse</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          {([
            { value: 'small', label: 'Klein' },
            { value: 'medium', label: 'Mittel' },
            { value: 'large', label: 'Gross' },
          ] as const).map((opt) => (
            <label
              key={opt.value}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 'var(--spacing-2)',
                padding: 'var(--spacing-2) var(--spacing-4)',
                border: form.qr_size === opt.value ? '2px solid var(--color-primary, #6366f1)' : '2px solid var(--color-border, #ddd)',
                borderRadius: 'var(--radius-md)',
                cursor: 'pointer',
                background: form.qr_size === opt.value ? 'var(--color-primary-light, #eef2ff)' : 'transparent',
                transition: 'all 0.15s ease',
              }}
            >
              <input
                type="radio"
                name="qr_size"
                value={opt.value}
                checked={form.qr_size === opt.value}
                onChange={() => setForm((prev) => ({ ...prev, qr_size: opt.value }))}
                style={{ accentColor: 'var(--color-primary, #6366f1)' }}
              />
              <span style={{ fontWeight: 500, fontSize: 'var(--font-size-sm)' }}>{opt.label}</span>
            </label>
          ))}
        </div>
      </div>

      {/* Printer Type */}
      <div className="sp-card">
        <h3 className="sp-card__title">Druckertyp</h3>
        <div className="sp-grid">
          <div className="sp-field sp-full">
            <label className="sp-label">Bevorzugter Drucker</label>
            <select
              className="sp-select"
              value={form.printer_type}
              onChange={(e) =>
                setForm((prev) => ({
                  ...prev,
                  printer_type: e.target.value as 'browser' | 'zebra',
                }))
              }
            >
              <option value="browser">Browser-Druck (Standard)</option>
              <option value="zebra">Zebra Thermodrucker (ZPL)</option>
            </select>
          </div>
          {form.printer_type === 'zebra' && (
            <div className="sp-field sp-full">
              <p style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', margin: 0 }}>
                Fuer Zebra-Drucker wird ZPL II Code generiert. Stellen Sie sicher, dass Ihr Zebra-Drucker ueber das Netzwerk erreichbar ist.
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="sp-footer">
        {saveSuccess && <span className="sp-msg--success">Einstellungen gespeichert!</span>}
        {saveError && <span className="sp-msg--error">Fehler beim Speichern</span>}
        <button
          className="sp-btn sp-btn--secondary"
          onClick={handlePrintPreview}
          style={{ marginRight: 'var(--spacing-2)' }}
        >
          Vorschau drucken
        </button>
        <button
          className="sp-btn sp-btn--primary"
          onClick={handleSave}
          disabled={saveMutation.isPending}
        >
          {saveMutation.isPending ? 'Speichern...' : 'Speichern'}
        </button>
      </div>
    </div>
  )
}

export default LabelSettingsPage
