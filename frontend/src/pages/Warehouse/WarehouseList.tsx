import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Plus, Warehouse as WarehouseIcon, MapPin, ArrowRightLeft, ClipboardCheck } from 'lucide-react';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as warehouseApi from '@/services/warehouse';
import type { Warehouse, Zone, StockLocation } from '@/types/warehouse';
import './WarehouseList.scss';

type Tab = 'warehouses' | 'zones' | 'locations';

const warehouseColumns: Column<Warehouse>[] = [
  { key: 'name', label: 'Name' },
  { key: 'code', label: 'Code', render: (row) => row.code || '-' },
  { key: 'address', label: 'Adresse', render: (row) => row.address || '-' },
  {
    key: 'is_active',
    label: 'Status',
    render: (row) => <StatusBadge status={row.is_active ? 'active' : 'inactive'} />,
  },
];

const zoneColumns: Column<Zone>[] = [
  { key: 'name', label: 'Name' },
  { key: 'code', label: 'Code', render: (row) => row.code || '-' },
  {
    key: 'climate_controlled',
    label: 'Klima',
    render: (row) => (row.climate_controlled ? 'Ja' : 'Nein'),
  },
  {
    key: 'max_weight_kg',
    label: 'Max. Gewicht',
    render: (row) => (row.max_weight_kg ? `${row.max_weight_kg} kg` : '-'),
  },
];

const locationColumns: Column<StockLocation>[] = [
  { key: 'code', label: 'Code' },
  { key: 'barcode', label: 'Barcode', render: (row) => row.barcode || '-' },
  { key: 'level', label: 'Ebene', render: (row) => row.level ?? '-' },
  { key: 'bay', label: 'Fach', render: (row) => row.bay ?? '-' },
  {
    key: 'is_active',
    label: 'Status',
    render: (row) => <StatusBadge status={row.is_active ? 'active' : 'inactive'} />,
  },
];

export default function WarehouseList() {
  const [activeTab, setActiveTab] = useState<Tab>('warehouses');

  const { data: warehousesData, isLoading: warehousesLoading } = useQuery({
    queryKey: ['warehouses'],
    queryFn: () => warehouseApi.listWarehouses(),
    enabled: activeTab === 'warehouses',
  });

  const { data: zonesData, isLoading: zonesLoading } = useQuery({
    queryKey: ['zones'],
    queryFn: () => warehouseApi.listZones(),
    enabled: activeTab === 'zones',
  });

  const { data: locationsData, isLoading: locationsLoading } = useQuery({
    queryKey: ['locations'],
    queryFn: () => warehouseApi.listLocations({ page: 1, per_page: 100 }),
    enabled: activeTab === 'locations',
  });

  const warehouses: Warehouse[] = Array.isArray(warehousesData) ? warehousesData : [];
  const zones: Zone[] = Array.isArray(zonesData) ? zonesData : [];
  const locations: StockLocation[] = Array.isArray(locationsData) ? locationsData : [];

  return (
    <PageWrapper title="Lager">
      <div className="warehouse-tabs">
        <button
          className={`warehouse-tabs__tab ${activeTab === 'warehouses' ? 'warehouse-tabs__tab--active' : ''}`}
          onClick={() => setActiveTab('warehouses')}
        >
          <WarehouseIcon size={16} />
          Lager
        </button>
        <button
          className={`warehouse-tabs__tab ${activeTab === 'zones' ? 'warehouse-tabs__tab--active' : ''}`}
          onClick={() => setActiveTab('zones')}
        >
          <MapPin size={16} />
          Zonen
        </button>
        <button
          className={`warehouse-tabs__tab ${activeTab === 'locations' ? 'warehouse-tabs__tab--active' : ''}`}
          onClick={() => setActiveTab('locations')}
        >
          <ClipboardCheck size={16} />
          Stellplaetze
        </button>
      </div>

      {activeTab === 'warehouses' && (
        <DataTable<Warehouse>
          columns={warehouseColumns}
          data={warehouses}
          loading={warehousesLoading}
          emptyMessage="Keine Lager vorhanden"
        />
      )}

      {activeTab === 'zones' && (
        <DataTable<Zone>
          columns={zoneColumns}
          data={zones}
          loading={zonesLoading}
          emptyMessage="Keine Zonen vorhanden"
        />
      )}

      {activeTab === 'locations' && (
        <DataTable<StockLocation>
          columns={locationColumns}
          data={locations}
          loading={locationsLoading}
          emptyMessage="Keine Stellplaetze vorhanden"
        />
      )}
    </PageWrapper>
  );
}
