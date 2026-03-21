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
  const limit = 20

  const { data: projectData, isLoading, error } = useQuery({
    queryKey: ['project-list', page, searchQuery, selectedStatus],
    queryFn: () => projectApi.list(page, limit),
    staleTime: 1000 * 60 * 5,
  })

  const filteredData = (projectData?.data || []).filter((project: Project) => {
    const matchesSearch =
      !searchQuery ||
      project.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      project.client_name?.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesStatus = !selectedStatus || project.status === selectedStatus

    return matchesSearch && matchesStatus
  })

  const columns: Column<Project>[] = [
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
      render: (status: ProjectStatus) => (
        <StatusBadge status={status} />
      ),
    },
    {
      key: 'start_date',
      label: 'Startdatum',
      render: (date: string) =>
        new Date(date).toLocaleDateString('de-DE'),
    },
    {
      key: 'budget',
      label: 'Budget',
      render: (budget?: number) => (budget ? `€${budget.toFixed(2)}` : '—'),
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
        <div style={{ display: 'flex', gap: 'var(--spacing-2)', marginBottom: 'var(--spacing-4)', flexWrap: 'wrap' }}>
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

      <DataTable<Project>
        columns={columns}
        data={filteredData}
        rowKey="id"
        loading={isLoading}
        onRowClick={(project) => navigate(`/projects/${project.id}`)}
        pagination={{
          page,
          total: projectData?.total || 0,
          limit,
          onPageChange: setPage,
        }}
      />
    </div>
  )
}

export default ProjectListPage
