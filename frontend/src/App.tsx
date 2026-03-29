import React, { Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import { useAuthStore } from '@/stores/authStore';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

// Lazy-loaded pages
const LoginPage = React.lazy(() => import('@/pages/Login/LoginPage'));
const DashboardPage = React.lazy(() => import('@/pages/Dashboard/DashboardPage'));
const EquipmentList = React.lazy(() => import('@/pages/Equipment/EquipmentList'));
const EquipmentDetail = React.lazy(() => import('@/pages/Equipment/EquipmentDetail'));
const EquipmentForm = React.lazy(() => import('@/pages/Equipment/EquipmentForm'));
const ProjectList = React.lazy(() => import('@/pages/Projects/ProjectList'));
const ProjectDetail = React.lazy(() => import('@/pages/Projects/ProjectDetail'));
const ProjectForm = React.lazy(() => import('@/pages/Projects/ProjectForm'));
const CustomerList = React.lazy(() => import('@/pages/Customers/CustomerList'));
const CustomerDetail = React.lazy(() => import('@/pages/Customers/CustomerDetail'));
const CustomerForm = React.lazy(() => import('@/pages/Customers/CustomerForm'));
const InvoiceList = React.lazy(() => import('@/pages/Invoices/InvoiceList'));
const InvoiceDetail = React.lazy(() => import('@/pages/Invoices/InvoiceDetail'));
const InvoiceForm = React.lazy(() => import('@/pages/Invoices/InvoiceForm'));
const ScannerPage = React.lazy(() => import('@/pages/Scanner/ScannerPage'));
const WarehouseList = React.lazy(() => import('@/pages/Warehouse/WarehouseList'));
const WarehouseMovements = React.lazy(() => import('@/pages/Warehouse/WarehouseMovements'));
const WarehouseInventory = React.lazy(() => import('@/pages/Warehouse/WarehouseInventory'));
const CrewList = React.lazy(() => import('@/pages/Crew/CrewList'));
const DocumentList = React.lazy(() => import('@/pages/Documents/DocumentList'));
const ExpenseList = React.lazy(() => import('@/pages/Expenses/ExpenseList'));
const ReportingList = React.lazy(() => import('@/pages/Reporting/ReportingList'));
const MaintenanceList = React.lazy(() => import('@/pages/Maintenance/MaintenanceList'));
const TransportList = React.lazy(() => import('@/pages/Transport/TransportList'));
const QuoteList = React.lazy(() => import('@/pages/Quotes/QuoteList'));
const QuoteDetail = React.lazy(() => import('@/pages/Quotes/QuoteDetail'));
const QuoteForm = React.lazy(() => import('@/pages/Quotes/QuoteForm'));
const AccountPage = React.lazy(() => import('@/pages/Account/AccountPage'));

// Layout (already exists with Sidebar + Header)
const AppLayout = React.lazy(() => import('@/components/Layout/AppLayout'));

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

function LoadingFallback() {
  return (
    <div className="loading-fallback">
      <div className="loading-spinner" />
    </div>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Suspense fallback={<LoadingFallback />}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />

            <Route
              path="/*"
              element={
                <ProtectedRoute>
                  <AppLayout />
                </ProtectedRoute>
              }
            >
              <Route index element={<DashboardPage />} />
              <Route path="equipment" element={<EquipmentList />} />
              <Route path="equipment/new" element={<EquipmentForm />} />
              <Route path="equipment/:id" element={<EquipmentDetail />} />
              <Route path="equipment/:id/edit" element={<EquipmentForm />} />
              <Route path="projects" element={<ProjectList />} />
              <Route path="projects/new" element={<ProjectForm />} />
              <Route path="projects/:id" element={<ProjectDetail />} />
              <Route path="projects/:id/edit" element={<ProjectForm />} />
              <Route path="customers" element={<CustomerList />} />
              <Route path="customers/new" element={<CustomerForm />} />
              <Route path="customers/:id" element={<CustomerDetail />} />
              <Route path="customers/:id/edit" element={<CustomerForm />} />
              <Route path="quotes" element={<QuoteList />} />
              <Route path="quotes/new" element={<QuoteForm />} />
              <Route path="quotes/:id" element={<QuoteDetail />} />
              <Route path="quotes/:id/edit" element={<QuoteForm />} />
              <Route path="invoices" element={<InvoiceList />} />
              <Route path="invoices/new" element={<InvoiceForm />} />
              <Route path="invoices/:id" element={<InvoiceDetail />} />
              <Route path="invoices/:id/edit" element={<InvoiceForm />} />
              <Route path="scanner" element={<ScannerPage />} />
              <Route path="warehouse" element={<WarehouseList />} />
              <Route path="warehouse/movements" element={<WarehouseMovements />} />
              <Route path="warehouse/inventory" element={<WarehouseInventory />} />
              <Route path="crew" element={<CrewList />} />
              <Route path="documents" element={<DocumentList />} />
              <Route path="expenses" element={<ExpenseList />} />
              <Route path="reporting" element={<ReportingList />} />
              <Route path="maintenance" element={<MaintenanceList />} />
              <Route path="transport" element={<TransportList />} />
              <Route path="account" element={<AccountPage />} />
            </Route>
          </Routes>
        </Suspense>
      </BrowserRouter>

      <Toaster
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: '#1a1d2e',
            color: '#f1f5f9',
            border: '1px solid rgba(255, 255, 255, 0.08)',
          },
        }}
      />
    </QueryClientProvider>
  );
}
