# MyRMS — Frontend-Komponenten (React)

**Stand:** 20. März 2026
**Version:** 1.0

---

## Überblick

Das MyRMS-Frontend ist eine **Progressive Web App (PWA)** auf Basis von React 18, TypeScript und Sass/SCSS. Es ist sowohl für Desktop-Nutzung (Marco, Thomas) als auch für mobile Nutzung (Lisa mit Zebra TC21, Kevin mit iPhone) optimiert.

---

## Tech-Stack Frontend

| Technologie | Paket | Begründung |
|------------|-------|-----------|
| Framework | React 18 | Concurrent Mode, Suspense, Server Components |
| Sprache | TypeScript 5.x | Typsicherheit, IntelliSense, Refactoring |
| Build-Tool | Vite 5 | Schnelles HMR, ESM-native, optimiertes Bundling |
| Styling | Sass/SCSS + CSS Modules | Keine Utility-CSS-Bloat, volle Kontrolle |
| UI-Bibliothek | Radix UI (headless) + eigenes Design-System | Zugänglichkeit (ARIA), vollständig anpassbar |
| State Management | Zustand + React Query (TanStack Query v5) | Leichtgewichtig, Server-State getrennt von Client-State |
| Routing | TanStack Router v1 | Type-safe Routes, File-based Routing |
| Formulare | React Hook Form + Zod | Performant, Schema-Validierung |
| Tabellen | TanStack Table v8 | Virtualisierung für große Datensätze |
| Charts | Recharts | React-nativ, einfach anpassbar |
| Datum/Zeit | date-fns | Kleines Bundle, Tree-shakeable |
| Internationalisierung | react-i18next | DE/EN von Anfang an |
| WebSocket | native WebSocket + Zustand | Real-time Notifications |
| Barcode-Scan | @zxing/browser | Kamera-Scanner für iOS/Android |
| PWA | vite-plugin-pwa (Workbox) | Service Worker, Offline-Support |
| Icons | Lucide React | Konsistentes Icon-Set, Tree-shakeable |
| Drag & Drop | dnd-kit | Modular, barrierefrei |
| PDF-Anzeige | react-pdf | Dokument-Vorschau |
| Testing | Vitest + React Testing Library | Unit + Integration Tests |
| E2E-Tests | Playwright | Browser-Tests |

---

## Design-System

### Tokens (SCSS-Variablen)

```scss
// tokens/_colors.scss
:root {
  // Primärfarben
  --color-primary-50:   #eff6ff;
  --color-primary-100:  #dbeafe;
  --color-primary-500:  #3b82f6;
  --color-primary-600:  #2563eb;
  --color-primary-700:  #1d4ed8;

  // Graustufen
  --color-gray-50:  #f9fafb;
  --color-gray-100: #f3f4f6;
  --color-gray-200: #e5e7eb;
  --color-gray-300: #d1d5db;
  --color-gray-400: #9ca3af;
  --color-gray-500: #6b7280;
  --color-gray-700: #374151;
  --color-gray-900: #111827;

  // Status-Farben
  --color-success: #22c55e;
  --color-warning: #f59e0b;
  --color-danger:  #ef4444;
  --color-info:    #3b82f6;

  // Scanner-Feedback (besonders groß & kontrastreich)
  --scanner-ok:     #16a34a;    // Kräftiges Grün
  --scanner-error:  #dc2626;    // Kräftiges Rot
  --scanner-warn:   #d97706;    // Kräftiges Gelb

  // Spacing
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-6: 24px;
  --space-8: 32px;

  // Typography
  --font-family: 'Inter', system-ui, sans-serif;
  --font-size-sm:   14px;
  --font-size-base: 16px;
  --font-size-lg:   18px;
  --font-size-xl:   20px;
  --font-size-2xl:  24px;

  // Border Radius
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;

  // Shadows
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
}

// Dark Mode
[data-theme="dark"] {
  --color-gray-50:  #111827;
  --color-gray-100: #1f2937;
  --color-gray-900: #f9fafb;
  // ...
}
```

### Basis-Komponenten (Design System)

```
src/components/ui/
├── Button/           Button, IconButton, ButtonGroup
├── Input/            Input, Textarea, Select, Checkbox, Switch, Radio
├── Form/             FormField, FormLabel, FormError, FormHint
├── Card/             Card, CardHeader, CardBody, CardFooter
├── Modal/            Modal, Dialog, Drawer, Sheet (mobile)
├── Table/            DataTable (TanStack), TableColumn, Pagination
├── Badge/            Badge, StatusBadge, CountBadge
├── Avatar/           Avatar, AvatarGroup
├── Tooltip/          Tooltip, Popover
├── Toast/            Toast, ToastContainer (Sonner)
├── Loading/          Spinner, Skeleton, ProgressBar
├── Empty/            EmptyState, ErrorState
├── Tabs/             Tabs, TabList, Tab, TabPanel
├── Accordion/        Accordion, AccordionItem
├── Command/          CommandPalette (⌘K Suche)
├── DatePicker/       DatePicker, DateRangePicker, TimePicker
├── FileUpload/       FileUpload, ImageUpload, DragDropZone
└── ColorPicker/      ColorPicker (für Kategorien)
```

---

## Routing-Struktur (TanStack Router)

```
/ (Layout: AppShell)
├── /login                        Nicht authentifiziert
├── /accept-invitation/:token     Einladung annehmen
├── /forgot-password
├── /reset-password/:token
│
├── / (Dashboard)                 Rollenbasiertes Dashboard
│
├── /projects                     Projekte
│   ├── /projects/new             Neues Projekt
│   ├── /projects/:id             Projektdetails
│   ├── /projects/:id/edit        Projekt bearbeiten
│   ├── /projects/:id/packing-list  Packliste
│   ├── /projects/:id/timeline    Gantt-Ansicht
│   └── /calendar                 Kalender-Ansicht
│
├── /equipment                    Inventar
│   ├── /equipment/new            Neues Equipment
│   ├── /equipment/:id            Equipment-Detail
│   ├── /equipment/:id/edit       Equipment bearbeiten
│   ├── /equipment/:id/history    Event-Historie
│   └── /equipment/import         Bulk-Import
│
├── /scanner                      Scanner-UI (Zebra/Handy optimiert)
│   ├── /scanner/checkin
│   ├── /scanner/checkout
│   └── /scanner/locate
│
├── /warehouse                    Lagerverwaltung
│   ├── /warehouse/locations      Lagerorte
│   ├── /warehouse/movements      Warenbewegungen
│   └── /warehouse/inventory      Inventur
│
├── /invoices                     Rechnungen
│   ├── /invoices/new             Neue Rechnung
│   ├── /invoices/:id             Rechnungsdetail
│   ├── /quotes                   Angebote
│   ├── /quotes/new
│   └── /quotes/:id
│
├── /crew                         Personal
│   ├── /crew/new
│   ├── /crew/:id
│   └── /crew/schedule            Wochenplan
│
├── /maintenance                  Wartung
│   ├── /maintenance/tasks
│   ├── /maintenance/plans
│   └── /maintenance/echeck       E-Check / DGUV V3
│
├── /transport                    Transport
│   ├── /transport/vehicles
│   └── /transport/tours
│
├── /federation                   Federation
│
├── /ai                           KI-Funktionen
│
├── /reports                      Reports & Dashboard
│
└── /settings                     Einstellungen
    ├── /settings/company         Firmendaten
    ├── /settings/users           Benutzer
    ├── /settings/roles           Rollen
    ├── /settings/categories      Kategorien
    ├── /settings/notifications   Benachrichtigungs-Präferenzen
    ├── /settings/integrations    Integrationen (KI, E-Mail)
    └── /settings/backup          Backup
```

---

## State-Management

### Architektur: Server State vs. Client State

```
Server State (React Query / TanStack Query):
  - Alle API-Daten (Equipment, Projekte, Rechnungen, ...)
  - Automatisches Caching, Background-Refetch, Optimistic Updates
  - Deklarativ: useQuery(), useMutation()

Client State (Zustand):
  - UI-State (Sidebar offen/zu, Ausgewählte Zeilen)
  - Authentifizierungs-State (currentUser, permissions)
  - Scanner-State (aktueller Job, Offline-Queue)
  - Notification-State (ungelesene Nachrichten)
  - Theme (light/dark)
```

### Zustand-Stores

```typescript
// src/stores/auth.store.ts
interface AuthStore {
  user: User | null;
  tenant: Tenant | null;
  permissions: string[];
  isAuthenticated: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
}

// src/stores/scanner.store.ts
interface ScannerStore {
  currentProjectId: string | null;
  scanMode: 'checkout' | 'checkin' | 'inventory' | 'locate';
  offlineQueue: OfflineScan[];
  lastScanResult: ScanResult | null;
  isSyncing: boolean;
  setProject: (projectId: string) => void;
  addToOfflineQueue: (scan: OfflineScan) => void;
  syncOfflineQueue: () => Promise<void>;
}

// src/stores/notifications.store.ts
interface NotificationStore {
  notifications: Notification[];
  unreadCount: number;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  addNotification: (notification: Notification) => void;
}

// src/stores/ui.store.ts
interface UIStore {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark' | 'system';
  activeFilters: Record<string, FilterValue>;
  toggleSidebar: () => void;
  setTheme: (theme: Theme) => void;
}
```

### React Query Patterns

```typescript
// src/queries/equipment.queries.ts
export const equipmentKeys = {
  all: ['equipment'] as const,
  list: (filters: EquipmentFilters) => ['equipment', 'list', filters] as const,
  detail: (id: string) => ['equipment', 'detail', id] as const,
  availability: (id: string, from: string, to: string) =>
    ['equipment', 'availability', id, from, to] as const,
};

export function useEquipmentList(filters: EquipmentFilters) {
  return useQuery({
    queryKey: equipmentKeys.list(filters),
    queryFn: () => api.equipment.list(filters),
    staleTime: 30_000,          // 30 Sekunden frisch
    gcTime: 5 * 60 * 1000,     // 5 Minuten im Cache
  });
}

export function useEquipmentDetail(id: string) {
  return useQuery({
    queryKey: equipmentKeys.detail(id),
    queryFn: () => api.equipment.getById(id),
    enabled: !!id,
  });
}

export function useCheckOutEquipment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CheckOutData) => api.equipment.checkOut(data),
    onMutate: async (data) => {
      // Optimistic Update: sofort UI aktualisieren
      await queryClient.cancelQueries({ queryKey: equipmentKeys.detail(data.equipmentId) });
      const previous = queryClient.getQueryData(equipmentKeys.detail(data.equipmentId));
      queryClient.setQueryData(equipmentKeys.detail(data.equipmentId), (old: Equipment) => ({
        ...old, status: 'rented',
      }));
      return { previous };
    },
    onError: (err, data, context) => {
      // Rollback bei Fehler
      queryClient.setQueryData(equipmentKeys.detail(data.equipmentId), context?.previous);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: equipmentKeys.all });
    },
  });
}
```

---

## Kernkomponenten

### AppShell (Haupt-Layout)

```typescript
// src/layouts/AppShell.tsx
export function AppShell() {
  const { sidebarCollapsed } = useUIStore();
  const { unreadCount } = useNotificationStore();

  return (
    <div className={styles.shell} data-sidebar={sidebarCollapsed ? 'collapsed' : 'expanded'}>
      <Sidebar />
      <main className={styles.main}>
        <TopBar notificationCount={unreadCount} />
        <div className={styles.content}>
          <Outlet />  {/* TanStack Router */}
        </div>
      </main>
      <NotificationCenter />
      <CommandPalette />  {/* ⌘K globale Suche */}
    </div>
  );
}
```

### Sidebar

```typescript
// src/layouts/Sidebar.tsx
// Navigationspunkte variieren je nach Rolle des Users
const navigationItems = [
  { icon: LayoutDashboard, label: 'Dashboard', to: '/', permission: null },
  { icon: FolderOpen, label: 'Projekte', to: '/projects', permission: 'project.read' },
  { icon: Package, label: 'Equipment', to: '/equipment', permission: 'equipment.read' },
  { icon: Scan, label: 'Scanner', to: '/scanner', permission: 'scanner.use' },
  { icon: Warehouse, label: 'Lager', to: '/warehouse', permission: 'warehouse.read' },
  { icon: Receipt, label: 'Rechnungen', to: '/invoices', permission: 'invoice.read' },
  { icon: Users, label: 'Personal', to: '/crew', permission: 'crew.read' },
  { icon: Wrench, label: 'Wartung', to: '/maintenance', permission: 'maintenance.read' },
  { icon: Truck, label: 'Transport', to: '/transport', permission: 'equipment.read' },
  { icon: Network, label: 'Federation', to: '/federation', permission: 'federation.read' },
  { icon: BarChart3, label: 'Reports', to: '/reports', permission: 'reports.read' },
  { icon: Settings, label: 'Einstellungen', to: '/settings', permission: 'admin.users' },
];
```

### Dashboard-Komponente (Marcos 30-Sekunden-Überblick)

```typescript
// src/pages/Dashboard/Dashboard.tsx
export function Dashboard() {
  const { user } = useAuthStore();
  const today = new Date();
  const weekEnd = addDays(today, 7);

  const { data: kpis } = useKPIs();
  const { data: upcomingProjects } = useProjects({
    dateFrom: today.toISOString(),
    dateTo: weekEnd.toISOString(),
    status: ['confirmed', 'in_preparation', 'active'],
  });
  const { data: openInvoices } = useOpenPositions({ limit: 5 });
  const { data: maintenanceDue } = useMaintenanceDue({ days: 30 });
  const { data: conflictsCount } = useAvailabilityConflicts();

  return (
    <div className={styles.dashboard}>
      {/* KPI-Karten (oben) */}
      <div className={styles.kpiGrid}>
        <KPICard
          title="Aktive Projekte"
          value={kpis?.activeProjects}
          icon={<FolderOpen />}
          trend={kpis?.projectsTrend}
        />
        <KPICard
          title="Offene Rechnungen"
          value={kpis?.openInvoicesAmount}
          format="currency"
          icon={<Receipt />}
          variant={kpis?.overdueCount > 0 ? 'warning' : 'default'}
        />
        <KPICard
          title="Equipment-Auslastung"
          value={kpis?.utilizationRate}
          format="percent"
          icon={<Package />}
        />
        {conflictsCount > 0 && (
          <KPICard
            title="Doppelbuchungen!"
            value={conflictsCount}
            icon={<AlertTriangle />}
            variant="danger"
            to="/projects?filter=conflicts"
          />
        )}
      </div>

      {/* Diese Woche: Projekte */}
      <Card>
        <CardHeader>Diese Woche</CardHeader>
        <CardBody>
          <ProjectTimelineWidget projects={upcomingProjects} />
        </CardBody>
      </Card>

      {/* Zwei-Spalten-Layout (Desktop) */}
      <div className={styles.twoColumn}>
        <Card>
          <CardHeader>Offene Rechnungen</CardHeader>
          <OpenInvoicesWidget invoices={openInvoices} />
        </Card>
        <Card>
          <CardHeader>Wartung fällig</CardHeader>
          <MaintenanceDueWidget items={maintenanceDue} />
        </Card>
      </div>
    </div>
  );
}
```

### Equipment-Liste (Inventory)

```typescript
// src/pages/Equipment/EquipmentList.tsx
export function EquipmentList() {
  const [filters, setFilters] = useState<EquipmentFilters>({});
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const { data, isLoading, fetchNextPage, hasNextPage } = useEquipmentListInfinite(filters);

  const columns = useMemo(() => [
    columnHelper.accessor('internalNumber', {
      header: 'Nr.',
      cell: info => <span className={styles.mono}>{info.getValue()}</span>,
      size: 80,
    }),
    columnHelper.accessor('name', {
      header: 'Bezeichnung',
      cell: info => (
        <div className={styles.nameCell}>
          {info.row.original.images?.[0] && (
            <img src={info.row.original.images[0]} className={styles.thumbnail} />
          )}
          <div>
            <span className={styles.name}>{info.getValue()}</span>
            <span className={styles.manufacturer}>{info.row.original.manufacturer}</span>
          </div>
        </div>
      ),
    }),
    columnHelper.accessor('status', {
      header: 'Status',
      cell: info => <EquipmentStatusBadge status={info.getValue()} />,
      size: 120,
    }),
    columnHelper.accessor('quantityAvailable', {
      header: 'Verfügbar',
      cell: info => (
        <QuantityDisplay
          available={info.getValue()}
          total={info.row.original.quantityTotal}
        />
      ),
      size: 100,
    }),
    columnHelper.accessor('location', {
      header: 'Lagerplatz',
      cell: info => info.getValue()?.name ?? '—',
    }),
  ], []);

  return (
    <div>
      <PageHeader
        title="Equipment"
        actions={
          <>
            <Button variant="outline" onClick={() => navigate('/equipment/import')}>
              Import
            </Button>
            <Button onClick={() => navigate('/equipment/new')}>
              <Plus size={16} /> Neues Equipment
            </Button>
          </>
        }
      />

      <FilterBar
        filters={equipmentFilterConfig}
        value={filters}
        onChange={setFilters}
      />

      {selectedIds.length > 0 && (
        <BulkActionBar
          selectedCount={selectedIds.length}
          actions={[
            { label: 'Labels drucken', icon: <Printer />, onClick: () => printLabels(selectedIds) },
            { label: 'Lagerplatz ändern', icon: <MapPin />, onClick: () => setShowMoveModal(true) },
          ]}
        />
      )}

      <DataTable
        data={data?.pages.flatMap(p => p.items) ?? []}
        columns={columns}
        isLoading={isLoading}
        onRowClick={(row) => navigate(`/equipment/${row.original.id}`)}
        rowSelection={{ value: selectedIds, onChange: setSelectedIds }}
        virtualizing={{ estimateSize: () => 60, overscan: 10 }}
        infiniteScroll={{ fetchNextPage, hasNextPage }}
      />
    </div>
  );
}
```

### Scanner-Interface (Optimiert für Lisa / Zebra TC21)

```typescript
// src/pages/Scanner/ScannerInterface.tsx
// Speziell für große Buttons, Offline-Modus, Zebra-Hardware-Scanner

export function ScannerInterface() {
  const { scanMode, currentProjectId } = useScannerStore();
  const [lastResult, setLastResult] = useState<ScanResult | null>(null);
  const cameraRef = useRef<HTMLVideoElement>(null);
  const audioFeedback = useAudioFeedback();

  // ZXing Barcode-Scanner (Kamera)
  const { startScan, stopScan } = useBarcodeScanner(cameraRef, {
    onSuccess: (barcode) => handleScan(barcode),
    formats: ['QR_CODE', 'CODE_128', 'CODE_39', 'EAN_13'],
  });

  // Zebra: Hardware-Scanner-Taste (DataWedge Intent)
  useZebraHardwareScanner({
    onScan: (barcode) => handleScan(barcode),
  });

  const handleScan = async (barcode: string) => {
    const result = await processScan(barcode, scanMode, currentProjectId);
    setLastResult(result);

    // Visuelles + akustisches Feedback
    if (result.status === 'ok') {
      audioFeedback.play('beep_ok');
    } else if (result.status === 'error') {
      audioFeedback.play('beep_error');
    } else {
      audioFeedback.play('beep_warn');
    }
  };

  return (
    <div className={clsx(styles.scanner, styles[`scanner--${lastResult?.status ?? 'idle'}`])}>

      {/* Feedback-Bereich (oben, sehr groß) */}
      <div className={styles.feedback}>
        {lastResult ? (
          <ScanFeedbackDisplay result={lastResult} />
        ) : (
          <div className={styles.waitingForScan}>
            <ScanIcon size={64} />
            <span>Bereit zum Scannen</span>
          </div>
        )}
      </div>

      {/* Kamera-Vorschau */}
      <div className={styles.cameraWrapper}>
        <video ref={cameraRef} className={styles.camera} autoPlay playsInline />
        <ScannerOverlay />
      </div>

      {/* Scan-Modus-Auswahl (große Buttons) */}
      <div className={styles.modeSelector}>
        <ScanModeButton
          mode="checkout"
          label="Ausgabe"
          icon={<ArrowUpRight size={32} />}
          active={scanMode === 'checkout'}
        />
        <ScanModeButton
          mode="checkin"
          label="Rückgabe"
          icon={<ArrowDownLeft size={32} />}
          active={scanMode === 'checkin'}
        />
        <ScanModeButton
          mode="locate"
          label="Suchen"
          icon={<MapPin size={32} />}
          active={scanMode === 'locate'}
        />
      </div>

      {/* Offline-Indikator */}
      <OfflineIndicator />
    </div>
  );
}

// Scan-Feedback: Vollbild-Farbe + Text (für schnelles Erkennen)
function ScanFeedbackDisplay({ result }: { result: ScanResult }) {
  return (
    <div className={clsx(styles.feedbackContent, styles[result.status])}>
      <div className={styles.icon}>
        {result.status === 'ok' && <CheckCircle size={80} />}
        {result.status === 'error' && <XCircle size={80} />}
        {result.status === 'warning' && <AlertTriangle size={80} />}
      </div>
      <h2 className={styles.equipmentName}>{result.equipmentName}</h2>
      <p className={styles.message}>{result.message}</p>
      {result.location && (
        <p className={styles.location}>
          <MapPin size={20} /> {result.location}
        </p>
      )}
    </div>
  );
}
```

### Packliste-Komponente (Interaktiv)

```typescript
// src/pages/Projects/PackingList.tsx
export function PackingList({ projectId }: { projectId: string }) {
  const { data: project } = useProject(projectId);
  const { data: packingList } = useProjectPackingList(projectId);
  const updateItem = useUpdateProjectEquipment();

  // Sortierung nach Lagerplatz (routenoptimiert für Lisa)
  const sortedItems = useMemo(() =>
    sortByLocation(packingList?.items ?? []),
    [packingList]
  );

  return (
    <div>
      <div className={styles.header}>
        <h2>{project?.name}</h2>
        <PackingListProgress items={packingList?.items ?? []} />
        <div className={styles.actions}>
          <Button variant="outline" onClick={() => printPackingList(projectId)}>
            <Printer size={16} /> Drucken
          </Button>
          <Button variant="outline" onClick={() => exportPackingList(projectId)}>
            <Download size={16} /> Export
          </Button>
        </div>
      </div>

      {/* Abschnitte nach Kategorie */}
      {packingList?.sections.map(section => (
        <PackingListSection key={section.id} section={section}>
          {sortedItems
            .filter(item => item.sectionId === section.id)
            .map(item => (
              <PackingListItem
                key={item.id}
                item={item}
                onStatusChange={(status) =>
                  updateItem.mutate({ itemId: item.id, status })
                }
              />
            ))}
        </PackingListSection>
      ))}

      <AvailabilityConflicts projectId={projectId} />
    </div>
  );
}

function PackingListItem({ item, onStatusChange }: PackingListItemProps) {
  return (
    <div className={clsx(styles.item, styles[`item--${item.status}`])}>
      <Checkbox
        checked={item.status === 'checked_out'}
        onCheckedChange={(checked) =>
          onStatusChange(checked ? 'checked_out' : 'planned')
        }
      />
      <div className={styles.itemInfo}>
        <span className={styles.name}>{item.equipmentName}</span>
        <span className={styles.location}>{item.location?.name}</span>
      </div>
      <span className={styles.quantity}>× {item.quantity}</span>
      <EquipmentStatusBadge status={item.equipmentStatus} />
    </div>
  );
}
```

### Rechnungs-Formular (Für Thomas)

```typescript
// src/pages/Invoices/InvoiceForm.tsx
// Tastatur-Navigation, klare Abläufe, keine Überraschungen

export function InvoiceForm({ invoiceId }: { invoiceId?: string }) {
  const isEdit = !!invoiceId;
  const { data: invoice } = useInvoice(invoiceId, { enabled: isEdit });

  const form = useForm<InvoiceFormData>({
    resolver: zodResolver(invoiceSchema),
    defaultValues: invoice ?? defaultInvoiceValues,
  });

  const {
    fields: lineItems,
    append: addLineItem,
    remove: removeLineItem,
    move: moveLineItem,
  } = useFieldArray({ control: form.control, name: 'items' });

  const subtotal = useWatch({ control: form.control, name: 'items' })
    .reduce((sum, item) => sum + item.quantity * item.unitPrice * (1 - item.discountPct / 100), 0);
  const taxAmount = subtotal * (form.watch('taxRate') / 100);
  const total = subtotal + taxAmount;

  return (
    <form onSubmit={form.handleSubmit(onSubmit)}>
      <div className={styles.grid}>
        {/* Kopf-Daten */}
        <FormField name="customerId" label="Kunde" required>
          <CustomerSelect control={form.control} name="customerId" />
        </FormField>
        <FormField name="projectId" label="Projekt">
          <ProjectSelect control={form.control} name="projectId" />
        </FormField>
        <FormField name="invoiceDate" label="Rechnungsdatum" required>
          <DatePicker control={form.control} name="invoiceDate" />
        </FormField>
        <FormField name="dueDate" label="Fälligkeitsdatum" required>
          <DatePicker control={form.control} name="dueDate" />
        </FormField>
      </div>

      {/* Positionen (Drag & Drop für Sortierung) */}
      <DndContext
        sensors={sensors}
        onDragEnd={(event) => moveLineItem(oldIndex, newIndex)}
      >
        <SortableContext items={lineItems.map(f => f.id)}>
          <table className={styles.lineItems}>
            <thead>
              <tr>
                <th>Pos.</th>
                <th>Beschreibung</th>
                <th>Menge</th>
                <th>Einheit</th>
                <th>Einzelpreis</th>
                <th>Rabatt %</th>
                <th>Gesamt</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {lineItems.map((field, index) => (
                <InvoiceLineItem
                  key={field.id}
                  index={index}
                  control={form.control}
                  onRemove={() => removeLineItem(index)}
                />
              ))}
            </tbody>
          </table>
        </SortableContext>
      </DndContext>

      <Button type="button" variant="ghost" onClick={() => addLineItem(defaultLineItem)}>
        <Plus size={16} /> Position hinzufügen
      </Button>

      {/* Summenbereich */}
      <div className={styles.totals}>
        <TotalsRow label="Zwischensumme" value={subtotal} />
        <TotalsRow label={`MwSt. ${form.watch('taxRate')}%`} value={taxAmount} />
        <TotalsRow label="Gesamt" value={total} bold />
      </div>

      <div className={styles.actions}>
        <Button type="button" variant="outline" onClick={() => navigate(-1)}>
          Abbrechen
        </Button>
        <Button type="submit" variant="outline">
          Entwurf speichern
        </Button>
        <Button type="button" onClick={() => handleSendInvoice()}>
          <Send size={16} /> Rechnung versenden
        </Button>
      </div>
    </form>
  );
}
```

### Real-time Notifications (WebSocket)

```typescript
// src/hooks/useWebSocket.ts
export function useWebSocket() {
  const { user } = useAuthStore();
  const { addNotification } = useNotificationStore();
  const queryClient = useQueryClient();
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!user) return;

    const token = getAccessToken();
    const ws = new WebSocket(
      `${import.meta.env.VITE_WS_URL}/ws/notifications?token=${token}`
    );

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data) as WSMessage;

      switch (message.type) {
        case 'notification':
          // In-App Notification anzeigen
          addNotification(message.payload);
          toast(message.payload.title, { description: message.payload.body });
          break;

        case 'equipment_updated':
          // React Query Cache invalidieren → automatischer Refetch
          queryClient.invalidateQueries({
            queryKey: equipmentKeys.detail(message.payload.equipmentId)
          });
          break;

        case 'project_conflict':
          // Doppelbuchungs-Warnung sofort anzeigen
          toast.warning('Doppelbuchung erkannt!', {
            description: message.payload.message,
            action: { label: 'Ansehen', onClick: () => navigate('/projects') },
          });
          break;
      }
    };

    ws.onclose = () => {
      // Reconnect nach 3 Sekunden
      setTimeout(() => setupWebSocket(), 3000);
    };

    wsRef.current = ws;
    return () => ws.close();
  }, [user]);
}
```

---

## Responsive Design & Mobile-Optimierung

### Breakpoints

```scss
// Breakpoints
$breakpoint-sm: 640px;    // Kleines Handy (Kevin iPhone)
$breakpoint-md: 768px;    // Zebra TC21 / Tablet
$breakpoint-lg: 1024px;   // Desktop klein
$breakpoint-xl: 1280px;   // Desktop standard (Marco)
$breakpoint-2xl: 1536px;  // Desktop groß (Thomas 27")

// Mixins
@mixin mobile { @media (max-width: #{$breakpoint-md - 1}) { @content; } }
@mixin tablet { @media (min-width: $breakpoint-md) and (max-width: #{$breakpoint-lg - 1}) { @content; } }
@mixin desktop { @media (min-width: $breakpoint-lg) { @content; } }
```

### Mobile-spezifisches Layout

```scss
// Dashboard: 1 Spalte auf Mobile, 2-3 auf Desktop
.kpiGrid {
  display: grid;
  grid-template-columns: 1fr;  // Mobile: 1 Spalte

  @include tablet { grid-template-columns: repeat(2, 1fr); }
  @include desktop { grid-template-columns: repeat(4, 1fr); }
}

// Scanner-UI: Vollbild auf Mobile
.scanner {
  @include mobile {
    min-height: 100dvh;           // Dynamic viewport height
    padding-bottom: env(safe-area-inset-bottom);  // iOS notch
  }
}

// Tabellen: Horizontal scrollbar auf Mobile
.tableWrapper {
  @include mobile {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }
}

// Buttons: Mindestgröße für Touch-Targets
.button {
  min-height: 44px;   // Apple-Empfehlung: 44pt
  min-width: 44px;

  @include mobile {
    min-height: 52px;  // Noch größer auf Mobile (Arbeitshandschuhe!)
    padding: var(--space-4) var(--space-6);
  }
}
```

### PWA-Konfiguration

```typescript
// vite.config.ts
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
  plugins: [
    VitePWA({
      registerType: 'autoUpdate',
      workbox: {
        // Offline-fähige Ressourcen
        runtimeCaching: [
          {
            urlPattern: /^\/api\/equipment\/.+\/info$/,
            handler: 'CacheFirst',
            options: {
              cacheName: 'equipment-info',
              expiration: { maxEntries: 1000, maxAgeSeconds: 3600 },
            },
          },
          {
            urlPattern: /^\/api\/projects\/.+\/packing-list$/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'packing-lists',
              expiration: { maxAgeSeconds: 900 },  // 15 Minuten
            },
          },
        ],
      },
      manifest: {
        name: 'MyRMS',
        short_name: 'MyRMS',
        description: 'Rental Management System',
        theme_color: '#2563eb',
        background_color: '#ffffff',
        display: 'standalone',
        orientation: 'portrait-primary',
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any maskable' },
        ],
      },
    }),
  ],
});
```

---

## Internationalisierung (DE/EN)

```typescript
// src/i18n/de.json (Auszug)
{
  "equipment": {
    "status": {
      "available": "Verfügbar",
      "rented": "Verliehen",
      "defect": "Defekt",
      "maintenance": "Wartung",
      "retired": "Ausgemustert"
    },
    "actions": {
      "checkOut": "Ausgabe",
      "checkIn": "Rückgabe",
      "edit": "Bearbeiten",
      "printLabel": "Label drucken"
    }
  },
  "scanner": {
    "feedback": {
      "ok": "OK — {{name}}",
      "alreadyCheckedOut": "BEREITS AUSGEGEBEN",
      "defect": "ACHTUNG: Gerät defekt!",
      "wrongJob": "Falscher Job: gehört zu {{projectName}}"
    }
  }
}

// src/i18n/en.json (für englischsprachige Freelancer)
{
  "equipment": {
    "status": {
      "available": "Available",
      "rented": "Rented Out",
      // ...
    }
  }
}
```

---

## Barrierefreiheit (Accessibility)

Alle Komponenten folgen WCAG 2.1 AA:

- **Tastatur-Navigation**: Alle interaktiven Elemente per Tab erreichbar
- **ARIA-Labels**: Korrekte Rollen und Beschriftungen für Screen-Reader
- **Fokus-Management**: Focus Trap in Modals, Rückkehr nach Schließen
- **Farbkontrast**: Mindest-Kontrastverhältnis 4.5:1 (Text) / 3:1 (große Schrift)
- **Animationen**: `prefers-reduced-motion` wird respektiert
- **Formulare**: Fehlermeldungen mit `aria-describedby` verknüpft

```typescript
// Beispiel: AccessibleDataTable
<table
  role="grid"
  aria-label={`${title} — ${data.length} Einträge`}
  aria-rowcount={totalCount}
>
  <caption className={visuallyHidden}>{description}</caption>
  {/* ... */}
</table>
```

---

## Command Palette (⌘K globale Suche)

```typescript
// src/components/CommandPalette/CommandPalette.tsx
// Globale Suche: ⌘K öffnet, sofortiger Zugriff auf alles

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();

  // Keyboard-Shortcut
  useHotkeys('mod+k', () => setOpen(true));

  const commands: Command[] = [
    // Navigation
    { id: 'nav-dashboard', label: 'Dashboard', icon: LayoutDashboard, action: () => navigate('/') },
    { id: 'nav-new-project', label: 'Neues Projekt', icon: Plus, action: () => navigate('/projects/new') },
    // ...
  ];

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Suchen... (Equipment, Projekte, Kunden)" />
      <CommandList>
        <SearchResults />
        <CommandGroup heading="Navigation">
          {commands.map(cmd => (
            <CommandItem key={cmd.id} onSelect={cmd.action}>
              <cmd.icon size={16} />
              {cmd.label}
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  );
}
```

---

## Performance-Optimierungen

### Code-Splitting und Lazy Loading

```typescript
// src/router.tsx — Alle Seiten werden lazy geladen
const DashboardPage = lazy(() => import('./pages/Dashboard'));
const EquipmentListPage = lazy(() => import('./pages/Equipment/EquipmentList'));
const ScannerPage = lazy(() => import('./pages/Scanner/ScannerInterface'));
// ... alle anderen Seiten

// Route mit Suspense
<Route
  path="/equipment"
  element={
    <Suspense fallback={<PageSkeleton />}>
      <EquipmentListPage />
    </Suspense>
  }
/>
```

### Virtualisierung für große Listen

```typescript
// Für Equipment-Listen mit 1000+ Einträgen: TanStack Virtual
import { useVirtualizer } from '@tanstack/react-virtual';

const rowVirtualizer = useVirtualizer({
  count: data.length,
  getScrollElement: () => parentRef.current,
  estimateSize: () => 60,      // Geschätzte Zeilenhöhe
  overscan: 5,                  // Vorher gerenderte Zeilen
});
```

### Image-Optimierung

```typescript
// Lazy Loading + WebP für Equipment-Bilder
<img
  src={imageUrl}
  loading="lazy"
  decoding="async"
  srcSet={`${imageUrl}?w=400 400w, ${imageUrl}?w=800 800w`}
  sizes="(max-width: 768px) 100vw, 400px"
  alt={equipmentName}
/>
```

---

## Testing-Strategie

### Unit Tests (Vitest + React Testing Library)

```typescript
// Beispiel: Equipment-Status-Badge
describe('EquipmentStatusBadge', () => {
  it('zeigt korrekten Text für Status "available"', () => {
    render(<EquipmentStatusBadge status="available" />);
    expect(screen.getByText('Verfügbar')).toBeInTheDocument();
  });

  it('hat korrekten Stil für Status "defect"', () => {
    render(<EquipmentStatusBadge status="defect" />);
    const badge = screen.getByRole('status');
    expect(badge).toHaveClass('badge--danger');
  });
});
```

### E2E-Tests (Playwright)

```typescript
// Beispiel: Kompletterfahren Check-Out
test('Equipment Check-Out Workflow', async ({ page }) => {
  await page.goto('/scanner');
  await page.getByRole('button', { name: 'Ausgabe' }).click();

  // Barcode-Eingabe simulieren (kein physischer Scanner in Tests)
  await page.getByTestId('barcode-input').fill('LT-0042');
  await page.keyboard.press('Enter');

  // Feedback prüfen
  await expect(page.getByText('OK')).toBeVisible();
  await expect(page.getByTestId('feedback-color')).toHaveClass('scanner--ok');
});
```
