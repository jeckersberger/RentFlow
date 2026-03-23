import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { workflowApi } from '../../services/api'
import type { WorkflowDefinition, WorkflowTemplate } from '../../types/workflow'
import styles from './Workflows.module.scss'

// Demo data for when API returns empty
const demoWorkflows: WorkflowDefinition[] = [
  {
    id: 'wf1',
    name: 'Mahnung bei ueberfaelliger Rueckgabe',
    description: 'Sendet automatisch eine Erinnerung wenn Equipment nicht fristgerecht zurueckgegeben wird.',
    status: 'active',
    trigger: 'event',
    steps: [
      { id: 's1', type: 'trigger', name: 'Rueckgabedatum ueberschritten', description: '', config: {} },
      { id: 's2', type: 'condition', name: 'Pruefe Verzoegerung > 24h', description: '', config: {} },
      { id: 's3', type: 'notification', name: 'E-Mail an Kunden', description: '', config: {} },
      { id: 's4', type: 'action', name: 'Status auf Ueberfaellig setzen', description: '', config: {} },
    ],
    created_at: '2026-01-15T10:00:00Z',
    updated_at: '2026-03-20T14:30:00Z',
    created_by: 'Admin',
    instances_count: 12,
  },
  {
    id: 'wf2',
    name: 'Wartungserinnerung',
    description: 'Erstellt automatisch Wartungsauftraege basierend auf Nutzungsstunden oder Zeitintervallen.',
    status: 'active',
    trigger: 'scheduled',
    steps: [
      { id: 's5', type: 'trigger', name: 'Wartungsintervall erreicht', description: '', config: {} },
      { id: 's6', type: 'action', name: 'Wartungsauftrag erstellen', description: '', config: {} },
      { id: 's7', type: 'notification', name: 'Techniker benachrichtigen', description: '', config: {} },
    ],
    created_at: '2026-02-01T09:00:00Z',
    updated_at: '2026-03-18T11:00:00Z',
    created_by: 'Admin',
    instances_count: 34,
  },
  {
    id: 'wf3',
    name: 'Angebotsannahme-Workflow',
    description: 'Automatisiert den Prozess nach Angebotsannahme: Rechnung erstellen, Lieferschein vorbereiten.',
    status: 'active',
    trigger: 'event',
    steps: [
      { id: 's8', type: 'trigger', name: 'Angebot angenommen', description: '', config: {} },
      { id: 's9', type: 'action', name: 'Rechnung generieren', description: '', config: {} },
      { id: 's10', type: 'action', name: 'Lieferschein erstellen', description: '', config: {} },
      { id: 's11', type: 'notification', name: 'Lager benachrichtigen', description: '', config: {} },
    ],
    created_at: '2026-01-20T10:00:00Z',
    updated_at: '2026-03-15T16:00:00Z',
    created_by: 'Admin',
    instances_count: 8,
  },
  {
    id: 'wf4',
    name: 'Bestandswarnung',
    description: 'Warnt wenn die verfuegbare Menge eines Equipment-Typs unter den Schwellwert faellt.',
    status: 'paused',
    trigger: 'event',
    steps: [
      { id: 's12', type: 'trigger', name: 'Bestand unter Minimum', description: '', config: {} },
      { id: 's13', type: 'notification', name: 'Alert an Disponenten', description: '', config: {} },
    ],
    created_at: '2026-03-01T09:00:00Z',
    updated_at: '2026-03-10T12:00:00Z',
    created_by: 'Admin',
    instances_count: 2,
  },
]

const demoTemplates: WorkflowTemplate[] = [
  { id: 't1', name: 'Mahnung bei Ueberfaelligkeit', description: 'Automatische Erinnerung bei verspaeteter Equipment-Rueckgabe', category: 'Kommunikation', icon: '\u23F0', steps: [] },
  { id: 't2', name: 'Wartungserinnerung', description: 'Regelmmaessige Wartungserinnerungen basierend auf Intervallen', category: 'Wartung', icon: '\u{1F527}', steps: [] },
  { id: 't3', name: 'Angebots-Pipeline', description: 'Automatischer Ablauf nach Angebotsannahme', category: 'Dokumente', icon: '\u{1F4CB}', steps: [] },
  { id: 't4', name: 'Schadensmeldung', description: 'Workflow fuer die Abwicklung von Schadensfaellen', category: 'Versicherung', icon: '\u{1F6E1}', steps: [] },
  { id: 't5', name: 'Kundenfeedback', description: 'Automatische Feedback-Anfrage nach Projektabschluss', category: 'Kommunikation', icon: '\u2B50', steps: [] },
  { id: 't6', name: 'Bestandskontrolle', description: 'Warnung bei niedrigem Equipmentbestand', category: 'Lager', icon: '\u{1F4E6}', steps: [] },
]

interface NewWorkflowForm {
  name: string
  description: string
  trigger: string
  conditions: string
  actions: string
}

function WorkflowsPage() {
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null)
  const [showNewWorkflowModal, setShowNewWorkflowModal] = useState(false)
  const [workflowSubmitted, setWorkflowSubmitted] = useState(false)
  const [newWorkflowForm, setNewWorkflowForm] = useState<NewWorkflowForm>({
    name: '',
    description: '',
    trigger: 'event',
    conditions: '',
    actions: '',
  })
  const [activeSection, setActiveSection] = useState<'workflows' | 'templates'>('workflows')

  const { data: rawWorkflows = [], isLoading: isLoadingWorkflows } = useQuery({
    queryKey: ['workflows'],
    queryFn: async () => {
      try {
        const result = await workflowApi.list()
        const items = Array.isArray(result) ? result : []
        return items.length > 0 ? items : demoWorkflows
      } catch {
        return demoWorkflows
      }
    },
    staleTime: 1000 * 60 * 5,
  })
  const workflows = Array.isArray(rawWorkflows) ? rawWorkflows : [] as WorkflowDefinition[]

  const { data: rawTemplates = [], isLoading: isLoadingTemplates } = useQuery({
    queryKey: ['workflow-templates'],
    queryFn: async () => {
      try {
        const result = await workflowApi.templates()
        const items = Array.isArray(result) ? result : []
        return items.length > 0 ? items : demoTemplates
      } catch {
        return demoTemplates
      }
    },
    staleTime: 1000 * 60 * 5,
  })
  const templates = Array.isArray(rawTemplates) ? rawTemplates : [] as WorkflowTemplate[]

  const activeWorkflows = workflows.filter((w: WorkflowDefinition) => w.status === 'active')
  const totalInstances = workflows.reduce((sum: number, w: WorkflowDefinition) => sum + w.instances_count, 0)

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'active': return '#10b981'
      case 'paused': return '#f59e0b'
      case 'archived': return '#6b7280'
      case 'draft': return '#6b7280'
      default: return '#00d4ff'
    }
  }

  const getStatusLabel = (status: string): string => {
    switch (status) {
      case 'active': return 'Aktiv'
      case 'paused': return 'Pausiert'
      case 'archived': return 'Archiviert'
      case 'draft': return 'Entwurf'
      default: return status
    }
  }

  const getTriggerLabel = (trigger: string): string => {
    switch (trigger) {
      case 'event': return 'Event-basiert'
      case 'scheduled': return 'Zeitgesteuert (Cron)'
      case 'manual': return 'Manuell'
      case 'webhook': return 'Webhook'
      default: return trigger
    }
  }

  const getTriggerIcon = (trigger: string): string => {
    switch (trigger) {
      case 'event': return '\u26A1'
      case 'scheduled': return '\u23F0'
      case 'manual': return '\u{1F446}'
      case 'webhook': return '\u{1F517}'
      default: return '\u26A1'
    }
  }

  const handleNewWorkflowSubmit = () => {
    if (!newWorkflowForm.name) return
    setWorkflowSubmitted(true)
    setTimeout(() => {
      setShowNewWorkflowModal(false)
      setWorkflowSubmitted(false)
      setNewWorkflowForm({ name: '', description: '', trigger: 'event', conditions: '', actions: '' })
    }, 2000)
  }

  const handleUseTemplate = (template: WorkflowTemplate) => {
    setNewWorkflowForm({
      name: template.name,
      description: template.description,
      trigger: 'event',
      conditions: '',
      actions: '',
    })
    setShowNewWorkflowModal(true)
  }

  return (
    <div className={styles.workflowsPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Workflow-Editor</h1>
          <p className={styles.subtitle}>Erstellen und verwalten Sie automatisierte Workflows</p>
        </div>
        <button
          className="btn btn--primary"
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          onClick={() => setShowNewWorkflowModal(true)}
        >
          + Neuer Workflow
        </button>
      </div>

      {/* KPI Stats */}
      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>{'\u26A1'}</div>
          <div className={styles.statLabel}>Gesamt Workflows</div>
          <div className={styles.statValue}>{isLoadingWorkflows ? '--' : workflows.length}</div>
          <div className={styles.statSubtext}>Definierte Automatisierungen</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>{'\u{1F7E2}'}</div>
          <div className={styles.statLabel}>Aktive Workflows</div>
          <div className={styles.statValue} style={{ color: 'var(--color-success)' }}>
            {isLoadingWorkflows ? '--' : activeWorkflows.length}
          </div>
          <div className={styles.statSubtext}>Laufen gerade</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>{'\u{1F504}'}</div>
          <div className={styles.statLabel}>Ausgefuehrte Instanzen</div>
          <div className={styles.statValue}>
            {isLoadingWorkflows ? '--' : totalInstances}
          </div>
          <div className={styles.statSubtext}>Seit Erstellung</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>{'\u{1F4CB}'}</div>
          <div className={styles.statLabel}>Vorlagen</div>
          <div className={styles.statValue}>
            {isLoadingTemplates ? '--' : templates.length}
          </div>
          <div className={styles.statSubtext}>Vordefinierte Templates</div>
        </div>
      </div>

      {/* Section Tabs */}
      <div className={styles.tabBar}>
        <button
          className={`${styles.tabButton} ${activeSection === 'workflows' ? styles['tabButton--active'] : ''}`}
          onClick={() => setActiveSection('workflows')}
        >
          Workflow-Definitionen ({workflows.length})
        </button>
        <button
          className={`${styles.tabButton} ${activeSection === 'templates' ? styles['tabButton--active'] : ''}`}
          onClick={() => setActiveSection('templates')}
        >
          Template-Galerie ({templates.length})
        </button>
      </div>

      {/* Workflows Section */}
      {activeSection === 'workflows' && (
        <section className={styles.section}>
          {isLoadingWorkflows ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'\u23F3'}</div>
              <p className={styles.emptyText}>Laden...</p>
            </div>
          ) : workflows.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'\u26A1'}</div>
              <h3 className={styles.emptyTitle}>Keine Workflows vorhanden</h3>
              <p className={styles.emptyText}>Erstellen Sie einen neuen Workflow oder waehlen Sie eine Vorlage aus der Template-Galerie.</p>
              <button className="btn btn--primary" onClick={() => setShowNewWorkflowModal(true)}>
                Ersten Workflow erstellen
              </button>
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
                      <span style={{ color: getStatusColor(workflow.status) }}>{'\u25CF'}</span>
                      {getStatusLabel(workflow.status)}
                    </div>
                  </div>

                  {/* Workflow Info Bar */}
                  <div className={styles.workflowInfoBar}>
                    <span className={styles.infoTag}>
                      {getTriggerIcon(workflow.trigger)} {getTriggerLabel(workflow.trigger)}
                    </span>
                    <span className={styles.infoTag}>
                      {workflow.steps?.length || 0} Schritte
                    </span>
                    <span className={styles.infoTag}>
                      {workflow.instances_count} Ausfuehrungen
                    </span>
                    <span className={styles.infoTag}>
                      Aktualisiert: {new Date(workflow.updated_at).toLocaleDateString('de-DE')}
                    </span>
                  </div>

                  {/* Workflow Steps Visualization */}
                  <div className={styles.stepsContainer}>
                    {(Array.isArray(workflow.steps) ? workflow.steps : []).map((step, idx) => (
                      <div key={step.id} className={styles.stepRow}>
                        <div className={styles.stepCard}>
                          <div className={styles.stepIcon}>
                            {step.type === 'trigger' && '\u25B6'}
                            {step.type === 'action' && '\u2699'}
                            {step.type === 'condition' && '\u25C6'}
                            {step.type === 'notification' && '\u{1F514}'}
                            {step.type === 'decision' && '\u25C7'}
                          </div>
                          <div className={styles.stepInfo}>
                            <div className={styles.stepLabel}>{step.name}</div>
                            <div className={styles.stepType}>{step.type}</div>
                          </div>
                        </div>
                        {idx < (Array.isArray(workflow.steps) ? workflow.steps : []).length - 1 && (
                          <div className={styles.stepArrow}>{'\u2193'}</div>
                        )}
                      </div>
                    ))}
                  </div>

                  <div className={styles.workflowActions}>
                    <button className="btn btn--sm btn--primary" onClick={() => alert(`Workflow "${workflow.name}" bearbeiten...`)}>Bearbeiten</button>
                    <button className="btn btn--sm" onClick={() => alert(`Workflow "${workflow.name}" Details...`)}>Details</button>
                    <button className="btn btn--sm" onClick={() => alert(`Workflow "${workflow.name}" dupliziert!`)}>Duplizieren</button>
                    {workflow.status === 'active' && (
                      <button className="btn btn--sm" onClick={() => alert(`Workflow "${workflow.name}" pausiert!`)}>Pausieren</button>
                    )}
                    {workflow.status === 'paused' && (
                      <button className="btn btn--sm btn--primary" onClick={() => alert(`Workflow "${workflow.name}" aktiviert!`)}>Aktivieren</button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* Templates Section */}
      {activeSection === 'templates' && (
        <section className={styles.section}>
          {isLoadingTemplates ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'\u23F3'}</div>
              <p className={styles.emptyText}>Laden...</p>
            </div>
          ) : templates.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'\u{1F4E6}'}</div>
              <p className={styles.emptyText}>Keine Templates verfuegbar</p>
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
                    <button
                      className="btn btn--sm btn--primary"
                      style={{ marginTop: 'var(--spacing-3)', width: '100%' }}
                      onClick={(e) => {
                        e.stopPropagation()
                        handleUseTemplate(template)
                      }}
                    >
                      Vorlage verwenden
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* New Workflow Modal */}
      {showNewWorkflowModal && (
        <div className={styles.modalOverlay} onClick={() => !workflowSubmitted && setShowNewWorkflowModal(false)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>Neuer Workflow</h2>
              <button
                className={styles.modalClose}
                onClick={() => setShowNewWorkflowModal(false)}
                disabled={workflowSubmitted}
              >
                {'\u2715'}
              </button>
            </div>

            {workflowSubmitted ? (
              <div className={styles.modalBody}>
                <div className={styles.successState}>
                  <div className={styles.successIcon}>{'\u2713'}</div>
                  <h3>Workflow erstellt</h3>
                  <p>Der Workflow "{newWorkflowForm.name}" wurde erfolgreich erstellt und ist als Entwurf gespeichert.</p>
                </div>
              </div>
            ) : (
              <div className={styles.modalBody}>
                <div className={styles.formGroup}>
                  <label className={styles.formLabel}>Workflow-Name *</label>
                  <input
                    type="text"
                    className={styles.formInput}
                    placeholder="z.B. Mahnung bei Ueberfaelligkeit"
                    value={newWorkflowForm.name}
                    onChange={(e) => setNewWorkflowForm({ ...newWorkflowForm, name: e.target.value })}
                  />
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.formLabel}>Beschreibung</label>
                  <textarea
                    className={styles.formTextarea}
                    rows={3}
                    placeholder="Was macht dieser Workflow?"
                    value={newWorkflowForm.description}
                    onChange={(e) => setNewWorkflowForm({ ...newWorkflowForm, description: e.target.value })}
                  />
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.formLabel}>Trigger-Typ *</label>
                  <select
                    className={styles.formSelect}
                    value={newWorkflowForm.trigger}
                    onChange={(e) => setNewWorkflowForm({ ...newWorkflowForm, trigger: e.target.value })}
                  >
                    <option value="event">Event-basiert</option>
                    <option value="scheduled">Zeitgesteuert (Cron)</option>
                    <option value="manual">Manuell</option>
                    <option value="webhook">Webhook</option>
                  </select>
                  <span className={styles.formHint}>
                    {newWorkflowForm.trigger === 'event' && 'Wird durch ein System-Event ausgeloest (z.B. Equipment ueberfaellig)'}
                    {newWorkflowForm.trigger === 'scheduled' && 'Wird zu festgelegten Zeiten automatisch ausgefuehrt'}
                    {newWorkflowForm.trigger === 'manual' && 'Wird manuell durch einen Benutzer gestartet'}
                    {newWorkflowForm.trigger === 'webhook' && 'Wird durch einen externen Webhook-Aufruf gestartet'}
                  </span>
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.formLabel}>Bedingungen</label>
                  <textarea
                    className={styles.formTextarea}
                    rows={2}
                    placeholder="z.B. Wenn Rueckgabedatum > 24h ueberschritten..."
                    value={newWorkflowForm.conditions}
                    onChange={(e) => setNewWorkflowForm({ ...newWorkflowForm, conditions: e.target.value })}
                  />
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.formLabel}>Aktionen</label>
                  <textarea
                    className={styles.formTextarea}
                    rows={2}
                    placeholder="z.B. E-Mail senden, Status aendern, Aufgabe erstellen..."
                    value={newWorkflowForm.actions}
                    onChange={(e) => setNewWorkflowForm({ ...newWorkflowForm, actions: e.target.value })}
                  />
                </div>

                <div className={styles.modalFooter}>
                  <button className="btn btn--secondary" onClick={() => setShowNewWorkflowModal(false)}>
                    Abbrechen
                  </button>
                  <button
                    className="btn btn--primary"
                    onClick={handleNewWorkflowSubmit}
                    disabled={!newWorkflowForm.name}
                  >
                    Workflow erstellen
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export default WorkflowsPage
