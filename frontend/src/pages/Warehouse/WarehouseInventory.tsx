import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as warehouseApi from '@/services/warehouse';
import type { InventoryCheck } from '@/types/warehouse';
import './WarehouseInventory.scss';

const columns: Column<InventoryCheck>[] = [
  {
    key: 'status',
    label: 'Status',
    render: (row) => <StatusBadge status={row.status} />,
  },
  { key: 'zone_id', label: 'Zone', render: (row) => row.zone_id?.slice(0, 8) || 'Alle' },
  { key: 'expected_count', label: 'Soll' },
  { key: 'actual_count', label: 'Ist' },
  {
    key: 'discrepancy_count',
    label: 'Differenz',
    render: (row) => {
      const diff = row.discrepancy_count;
      if (diff === 0) return <span style={{ color: '#10b981' }}>0</span>;
      return <span style={{ color: '#ef4444' }}>{diff}</span>;
    },
  },
  {
    key: 'started_at',
    label: 'Gestartet',
    render: (row) =>
      new Date(row.started_at).toLocaleString('de-DE', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      }),
  },
  {
    key: 'completed_at',
    label: 'Abgeschlossen',
    render: (row) =>
      row.completed_at
        ? new Date(row.completed_at).toLocaleString('de-DE', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
          })
        : '-',
  },
];

export default function WarehouseInventory() {
  const { data, isLoading } = useQuery({
    queryKey: ['inventory-checks'],
    queryFn: () => warehouseApi.listInventoryChecks({ page: 1, per_page: 50 }),
  });

  const checks: InventoryCheck[] = Array.isArray(data) ? data : [];

  return (
    <PageWrapper title="Inventur">
      <DataTable<InventoryCheck>
        columns={columns}
        data={checks}
        loading={isLoading}
        emptyMessage="Keine Inventuren vorhanden"
      />
    </PageWrapper>
  );
}
