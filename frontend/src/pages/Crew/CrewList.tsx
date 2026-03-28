import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as crewApi from '@/services/crew';
import type { CrewMember } from '@/types/crew';
import './CrewList.scss';

const columns: Column<CrewMember>[] = [
  { key: 'first_name', label: 'Vorname' },
  { key: 'last_name', label: 'Nachname' },
  { key: 'email', label: 'Email', render: (row) => row.email || '-' },
  { key: 'phone', label: 'Telefon', render: (row) => row.phone || '-' },
  { key: 'role', label: 'Rolle' },
  {
    key: 'hourly_rate',
    label: 'Stundensatz',
    render: (row) => `${(row.hourly_rate / 100).toFixed(2)} €`,
  },
  {
    key: 'is_active',
    label: 'Status',
    render: (row) => <StatusBadge status={row.is_active ? 'active' : 'cancelled'} />,
  },
];

export default function CrewList() {
  const { data, isLoading } = useQuery({
    queryKey: ['crew-members'],
    queryFn: crewApi.listCrewMembers,
  });

  const items: CrewMember[] = Array.isArray(data) ? data : [];

  return (
    <PageWrapper title="Crew">
      <DataTable<CrewMember>
        columns={columns}
        data={items}
        loading={isLoading}
        emptyMessage="Keine Crew-Mitglieder vorhanden"
      />
    </PageWrapper>
  );
}
