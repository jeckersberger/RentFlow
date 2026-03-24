import { useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'
import { Theme } from '@radix-ui/themes'
import '@radix-ui/themes/styles.css'
import ErrorBoundary from './components/ErrorBoundary/ErrorBoundary'
import InstallPrompt from './components/InstallPrompt/InstallPrompt'
import OfflineIndicator from './components/OfflineIndicator/OfflineIndicator'
import { useAuthStore } from './stores/authStore'
import { useModuleStore } from './stores/moduleStore'
import { useThemeStore, initializeTheme } from './stores/themeStore'
import { setupApi, configApi } from './services/api'
import MainLayout from './components/Layout/MainLayout'
import { ToastContainer } from './components/Toast/Toast'
import { CommandPalette } from './components/CommandPalette/CommandPalette'
import { ShortcutsHelp } from './components/ShortcutsHelp/ShortcutsHelp'
import { KeyboardShortcutsProvider } from './components/KeyboardShortcutsProvider'
import LoginPage from './pages/Login'
import ForgotPasswordPage from './pages/ForgotPassword'
import ResetPasswordPage from './pages/ResetPassword'
import BookingResponsePage from './pages/BookingResponse'
import SetupWizard from './pages/Setup/SetupWizard'
import DashboardPage from './pages/Dashboard'

// Equipment Pages
import EquipmentListPage from './pages/Equipment/EquipmentList'
import EquipmentDetailPage from './pages/Equipment/EquipmentDetail'
import EquipmentFormPage from './pages/Equipment/EquipmentForm'
import EquipmentLabelsPage from './pages/Equipment/EquipmentLabels'
import EquipmentImportPage from './pages/Equipment/EquipmentImport'
import EquipmentTimelinePage from './pages/Equipment/EquipmentTimeline'

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
import InventoryPage from './pages/Warehouse/InventoryPage'

// Settings Pages
import SettingsLayout from './pages/Settings/SettingsLayout'
import CompanyPage from './pages/Settings/pages/CompanyPage'
import UsersPage from './pages/Settings/pages/UsersPage'
import RolesPage from './pages/Settings/pages/RolesPage'
import IntegrationsPage from './pages/Settings/pages/IntegrationsPage'
import BackupsPage from './pages/Settings/pages/BackupsPage'
import LocalePage from './pages/Settings/pages/LocalePage'
import NumberSequencesPage from './pages/Settings/pages/NumberSequencesPage'
import ProjectTypesPage from './pages/Settings/pages/ProjectTypesPage'
import CategoriesPage from './pages/Settings/pages/CategoriesPage'
import CustomFieldsPage from './pages/Settings/pages/CustomFieldsPage'
import EmailSettingsPage from './pages/Settings/pages/EmailSettingsPage'
import DocumentTemplatesPage from './pages/Settings/pages/DocumentTemplatesPage'
import TaxExemptionPage from './pages/Settings/pages/TaxExemptionPage'
import VatSchemesPage from './pages/Settings/pages/VatSchemesPage'
import PaymentTermsPage from './pages/Settings/pages/PaymentTermsPage'
import BankDetailsPage from './pages/Settings/pages/BankDetailsPage'
import TermsConditionsPage from './pages/Settings/pages/TermsConditionsPage'
import ModulesPage from './pages/Settings/pages/ModulesPage'
import LabelSettingsPage from './pages/Settings/pages/LabelSettingsPage'
import EmailAccountsPage from './pages/Settings/pages/EmailAccountsPage'
import NotificationSettingsPage from './pages/Settings/pages/NotificationSettingsPage'

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

// Contacts Pages
import ContactsPage from './pages/Contacts/ContactsPage'

// Quotes Pages
import QuotesPage from './pages/Quotes/QuotesPage'
import QuoteDetailPage from './pages/Quotes/QuoteDetail'

// Real Pages (formerly placeholders)
import CalendarPage from './pages/Calendar/CalendarPage'
import ShortagesPage from './pages/Shortages/ShortagesPage'
import WorkshopPage from './pages/Workshop/WorkshopPage'
import TimeTrackingPage from './pages/TimeTracking/TimeTrackingPage'
import MailInboxPage from './pages/Mail/MailInboxPage'
import MailSentPage from './pages/Mail/MailSentPage'
import MailComposePage from './pages/Mail/MailComposePage'

// Remaining Placeholder Pages
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
import NotFoundPage from './pages/NotFound/NotFoundPage'
import ProfilePage from './pages/Profile/ProfilePage'

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
    setupApi.getStatus()
      .then((status) => {
        if (status.is_completed && location.pathname === '/setup') {
          navigate('/login', { replace: true })
        } else if (!status.is_completed && location.pathname !== '/setup') {
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
  const setEnabledModules = useModuleStore((s) => s.setEnabledModules)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  useEffect(() => {
    initializeTheme()
  }, [])

  // Load module config when authenticated
  useEffect(() => {
    if (isAuthenticated) {
      configApi.get('modules.enabled')
        .then((data) => {
          if (Array.isArray(data)) {
            setEnabledModules(data)
          }
        })
        .catch(() => {
          // API failed — use defaults and mark as loaded so sidebar filters work
          setEnabledModules(['warehouse', 'projects', 'finance'])
        })
    }
  }, [isAuthenticated, setEnabledModules])

  return (
    <ErrorBoundary>
    <QueryClientProvider client={queryClient}>
      <Theme appearance={isDarkMode ? 'dark' : 'light'} accentColor="cyan" grayColor="slate" panelBackground="translucent">
        <BrowserRouter>
          <SetupCheck>
          <Routes>
          <Route path="/setup" element={<SetupWizard />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/forgot-password" element={<ForgotPasswordPage />} />
          <Route path="/reset-password/:token" element={<ResetPasswordPage />} />
          <Route path="/booking/:token" element={<BookingResponsePage />} />

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

            {/* Calendar */}
            <Route path="calendar" element={<CalendarPage />} />

            {/* Equipment Routes */}
            <Route path="equipment" element={<EquipmentListPage />} />
            <Route path="equipment/labels" element={<EquipmentLabelsPage />} />
            <Route path="equipment/import" element={<EquipmentImportPage />} />
            <Route path="equipment/timeline" element={<EquipmentTimelinePage />} />
            <Route path="equipment/new" element={<EquipmentFormPage />} />
            <Route path="equipment/:id" element={<EquipmentDetailPage />} />
            <Route path="equipment/:id/edit" element={<EquipmentFormPage />} />

            {/* Project Routes */}
            <Route path="projects" element={<ProjectListPage />} />
            <Route path="projects/new" element={<ProjectFormPage />} />
            <Route path="projects/:id" element={<ProjectDetailPage />} />
            <Route path="projects/:id/edit" element={<ProjectFormPage />} />

            {/* Shortages */}
            <Route path="shortages" element={<ShortagesPage />} />

            {/* Invoice Routes */}
            <Route path="invoices" element={<InvoiceListPage />} />
            <Route path="invoices/new" element={<InvoiceFormPage />} />
            <Route path="invoices/:id" element={<InvoiceDetailPage />} />
            <Route path="invoices/:id/edit" element={<InvoiceFormPage />} />

            {/* Quotes */}
            <Route path="quotes" element={<QuotesPage />} />
            <Route path="quotes/:id" element={<QuoteDetailPage />} />

            {/* Contacts */}
            <Route path="contacts" element={<ContactsPage />} />

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

            {/* Workshop */}
            <Route path="workshop" element={<WorkshopPage />} />

            {/* Crew Routes */}
            <Route path="crew" element={<CrewPage />} />
            <Route path="crew/:id" element={<CrewDetail />} />
            <Route path="crew/new" element={<NewCrewMemberPage />} />
            <Route path="my-assignments" element={<MyAssignments />} />

            {/* Time Tracking */}
            <Route path="time-tracking" element={<TimeTrackingPage />} />

            {/* Documents Routes */}
            <Route path="documents" element={<DocumentsPage />} />
            <Route path="documents/:id" element={<DocumentDetailPage />} />
            <Route path="documents/new" element={<NewDocumentPage />} />

            {/* Insurance Routes */}
            <Route path="insurance" element={<InsurancePage />} />
            <Route path="insurance/claims/:id" element={<InsuranceDetail />} />
            <Route path="insurance/new-claim" element={<NewClaimPage />} />

            {/* Mail */}
            <Route path="mail/inbox" element={<MailInboxPage />} />
            <Route path="mail/sent" element={<MailSentPage />} />
            <Route path="mail/compose" element={<MailComposePage />} />

            {/* Reports Routes */}
            <Route path="reports" element={<ReportsPage />} />

            {/* Scanner & Warehouse */}
            <Route path="scanner" element={<ScannerPage />} />
            <Route path="warehouse" element={<WarehouseViewPage />} />
            <Route path="inventory" element={<InventoryPage />} />

            {/* Settings - Nested Routes */}
            <Route path="settings" element={<SettingsLayout />}>
              <Route index element={<CompanyPage />} />
              <Route path="company" element={<CompanyPage />} />
              <Route path="users" element={<UsersPage />} />
              <Route path="roles" element={<RolesPage />} />
              <Route path="integrations" element={<IntegrationsPage />} />
              <Route path="backups" element={<BackupsPage />} />
              <Route path="locale" element={<LocalePage />} />
              <Route path="number-sequences" element={<NumberSequencesPage />} />
              <Route path="project-types" element={<ProjectTypesPage />} />
              <Route path="categories" element={<CategoriesPage />} />
              <Route path="custom-fields" element={<CustomFieldsPage />} />
              <Route path="email" element={<EmailSettingsPage />} />
              <Route path="document-templates" element={<DocumentTemplatesPage />} />
              <Route path="tax-exemption" element={<TaxExemptionPage />} />
              <Route path="vat-schemes" element={<VatSchemesPage />} />
              <Route path="payment-terms" element={<PaymentTermsPage />} />
              <Route path="bank-details" element={<BankDetailsPage />} />
              <Route path="terms-conditions" element={<TermsConditionsPage />} />
              <Route path="modules" element={<ModulesPage />} />
              <Route path="labels" element={<LabelSettingsPage />} />
              <Route path="email-accounts" element={<EmailAccountsPage />} />
              <Route path="notifications" element={<NotificationSettingsPage />} />
            </Route>

            {/* Phase 4 Routes */}
            <Route path="ai" element={<AIPage />} />
            <Route path="workflows" element={<WorkflowsPage />} />
            <Route path="federation" element={<FederationPage />} />
            <Route path="audit" element={<AuditPage />} />
            <Route path="admin" element={<AdminPage />} />

            {/* Profile */}
            <Route path="profile" element={<ProfilePage />} />

            {/* 404 Catch-all */}
            <Route path="*" element={<NotFoundPage />} />
          </Route>
        </Routes>
          </SetupCheck>
        <ToastContainer />
        <CommandPalette />
        <KeyboardShortcutsProvider />
        <ShortcutsHelp />
      </BrowserRouter>
      <OfflineIndicator />
      <InstallPrompt />
      </Theme>
    </QueryClientProvider>
    </ErrorBoundary>
  )
}

export default App
