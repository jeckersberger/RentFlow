import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as projectApi from '@/services/projects';
import type { Project } from '@/types/project';
import './ProjectList.scss';

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('de-DE');
}

const columns: Column<Project>[] = [
  { key: 'name', label: 'Name' },
  {
    key: 'status',
    label: 'Status',
    render: (row) => <StatusBadge status={row.status} />,
  },
  {
    key: 'start_date',
    label: 'Startdatum',
    render: (row) => formatDate(row.start_date),
  },
  {
    key: 'end_date',
    label: 'Enddatum',
    render: (row) => formatDate(row.end_date),
  },
];

export default function ProjectList() {
  const navigate = useNavigate();

  const { data, isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectApi.list({ page: 1, per_page: 50 }),
  });

  const items: Project[] = Array.isArray(data) ? data : [];

  const addButton = (
    <button
      className="btn btn--primary"
      onClick={() => navigate('/projects/new')}
    >
      <Plus size={16} />
      <span>Neu</span>
    </button>
  );

  return (
    <PageWrapper title="Projekte" actions={addButton}>
      <DataTable<Project>
        columns={columns}
        data={items}
        loading={isLoading}
        onRowClick={(row) => navigate(`/projects/${row.id}`)}
        emptyMessage="Keine Projekte vorhanden"
      />
    </PageWrapper>
  );
}
