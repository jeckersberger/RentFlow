import axios, { AxiosInstance } from 'axios'
import { useAuthStore } from '../stores/authStore'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:80'

// Create axios instance
export const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor - add JWT token
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor - handle 401 and refresh token
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// ============================================================================
// MOCK MODE - enables frontend without backend
// ============================================================================
const MOCK_MODE = !import.meta.env.VITE_API_BASE_URL || import.meta.env.VITE_MOCK === 'true'

const mockDelay = <T>(data: T, ms = 300): Promise<T> =>
  new Promise((resolve) => setTimeout(() => resolve(data), ms))

// Auth API endpoints
export const authApi = {
  login: (email: string, password: string) => {
    if (MOCK_MODE) {
      if (email === 'admin@example.com' && password === 'password') {
        return mockDelay({
          token: 'mock-jwt-token-rentflow-2026',
          user: { id: '1', email: 'admin@example.com', name: 'Marco Berger' },
        })
      }
      return Promise.reject({ response: { status: 401 } })
    }
    return api.post('/api/v1/auth/login', { email, password }).then(res => res.data)
  },

  register: (email: string, password: string, name: string) =>
    api.post('/api/v1/auth/register', { email, password, name }).then(res => res.data),

  logout: () =>
    MOCK_MODE ? mockDelay({ success: true }) : api.post('/api/v1/auth/logout').then(res => res.data),

  getCurrentUser: () =>
    MOCK_MODE
      ? mockDelay({ id: '1', email: 'admin@example.com', name: 'Marco Berger' })
      : api.get('/api/v1/auth/me').then(res => res.data),

  refreshToken: () =>
    api.post('/api/v1/auth/refresh').then(res => res.data),
}

// ============================================================================
// MOCK DATA
// ============================================================================
const mockEquipment = [
  { id: '1', name: 'JBL VTX A12', sku: 'SPK-001', barcode: 'RF-SPK-001', category: 'audio', status: 'available', location: 'Lager A - Regal 1', description: 'Line Array Lautsprecher, 2x 12"', quantity: 8, price_daily: 120, price_weekly: 600, price_monthly: 1800, created_at: '2025-09-15T10:00:00Z', updated_at: '2026-03-01T14:30:00Z' },
  { id: '2', name: 'Shure SM58', sku: 'MIC-001', barcode: 'RF-MIC-001', category: 'audio', status: 'available', location: 'Lager A - Regal 3', description: 'Dynamisches Gesangsmikrofon', quantity: 24, price_daily: 15, price_weekly: 75, price_monthly: 200, created_at: '2025-09-15T10:00:00Z', updated_at: '2026-02-20T09:00:00Z' },
  { id: '3', name: 'MA Lighting grandMA3', sku: 'LGT-001', barcode: 'RF-LGT-001', category: 'lighting', status: 'checked_out', location: 'Projekt: Stadtfest München', description: 'Lichtmischpult, Full-Size', quantity: 2, price_daily: 250, price_weekly: 1200, price_monthly: 3500, created_at: '2025-10-01T08:00:00Z', updated_at: '2026-03-18T16:00:00Z' },
  { id: '4', name: 'Martin MAC Aura XB', sku: 'LGT-002', barcode: 'RF-LGT-002', category: 'lighting', status: 'checked_out', location: 'Projekt: Stadtfest München', description: 'LED Moving Head Wash', quantity: 16, price_daily: 85, price_weekly: 400, price_monthly: 1200, created_at: '2025-10-01T08:00:00Z', updated_at: '2026-03-18T16:00:00Z' },
  { id: '5', name: 'Blackmagic ATEM Mini Extreme', sku: 'VID-001', barcode: 'RF-VID-001', category: 'video', status: 'available', location: 'Lager B - Regal 1', description: 'Video-Mischpult, 8 HDMI Inputs', quantity: 3, price_daily: 180, price_weekly: 850, price_monthly: 2500, created_at: '2025-11-10T12:00:00Z', updated_at: '2026-01-15T11:00:00Z' },
  { id: '6', name: 'Prolyte X30V Truss 3m', sku: 'STG-001', barcode: 'RF-STG-001', category: 'stage', status: 'available', location: 'Lager C - Boden', description: 'Traversensystem, Aluminium', quantity: 40, price_daily: 25, price_weekly: 120, price_monthly: 350, created_at: '2025-09-20T09:00:00Z', updated_at: '2026-03-10T08:00:00Z' },
  { id: '7', name: 'Yamaha CL5', sku: 'AUD-002', barcode: 'RF-AUD-002', category: 'audio', status: 'maintenance', location: 'Werkstatt', description: 'Digitales Mischpult, 72 Kanäle', quantity: 1, price_daily: 350, price_weekly: 1600, price_monthly: 4500, created_at: '2025-09-15T10:00:00Z', updated_at: '2026-03-20T10:00:00Z' },
  { id: '8', name: 'Canon EOS R5', sku: 'VID-002', barcode: 'RF-VID-002', category: 'video', status: 'reserved', location: 'Lager B - Regal 2', description: 'Vollformat Kamera, 8K Video', quantity: 4, price_daily: 200, price_weekly: 950, price_monthly: 2800, created_at: '2025-12-01T14:00:00Z', updated_at: '2026-03-19T09:00:00Z' },
  { id: '9', name: 'Chainmaster BGV-D8+ 1t', sku: 'STG-002', barcode: 'RF-STG-002', category: 'stage', status: 'available', location: 'Lager C - Regal 1', description: 'Kettenzug, 1 Tonne, BGV-D8+', quantity: 12, price_daily: 45, price_weekly: 200, price_monthly: 600, created_at: '2025-10-15T11:00:00Z', updated_at: '2026-02-28T13:00:00Z' },
  { id: '10', name: 'Robe MegaPointe', sku: 'LGT-003', barcode: 'RF-LGT-003', category: 'lighting', status: 'available', location: 'Lager A - Regal 5', description: 'Moving Head Spot/Beam/Wash', quantity: 20, price_daily: 95, price_weekly: 450, price_monthly: 1300, created_at: '2025-11-20T10:00:00Z', updated_at: '2026-03-05T15:00:00Z' },
]

const mockProjects = [
  { id: '1', name: 'Stadtfest München 2026', client: 'Stadt München', status: 'active', start_date: '2026-04-15', end_date: '2026-04-18', location: 'Marienplatz, München', budget: 45000, description: 'Jährliches Stadtfest mit 3 Bühnen, Full-Service Veranstaltungstechnik', created_at: '2026-01-10T09:00:00Z' },
  { id: '2', name: 'Firmen-Gala TechCorp', client: 'TechCorp GmbH', status: 'planning', start_date: '2026-05-20', end_date: '2026-05-20', location: 'Hilton Hotel, Frankfurt', budget: 18500, description: 'Firmengala für 500 Gäste, Ton/Licht/Video', created_at: '2026-02-15T14:00:00Z' },
  { id: '3', name: 'Open Air Festival Bodensee', client: 'Festival GmbH', status: 'planning', start_date: '2026-07-10', end_date: '2026-07-12', location: 'Seepark, Konstanz', budget: 85000, description: '3-Tages Festival, 2 Stages, Camping', created_at: '2026-02-20T11:00:00Z' },
  { id: '4', name: 'Produktlaunch AutoBrand', client: 'AutoBrand AG', status: 'completed', start_date: '2026-02-28', end_date: '2026-02-28', location: 'Messe Stuttgart', budget: 32000, description: 'Produktpräsentation neues E-Auto, LED-Wall + Audio', created_at: '2025-12-05T10:00:00Z' },
  { id: '5', name: 'Hochzeit Familie Weber', client: 'Familie Weber', status: 'completed', start_date: '2026-03-08', end_date: '2026-03-08', location: 'Schloss Neuschwanstein', budget: 8500, description: 'Hochzeitsfeier, DJ-Setup + Beleuchtung', created_at: '2026-01-20T16:00:00Z' },
]

const mockInvoices = [
  { id: '1', invoice_number: 'RF-2026-001', client: 'AutoBrand AG', project_name: 'Produktlaunch AutoBrand', status: 'paid', amount: 32000, tax: 6080, total: 38080, issue_date: '2026-03-01', due_date: '2026-03-31', paid_date: '2026-03-15', items: [{ description: 'LED-Wall 6x3m (3 Tage)', quantity: 1, price: 4500, total: 4500 }, { description: 'Line Array System (3 Tage)', quantity: 2, price: 3600, total: 7200 }, { description: 'Lichttechnik Paket', quantity: 1, price: 8500, total: 8500 }, { description: 'Techniker (3 Tage)', quantity: 4, price: 2950, total: 11800 }] },
  { id: '2', invoice_number: 'RF-2026-002', client: 'Familie Weber', project_name: 'Hochzeit Familie Weber', status: 'paid', amount: 8500, tax: 1615, total: 10115, issue_date: '2026-03-09', due_date: '2026-04-09', paid_date: '2026-03-20', items: [{ description: 'DJ-Setup komplett', quantity: 1, price: 3500, total: 3500 }, { description: 'Ambientebeleuchtung', quantity: 1, price: 2800, total: 2800 }, { description: 'Techniker', quantity: 2, price: 1100, total: 2200 }] },
  { id: '3', invoice_number: 'RF-2026-003', client: 'Stadt München', project_name: 'Stadtfest München 2026', status: 'draft', amount: 45000, tax: 8550, total: 53550, issue_date: '2026-03-22', due_date: '2026-04-22', paid_date: null, items: [{ description: 'PA-System Hauptbühne', quantity: 1, price: 12000, total: 12000 }, { description: 'Lichttechnik 3 Bühnen', quantity: 1, price: 15000, total: 15000 }, { description: 'Bühne + Truss', quantity: 1, price: 8000, total: 8000 }, { description: 'Techniker-Team (4 Tage)', quantity: 8, price: 1250, total: 10000 }] },
  { id: '4', invoice_number: 'RF-2026-004', client: 'TechCorp GmbH', project_name: 'Firmen-Gala TechCorp', status: 'sent', amount: 18500, tax: 3515, total: 22015, issue_date: '2026-03-18', due_date: '2026-04-18', paid_date: null, items: [{ description: 'Audio-Paket Gala', quantity: 1, price: 6500, total: 6500 }, { description: 'Licht-Design Gala', quantity: 1, price: 7000, total: 7000 }, { description: 'Video/Streaming', quantity: 1, price: 5000, total: 5000 }] },
  { id: '5', invoice_number: 'RF-2026-005', client: 'Festival GmbH', project_name: 'Open Air Festival Bodensee', status: 'overdue', amount: 25000, tax: 4750, total: 29750, issue_date: '2026-02-01', due_date: '2026-03-01', paid_date: null, items: [{ description: 'Anzahlung Festival-Paket (30%)', quantity: 1, price: 25000, total: 25000 }] },
]

// Equipment API endpoints
export const equipmentApi = {
  list: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockEquipment.slice((page - 1) * limit, page * limit), total: mockEquipment.length, page, limit })
      : api.get('/api/v1/equipment', { params: { page, limit } }).then(res => res.data),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockEquipment.find(e => e.id === id) || mockEquipment[0])
      : api.get(`/api/v1/equipment/${id}`).then(res => res.data),

  create: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()) }) : api.post('/api/v1/equipment', data).then(res => res.data),

  update: (id: string, data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id }) : api.put(`/api/v1/equipment/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    MOCK_MODE ? mockDelay({ success: true, id }) : api.delete(`/api/v1/equipment/${id}`).then(res => res.data),

  getPrice: (id: string, days: number, discount = 0) =>
    api.get(`/api/v1/equipment/${id}/price`, { params: { days, discount } }).then(res => res.data),

  getAvailability: (id: string, start: string, end: string) =>
    api.get(`/api/v1/equipment/${id}/availability`, { params: { start, end } }).then(res => res.data),

  getQRCode: (id: string) =>
    api.get(`/api/v1/equipment/${id}/qr-code`, { responseType: 'blob' }).then(res => res.data),

  getHistory: (id: string, page = 1, limit = 50) =>
    api.get(`/api/v1/equipment/${id}/history`, { params: { page, limit } }).then(res => res.data),

  importCSV: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/api/v1/equipment/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then(res => res.data)
  },

  // Equipment per Barcode suchen (für Scanner)
  getByBarcode: (barcode: string) =>
    MOCK_MODE
      ? mockDelay(mockEquipment.find(e => e.barcode === barcode))
      : api.get(`/api/v1/equipment/barcode/${barcode}`).then(res => res.data),

  // Equipment per Suchbegriff finden
  search: (query: string, limit = 20) =>
    MOCK_MODE
      ? mockDelay(mockEquipment.filter(e => e.name.toLowerCase().includes(query.toLowerCase())).slice(0, limit))
      : api.get('/api/v1/equipment/search', { params: { q: query, limit } }).then(res => res.data),

  // Equipment einem Projekt zuweisen (Check-Out)
  checkOut: (equipmentId: string, projectId: string) =>
    api.post(`/api/v1/equipment/${equipmentId}/check-out`, { project_id: projectId }).then(res => res.data),

  // Equipment vom Projekt zurückgeben (Check-In)
  checkIn: (equipmentId: string) =>
    api.post(`/api/v1/equipment/${equipmentId}/check-in`, {}).then(res => res.data),
}

// Project API endpoints
export const projectApi = {
  list: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockProjects.slice((page - 1) * limit, page * limit), total: mockProjects.length, page, limit })
      : api.get('/api/v1/projects', { params: { page, limit } }).then(res => res.data),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockProjects.find(p => p.id === id) || mockProjects[0])
      : api.get(`/api/v1/projects/${id}`).then(res => res.data),

  create: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()) }) : api.post('/api/v1/projects', data).then(res => res.data),

  update: (id: string, data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id }) : api.put(`/api/v1/projects/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    MOCK_MODE ? mockDelay({ success: true, id }) : api.delete(`/api/v1/projects/${id}`).then(res => res.data),

  getCalendar: (start: string, end: string) =>
    api.get('/api/v1/projects/calendar', { params: { start, end } }).then(res => res.data),

  getConflicts: (equipmentId: string, start: string, end: string) =>
    api.get(`/api/v1/projects/conflicts/${equipmentId}`, { params: { start, end } }).then(res => res.data),

  copyProject: (id: string) =>
    api.post(`/api/v1/projects/${id}/copy`, {}).then(res => res.data),

  getPackingListHTML: (id: string) =>
    api.get(`/api/v1/projects/${id}/packing-list`, { responseType: 'blob' }).then(res => res.data),
}

// Invoice API endpoints
export const invoiceApi = {
  list: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockInvoices.slice((page - 1) * limit, page * limit), total: mockInvoices.length, page, limit })
      : api.get('/api/v1/invoices', { params: { page, limit } }).then(res => res.data),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockInvoices.find(i => i.id === id) || mockInvoices[0])
      : api.get(`/api/v1/invoices/${id}`).then(res => res.data),

  create: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()) }) : api.post('/api/v1/invoices', data).then(res => res.data),

  update: (id: string, data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id }) : api.put(`/api/v1/invoices/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    MOCK_MODE ? mockDelay({ success: true, id }) : api.delete(`/api/v1/invoices/${id}`).then(res => res.data),

  getOpen: () =>
    MOCK_MODE ? mockDelay(mockInvoices.filter(i => i.status !== 'paid')) : api.get('/api/v1/invoices/open', {}).then(res => res.data),

  getPDF: (id: string) =>
    api.get(`/api/v1/invoices/${id}/pdf`, { responseType: 'blob' }).then(res => res.data),

  createFromProject: (projectId: string) =>
    api.post(`/api/v1/invoices/from-project/${projectId}`, {}).then(res => res.data),

  getQuotePDF: (quoteId: string) =>
    api.get(`/api/v1/invoices/${quoteId}/quote-pdf`, { responseType: 'blob' }).then(res => res.data),
}

// Customers API endpoints
export const customerApi = {
  list: (page = 1, limit = 50) =>
    api.get('/api/v1/customers', { params: { page, limit } }).then(res => res.data),

  getById: (id: string) =>
    api.get(`/api/v1/customers/${id}`).then(res => res.data),

  create: (data: unknown) =>
    api.post('/api/v1/customers', data).then(res => res.data),

  update: (id: string, data: unknown) =>
    api.put(`/api/v1/customers/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    api.delete(`/api/v1/customers/${id}`).then(res => res.data),
}

export default api
