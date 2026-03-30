import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as crewApi from '@/services/crew';
import type { CrewMember } from '@/types/crew';
import './CrewList.scss';

type ViewMode = 'list' | 'availability';

const WEEKDAYS = ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So'] as const;

function getWeekDates(): string[] {
  const now = new Date();
  const dayOfWeek = now.getDay();
  // Monday = 0 offset, Sunday = 6
  const mondayOffset = dayOfWeek === 0 ? -6 : 1 - dayOfWeek;
  const monday = new Date(now);
  monday.setDate(now.getDate() + mondayOffset);

  return WEEKDAYS.map((_, i) => {
    const date = new Date(monday);
    date.setDate(monday.getDate() + i);
    return date.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' });
  });
}

function AvailabilityGrid({ members }: { members: CrewMember[] }) {
  const weekDates = useMemo(() => getWeekDates(), []);

  if (members.length === 0) {
    return (
      <div className="availability-empty">
        Keine Crew-Mitglieder vorhanden
      </div>
    );
  }

  return (
    <div className="availability-grid">
      <table className="availability-grid__table">
        <thead>
          <tr>
            <th className="availability-grid__name-col">Mitarbeiter</th>
            {WEEKDAYS.map((day, i) => (
              <th key={day} className="availability-grid__day-col">
                <div className="availability-grid__day-name">{day}</div>
                <div className="availability-grid__day-date">{weekDates[i]}</div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {members.map((member) => (
            <tr key={member.id}>
              <td className="availability-grid__name">
                {member.first_name} {member.last_name}
                <span className="availability-grid__role">{member.role}</span>
              </td>
              {WEEKDAYS.map((day) => (
                <td key={day} className="availability-grid__cell">
                  <div
                    className="availability-grid__dot availability-grid__dot--unknown"
                    title="Keine Verfuegbarkeitsdaten"
                  />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      <div className="availability-grid__legend">
        <span className="availability-grid__legend-item">
          <span className="availability-grid__dot availability-grid__dot--available" /> Verfuegbar
        </span>
        <span className="availability-grid__legend-item">
          <span className="availability-grid__dot availability-grid__dot--unavailable" /> Nicht verfuegbar
        </span>
        <span className="availability-grid__legend-item">
          <span className="availability-grid__dot availability-grid__dot--unknown" /> Unbekannt
        </span>
      </div>
    </div>
  );
}

const columns: Column<CrewMember>[] = [
  { key: 'first_name', label: 'Vorname' },
  { key: 'last_name', label: 'Nachname' },
  { key: 'email', label: 'Email', render: (row) => row.email || '-' },
  { key: 'phone', label: 'Telefon', render: (row) => row.phone || '-' },
  { key: 'role', label: 'Rolle' },
  {
    key: 'hourly_rate',
    label: 'Stundensatz',
    render: (row) => `${(row.hourly_rate / 100).toFixed(2)} EUR`,
  },
  {
    key: 'is_active',
    label: 'Status',
    render: (row) => <StatusBadge status={row.is_active ? 'active' : 'cancelled'} />,
  },
];

export default function CrewList() {
  const [view, setView] = useState<ViewMode>('list');

  const { data, isLoading } = useQuery({
    queryKey: ['crew-members'],
    queryFn: crewApi.listCrewMembers,
  });

  const items: CrewMember[] = Array.isArray(data) ? data : [];

  return (
    <PageWrapper title="Crew">
      <div className="crew-view-toggle">
        <button
          className={`crew-view-toggle__btn ${view === 'list' ? 'crew-view-toggle__btn--active' : ''}`}
          onClick={() => setView('list')}
        >
          Liste
        </button>
        <button
          className={`crew-view-toggle__btn ${view === 'availability' ? 'crew-view-toggle__btn--active' : ''}`}
          onClick={() => setView('availability')}
        >
          Verfuegbarkeit
        </button>
      </div>

      {view === 'list' && (
        <DataTable<CrewMember>
          columns={columns}
          data={items}
          loading={isLoading}
          emptyMessage="Keine Crew-Mitglieder vorhanden"
        />
      )}

      {view === 'availability' && (
        isLoading
          ? <div className="availability-loading">Crew wird geladen...</div>
          : <AvailabilityGrid members={items} />
      )}
    </PageWrapper>
  );
}
