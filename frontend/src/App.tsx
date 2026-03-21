import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'
import { Theme } from '@radix-ui/themes'
import '@radix-ui/themes/styles.css'
import { useAuthStore } from './stores/authStore'
import { initializeTheme } from './stores/themeStore'
import MainLayout from './components/Layout/MainLayout'
import { ToastContainer } from './components/Toast/Toast'
import { CommandPalette } from './components/CommandPalette/CommandPalette'
import LoginPage from './pages/Login'
import DashboardPage from './pages/Dashboard'

// Equipment Pages
import EquipmentListPage from './pages/Equipment/EquipmentList'
import EquipmentDetailPage from './pages/Equipment/EquipmentDetail'
import EquipmentFormPage from './pages/Equipment/EquipmentForm'

// Project Pages
import ProjectListPage from './pages/Projects/ProjectList'
import ProjectDetailPage from './pages/Projects/ProjectDetail'

// Invoice Pages
import InvoiceListPage from './pages/Invoices/InvoiceList'
import InvoiceDetailPage from './pages/Invoices/InvoiceDetail'

// Scanner Page
import ScannerPage from './pages/Scanner/ScannerPage'

// Warehouse Page
import WarehouseViewPage from './pages/Warehouse/WarehousePage'

// Settings Page
import SettingsPage from './pages/Settings/SettingsPage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10, // 10 minutes (formerly cacheTime)
    },
  },
})

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function App() {
  useEffect(() => {
    initializeTheme()
  }, [])

  const isDarkMode = document.documentElement.classList.contains('dark-mode')

  return (
    <QueryClientProvider client={queryClient}>
      <Theme appearance={isDarkMode ? 'dark' : 'light'} accentColor="blue" grayColor="slate">
        <BrowserRouter>
          <Routes>
          <Route path="/login" element={<LoginPage />} />

          <Route
            path="/*"
            element={
              <PrivateRoute>
                <MainLayout />
              </PrivateRoute>
            }
          >
            {/* Dashboard */}
            <Route index element={<DashboardPage />} />
            <Route path="dashboard" element={<DashboardPage />} />

            {/* Equipment Routes */}
            <Route path="equipment" element={<EquipmentListPage />} />
            <Route path="equipment/new" element={<EquipmentFormPage />} />
            <Route path="equipment/:id" element={<EquipmentDetailPage />} />
            <Route path="equipment/:id/edit" element={<EquipmentFormPage />} />

            {/* Project Routes */}
            <Route path="projects" element={<ProjectListPage />} />
            <Route path="projects/new" element={<div>Neues Projekt - Formular</div>} />
            <Route path="projects/:id" element={<ProjectDetailPage />} />
            <Route path="projects/:id/edit" element={<div>Projekt bearbeiten - Formular</div>} />

            {/* Invoice Routes */}
            <Route path="invoices" element={<InvoiceListPage />} />
            <Route path="invoices/new" element={<div>Neue Rechnung - Formular</div>} />
            <Route path="invoices/:id" element={<InvoiceDetailPage />} />
            <Route path="invoices/:id/edit" element={<div>Rechnung bearbeiten - Formular</div>} />

            {/* Scanner & Warehouse */}
            <Route path="scanner" element={<ScannerPage />} />
            <Route path="warehouse" element={<WarehouseViewPage />} />

            {/* Settings */}
            <Route path="settings" element={<SettingsPage />} />
          </Route>
        </Routes>
        <ToastContainer />
        <CommandPalette />
      </BrowserRouter>
      </Theme>
    </QueryClientProvider>
  )
}

export default App
