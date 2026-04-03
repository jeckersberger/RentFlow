import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { projectApi } from '../../services/api'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { OverviewTab } from './tabs/OverviewTab'
import { EquipmentTab } from './tabs/EquipmentTab'
import { CrewTab } from './tabs/CrewTab'
import { FinanceTab } from './tabs/FinanceTab'
import { TransportTab } from './tabs/TransportTab'
import { DocumentsTab } from './tabs/DocumentsTab'
import { CommunicationTab } from './tabs/CommunicationTab'
import { HistoryTab } from './tabs/HistoryTab'
import './ProjectDetail.scss'

type TabId = 'overview' | 'equipment' | 'crew' | 'finance' | 'transport' | 'documents' | 'communication' | 'history'

interface TabDef {
  id: TabId
  label: string
  icon: string
}

const TABS: TabDef[] = [
  { id: 'overview', label: 'Übersicht', icon: '\u{1F4CB}' },
  { id: 'equipment', label: 'Equipment', icon: '\u{1F3A4}' },
  { id: 'crew', label: 'Crew', icon: '\u{1F465}' },
  { id: 'finance', label: 'Finanzen', icon: '\u{1F4B0}' },
  { id: 'transport', label: 'Transport', icon: '\u{1F69A}' },
  { id: 'documents', label: 'Dokumente', icon: '\u{1F4C4}' },
  { id: 'communication', label: 'Kommunikation', icon: '\u{1F4AC}' },
  { id: 'history', label: 'Verlauf', icon: '\u{1F553}' },
]

// Status-based color indicator
function getStatusColor(status: string): string {
  switch (status) {
    case 'confirmed': return '#10b981'
    case 'in_progress': return '#f59e0b'
    case 'completed': return '#06b6d4'
    case 'cancelled': return '#ef4444'
    case 'invoiced': return '#8b5cf6'
    case 'quoted': return '#3b82f6'
    default: return '#6b7280'
  }
}

function ProjectDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [activeTab, setActiveTab] = useState<TabId>('overview')

  const { data: projectRaw, isLoading, error } = useQuery({
    queryKey: ['project', id],
    queryFn: () => projectApi.getById(id!),
    enabled: !!id,
  })

  // Normalize backend field names
  const project = projectRaw ? {
    ...projectRaw,
    name: projectRaw.name || projectRaw.title || '---',
    status: projectRaw.status || 'draft',
    start_date: projectRaw.start_date || projectRaw.startDate || '',
    end_date: projectRaw.end_date || projectRaw.endDate || '',
    client_name: projectRaw.client_name || projectRaw.contact_name || '',
  } : null

  if (isLoading) {
    return <div className="loading">Wird geladen...</div>
  }

  if (error) {
    return <div className="error">Fehler beim Laden des Projekts</div>
  }

  if (!project) {
    return <div className="error">Projekt nicht gefunden</div>
  }

  const formatDateShort = (date: string) => {
    if (!date) return '---'
    const d = new Date(date)
    return isNaN(d.getTime()) ? '---' : d.toLocaleDateString('de-DE')
  }

  const renderActiveTab = () => {
    switch (activeTab) {
      case 'overview': return <OverviewTab project={project} />
      case 'equipment': return <EquipmentTab project={project} />
      case 'crew': return <CrewTab project={project} />
      case 'finance': return <FinanceTab project={project} />
      case 'transport': return <TransportTab project={project} />
      case 'documents': return <DocumentsTab project={project} />
      case 'communication': return <CommunicationTab project={project} />
      case 'history': return <HistoryTab project={project} />
      default: return <OverviewTab project={project} />
    }
  }

  return (
    <div className="projectDetail">
      {/* Project Header */}
      <div className="header">
        <div className="headerLeft">
          <div
            className="colorIndicator"
            style={{ backgroundColor: getStatusColor(project.status) }}
          />
          <div className="headerInfo">
            <p className="projectNumber">
              Projekt #{project.id?.slice(0, 8).toUpperCase() || '---'}
            </p>
            <h1 className="projectName">{project.name}</h1>
            <div className="headerMeta">
              <StatusBadge status={project.status} />
              <span className="headerDates">
                {formatDateShort(project.start_date)} &ndash; {formatDateShort(project.end_date)}
              </span>
              {project.client_name && (
                <span className="headerDates">
                  | {project.client_name}
                </span>
              )}
            </div>
          </div>
        </div>
        <div className="headerActions">
          <button
            className="btn btn--secondary"
            onClick={() => navigate('/projects')}
          >
            Zurück
          </button>
          <button
            className="btn btn--primary"
            onClick={() => navigate(`/projects/${id}/edit`)}
          >
            Bearbeiten
          </button>
        </div>
      </div>

      {/* Tab Bar */}
      <div className="tabBar">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            className={`tab ${activeTab === tab.id ? 'tabActive' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            <span className="tabIcon">{tab.icon}</span>
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="tabContent">
        {renderActiveTab()}
      </div>
    </div>
  )
}

export default ProjectDetailPage
