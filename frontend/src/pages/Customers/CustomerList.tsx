import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as customerApi from '@/services/customers';
import type { Customer } from '@/types/customer';
import './CustomerList.scss';

const columns: Column<Customer>[] = [
  { key: 'company_name', label: 'Firma' },
  { key: 'email', label: 'Email', render: (row) => row.email || '-' },
  { key: 'phone', label: 'Telefon', render: (row) => row.phone || '-' },
  {
    key: 'is_active',
    label: 'Status',
    render: (row) => (
      <StatusBadge
        status={row.is_active ? 'active' : 'cancelled'}
      />
    ),
  },
];

export default function CustomerList() {
  const navigate = useNavigate();

  const { data, isLoading } = useQuery({
    queryKey: ['customers'],
    queryFn: () => customerApi.list({ page: 1, per_page: 50 }),
  });

  const items: Customer[] = Array.isArray(data) ? data : [];

  const addButton = (
    <button
      className="btn btn--primary"
      onClick={() => navigate('/customers/new')}
    >
      <Plus size={16} />
      <span>Neu</span>
    </button>
  );

  return (
    <PageWrapper title="Kunden" actions={addButton}>
      <DataTable<Customer>
        columns={columns}
        data={items}
        loading={isLoading}
        onRowClick={(row) => navigate(`/customers/${row.id}`)}
        emptyMessage="Keine Kunden vorhanden"
      />
    </PageWrapper>
  );
}
