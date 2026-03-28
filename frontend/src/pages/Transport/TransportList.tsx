import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import api from '@/services/api';
import './TransportList.scss';

interface TransportOrder {
  id: string;
  type: string;
  status: string;
  pickup_address: string;
  delivery_address: string;
  scheduled_at: string;
  created_at: string;
}

const columns: Column<TransportOrder>[] = [
  { key: 'type', label: 'Typ' },
  { key: 'pickup_address', label: 'Abholung', render: (row) => row.pickup_address || '-' },
  { key: 'delivery_address', label: 'Lieferung', render: (row) => row.delivery_address || '-' },
  { key: 'scheduled_at', label: 'Geplant', render: (row) => row.scheduled_at ? new Date(row.scheduled_at).toLocaleDateString('de-DE') : '-' },
  { key: 'status', label: 'Status', render: (row) => <StatusBadge status={row.status} /> },
];

export default function TransportList() {
  const { data, isLoading } = useQuery({
    queryKey: ['transport-orders'],
    queryFn: () => api.get('/api/v1/transport-orders') as unknown as TransportOrder[],
  });

  const items: TransportOrder[] = Array.isArray(data) ? data : [];

  return (
    <PageWrapper title="Transport">
      <DataTable<TransportOrder>
        columns={columns}
        data={items}
        loading={isLoading}
        emptyMessage="Keine Transportauftraege vorhanden"
      />
    </PageWrapper>
  );
}
