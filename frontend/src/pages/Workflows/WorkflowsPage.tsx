import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { WorkflowDefinition, WorkflowTemplate } from '../../types/workflow'
import styles from './Workflows.module.scss'

const mockWorkflowDefinitions: WorkflowDefinition[] = [
  {
    id: '1',
    name: 'Onboarding neuer Mitarbeiter',
    description: 'Automatischer Onboarding-Prozess mit Qualifikationsprüfung',
    status: 'active',
    trigger: 'manual',
    steps: [
      { id: 's1', type: 'trigger', name: 'Mitarbeiter hinzufügen', description: '', config: {} },
      { id: 's2', type: 'action', name: 'Willkommens-Email', description: '', config: {} },
      { id: 's3', type: 'action', name: 'Qualifikationen abfragen', description: '', config: {} },
      { id: 's4', type: 'notification', name: 'Team benachrichtigen', description: '', config: {} },
    ],
    created_at: '2026-01-15T10:00:00Z',
    updated_at: '2026-03-15T14:30:00Z',
    created_by: 'admin@rentflow.de',
    instances_count: 24,
  },
  {
    id: '2',
    name: 'Wartungserinnerungen',
    description: 'Regelmäßige Erinnerungen für anstehende Wartungen',
    status: 'active',
    trigger: 'scheduled',
    steps: [
      { id: 's5', type: 'trigger', name: 'Wöchentlicher Check', description: '', config: {} },
      { id: 's6', type: 'condition', name: 'Wartung überdue?', description: '', config: {} },
      { id: 's7', type: 'action', name: 'Email an Techniker', description: '', config: {} },
      { id: 's8', type: 'action', name: 'Task erstellen', description: '', config: {} },
    ],
    created_at: '2026-02-01T08:00:00Z',
    updated_at: '2026-03-20T11:00:00Z',
    created_by: 'service@rentflow.de',
    instances_count: 156,
  },
  {
    id: '3',
    name: 'Ausrüstungswarnung',
    description: 'Benachrichtigung bei niedriger Verfügbarkeit',
    status: 'paused',
    trigger: 'event',
    steps: [
      { id: 's9', type: 'trigger', name: 'Bestand-Event', description: '', config: {} },
      { id: 's10', type: 'condition', name: 'Minimum erreicht?', description: '', config: {} },
      { id: 's11', type: 'notification', name: 'Manager benachrichtigen', description: '', config: {} },
      { id: 's12', type: 'action', name: 'Nachbestellung vorschlagen', description: '', config: {} },
    ],
    created_at: '2026-02-10T09:00:00Z',
    updated_at: '2026-03-18T16:45:00Z',
    created_by: 'inventory@rentflow.de',
    instances_count: 42,
  },
]

const mockWorkflowTemplates: WorkflowTemplate[] = [
  {
    id: 't1',
    name: 'Onboarding Prozess',
    description: 'Standard Mitarbeiter Onboarding',
    category: 'HR',
    icon: '👥',
    steps: [
      { id: 'st1', type: 'trigger', name: 'Start', description: '', config: {} },
      { id: 'st2', type: 'action', name: 'Welcome Email', description: '', config: {} },
      { id: 'st3', type: 'decision', name: 'Qualifications?', description: '', config: {} },
    ],
  },
  {
    id: 't2',
    name: 'Wartungs-Reminder',
    description: 'Regelmäßige Wartungsprüfungen',
    category: 'Maintenance',
    icon: '🔧',
    steps: [
      { id: 'st4', type: 'trigger', name: 'Zeitplan', description: '', config: {} },
      { id: 'st5', type: 'condition', name: 'Prüfen', description: '', config: {} },
      { id: 'st6', type: 'action', name: 'Aufgabe erstellen', description: '', config: {} },
    ],
  },
  {
    id: 't3',
    name: 'Bestandswarnung',
    description: 'Alarmierung bei niedrigen Beständen',
    category: 'Inventory',
    icon: '📦',
    steps: [
      { id: 'st7', type: 'trigger', name: 'Bestandsänderung', description: '', config: {} },
      { id: 'st8', type: 'condition', name: 'Mindestmenge erreicht?', description: '', config: {} },
      { id: 'st9', type: 'notification', name: 'Benachrichtigung', description: '', config: {} },
    ],
  },
]

function WorkflowsPage() {
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null)

  const { data: workflows = mockWorkflowDefinitions } = useQuery({
    queryKey: ['workflows'],
    queryFn: async () => mockWorkflowDefinitions,
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
          <div className={styles.statValue}>{workflows.length}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Aktiv</div>
          <div className={styles.statValue} style={{ color: 'var(--color-success)' }}>
            {workflows.filter(w => w.status === 'active').length}
          </div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Instanzen läuft</div>
          <div className={styles.statValue}>{workflows.reduce((sum, w) => sum + w.instances_count, 0)}</div>
        </div>
      </div>

      <div className={styles.mainGrid}>
        {/* Workflows List */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Aktive Workflow-Definitionen</h2>

          {workflows.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>⚡</div>
              <h3 className={styles.emptyTitle}>Keine Workflows vorhanden</h3>
              <p className={styles.emptyText}>Erstellen Sie einen neuen Workflow, um zu beginnen.</p>
            </div>
          ) : (
            <div className={styles.workflowsList}>
              {workflows.map(workflow => (
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

          <div className={styles.templatesGrid}>
            {mockWorkflowTemplates.map(template => (
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
        </section>
      </div>
    </div>
  )
}

export default WorkflowsPage
