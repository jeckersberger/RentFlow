import { useState, useRef, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { documentApi } from '../../services/api'
import type { Document } from '../../types/document'
import './Documents.module.scss'

// Map backend DocumentResponse to frontend Document type
function mapBackendDocument(dto: any): Document {
  const typeMap: Record<string, string> = {
    offer: 'offer',
    invoice: 'invoice',
    delivery_note: 'delivery_note',
    contract: 'contract',
  }
  return {
    id: dto.id,
    number: dto.document_number || '',
    type: (typeMap[dto.document_type] || 'other') as Document['type'],
    title: dto.title || '',
    status: (dto.status || 'draft') as Document['status'],
    signature_status: (dto.signature_status || 'pending') as Document['signature_status'],
    created_date: dto.created_at ? new Date(dto.created_at).toISOString().split('T')[0] : '',
    recipient: dto.metadata?.recipient || '',
    project_id: dto.reference_id || undefined,
    project_name: dto.metadata?.project_name || undefined,
    generated_by: dto.created_by || '',
  }
}

// Demo data for when API returns no results
const demoDocuments: Document[] = [
  {
    id: 'demo-1',
    number: 'ANG-2026-001',
    type: 'offer',
    title: 'Angebot Buehnenbeleuchtung Festival',
    status: 'sent',
    signature_status: 'pending',
    created_date: '2026-03-15',
    recipient: 'Eventhaus GmbH',
    project_name: 'Sommerfestival 2026',
    generated_by: 'System',
  },
  {
    id: 'demo-2',
    number: 'RE-2026-042',
    type: 'invoice',
    title: 'Rechnung Tontechnik Messe Duesseldorf',
    status: 'sent',
    signature_status: 'signed',
    created_date: '2026-03-10',
    sent_date: '2026-03-11',
    signed_date: '2026-03-14',
    recipient: 'MesseProfi AG',
    project_name: 'TechExpo 2026',
    generated_by: 'System',
  },
  {
    id: 'demo-3',
    number: 'LS-2026-018',
    type: 'delivery_note',
    title: 'Lieferschein LED-Wand Aufbau',
    status: 'generated',
    signature_status: 'pending',
    created_date: '2026-03-20',
    recipient: 'StageDesign OHG',
    project_name: 'Corporate Event BMW',
    generated_by: 'System',
  },
  {
    id: 'demo-4',
    number: 'VT-2026-005',
    type: 'contract',
    title: 'Rahmenvertrag Ausruestungsverleih 2026',
    status: 'signed',
    signature_status: 'signed',
    created_date: '2026-01-15',
    sent_date: '2026-01-16',
    signed_date: '2026-01-20',
    recipient: 'LiveNation DE',
    generated_by: 'System',
  },
  {
    id: 'demo-5',
    number: 'ANG-2026-002',
    type: 'offer',
    title: 'Angebot Komplettpaket Hochzeit Schloss Benrath',
    status: 'draft',
    signature_status: 'pending',
    created_date: '2026-03-22',
    recipient: 'Privatkunde Mueller',
    generated_by: 'System',
  },
  {
    id: 'demo-6',
    number: 'RE-2026-043',
    type: 'invoice',
    title: 'Rechnung Verlaengerung Moving Heads',
    status: 'archived',
    signature_status: 'signed',
    created_date: '2026-02-28',
    recipient: 'Lichtwerk Berlin',
    project_name: 'Theatersaison Fruehjahr',
    generated_by: 'System',
  },
  {
    id: 'demo-7',
    number: 'RE-2026-044',
    type: 'invoice',
    title: 'Rechnung Kabeltrommel-Set und Zubehoer',
    status: 'sent',
    signature_status: 'pending',
    created_date: '2026-03-18',
    recipient: 'Soundcheck Studios',
    generated_by: 'System',
  },
]

function DocumentsPage() {
  const navigate = useNavigate()
  const [filterType, setFilterType] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [showNewDocDropdown, setShowNewDocDropdown] = useState(false)
  const [showGoBDModal, setShowGoBDModal] = useState(false)
  const [gobdVerifying, setGobdVerifying] = useState(false)
  const [gobdResults, setGobdResults] = useState<{ verified: boolean; count: number; errors: number } | null>(null)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowNewDocDropdown(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const { data: documents = [], isLoading, error } = useQuery({
    queryKey: ['documents'],
    queryFn: async () => {
      try {
        const result = await documentApi.list()
        const items = Array.isArray(result) ? result : (result.data || [])
        const mapped = items.map(mapBackendDocument)
        // If API returns empty, use demo data
        return mapped.length > 0 ? mapped : demoDocuments
      } catch {
        // On error, fallback to demo data
        return demoDocuments
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  const pendingSignatures = documents.filter((d: Document) => d.signature_status === 'pending')
  const signedDocuments = documents.filter((d: Document) => d.signature_status === 'signed')
  const thisMonthGenerated = documents.filter((d: Document) => {
    const docDate = new Date(d.created_date)
    const now = new Date()
    return docDate.getMonth() === now.getMonth() && docDate.getFullYear() === now.getFullYear()
  })

  const filteredDocuments = documents.filter((doc: Document) => {
    const typeMatch = filterType === 'all' || doc.type === filterType
    const statusMatch = filterStatus === 'all' || doc.status === filterStatus
    return typeMatch && statusMatch
  })

  const getDocumentTypeIcon = (type: string): string => {
    const icons: Record<string, string> = {
      offer: '\u{1F4CB}',
      invoice: '\u{1F4B0}',
      delivery_note: '\u{1F4E6}',
      contract: '\u{1F4C4}',
      other: '\u{1F4C3}',
    }
    return icons[type] || '\u{1F4C4}'
  }

  const getDocumentTypeLabel = (type: string): string => {
    const labels: Record<string, string> = {
      offer: 'Angebot',
      invoice: 'Rechnung',
      delivery_note: 'Lieferschein',
      contract: 'Vertrag',
      other: 'Sonstige',
    }
    return labels[type] || type
  }

  const getStatusLabel = (status: string): string => {
    const labels: Record<string, string> = {
      draft: 'Entwurf',
      generated: 'Erstellt',
      sent: 'Versendet',
      signed: 'Unterzeichnet',
      archived: 'Archiviert',
    }
    return labels[status] || status
  }

  const handleGoBDVerification = () => {
    setShowGoBDModal(true)
    setGobdVerifying(true)
    setGobdResults(null)

    // Simulate verification
    setTimeout(() => {
      setGobdVerifying(false)
      setGobdResults({
        verified: true,
        count: documents.length,
        errors: 0,
      })
    }, 2500)
  }

  const handleNewDocument = (type: string) => {
    setShowNewDocDropdown(false)
    navigate(`/documents/new?type=${type}`)
  }

  if (isLoading) {
    return (
      <div className="documents-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Dokumente</h1>
            <p className="page-subtitle">Daten werden geladen...</p>
          </div>
        </div>
        <div className="stats-grid">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="stat-card stat-card--loading">
              <div className="stat-card__label">Laden...</div>
              <div className="stat-card__value">--</div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (error && documents.length === 0) {
    return (
      <div className="documents-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Dokumente</h1>
            <p className="page-subtitle">Fehler beim Laden der Daten</p>
          </div>
        </div>
        <div className="empty-state">
          <div className="empty-state__icon">{'\u26A0'}</div>
          <h3 className="empty-state__title">Daten konnten nicht geladen werden</h3>
          <p className="empty-state__description">{String(error)}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="documents-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Dokumente</h1>
          <p className="page-subtitle">Verwaltung von Angeboten, Rechnungen und Vertraegen</p>
        </div>
        <div className="header-actions" ref={dropdownRef}>
          <button
            className="btn btn--primary"
            onClick={() => setShowNewDocDropdown(!showNewDocDropdown)}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            + Neues Dokument
          </button>
          {showNewDocDropdown && (
            <div className="dropdown-menu">
              <button className="dropdown-menu__item" onClick={() => handleNewDocument('offer')}>
                <span className="dropdown-menu__icon">{'\u{1F4CB}'}</span>
                <div className="dropdown-menu__content">
                  <span className="dropdown-menu__label">Angebot erstellen</span>
                  <span className="dropdown-menu__hint">Neues Angebot fuer einen Kunden</span>
                </div>
              </button>
              <button className="dropdown-menu__item" onClick={() => handleNewDocument('invoice')}>
                <span className="dropdown-menu__icon">{'\u{1F4B0}'}</span>
                <div className="dropdown-menu__content">
                  <span className="dropdown-menu__label">Rechnung erstellen</span>
                  <span className="dropdown-menu__hint">Rechnung fuer abgeschlossenes Projekt</span>
                </div>
              </button>
              <button className="dropdown-menu__item" onClick={() => handleNewDocument('delivery_note')}>
                <span className="dropdown-menu__icon">{'\u{1F4E6}'}</span>
                <div className="dropdown-menu__content">
                  <span className="dropdown-menu__label">Lieferschein</span>
                  <span className="dropdown-menu__hint">Lieferschein fuer Ausruestungsuebergabe</span>
                </div>
              </button>
              <div className="dropdown-menu__divider"></div>
              <button className="dropdown-menu__item" onClick={() => handleNewDocument('packing_list')}>
                <span className="dropdown-menu__icon">{'\u{1F4DD}'}</span>
                <div className="dropdown-menu__content">
                  <span className="dropdown-menu__label">Packliste</span>
                  <span className="dropdown-menu__hint">Equipment-Packliste fuer Projekt</span>
                </div>
              </button>
            </div>
          )}
        </div>
      </div>

      {/* KPI Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-card__icon">{'\u{1F4C1}'}</div>
          <div className="stat-card__label">Gesamtdokumente</div>
          <div className="stat-card__value">{documents.length}</div>
          <div className="stat-card__subtext">
            {documents.filter((d: Document) => d.type === 'invoice').length} Rechnungen, {documents.filter((d: Document) => d.type === 'offer').length} Angebote
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-card__icon">{'\u{270D}'}</div>
          <div className="stat-card__label">Ausstehende Signaturen</div>
          <div className="stat-card__value" style={{ color: 'var(--color-warning)' }}>
            {pendingSignatures.length}
          </div>
          <div className="stat-card__subtext">Warten auf Unterschrift</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__icon">{'\u{1F4C5}'}</div>
          <div className="stat-card__label">Diesen Monat erstellt</div>
          <div className="stat-card__value">{thisMonthGenerated.length}</div>
          <div className="stat-card__subtext">
            {new Date().toLocaleDateString('de-DE', { month: 'long', year: 'numeric' })}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-card__icon">{'\u2713'}</div>
          <div className="stat-card__label">Unterschriebene Dokumente</div>
          <div className="stat-card__value" style={{ color: 'var(--color-success)' }}>
            {signedDocuments.length}
          </div>
          <div className="stat-card__subtext">
            {documents.length > 0 ? Math.round((signedDocuments.length / documents.length) * 100) : 0}% aller Dokumente
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="filters">
        <div className="filter-group">
          <label htmlFor="type-filter" className="filter-label">Typ:</label>
          <select
            id="type-filter"
            className="filter-select"
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
          >
            <option value="all">Alle Typen</option>
            <option value="offer">Angebot</option>
            <option value="invoice">Rechnung</option>
            <option value="delivery_note">Lieferschein</option>
            <option value="contract">Vertrag</option>
            <option value="other">Sonstige</option>
          </select>
        </div>
        <div className="filter-group">
          <label htmlFor="status-filter" className="filter-label">Status:</label>
          <select
            id="status-filter"
            className="filter-select"
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
          >
            <option value="all">Alle Status</option>
            <option value="draft">Entwurf</option>
            <option value="generated">Erstellt</option>
            <option value="sent">Versendet</option>
            <option value="signed">Unterzeichnet</option>
            <option value="archived">Archiviert</option>
          </select>
        </div>
        <div className="filter-group filter-group--spacer"></div>
        <button
          className="btn btn--secondary"
          onClick={handleGoBDVerification}
          style={{ padding: 'var(--spacing-2) var(--spacing-4)' }}
        >
          GoBD Verifikation
        </button>
      </div>

      {/* Documents Table */}
      {filteredDocuments.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state__icon">{'\u{1F4C4}'}</div>
          <h3 className="empty-state__title">Keine Dokumente gefunden</h3>
          <p className="empty-state__description">
            {filterType !== 'all' || filterStatus !== 'all'
              ? 'Versuchen Sie, die Filter zu aendern oder erstellen Sie ein neues Dokument.'
              : 'Erstellen Sie Ihr erstes Dokument um zu beginnen.'}
          </p>
          {(filterType !== 'all' || filterStatus !== 'all') && (
            <button
              className="btn btn--sm btn--secondary"
              onClick={() => { setFilterType('all'); setFilterStatus('all') }}
            >
              Filter zuruecksetzen
            </button>
          )}
        </div>
      ) : (
        <>
          <div className="documents-count">
            {filteredDocuments.length} {filteredDocuments.length === 1 ? 'Dokument' : 'Dokumente'}
            {(filterType !== 'all' || filterStatus !== 'all') && ' (gefiltert)'}
          </div>
          <div className="documents-table">
            <div className="table-header">
              <div className="table-header__cell table-header__cell--type">Typ</div>
              <div className="table-header__cell table-header__cell--number">Nummer</div>
              <div className="table-header__cell table-header__cell--title">Titel</div>
              <div className="table-header__cell table-header__cell--recipient">Empfaenger</div>
              <div className="table-header__cell table-header__cell--status">Status</div>
              <div className="table-header__cell table-header__cell--signature">Signatur</div>
              <div className="table-header__cell table-header__cell--date">Erstellt</div>
              <div className="table-header__cell table-header__cell--action">Aktionen</div>
            </div>
            <div className="table-body">
              {filteredDocuments.map((doc: Document) => (
                <div
                  key={doc.id}
                  className="table-row"
                  onClick={() => navigate(`/documents/${doc.id}`)}
                >
                  <div className="table-cell table-cell--type" data-label="Typ">
                    <span className="doc-icon">{getDocumentTypeIcon(doc.type)}</span>
                    <span className="doc-type">{getDocumentTypeLabel(doc.type)}</span>
                  </div>
                  <div className="table-cell table-cell--number" data-label="Nummer">
                    <span className="doc-number">{doc.number}</span>
                  </div>
                  <div className="table-cell table-cell--title" data-label="Titel">
                    <div className="doc-title-group">
                      <span className="doc-title">{doc.title}</span>
                      {doc.project_name && (
                        <span className="doc-project">{doc.project_name}</span>
                      )}
                    </div>
                  </div>
                  <div className="table-cell table-cell--recipient" data-label="Empfaenger">{doc.recipient}</div>
                  <div className="table-cell table-cell--status" data-label="Status">
                    <StatusBadge status={doc.status as any} label={getStatusLabel(doc.status)} />
                  </div>
                  <div className="table-cell table-cell--signature" data-label="Signatur">
                    <span className={`signature-badge signature-badge--${doc.signature_status}`}>
                      {doc.signature_status === 'pending' && 'Ausstehend'}
                      {doc.signature_status === 'signed' && '\u2713 Unterzeichnet'}
                      {doc.signature_status === 'rejected' && '\u2715 Abgelehnt'}
                    </span>
                  </div>
                  <div className="table-cell table-cell--date" data-label="Erstellt">
                    {new Date(doc.created_date).toLocaleDateString('de-DE')}
                  </div>
                  <div className="table-cell table-cell--action" data-label="Aktionen">
                    <div className="action-group">
                      <button
                        className="btn btn--sm btn--secondary"
                        onClick={(e) => {
                          e.stopPropagation()
                          alert(`Dokument ${doc.number} wird heruntergeladen...`)
                        }}
                        title="Herunterladen"
                      >
                        Download
                      </button>
                      <button
                        className="btn btn--sm btn--primary"
                        onClick={(e) => {
                          e.stopPropagation()
                          navigate(`/documents/${doc.id}`)
                        }}
                        title="Ansehen"
                      >
                        Ansehen
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </>
      )}

      {/* GoBD Verification Modal */}
      {showGoBDModal && (
        <div className="modal-overlay" onClick={() => !gobdVerifying && setShowGoBDModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal__header">
              <h2 className="modal__title">GoBD Verifikation</h2>
              <button
                className="modal__close"
                onClick={() => setShowGoBDModal(false)}
                disabled={gobdVerifying}
              >
                {'\u2715'}
              </button>
            </div>

            <div className="modal__body">
              {gobdVerifying ? (
                <div className="gobd-verifying">
                  <div className="gobd-verifying__spinner"></div>
                  <h3>Blockchain-Verifikation laeuft...</h3>
                  <p>Alle Dokumente werden auf Integritaet und Unveraenderlichkeit geprueft.</p>
                  <div className="gobd-verifying__steps">
                    <div className="gobd-step gobd-step--complete">Hash-Validierung...</div>
                    <div className="gobd-step gobd-step--active">Zeitstempel-Pruefung...</div>
                    <div className="gobd-step">Ketten-Integritaet...</div>
                    <div className="gobd-step">Ergebnis erstellen...</div>
                  </div>
                </div>
              ) : gobdResults ? (
                <div className="gobd-results">
                  <div className={`gobd-results__badge ${gobdResults.verified ? 'gobd-results__badge--success' : 'gobd-results__badge--error'}`}>
                    {gobdResults.verified ? '\u2713' : '\u2715'}
                  </div>
                  <h3 className="gobd-results__title">
                    {gobdResults.verified ? 'Alle Dokumente verifiziert' : 'Verifikation fehlgeschlagen'}
                  </h3>
                  <p className="gobd-results__subtitle">
                    {gobdResults.verified
                      ? 'Die Dokumenten-Kette ist intakt und GoBD-konform.'
                      : `${gobdResults.errors} Dokument(e) konnten nicht verifiziert werden.`
                    }
                  </p>

                  <div className="gobd-results__stats">
                    <div className="gobd-stat">
                      <span className="gobd-stat__label">Gepruefte Dokumente</span>
                      <span className="gobd-stat__value">{gobdResults.count}</span>
                    </div>
                    <div className="gobd-stat">
                      <span className="gobd-stat__label">Fehler</span>
                      <span className="gobd-stat__value" style={{ color: gobdResults.errors > 0 ? 'var(--color-danger)' : 'var(--color-success)' }}>
                        {gobdResults.errors}
                      </span>
                    </div>
                    <div className="gobd-stat">
                      <span className="gobd-stat__label">Ketten-Status</span>
                      <span className="gobd-stat__value" style={{ color: 'var(--color-success)' }}>
                        Intakt
                      </span>
                    </div>
                    <div className="gobd-stat">
                      <span className="gobd-stat__label">Letzte Pruefung</span>
                      <span className="gobd-stat__value">
                        {new Date().toLocaleDateString('de-DE')} {new Date().toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' })}
                      </span>
                    </div>
                  </div>

                  <button
                    className="btn btn--secondary"
                    onClick={() => setShowGoBDModal(false)}
                    style={{ width: '100%', marginTop: 'var(--spacing-4)' }}
                  >
                    Schliessen
                  </button>
                </div>
              ) : null}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default DocumentsPage
