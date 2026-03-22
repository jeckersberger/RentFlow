import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import type { Document } from '../../types/document'
import './Documents.module.scss'

const mockDocuments: Document[] = [
  {
    id: '1',
    number: 'ANG-2026-001',
    type: 'offer',
    title: 'Angebot: Ausrüstung Stadtfest München',
    status: 'generated',
    signature_status: 'pending',
    created_date: '2026-03-15',
    sent_date: '2026-03-15',
    project_id: 'p1',
    project_name: 'Stadtfest München 2026',
    recipient: 'München Stadtfest GmbH',
    generated_by: 'admin',
  },
  {
    id: '2',
    number: 'REC-2026-145',
    type: 'invoice',
    title: 'Rechnung: Bodensee Festival',
    status: 'signed',
    signature_status: 'signed',
    created_date: '2026-03-10',
    sent_date: '2026-03-10',
    signed_date: '2026-03-12',
    project_id: 'p3',
    project_name: 'Open Air Festival Bodensee',
    recipient: 'Festival Bodensee AG',
    generated_by: 'admin',
  },
  {
    id: '3',
    number: 'LN-2026-087',
    type: 'delivery_note',
    title: 'Lieferschein: TechCorp Event',
    status: 'sent',
    signature_status: 'pending',
    created_date: '2026-03-20',
    sent_date: '2026-03-21',
    project_id: 'p2',
    project_name: 'Firmen-Gala TechCorp',
    recipient: 'TechCorp AG',
    generated_by: 'user1',
  },
  {
    id: '4',
    number: 'VER-2026-003',
    type: 'contract',
    title: 'Vertrag: Langzeitmietung Ausrüstung',
    status: 'signed',
    signature_status: 'signed',
    created_date: '2026-03-05',
    sent_date: '2026-03-05',
    signed_date: '2026-03-08',
    recipient: 'RentFlow Customer',
    generated_by: 'admin',
  },
  {
    id: '5',
    number: 'ANG-2026-002',
    type: 'offer',
    title: 'Angebot: Spezialausrüstung',
    status: 'draft',
    signature_status: 'pending',
    created_date: '2026-03-22',
    recipient: 'Potential Client',
    generated_by: 'user1',
  },
  {
    id: '6',
    number: 'REC-2026-146',
    type: 'invoice',
    title: 'Rechnung: Wartungsservice',
    status: 'sent',
    signature_status: 'pending',
    created_date: '2026-03-18',
    sent_date: '2026-03-20',
    recipient: 'RentFlow Customer',
    generated_by: 'admin',
  },
  {
    id: '7',
    number: 'LN-2026-088',
    type: 'delivery_note',
    title: 'Lieferschein: Wartung Abschluss',
    status: 'archived',
    signature_status: 'signed',
    created_date: '2026-03-01',
    signed_date: '2026-03-02',
    recipient: 'Customer A',
    generated_by: 'user2',
  },
  {
    id: '8',
    number: 'ANG-2026-003',
    type: 'offer',
    title: 'Angebot: Miete Kran',
    status: 'sent',
    signature_status: 'rejected',
    created_date: '2026-03-19',
    sent_date: '2026-03-19',
    recipient: 'Customer B',
    generated_by: 'admin',
  },
  {
    id: '9',
    number: 'DOC-2026-001',
    type: 'other',
    title: 'Bedingungen und Konditionen',
    status: 'archived',
    signature_status: 'signed',
    created_date: '2026-02-01',
    recipient: 'Internal',
    generated_by: 'admin',
  },
  {
    id: '10',
    number: 'REC-2026-147',
    type: 'invoice',
    title: 'Rechnung: Februar 2026',
    status: 'draft',
    signature_status: 'pending',
    created_date: '2026-03-22',
    recipient: 'Monthly Billing',
    generated_by: 'accounting',
  },
]

function DocumentsPage() {
  const navigate = useNavigate()
  const [filterType, setFilterType] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')

  // TODO: Replace with documentsApi when backend endpoint is available
  const { data: documents = [], isLoading, error } = useQuery({
    queryKey: ['documents'],
    queryFn: async () => {
      // No documents API endpoint available yet - using mock data
      return mockDocuments
    },
    staleTime: 1000 * 60 * 5,
  })

  const pendingSignatures = documents.filter(d => d.signature_status === 'pending')
  const thisMonthGenerated = documents.filter(d => {
    const docDate = new Date(d.created_date)
    const now = new Date()
    return docDate.getMonth() === now.getMonth() && docDate.getFullYear() === now.getFullYear()
  })

  const filteredDocuments = documents.filter(doc => {
    const typeMatch = filterType === 'all' || doc.type === filterType
    const statusMatch = filterStatus === 'all' || doc.status === filterStatus
    return typeMatch && statusMatch
  })

  const getDocumentTypeIcon = (type: string): string => {
    const icons: Record<string, string> = {
      offer: '📋',
      invoice: '💰',
      delivery_note: '📦',
      contract: '📄',
      other: '📃',
    }
    return icons[type] || '📄'
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
            <div key={i} className="stat-card">
              <div className="stat-card__label">Laden...</div>
              <div className="stat-card__value">--</div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="documents-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Dokumente</h1>
            <p className="page-subtitle">Fehler beim Laden der Daten</p>
          </div>
        </div>
        <div className="empty-state">
          <div className="empty-state__icon">&#x26A0;</div>
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
          <p className="page-subtitle">Verwaltung von Angeboten, Rechnungen und Verträgen</p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => navigate('/documents/new')}
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
        >
          ➕ Neues Dokument
        </button>
      </div>

      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-card__label">Gesamtdokumente</div>
          <div className="stat-card__value">{documents.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Ausstehende Signaturen</div>
          <div className="stat-card__value" style={{ color: 'var(--color-warning)' }}>
            {pendingSignatures.length}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Diesen Monat erstellt</div>
          <div className="stat-card__value">{thisMonthGenerated.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Unterschriebene Dokumente</div>
          <div className="stat-card__value" style={{ color: 'var(--color-success)' }}>
            {documents.filter(d => d.signature_status === 'signed').length}
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
            <option value="all">Alle Statuses</option>
            <option value="draft">Entwurf</option>
            <option value="generated">Erstellt</option>
            <option value="sent">Versendet</option>
            <option value="signed">Unterzeichnet</option>
            <option value="archived">Archiviert</option>
          </select>
        </div>
        <button
          className="btn btn--secondary"
          onClick={() => alert('GoBD-Verifikation wird durchgeführt...')}
          style={{ padding: 'var(--spacing-2) var(--spacing-4)' }}
        >
          🔗 GoBD Verifikation
        </button>
      </div>

      {/* Documents Table */}
      {filteredDocuments.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state__icon">📄</div>
          <h3 className="empty-state__title">Keine Dokumente gefunden</h3>
          <p className="empty-state__description">Versuchen Sie, die Filter zu ändern oder erstellen Sie ein neues Dokument.</p>
        </div>
      ) : (
        <div className="documents-table">
          <div className="table-header">
            <div className="table-header__cell table-header__cell--type">Typ</div>
            <div className="table-header__cell table-header__cell--number">Nummer</div>
            <div className="table-header__cell table-header__cell--title">Titel</div>
            <div className="table-header__cell table-header__cell--recipient">Empfänger</div>
            <div className="table-header__cell table-header__cell--status">Status</div>
            <div className="table-header__cell table-header__cell--signature">Signatur</div>
            <div className="table-header__cell table-header__cell--date">Erstellt</div>
            <div className="table-header__cell table-header__cell--action"></div>
          </div>
          <div className="table-body">
            {filteredDocuments.map(doc => (
              <div
                key={doc.id}
                className="table-row"
                onClick={() => navigate(`/documents/${doc.id}`)}
              >
                <div className="table-cell table-cell--type">
                  <span className="doc-icon">{getDocumentTypeIcon(doc.type)}</span>
                  <span className="doc-type">{getDocumentTypeLabel(doc.type)}</span>
                </div>
                <div className="table-cell table-cell--number">{doc.number}</div>
                <div className="table-cell table-cell--title">{doc.title}</div>
                <div className="table-cell table-cell--recipient">{doc.recipient}</div>
                <div className="table-cell table-cell--status">
                  <StatusBadge status={doc.status as any} />
                </div>
                <div className="table-cell table-cell--signature">
                  <span className={`signature-badge signature-badge--${doc.signature_status}`}>
                    {doc.signature_status === 'pending' && '⏳ Ausstehend'}
                    {doc.signature_status === 'signed' && '✓ Unterzeichnet'}
                    {doc.signature_status === 'rejected' && '✕ Abgelehnt'}
                  </span>
                </div>
                <div className="table-cell table-cell--date">
                  {new Date(doc.created_date).toLocaleDateString('de-DE')}
                </div>
                <div className="table-cell table-cell--action">
                  <button
                    className="btn btn--sm btn--primary"
                    onClick={(e) => {
                      e.stopPropagation()
                      navigate(`/documents/${doc.id}`)
                    }}
                  >
                    Ansehen
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export default DocumentsPage
