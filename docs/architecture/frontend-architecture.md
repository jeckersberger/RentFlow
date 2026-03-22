# Frontend-Architektur: Lagerverwaltungs- und Rechnungssoftware

**Tech Stack:** React + TypeScript + Sass/SCSS | Vite | i18n (DE/EN) von Anfang an

---

## 1. Projekt-Struktur

```
src/
├── App.tsx                          # Root-Komponente mit Router
├── main.tsx                         # Entry Point
├── vite-env.d.ts
├── index.scss                       # Globale Styles
│
├── components/                      # Shared & Feature-Components
│   ├── atoms/                       # Primitive UI-Elemente
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   ├── Badge.tsx
│   │   ├── Avatar.tsx
│   │   ├── Icon.tsx
│   │   └── _atoms.scss
│   │
│   ├── molecules/                   # Kombinationen von Atoms
│   │   ├── FormField.tsx            # Label + Input + Error
│   │   ├── SearchBar.tsx
│   │   ├── Pagination.tsx
│   │   ├── BreadcrumbNav.tsx
│   │   ├── ConfirmDialog.tsx
│   │   ├── Toast.tsx
│   │   └── _molecules.scss
│   │
│   ├── organisms/                   # Feature-Komponenten
│   │   ├── Header.tsx               # Navbar mit Rolle-Anpassung
│   │   ├── Sidebar.tsx              # Adaptive Navigation
│   │   ├── RoleGuard.tsx            # Rollen-Wrapper
│   │   ├── OfflineIndicator.tsx
│   │   └── _organisms.scss
│   │
│   └── layouts/                     # Layout-Container
│       ├── AppShell.tsx             # Haupt-Layout mit Header/Sidebar
│       ├── DashboardLayout.tsx
│       ├── ScannerLayout.tsx        # Zebra-optimiert
│       ├── MobileLayout.tsx
│       └── _layouts.scss
│
├── features/                        # Feature-Module pro Rolle/Domain
│   ├── dashboard/
│   │   ├── pages/
│   │   │   └── DashboardPage.tsx
│   │   ├── components/
│   │   │   ├── KPICard.tsx
│   │   │   ├── RecentOrders.tsx
│   │   │   └── EquipmentStatus.tsx
│   │   ├── hooks/
│   │   │   └── useDashboardData.ts
│   │   ├── types.ts
│   │   ├── api.ts                   # Feature-spezifische API-Calls
│   │   └── dashboard.scss
│   │
│   ├── warehouse/                   # Scanner-Modul (Lisa)
│   │   ├── pages/
│   │   │   ├── ScannerPage.tsx
│   │   │   ├── InventoryPage.tsx
│   │   │   └── StockMovementPage.tsx
│   │   ├── components/
│   │   │   ├── ScanInput.tsx        # Scanner-Input mit Event-Handling
│   │   │   ├── ScanResult.tsx       # Feedback-Komponente
│   │   │   ├── InventoryTable.tsx   # Virtual List
│   │   │   └── LocationGrid.tsx
│   │   ├── hooks/
│   │   │   ├── useScanner.ts        # Scanner-Integration
│   │   │   ├── useOfflineStorage.ts # IndexedDB
│   │   │   └── useScanSync.ts       # Offline-Sync
│   │   ├── offline/
│   │   │   ├── db.ts                # IndexedDB Schema
│   │   │   └── syncQueue.ts
│   │   ├── types.ts
│   │   ├── api.ts
│   │   └── warehouse.scss
│   │
│   ├── quotes/                      # Angebotsmodul (Marco)
│   │   ├── pages/
│   │   │   ├── QuotesListPage.tsx
│   │   │   ├── CreateQuotePage.tsx
│   │   │   └── QuoteDetailPage.tsx
│   │   ├── components/
│   │   │   ├── QuoteForm.tsx
│   │   │   ├── EquipmentSelector.tsx
│   │   │   └── PricingCalculator.tsx
│   │   ├── hooks/
│   │   │   └── useQuoteForm.ts
│   │   ├── types.ts
│   │   ├── api.ts
│   │   └── quotes.scss
│   │
│   ├── accounting/                  # Finanzen (Thomas)
│   │   ├── pages/
│   │   │   ├── InvoiceListPage.tsx
│   │   │   ├── InvoiceDetailPage.tsx
│   │   │   └── ReportPage.tsx
│   │   ├── components/
│   │   │   ├── InvoiceTable.tsx     # Excel-ähnlich + Tastatur-Nav
│   │   │   ├── LineItemEditor.tsx   # Inline-Editing
│   │   │   └── PrintPreview.tsx
│   │   ├── hooks/
│   │   │   └── useInvoiceForm.ts
│   │   ├── types.ts
│   │   ├── api.ts
│   │   └── accounting.scss
│   │
│   └── freelancer/                  # Mobil (Kevin)
│       ├── pages/
│       │   ├── TasksPage.tsx
│       │   ├── PacklistPage.tsx
│       │   └── PhotoLogPage.tsx
│       ├── components/
│       │   ├── TaskCard.tsx
│       │   ├── PacklistCheckbox.tsx
│       │   └── CameraCapture.tsx
│       ├── hooks/
│       │   └── useCameraScanner.ts
│       ├── types.ts
│       ├── api.ts
│       └── freelancer.scss
│
├── hooks/                           # Global Custom Hooks
│   ├── useAuth.ts
│   ├── useRole.ts
│   ├── useDeviceType.ts             # Desktop/Mobile/Tablet/Zebra-Detection
│   ├── useResponsive.ts             # Breakpoint-Observer
│   ├── useNetworkStatus.ts
│   ├── useTheme.ts
│   └── useLocalStorage.ts
│
├── stores/                          # State Management (Zustand)
│   ├── authStore.ts                 # Auth-State, User, Role
│   ├── uiStore.ts                   # UI: Modal, Sidebar, Theme
│   ├── offlineStore.ts              # Offline-Mode + Queue
│   ├── notificationStore.ts         # Toast-Zentrale
│   └── scanStore.ts                 # Scanner-State (aktive Scans, Ergebnisse)
│
├── api/                             # API-Integration (TanStack Query)
│   ├── client.ts                    # Axios-Instance mit Token
│   ├── queries.ts                   # useQuery-Hooks
│   ├── mutations.ts                 # useMutation-Hooks
│   ├── queryClient.ts               # QueryClient Setup
│   └── endpoints.ts                 # Konstanten für API-Routes
│
├── lib/                             # Utilities & Helper
│   ├── cn.ts                        # classname-Merger
│   ├── formatters.ts                # Datum, Währung, Mengeneinheiten
│   ├── validators.ts                # Form-Validators
│   ├── storage.ts                   # LocalStorage-Wrapper
│   ├── logger.ts                    # Console-Logger
│   └── scanner.ts                   # Scanner-Utils (Barcode-Parser)
│
├── i18n/                            # Internationalisierung
│   ├── i18n.ts                      # i18next Config
│   ├── de/                          # Deutsche Übersetzungen
│   │   ├── common.json
│   │   ├── dashboard.json
│   │   ├── warehouse.json
│   │   ├── quotes.json
│   │   ├── accounting.json
│   │   ├── freelancer.json
│   │   └── errors.json
│   └── en/                          # English Translations
│       ├── common.json
│       ├── dashboard.json
│       ├── warehouse.json
│       ├── quotes.json
│       ├── accounting.json
│       ├── freelancer.json
│       └── errors.json
│
├── styles/                          # Globale & Theme-Styles
│   ├── variables.scss               # CSS Custom Properties
│   ├── breakpoints.scss             # Responsive Breakpoints
│   ├── mixins.scss                  # SCSS-Funktionen
│   ├── reset.scss                   # Browser-Reset
│   ├── typography.scss
│   ├── colors.scss
│   ├── themes/
│   │   ├── light.scss               # Light-Mode Variablen
│   │   ├── dark.scss                # Dark-Mode Variablen
│   │   └── highcontrast.scss        # High-Contrast für Lager
│   └── responsive/
│       ├── desktop.scss
│       ├── tablet.scss
│       ├── mobile.scss
│       └── scanner.scss             # Zebra TC21 Spezifika
│
├── types/                           # Globale TypeScript-Typen
│   ├── common.ts
│   ├── entities.ts                  # Equipment, Order, Invoice, etc.
│   ├── api.ts                       # API-Response-Types
│   └── auth.ts
│
├── workers/                         # Web Workers
│   ├── offline-sync.worker.ts       # IndexedDB-Sync im Hintergrund
│   └── data-processor.worker.ts     # Datenverarbeitung (Tabellen-Sorting)
│
├── services/                        # Business-Logik
│   ├── offlineService.ts            # Offline-Mode-Manager
│   ├── syncService.ts               # Sync-Koordination
│   ├── authService.ts               # Auth-Logik
│   ├── reportService.ts             # Report-Generation (Thomas)
│   └── priceService.ts              # Kalkulation (Marco)
│
└── routes/                          # Route-Definitionen
    ├── index.tsx                    # Rollen-basierte Route-Guards
    └── protectedRoutes.ts           # Protected-Route-Wrapper

public/
├── manifest.json                    # PWA Manifest
├── icons/
│   ├── icon-192.png
│   ├── icon-512.png
│   └── favicon.ico
└── offline.html                     # Offline-Fallback

tests/
├── unit/
│   ├── components/
│   ├── hooks/
│   └── utils/
├── integration/
│   ├── features/
│   └── flows/
└── e2e/
    └── specs/

vite.config.ts
tsconfig.json
package.json
```

**Ordner-Erklärungen:**

- **`/components`**: Wiederverwendbare UI-Komponenten nach Atomic Design. Atoms sind die kleinsten, völlig datenlos. Molecules kombinieren Atoms. Organisms sind komplexe Features. Layouts definieren die Page-Struktur.

- **`/features`**: Feature-Module nach Business-Domain. Jedes Modul ist selbstständig mit Pages, Components, Hooks, Types, API. Starke Kohäsion, lose Kopplung zwischen Features.

- **`/hooks`**: Custom React Hooks für Cross-cutting Concerns (Auth, Responsive, Network).

- **`/stores`**: Zustand für globalen UI-State (nicht für Daten! Das macht React Query).

- **`/api`**: Zentrale API-Integration mit TanStack Query. Queries für GETs, Mutations für POST/PUT/DELETE.

- **`/lib`**: Reine Utilities ohne React-Dependencies.

- **`/i18n`**: Namespacing pro Modul + Common. Key-Pattern: `feature:key`.

- **`/styles`**: CSS Custom Properties als Single Source of Truth. Theme-Variablen für Light/Dark/HighContrast.

- **`/workers`**: Web Workers für Hintergrund-Operationen (IndexedDB-Sync, Datenverarbeitung).

- **`/services`**: Business-Logik-Layer über API. Koordiniert mehrere API-Calls, Offline-Strategie, etc.

---

## 2. Komponenten-Hierarchie

### App-Shell

```tsx
// App.tsx
<QueryClientProvider>
  <AuthProvider>
    <ThemeProvider>
      <AppRouter />
    </ThemeProvider>
  </AuthProvider>
</QueryClientProvider>
```

### Layout-Baum nach Rolle

**Desktop-Rolle (Marco/Thomas):**
```
AppShell
├── Header (Navbar mit Logo, User-Menu, Search)
├── Sidebar (Navigation angepasst an Rolle)
└── MainContent
    └── Page (z.B. DashboardPage, InvoiceListPage)
```

**Zebra-Handheld (Lisa):**
```
ScannerLayout (Vollbild, keine Header/Sidebar)
├── ScanInput (großer Input-Bereich)
├── ScanResult (Farb-Feedback: grün/rot/gelb)
└── QuickActions (2-3 Buttons max)
```

**Mobile (Kevin):**
```
MobileLayout
├── Header (Compact, Back-Button, Menü-Icon)
├── MainContent (Full-Width)
└── BottomNav (Icon-Tabs für Hauptfunktionen)
```

### Shared Component Library

```
atoms/
  Button (variant: primary|secondary|danger|outline, size: sm|md|lg, loading, disabled)
  Input (type, placeholder, error, disabled, autoFocus)
  Badge (variant: success|warning|error|info, size: sm|md)
  Avatar (src, initials, size)
  Icon (name, size, color)
  Spinner (size, color)

molecules/
  FormField (label, required, error, hint, children)
  SearchBar (placeholder, onSearch, debounce)
  Pagination (current, total, pageSize, onChange)
  BreadcrumbNav (items: {label, href}[])
  ConfirmDialog (title, message, onConfirm, onCancel, isDangerous)
  Toast (variant, message, action, autoClose)
  Tooltip (content, position: top|bottom|left|right)
  Tabs (tabs: {label, id}[], onChange)

organisms/
  Header (sticky, role-aware)
  Sidebar (adaptive: desktop=wide, mobile=drawer, zebra=none)
  RoleGuard (allowedRoles, children, fallback)
  OfflineIndicator (show when navigator.onLine === false)
  UserMenu (avatar, dropdown with Logout)
```

**Varianten-Pattern für Button:**
```tsx
<Button variant="primary" size="lg" loading={isLoading}>
  {t('common:save')}
</Button>
```

---

## 3. Routing

### Route-Struktur mit Rollen-Guards

```tsx
// routes/index.tsx
import { createBrowserRouter } from 'react-router-dom';
import { RoleGuard } from '../components/organisms/RoleGuard';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppShell />,
    errorElement: <ErrorPage />,
    children: [
      // PUBLIC
      { path: 'login', element: <LoginPage /> },
      { path: 'register', element: <RegisterPage /> },

      // DASHBOARD (alle Rollen)
      {
        path: 'dashboard',
        element: <RoleGuard allowedRoles={['geschaeftsführung', 'buchhaltung', 'lager', 'freelancer']} />,
        children: [
          { index: true, element: <DashboardPage /> },
        ],
      },

      // WAREHOUSE (Lager-Rolle)
      {
        path: 'warehouse',
        element: <RoleGuard allowedRoles={['lager']} />,
        children: [
          { index: true, element: <ScannerPage /> },
          { path: 'inventory', element: <InventoryPage /> },
          { path: 'stock/:id', element: <StockDetailPage /> },
          { path: 'movements', element: <StockMovementPage /> },
        ],
      },

      // QUOTES (Geschäftsführung)
      {
        path: 'quotes',
        element: <RoleGuard allowedRoles={['geschaeftsführung']} />,
        children: [
          { index: true, element: <QuotesListPage /> },
          { path: 'new', element: <CreateQuotePage /> },
          { path: ':id', element: <QuoteDetailPage /> },
          { path: ':id/edit', element: <EditQuotePage /> },
        ],
      },

      // ACCOUNTING (Buchhaltung)
      {
        path: 'accounting',
        element: <RoleGuard allowedRoles={['buchhaltung']} />,
        children: [
          { path: 'invoices', element: <InvoiceListPage /> },
          { path: 'invoices/:id', element: <InvoiceDetailPage /> },
          { path: 'invoices/:id/print', element: <PrintPreviewPage /> },
          { path: 'reports', element: <ReportPage /> },
          { path: 'dashboard', element: <AccountingDashboardPage /> },
        ],
      },

      // FREELANCER (Mobile-optimiert)
      {
        path: 'tasks',
        element: <RoleGuard allowedRoles={['freelancer']} />,
        children: [
          { index: true, element: <TasksPage /> },
          { path: ':id', element: <TaskDetailPage /> },
          { path: ':id/packlist', element: <PacklistPage /> },
          { path: ':id/photos', element: <PhotoLogPage /> },
        ],
      },

      // SETTINGS (alle Rollen)
      {
        path: 'settings',
        children: [
          { path: 'profile', element: <ProfilePage /> },
          { path: 'preferences', element: <PreferencesPage /> },
          { path: 'notifications', element: <NotificationsPage /> },
        ],
      },
    ],
  },

  // FALLBACK
  { path: '*', element: <NotFoundPage /> },
]);
```

**RoleGuard Komponente:**
```tsx
// components/organisms/RoleGuard.tsx
export const RoleGuard: React.FC<RoleGuardProps> = ({
  allowedRoles,
  children,
  fallback = <AccessDeniedPage />,
}) => {
  const { role } = useAuth();

  if (!allowedRoles.includes(role)) {
    return fallback;
  }

  return children;
};
```

**Nested Routes-Pattern:**
- Jeder Feature-Bereich hat eigene Nested-Routes
- Layout-Wrapper innerhalb der AppShell
- Lazy Loading per Route mit React.lazy()

---

## 4. State Management

### Architektur-Entscheidung

```
┌─────────────────────────────────────────────┐
│  REMOTE STATE (Server-Truth)                │
│  - TanStack Query (React Query)             │
│  - Cache: GET Requests                      │
│  - Invalidation auf Mutations               │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  LOCAL STATE                                │
│  - Zustand: UI-State nur                    │
│  - Context API: Auth                        │
│  - React Hook Form: Form-State              │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  OFFLINE STATE (IndexedDB)                  │
│  - useSyncExternalStore für Sync            │
│  - Sync-Queue in localStorage               │
└─────────────────────────────────────────────┘
```

### TanStack Query (Server State)

```tsx
// api/queries.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

// GET: Equipment List
export const useEquipmentList = (filters?: EquipmentFilters) => {
  return useQuery({
    queryKey: ['equipment', filters],
    queryFn: () => client.get('/api/equipment', { params: filters }),
    staleTime: 5 * 60 * 1000, // 5 Min
    gcTime: 10 * 60 * 1000,    // 10 Min (vormals cacheTime)
    enabled: !!filters, // bedingte Queries
  });
};

// POST: Create Invoice
export const useCreateInvoice = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateInvoiceRequest) =>
      client.post('/api/invoices', data),
    onSuccess: () => {
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
    onError: (error) => {
      notificationStore.addError(error.message);
    },
  });
};
```

### Zustand (UI State)

```tsx
// stores/uiStore.ts
import { create } from 'zustand';

interface UIState {
  sidebarOpen: boolean;
  toggleSidebar: () => void;

  theme: 'light' | 'dark' | 'highcontrast';
  setTheme: (theme: Theme) => void;

  modals: Record<string, boolean>;
  openModal: (id: string) => void;
  closeModal: (id: string) => void;
}

export const useUIStore = create<UIState>((set) => ({
  sidebarOpen: true,
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),

  theme: 'light',
  setTheme: (theme) => set({ theme }),

  modals: {},
  openModal: (id) => set((s) => ({ modals: { ...s.modals, [id]: true } })),
  closeModal: (id) => set((s) => ({ modals: { ...s.modals, [id]: false } })),
}));
```

### Auth Context

```tsx
// contexts/AuthContext.tsx
interface AuthContextType {
  user: User | null;
  role: Role;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    // Validiere Token beim Mount
    const token = localStorage.getItem('token');
    if (token) {
      validateToken(token).then(setUser);
    }
  }, []);

  return (
    <AuthContext.Provider value={{
      user,
      role: user?.role || 'freelancer',
      isAuthenticated: !!user,
      login: async (email, password) => {
        const { token, user } = await client.post('/auth/login', { email, password });
        localStorage.setItem('token', token);
        setUser(user);
      },
      logout: () => {
        localStorage.removeItem('token');
        setUser(null);
      },
    }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider');
  return ctx;
};
```

---

## 5. Offline-Strategie

### Architecture für Scanner-Modus

```
┌─────────────────────────────────────────────┐
│  Scanner (Zebra TC21 / Handheld)            │
│  - Barcode-Input → ScanInput Component      │
│  - Offline-First: alle Scans lokal speichern│
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  IndexedDB                                  │
│  - scans (id, barcode, qty, timestamp)      │
│  - inventory (equipmentId, qty, location)   │
│  - syncQueue (action, data, status)         │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  Web Worker (offline-sync.worker.ts)        │
│  - Beobachtet: navigator.onLine             │
│  - Batch-Sync alle 30s bei Reconnect        │
│  - Retry mit Exponential Backoff            │
└─────────────────────────────────────────────�────────┐
                    ↓                                   │
┌─────────────────────────────────────────────┐        │
│  Server API                                 │        │
│  - POST /api/warehouse/scans/batch          │        │
│  - POST /api/warehouse/inventory/sync       │        │
│  - Optimistic Response-Handling             │────────┘
└─────────────────────────────────────────────┘
```

### IndexedDB Schema

```typescript
// warehouse/offline/db.ts
import { openDB } from 'idb';

export const initDB = async () => {
  return openDB('event-tech-warehouse', 1, {
    upgrade(db) {
      // Store für Scans
      const scanStore = db.createObjectStore('scans', {
        keyPath: 'id',
        autoIncrement: true,
      });
      scanStore.createIndex('timestamp', 'timestamp');
      scanStore.createIndex('syncStatus', 'syncStatus'); // 'pending' | 'synced'

      // Store für lokale Inventur
      const inventoryStore = db.createObjectStore('inventory', {
        keyPath: 'equipmentId',
      });
      inventoryStore.createIndex('location', 'location');

      // Store für Sync-Queue
      const queueStore = db.createObjectStore('syncQueue', {
        keyPath: 'id',
        autoIncrement: true,
      });
      queueStore.createIndex('status', 'status'); // 'pending' | 'processing' | 'failed'
      queueStore.createIndex('createdAt', 'createdAt');
    },
  });
};

export interface StoredScan {
  id?: number;
  barcode: string;
  equipmentId: string;
  quantity: number;
  action: 'in' | 'out'; // Rein/Raus
  location?: string;
  timestamp: number;
  syncStatus: 'pending' | 'synced';
}

export interface StoredInventory {
  equipmentId: string;
  quantity: number;
  location: string;
  lastUpdated: number;
}

export interface SyncQueueItem {
  id?: number;
  action: 'scan' | 'inventory-update' | 'location-change';
  data: any;
  status: 'pending' | 'processing' | 'failed';
  createdAt: number;
  retries: number;
}
```

### useOfflineStorage Hook

```typescript
// warehouse/hooks/useOfflineStorage.ts
export const useOfflineStorage = () => {
  const [db, setDb] = useState<IDBDatabase | null>(null);

  useEffect(() => {
    initDB().then(setDb);
  }, []);

  return {
    // Store Scan in IndexedDB
    addScan: async (scan: Omit<StoredScan, 'id'>) => {
      const result = await db?.add('scans', {
        ...scan,
        syncStatus: 'pending',
      });
      return result;
    },

    // Fetch pending Scans für Batch-Sync
    getPendingScans: async () => {
      const index = db?.transaction('scans').store.index('syncStatus');
      return await index?.getAll('pending');
    },

    // Update Scan nach erfolgreicher Sync
    markScanSynced: async (scanId: number) => {
      const tx = db?.transaction('scans', 'readwrite');
      const scan = await tx?.store.get(scanId);
      if (scan) {
        scan.syncStatus = 'synced';
        await tx?.store.put(scan);
      }
    },
  };
};
```

### Web Worker für Hintergrund-Sync

```typescript
// workers/offline-sync.worker.ts
let db: IDBDatabase;

self.addEventListener('message', async (event) => {
  const { type } = event.data;

  if (type === 'START_SYNC') {
    // Beobachte Online/Offline Status
    self.addEventListener('online', trySyncBatch);
    // Auch regelmäßig versuchen (alle 30s)
    setInterval(() => {
      if (navigator.onLine) {
        trySyncBatch();
      }
    }, 30000);
  }
});

async function trySyncBatch() {
  try {
    const pendingScans = await db
      .transaction('scans')
      .store.index('syncStatus')
      .getAll('pending');

    if (pendingScans.length === 0) return;

    // Batch-Request an Server
    const response = await fetch('/api/warehouse/scans/batch', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ scans: pendingScans }),
    });

    if (!response.ok) throw new Error('Sync failed');

    // Markiere alle als synced
    const tx = db.transaction('scans', 'readwrite');
    for (const scan of pendingScans) {
      scan.syncStatus = 'synced';
      await tx.store.put(scan);
    }

    // Benachrichtige Main Thread
    self.postMessage({ type: 'SYNC_SUCCESS', count: pendingScans.length });
  } catch (error) {
    // Exponential Backoff
    setTimeout(trySyncBatch, Math.min(1000 * Math.pow(2, retries), 30000));
  }
}
```

### useSyncExternalStore für Offline Status

```typescript
// hooks/useNetworkStatus.ts
export const useNetworkStatus = () => {
  return useSyncExternalStore(
    (subscribe) => {
      window.addEventListener('online', () => subscribe());
      window.addEventListener('offline', () => subscribe());
      return () => {
        window.removeEventListener('online', subscribe);
        window.removeEventListener('offline', subscribe);
      };
    },
    () => navigator.onLine,
  );
};

// In Scanner-Component
export const ScannerPage = () => {
  const isOnline = useNetworkStatus();

  return (
    <>
      {!isOnline && <OfflineIndicator />}
      <ScanInput />
    </>
  );
};
```

---

## 6. i18n-Strategie

### react-i18next Setup mit Namespacing

```typescript
// i18n/i18n.ts
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

// Dynamische Imports pro Namespace
const resources = {
  de: {
    common: () => import('./de/common.json'),
    dashboard: () => import('./de/dashboard.json'),
    warehouse: () => import('./de/warehouse.json'),
    quotes: () => import('./de/quotes.json'),
    accounting: () => import('./de/accounting.json'),
    freelancer: () => import('./de/freelancer.json'),
    errors: () => import('./de/errors.json'),
  },
  en: {
    common: () => import('./en/common.json'),
    dashboard: () => import('./en/dashboard.json'),
    // ... etc
  },
};

i18n
  .use(initReactI18next)
  .init({
    fallbackLng: 'de',
    defaultNS: 'common',
    interpolation: { escapeValue: false },
    ns: ['common', 'dashboard', 'warehouse', 'quotes', 'accounting', 'freelancer'],
    resources, // oder mit i18next-http-backend für dynamisches Laden
    react: {
      useSuspense: false, // für Offline-Modus
    },
  });

export default i18n;
```

### Namespace-Pattern in JSON

```json
// i18n/de/warehouse.json
{
  "scanner": {
    "title": "Barcode-Scanner",
    "placeholder": "Barcode scannen oder eingeben",
    "success": "Scan erfolgreich",
    "error": "Artikel nicht gefunden"
  },
  "inventory": {
    "title": "Bestand",
    "columns": {
      "equipment": "Artikel",
      "quantity": "Menge",
      "location": "Lagerort"
    }
  }
}
```

### Verwendung in Komponenten

```tsx
// warehouse/components/ScanInput.tsx
import { useTranslation } from 'react-i18next';

export const ScanInput = () => {
  const { t } = useTranslation(['warehouse', 'common']);

  return (
    <input
      placeholder={t('warehouse:scanner.placeholder')}
      aria-label={t('warehouse:scanner.title')}
    />
  );
};
```

### Sprach-Umschalter

```tsx
// components/atoms/LanguageSwitcher.tsx
export const LanguageSwitcher = () => {
  const { i18n } = useTranslation();

  return (
    <select
      value={i18n.language}
      onChange={(e) => i18n.changeLanguage(e.target.value)}
    >
      <option value="de">Deutsch</option>
      <option value="en">English</option>
    </select>
  );
};
```

---

## 7. Responsive Design

### Breakpoints Definition

```scss
// styles/breakpoints.scss
$breakpoints: (
  'xs': 0,      // Mobile: Zebra TC21 (320px)
  'sm': 640px,  // iPhone (640px+)
  'md': 768px,  // iPad (768px+)
  'lg': 1024px, // Desktop (1024px+)
  'xl': 1280px, // Large Desktop (1280px+)
);

@mixin media($breakpoint) {
  @media (min-width: map-get($breakpoints, $breakpoint)) {
    @content;
  }
}

// Verwendung:
.sidebar {
  display: none;

  @include media('lg') {
    display: block;
    width: 250px;
  }
}
```

### Device Detection Hook

```typescript
// hooks/useDeviceType.ts
export type DeviceType = 'zebra' | 'mobile' | 'tablet' | 'desktop';

export const useDeviceType = (): DeviceType => {
  const [device, setDevice] = useState<DeviceType>('desktop');

  useEffect(() => {
    const detectDevice = () => {
      const ua = navigator.userAgent;
      const width = window.innerWidth;

      // Zebra TC21 Detection (Android + kleine Screen)
      if (ua.includes('Android') && width < 500) {
        setDevice('zebra');
      }
      // Mobile
      else if (width < 768) {
        setDevice('mobile');
      }
      // Tablet
      else if (width < 1024) {
        setDevice('tablet');
      }
      // Desktop
      else {
        setDevice('desktop');
      }
    };

    detectDevice();
    window.addEventListener('resize', detectDevice);
    return () => window.removeEventListener('resize', detectDevice);
  }, []);

  return device;
};
```

### Layout pro Device-Typ

```tsx
// layouts/AppShell.tsx
export const AppShell = () => {
  const device = useDeviceType();
  const { sidebarOpen } = useUIStore();

  if (device === 'zebra') {
    return <ScannerLayout />;
  }

  if (device === 'mobile') {
    return <MobileLayout />;
  }

  return (
    <div className="app-shell">
      <Header />
      <div className="app-content">
        {sidebarOpen && <Sidebar />}
        <main className="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
};
```

### Responsive Table für Thomas (Buchhaltung)

```tsx
// accounting/components/InvoiceTable.tsx
export const InvoiceTable = ({ invoices }: Props) => {
  const device = useDeviceType();

  // Desktop: Klassische Tabelle
  if (device === 'desktop') {
    return (
      <table className="invoice-table">
        <thead>
          <tr>
            <th>{t('accounting:columns.id')}</th>
            <th>{t('accounting:columns.client')}</th>
            <th>{t('accounting:columns.amount')}</th>
            <th>{t('accounting:columns.status')}</th>
            <th>{t('accounting:columns.actions')}</th>
          </tr>
        </thead>
        <tbody>
          {invoices.map((inv) => (
            <tr key={inv.id}>
              <td>{inv.id}</td>
              <td>{inv.client}</td>
              <td>{formatCurrency(inv.amount)}</td>
              <td><Badge variant={statusVariant(inv.status)} /></td>
              <td>
                <Button onClick={() => handleEdit(inv)}>Edit</Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }

  // Mobile: Karten-Ansicht
  return invoices.map((inv) => (
    <Card key={inv.id} className="invoice-card">
      <h3>{inv.id}</h3>
      <p>{inv.client}</p>
      <p className="amount">{formatCurrency(inv.amount)}</p>
      <Badge variant={statusVariant(inv.status)} />
      <Button onClick={() => handleEdit(inv)}>Edit</Button>
    </Card>
  ));
};
```

---

## 8. Theming

### CSS Custom Properties Setup

```scss
// styles/variables.scss
:root {
  // Farben - Light Mode
  --color-primary: #0070f3;
  --color-secondary: #7c3aed;
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-error: #ef4444;
  --color-info: #3b82f6;

  --color-bg-primary: #ffffff;
  --color-bg-secondary: #f9fafb;
  --color-bg-tertiary: #f3f4f6;

  --color-text-primary: #1f2937;
  --color-text-secondary: #6b7280;
  --color-text-tertiary: #9ca3af;

  --color-border: #e5e7eb;

  // Spacing
  --spacing-xs: 0.25rem;  // 4px
  --spacing-sm: 0.5rem;   // 8px
  --spacing-md: 1rem;     // 16px
  --spacing-lg: 1.5rem;   // 24px
  --spacing-xl: 2rem;     // 32px

  // Typography
  --font-size-xs: 0.75rem;    // 12px
  --font-size-sm: 0.875rem;   // 14px
  --font-size-md: 1rem;       // 16px
  --font-size-lg: 1.125rem;   // 18px
  --font-size-xl: 1.25rem;    // 20px
  --font-size-2xl: 1.5rem;    // 24px

  --font-weight-regular: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;
  --font-weight-bold: 700;

  // Shadows
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1);

  // Radius
  --radius-sm: 0.25rem;
  --radius-md: 0.375rem;
  --radius-lg: 0.5rem;

  // Transitions
  --transition-fast: 150ms ease-out;
  --transition-normal: 300ms ease-out;
}
```

### Dark Mode

```scss
// styles/themes/dark.scss
@media (prefers-color-scheme: dark) {
  :root {
    --color-bg-primary: #1f2937;
    --color-bg-secondary: #111827;
    --color-bg-tertiary: #0f172a;

    --color-text-primary: #f9fafb;
    --color-text-secondary: #d1d5db;
    --color-text-tertiary: #9ca3af;

    --color-border: #374151;
  }
}

// Oder per Daten-Attribut für manuellen Toggle
[data-theme='dark'] {
  --color-bg-primary: #1f2937;
  --color-bg-secondary: #111827;
  // ...
}
```

### High-Contrast Mode für Lager

```scss
// styles/themes/highcontrast.scss
[data-theme='highcontrast'] {
  --color-primary: #0000ff;     // Knalliges Blau
  --color-success: #008000;     // Knalliges Grün
  --color-error: #ff0000;       // Knalliges Rot
  --color-warning: #ff8800;     // Knalliges Orange

  --color-bg-primary: #ffffff;
  --color-text-primary: #000000;
  --color-border: #000000;

  // Fette Border für bessere Sichtbarkeit
  --border-width: 2px;

  // Größere Fonts
  --font-size-md: 1.125rem;
  --font-size-lg: 1.25rem;
}
```

### Theme-Hook

```typescript
// hooks/useTheme.ts
export const useTheme = () => {
  const { theme, setTheme } = useUIStore();

  useEffect(() => {
    const root = document.documentElement;
    root.setAttribute('data-theme', theme);

    // Falls auch prefers-color-scheme ändern
    if (theme === 'dark') {
      root.style.colorScheme = 'dark';
    } else if (theme === 'light') {
      root.style.colorScheme = 'light';
    }
  }, [theme]);

  return { theme, setTheme };
};
```

---

## 9. Scanner-Integration

### Zebra TC21 Handheld Scanner

**Hardware-Eigenschaften:**
- Physischer Scanner-Button (Hardcoded Key: Enter/Return)
- Kleine Screen: 4.7" (480x800px)
- Android OS, WebView-basiert
- Dicke Arbeitshandschuhe → große Touch-Targets (min 44x44px)

**Integration:**

```typescript
// warehouse/hooks/useScanner.ts
export const useScanner = () => {
  const [barcode, setBarcode] = useState('');
  const [lastScanTime, setLastScanTime] = useState(0);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Zebra TC21: Scanner gibt 'Enter' am Ende aus
      if (event.key === 'Enter' && barcode.trim()) {
        // Bounce: doppelte Events innerhalb 100ms ignorieren
        if (Date.now() - lastScanTime < 100) {
          return;
        }

        setLastScanTime(Date.now());

        // Löse Scan-Event aus
        processScan(barcode);
        setBarcode(''); // Reset für nächsten Scan
      } else if (event.key !== 'Enter') {
        // Normal Keyboard Input
        setBarcode((prev) => prev + event.key);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [barcode, lastScanTime]);

  return {
    barcode,
    setBarcode,
  };
};
```

**ScanInput Component:**

```tsx
// warehouse/components/ScanInput.tsx
export const ScanInput = () => {
  const { barcode, setBarcode } = useScanner();
  const { mutate: submitScan, isPending } = useScanMutation();
  const { addScan } = useOfflineStorage();
  const device = useDeviceType();

  const handleScan = async (code: string) => {
    // Offline-First: speichere lokal
    await addScan({
      barcode: code,
      equipment_id: extractEquipmentId(code), // SKU-Parser
      quantity: 1,
      action: 'in', // oder 'out'
      timestamp: Date.now(),
    });

    // Versuche zu syncen wenn online
    if (navigator.onLine) {
      submitScan({ barcode: code });
    }
  };

  // Zebra TC21: Großer Input-Bereich, kein Label nötig
  if (device === 'zebra') {
    return (
      <div className="scan-input--zebra">
        <input
          ref={inputRef}
          value={barcode}
          onChange={(e) => setBarcode(e.target.value)}
          placeholder={t('warehouse:scanner.placeholder')}
          autoFocus
          className="scan-input__field--zebra"
          style={{
            fontSize: '24px',           // Groß für schnelle Erfassung
            padding: '20px',
            width: '100%',
            minHeight: '80px',
          }}
        />
        <div className="scan-input__feedback">
          {isPending && <Spinner />}
        </div>
      </div>
    );
  }

  // Standard Mobile/Desktop
  return (
    <FormField label={t('warehouse:scanner.title')}>
      <input
        value={barcode}
        onChange={(e) => setBarcode(e.target.value)}
        placeholder={t('warehouse:scanner.placeholder')}
      />
    </FormField>
  );
};
```

### Kamera-Scan für iPhones (Kevin, Freelancer)

```typescript
// freelancer/hooks/useCameraScanner.ts
export const useCameraScanner = () => {
  const [isScanning, setIsScanning] = useState(false);

  const startCamera = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'environment' },
      });
      setIsScanning(true);
      return stream;
    } catch (error) {
      // Fallback: manueller Input
      console.warn('Camera not available');
    }
  };

  return {
    startCamera,
    isScanning,
  };
};
```

**CameraCapture Komponente (mit jsQR Library):**

```tsx
// freelancer/components/CameraCapture.tsx
import jsQR from 'jsqr';

export const CameraCapture = ({ onScan }: Props) => {
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const { startCamera } = useCameraScanner();

  useEffect(() => {
    let animationFrameId: number;

    const scanLoop = () => {
      const video = videoRef.current;
      const canvas = canvasRef.current;

      if (video && canvas) {
        const ctx = canvas.getContext('2d')!;
        ctx.drawImage(video, 0, 0, canvas.width, canvas.height);

        const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
        const code = jsQR(imageData.data, canvas.width, canvas.height);

        if (code) {
          onScan(code.data);
          // Stop scanning
          return;
        }
      }

      animationFrameId = requestAnimationFrame(scanLoop);
    };

    const setup = async () => {
      const stream = await startCamera();
      if (stream && videoRef.current) {
        videoRef.current.srcObject = stream;
        videoRef.current.onloadedmetadata = () => {
          videoRef.current!.play();
          scanLoop();
        };
      }
    };

    setup();

    return () => {
      cancelAnimationFrame(animationFrameId);
      if (videoRef.current?.srcObject) {
        (videoRef.current.srcObject as MediaStream)
          .getTracks()
          .forEach((t) => t.stop());
      }
    };
  }, []);

  return (
    <div className="camera-capture">
      <video ref={videoRef} style={{ display: 'none' }} />
      <canvas ref={canvasRef} style={{ display: 'none' }} />
      <div className="camera-capture__viewfinder">
        <p>{t('freelancer:capture.aim')}</p>
      </div>
    </div>
  );
};
```

---

## 10. Echtzeit-Updates

### WebSocket für Packlisten (Multi-User)

```typescript
// services/realtimeService.ts
import { io, Socket } from 'socket.io-client';

class RealtimeService {
  private socket: Socket;

  constructor() {
    this.socket = io(import.meta.env.VITE_API_URL, {
      reconnection: true,
      reconnectionDelay: 1000,
      reconnectionDelayMax: 5000,
      reconnectionAttempts: Infinity,
    });
  }

  // Subscribe auf Packlisten-Updates
  onPacklistUpdate(
    packlistId: string,
    callback: (update: PacklistUpdate) => void
  ) {
    this.socket.on(`packlist:${packlistId}`, callback);
  }

  // User checkt Item in Packliste
  checkPacklistItem(packlistId: string, itemId: string) {
    this.socket.emit('packlist:check', { packlistId, itemId });
  }

  disconnect() {
    this.socket.disconnect();
  }
}

export const realtimeService = new RealtimeService();
```

**Hook für Packlisten:**

```typescript
// freelancer/hooks/usePacklistRealtime.ts
export const usePacklistRealtime = (packlistId: string) => {
  const [items, setItems] = useState<PacklistItem[]>([]);

  useEffect(() => {
    // Initial Load
    fetchPacklist(packlistId).then(setItems);

    // Subscribe zu Live-Updates
    const unsubscribe = realtimeService.onPacklistUpdate(
      packlistId,
      (update: PacklistUpdate) => {
        setItems((prev) =>
          prev.map((item) =>
            item.id === update.itemId
              ? { ...item, checked: update.checked, checkedBy: update.checkedBy }
              : item,
          ),
        );
      },
    );

    return unsubscribe;
  }, [packlistId]);

  return items;
};
```

### Server-Sent Events (SSE) als Alternative

```typescript
// hooks/useSSEStream.ts
export const useSSEStream = (endpoint: string) => {
  const [data, setData] = useState<any>(null);

  useEffect(() => {
    const eventSource = new EventSource(endpoint);

    eventSource.addEventListener('update', (event) => {
      setData(JSON.parse(event.data));
    });

    eventSource.addEventListener('error', () => {
      eventSource.close();
    });

    return () => eventSource.close();
  }, [endpoint]);

  return data;
};
```

---

## 11. Performance

### Code-Splitting per Route

```tsx
// routes/index.tsx
import { lazy, Suspense } from 'react';

const DashboardPage = lazy(() =>
  import('../features/dashboard/pages/DashboardPage'),
);
const WarehousePage = lazy(() =>
  import('../features/warehouse/pages/ScannerPage'),
);

export const router = createBrowserRouter([
  {
    path: 'dashboard',
    element: (
      <Suspense fallback={<LoadingSpinner />}>
        <DashboardPage />
      </Suspense>
    ),
  },
]);
```

### Lazy Loading von Features

```typescript
// useFeatureLoader Hook
export const useFeatureLoader = (featureId: string) => {
  return useMemo(() => {
    if (!featureId) return null;
    return lazy(() =>
      import(`../features/${featureId}`).catch(() => ({
        default: NotFoundComponent,
      })),
    );
  }, [featureId]);
};
```

### Virtual List für große Tabellen (Thomas)

```tsx
// accounting/components/InvoiceTable.tsx
import { FixedSizeList } from 'react-window';

export const InvoiceTable = ({ invoices }: Props) => {
  const Row = ({ index, style }: any) => {
    const invoice = invoices[index];
    return (
      <div style={style} className="invoice-row">
        <div>{invoice.id}</div>
        <div>{invoice.client}</div>
        <div>{formatCurrency(invoice.amount)}</div>
      </div>
    );
  };

  return (
    <FixedSizeList
      height={600}
      itemCount={invoices.length}
      itemSize={50}
      width="100%"
    >
      {Row}
    </FixedSizeList>
  );
};
```

### Debouncing für SearchBar

```typescript
// hooks/useDebounce.ts
export const useDebounce = <T,>(value: T, delay = 300) => {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(handler);
  }, [value, delay]);

  return debouncedValue;
};

// Verwendung
export const EquipmentSearch = () => {
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);

  const { data: results } = useQuery({
    queryKey: ['equipment', debouncedSearch],
    queryFn: () => client.get('/api/equipment', { params: { q: debouncedSearch } }),
    enabled: debouncedSearch.length > 2,
  });

  return (
    <>
      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Equipment suchen..."
      />
      {results && <EquipmentList items={results} />}
    </>
  );
};
```

### Image Optimization

```tsx
// Lazy Loading mit native loading attribute
<img
  src={imageUrl}
  alt="description"
  loading="lazy"
  width={400}
  height={300}
/>

// oder mit next-gen Formate
<picture>
  <source srcSet={imageUrl} type="image/webp" />
  <img src={imageFallback} alt="description" loading="lazy" />
</picture>
```

---

## 12. Testing-Strategie

### Unit Tests mit Vitest

```typescript
// tests/unit/utils/formatters.test.ts
import { describe, it, expect } from 'vitest';
import { formatCurrency, formatDate } from '../../../lib/formatters';

describe('formatters', () => {
  it('formats currency correctly', () => {
    expect(formatCurrency(1234.56, 'de-DE')).toBe('1.234,56 €');
    expect(formatCurrency(1234.56, 'en-US')).toBe('$1,234.56');
  });

  it('formats dates correctly', () => {
    const date = new Date('2026-03-20');
    expect(formatDate(date, 'de-DE')).toBe('20. März 2026');
  });
});
```

### Integration Tests mit Testing Library

```typescript
// tests/integration/features/warehouse/scanner.test.tsx
import { render, screen, userEvent } from '@testing-library/react';
import { QueryClientProvider } from '@tanstack/react-query';
import { ScannerPage } from '../../../features/warehouse/pages/ScannerPage';

describe('Scanner Integration', () => {
  it('should capture barcode and submit scan', async () => {
    const { queryClient } = setup();

    render(
      <QueryClientProvider client={queryClient}>
        <ScannerPage />
      </QueryClientProvider>,
    );

    const input = screen.getByRole('textbox');
    await userEvent.type(input, '123456789{Enter}');

    expect(screen.getByText(/Scan erfolgreich/i)).toBeInTheDocument();
  });

  it('should handle offline mode', async () => {
    // Mock navigator.onLine
    Object.defineProperty(navigator, 'onLine', { value: false });

    // Render & test offline behavior
    render(<ScannerPage />);

    expect(screen.getByText(/Offline-Modus/i)).toBeInTheDocument();
  });
});
```

### E2E Tests mit Playwright

```typescript
// tests/e2e/warehouse-scanner.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Warehouse Scanner Flow', () => {
  test('should scan equipment in/out and sync on reconnect', async ({
    page,
  }) => {
    // Login
    await page.goto('/login');
    await page.fill('input[type="email"]', 'lisa@company.de');
    await page.fill('input[type="password"]', 'password');
    await page.click('button:has-text("Login")');

    // Navigate to Scanner
    await page.click('a:has-text("Scanner")');

    // Simulate offline
    await page.context().setOffline(true);

    // Scan barcode
    await page.fill('input[placeholder*="Barcode"]', '123456789');
    await page.keyboard.press('Enter');

    expect(page.locator('.scan-result--success')).toBeVisible();

    // Check IndexedDB storage
    const pendingScans = await page.evaluate(() => {
      return (window as any).__db
        ?.transaction('scans')
        .store.getAll('pending');
    });
    expect(pendingScans.length).toBe(1);

    // Go online & verify sync
    await page.context().setOffline(false);
    await page.waitForTimeout(5000); // Warte auf Sync

    const syncedScans = await page.evaluate(() => {
      return (window as any).__db
        ?.transaction('scans')
        .store.getAll('synced');
    });
    expect(syncedScans.length).toBe(1);
  });
});
```

---

## 13. Design System: Basis-Komponenten

### Button Komponente

```tsx
// components/atoms/Button.tsx
import React from 'react';
import './Button.scss';

interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  fullWidth?: boolean;
  icon?: React.ReactNode;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      variant = 'primary',
      size = 'md',
      loading = false,
      fullWidth = false,
      icon,
      children,
      ...props
    },
    ref,
  ) => {
    return (
      <button
        ref={ref}
        className={cn(
          'btn',
          `btn--${variant}`,
          `btn--${size}`,
          { 'btn--full-width': fullWidth },
          { 'btn--loading': loading },
        )}
        disabled={loading || props.disabled}
        {...props}
      >
        {loading && <Spinner size={size} />}
        {icon && <span className="btn__icon">{icon}</span>}
        {children}
      </button>
    );
  },
);

Button.displayName = 'Button';
```

**Button Styles:**

```scss
// components/atoms/Button.scss
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);

  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-md);
  border: none;
  cursor: pointer;
  transition: all var(--transition-fast);

  &--primary {
    background-color: var(--color-primary);
    color: white;

    &:hover:not(:disabled) {
      background-color: darken($color-primary, 10%);
      box-shadow: var(--shadow-md);
    }

    &:active:not(:disabled) {
      transform: scale(0.98);
    }
  }

  &--secondary {
    background-color: var(--color-bg-secondary);
    color: var(--color-text-primary);
    border: 1px solid var(--color-border);

    &:hover:not(:disabled) {
      background-color: var(--color-bg-tertiary);
    }
  }

  &--danger {
    background-color: var(--color-error);
    color: white;

    &:hover:not(:disabled) {
      background-color: darken($color-error, 10%);
    }
  }

  &--outline {
    border: 2px solid var(--color-primary);
    color: var(--color-primary);
    background-color: transparent;

    &:hover:not(:disabled) {
      background-color: rgba(0, 112, 243, 0.05);
    }
  }

  &--ghost {
    background-color: transparent;
    color: var(--color-text-primary);

    &:hover:not(:disabled) {
      background-color: var(--color-bg-secondary);
    }
  }

  // Größen
  &--sm {
    padding: var(--spacing-xs) var(--spacing-sm);
    font-size: var(--font-size-sm);
    min-height: 32px;

    // Für Zebra: mindestens 44px
    @include media('xs') {
      min-height: 44px;
      padding: var(--spacing-sm) var(--spacing-md);
    }
  }

  &--md {
    padding: var(--spacing-sm) var(--spacing-md);
    font-size: var(--font-size-md);
    min-height: 40px;
  }

  &--lg {
    padding: var(--spacing-md) var(--spacing-lg);
    font-size: var(--font-size-lg);
    min-height: 48px;
  }

  &--full-width {
    width: 100%;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &--loading {
    pointer-events: none;
  }
}
```

### Input Komponente

```tsx
// components/atoms/Input.tsx
interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: string;
  label?: string;
  hint?: string;
  required?: boolean;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ error, label, hint, required, ...props }, ref) => {
    const id = props.id || `input-${Math.random()}`;

    return (
      <div className="input-wrapper">
        {label && (
          <label htmlFor={id} className="input__label">
            {label}
            {required && <span className="input__required">*</span>}
          </label>
        )}
        <input
          ref={ref}
          id={id}
          className={cn('input', { 'input--error': error })}
          aria-invalid={!!error}
          aria-describedby={error ? `${id}-error` : undefined}
          {...props}
        />
        {error && (
          <span id={`${id}-error`} className="input__error">
            {error}
          </span>
        )}
        {hint && !error && (
          <span className="input__hint">{hint}</span>
        )}
      </div>
    );
  },
);

Input.displayName = 'Input';
```

### Badge Komponente

```tsx
// components/atoms/Badge.tsx
interface BadgeProps {
  variant?: 'success' | 'warning' | 'error' | 'info' | 'default';
  size?: 'sm' | 'md' | 'lg';
  children: React.ReactNode;
}

export const Badge: React.FC<BadgeProps> = ({
  variant = 'default',
  size = 'md',
  children,
}) => {
  return (
    <span className={cn('badge', `badge--${variant}`, `badge--${size}`)}>
      {children}
    </span>
  );
};
```

**Badge Styles (Zebra-optimiert: große Schrift):**

```scss
// components/atoms/_atoms.scss
.badge {
  display: inline-block;
  border-radius: var(--radius-md);
  font-weight: var(--font-weight-semibold);
  white-space: nowrap;

  &--success {
    background-color: var(--color-success);
    color: white;
  }

  &--warning {
    background-color: var(--color-warning);
    color: white;
  }

  &--error {
    background-color: var(--color-error);
    color: white;
  }

  &--info {
    background-color: var(--color-info);
    color: white;
  }

  // Größen
  &--sm {
    padding: 2px 8px;
    font-size: var(--font-size-xs);
  }

  &--md {
    padding: 4px 12px;
    font-size: var(--font-size-sm);
  }

  &--lg {
    padding: 8px 16px;
    font-size: var(--font-size-md);

    // Zebra: noch größer
    @include media('xs') {
      padding: 12px 20px;
      font-size: var(--font-size-lg);
    }
  }
}
```

### Table Komponente (Thomas' Lieblings-Komponente)

```tsx
// components/molecules/Table.tsx
interface TableProps {
  columns: ColumnDef[];
  data: any[];
  onSort?: (column: string, direction: 'asc' | 'desc') => void;
  selectable?: boolean;
}

export const Table: React.FC<TableProps> = ({
  columns,
  data,
  onSort,
  selectable = false,
}) => {
  const [sortBy, setSortBy] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');

  const handleSort = (columnId: string) => {
    const newDir = sortBy === columnId && sortDir === 'asc' ? 'desc' : 'asc';
    setSortBy(columnId);
    setSortDir(newDir);
    onSort?.(columnId, newDir);
  };

  return (
    <div className="table-wrapper">
      <table className="table" role="grid">
        <thead className="table__head">
          <tr>
            {selectable && (
              <th className="table__cell table__cell--checkbox">
                <input type="checkbox" />
              </th>
            )}
            {columns.map((col) => (
              <th
                key={col.id}
                className={cn('table__cell table__cell--header', {
                  'table__cell--sortable': col.sortable,
                })}
                onClick={() => col.sortable && handleSort(col.id)}
                role="columnheader"
              >
                {col.label}
                {col.sortable && (
                  <span className="table__sort-icon">
                    {sortBy === col.id && (
                      <Icon name={sortDir === 'asc' ? 'chevron-up' : 'chevron-down'} />
                    )}
                  </span>
                )}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="table__body">
          {data.map((row, rowIdx) => (
            <tr key={rowIdx} className="table__row">
              {selectable && (
                <td className="table__cell table__cell--checkbox">
                  <input type="checkbox" />
                </td>
              )}
              {columns.map((col) => (
                <td
                  key={col.id}
                  className="table__cell"
                  data-label={col.label}
                >
                  {col.render ? col.render(row) : row[col.id]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
```

**Table Styles (Keyboard-Navigation für Thomas):**

```scss
// components/molecules/_molecules.scss
.table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-md);

  &__head {
    background-color: var(--color-bg-secondary);
    border-bottom: 2px solid var(--color-border);
  }

  &__cell {
    padding: var(--spacing-md);
    text-align: left;
    border-bottom: 1px solid var(--color-border);

    &--header {
      font-weight: var(--font-weight-semibold);
      user-select: none;
    }

    &--sortable {
      cursor: pointer;

      &:hover {
        background-color: var(--color-bg-tertiary);
      }
    }
  }

  &__row {
    &:hover {
      background-color: var(--color-bg-secondary);
    }

    // Keyboard Focus für Tab-Navigation
    &:focus-within {
      outline: 2px solid var(--color-primary);
      outline-offset: -2px;
    }
  }
}

// Mobile: Karten-Ansicht
@include media('md') {
  .table {
    &__head {
      display: none;
    }

    &__body {
      display: grid;
      gap: var(--spacing-md);
    }

    &__row {
      display: block;
      border: 1px solid var(--color-border);
      border-radius: var(--radius-lg);
      padding: var(--spacing-md);

      &:hover {
        background-color: transparent;
        box-shadow: var(--shadow-md);
      }
    }

    &__cell {
      display: grid;
      grid-template-columns: 150px 1fr;
      gap: var(--spacing-md);
      padding: var(--spacing-sm) 0;
      border: none;

      &::before {
        content: attr(data-label);
        font-weight: var(--font-weight-semibold);
      }
    }
  }
}
```

### Toast/Notification System

```tsx
// components/molecules/Toast.tsx
interface ToastProps {
  id: string;
  variant: 'success' | 'error' | 'warning' | 'info';
  message: string;
  action?: { label: string; onClick: () => void };
  autoClose?: number; // ms
}

export const Toast: React.FC<ToastProps> = ({
  id,
  variant,
  message,
  action,
  autoClose = 5000,
}) => {
  const { removeNotification } = useNotificationStore();

  useEffect(() => {
    if (autoClose) {
      const timer = setTimeout(() => removeNotification(id), autoClose);
      return () => clearTimeout(timer);
    }
  }, [id, autoClose, removeNotification]);

  return (
    <div className={cn('toast', `toast--${variant}`)} role="alert">
      <span className="toast__message">{message}</span>
      {action && (
        <button
          className="toast__action"
          onClick={action.onClick}
        >
          {action.label}
        </button>
      )}
      <button
        className="toast__close"
        onClick={() => removeNotification(id)}
        aria-label="Close"
      >
        ✕
      </button>
    </div>
  );
};

// Container für alle Toasts
export const ToastContainer = () => {
  const notifications = useNotificationStore((s) => s.notifications);

  return (
    <div className="toast-container" aria-live="polite" aria-atomic="true">
      {notifications.map((notif) => (
        <Toast key={notif.id} {...notif} />
      ))}
    </div>
  );
};
```

**Toast Styles:**

```scss
.toast-container {
  position: fixed;
  bottom: var(--spacing-lg);
  right: var(--spacing-lg);
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);

  // Mobile: oben, Vollbreite
  @include media('md') {
    bottom: auto;
    top: var(--spacing-lg);
    left: var(--spacing-md);
    right: var(--spacing-md);
  }
}

.toast {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  background-color: var(--color-bg-primary);
  min-width: 320px;
  animation: slideIn 300ms ease-out;

  &--success {
    border-left: 4px solid var(--color-success);
  }

  &--error {
    border-left: 4px solid var(--color-error);
  }

  &--warning {
    border-left: 4px solid var(--color-warning);
  }

  &__message {
    flex: 1;
    font-size: var(--font-size-md);
  }

  &__action {
    background-color: transparent;
    border: none;
    color: var(--color-primary);
    font-weight: var(--font-weight-semibold);
    cursor: pointer;
    padding: 0;

    &:hover {
      text-decoration: underline;
    }
  }

  &__close {
    background-color: transparent;
    border: none;
    color: var(--color-text-secondary);
    cursor: pointer;
    font-size: 20px;
    padding: 0;

    &:hover {
      color: var(--color-text-primary);
    }
  }
}

@keyframes slideIn {
  from {
    transform: translateX(400px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
```

---

## Summary: Architektur-Highlights

✅ **Scalable**: Feature-basierte Struktur mit klaren Grenzen
✅ **Accessible**: Keyboard-Navigation (Thomas), High-Contrast (Lisa), Mobile-First
✅ **Performant**: Code-Splitting, Virtual Lists, Debouncing, Web Workers
✅ **Offline-Ready**: IndexedDB + Sync-Queue für Scanner
✅ **Responsive**: Device-Detection + Breakpoint-System für alle 4 Personas
✅ **Themable**: CSS Custom Properties für Light/Dark/HighContrast
✅ **Testable**: Vitest + Testing Library + Playwright für alle Layer
✅ **i18n-Ready**: Namespace-Struktur pro Feature + DE/EN von Anfang an

