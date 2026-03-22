import { useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'
import { Theme } from '@radix-ui/themes'
import '@radix-ui/themes/styles.css'
import { useAuthStore } from './stores/authStore'
import { useThemeStore, initializeTheme } from './stores/themeStore'
import { setupApi } from './services/api'
import MainLayout from './components/Layout/MainLayout'
import { ToastContainer } from './components/Toast/Toast'
import { CommandPalette } from './components/CommandPalette/CommandPalette'
import LoginPage from './pages/Login'
import SetupWizard from './pages/Setup/SetupWizard'
import DashboardPage from './pages/Dashboard'

// Equipment Pages
import EquipmentListPage from './pages/Equipment/EquipmentList'
import EquipmentDetailPage from './pages/Equipment/EquipmentDetail'
import EquipmentFormPage from './pages/Equipment/EquipmentForm'
import EquipmentLabelsPage from './pages/Equipment/EquipmentLabels'

// Project Pages
import ProjectListPage from './pages/Projects/ProjectList'
import ProjectDetailPage from './pages/Projects/ProjectDetail'
import ProjectFormPage from './pages/Projects/ProjectForm'

// Invoice Pages
import InvoiceListPage from './pages/Invoices/InvoiceList'
import InvoiceDetailPage from './pages/Invoices/InvoiceDetail'
import InvoiceFormPage from './pages/Invoices/InvoiceForm'

// Scanner Page
import ScannerPage from './pages/Scanner/ScannerPage'

// Warehouse Page
import WarehouseViewPage from './pages/Warehouse/WarehousePage'

// Settings Page
import SettingsPage from './pages/Settings/SettingsPage'

// Transport Pages
import TransportPage from './pages/Transport/TransportPage'
import TransportDetailPage from './pages/Transport/TransportDetail'

// Maintenance Pages
import MaintenancePage from './pages/Maintenance/MaintenancePage'
import MaintenanceDetailPage from './pages/Maintenance/MaintenanceDetail'

// Crew Pages
import CrewPage from './pages/Crew/CrewPage'
import CrewDetail from './pages/Crew/CrewDetail'
import MyAssignments from './pages/Crew/MyAssignments'

// Documents Pages
import DocumentsPage from './pages/Documents/DocumentsPage'

// Insurance Pages
import InsurancePage from './pages/Insurance/InsurancePage'
import InsuranceDetail from './pages/Insurance/InsuranceDetail'

// Reports Pages
import ReportsPage from './pages/Reports/ReportsPage'

// Placeholder Pages
import {
  NewTourPage,
  NewVehiclePage,
  VehicleDetailPage,
  NewMaintenanceTaskPage,
  NewECheckPage,
  MaintenancePlanDetailPage,
  NewCrewMemberPage,
  DocumentDetailPage,
  NewDocumentPage,
  NewClaimPage,
} from './pages/PlaceholderPages'

// Phase 4 Pages
import AIPage from './pages/AI/AIPage'
import WorkflowsPage from './pages/Workflows/WorkflowsPage'
import FederationPage from './pages/Federation/FederationPage'
import AuditPage from './pages/Audit/AuditPage'
import AdminPage from './pages/Admin/AdminPage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10, // 10 minutes (formerly cacheTime)
    },
  },
})

function SetupCheck({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate()
  const location = useLocation()
  const [checked, setChecked] = useState(false)

  useEffect(() => {
    if (location.pathname === '/setup') {
      setChecked(true)
      return
    }
    setupApi.getStatus()
      .then((status) => {
        if (!status.is_completed) {
          navigate('/setup', { replace: true })
        }
        setChecked(true)
      })
      .catch(() => {
        setChecked(true)
      })
  }, [navigate, location.pathname])

  if (!checked) return null
  return <>{children}</>
}

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function App() {
  const isDarkMode = useThemeStore((s) => s.isDarkMode)

  useEffect(() => {
    initializeTheme()
  }, [])

  return (
    <QueryClientProvider client={queryClient}>
      <Theme appearance={isDarkMode ? 'dark' : 'light'} accentColor="cyan" grayColor="slate" panelBackground="translucent">
        <BrowserRouter>
          <SetupCheck>
          <Routes>
          <Route path="/setup" element={<SetupWizard />} />
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
            <Route path="equipment/labels" element={<EquipmentLabelsPage />} />
            <Route path="equipment/new" element={<EquipmentFormPage />} />
            <Route path="equipment/:id" element={<EquipmentDetailPage />} />
            <Route path="equipment/:id/edit" element={<EquipmentFormPage />} />

            {/* Project Routes */}
            <Route path="projects" element={<ProjectListPage />} />
            <Route path="projects/new" element={<ProjectFormPage />} />
            <Route path="projects/:id" element={<ProjectDetailPage />} />
            <Route path="projects/:id/edit" element={<ProjectFormPage />} />

            {/* Invoice Routes */}
            <Route path="invoices" element={<InvoiceListPage />} />
            <Route path="invoices/new" element={<InvoiceFormPage />} />
            <Route path="invoices/:id" element={<InvoiceDetailPage />} />
            <Route path="invoices/:id/edit" element={<InvoiceFormPage />} />

            {/* Transport Routes */}
            <Route path="transport" element={<TransportPage />} />
            <Route path="transport/tours/:id" element={<TransportDetailPage />} />
            <Route path="transport/new-tour" element={<NewTourPage />} />
            <Route path="transport/new-vehicle" element={<NewVehiclePage />} />
            <Route path="transport/vehicles/:id" element={<VehicleDetailPage />} />

            {/* Maintenance Routes */}
            <Route path="maintenance" element={<MaintenancePage />} />
            <Route path="maintenance/tasks/:id" element={<MaintenanceDetailPage />} />
            <Route path="maintenance/new-task" element={<NewMaintenanceTaskPage />} />
            <Route path="maintenance/new-echeck" element={<NewECheckPage />} />
            <Route path="maintenance/plans/:id" element={<MaintenancePlanDetailPage />} />

            {/* Crew Routes */}
            <Route path="crew" element={<CrewPage />} />
            <Route path="crew/:id" element={<CrewDetail />} />
            <Route path="crew/new" element={<NewCrewMemberPage />} />
            <Route path="my-assignments" element={<MyAssignments />} />

            {/* Documents Routes */}
            <Route path="documents" element={<DocumentsPage />} />
            <Route path="documents/:id" element={<DocumentDetailPage />} />
            <Route path="documents/new" element={<NewDocumentPage />} />

            {/* Insurance Routes */}
            <Route path="insurance" element={<InsurancePage />} />
            <Route path="insurance/claims/:id" element={<InsuranceDetail />} />
            <Route path="insurance/new-claim" element={<NewClaimPage />} />

            {/* Reports Routes */}
            <Route path="reports" element={<ReportsPage />} />

            {/* Scanner & Warehouse */}
            <Route path="scanner" element={<ScannerPage />} />
            <Route path="warehouse" element={<WarehouseViewPage />} />

            {/* Settings */}
            <Route path="settings" element={<SettingsPage />} />

            {/* Phase 4 Routes */}
            <Route path="ai" element={<AIPage />} />
            <Route path="workflows" element={<WorkflowsPage />} />
            <Route path="federation" element={<FederationPage />} />
            <Route path="audit" element={<AuditPage />} />
            <Route path="admin" element={<AdminPage />} />
          </Route>
        </Routes>
          </SetupCheck>
        <ToastContainer />
        <CommandPalette />
      </BrowserRouter>
      </Theme>
    </QueryClientProvider>
  )
}

export default App
