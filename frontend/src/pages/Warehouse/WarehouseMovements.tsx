import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import * as warehouseApi from '@/services/warehouse';
import type { Movement } from '@/types/warehouse';
import './WarehouseMovements.scss';

const columns: Column<Movement>[] = [
  { key: 'equipment_id', label: 'Equipment', render: (row) => row.equipment_id?.slice(0, 8) || '-' },
  { key: 'from_location_id', label: 'Von', render: (row) => row.from_location_id?.slice(0, 8) || 'Extern' },
  { key: 'to_location_id', label: 'Nach', render: (row) => row.to_location_id?.slice(0, 8) || 'Extern' },
  { key: 'quantity', label: 'Menge' },
  { key: 'reason', label: 'Grund', render: (row) => row.reason || '-' },
  {
    key: 'created_at',
    label: 'Datum',
    render: (row) =>
      new Date(row.created_at).toLocaleString('de-DE', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      }),
  },
];

export default function WarehouseMovements() {
  const { data, isLoading } = useQuery({
    queryKey: ['movements'],
    queryFn: () => warehouseApi.listMovements({ page: 1, per_page: 50 }),
  });

  const movements: Movement[] = Array.isArray(data) ? data : [];

  return (
    <PageWrapper title="Warenbewegungen">
      <DataTable<Movement>
        columns={columns}
        data={movements}
        loading={isLoading}
        emptyMessage="Keine Warenbewegungen vorhanden"
      />
    </PageWrapper>
  );
}
