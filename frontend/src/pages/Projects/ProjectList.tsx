import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { FolderOpen } from 'lucide-react'
import { projectApi } from '../../services/api'
import { DataTable, Column } from '../../components/DataTable/DataTable'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Input } from '../../components/Form/Input'
import EmptyState from '../../components/EmptyState/EmptyState'
import ErrorState from '../../components/ErrorState/ErrorState'
import { SkeletonTable, SkeletonCard } from '../../components/Skeleton/SkeletonLoader'
import { Project, ProjectStatus } from '../../types/project'
import { generateCSV, downloadCSV, formatDateForExport } from '../../utils/csvExport'
import '../Equipment/Equipment.scss'

const STATUS_TABS: Array<{ value: ProjectStatus | ''; label: string }> = [
  { value: '', label: 'Alle' },
  { value: 'draft', label: 'Entwurf' },
  { value: 'confirmed', label: 'Bestätigt' },
  { value: 'in_progress', label: 'In Bearbeitung' },
  { value: 'completed', label: 'Abgeschlossen' },
  { value: 'cancelled', label: 'Storniert' },
]

function ProjectListPage() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState<ProjectStatus | ''>('')
  const [viewMode, setViewMode] = useState<'list' | 'calendar'>('list')
  const limit = 20

  const { data: projectData, isLoading: _isLoading, error } = useQuery({
    queryKey: ['project-list', page, searchQuery, selectedStatus],
    queryFn: () => projectApi.list(page, limit),
    staleTime: 1000 * 60 * 5,
  })

  // Handle both response formats: { items: [...] } (mock) and { data: [...] } (real API)
  const _filteredData = (projectData?.items || projectData?.data || []).filter((project: Project) => {
    const matchesSearch =
      !searchQuery ||
      project.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      project.client_name?.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesStatus = !selectedStatus || project.status === selectedStatus

    return matchesSearch && matchesStatus
  })

  const exportProjectsCSV = () => {
    const PROJECT_HEADERS = [
      { key: 'name', label: 'Projektname' },
      { key: 'client_name', label: 'Kunde' },
      { key: 'status', label: 'Status' },
      { key: 'start_date', label: 'Startdatum' },
      { key: 'end_date', label: 'Enddatum' },
      { key: 'budget', label: 'Budget' },
      { key: 'location', label: 'Ort' },
      { key: 'description', label: 'Beschreibung' },
    ]
    const STATUS_LABELS: Record<string, string> = {
      draft: 'Entwurf', quoted: 'Angebot', confirmed: 'Bestätigt',
      in_progress: 'In Bearbeitung', completed: 'Abgeschlossen',
      cancelled: 'Storniert', invoiced: 'Berechnet',
    }
    const rows = _filteredData.map((p: Project) => ({
      name: p.name || '',
      client_name: p.client_name || '',
      status: STATUS_LABELS[p.status] || p.status || '',
      start_date: p.start_date ? new Date(p.start_date).toLocaleDateString('de-DE') : '',
      end_date: p.end_date ? new Date(p.end_date).toLocaleDateString('de-DE') : '',
      budget: p.budget != null ? String(p.budget) : '',
      location: p.location || '',
      description: p.description || '',
    }))
    const csv = generateCSV(PROJECT_HEADERS, rows)
    downloadCSV(csv, `projekte_export_${formatDateForExport()}.csv`)
  }

  const _columns: Column<Project>[] = [
    {
      key: 'name',
      label: 'Projektname',
      sortable: true,
    },
    {
      key: 'client_name',
      label: 'Kunde',
    },
    {
      key: 'status',
      label: 'Status',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      render: (status: any) => (
        <StatusBadge status={status as ProjectStatus} />
      ),
    },
    {
      key: 'start_date',
      label: 'Startdatum',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      render: (date: any) =>
        new Date(date).toLocaleDateString('de-DE'),
    },
    {
      key: 'budget',
      label: 'Budget',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      render: (budget: any) => (budget ? budget.toLocaleString('de-DE', { style: 'currency', currency: 'EUR' }) : '—'),
    },
  ]

  return (
    <div className="project-list-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Projekte</h1>
          <p className="page-subtitle">Verwalten Sie alle Ihre Veranstaltungsprojekte</p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          <button
            className="btn btn--secondary"
            onClick={exportProjectsCSV}
            disabled={_filteredData.length === 0}
          >
            Exportieren
          </button>
          <button
            className="btn btn--primary"
            onClick={() => navigate('/projects/new')}
          >
            + Neues Projekt
          </button>
        </div>
      </div>

      <div style={{ marginBottom: 'var(--spacing-6)' }}>
        <div style={{ display: 'flex', gap: 'var(--spacing-2)', marginBottom: 'var(--spacing-4)', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', gap: 'var(--spacing-2)', flexWrap: 'wrap' }}>
            {STATUS_TABS.map((tab) => (
              <button
                key={tab.value}
                className={`btn ${selectedStatus === tab.value ? 'btn--primary' : 'btn--secondary'}`}
                onClick={() => {
                  setSelectedStatus(tab.value)
                  setPage(1)
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <div style={{ display: 'flex', gap: 'var(--spacing-2)' }}>
            <button
              className={`btn ${viewMode === 'list' ? 'btn--primary' : 'btn--secondary'}`}
              onClick={() => setViewMode('list')}
            >
              📝 Liste
            </button>
            <button
              className={`btn ${viewMode === 'calendar' ? 'btn--primary' : 'btn--secondary'}`}
              onClick={() => setViewMode('calendar')}
            >
              📅 Kalender
            </button>
          </div>
        </div>

        <Input
          type="text"
          placeholder="Nach Projektname oder Kunde suchen..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value)
            setPage(1)
          }}
        />
      </div>

      {error ? (
        <ErrorState
          variant="generic"
          title="Fehler beim Laden der Projekte"
          description="Die Projektdaten konnten nicht geladen werden. Bitte versuchen Sie es erneut."
          onRetry={() => window.location.reload()}
          compact
        />
      ) : _isLoading ? (
        viewMode === 'list' ? <SkeletonTable rows={6} columns={5} /> : <SkeletonCard count={6} />
      ) : _filteredData.length === 0 ? (
        <EmptyState
          icon={FolderOpen}
          title={selectedStatus ? 'Keine Projekte mit diesem Status' : 'Noch keine Projekte'}
          description={selectedStatus ? 'Versuchen Sie einen anderen Filter.' : 'Erstellen Sie Ihr erstes Projekt, um loszulegen.'}
          action={selectedStatus ? undefined : { label: 'Neues Projekt', href: '/projects/new' }}
        />
      ) : viewMode === 'list' ? (
        <DataTable<Project>
          columns={_columns}
          data={_filteredData}
          rowKey="id"
          loading={_isLoading}
          onRowClick={(project) => navigate(`/projects/${project.id}`)}
          pagination={{
            page,
            total: projectData?.total || 0,
            limit,
            onPageChange: setPage,
          }}
        />
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: 'var(--spacing-4)' }}>
          {_filteredData.map((project: Project) => (
              <div
                key={project.id}
                style={{
                  padding: 'var(--spacing-4)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius-card)',
                  backgroundColor: 'var(--color-bg-secondary)',
                  backdropFilter: 'blur(12px)',
                  cursor: 'pointer',
                  transition: 'all 0.2s ease',
                  boxShadow: 'var(--shadow-card)',
                }}
                onClick={() => navigate(`/projects/${project.id}`)}
                onMouseEnter={(e) => {
                  e.currentTarget.style.boxShadow = '0 0 20px rgba(0, 212, 255, 0.15)'
                  e.currentTarget.style.borderColor = 'var(--color-primary)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.boxShadow = 'var(--shadow-card)'
                  e.currentTarget.style.borderColor = 'var(--color-border)'
                }}
              >
                <h3 style={{ margin: '0 0 var(--spacing-2) 0', color: 'var(--color-text-primary)' }}>
                  {project.name}
                </h3>
                <p style={{ margin: '0 0 var(--spacing-2) 0', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                  {project.client_name || 'Kein Kunde'}
                </p>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-3)' }}>
                  <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                    {project.start_date ? new Date(project.start_date).toLocaleDateString('de-DE') : '---'}
                    {project.end_date && project.end_date !== project.start_date && (
                      <> — {new Date(project.end_date).toLocaleDateString('de-DE')}</>
                    )}
                  </span>
                  <StatusBadge status={project.status as ProjectStatus} />
                </div>
                {project.budget != null && project.budget > 0 && (
                  <p style={{ margin: 0, fontSize: 'var(--font-size-sm)', fontWeight: 'var(--font-weight-semibold)', color: 'var(--color-primary)' }}>
                    {project.budget.toLocaleString('de-DE', { style: 'currency', currency: 'EUR' })}
                  </p>
                )}
              </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default ProjectListPage
