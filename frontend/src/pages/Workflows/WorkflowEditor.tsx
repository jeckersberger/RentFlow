import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import {
  GitBranch,
  Zap,
  Filter,
  Play,
  CheckCircle,
  FileText,
  ArrowLeft,
  Copy,
} from 'lucide-react';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { Modal } from '@/components/Modal/Modal';
import api from '@/services/api';
import { workflowService, WorkflowDefinition } from '@/services/workflow';
import './WorkflowEditor.scss';

// ── Types ──

interface WorkflowStep {
  name: string;
  role: string;
}

interface WorkflowTemplate {
  id: string;
  name: string;
  description?: string;
  type?: string;
  steps?: WorkflowStep[];
}

type NodeVariant = 'trigger' | 'condition' | 'action' | 'end';

interface NodeDisplay {
  variant: NodeVariant;
  label: string;
  sub: string;
  icon: React.ReactNode;
}

// ── Helpers ──

function mapRoleToVariant(role: string): NodeVariant {
  const lower = role.toLowerCase();
  if (lower.includes('trigger') || lower.includes('start')) return 'trigger';
  if (lower.includes('condition') || lower.includes('check') || lower.includes('filter')) return 'condition';
  if (lower.includes('end') || lower.includes('complete') || lower.includes('finish')) return 'end';
  return 'action';
}

function iconForVariant(variant: NodeVariant): React.ReactNode {
  switch (variant) {
    case 'trigger':
      return <Zap size={18} />;
    case 'condition':
      return <Filter size={18} />;
    case 'action':
      return <Play size={18} />;
    case 'end':
      return <CheckCircle size={18} />;
  }
}

function variantLabel(variant: NodeVariant): string {
  switch (variant) {
    case 'trigger':
      return 'Trigger';
    case 'condition':
      return 'Bedingung';
    case 'action':
      return 'Aktion';
    case 'end':
      return 'Ende';
  }
}

function buildNodes(steps: WorkflowStep[]): NodeDisplay[] {
  if (steps.length === 0) return [];

  const nodes: NodeDisplay[] = [];

  // Always start with a trigger node
  const first = steps[0];
  const firstVariant = mapRoleToVariant(first.role);
  if (firstVariant !== 'trigger') {
    nodes.push({
      variant: 'trigger',
      label: 'Trigger',
      sub: 'Workflow-Start',
      icon: iconForVariant('trigger'),
    });
  }

  for (const step of steps) {
    const variant = mapRoleToVariant(step.role);
    nodes.push({
      variant,
      label: step.name,
      sub: variantLabel(variant) + (step.role ? ` (${step.role})` : ''),
      icon: iconForVariant(variant),
    });
  }

  // Always end with an end node if the last isn't one
  const lastVariant = nodes[nodes.length - 1]?.variant;
  if (lastVariant !== 'end') {
    nodes.push({
      variant: 'end',
      label: 'Ende',
      sub: 'Workflow abgeschlossen',
      icon: iconForVariant('end'),
    });
  }

  return nodes;
}

function formatTriggerType(type?: string): string {
  if (!type) return '-';
  const map: Record<string, string> = {
    manual: 'Manuell',
    event: 'Ereignis',
    schedule: 'Zeitplan',
    webhook: 'Webhook',
  };
  return map[type] ?? type;
}

// ── Component ──

export default function WorkflowEditor() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [templateModalOpen, setTemplateModalOpen] = useState(false);

  // Fetch workflow definitions
  const { data: definitionsRaw, isLoading: defsLoading } = useQuery({
    queryKey: ['workflow-definitions'],
    queryFn: workflowService.listDefinitions,
  });

  const definitions: WorkflowDefinition[] = Array.isArray(definitionsRaw)
    ? definitionsRaw
    : [];

  const selected = definitions.find((d) => d.id === selectedId) ?? null;

  // Fetch templates (only when modal is open)
  const { data: templatesRaw, isLoading: templatesLoading } = useQuery({
    queryKey: ['workflow-templates'],
    queryFn: () => api.get('/api/v1/workflow-templates') as unknown as WorkflowTemplate[],
    enabled: templateModalOpen,
  });

  const templates: WorkflowTemplate[] = Array.isArray(templatesRaw)
    ? templatesRaw
    : [];

  // Mutation: create from template
  const createFromTemplate = useMutation({
    mutationFn: (template: WorkflowTemplate) =>
      workflowService.createDefinition({
        name: `${template.name} (Kopie)`,
        type: template.type ?? 'manual',
        steps: template.steps ?? [],
      }),
    onSuccess: (result) => {
      toast.success('Workflow aus Vorlage erstellt');
      queryClient.invalidateQueries({ queryKey: ['workflow-definitions'] });
      setTemplateModalOpen(false);
      const created = result as unknown as WorkflowDefinition;
      if (created?.id) {
        setSelectedId(created.id);
      }
    },
    onError: () => {
      toast.error('Fehler beim Erstellen des Workflows');
    },
  });

  const handleSelectDefinition = useCallback((id: string) => {
    setSelectedId(id);
  }, []);

  const handleTemplateSelect = useCallback(
    (template: WorkflowTemplate) => {
      createFromTemplate.mutate(template);
    },
    [createFromTemplate],
  );

  const steps = selected?.steps ?? [];
  const nodes = buildNodes(steps);
  const isActive = Array.isArray(steps) && steps.length > 0;

  return (
    <PageWrapper
      title="Workflow-Editor"
      actions={
        <button
          className="wf-editor__template-btn"
          style={{
            background: 'transparent',
            color: 'var(--text-secondary, #94a3b8)',
            border: '1px solid rgba(255,255,255,0.08)',
          }}
          onClick={() => navigate('/workflows')}
        >
          <ArrowLeft size={16} />
          Zurueck
        </button>
      }
    >
      {defsLoading ? (
        <div className="wf-editor__loading">
          <div className="loading-spinner" />
          <span>Workflows werden geladen...</span>
        </div>
      ) : (
        <div className="wf-editor">
          {/* Left panel: definition list */}
          <div className="wf-editor__sidebar">
            <div className="wf-editor__sidebar-header">Definitionen</div>
            <div className="wf-editor__sidebar-list">
              {definitions.length === 0 ? (
                <div className="wf-editor__sidebar-empty">
                  Keine Workflows vorhanden
                </div>
              ) : (
                definitions.map((def) => (
                  <div
                    key={def.id}
                    className={`wf-editor__sidebar-item ${
                      selectedId === def.id ? 'wf-editor__sidebar-item--active' : ''
                    }`}
                    onClick={() => handleSelectDefinition(def.id)}
                  >
                    <GitBranch size={16} className="wf-editor__sidebar-icon" />
                    <span>{def.name}</span>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Right panel: detail */}
          <div className="wf-editor__detail">
            {!selected ? (
              <div className="wf-editor__detail-empty">
                <GitBranch size={48} />
                <span>Workflow auswaehlen oder aus Vorlage erstellen</span>
              </div>
            ) : (
              <>
                {/* Header */}
                <div className="wf-editor__detail-header">
                  <div className="wf-editor__detail-title-row">
                    <h3 className="wf-editor__detail-name">{selected.name}</h3>
                    <div className="wf-editor__toggle">
                      <span className="wf-editor__toggle-label">
                        {isActive ? 'Aktiv' : 'Inaktiv'}
                      </span>
                      <div
                        className={`wf-editor__toggle-track ${
                          isActive ? 'wf-editor__toggle-track--active' : ''
                        }`}
                      >
                        <div className="wf-editor__toggle-thumb" />
                      </div>
                    </div>
                  </div>
                  <div className="wf-editor__detail-meta">
                    <div className="wf-editor__meta-item">
                      Trigger: <span>{formatTriggerType(selected.type)}</span>
                    </div>
                    <div className="wf-editor__meta-item">
                      Schritte: <span>{steps.length}</span>
                    </div>
                    <div className="wf-editor__meta-item">
                      Erstellt:{' '}
                      <span>
                        {selected.created_at
                          ? new Date(selected.created_at).toLocaleDateString('de-DE')
                          : '-'}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Steps */}
                <div className="wf-editor__steps">
                  <h4 className="wf-editor__steps-title">Ablauf</h4>
                  {nodes.length === 0 ? (
                    <div className="wf-editor__no-steps">
                      Keine Schritte definiert
                    </div>
                  ) : (
                    <div className="wf-editor__node-list">
                      {nodes.map((node, idx) => (
                        <div key={idx} className="wf-editor__node">
                          <div className="wf-editor__node-connector">
                            {idx > 0 && <div className="wf-editor__node-line" />}
                            <div
                              className={`wf-editor__node-circle wf-editor__node-circle--${node.variant}`}
                            >
                              {node.icon}
                            </div>
                            {idx < nodes.length - 1 && (
                              <div className="wf-editor__node-line" />
                            )}
                          </div>
                          <div className="wf-editor__node-body">
                            <span className="wf-editor__node-label">
                              {node.label}
                            </span>
                            <span className="wf-editor__node-sub">
                              {node.sub}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                {/* Bottom */}
                <div className="wf-editor__bottom">
                  <button
                    className="wf-editor__template-btn"
                    onClick={() => setTemplateModalOpen(true)}
                  >
                    <Copy size={16} />
                    Aus Vorlage erstellen
                  </button>
                </div>
              </>
            )}

            {/* Show template button even when nothing selected */}
            {!selected && (
              <div className="wf-editor__bottom">
                <button
                  className="wf-editor__template-btn"
                  onClick={() => setTemplateModalOpen(true)}
                >
                  <Copy size={16} />
                  Aus Vorlage erstellen
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Template selection modal */}
      <Modal
        isOpen={templateModalOpen}
        onClose={() => setTemplateModalOpen(false)}
        title="Workflow aus Vorlage erstellen"
        size="md"
      >
        {templatesLoading ? (
          <div className="wf-editor__template-loading">
            <div className="loading-spinner" />
            <span>Vorlagen werden geladen...</span>
          </div>
        ) : templates.length === 0 ? (
          <div className="wf-editor__template-empty">
            Keine Vorlagen verfuegbar
          </div>
        ) : (
          <div className="wf-editor__template-list">
            {templates.map((tpl) => (
              <div
                key={tpl.id}
                className="wf-editor__template-item"
                onClick={() => handleTemplateSelect(tpl)}
              >
                <div className="wf-editor__template-item-icon">
                  <FileText size={18} />
                </div>
                <div className="wf-editor__template-item-info">
                  <span className="wf-editor__template-item-name">
                    {tpl.name}
                  </span>
                  {tpl.description && (
                    <span className="wf-editor__template-item-desc">
                      {tpl.description}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </Modal>
    </PageWrapper>
  );
}
