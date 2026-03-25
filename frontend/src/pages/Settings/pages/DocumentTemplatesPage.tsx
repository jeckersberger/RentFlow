import { useState } from 'react'
import { FileText, Eye, Download, Palette, X } from 'lucide-react'
import '../Settings.scss'

const templates = [
  { type: 'invoice', label: 'Rechnung', description: 'Standard-Rechnungsvorlage mit Logo und Bankdaten', status: 'aktiv', previewPath: '/api/v1/invoices' },
  { type: 'quote', label: 'Angebot', description: 'Angebotsvorlage mit Artikelpositionen und Gültigkeit', status: 'aktiv' },
  { type: 'reminder_1', label: 'Zahlungserinnerung', description: 'Erste freundliche Zahlungserinnerung', status: 'aktiv' },
  { type: 'reminder_2', label: 'Mahnung (2. Stufe)', description: 'Zweite Mahnung mit Fristsetzung', status: 'aktiv' },
  { type: 'reminder_3', label: 'Letzte Mahnung', description: 'Letzte Mahnung vor rechtlichen Schritten', status: 'aktiv' },
  { type: 'credit_note', label: 'Gutschrift', description: 'Gutschrift / Stornierung einer Rechnung', status: 'aktiv' },
  { type: 'packing_list', label: 'Packliste', description: 'Packliste für Equipment-Zusammenstellung', status: 'aktiv' },
  { type: 'delivery_note', label: 'Lieferschein', description: 'Lieferschein mit Equipment und Unterschrift', status: 'aktiv' },
]

function DocumentTemplatesPage() {
  const [previewType, setPreviewType] = useState<string | null>(null)
  const [showCustomize, setShowCustomize] = useState(false)

  const selectedTemplate = templates.find(t => t.type === previewType)

  return (
    <div className="sp-page">
      <div className="sp-header">
        <div>
          <h1>Dokumentvorlagen</h1>
          <p>Verwalten Sie Ihre Vorlagen für Rechnungen, Angebote, Mahnungen und mehr.</p>
        </div>
        <button className="sp-btn sp-btn--primary" onClick={() => setShowCustomize(true)}>
          <Palette size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
          Design anpassen
        </button>
      </div>

      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="sp-table">
          <thead>
            <tr>
              <th></th>
              <th>Vorlage</th>
              <th>Beschreibung</th>
              <th>Status</th>
              <th style={{ width: 180 }}>Aktionen</th>
            </tr>
          </thead>
          <tbody>
            {templates.map((t) => (
              <tr key={t.type}>
                <td style={{ width: 40 }}>
                  <FileText size={18} style={{ color: 'var(--color-primary)', opacity: 0.7 }} />
                </td>
                <td style={{ fontWeight: 500 }}>{t.label}</td>
                <td style={{ color: 'var(--color-text-secondary)' }}>{t.description}</td>
                <td>
                  <span className="badge badge--success" style={{ fontSize: 'var(--font-size-xs)' }}>
                    {t.status}
                  </span>
                </td>
                <td>
                  <div style={{ display: 'flex', gap: 6 }}>
                    <button
                      className="sp-btn sp-btn--secondary"
                      style={{ padding: '4px 10px', fontSize: 'var(--font-size-xs)' }}
                      onClick={() => setPreviewType(t.type)}
                    >
                      <Eye size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                      Vorschau
                    </button>
                    <button
                      className="sp-btn sp-btn--secondary"
                      style={{ padding: '4px 10px', fontSize: 'var(--font-size-xs)' }}
                      onClick={() => {
                        const w = window.open('', '_blank')
                        if (w) {
                          w.document.write(getSampleHTML(t.type))
                          w.document.close()
                          w.print()
                        }
                      }}
                    >
                      <Download size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                      Drucken
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Template Preview Modal */}
      {previewType && selectedTemplate && (
        <div style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          background: 'rgba(0,0,0,0.7)', zIndex: 9999,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }} onClick={() => setPreviewType(null)}>
          <div style={{
            background: 'white', borderRadius: 12, width: '90%', maxWidth: 900,
            height: '85vh', display: 'flex', flexDirection: 'column',
            overflow: 'hidden', boxShadow: '0 20px 60px rgba(0,0,0,0.5)',
          }} onClick={e => e.stopPropagation()}>
            <div style={{
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              padding: '16px 24px', borderBottom: '1px solid #e5e7eb',
              background: '#f9fafb',
            }}>
              <h3 style={{ margin: 0, color: '#111827' }}>
                Vorschau: {selectedTemplate.label}
              </h3>
              <button onClick={() => setPreviewType(null)} style={{
                background: 'none', border: 'none', cursor: 'pointer', padding: 4,
              }}>
                <X size={20} color="#6b7280" />
              </button>
            </div>
            <div style={{ flex: 1, overflow: 'auto', padding: 0 }}>
              <iframe
                srcDoc={getSampleHTML(previewType)}
                style={{ width: '100%', height: '100%', border: 'none' }}
                title={`Vorschau ${selectedTemplate.label}`}
              />
            </div>
          </div>
        </div>
      )}

      {/* Design Customization Modal */}
      {showCustomize && (
        <div style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          background: 'rgba(0,0,0,0.7)', zIndex: 9999,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }} onClick={() => setShowCustomize(false)}>
          <div style={{
            background: 'var(--color-bg-card)', borderRadius: 12,
            width: '90%', maxWidth: 600, maxHeight: '80vh',
            overflow: 'auto', padding: 24,
          }} onClick={e => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 20 }}>
              <h2 style={{ margin: 0 }}>Design anpassen</h2>
              <button onClick={() => setShowCustomize(false)} style={{
                background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-secondary)',
              }}>
                <X size={20} />
              </button>
            </div>
            <p style={{ color: 'var(--color-text-secondary)', marginBottom: 24, fontSize: 'var(--font-size-sm)' }}>
              Diese Einstellungen werden auf alle Dokumentvorlagen angewendet. Die Daten werden aus Ihrem Impressum übernommen.
            </p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <div>
                <label className="sp-label">Primärfarbe</label>
                <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                  <input type="color" defaultValue="#00d4ff" style={{ width: 40, height: 32, border: 'none', cursor: 'pointer' }} />
                  <input type="text" className="sp-input" defaultValue="#00d4ff" style={{ flex: 1 }} />
                </div>
              </div>
              <div>
                <label className="sp-label">Logo-URL</label>
                <input type="text" className="sp-input" placeholder="https://example.com/logo.png" />
              </div>
              <div>
                <label className="sp-label">Schriftart</label>
                <select className="sp-input">
                  <option>Arial (Standard)</option>
                  <option>Helvetica</option>
                  <option>Open Sans</option>
                  <option>Roboto</option>
                </select>
              </div>
              <div>
                <label className="sp-label">Fußzeile (alle Dokumente)</label>
                <textarea className="sp-input" rows={3} defaultValue="JE-Sound & Light | Inhaber: Janis Eckersberger" style={{ resize: 'vertical' }} />
              </div>
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginTop: 24 }}>
              <button className="sp-btn sp-btn--secondary" onClick={() => setShowCustomize(false)}>Abbrechen</button>
              <button className="sp-btn sp-btn--primary" onClick={() => setShowCustomize(false)}>Speichern</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function getSampleHTML(type: string): string {
  const styles = `
    body { font-family: Arial, sans-serif; max-width: 850px; margin: 0 auto; padding: 40px; color: #333; }
    .header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 40px; }
    .company { font-size: 12px; color: #666; }
    .doc-title { font-size: 24px; font-weight: bold; color: #00d4ff; margin: 0 0 5px 0; }
    .meta { text-align: right; font-size: 13px; }
    .meta-label { color: #999; }
    .client { margin: 30px 0; padding: 15px; background: #f8f9fa; border-radius: 6px; }
    table { width: 100%; border-collapse: collapse; margin: 30px 0; }
    th { background: #f0f0f0; border-bottom: 2px solid #00d4ff; padding: 10px; text-align: left; font-size: 13px; }
    td { border-bottom: 1px solid #eee; padding: 10px; font-size: 13px; }
    .total { text-align: right; margin-top: 20px; }
    .total-row { display: flex; justify-content: flex-end; gap: 40px; padding: 5px 0; }
    .total-final { font-weight: bold; font-size: 18px; border-top: 2px solid #333; padding-top: 10px; }
    .footer { margin-top: 60px; border-top: 1px solid #ddd; padding-top: 15px; font-size: 11px; color: #999; text-align: center; }
    .bank { margin-top: 30px; padding: 15px; background: #f0f7ff; border: 1px solid #b3d4fc; border-radius: 6px; font-size: 12px; }
    @media print { body { padding: 0; } }
  `

  const header = `
    <div class="header">
      <div>
        <div class="doc-title">${getDocTitle(type)}</div>
        <div class="company">
          JE-Sound & Light<br>
          Janis Eckersberger<br>
          Musterstraße 1, 12345 Musterstadt
        </div>
      </div>
      <div class="meta">
        <div><span class="meta-label">Nummer:</span> RF-2026-0001</div>
        <div><span class="meta-label">Datum:</span> 25.03.2026</div>
        <div><span class="meta-label">Fällig:</span> 24.04.2026</div>
      </div>
    </div>
  `

  const client = `
    <div class="client">
      <strong>Musterfirma GmbH</strong><br>
      Max Mustermann<br>
      Beispielweg 42, 80331 München
    </div>
  `

  const items = `
    <table>
      <thead><tr><th>Pos.</th><th>Beschreibung</th><th style="text-align:right">Menge</th><th style="text-align:right">Einzelpreis</th><th style="text-align:right">Gesamt</th></tr></thead>
      <tbody>
        <tr><td>1</td><td>PA-System QSC K12.2 (2x)</td><td style="text-align:right">2 Tage</td><td style="text-align:right">120,00 €</td><td style="text-align:right">240,00 €</td></tr>
        <tr><td>2</td><td>Lichtset Moving Heads (4x)</td><td style="text-align:right">2 Tage</td><td style="text-align:right">200,00 €</td><td style="text-align:right">400,00 €</td></tr>
        <tr><td>3</td><td>Techniker (Aufbau + Betreuung)</td><td style="text-align:right">8 Std.</td><td style="text-align:right">45,00 €</td><td style="text-align:right">360,00 €</td></tr>
      </tbody>
    </table>
    <div class="total">
      <div class="total-row"><span>Nettobetrag:</span><span>1.000,00 €</span></div>
      <div class="total-row" style="font-size:12px;color:#999"><span>Gemäß §19 UStG wird keine Umsatzsteuer berechnet.</span></div>
      <div class="total-row total-final"><span>Gesamtbetrag:</span><span>1.000,00 €</span></div>
    </div>
  `

  const bank = `
    <div class="bank">
      <strong>Bankverbindung:</strong> Sparkasse Musterstadt | IBAN: DE89 3704 0044 0532 0130 00 | BIC: COBADEFFXXX
    </div>
  `

  const footer = `<div class="footer">JE-Sound & Light | Inhaber: Janis Eckersberger | Steuernummer: 123/456/78901</div>`

  if (type === 'packing_list') {
    return `<!DOCTYPE html><html><head><style>${styles}</style></head><body>
      ${header.replace(getDocTitle('invoice'), 'PACKLISTE')}
      <div class="client"><strong>Projekt:</strong> Firmen-Gala TechCorp<br><strong>Datum:</strong> 28.03.2026<br><strong>Verantwortlich:</strong> Janis Eckersberger</div>
      <table>
        <thead><tr><th>☐</th><th>Equipment</th><th>SKU</th><th style="text-align:right">Menge</th><th>Zustand</th></tr></thead>
        <tbody>
          <tr><td>☐</td><td>QSC K12.2</td><td>QSC-K12-001</td><td style="text-align:right">2</td><td>Gut</td></tr>
          <tr><td>☐</td><td>Moving Head Wash</td><td>MH-W-003</td><td style="text-align:right">4</td><td>Gut</td></tr>
          <tr><td>☐</td><td>XLR Kabel 10m</td><td>XLR-10-012</td><td style="text-align:right">6</td><td>Gut</td></tr>
          <tr><td>☐</td><td>Stativ Speaker</td><td>ST-SP-008</td><td style="text-align:right">2</td><td>Gut</td></tr>
        </tbody>
      </table>
      <div style="margin-top:60px;border-top:1px solid #ccc;padding-top:20px">
        <div style="display:flex;justify-content:space-between">
          <div><strong>Gepackt von:</strong> ____________________</div>
          <div><strong>Datum:</strong> ____________________</div>
          <div><strong>Unterschrift:</strong> ____________________</div>
        </div>
      </div>
      ${footer}
    </body></html>`
  }

  if (type === 'delivery_note') {
    return `<!DOCTYPE html><html><head><style>${styles}</style></head><body>
      ${header.replace(getDocTitle('invoice'), 'LIEFERSCHEIN')}
      ${client}
      <table>
        <thead><tr><th>Pos.</th><th>Beschreibung</th><th>Seriennummer</th><th style="text-align:right">Menge</th></tr></thead>
        <tbody>
          <tr><td>1</td><td>QSC K12.2</td><td>QSC-K12-001</td><td style="text-align:right">2</td></tr>
          <tr><td>2</td><td>Moving Head Wash</td><td>MH-W-003</td><td style="text-align:right">4</td></tr>
        </tbody>
      </table>
      <div style="margin-top:60px">
        <div style="display:flex;justify-content:space-between">
          <div style="width:45%"><strong>Übergabe:</strong><br><br>____________________<br>Datum, Unterschrift Absender</div>
          <div style="width:45%"><strong>Empfang:</strong><br><br>____________________<br>Datum, Unterschrift Empfänger</div>
        </div>
      </div>
      ${footer}
    </body></html>`
  }

  return `<!DOCTYPE html><html><head><style>${styles}</style></head><body>${header}${client}${items}${bank}${footer}</body></html>`
}

function getDocTitle(type: string): string {
  switch (type) {
    case 'invoice': return 'RECHNUNG'
    case 'quote': return 'ANGEBOT'
    case 'reminder_1': return 'ZAHLUNGSERINNERUNG'
    case 'reminder_2': return 'MAHNUNG (2. STUFE)'
    case 'reminder_3': return 'LETZTE MAHNUNG'
    case 'credit_note': return 'GUTSCHRIFT'
    case 'packing_list': return 'PACKLISTE'
    case 'delivery_note': return 'LIEFERSCHEIN'
    default: return 'DOKUMENT'
  }
}

export default DocumentTemplatesPage
