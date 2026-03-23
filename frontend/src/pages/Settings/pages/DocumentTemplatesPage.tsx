import { FileText, Download, Lock } from 'lucide-react'
import '../Settings.scss'

const templates = [
  { type: 'invoice', label: 'Rechnung', description: 'Standard-Rechnungsvorlage mit Logo und Bankdaten', status: 'aktiv' },
  { type: 'quote', label: 'Angebot', description: 'Angebotsvorlage mit Artikelpositionen und Gültigkeit', status: 'aktiv' },
  { type: 'reminder_1', label: 'Zahlungserinnerung', description: 'Erste freundliche Zahlungserinnerung', status: 'aktiv' },
  { type: 'reminder_2', label: 'Mahnung (2. Stufe)', description: 'Zweite Mahnung mit Fristsetzung', status: 'aktiv' },
  { type: 'reminder_3', label: 'Letzte Mahnung', description: 'Letzte Mahnung vor rechtlichen Schritten', status: 'aktiv' },
  { type: 'credit_note', label: 'Gutschrift', description: 'Gutschrift / Stornierung einer Rechnung', status: 'aktiv' },
  { type: 'packing_list', label: 'Packliste', description: 'Packliste für Equipment-Zusammenstellung', status: 'aktiv' },
  { type: 'delivery_note', label: 'Lieferschein', description: 'Lieferschein mit Equipment und Unterschrift', status: 'aktiv' },
]

function DocumentTemplatesPage() {
  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Dokumentvorlagen</h1>
        <p>Verwalten Sie Ihre Vorlagen für Rechnungen, Angebote, Mahnungen und mehr.</p>
      </div>

      <div style={{
        display: 'flex', alignItems: 'center', gap: 'var(--spacing-3)',
        padding: 'var(--spacing-4)', borderRadius: 'var(--radius-md)',
        background: 'rgba(99, 102, 241, 0.08)', border: '1px solid rgba(99, 102, 241, 0.2)',
        marginBottom: 'var(--spacing-5)', color: 'var(--color-text-secondary)',
        fontSize: 'var(--font-size-sm)',
      }}>
        <Lock size={16} style={{ flexShrink: 0, color: 'var(--color-primary)' }} />
        <span>Der visuelle Vorlagen-Editor ist <strong>bald verfügbar</strong>. Aktuell werden die Standard-Vorlagen verwendet.</span>
      </div>

      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="sp-table">
          <thead>
            <tr>
              <th></th>
              <th>Vorlage</th>
              <th>Beschreibung</th>
              <th>Status</th>
              <th style={{ width: 100 }}>Aktion</th>
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
                  <button className="sp-btn sp-btn--secondary" disabled style={{ padding: '4px 10px', fontSize: 'var(--font-size-xs)', opacity: 0.5, cursor: 'not-allowed' }}>
                    <Download size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                    Bearbeiten
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export default DocumentTemplatesPage
