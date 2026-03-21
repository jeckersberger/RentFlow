import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { projectApi } from '../../services/api'
import { DataTable, Column } from '../../components/DataTable/DataTable'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Input } from '../../components/Form/Input'
import { Project, ProjectStatus } from '../../types/project'
import '../Equipment/Equipment.module.scss'

const STATUS_TABS: Array<{ value: ProjectStatus | ''; label: string }> = [
  { value: '', label: 'Alle' },
  { value: 'draft', label: 'Entwurf' },
  { value: 'confirmed', label: 'Bestätigt' },
  { value: 'in_progress', label: 'In Bearbeitung' },
  { value: 'completed', label: 'Abgeschlossen' },
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

  const _filteredData = (projectData?.data || []).filter((project: Project) => {
    const matchesSearch =
      !searchQuery ||
      project.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      project.client_name?.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesStatus = !selectedStatus || project.status === selectedStatus

    return matchesSearch && matchesStatus
  })

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
      render: (budget: any) => (budget ? `€${budget.toFixed(2)}` : '—'),
    },
  ]

  return (
    <div className="project-list-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Projekte</h1>
          <p className="page-subtitle">Verwalten Sie alle Ihre Veranstaltungsprojekte</p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => navigate('/projects/new')}
        >
          + Neues Projekt
        </button>
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

      {error && (
        <div className="error-message" role="alert">
          Fehler beim Laden der Projekte. Bitte versuchen Sie es später erneut.
        </div>
      )}

      {viewMode === 'list' ? (
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
          {_filteredData.length === 0 ? (
            <p style={{ gridColumn: '1 / -1', textAlign: 'center', color: 'var(--color-text-secondary)' }}>
              Keine Projekte gefunden
            </p>
          ) : (
            _filteredData.map((project: Project) => (
              <div
                key={project.id}
                style={{
                  padding: 'var(--spacing-4)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius-card)',
                  backgroundColor: 'var(--color-bg-primary)',
                  cursor: 'pointer',
                  transition: 'all var(--transition-fast)',
                }}
                onClick={() => navigate(`/projects/${project.id}`)}
                onMouseEnter={(e) => (e.currentTarget.style.boxShadow = 'var(--shadow-lg)')}
                onMouseLeave={(e) => (e.currentTarget.style.boxShadow = 'var(--shadow-card)')}
              >
                <h3 style={{ margin: '0 0 var(--spacing-2) 0', color: 'var(--color-text-primary)' }}>
                  {project.name}
                </h3>
                <p style={{ margin: '0 0 var(--spacing-2) 0', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                  {project.client_name}
                </p>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-3)' }}>
                  <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                    {new Date(project.start_date).toLocaleDateString('de-DE')}
                  </span>
                  <StatusBadge status={project.status as ProjectStatus} />
                </div>
                {project.budget && (
                  <p style={{ margin: 0, fontSize: 'var(--font-size-sm)', fontWeight: 'var(--font-weight-semibold)', color: 'var(--color-primary)' }}>
                    €{project.budget.toFixed(2)}
                  </p>
                )}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  )
}

export default ProjectListPage
