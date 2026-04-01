import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { documentApi } from '../../services/api'
import '../Equipment/Equipment.scss'

function DocumentDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()

  const {
    data: document,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ['document', id],
    queryFn: () => documentApi.getById(id!),
    enabled: !!id,
  })

  if (isLoading) {
    return (
      <div className="equipment-detail-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Wird geladen...</h1>
          </div>
        </div>
        <div className="detail-card">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              style={{
                height: '1.5rem',
                background: 'var(--color-bg-tertiary)',
                borderRadius: 'var(--radius-md)',
                marginBottom: 'var(--spacing-4)',
              }}
            />
          ))}
        </div>
      </div>
    )
  }

  if (isError || !document) {
    return (
      <div className="equipment-detail-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Dokument nicht gefunden</h1>
            <p className="page-subtitle">
              Das angeforderte Dokument konnte nicht geladen werden.
            </p>
          </div>
        </div>
        <button
          className="btn btn--secondary"
          onClick={() => navigate('/documents')}
        >
          Zurueck zu Dokumenten
        </button>
      </div>
    )
  }

  const formatDate = (dateStr: string | undefined) => {
    if (!dateStr) return '-'
    try {
      return new Date(dateStr).toLocaleDateString('de-DE', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    } catch {
      return dateStr
    }
  }

  return (
    <div className="equipment-detail-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">
            {document.title || document.name || 'Dokument'}
          </h1>
          <p className="page-subtitle">
            Dokument-ID: {document.id}
          </p>
        </div>
      </div>

      <div className="detail-grid">
        <div className="detail-card">
          <h3 className="detail-card__title">Dokumentinformationen</h3>
          <div className="detail-card__content">
            <div className="detail-card__row">
              <span className="detail-card__row-label">Name</span>
              <span className="detail-card__row-value">
                {document.title || document.name || '-'}
              </span>
            </div>
            <div className="detail-card__row">
              <span className="detail-card__row-label">Typ</span>
              <span className="detail-card__row-value">
                {document.document_type || '-'}
              </span>
            </div>
            <div className="detail-card__row">
              <span className="detail-card__row-label">Vorlage</span>
              <span className="detail-card__row-value">
                {document.template_id || '-'}
              </span>
            </div>
            <div className="detail-card__row">
              <span className="detail-card__row-label">Dokumentnr.</span>
              <span className="detail-card__row-value">
                {document.document_number || '-'}
              </span>
            </div>
            <div className="detail-card__row">
              <span className="detail-card__row-label">Erstellt am</span>
              <span className="detail-card__row-value">
                {formatDate(document.created_at)}
              </span>
            </div>
            <div className="detail-card__row">
              <span className="detail-card__row-label">Status</span>
              <span className="detail-card__row-value">
                {document.status || '-'}
              </span>
            </div>
          </div>

          <div className="detail-actions">
            <button
              className="btn btn--secondary"
              onClick={() => navigate('/documents')}
            >
              Zurueck
            </button>
            {document.download_url && (
              <a
                href={document.download_url}
                target="_blank"
                rel="noopener noreferrer"
                className="btn btn--primary"
              >
                Herunterladen
              </a>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default DocumentDetailPage
