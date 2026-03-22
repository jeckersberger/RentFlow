import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { workflowApi } from '../../services/api'
import type { WorkflowDefinition, WorkflowTemplate } from '../../types/workflow'
import styles from './Workflows.module.scss'

function WorkflowsPage() {
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null)

  // Fetch workflow definitions
  const { data: workflows = [], isLoading: isLoadingWorkflows } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => workflowApi.list(),
    staleTime: 1000 * 60 * 5,
  })

  // Fetch workflow templates
  const { data: templates = [], isLoading: isLoadingTemplates } = useQuery({
    queryKey: ['workflow-templates'],
    queryFn: () => workflowApi.templates(),
    staleTime: 1000 * 60 * 5,
  })

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'active':
        return '#10b981'
      case 'paused':
        return '#f59e0b'
      case 'archived':
        return '#6b7280'
      default:
        return '#00d4ff'
    }
  }

  const getStatusLabel = (status: string): string => {
    switch (status) {
      case 'active':
        return 'Aktiv'
      case 'paused':
        return 'Pausiert'
      case 'archived':
        return 'Archiviert'
      default:
        return status
    }
  }

  return (
    <div className={styles.workflowsPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>⚡ Workflow-Editor</h1>
          <p className={styles.subtitle}>Erstellen und verwalten Sie automatisierte Workflows</p>
        </div>
        <button className="btn btn--primary" style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}>
          ➕ Neuer Workflow
        </button>
      </div>

      {/* Stats */}
      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Gesamt Workflows</div>
          <div className={styles.statValue}>{isLoadingWorkflows ? '⏳' : workflows.length}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Aktiv</div>
          <div className={styles.statValue} style={{ color: 'var(--color-success)' }}>
            {isLoadingWorkflows ? '⏳' : workflows.filter((w: WorkflowDefinition) => w.status === 'active').length}
          </div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Instanzen läuft</div>
          <div className={styles.statValue}>{isLoadingWorkflows ? '⏳' : workflows.reduce((sum: number, w: WorkflowDefinition) => sum + w.instances_count, 0)}</div>
        </div>
      </div>

      <div className={styles.mainGrid}>
        {/* Workflows List */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Aktive Workflow-Definitionen</h2>

          {isLoadingWorkflows ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>⏳</div>
              <p className={styles.emptyText}>Laden...</p>
            </div>
          ) : workflows.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>⚡</div>
              <h3 className={styles.emptyTitle}>Keine Workflows vorhanden</h3>
              <p className={styles.emptyText}>Erstellen Sie einen neuen Workflow, um zu beginnen.</p>
            </div>
          ) : (
            <div className={styles.workflowsList}>
              {workflows.map((workflow: WorkflowDefinition) => (
                <div key={workflow.id} className={styles.workflowCard}>
                  <div className={styles.workflowHeader}>
                    <div>
                      <h3 className={styles.workflowName}>{workflow.name}</h3>
                      <p className={styles.workflowDesc}>{workflow.description}</p>
                    </div>
                    <div
                      className={styles.statusBadge}
                      style={{ backgroundColor: `${getStatusColor(workflow.status)}20`, borderColor: getStatusColor(workflow.status) }}
                    >
                      <span style={{ color: getStatusColor(workflow.status) }}>●</span>
                      {getStatusLabel(workflow.status)}
                    </div>
                  </div>

                  {/* Workflow Steps Visualization */}
                  <div className={styles.stepsContainer}>
                    {workflow.steps.map((step, idx) => (
                      <div key={step.id}>
                        <div className={styles.stepCard}>
                          <div className={styles.stepIcon}>
                            {step.type === 'trigger' && '▶'}
                            {step.type === 'action' && '⚙'}
                            {step.type === 'condition' && '◆'}
                            {step.type === 'notification' && '🔔'}
                            {step.type === 'decision' && '◇'}
                          </div>
                          <div className={styles.stepLabel}>{step.name}</div>
                        </div>
                        {idx < workflow.steps.length - 1 && <div className={styles.stepArrow}>↓</div>}
                      </div>
                    ))}
                  </div>

                  <div className={styles.workflowMeta}>
                    <span className={styles.metaItem}>
                      <strong>Trigger:</strong> {workflow.trigger}
                    </span>
                    <span className={styles.metaItem}>
                      <strong>Instanzen:</strong> {workflow.instances_count}
                    </span>
                    <span className={styles.metaItem}>
                      <strong>Erstellt von:</strong> {workflow.created_by}
                    </span>
                  </div>

                  <div className={styles.workflowActions}>
                    <button className="btn btn--sm btn--primary">Bearbeiten</button>
                    <button className="btn btn--sm">Details</button>
                    <button className="btn btn--sm">Duplikate</button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Templates Gallery */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Template-Galerie</h2>

          {isLoadingTemplates ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>⏳</div>
              <p className={styles.emptyText}>Laden...</p>
            </div>
          ) : templates.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>📦</div>
              <p className={styles.emptyText}>Keine Templates verfügbar</p>
            </div>
          ) : (
            <div className={styles.templatesGrid}>
              {templates.map((template: WorkflowTemplate) => (
                <div
                  key={template.id}
                  className={`${styles.templateCard} ${selectedTemplate === template.id ? styles['templateCard--selected'] : ''}`}
                  onClick={() => setSelectedTemplate(selectedTemplate === template.id ? null : template.id)}
                >
                  <div className={styles.templateIcon}>{template.icon}</div>
                  <h3 className={styles.templateName}>{template.name}</h3>
                  <p className={styles.templateDesc}>{template.description}</p>
                  <span className={styles.templateCategory}>{template.category}</span>
                  {selectedTemplate === template.id && (
                    <button className="btn btn--sm btn--primary" style={{ marginTop: 'var(--spacing-3)', width: '100%' }}>
                      Verwenden
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}

export default WorkflowsPage
