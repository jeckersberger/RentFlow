# RentFlow Frontend - Complete Implementation

## Overview

A professional, enterprise-grade React 18 + TypeScript frontend application for event equipment rental management. Built with modern tooling, responsive design, and comprehensive type safety.

## Architecture

### Core Technologies
- **React 18** with Hooks for state management
- **TypeScript** for type safety
- **React Router v6** for navigation
- **TanStack Query (React Query)** for data fetching and caching
- **Zustand** for global auth state management
- **SCSS Modules** for scoped styling
- **Axios** for HTTP requests with interceptors

### Project Structure

```
frontend/src/
├── components/
│   ├── Layout/
│   │   ├── MainLayout.tsx       # Main app layout wrapper
│   │   ├── Header.tsx           # Top navigation bar
│   │   ├── Sidebar.tsx          # Left navigation sidebar
│   │   └── *.scss               # Layout styles
│   ├── DataTable/
│   │   ├── DataTable.tsx        # Reusable data table component
│   │   └── DataTable.module.scss
│   ├── Form/
│   │   ├── Input.tsx            # Text input component
│   │   ├── Select.tsx           # Dropdown component
│   │   ├── TextArea.tsx         # Text area component
│   │   ├── FileUpload.tsx       # File upload component
│   │   └── Form.module.scss
│   ├── Modal/
│   │   ├── Modal.tsx            # Reusable modal dialog
│   │   └── Modal.module.scss
│   ├── KPICard/
│   │   ├── KPICard.tsx          # Dashboard KPI card
│   │   └── KPICard.module.scss
│   └── StatusBadge/
│       ├── StatusBadge.tsx      # Status indicator badge
│       └── StatusBadge.module.scss
├── pages/
│   ├── Login.tsx                # Authentication page
│   ├── Dashboard.tsx            # Main dashboard
│   ├── Equipment/
│   │   ├── EquipmentList.tsx    # Equipment list with filters
│   │   ├── EquipmentDetail.tsx  # Equipment detail view
│   │   ├── EquipmentForm.tsx    # Equipment create/edit form
│   │   └── Equipment.module.scss
│   ├── Projects/
│   │   ├── ProjectList.tsx      # Project list with status filters
│   │   ├── ProjectDetail.tsx    # Project detail view
│   │   └── [ProjectForm.tsx]    # Placeholder for form
│   ├── Invoices/
│   │   ├── InvoiceList.tsx      # Invoice list with overdue highlighting
│   │   ├── InvoiceDetail.tsx    # Invoice detail with line items
│   │   └── [InvoiceForm.tsx]    # Placeholder for form
│   ├── Scanner/
│   │   ├── ScannerPage.tsx      # Barcode scanner interface
│   │   └── Scanner.module.scss
│   ├── Warehouse/
│   │   ├── WarehousePage.tsx    # Warehouse tree view
│   │   └── Warehouse.module.scss
│   └── Settings/
│       ├── SettingsPage.tsx     # System settings with tabs
│       └── Settings.module.scss
├── services/
│   └── api.ts                   # API client with auth interceptor
├── stores/
│   └── authStore.ts             # Zustand auth store (persistent)
├── types/
│   ├── equipment.ts             # Equipment types
│   ├── project.ts               # Project types
│   ├── invoice.ts               # Invoice types
│   ├── scanner.ts               # Scanner types
│   └── warehouse.ts             # Warehouse types
├── styles/
│   ├── variables.scss           # Design tokens (colors, spacing, etc.)
│   └── global.scss              # Global styles
├── App.tsx                      # Route configuration
└── main.tsx                     # Entry point
```

## Features Implemented

### 1. Authentication
- Email/password login form
- "Remember me" checkbox
- Error handling and validation
- Token-based auth with Bearer scheme
- Persistent session with Zustand
- Auth interceptor on all API requests

### 2. Dashboard
- 4 KPI cards (Total Equipment, Available, Active Projects, Pending Invoices)
- Equipment distribution chart (bar charts using div-based styling)
- Quick action buttons (New Equipment, Project, Invoice, Scanner)
- Recent activity feed with timestamps
- Responsive grid layout

### 3. Equipment Management
- **List View**: Searchable/filterable equipment table
  - Search by name or SKU
  - Filter by category and status
  - Pagination support
  - Sortable columns
  - Click rows to navigate to detail view

- **Detail View**: Complete equipment information
  - Display all fields with status badges
  - Status change modal
  - Edit button linking to form

- **Form**: Create/edit equipment
  - All fields: name, SKU, barcode, category, location, prices
  - Image upload support
  - Form validation
  - Success/error handling

### 4. Project Management
- **List View**: Projects with status filter tabs
  - Tab-based filtering (All, Draft, Confirmed, In Progress, Completed)
  - Search by project name or client
  - Table with status, dates, budget
  - Pagination

- **Detail View**: Project information
  - Client, dates, location, budget
  - Expandable sections: Info, Packlist, Crew
  - Edit button

### 5. Invoice Management
- **List View**: Invoices with overdue highlighting
  - Status filter tabs (All, Draft, Sent, Paid, Overdue)
  - Search by invoice number or client
  - Overdue dates shown in red
  - Pagination

- **Detail View**: Complete invoice with line items
  - Header: number, date, client, status
  - Line items table with calculations
  - Summary box: subtotal, tax, total
  - Action buttons: Send, Mark Paid, Cancel
  - PDF download placeholder

### 6. Scanner
- Large barcode input field with focus
- Scan type selector (Check-in, Check-out, Inventory)
- Project selector for check-out operations
- Recent scans history (last 10)
- Batch mode toggle
- Today's scan count and warehouse stats

### 7. Warehouse Management
- Tree view of warehouse structure
  - Site → Room → Rack → Shelf hierarchy
  - Collapsible/expandable locations
  - Equipment count per location
  - Search within warehouse

- Detail panel for selected location
  - Type, equipment count, capacity %
  - Sub-locations list

### 8. Settings
- Tabbed interface (Company, Users, Categories, Notifications)
- Company info (name, address, phone, email)
- User management table with roles and status
- Category management
- Notification preferences with toggles

## Component Hierarchy

### Shared Components

#### DataTable
Generic, reusable table component with:
- Sortable columns (with callback)
- Pagination controls
- Loading state (skeleton)
- Empty state messaging
- Custom cell rendering
- German labels ("Seite 1 von 5", "Nächste", etc.)

#### Form Components
- **Input**: Text/email/number inputs with validation
- **Select**: Dropdown with options
- **TextArea**: Multi-line text input
- **FileUpload**: Drag-and-drop file upload with preview
All with error display, labels, and helper text

#### Modal
- Configurable size (sm, md, lg)
- Header with close button
- Content area
- Footer for actions
- Backdrop click to close
- Prevents body scroll when open

#### StatusBadge
Color-coded badges for status indicators:
- Success (green): available, confirmed, paid
- Warning (amber): reserved, draft, in_progress
- Danger (red): maintenance, overdue, retired
- Info (cyan): checked_out, sent
- Secondary (gray): default

#### KPICard
Dashboard statistics card with:
- Icon support
- Value display (with skeleton loading)
- Change indicator (up/down percentage)
- Hover effect

## Styling System

### Design Tokens (variables.scss)
- **Colors**: Primary blue, secondary grays, status colors (success, warning, danger, info)
- **Spacing**: 4px base unit (spacing-1 through spacing-32)
- **Typography**: Font families, sizes (xs to 5xl), weights, line heights
- **Border Radius**: From xs (4px) to full (9999px)
- **Shadows**: From sm to 2xl
- **Animations**: Fast (150ms), normal (250ms), slow (350ms)
- **Z-Index Scale**: Organized hierarchy from base (0) to notification (1080)
- **Breakpoints**: Mobile-first (xs: 320px, sm: 640px, md: 768px, lg: 1024px, xl: 1280px, 2xl: 1536px)

### CSS Modules
Each component and page has its own `.module.scss` file:
- No global namespace pollution
- Component-scoped styles
- Imported directly into components
- SCSS features: nesting, variables, mixins

### Dark Mode Support
All colors automatically adjust based on `prefers-color-scheme: dark` media query

### Responsive Design
- Mobile-first approach
- Flexbox and CSS Grid for layouts
- Breakpoint-specific media queries
- Touch-friendly form controls and buttons

## API Integration

### Auth Interceptor
```typescript
api.interceptors.request.use(config => {
  const token = useAuthStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)
```

### Service Methods
- `equipmentApi.list()` - Paginated list with filters
- `equipmentApi.getById(id)` - Single record
- `equipmentApi.create(data)` - Create new
- `equipmentApi.update(id, data)` - Update existing
- Similar methods for projects, invoices, etc.

## State Management

### Zustand Auth Store
Persistent store with:
- User info (id, email, name)
- Auth token
- Login/logout methods
- LocalStorage persistence
- Automatic recovery on page reload

## Type Safety

### Equipment Types
```typescript
type EquipmentStatus = 'available' | 'reserved' | 'checked_out' | 'maintenance' | 'retired'
interface Equipment { ... }
interface CreateEquipmentDTO { ... }
interface UpdateEquipmentDTO { ... }
```

Similar type definitions for:
- Projects (ProjectStatus: draft, confirmed, in_progress, completed, cancelled)
- Invoices (InvoiceStatus: draft, sent, paid, overdue, cancelled)
- Scanner (ScanType: check_in, check_out, inventory)
- Warehouse (WarehouseLocation hierarchy)

## Routing

Main routes handled by App.tsx:
```
/                          → Dashboard
/login                     → Login
/equipment                 → Equipment List
/equipment/new             → Equipment Form (Create)
/equipment/:id             → Equipment Detail
/equipment/:id/edit        → Equipment Form (Edit)
/projects                  → Project List
/projects/:id              → Project Detail
/invoices                  → Invoice List
/invoices/:id              → Invoice Detail
/scanner                   → Scanner Page
/warehouse                 → Warehouse View
/settings                  → Settings Page
```

## Navigation

### Sidebar (Left)
- Dashboard (📊)
- Ausrüstung/Equipment (📦)
- Projekte/Projects (📋)
- Rechnungen/Invoices (💰)
- Scanner (📱)
- Lager/Warehouse (🏢)
- Einstellungen/Settings (⚙️)

### Header (Top)
- App logo/title (left)
- User profile with email (center-right)
- Logout button (right)

## German Localization

All UI labels and messages use German:
- "Ausrüstung" (Equipment)
- "Projekte" (Projects)
- "Rechnungen" (Invoices)
- "Lager" (Warehouse)
- "Speichern" (Save)
- "Abbrechen" (Cancel)
- "Bearbeiten" (Edit)
- Status translations
- Validation messages in German

## Validation

### Form Validation
- Email format validation
- Required field checks
- Number range validation (prices >= 0)
- File type/size validation for uploads
- Real-time error clearing on input change

### API Error Handling
- Display user-friendly error messages
- Network error handling
- 401 redirect to login
- Form submission error state

## Performance Optimizations

### React Query (TanStack Query)
- Automatic caching with 5-minute stale time
- Garbage collection after 10 minutes
- Request deduplication
- Background refetching
- Pagination support

### Code Splitting
- Lazy-loaded pages via React Router
- Component-level code splitting ready

### Rendering
- Memoization-ready component structure
- Conditional rendering for large lists
- Pagination to limit DOM nodes

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- ES2020+ JavaScript features
- CSS Grid and Flexbox
- Local Storage
- Fetch API

## Development Commands

```bash
# Development server
npm run dev

# Type checking
npx tsc --noEmit

# Build for production
npm run build

# Preview production build
npm run preview
```

## File Sizes Reference

- Login page: ~8KB
- Dashboard: ~12KB
- Equipment pages: ~25KB total
- Project pages: ~18KB total
- Invoice pages: ~20KB total
- Scanner: ~8KB
- Warehouse: ~9KB
- Settings: ~9KB
- Shared components: ~35KB total

## Future Enhancements

1. Project/Invoice form pages (currently placeholders)
2. Crew management pages
3. Expense tracking
4. Reports and analytics
5. Real camera integration for scanner
6. Export to CSV/PDF
7. Real-time updates via WebSockets
8. Offline mode
9. Advanced filtering and search
10. Batch operations
11. Undo/redo functionality
12. Audit logging

## Notes

- All components follow React best practices
- TypeScript strict mode enabled
- Accessibility considerations (semantic HTML, ARIA labels)
- Mobile-responsive design tested
- Dark mode support included
- Form handling with Zustand for auth, React Query for data
- No external UI library dependency (custom components)
