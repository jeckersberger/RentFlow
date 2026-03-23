import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import '../Settings.module.scss'

interface LabelConfig {
  default_size: 'small' | 'medium' | 'large' | 'flightcase'
  show_company_name: boolean
  show_qr_code: boolean
  show_barcode: boolean
  show_name: boolean
  show_sku: boolean
  printer_type: 'browser' | 'zebra'
}

const DEFAULT_CONFIG: LabelConfig = {
  default_size: 'medium',
  show_company_name: true,
  show_qr_code: true,
  show_barcode: true,
  show_name: true,
  show_sku: true,
  printer_type: 'browser',
}

function LabelSettingsPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [saveError, setSaveError] = useState(false)

  const [form, setForm] = useState<LabelConfig>(DEFAULT_CONFIG)

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'labels'],
    queryFn: () => configApi.get('labels'),
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
    mutationFn: (data: LabelConfig) => configApi.set('labels', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'labels'] })
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

  if (isLoading) {
    return <div className="sp-loading">Label-Einstellungen werden geladen...</div>
  }

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Label-Einstellungen</h1>
        <p>Konfigurieren Sie das Standardverhalten f\u00fcr den Label-Druck.</p>
      </div>

      {/* Default Label Size */}
      <div className="sp-card">
        <h3 className="sp-card__title">Standard-Labelgr\u00f6\u00dfe</h3>
        <div className="sp-grid">
          <div className="sp-field sp-full">
            <label className="sp-label">Standardformat</label>
            <select
              className="sp-select"
              value={form.default_size}
              onChange={(e) =>
                setForm((prev) => ({
                  ...prev,
                  default_size: e.target.value as LabelConfig['default_size'],
                }))
              }
            >
              <option value="small">Klein (25x15mm) \u2014 nur QR-Code</option>
              <option value="medium">Mittel (50x25mm) \u2014 QR + Name + SKU</option>
              <option value="large">Gro\u00df (100x50mm) \u2014 QR + Barcode + alle Infos</option>
              <option value="flightcase">Flightcase (A5) \u2014 Inhaltsliste</option>
            </select>
          </div>
        </div>
      </div>

      {/* Label Content */}
      <div className="sp-card">
        <h3 className="sp-card__title">Label-Inhalt</h3>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
          <label className="sp-toggle" onClick={() => toggleField('show_company_name')}>
            <div className="sp-toggle__label">
              <strong>Firmenname anzeigen</strong>
              <span>Zeigt den Firmennamen auf dem Label an</span>
            </div>
            <input type="checkbox" checked={form.show_company_name} readOnly />
          </label>

          <label className="sp-toggle" onClick={() => toggleField('show_qr_code')}>
            <div className="sp-toggle__label">
              <strong>QR-Code anzeigen</strong>
              <span>QR-Code mit Equipment-ID oder Barcode</span>
            </div>
            <input type="checkbox" checked={form.show_qr_code} readOnly />
          </label>

          <label className="sp-toggle" onClick={() => toggleField('show_barcode')}>
            <div className="sp-toggle__label">
              <strong>Barcode anzeigen</strong>
              <span>Code128-Barcode mit SKU/Barcode-Wert</span>
            </div>
            <input type="checkbox" checked={form.show_barcode} readOnly />
          </label>

          <label className="sp-toggle" onClick={() => toggleField('show_name')}>
            <div className="sp-toggle__label">
              <strong>Ger\u00e4tename anzeigen</strong>
              <span>Name der Ausr\u00fcstung</span>
            </div>
            <input type="checkbox" checked={form.show_name} readOnly />
          </label>

          <label className="sp-toggle" onClick={() => toggleField('show_sku')}>
            <div className="sp-toggle__label">
              <strong>SKU / Code anzeigen</strong>
              <span>Artikelnummer oder Barcode als Text</span>
            </div>
            <input type="checkbox" checked={form.show_sku} readOnly />
          </label>
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
                F\u00fcr Zebra-Drucker wird ZPL II Code generiert. Das Backend unterst\u00fctzt das
                Standardformat 50x25mm. Stellen Sie sicher, dass Ihr Zebra-Drucker \u00fcber
                das Netzwerk erreichbar ist.
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Save */}
      <div className="sp-footer">
        {saveSuccess && <span className="sp-msg--success">Einstellungen gespeichert!</span>}
        {saveError && <span className="sp-msg--error">Fehler beim Speichern</span>}
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
