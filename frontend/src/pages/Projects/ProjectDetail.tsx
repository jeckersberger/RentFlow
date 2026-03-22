import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { projectApi } from '../../services/api'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import '../Equipment/Equipment.module.scss'

function ProjectDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({
    info: true,
    packlist: true,
    crew: false,
  })

  const { data: project, isLoading, error } = useQuery({
    queryKey: ['project', id],
    queryFn: () => projectApi.getById(id!),
    enabled: !!id,
  })

  const toggleSection = (section: string) => {
    setExpandedSections((prev) => ({
      ...prev,
      [section]: !prev[section],
    }))
  }

  if (isLoading) return <div className="project-detail-page">Wird geladen...</div>
  if (error) return <div className="error-message">Fehler beim Laden des Projekts</div>
  if (!project) return <div className="error-message">Projekt nicht gefunden</div>

  return (
    <div className="project-detail-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{project.name}</h1>
          <p className="page-subtitle">
            <StatusBadge status={project.status} />
          </p>
        </div>
        <button
          className="btn btn--secondary"
          onClick={() => navigate(`/projects/${id}/edit`)}
        >
          Bearbeiten
        </button>
      </div>

      <div className="detail-card">
        <div
          className="detail-card__title"
          style={{ cursor: 'pointer' }}
          onClick={() => toggleSection('info')}
        >
          📋 Projektinformationen
        </div>

        {expandedSections.info && (
          <div className="detail-card__content">
            <div className="detail-card__row">
              <span className="detail-card__row-label">Kunde</span>
              <span className="detail-card__row-value">
                {project.client_name || '—'}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Startdatum</span>
              <span className="detail-card__row-value">
                {new Date(project.start_date).toLocaleDateString('de-DE')}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Enddatum</span>
              <span className="detail-card__row-value">
                {new Date(project.end_date).toLocaleDateString('de-DE')}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Standort</span>
              <span className="detail-card__row-value">
                {project.location
                  || (project.venue_address && [project.venue_address.street, project.venue_address.city, project.venue_address.country].filter(Boolean).join(', '))
                  || '—'}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Budget</span>
              <span className="detail-card__row-value">
                {project.budget ? `€${project.budget.toFixed(2)}` : '—'}
              </span>
            </div>

            {project.description && (
              <div className="detail-card__row">
                <span className="detail-card__row-label">Beschreibung</span>
                <span className="detail-card__row-value">
                  {project.description}
                </span>
              </div>
            )}
          </div>
        )}
      </div>

      <div className="detail-card" style={{ marginTop: 'var(--spacing-4)' }}>
        <div
          className="detail-card__title"
          style={{ cursor: 'pointer' }}
          onClick={() => toggleSection('packlist')}
        >
          📦 Packliste
        </div>

        {expandedSections.packlist && (
          <div className="detail-card__content">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
              <p style={{ color: 'var(--color-text-secondary)', margin: 0 }}>
                Packliste wird in der erweiterten Version angezeigt
              </p>
            </div>
          </div>
        )}
      </div>

      <div className="detail-card" style={{ marginTop: 'var(--spacing-4)' }}>
        <div
          className="detail-card__title"
          style={{ cursor: 'pointer' }}
          onClick={() => toggleSection('crew')}
        >
          👥 Crew & Team
        </div>

        {expandedSections.crew && (
          <div className="detail-card__content">
            <p style={{ color: 'var(--color-text-secondary)', margin: 0 }}>
              Crew-Zuordnung wird in der erweiterten Version angezeigt
            </p>
          </div>
        )}
      </div>
    </div>
  )
}

export default ProjectDetailPage
