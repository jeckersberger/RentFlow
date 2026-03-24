import { lazy, Suspense, useEffect, useState } from 'react'
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

// Eagerly loaded pages (first pages users see — no lazy loading)
import LoginPage from './pages/Login'
import ForgotPasswordPage from './pages/ForgotPassword'
import ResetPasswordPage from './pages/ResetPassword'
import BookingResponsePage from './pages/BookingResponse'
import SetupWizard from './pages/Setup/SetupWizard'

// Lazy-loaded pages (code-split for faster initial load)
const DashboardPage = lazy(() => import('./pages/Dashboard'))

// Equipment Pages
const EquipmentListPage = lazy(() => import('./pages/Equipment/EquipmentList'))
const EquipmentDetailPage = lazy(() => import('./pages/Equipment/EquipmentDetail'))
const EquipmentFormPage = lazy(() => import('./pages/Equipment/EquipmentForm'))
const EquipmentLabelsPage = lazy(() => import('./pages/Equipment/EquipmentLabels'))
const EquipmentImportPage = lazy(() => import('./pages/Equipment/EquipmentImport'))
const EquipmentTimelinePage = lazy(() => import('./pages/Equipment/EquipmentTimeline'))

// Project Pages
const ProjectListPage = lazy(() => import('./pages/Projects/ProjectList'))
const ProjectDetailPage = lazy(() => import('./pages/Projects/ProjectDetail'))
const ProjectFormPage = lazy(() => import('./pages/Projects/ProjectForm'))

// Invoice Pages
const InvoiceListPage = lazy(() => import('./pages/Invoices/InvoiceList'))
const InvoiceDetailPage = lazy(() => import('./pages/Invoices/InvoiceDetail'))
const InvoiceEditor = lazy(() => import('./pages/Invoices/InvoiceEditor'))

// Scanner Page
const ScannerPage = lazy(() => import('./pages/Scanner/ScannerPage'))

// Warehouse Page
const WarehouseViewPage = lazy(() => import('./pages/Warehouse/WarehousePage'))
const InventoryPage = lazy(() => import('./pages/Warehouse/InventoryPage'))

// Settings Pages
const SettingsLayout = lazy(() => import('./pages/Settings/SettingsLayout'))
const CompanyPage = lazy(() => import('./pages/Settings/pages/CompanyPage'))
const UsersPage = lazy(() => import('./pages/Settings/pages/UsersPage'))
const RolesPage = lazy(() => import('./pages/Settings/pages/RolesPage'))
const IntegrationsPage = lazy(() => import('./pages/Settings/pages/IntegrationsPage'))
const BackupsPage = lazy(() => import('./pages/Settings/pages/BackupsPage'))
const LocalePage = lazy(() => import('./pages/Settings/pages/LocalePage'))
const NumberSequencesPage = lazy(() => import('./pages/Settings/pages/NumberSequencesPage'))
const ProjectTypesPage = lazy(() => import('./pages/Settings/pages/ProjectTypesPage'))
const CategoriesPage = lazy(() => import('./pages/Settings/pages/CategoriesPage'))
const CustomFieldsPage = lazy(() => import('./pages/Settings/pages/CustomFieldsPage'))
const EmailSettingsPage = lazy(() => import('./pages/Settings/pages/EmailSettingsPage'))
const DocumentTemplatesPage = lazy(() => import('./pages/Settings/pages/DocumentTemplatesPage'))
const TaxExemptionPage = lazy(() => import('./pages/Settings/pages/TaxExemptionPage'))
const VatSchemesPage = lazy(() => import('./pages/Settings/pages/VatSchemesPage'))
const PaymentTermsPage = lazy(() => import('./pages/Settings/pages/PaymentTermsPage'))
const BankDetailsPage = lazy(() => import('./pages/Settings/pages/BankDetailsPage'))
const TermsConditionsPage = lazy(() => import('./pages/Settings/pages/TermsConditionsPage'))
const ModulesPage = lazy(() => import('./pages/Settings/pages/ModulesPage'))
const LabelSettingsPage = lazy(() => import('./pages/Settings/pages/LabelSettingsPage'))
const EmailAccountsPage = lazy(() => import('./pages/Settings/pages/EmailAccountsPage'))
const NotificationSettingsPage = lazy(() => import('./pages/Settings/pages/NotificationSettingsPage'))
const ScannerDevicesPage = lazy(() => import('./pages/Settings/pages/ScannerDevicesPage'))

// Transport Pages
const TransportPage = lazy(() => import('./pages/Transport/TransportPage'))
const TransportDetailPage = lazy(() => import('./pages/Transport/TransportDetail'))

// Maintenance Pages
const MaintenancePage = lazy(() => import('./pages/Maintenance/MaintenancePage'))
const MaintenanceDetailPage = lazy(() => import('./pages/Maintenance/MaintenanceDetail'))

// Crew Pages
const CrewPage = lazy(() => import('./pages/Crew/CrewPage'))
const CrewDetail = lazy(() => import('./pages/Crew/CrewDetail'))
const MyAssignments = lazy(() => import('./pages/Crew/MyAssignments'))

// Documents Pages
const DocumentsPage = lazy(() => import('./pages/Documents/DocumentsPage'))

// Insurance Pages
const InsurancePage = lazy(() => import('./pages/Insurance/InsurancePage'))
const InsuranceDetail = lazy(() => import('./pages/Insurance/InsuranceDetail'))

// Reports Pages
const ReportsPage = lazy(() => import('./pages/Reports/ReportsPage'))

// Contacts Pages
const ContactsPage = lazy(() => import('./pages/Contacts/ContactsPage'))

// Expenses Pages
const ExpensesPage = lazy(() => import('./pages/Expenses/ExpensesPage'))

// Quotes Pages
const QuotesPage = lazy(() => import('./pages/Quotes/QuotesPage'))
const QuoteNewPage = lazy(() => import('./pages/Quotes/QuoteNew'))
const QuoteDetailPage = lazy(() => import('./pages/Quotes/QuoteDetail'))

// Real Pages (formerly placeholders)
const CalendarPage = lazy(() => import('./pages/Calendar/CalendarPage'))
const ShortagesPage = lazy(() => import('./pages/Shortages/ShortagesPage'))
const WorkshopPage = lazy(() => import('./pages/Workshop/WorkshopPage'))
const TimeTrackingPage = lazy(() => import('./pages/TimeTracking/TimeTrackingPage'))
const MailInboxPage = lazy(() => import('./pages/Mail/MailInboxPage'))
const MailSentPage = lazy(() => import('./pages/Mail/MailSentPage'))
const MailComposePage = lazy(() => import('./pages/Mail/MailComposePage'))

// Remaining Placeholder Pages — lazy-loaded as a single chunk
const NewTourPage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.NewTourPage })))
const NewVehiclePage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.NewVehiclePage })))
const VehicleDetailPage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.VehicleDetailPage })))
const NewMaintenanceTaskPage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.NewMaintenanceTaskPage })))
const NewECheckPage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.NewECheckPage })))
const MaintenancePlanDetailPage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.MaintenancePlanDetailPage })))
const NewCrewMemberPage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.NewCrewMemberPage })))
const DocumentDetailPage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.DocumentDetailPage })))
const NewDocumentPage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.NewDocumentPage })))
const NewClaimPage = lazy(() => import('./pages/PlaceholderPages').then(m => ({ default: m.NewClaimPage })))

// Phase 4 Pages
const AIPage = lazy(() => import('./pages/AI/AIPage'))
const WorkflowsPage = lazy(() => import('./pages/Workflows/WorkflowsPage'))
const FederationPage = lazy(() => import('./pages/Federation/FederationPage'))
const AuditPage = lazy(() => import('./pages/Audit/AuditPage'))
const AdminPage = lazy(() => import('./pages/Admin/AdminPage'))
const NotFoundPage = lazy(() => import('./pages/NotFound/NotFoundPage'))
const ProfilePage = lazy(() => import('./pages/Profile/ProfilePage'))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10, // 10 minutes (formerly cacheTime)
      refetchOnWindowFocus: false,
      retry: 1,
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
          <Suspense fallback={<div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', color: 'var(--gray-11)' }}>Laden...</div>}>
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
            <Route path="invoices/new" element={<InvoiceEditor />} />
            <Route path="invoices/:id" element={<InvoiceDetailPage />} />
            <Route path="invoices/:id/edit" element={<InvoiceEditor />} />

            {/* Quotes */}
            <Route path="quotes" element={<QuotesPage />} />
            <Route path="quotes/new" element={<QuoteNewPage />} />
            <Route path="quotes/:id" element={<QuoteDetailPage />} />

            {/* Contacts */}
            <Route path="contacts" element={<ContactsPage />} />

            {/* Expenses */}
            <Route path="expenses" element={<ExpensesPage />} />

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
              <Route path="scanner-devices" element={<ScannerDevicesPage />} />
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
          </Suspense>
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
