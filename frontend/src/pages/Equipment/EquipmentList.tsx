import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as equipmentApi from '@/services/equipment';
import type { Equipment } from '@/types/equipment';
import './EquipmentList.scss';

const columns: Column<Equipment>[] = [
  { key: 'name', label: 'Name' },
  {
    key: 'category_id',
    label: 'Kategorie',
    render: (row) => row.category_id || '-',
  },
  {
    key: 'status',
    label: 'Status',
    render: (row) => <StatusBadge status={row.status} />,
  },
  { key: 'barcode', label: 'Barcode', render: (row) => row.barcode || '-' },
  {
    key: 'condition',
    label: 'Zustand',
    render: (row) => <StatusBadge status={row.condition} />,
  },
];

export default function EquipmentList() {
  const navigate = useNavigate();

  const { data, isLoading } = useQuery({
    queryKey: ['equipment'],
    queryFn: () => equipmentApi.list({ page: 1, per_page: 50 }),
  });

  // The api interceptor unwraps response.data.data -> returns the array directly
  const items: Equipment[] = Array.isArray(data) ? data : [];

  const addButton = (
    <button
      className="btn btn--primary"
      onClick={() => navigate('/equipment/new')}
    >
      <Plus size={16} />
      <span>Neu</span>
    </button>
  );

  return (
    <PageWrapper title="Equipment" actions={addButton}>
      <DataTable<Equipment>
        columns={columns}
        data={items}
        loading={isLoading}
        onRowClick={(row) => navigate(`/equipment/${row.id}`)}
        emptyMessage="Kein Equipment vorhanden"
      />
    </PageWrapper>
  );
}
