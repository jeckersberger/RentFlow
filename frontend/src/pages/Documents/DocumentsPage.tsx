import { useState } from 'react'
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
    signature_status: 'pending' as Document['signature_status'],
    created_date: dto.created_at ? new Date(dto.created_at).toISOString().split('T')[0] : '',
    recipient: dto.metadata?.recipient || '',
    project_id: dto.reference_id || undefined,
    generated_by: dto.created_by || '',
  }
}

function DocumentsPage() {
  const navigate = useNavigate()
  const [filterType, setFilterType] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')

  const { data: documents = [], isLoading, error } = useQuery({
    queryKey: ['documents'],
    queryFn: async () => {
      const result = await documentApi.list()
      const items = Array.isArray(result) ? result : (result.data || [])
      return items.map(mapBackendDocument)
    },
    staleTime: 1000 * 60 * 5,
  })

  const pendingSignatures = documents.filter((d: any) => d.signature_status === 'pending')
  const thisMonthGenerated = documents.filter((d: any) => {
    const docDate = new Date(d.created_date)
    const now = new Date()
    return docDate.getMonth() === now.getMonth() && docDate.getFullYear() === now.getFullYear()
  })

  const filteredDocuments = documents.filter((doc: any) => {
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
            {documents.filter((d: any) => d.signature_status === 'signed').length}
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
            {filteredDocuments.map((doc: any) => (
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
