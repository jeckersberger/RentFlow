import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { GitBranch } from 'lucide-react';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import { workflowService, WorkflowDefinition } from '@/services/workflow';
import './WorkflowList.scss';

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('de-DE');
}

function formatTrigger(type?: string): string {
  if (!type) return '-';
  const map: Record<string, string> = {
    manual: 'Manuell',
    event: 'Ereignis',
    schedule: 'Zeitplan',
    webhook: 'Webhook',
  };
  return map[type] ?? type;
}

const columns: Column<WorkflowDefinition>[] = [
  { key: 'name', label: 'Name' },
  {
    key: 'type',
    label: 'Trigger',
    render: (row) => formatTrigger(row.type),
  },
  {
    key: 'steps',
    label: 'Status',
    render: (row) => {
      const stepCount = Array.isArray(row.steps) ? row.steps.length : 0;
      return stepCount > 0
        ? <StatusBadge status="active" />
        : <StatusBadge status="draft" />;
    },
  },
  {
    key: 'created_at',
    label: 'Erstellt am',
    render: (row) => formatDate(row.created_at),
  },
];

export default function WorkflowList() {
  const navigate = useNavigate();

  const { data, isLoading } = useQuery({
    queryKey: ['workflow-definitions'],
    queryFn: workflowService.listDefinitions,
  });

  const items: WorkflowDefinition[] = Array.isArray(data) ? data : [];

  const editorButton = (
    <button
      className="btn btn--primary"
      onClick={() => navigate('/workflows/editor')}
    >
      <GitBranch size={16} />
      <span>Workflow-Editor</span>
    </button>
  );

  return (
    <PageWrapper title="Workflows" actions={editorButton}>
      <DataTable<WorkflowDefinition>
        columns={columns}
        data={items}
        loading={isLoading}
        onRowClick={() => navigate('/workflows/editor')}
        emptyMessage="Keine Workflows vorhanden"
      />
    </PageWrapper>
  );
}
