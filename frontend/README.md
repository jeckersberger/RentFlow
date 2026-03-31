# RentFlow Frontend

A modern React 18 + TypeScript + Vite frontend for the RentFlow event and equipment rental management system.

## Technology Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Fast build tool and dev server
- **React Router v6** - Client-side routing
- **TanStack Query (React Query)** - Server state management
- **Zustand** - Client state management
- **Axios** - HTTP client
- **SCSS** - Styling with design tokens
- **ESLint** - Code quality

## Project Structure

```
frontend/
├── public/                  # Static assets
├── src/
│   ├── components/
│   │   ├── Layout/         # Main layout components
│   │   │   ├── MainLayout.tsx
│   │   │   ├── Header.tsx
│   │   │   └── Sidebar.tsx
│   │   └── ...
│   ├── pages/              # Page components
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   └── ...
│   ├── services/           # API services
│   │   └── api.ts
│   ├── stores/             # Zustand stores
│   │   └── authStore.ts
│   ├── styles/             # Global styles
│   │   ├── global.scss
│   │   └── variables.scss
│   ├── types/              # TypeScript types
│   ├── hooks/              # Custom React hooks
│   ├── App.tsx             # Root component
│   └── main.tsx            # Entry point
├── index.html              # HTML template
├── package.json
├── tsconfig.json
├── vite.config.ts
├── nginx.conf              # Production server config
├── Dockerfile              # Container build
└── README.md
```

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

```bash
cd frontend
npm install
```

### Development

```bash
# Start development server (hot reload)
npm run dev

# The app will be available at http://localhost:3000
# API requests will be proxied to http://localhost:80
```

### Build

```bash
# Create optimized production build
npm run build

# Preview production build locally
npm run preview
```

### Type Checking

```bash
# Check TypeScript types
npm run type-check
```

### Linting

```bash
# Run ESLint
npm run lint
```

## Environment Variables

Create a `.env.local` file for local development:

```bash
# API Configuration
VITE_API_BASE_URL=http://localhost:80

# Feature flags
VITE_ENABLE_DEBUG=false
VITE_ENABLE_ANALYTICS=true

# App configuration
VITE_APP_NAME=RentFlow
VITE_APP_VERSION=1.0.0
```

See `.env.example` for all available variables.

## Architecture

### State Management

#### Authentication (Zustand)
Located in `src/stores/authStore.ts`:
- Stores JWT token and user info
- Persists to localStorage
- Auto-clears on 401 responses

```typescript
const { token, user, isAuthenticated, login, logout } = useAuthStore()
```

#### Server State (TanStack Query)
Used for API data fetching:
- Automatic caching and refetching
- Background synchronization
- Optimistic updates support

```typescript
const { data, isLoading, error } = useQuery({
  queryKey: ['equipment'],
  queryFn: () => equipmentApi.list(),
})
```

### API Integration

All API calls go through `src/services/api.ts`:

```typescript
// Automatic JWT token injection
// Automatic 401 handling (redirects to login)
// Interceptors for request/response transformation

import { api, authApi, equipmentApi } from '@/services/api'

// Make requests
const user = await authApi.login(email, password)
const equipment = await equipmentApi.list()
```

### Routing

React Router with lazy loading support:

```tsx
<BrowserRouter>
  <Routes>
    <Route path="/login" element={<LoginPage />} />
    <Route path="/" element={<PrivateRoute><MainLayout /></PrivateRoute>}>
      <Route index element={<DashboardPage />} />
      <Route path="equipment" element={<EquipmentPage />} />
    </Route>
  </Routes>
</BrowserRouter>
```

### Styling

Design tokens-based SCSS with CSS custom properties:

```scss
// Colors
--color-primary: #3b82f6
--color-success: #10b981
--color-danger: #ef4444

// Typography
--font-size-base: 1rem
--font-weight-bold: 700
--line-height-normal: 1.5

// Spacing (4px base unit)
--spacing-4: 1rem // 16px
--spacing-8: 2rem // 32px

// Components
--button-height-md: 2.5rem
--input-height: 2.5rem
--sidebar-width: 16rem

// Dark mode support
@media (prefers-color-scheme: dark) {
  --color-bg-primary: var(--color-gray-900)
  --color-text-primary: var(--color-gray-50)
}
```

See `src/styles/variables.scss` for complete design system.

## Features

### Authentication
- Email/password login
- JWT token management
- Automatic token refresh (implement on backend)
- Protected routes

### Dashboard
- Overview metrics (equipment, projects, invoices, crew)
- Quick action buttons
- Responsive layout

### Sidebar Navigation
- Mobile-responsive hamburger menu
- Active route highlighting
- Service module navigation

### Form Handling
- Controlled components
- Error states
- Loading states
- Accessibility labels

## Adding New Pages

1. Create component in `src/pages/NewPage.tsx`:

```typescript
function NewPage() {
  return <div className="new-page">...</div>
}
export default NewPage
```

2. Add route in `src/App.tsx`:

```tsx
<Route path="new-page" element={<NewPage />} />
```

3. Add navigation in `src/components/Layout/Sidebar.tsx`:

```typescript
const navItems = [
  { label: 'New Page', href: '/new-page', icon: '📄' },
]
```

## Adding New API Endpoints

1. Add to `src/services/api.ts`:

```typescript
export const newApi = {
  list: () => api.get('/api/v1/new-resource').then(res => res.data),
  getById: (id: string) => api.get(`/api/v1/new-resource/${id}`).then(res => res.data),
  create: (data: unknown) => api.post('/api/v1/new-resource', data).then(res => res.data),
}
```

2. Use in component:

```typescript
const { data } = useQuery({
  queryKey: ['new-resource'],
  queryFn: () => newApi.list(),
})
```

## Performance Optimization

### Code Splitting
Routes are automatically code-split by Vite. Use React.lazy for component-level splitting:

```typescript
const HeavyComponent = React.lazy(() => import('./components/HeavyComponent'))

// Usage with Suspense
<Suspense fallback={<Loading />}>
  <HeavyComponent />
</Suspense>
```

### Image Optimization
- Use WebP format where possible
- Lazy load images below the fold
- Optimize for mobile sizes

### Caching
TanStack Query handles server state caching:
- 5 minute stale time by default
- 10 minute garbage collection time
- Configurable per query

### Bundle Analysis
```bash
npm run build
# Check dist/ folder size
```

## Accessibility (a11y)

- Semantic HTML
- ARIA labels where needed
- Keyboard navigation support
- Focus visible outlines
- Color contrast compliance

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Docker Deployment

The frontend is containerized with multi-stage build:

```bash
# Build
docker build -t rentflow-frontend:latest .

# Run
docker run -p 3000:80 rentflow-frontend:latest
```

Production features:
- Nginx reverse proxy
- Gzip compression
- Security headers
- SPA routing (all non-API routes → index.html)
- Cache-busting for static assets

## Troubleshooting

### Port 3000 already in use
```bash
npm run dev -- --port 3001
```

### CORS errors
Check that API proxy is configured in `vite.config.ts`:
```typescript
proxy: {
  '/api': {
    target: 'http://localhost:80',
    changeOrigin: true,
  }
}
```

### Modules not found
Ensure import paths match tsconfig.json paths:
```typescript
import Component from '@/components/Component'  // Good
import Component from './components/Component'  // Less ideal
```

### TypeScript errors
```bash
npm run type-check
```

## Development Tips

### Hot Module Replacement (HMR)
Vite provides instant HMR for code changes. Just save and browser refreshes automatically.

### React DevTools
Install React DevTools browser extension for component inspection.

### Redux DevTools
For debugging Zustand state, use browser console:
```javascript
// In console
window.__ZUSTAND_DEBUG__ = true
```

### Network Inspection
Use browser DevTools Network tab to inspect API calls and responses.

## Production Checklist

- [ ] Set `VITE_API_BASE_URL` to production API URL
- [ ] Disable debug logging (`VITE_ENABLE_DEBUG=false`)
- [ ] Update `ALLOWED_ORIGINS` in backend for CORS
- [ ] Configure SSL/TLS certificates
- [ ] Set up error tracking (Sentry, etc.)
- [ ] Configure CDN for static assets
- [ ] Set up monitoring and alerts
- [ ] Run security audit (`npm audit`)

## Contributing

1. Follow TypeScript strict mode
2. Use functional components with hooks
3. Keep components focused and testable
4. Use design tokens (SCSS variables)
5. Write semantic HTML
6. Test keyboard navigation
7. Verify dark mode compatibility

## License

Proprietary - RentFlow

## Support

For issues and questions, contact the development team.
