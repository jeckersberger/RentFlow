import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import api from '@/services/api';
import './MaintenanceList.scss';

interface MaintenanceTask {
  id: string;
  title: string;
  equipment_id: string;
  status: string;
  priority: string;
  due_date: string;
  cost: number;
  created_at: string;
}

const columns: Column<MaintenanceTask>[] = [
  { key: 'title', label: 'Aufgabe' },
  { key: 'priority', label: 'Prioritaet' },
  { key: 'due_date', label: 'Faellig', render: (row) => row.due_date || '-' },
  { key: 'cost', label: 'Kosten', render: (row) => row.cost ? `${(row.cost / 100).toFixed(2)} €` : '-' },
  { key: 'status', label: 'Status', render: (row) => <StatusBadge status={row.status} /> },
];

export default function MaintenanceList() {
  const { data, isLoading } = useQuery({
    queryKey: ['maintenance-tasks'],
    queryFn: () => api.get('/api/v1/maintenance-tasks') as unknown as MaintenanceTask[],
  });

  const items: MaintenanceTask[] = Array.isArray(data) ? data : [];

  return (
    <PageWrapper title="Wartung">
      <DataTable<MaintenanceTask>
        columns={columns}
        data={items}
        loading={isLoading}
        emptyMessage="Keine Wartungsaufgaben vorhanden"
      />
    </PageWrapper>
  );
}
