import axios, { AxiosInstance } from 'axios'
import { useAuthStore } from '../stores/authStore'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

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
    const tenantId = useAuthStore.getState().tenantId
    if (tenantId) {
      config.headers['X-Tenant-ID'] = tenantId
    }
    // Extract user ID from JWT for service-to-service auth
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split('.')[1]))
        if (payload.sub) {
          config.headers['X-User-ID'] = payload.sub
        }
      } catch { /* ignore decode errors */ }
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor - unwrap { data: ..., message: "..." } envelope from backend
let isRefreshing = false
let failedQueue: Array<{ resolve: (token: string) => void; reject: (err: unknown) => void }> = []

const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (token) prom.resolve(token)
    else prom.reject(error)
  })
  failedQueue = []
}

api.interceptors.response.use(
  (response) => {
    // Backend wraps responses as { data: <payload>, message: "..." }
    // Unwrap so that res.data gives the actual payload directly
    if (response.data && typeof response.data === 'object' && 'data' in response.data && 'message' in response.data) {
      response.data = response.data.data
    }
    return response
  },
  async (error) => {
    const originalRequest = error.config
    if (error.response?.status === 401 && !originalRequest._retry) {
      // Don't retry refresh or login requests
      if (originalRequest.url?.includes('/auth/refresh') || originalRequest.url?.includes('/auth/login')) {
        useAuthStore.getState().logout()
        window.location.href = '/login'
        return Promise.reject(error)
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        }).then((token) => {
          originalRequest.headers.Authorization = `Bearer ${token}`
          return api(originalRequest)
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const res = await api.post('/api/v1/auth/refresh')
        const newToken = res.data?.access_token || res.data?.token || res.data?.data?.access_token
        if (newToken) {
          useAuthStore.getState().setToken(newToken)
          originalRequest.headers.Authorization = `Bearer ${newToken}`
          processQueue(null, newToken)
          return api(originalRequest)
        }
        throw new Error('No token in refresh response')
      } catch (refreshError) {
        processQueue(refreshError, null)
        useAuthStore.getState().logout()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }
    return Promise.reject(error)
  }
)

// ============================================================================
// MOCK MODE - enables frontend without backend
// ============================================================================
const MOCK_MODE = import.meta.env.VITE_MOCK === 'true'

const mockDelay = <T>(data: T, ms = 300): Promise<T> =>
  new Promise((resolve) => setTimeout(() => resolve(data), ms))

// Auth API endpoints
export const authApi = {
  login: async (email: string, password: string) => {
    if (MOCK_MODE) {
      if (email === 'admin@example.com' && password === 'password') {
        return mockDelay({
          token: 'mock-jwt-token-rentflow-2026',
          user: { id: '1', email: 'admin@example.com', name: 'Marco Berger' },
        })
      }
      return Promise.reject({ response: { status: 401 } })
    }
    const res = await api.post('/api/v1/auth/login', { email, password, tenant_slug: localStorage.getItem('cd_tenant') || 'je-soundulight' })
    const accessToken = res.data?.data?.tokens?.access_token || res.data?.data?.access_token || res.data?.token || res.data?.access_token
    return {
      token: accessToken,
      user: { email, name: email.split('@')[0] }
    }
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

  updateProfile: (data: { first_name?: string; last_name?: string; email?: string; phone?: string }) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: '1', name: `${data.first_name || 'Marco'} ${data.last_name || 'Berger'}`, email: data.email || 'admin@example.com' })
      : api.put('/api/v1/auth/profile', data).then(res => res.data),

  changePassword: (data: { current_password: string; new_password: string }) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.put('/api/v1/auth/password', data).then(res => res.data),

  forgotPassword: (email: string) =>
    MOCK_MODE
      ? mockDelay({ message: 'Reset link generated' })
      : api.post('/api/v1/auth/forgot-password', { email }).then(res => res.data),

  resetPassword: (token: string, new_password: string) =>
    MOCK_MODE
      ? mockDelay({ message: 'Password reset successfully' })
      : api.post('/api/v1/auth/reset-password', { token, new_password }).then(res => res.data),

  getSessions: () =>
    MOCK_MODE
      ? mockDelay([
          { id: '1', browser: 'Chrome 122', ip: '192.168.1.42', last_active: '2026-03-23T10:30:00Z', current: true },
          { id: '2', browser: 'Firefox 124', ip: '10.0.0.15', last_active: '2026-03-22T18:15:00Z', current: false },
          { id: '3', browser: 'Safari 18', ip: '172.16.0.8', last_active: '2026-03-21T09:00:00Z', current: false },
        ])
      : api.get('/api/v1/auth/sessions').then(res => res.data),

  revokeOtherSessions: () =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.post('/api/v1/auth/sessions/revoke-others').then(res => res.data),

  generateQRToken: () =>
    MOCK_MODE
      ? mockDelay({ qr_token: 'mock-qr-token-12345', expires_at: new Date(Date.now() + 5 * 60 * 1000).toISOString() })
      : api.post('/api/v1/auth/qr-token').then(res => {
          // qr-token endpoint returns { qr_token, expires_at } without data/message envelope
          const d = res.data
          return { qr_token: d.qr_token, expires_at: d.expires_at }
        }),

  getQRStatus: (token: string) =>
    MOCK_MODE
      ? mockDelay({ status: 'pending', redeemed: false })
      : api.get(`/api/v1/auth/qr-status/${token}`).then(res => res.data as { status: string; redeemed: boolean }),
}

// ============================================================================
// MOCK DATA
// ============================================================================
const mockEquipment = [
  { id: '1', tenant_id: '1', name: 'JBL VTX A12', sku: 'SPK-001', barcode: 'RF-SPK-001', category_id: 'cat-audio', status: 'available', condition: 'good', location_id: 'loc-1', description: 'Line Array Lautsprecher, 2x 12"', rental_price_day: 120, rental_price_week: 600, weight: 35, dimensions: { length: 60, width: 40, height: 35, unit: 'cm' }, image_refs: [], tags: ['audio', 'pa'], custom_fields: {}, created_at: '2025-09-15T10:00:00Z', updated_at: '2026-03-01T14:30:00Z' },
  { id: '2', tenant_id: '1', name: 'Shure SM58', sku: 'MIC-001', barcode: 'RF-MIC-001', category_id: 'cat-audio', status: 'available', condition: 'good', location_id: 'loc-1', description: 'Dynamisches Gesangsmikrofon', rental_price_day: 15, rental_price_week: 75, weight: 0.33, dimensions: {}, image_refs: [], tags: ['audio', 'mikrofon'], custom_fields: {}, created_at: '2025-09-15T10:00:00Z', updated_at: '2026-02-20T09:00:00Z' },
  { id: '3', tenant_id: '1', name: 'MA Lighting grandMA3', sku: 'LGT-001', barcode: 'RF-LGT-001', category_id: 'cat-lighting', status: 'checked_out', condition: 'good', location_id: 'loc-2', description: 'Lichtmischpult, Full-Size', rental_price_day: 250, rental_price_week: 1200, weight: 28, dimensions: {}, image_refs: [], tags: ['licht', 'pult'], custom_fields: {}, created_at: '2025-10-01T08:00:00Z', updated_at: '2026-03-18T16:00:00Z' },
  { id: '4', tenant_id: '1', name: 'Martin MAC Aura XB', sku: 'LGT-002', barcode: 'RF-LGT-002', category_id: 'cat-lighting', status: 'checked_out', condition: 'good', location_id: 'loc-2', description: 'LED Moving Head Wash', rental_price_day: 85, rental_price_week: 400, weight: 6.5, dimensions: {}, image_refs: [], tags: ['licht', 'moving-head'], custom_fields: {}, created_at: '2025-10-01T08:00:00Z', updated_at: '2026-03-18T16:00:00Z' },
  { id: '5', tenant_id: '1', name: 'Blackmagic ATEM Mini Extreme', sku: 'VID-001', barcode: 'RF-VID-001', category_id: 'cat-video', status: 'available', condition: 'good', location_id: 'loc-3', description: 'Video-Mischpult, 8 HDMI Inputs', rental_price_day: 180, rental_price_week: 850, weight: 1.2, dimensions: {}, image_refs: [], tags: ['video', 'mischer'], custom_fields: {}, created_at: '2025-11-10T12:00:00Z', updated_at: '2026-01-15T11:00:00Z' },
  { id: '6', tenant_id: '1', name: 'Prolyte X30V Truss 3m', sku: 'STG-001', barcode: 'RF-STG-001', category_id: 'cat-stage', status: 'available', condition: 'good', location_id: 'loc-4', description: 'Traversensystem, Aluminium', rental_price_day: 25, rental_price_week: 120, weight: 12, dimensions: { length: 300, width: 30, height: 30, unit: 'cm' }, image_refs: [], tags: ['bühne', 'truss'], custom_fields: {}, created_at: '2025-09-20T09:00:00Z', updated_at: '2026-03-10T08:00:00Z' },
  { id: '7', tenant_id: '1', name: 'Yamaha CL5', sku: 'AUD-002', barcode: 'RF-AUD-002', category_id: 'cat-audio', status: 'in_maintenance', condition: 'fair', location_id: 'loc-5', description: 'Digitales Mischpult, 72 Kanäle', rental_price_day: 350, rental_price_week: 1600, weight: 42, dimensions: {}, image_refs: [], tags: ['audio', 'pult'], custom_fields: {}, created_at: '2025-09-15T10:00:00Z', updated_at: '2026-03-20T10:00:00Z' },
  { id: '8', tenant_id: '1', name: 'Canon EOS R5', sku: 'VID-002', barcode: 'RF-VID-002', category_id: 'cat-video', status: 'reserved', condition: 'new', location_id: 'loc-3', description: 'Vollformat Kamera, 8K Video', rental_price_day: 200, rental_price_week: 950, weight: 0.74, dimensions: {}, image_refs: [], tags: ['video', 'kamera'], custom_fields: {}, created_at: '2025-12-01T14:00:00Z', updated_at: '2026-03-19T09:00:00Z' },
  { id: '9', tenant_id: '1', name: 'Chainmaster BGV-D8+ 1t', sku: 'STG-002', barcode: 'RF-STG-002', category_id: 'cat-stage', status: 'available', condition: 'good', location_id: 'loc-4', description: 'Kettenzug, 1 Tonne, BGV-D8+', rental_price_day: 45, rental_price_week: 200, weight: 25, dimensions: {}, image_refs: [], tags: ['bühne', 'rigging'], custom_fields: {}, created_at: '2025-10-15T11:00:00Z', updated_at: '2026-02-28T13:00:00Z' },
  { id: '10', tenant_id: '1', name: 'Robe MegaPointe', sku: 'LGT-003', barcode: 'RF-LGT-003', category_id: 'cat-lighting', status: 'available', condition: 'good', location_id: 'loc-1', description: 'Moving Head Spot/Beam/Wash', rental_price_day: 95, rental_price_week: 450, weight: 11, dimensions: {}, image_refs: [], tags: ['licht', 'moving-head'], custom_fields: {}, created_at: '2025-11-20T10:00:00Z', updated_at: '2026-03-05T15:00:00Z' },
]

const mockProjects = [
  { id: '1', name: 'Stadtfest München 2026', client: 'Stadt München', status: 'active', start_date: '2026-04-15', end_date: '2026-04-18', location: 'Marienplatz, München', budget: 45000, description: 'Jährliches Stadtfest mit 3 Bühnen, Full-Service Veranstaltungstechnik', created_at: '2026-01-10T09:00:00Z' },
  { id: '2', name: 'Firmen-Gala TechCorp', client: 'TechCorp GmbH', status: 'planning', start_date: '2026-05-20', end_date: '2026-05-20', location: 'Hilton Hotel, Frankfurt', budget: 18500, description: 'Firmengala für 500 Gäste, Ton/Licht/Video', created_at: '2026-02-15T14:00:00Z' },
  { id: '3', name: 'Open Air Festival Bodensee', client: 'Festival GmbH', status: 'planning', start_date: '2026-07-10', end_date: '2026-07-12', location: 'Seepark, Konstanz', budget: 85000, description: '3-Tages Festival, 2 Stages, Camping', created_at: '2026-02-20T11:00:00Z' },
  { id: '4', name: 'Produktlaunch AutoBrand', client: 'AutoBrand AG', status: 'completed', start_date: '2026-02-28', end_date: '2026-02-28', location: 'Messe Stuttgart', budget: 32000, description: 'Produktpräsentation neues E-Auto, LED-Wall + Audio', created_at: '2025-12-05T10:00:00Z' },
  { id: '5', name: 'Hochzeit Familie Weber', client: 'Familie Weber', status: 'completed', start_date: '2026-03-08', end_date: '2026-03-08', location: 'Schloss Neuschwanstein', budget: 8500, description: 'Hochzeitsfeier, DJ-Setup + Beleuchtung', created_at: '2026-01-20T16:00:00Z' },
]

const mockInvoices = [
  { id: '1', number: 'RF-2026-001', client_id: 'c1', client_name: 'AutoBrand AG', client_email: 'info@autobrand.de', client_address: 'Industriestr. 12, 80333 München', project_id: 'p1', project_name: 'Produktlaunch AutoBrand', status: 'paid', subtotal: 32000, tax_total: 6080, total: 38080, issue_date: '2026-03-01', due_date: '2026-03-31', paid_date: '2026-03-15', payment_terms: '30', notes: 'Vielen Dank für Ihren Auftrag!', bank_name: 'Sparkasse München', bank_iban: 'DE89 3704 0044 0532 0130 00', bank_bic: 'COBADEFFXXX', bank_account_holder: 'JE-Sound&Light', created_at: '2026-03-01T10:00:00Z', updated_at: '2026-03-15T14:00:00Z', line_items: [{ id: 'li1', invoice_id: '1', name: 'LED-Wall 6x3m', description: 'LED-Wall 6x3m (3 Tage)', quantity: 1, unit_price: 4500, total: 4500, tax_rate: 19, tax_amount: 855 }, { id: 'li2', invoice_id: '1', name: 'Line Array System', description: 'Line Array System (3 Tage)', quantity: 2, unit_price: 3600, total: 7200, tax_rate: 19, tax_amount: 1368 }, { id: 'li3', invoice_id: '1', name: 'Lichttechnik Paket', description: 'Lichttechnik Paket', quantity: 1, unit_price: 8500, total: 8500, tax_rate: 19, tax_amount: 1615 }, { id: 'li4', invoice_id: '1', name: 'Techniker', description: 'Techniker (3 Tage)', quantity: 4, unit_price: 2950, total: 11800, tax_rate: 19, tax_amount: 2242 }] },
  { id: '2', number: 'RF-2026-002', client_id: 'c2', client_name: 'Familie Weber', client_email: 'weber@email.de', client_address: 'Rosenstr. 5, 81669 München', project_id: 'p2', project_name: 'Hochzeit Familie Weber', status: 'paid', subtotal: 8500, tax_total: 1615, total: 10115, issue_date: '2026-03-09', due_date: '2026-04-09', paid_date: '2026-03-20', payment_terms: '30', notes: '', bank_name: 'Sparkasse München', bank_iban: 'DE89 3704 0044 0532 0130 00', bank_bic: 'COBADEFFXXX', bank_account_holder: 'JE-Sound&Light', created_at: '2026-03-09T09:00:00Z', updated_at: '2026-03-20T11:00:00Z', line_items: [{ id: 'li5', invoice_id: '2', name: 'DJ-Setup komplett', description: 'DJ-Setup komplett', quantity: 1, unit_price: 3500, total: 3500, tax_rate: 19, tax_amount: 665 }, { id: 'li6', invoice_id: '2', name: 'Ambientebeleuchtung', description: 'Ambientebeleuchtung', quantity: 1, unit_price: 2800, total: 2800, tax_rate: 19, tax_amount: 532 }, { id: 'li7', invoice_id: '2', name: 'Techniker', description: 'Techniker', quantity: 2, unit_price: 1100, total: 2200, tax_rate: 19, tax_amount: 418 }] },
  { id: '3', number: 'RF-2026-003', client_id: 'c3', client_name: 'Stadt München', client_email: 'veranstaltungen@muenchen.de', client_address: 'Marienplatz 8, 80331 München', project_id: 'p3', project_name: 'Stadtfest München 2026', status: 'draft', subtotal: 45000, tax_total: 8550, total: 53550, issue_date: '2026-03-22', due_date: '2026-04-22', paid_date: null, payment_terms: '30', notes: 'Anzahlung 50% bei Auftragsbestätigung', bank_name: 'Sparkasse München', bank_iban: 'DE89 3704 0044 0532 0130 00', bank_bic: 'COBADEFFXXX', bank_account_holder: 'JE-Sound&Light', created_at: '2026-03-22T08:00:00Z', updated_at: '2026-03-22T08:00:00Z', line_items: [{ id: 'li8', invoice_id: '3', name: 'PA-System Hauptbühne', description: 'PA-System Hauptbühne', quantity: 1, unit_price: 12000, total: 12000, tax_rate: 19, tax_amount: 2280 }, { id: 'li9', invoice_id: '3', name: 'Lichttechnik 3 Bühnen', description: 'Lichttechnik 3 Bühnen', quantity: 1, unit_price: 15000, total: 15000, tax_rate: 19, tax_amount: 2850 }, { id: 'li10', invoice_id: '3', name: 'Bühne + Truss', description: 'Bühne + Truss', quantity: 1, unit_price: 8000, total: 8000, tax_rate: 19, tax_amount: 1520 }, { id: 'li11', invoice_id: '3', name: 'Techniker-Team', description: 'Techniker-Team (4 Tage)', quantity: 8, unit_price: 1250, total: 10000, tax_rate: 19, tax_amount: 1900 }] },
  { id: '4', number: 'RF-2026-004', client_id: 'c4', client_name: 'TechCorp GmbH', client_email: 'events@techcorp.de', client_address: 'Leopoldstr. 44, 80802 München', project_id: 'p4', project_name: 'Firmen-Gala TechCorp', status: 'sent', subtotal: 18500, tax_total: 3515, total: 22015, issue_date: '2026-03-18', due_date: '2026-04-18', paid_date: null, payment_terms: '30', notes: '', bank_name: 'Sparkasse München', bank_iban: 'DE89 3704 0044 0532 0130 00', bank_bic: 'COBADEFFXXX', bank_account_holder: 'JE-Sound&Light', created_at: '2026-03-18T10:00:00Z', updated_at: '2026-03-18T10:00:00Z', line_items: [{ id: 'li12', invoice_id: '4', name: 'Audio-Paket Gala', description: 'Audio-Paket Gala', quantity: 1, unit_price: 6500, total: 6500, tax_rate: 19, tax_amount: 1235 }, { id: 'li13', invoice_id: '4', name: 'Licht-Design Gala', description: 'Licht-Design Gala', quantity: 1, unit_price: 7000, total: 7000, tax_rate: 19, tax_amount: 1330 }, { id: 'li14', invoice_id: '4', name: 'Video/Streaming', description: 'Video/Streaming', quantity: 1, unit_price: 5000, total: 5000, tax_rate: 19, tax_amount: 950 }] },
  { id: '5', number: 'RF-2026-005', client_id: 'c5', client_name: 'Festival GmbH', client_email: 'buchhaltung@festival-gmbh.de', client_address: 'Seestr. 22, 88131 Lindau', project_id: 'p5', project_name: 'Open Air Festival Bodensee', status: 'overdue', subtotal: 25000, tax_total: 4750, total: 29750, issue_date: '2026-02-01', due_date: '2026-03-01', paid_date: null, payment_terms: '30', notes: 'Zahlungserinnerung wurde gesendet', bank_name: 'Sparkasse München', bank_iban: 'DE89 3704 0044 0532 0130 00', bank_bic: 'COBADEFFXXX', bank_account_holder: 'JE-Sound&Light', dunning_entries: [{ level: 1, date: '2026-03-08', sent: true }], created_at: '2026-02-01T09:00:00Z', updated_at: '2026-03-08T10:00:00Z', line_items: [{ id: 'li15', invoice_id: '5', name: 'Anzahlung Festival-Paket', description: 'Anzahlung Festival-Paket (30%)', quantity: 1, unit_price: 25000, total: 25000, tax_rate: 19, tax_amount: 4750 }] },
  { id: '6', number: 'RF-2026-006', client_id: 'c6', client_name: 'Messe Frankfurt GmbH', client_email: 'technik@messe-ffm.de', client_address: 'Ludwig-Erhard-Anlage 1, 60327 Frankfurt', project_id: 'p6', project_name: 'Messe-Auftritt Q2', status: 'partial', subtotal: 15000, tax_total: 2850, total: 17850, issue_date: '2026-03-05', due_date: '2026-04-05', paid_date: null, payment_terms: '30', notes: 'Erste Teilzahlung erhalten', bank_name: 'Sparkasse München', bank_iban: 'DE89 3704 0044 0532 0130 00', bank_bic: 'COBADEFFXXX', bank_account_holder: 'JE-Sound&Light', payments: [{ id: 'pay1', invoice_id: '6', amount: 10000, date: '2026-03-15', method: 'Überweisung', notes: 'Erste Rate' }], created_at: '2026-03-05T11:00:00Z', updated_at: '2026-03-15T14:00:00Z', line_items: [{ id: 'li16', invoice_id: '6', name: 'Messe-Technikpaket', description: 'Komplettes Audio/Video-Paket für Messestand', quantity: 1, unit_price: 15000, total: 15000, tax_rate: 19, tax_amount: 2850 }] },
]

// Transport/Fleet Mock Data
const mockVehicles = [
  { id: '1', tenant_id: '1', license_plate: 'M-RFA-001', name: 'MAN TGX Ladeflächenlänge 7,5m', capacity_kg: 18000, capacity_m3: 52, vehicle_type: 'Lastkraftwagen', status: 'available', dguv_last_check: '2025-12-15T10:00:00Z', dguv_next_check: '2026-12-15T10:00:00Z', created_at: '2025-03-01T09:00:00Z', updated_at: '2026-03-10T14:30:00Z' },
  { id: '2', tenant_id: '1', license_plate: 'M-RFA-002', name: 'Scania R500 V8 mit Aufbau', capacity_kg: 22000, capacity_m3: 65, vehicle_type: 'Lastkraftwagen', status: 'in_use', dguv_last_check: '2026-02-01T08:00:00Z', dguv_next_check: '2027-02-01T08:00:00Z', created_at: '2025-03-15T10:00:00Z', updated_at: '2026-03-18T11:00:00Z' },
  { id: '3', tenant_id: '1', license_plate: 'M-RFA-003', name: 'Sprinter VAN 5 Tonnen', capacity_kg: 4500, capacity_m3: 14, vehicle_type: 'Transporter', status: 'available', dguv_last_check: '2025-11-20T09:00:00Z', dguv_next_check: '2026-11-20T09:00:00Z', created_at: '2025-04-01T08:00:00Z', updated_at: '2026-03-01T16:00:00Z' },
  { id: '4', tenant_id: '1', license_plate: 'M-RFA-004', name: 'Iveco Daily 3,5t Behördenfahrzeug', capacity_kg: 3500, capacity_m3: 11, vehicle_type: 'Transporter', status: 'maintenance', dguv_last_check: '2025-09-10T14:00:00Z', dguv_next_check: '2026-09-10T14:00:00Z', created_at: '2025-05-01T07:00:00Z', updated_at: '2026-03-20T08:30:00Z' },
]

const mockTours = [
  { id: '1', tenant_id: '1', vehicle_id: '1', project_id: '1', driver_id: 'user-2', status: 'completed', departure_at: '2026-03-20T06:00:00Z', arrival_at: '2026-03-20T14:30:00Z', km_start: 125400, km_end: 125680, total_cost: 450.50, notes: 'Stadtfest München - Anlieferung Bühnenteile', created_at: '2026-03-19T15:00:00Z', updated_at: '2026-03-20T16:00:00Z' },
  { id: '2', tenant_id: '1', vehicle_id: '2', project_id: '1', driver_id: 'user-3', status: 'in_transit', departure_at: '2026-03-22T07:00:00Z', arrival_at: '2026-03-22T18:00:00Z', km_start: 98750, km_end: 99100, total_cost: 525.75, notes: 'Stadtfest München - Lichttechnik Transport', created_at: '2026-03-22T05:30:00Z', updated_at: '2026-03-22T12:15:00Z' },
  { id: '3', tenant_id: '1', vehicle_id: '3', project_id: '2', driver_id: 'user-2', status: 'planned', departure_at: '2026-05-18T08:00:00Z', arrival_at: '2026-05-18T14:00:00Z', km_start: 45600, km_end: 0, total_cost: 0, notes: 'Firmen-Gala TechCorp - Anlieferung', created_at: '2026-05-10T10:00:00Z', updated_at: '2026-05-10T10:00:00Z' },
]

const mockMaintenancePlans = [
  { id: '1', tenant_id: '1', equipment_id: '1', equipment_name: 'JBL VTX A12', plan_type: 'interval', interval_days: 90, interval_hours: null, name: 'Reinigung und Kontrolle', description: 'Regelmäßige Reinigung und visuelle Kontrolle der Lautsprecher', checklist_template_id: 'chk-1', last_executed_at: '2026-02-15T10:00:00Z', next_due_at: '2026-05-15T10:00:00Z', is_active: true, created_at: '2025-06-01T09:00:00Z', updated_at: '2026-02-15T10:00:00Z' },
  { id: '2', tenant_id: '1', equipment_id: '3', equipment_name: 'MA Lighting grandMA3', plan_type: 'interval', interval_days: 180, interval_hours: null, name: 'Software-Update und Kalibrierung', description: 'Betriebssystem-Updates und Farbkalibrierung durchführen', checklist_template_id: 'chk-2', last_executed_at: '2025-12-10T11:00:00Z', next_due_at: '2026-06-10T11:00:00Z', is_active: true, created_at: '2025-06-01T09:00:00Z', updated_at: '2025-12-10T11:00:00Z' },
  { id: '3', tenant_id: '1', equipment_id: '7', equipment_name: 'Yamaha CL5', plan_type: 'hours_based', interval_days: null, interval_hours: 100, name: 'Wartung nach Nutzung (alle 100h)', description: 'Inspektionen nach 100 Betriebsstunden durchführen', checklist_template_id: 'chk-3', last_executed_at: '2026-01-20T14:00:00Z', next_due_at: '2026-03-25T14:00:00Z', is_active: true, created_at: '2025-07-01T09:00:00Z', updated_at: '2026-01-20T14:00:00Z' },
]

const mockMaintenanceTasks = [
  { id: '1', tenant_id: '1', plan_id: '1', plan_name: 'Reinigung und Kontrolle', equipment_id: '1', equipment_name: 'JBL VTX A12', assigned_to: 'user-4', status: 'completed', priority: 'medium', scheduled_at: '2026-02-15T10:00:00Z', started_at: '2026-02-15T10:15:00Z', completed_at: '2026-02-15T11:45:00Z', notes: 'Alle Komponenten gereinigt, keine Schäden gefunden', checklist_data: { items: [{ name: 'Visuelle Kontrolle', check_type: 'yes_no', value: true, passed: true }, { name: 'Reinigung durchgeführt', check_type: 'yes_no', value: true, passed: true }] }, created_at: '2026-02-10T09:00:00Z', updated_at: '2026-02-15T11:45:00Z' },
  { id: '2', tenant_id: '1', plan_id: '3', plan_name: 'Wartung nach Nutzung (alle 100h)', equipment_id: '7', equipment_name: 'Yamaha CL5', assigned_to: null, status: 'planned', priority: 'high', scheduled_at: '2026-03-28T09:00:00Z', started_at: null, completed_at: null, notes: 'Wartung fällig nach Einsatz auf Stadtfest', checklist_data: null, created_at: '2026-03-10T14:00:00Z', updated_at: '2026-03-10T14:00:00Z' },
  { id: '3', tenant_id: '1', plan_id: '2', plan_name: 'Software-Update und Kalibrierung', equipment_id: '3', equipment_name: 'MA Lighting grandMA3', assigned_to: 'user-4', status: 'overdue', priority: 'critical', scheduled_at: '2026-03-15T10:00:00Z', started_at: null, completed_at: null, notes: 'Wichtig: v2.1.5 Update ausstehend', checklist_data: null, created_at: '2026-03-01T10:00:00Z', updated_at: '2026-03-22T08:00:00Z' },
]

const mockElectricalTests = [
  { id: '1', tenant_id: '1', task_id: '1', equipment_id: '1', equipment_name: 'JBL VTX A12', tester_id: 'user-5', test_type: 'vde_0702', test_date: '2026-02-15T11:00:00Z', next_test_date: '2027-02-15T11:00:00Z', result: 'passed', insulation_resistance_mohm: 2.5, protective_conductor_resistance_ohm: 0.15, leakage_current_ma: 1.2, visual_inspection_ok: true, functional_test_ok: true, certificate_number: 'VDE-RF-2026-001', notes: 'Alle Tests bestanden, Gerät freigegeben', created_at: '2026-02-15T11:00:00Z' },
  { id: '2', tenant_id: '1', task_id: null, equipment_id: '3', equipment_name: 'MA Lighting grandMA3', tester_id: 'user-5', test_type: 'vde_0701', test_date: '2025-12-10T13:00:00Z', next_test_date: '2026-12-10T13:00:00Z', result: 'conditional', insulation_resistance_mohm: 1.8, protective_conductor_resistance_ohm: 0.18, leakage_current_ma: 2.1, visual_inspection_ok: true, functional_test_ok: false, certificate_number: 'VDE-RF-2025-045', notes: 'FI-Schalter muss überprüft werden', created_at: '2025-12-10T13:00:00Z' },
]

const mockChecklists = [
  { id: 'chk-1', tenant_id: '1', name: 'Lautsprecher-Wartung', description: 'Checkliste für Lautsprecher-Inspektionen', version: 2, items: [{ name: 'Visuelle Kontrolle', description: 'Beschädigungen, Verschleiß prüfen', check_type: 'yes_no', required: true }, { name: 'Reinigung durchgeführt', description: 'Oberfläche und Anschlüsse reinigen', check_type: 'yes_no', required: true }, { name: 'Höhe Kopplungskondensator (V)', description: 'Spannungsmessung durchführen', check_type: 'measurement', required: false, unit: 'V', min_value: 0, max_value: 5 }], created_at: '2025-06-01T09:00:00Z' },
  { id: 'chk-2', tenant_id: '1', name: 'Lichttechnik-Inspektion', description: 'Checkliste für Lichttechnik-Wartung', version: 1, items: [{ name: 'Lampen-Betriebsstunden prüfen', description: 'Betriebsstunden auslesen und dokumentieren', check_type: 'measurement', required: true, unit: 'h' }, { name: 'Kühlgebläse prüfen', description: 'Geräusch und Luftstrom prüfen', check_type: 'yes_no', required: true }, { name: 'Software-Version', description: 'Aktuelle Softwareversion überprüfen', check_type: 'text', required: true }], created_at: '2025-06-01T09:00:00Z' },
]

const mockWarehouses = [
  { id: '1', tenant_id: '1', name: 'Lager München - Hauptstandort', location: 'Moosacher Str. 88, München', total_capacity_kg: 50000, total_capacity_m3: 500, zones: [{ id: 'z1', name: 'Zone A - Audio', type: 'zone', capacity_kg: 15000, capacity_m3: 150 }, { id: 'z2', name: 'Zone B - Lichttechnik', type: 'zone', capacity_kg: 20000, capacity_m3: 200 }, { id: 'z3', name: 'Zone C - Bühne', type: 'zone', capacity_kg: 15000, capacity_m3: 150 }], created_at: '2025-01-01T08:00:00Z', updated_at: '2026-03-10T09:00:00Z' },
]

const mockStockMovements = [
  { id: '1', equipment_id: '1', from_location: 'z1', to_location: 'Tour #1', movement_type: 'checkout', quantity: 8, timestamp: '2026-03-20T06:00:00Z', user_id: 'user-2' },
  { id: '2', equipment_id: '3', from_location: 'z2', to_location: 'Projekt: Stadtfest München', movement_type: 'checkout', quantity: 2, timestamp: '2026-03-18T16:00:00Z', user_id: 'user-1' },
]

// Equipment API endpoints
export const equipmentApi = {
  list: (params: { limit?: number; offset?: number; status?: string; category_id?: string; location_id?: string; type_id?: string } = {}) =>
    MOCK_MODE
      ? mockDelay({ data: mockEquipment.slice(params.offset || 0, (params.offset || 0) + (params.limit || 20)), total: mockEquipment.length, limit: params.limit || 20, offset: params.offset || 0 })
      : api.get('/api/v1/equipment', { params }).then(res => res.data),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockEquipment.find(e => e.id === id) || mockEquipment[0])
      : api.get(`/api/v1/equipment/${id}`).then(res => res.data),

  create: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()) }) : api.post('/api/v1/equipment', data).then(res => res.data),

  update: (id: string, data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id }) : api.put(`/api/v1/equipment/${id}`, data).then(res => res.data),

  changeStatus: (id: string, status: string, reason?: string) =>
    api.patch(`/api/v1/equipment/${id}/status`, { status, reason }).then(res => res.data),

  delete: (id: string) =>
    MOCK_MODE ? mockDelay({ success: true, id }) : api.delete(`/api/v1/equipment/${id}`).then(res => res.data),

  getPrice: (id: string, days: number, discount = 0) =>
    api.get(`/api/v1/equipment/${id}/price`, { params: { days, discount } }).then(res => res.data),

  getAvailability: (id: string, qty = 1) =>
    api.get(`/api/v1/equipment/${id}/availability`, { params: { qty } }).then(res => res.data),

  getQRCode: (id: string) =>
    api.get(`/api/v1/equipment/${id}/qr-code`, { responseType: 'blob' }).then(res => res.data),

  getHistory: (id: string, limit = 20, offset = 0) =>
    api.get(`/api/v1/equipment/${id}/history`, { params: { limit, offset } }).then(res => res.data),

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
      : api.get(`/api/v1/equipment/lookup/barcode/${barcode}`).then(res => res.data),

  // Equipment per RFID-Tag suchen (für UHF PDA Scanner)
  getByRfidTag: (tag: string) =>
    api.get(`/api/v1/equipment/lookup/rfid/${tag}`).then(res => res.data),

  // RFID-Tag einem Equipment zuordnen
  assignRfidTag: (equipmentId: string, rfidTag: string) =>
    api.patch(`/api/v1/equipment/${equipmentId}/rfid`, { rfid_tag: rfidTag }).then(res => res.data),

  // Equipment per Suchbegriff finden
  search: (query: string, limit = 20, offset = 0) =>
    MOCK_MODE
      ? mockDelay(mockEquipment.filter(e => e.name.toLowerCase().includes(query.toLowerCase())).slice(0, limit))
      : api.get('/api/v1/equipment/search', { params: { q: query, limit, offset } }).then(res => res.data),

  // Equipment einem Projekt zuweisen (Check-Out)
  checkOut: (equipmentId: string, projectId: string) =>
    MOCK_MODE
      ? mockDelay({ id: equipmentId, project_id: projectId, status: 'checked_out' })
      : api.post(`/api/v1/equipment/${equipmentId}/check-out`, { project_id: projectId }).then(res => res.data),

  // Equipment vom Projekt zurückgeben (Check-In)
  checkIn: (equipmentId: string) =>
    MOCK_MODE
      ? mockDelay({ id: equipmentId, status: 'available' })
      : api.post(`/api/v1/equipment/${equipmentId}/check-in`).then(res => res.data),

  // Zustandsbericht für Equipment aktualisieren
  updateCondition: (equipmentId: string, condition: string, notes?: string, reportedBy?: string) =>
    MOCK_MODE
      ? mockDelay({ id: equipmentId, condition, status: condition === 'damaged' ? 'in_maintenance' : 'available' })
      : api.patch(`/api/v1/equipment/${equipmentId}/condition`, { condition, notes, reported_by: reportedBy }).then(res => res.data),

  uploadImage: (equipmentId: string, file: File) => {
    const formData = new FormData()
    formData.append('image', file)
    return api.post(`/api/v1/equipment/${equipmentId}/images`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then(res => res.data)
  },

  listReservations: (equipmentId: string) =>
    api.get('/api/v1/reservations', { params: { equipment_id: equipmentId } }).then(res => {
      const data = res.data
      return Array.isArray(data) ? data : (data?.data ?? data?.items ?? [])
    }),
}

// Equipment Type API endpoints (Typ/Instanz-Hierarchie)
export const equipmentTypeApi = {
  list: (params: { limit?: number; offset?: number; category_id?: string } = {}) =>
    api.get('/api/v1/equipment-types', { params }).then(res => res.data),

  getById: (id: string) =>
    api.get(`/api/v1/equipment-types/${id}`).then(res => res.data),

  create: (data: object) =>
    api.post('/api/v1/equipment-types', data).then(res => res.data),

  update: (id: string, data: object) =>
    api.put(`/api/v1/equipment-types/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    api.delete(`/api/v1/equipment-types/${id}`).then(res => res.data),

  createItems: (id: string, count: number) =>
    api.post(`/api/v1/equipment-types/${id}/create-items`, { count }).then(res => res.data),
}

// Tenant API endpoints
export const tenantApi = {
  getById: (_id?: string) =>
    api.get('/api/v1/tenants/current').then(res => res.data),

  update: (_id: string, data: object) =>
    api.put('/api/v1/tenants/current', data).then(res => res.data),
}

// User API endpoints (auth-service)
export const userApi = {
  list: () =>
    api.get('/api/v1/users').then(res => res.data),

  create: (data: { name: string; email: string; password: string; role: string }) =>
    api.post('/api/v1/users', data).then(res => res.data),

  deactivate: (id: string) =>
    api.put(`/api/v1/users/${id}/deactivate`).then(res => res.data),

  activate: (id: string) =>
    api.put(`/api/v1/users/${id}/activate`).then(res => res.data),

  delete: (id: string) =>
    api.delete(`/api/v1/users/${id}`).then(res => res.data),
}

// Category API endpoints
export const categoryApi = {
  list: () =>
    api.get('/api/v1/categories').then(res => res.data),

  getById: (id: string) =>
    api.get(`/api/v1/categories/${id}`).then(res => res.data),

  create: (data: object) =>
    api.post('/api/v1/categories', data).then(res => res.data),

  update: (id: string, data: object) =>
    api.put(`/api/v1/categories/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    api.delete(`/api/v1/categories/${id}`).then(res => res.data),
}

// Project API endpoints
export const projectApi = {
  list: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockProjects.slice((page - 1) * limit, page * limit), total: mockProjects.length, page, limit })
      : api.get('/api/v1/projects', { params: { offset: (page - 1) * limit, limit } }).then(res => res.data),

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

  copyProject: (id: string) =>
    api.post(`/api/v1/projects/${id}/copy`, {}).then(res => res.data),

  getPackingListHTML: (id: string) =>
    api.get(`/api/v1/projects/${id}/packing-list/html`, { responseType: 'blob' }).then(res => res.data),

  getPackingListJSON: (id: string) =>
    api.get(`/api/v1/projects/${id}/packing-list`).then(res => res.data),
}

// Reservation API endpoints
export const reservationApi = {
  list: (projectId: string) =>
    api.get('/api/v1/reservations', { params: { project_id: projectId } }).then(res => res.data),

  create: (data: { project_id: string; equipment_id: string; start_date: string; end_date: string; force?: boolean }) =>
    api.post('/api/v1/reservations', data).then(res => res.data),

  delete: (id: string) =>
    api.delete(`/api/v1/reservations/${id}`).then(res => res.data),

  checkConflicts: (equipmentId: string, start: string, end: string, excludeProjectId?: string) =>
    api.get('/api/v1/reservations/conflicts', {
      params: {
        equipment_id: equipmentId,
        start,
        end,
        ...(excludeProjectId ? { exclude_project_id: excludeProjectId } : {}),
      },
    }).then(res => {
      const data = res.data
      return Array.isArray(data) ? data : (data?.data ?? data?.items ?? [])
    }),
}

// Packlist API endpoints
export const packlistApi = {
  list: (projectId: string) =>
    api.get('/api/v1/packlists', { params: { project_id: projectId } }).then(res => res.data),

  getById: (id: string) =>
    api.get(`/api/v1/packlists/${id}`).then(res => res.data),

  create: (data: { project_id: string; name: string }) =>
    api.post('/api/v1/packlists', data).then(res => res.data),

  addItem: (packlistId: string, data: { equipment_id: string; equipment_name: string; quantity: number }) =>
    api.post(`/api/v1/packlists/${packlistId}/items`, data).then(res => res.data),

  removeItem: (packlistId: string, equipmentId: string) =>
    api.delete(`/api/v1/packlists/${packlistId}/items`, { data: { equipment_id: equipmentId } }).then(res => res.data),

  markItemPacked: (packlistId: string, data: { equipment_id: string; quantity_packed: number }) =>
    api.patch(`/api/v1/packlists/${packlistId}/items/pack`, data).then(res => res.data),

  markItemReturned: (packlistId: string, data: { equipment_id: string; quantity_returned: number }) =>
    api.patch(`/api/v1/packlists/${packlistId}/items/return`, data).then(res => res.data),

  changeStatus: (packlistId: string, status: string) =>
    api.patch(`/api/v1/packlists/${packlistId}/status`, { status }).then(res => res.data),

  delete: (id: string) =>
    api.delete(`/api/v1/packlists/${id}`).then(res => res.data),
}

// Invoice API endpoints
export const invoiceApi = {
  list: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ data: mockInvoices.slice((page - 1) * limit, page * limit), total: mockInvoices.length, page, limit })
      : api.get('/api/v1/invoices', { params: { page, limit } }).then(res => {
          const result = res.data || { data: [], total: 0 }
          // Ensure result.data is always an array
          if (!Array.isArray(result.data)) result.data = []
          // Map backend DTO field names to frontend Invoice type + cents→EUR
          const c2e = (v: number) => v > 1000000 ? v / 100 : v // heuristic: values > 10000 EUR are likely cents
          result.data = result.data.map((inv: any) => {
            const rawSub = inv.subtotal ?? inv.sub_total ?? 0
            const rawTax = inv.tax_total ?? inv.tax_amount ?? 0
            const rawTotal = inv.total ?? 0
            const mapped = {
              ...inv,
              number: inv.number || inv.invoice_number || '',
              subtotal: c2e(rawSub),
              tax_total: c2e(rawTax),
              total: c2e(rawTotal),
              line_items: (inv.line_items || inv.items || []).map((li: any) => ({
                ...li,
                unit_price: c2e(li.unit_price || 0),
                total: c2e(li.total || 0),
                tax_amount: c2e(li.tax_amount || 0),
              })),
            }
            // Recalculate totals from line_items if total is 0 but items exist
            if (mapped.total === 0 && mapped.line_items.length > 0) {
              mapped.subtotal = mapped.line_items.reduce((s: number, i: any) => s + ((i.quantity || 0) * (i.unit_price || 0)), 0)
              mapped.tax_total = mapped.line_items.reduce((s: number, i: any) => s + ((i.tax_amount || 0) || ((i.quantity || 0) * (i.unit_price || 0) * ((i.tax_rate || 0) / 100))), 0)
              mapped.total = mapped.subtotal + mapped.tax_total
            }
            return mapped
          })
          return result
        }),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockInvoices.find(i => i.id === id) || mockInvoices[0])
      : api.get(`/api/v1/invoices/${id}`).then(res => {
          const inv = res.data
          if (inv && typeof inv === 'object') {
            const c2e = (v: number) => v > 1000000 ? v / 100 : v
            inv.number = inv.number || inv.invoice_number || ''
            inv.subtotal = c2e(inv.subtotal ?? inv.sub_total ?? 0)
            inv.tax_total = c2e(inv.tax_total ?? inv.tax_amount ?? 0)
            inv.total = c2e(inv.total ?? 0)
            inv.line_items = (inv.line_items || inv.items || []).map((li: any) => ({
              ...li,
              unit_price: c2e(li.unit_price || 0),
              total: c2e(li.total || 0),
              tax_amount: c2e(li.tax_amount || 0),
            }))
            // Recalculate totals from line_items if total is 0 but items exist
            if (inv.total === 0 && inv.line_items.length > 0) {
              inv.subtotal = inv.line_items.reduce((s: number, i: any) => s + ((i.quantity || 0) * (i.unit_price || 0)), 0)
              inv.tax_total = inv.line_items.reduce((s: number, i: any) => s + ((i.tax_amount || 0) || ((i.quantity || 0) * (i.unit_price || 0) * ((i.tax_rate || 0) / 100))), 0)
              inv.total = inv.subtotal + inv.tax_total
            }
          }
          return inv
        }),

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
    api.post(`/api/v1/invoices/actions/from-project/${projectId}`, {}).then(res => res.data),

  getQuotePDF: (quoteId: string) =>
    api.get(`/api/v1/invoices/${quoteId}/quote-pdf`, { responseType: 'blob' }).then(res => res.data),

  sendEmail: (id: string) =>
    MOCK_MODE ? mockDelay({ success: true }) : api.post(`/api/v1/invoices/${id}/send`).then(res => res.data),

  createDunning: (id: string, level: number, fee?: number, message?: string) =>
    MOCK_MODE
      ? mockDelay({
          id: `dun_${Date.now()}_${level}`,
          invoice_id: id,
          level,
          level_name: level === 1 ? 'Zahlungserinnerung' : level === 2 ? '1. Mahnung' : '2. Mahnung',
          fee: fee ?? (level === 1 ? 0 : level === 2 ? 5 : 10),
          sent_at: new Date().toISOString(),
          due_date: new Date(Date.now() + (level === 2 ? 14 : 7) * 86400000).toISOString(),
          notes: message || '',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          date: new Date().toISOString(),
          sent: true,
        })
      : api.post('/api/v1/dunning/send-reminder', { invoice_id: id, level: level === 1 ? 'reminder' : level === 2 ? 'dunning_1' : 'dunning_2', notes: message || '' }).then(res => res.data),

  getDunningHistory: (id: string) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get(`/api/v1/dunning/entries/${id}`).then(res => res.data),

  getOverdueInvoices: () =>
    api.get('/api/v1/dunning/overdue').then(res => res.data),

  runDunningCheck: () =>
    api.post('/api/v1/dunning/check').then(res => res.data),
}

// ============================================================================
// MAIL API endpoints
// ============================================================================
export const mailApi = {
  // Mailboxes
  listMailboxes: () =>
    MOCK_MODE
      ? mockDelay([
          { id: 'mb-1', type: 'general', email: 'info@berger-eventtechnik.de', label: 'Allgemein', imap_host: 'imap.strato.de', imap_port: 993, smtp_host: 'smtp.strato.de', smtp_port: 587, username: 'info@berger-eventtechnik.de', tls: true },
          { id: 'mb-2', type: 'invoices', email: 'rechnung@berger-eventtechnik.de', label: 'Rechnungen', imap_host: 'imap.strato.de', imap_port: 993, smtp_host: 'smtp.strato.de', smtp_port: 587, username: 'rechnung@berger-eventtechnik.de', tls: true },
          { id: 'mb-3', type: 'personal', email: 'marco@berger-eventtechnik.de', label: 'Persönlich', imap_host: 'imap.strato.de', imap_port: 993, smtp_host: 'smtp.strato.de', smtp_port: 587, username: 'marco@berger-eventtechnik.de', tls: true },
        ])
      : api.get('/api/v1/mail/mailboxes').then(res => res.data),

  createMailbox: (data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: `mb-${Date.now()}` })
      : api.post('/api/v1/mail/mailboxes', data).then(res => res.data),

  updateMailbox: (id: string, data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id })
      : api.put(`/api/v1/mail/mailboxes/${id}`, data).then(res => res.data),

  deleteMailbox: (id: string) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.delete(`/api/v1/mail/mailboxes/${id}`).then(res => res.data),

  testMailbox: (id: string) =>
    MOCK_MODE
      ? mockDelay({ success: true, message: 'Verbindung erfolgreich' })
      : api.post(`/api/v1/mail/mailboxes/${id}/test`).then(res => res.data),

  // Mails
  list: (params?: any) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/mail', { params }).then(res => res.data),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.get(`/api/v1/mail/messages/${id}`).then(res => res.data),

  markAsRead: (id: string) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.put(`/api/v1/mail/messages/${id}/read`).then(res => res.data),

  assignToProject: (id: string, projectId: string) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.put(`/api/v1/mail/messages/${id}/project`, { project_id: projectId }).then(res => res.data),

  send: (data: any) =>
    MOCK_MODE
      ? mockDelay({ success: true, id: `sent-${Date.now()}` })
      : api.post('/api/v1/mail/send', data).then(res => res.data),
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

// ============================================================================
// CONTACTS/CRM API endpoints
// ============================================================================
export const contactApi = {
  list: (params?: any) =>
    MOCK_MODE
      ? mockDelay({ data: [], total: 0 })
      : api.get('/api/v1/customers', { params }).then(res => res.data),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.get(`/api/v1/customers/${id}`).then(res => res.data),

  create: (data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()), created_at: new Date().toISOString(), updated_at: new Date().toISOString() })
      : api.post('/api/v1/customers', data).then(res => res.data),

  update: (id: string, data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id, updated_at: new Date().toISOString() })
      : api.put(`/api/v1/customers/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.delete(`/api/v1/customers/${id}`).then(res => res.data),
}

// ============================================================================
// QUOTES API endpoints
// ============================================================================
export const quoteApi = {
  list: (params?: any) =>
    MOCK_MODE
      ? mockDelay({ data: [], total: 0 })
      : api.get('/api/v1/quotes', { params }).then(res => res.data),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.get(`/api/v1/quotes/${id}`).then(res => res.data),

  create: (data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/quotes', data).then(res => res.data),

  updateStatus: (id: string, status: string) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.patch(`/api/v1/quotes/${id}/status`, { status }).then(res => res.data),

  send: (id: string) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.patch(`/api/v1/quotes/${id}/status`, { status: 'sent' }).then(res => res.data),

  accept: (id: string) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.patch(`/api/v1/quotes/${id}/status`, { status: 'accepted' }).then(res => res.data),

  reject: (id: string) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.patch(`/api/v1/quotes/${id}/status`, { status: 'declined' }).then(res => res.data),

  convertToInvoice: (id: string) =>
    MOCK_MODE
      ? mockDelay({ id: String(Date.now()) })
      : api.post(`/api/v1/quotes/${id}/convert-to-invoice`).then(res => res.data),

  getPdf: (id: string) =>
    api.get(`/api/v1/quotes/${id}/pdf`, { responseType: 'blob' }).then(res => res.data),
}

// ============================================================================
// BANKING API endpoints
// ============================================================================
export const bankingApi = {
  importCSV: (file: File) =>
    file.text().then(csvText =>
      api.post('/api/v1/banking/import', csvText, {
        headers: { 'Content-Type': 'text/csv' },
      }).then(res => res.data)
    ),

  autoMatch: () =>
    api.post('/api/v1/banking/auto-match').then(res => res.data),

  confirmMatch: (transactionId: string, invoiceId: string) =>
    api.post('/api/v1/banking/confirm-match', { transaction_id: transactionId, invoice_id: invoiceId }).then(res => res.data),

  listTransactions: (matched?: boolean) =>
    api.get('/api/v1/banking/transactions', { params: matched !== undefined ? { matched } : {} }).then(res => res.data),
}

// ============================================================================
// DATEV EXPORT API endpoints
// ============================================================================
export const datevApi = {
  exportCSV: (from: string, to: string, format = 'skr03') =>
    api.get('/api/v1/export/datev', {
      params: { from, to, format },
      responseType: 'blob',
    }).then(res => res.data),
}

// ============================================================================
// TRANSPORT/FLEET API endpoints
// ============================================================================
export const transportApi = {
  // Vehicle Management
  listVehicles: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockVehicles.slice((page - 1) * limit, page * limit), total: mockVehicles.length, page, limit })
      : api.get('/api/v1/transport/vehicles', { params: { page, limit } }).then(res => res.data),

  getVehicle: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockVehicles.find(v => v.id === id) || mockVehicles[0])
      : api.get(`/api/v1/transport/vehicles/${id}`).then(res => res.data),

  createVehicle: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()), status: 'available' }) : api.post('/api/v1/transport/vehicles', data).then(res => res.data),

  updateVehicle: (id: string, data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id }) : api.put(`/api/v1/transport/vehicles/${id}`, data).then(res => res.data),

  deleteVehicle: (id: string) =>
    MOCK_MODE ? mockDelay({ success: true, id }) : api.delete(`/api/v1/transport/vehicles/${id}`).then(res => res.data),

  // Tour Management
  listTours: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockTours.slice((page - 1) * limit, page * limit), total: mockTours.length, page, limit })
      : api.get('/api/v1/transport/tours', { params: { page, limit } }).then(res => res.data),

  getTour: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockTours.find(t => t.id === id) || mockTours[0])
      : api.get(`/api/v1/transport/tours/${id}`).then(res => res.data),

  createTour: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()), status: 'planned', equipment_items: [] }) : api.post('/api/v1/transport/tours', data).then(res => res.data),

  updateTour: (id: string, data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id }) : api.put(`/api/v1/transport/tours/${id}`, data).then(res => res.data),

  // Tour Equipment Management
  addEquipmentToTour: (tourId: string, equipmentId: string, weight_kg: number, volume_m3: number) =>
    MOCK_MODE
      ? mockDelay({ id: String(Date.now()), tour_id: tourId, equipment_id: equipmentId, weight_kg, volume_m3, loaded_at: new Date().toISOString(), unloaded_at: null })
      : api.post(`/api/v1/transport/tours/${tourId}/equipment`, { equipment_id: equipmentId, weight_kg, volume_m3 }).then(res => res.data),

  removeEquipmentFromTour: (tourId: string, equipmentId: string) =>
    MOCK_MODE
      ? mockDelay({ success: true, tour_id: tourId, equipment_id: equipmentId })
      : api.delete(`/api/v1/transport/tours/${tourId}/equipment/${equipmentId}`).then(res => res.data),

  // Tour Status Management
  startTour: (tourId: string, kmStart: number) =>
    MOCK_MODE
      ? mockDelay({ id: tourId, status: 'in_transit', km_start: kmStart, started_at: new Date().toISOString() })
      : api.post(`/api/v1/transport/tours/${tourId}/start`, { km_start: kmStart }).then(res => res.data),

  completeTour: (tourId: string, kmEnd: number, notes?: string) =>
    MOCK_MODE
      ? mockDelay({ id: tourId, status: 'completed', km_end: kmEnd, completed_at: new Date().toISOString(), notes })
      : api.post(`/api/v1/transport/tours/${tourId}/complete`, { km_end: kmEnd, notes }).then(res => res.data),

  // Tour Capacity
  getTourCapacity: (tourId: string) =>
    MOCK_MODE
      ? mockDelay({ vehicle_capacity_kg: 18000, vehicle_capacity_m3: 52, used_kg: 8500, used_m3: 24, remaining_kg: 9500, remaining_m3: 28 })
      : api.get(`/api/v1/transport/tours/${tourId}/capacity`).then(res => res.data),

  // Driver Logs
  createDriverLog: (tourId: string, driverId: string, startTime: string, endTime: string, breakMinutes: number, kmDriven: number, notes?: string) =>
    MOCK_MODE
      ? mockDelay({ id: String(Date.now()), tour_id: tourId, driver_id: driverId, start_time: startTime, end_time: endTime, break_minutes: breakMinutes, km_driven: kmDriven, notes })
      : api.post(`/api/v1/transport/tours/${tourId}/driver-log`, { driver_id: driverId, start_time: startTime, end_time: endTime, break_minutes: breakMinutes, km_driven: kmDriven, notes }).then(res => res.data),

  getDriverLogs: (tourId: string) =>
    MOCK_MODE
      ? mockDelay([{ id: '1', tour_id: tourId, driver_id: 'user-2', start_time: '2026-03-20T06:00:00Z', end_time: '2026-03-20T14:30:00Z', break_minutes: 45, km_driven: 280, notes: 'Fahrt ohne Besonderheiten' }])
      : api.get(`/api/v1/transport/tours/${tourId}/driver-logs`).then(res => res.data),
}

// ============================================================================
// INSURANCE API endpoints
// ============================================================================

// Helper to get current tenant ID for insurance service (uses query params)
const getInsuranceTenantId = (): string => {
  return useAuthStore.getState().tenantId || ''
}

export const insuranceApi = {
  // Policies
  listPolicies: () =>
    api.get('/api/v1/policies', { params: { tenant_id: getInsuranceTenantId() } }).then(res => res.data),

  getPolicy: (id: string) =>
    api.get(`/api/v1/policies/${id}`).then(res => res.data),

  createPolicy: (data: object) =>
    api.post('/api/v1/policies', { ...data, tenant_id: getInsuranceTenantId() }).then(res => res.data),

  updatePolicy: (id: string, data: object) =>
    api.put(`/api/v1/policies/${id}`, data).then(res => res.data),

  getActivePolicies: () =>
    api.get('/api/v1/policies/active', { params: { tenant_id: getInsuranceTenantId() } }).then(res => res.data),

  // Claims
  listClaims: () =>
    api.get('/api/v1/claims', { params: { tenant_id: getInsuranceTenantId() } }).then(res => res.data),

  getClaim: (id: string) =>
    api.get(`/api/v1/claims/${id}`).then(res => res.data),

  createClaim: (data: object) =>
    api.post('/api/v1/claims', { ...data, tenant_id: getInsuranceTenantId() }).then(res => res.data),

  updateClaim: (id: string, data: object) =>
    api.put(`/api/v1/claims/${id}`, data).then(res => res.data),

  submitClaim: (id: string) =>
    api.post(`/api/v1/claims/${id}/submit`).then(res => res.data),

  approveClaim: (id: string, data?: object) =>
    api.post(`/api/v1/claims/${id}/approve`, data).then(res => res.data),

  rejectClaim: (id: string, data?: object) =>
    api.post(`/api/v1/claims/${id}/reject`, data).then(res => res.data),

  settleClaim: (id: string, data?: object) =>
    api.post(`/api/v1/claims/${id}/settle`, data).then(res => res.data),

  // Claim Items
  getClaimItems: (claimId: string) =>
    api.get(`/api/v1/claims/${claimId}/items`).then(res => res.data),

  createClaimItem: (claimId: string, data: object) =>
    api.post(`/api/v1/claims/${claimId}/items`, data).then(res => res.data),

  // Dashboard
  getDashboard: () =>
    api.get('/api/v1/insurance/dashboard', { params: { tenant_id: getInsuranceTenantId() } }).then(res => res.data),
}

// ============================================================================
// MAINTENANCE API endpoints
// ============================================================================
export const maintenanceApi = {
  // Maintenance Plans
  listPlans: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockMaintenancePlans.slice((page - 1) * limit, page * limit), total: mockMaintenancePlans.length, page, limit })
      : api.get('/api/v1/maintenance/plans', { params: { page, limit } }).then(res => res.data),

  getPlan: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockMaintenancePlans.find(p => p.id === id) || mockMaintenancePlans[0])
      : api.get(`/api/v1/maintenance/plans/${id}`).then(res => res.data),

  createPlan: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()), is_active: true }) : api.post('/api/v1/maintenance/plans', data).then(res => res.data),

  updatePlan: (id: string, data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id }) : api.put(`/api/v1/maintenance/plans/${id}`, data).then(res => res.data),

  deletePlan: (id: string) =>
    MOCK_MODE ? mockDelay({ success: true, id }) : api.delete(`/api/v1/maintenance/plans/${id}`).then(res => res.data),

  // Maintenance Tasks
  listTasks: (page = 1, limit = 50, status?: string) =>
    MOCK_MODE
      ? mockDelay({ items: mockMaintenanceTasks.filter(t => !status || t.status === status).slice((page - 1) * limit, page * limit), total: mockMaintenanceTasks.length, page, limit })
      : api.get('/api/v1/maintenance/tasks', { params: { page, limit, status } }).then(res => res.data),

  getTask: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockMaintenanceTasks.find(t => t.id === id) || mockMaintenanceTasks[0])
      : api.get(`/api/v1/maintenance/tasks/${id}`).then(res => res.data),

  createTask: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()), status: 'planned' }) : api.post('/api/v1/maintenance/tasks', data).then(res => res.data),

  startTask: (id: string, notes?: string) =>
    MOCK_MODE
      ? mockDelay({ id, status: 'in_progress', started_at: new Date().toISOString(), notes })
      : api.put(`/api/v1/maintenance/tasks/${id}/start`, { notes }).then(res => res.data),

  completeTask: (id: string, checklistData?: object, notes?: string) =>
    MOCK_MODE
      ? mockDelay({ id, status: 'completed', completed_at: new Date().toISOString(), checklist_data: checklistData, notes })
      : api.put(`/api/v1/maintenance/tasks/${id}/complete`, { checklist_data: checklistData, notes }).then(res => res.data),

  getDueTasks: () =>
    MOCK_MODE
      ? mockDelay(mockMaintenanceTasks.filter(t => t.status === 'planned'))
      : api.get('/api/v1/maintenance/tasks/due').then(res => res.data),

  getOverdueTasks: () =>
    MOCK_MODE
      ? mockDelay(mockMaintenanceTasks.filter(t => t.status === 'overdue'))
      : api.get('/api/v1/maintenance/tasks/overdue').then(res => res.data),

  // Checklists
  listChecklists: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockChecklists.slice((page - 1) * limit, page * limit), total: mockChecklists.length, page, limit })
      : api.get('/api/v1/maintenance/checklists', { params: { page, limit } }).then(res => res.data),

  getChecklist: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockChecklists.find(c => c.id === id) || mockChecklists[0])
      : api.get(`/api/v1/maintenance/checklists/${id}`).then(res => res.data),

  createChecklist: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()), version: 1 }) : api.post('/api/v1/maintenance/checklists', data).then(res => res.data),

  // Electrical Tests (VDE / DGUV V3)
  createElectricalTest: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()), certificate_number: `VDE-RF-${new Date().getFullYear()}-${String(Date.now()).slice(-3)}` }) : api.post('/api/v1/maintenance/electrical-tests', data).then(res => res.data),

  listElectricalTests: (page = 1, limit = 50, equipmentId?: string) =>
    MOCK_MODE
      ? mockDelay({ items: (equipmentId ? mockElectricalTests.filter(t => t.equipment_id === equipmentId) : mockElectricalTests).slice((page - 1) * limit, page * limit), total: mockElectricalTests.length, page, limit })
      : api.get('/api/v1/maintenance/electrical-tests', { params: { page, limit, ...(equipmentId ? { equipment_id: equipmentId } : {}) } }).then(res => res.data),

  // Izytron XML Import
  importIzytronXML: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return MOCK_MODE
      ? mockDelay({ success: true, imported_records: 45, warnings: 2 })
      : api.post('/api/v1/maintenance/import/izytron', formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        }).then(res => res.data)
  },

  // Dashboard
  getDashboard: () =>
    MOCK_MODE
      ? mockDelay({
          due_tasks: mockMaintenanceTasks.filter(t => t.status === 'planned'),
          overdue_tasks: mockMaintenanceTasks.filter(t => t.status === 'overdue'),
          due_plans: mockMaintenancePlans,
          recent_tests: mockElectricalTests,
          stats: {
            total_plans: mockMaintenancePlans.length,
            active_plans: mockMaintenancePlans.filter(p => p.is_active).length,
            tasks_due: 2,
            tasks_overdue: 1,
            tests_this_month: 2,
          },
        })
      : api.get('/api/v1/maintenance/dashboard').then(res => res.data),

  // Equipment Maintenance Status
  getEquipmentMaintenanceStatus: (equipmentId: string) =>
    MOCK_MODE
      ? mockDelay({
          equipment_id: equipmentId,
          last_maintenance: '2026-02-15T10:00:00Z',
          next_maintenance: '2026-05-15T10:00:00Z',
          status: 'ok',
          tests_passed: 2,
          tests_conditional: 0,
          tests_failed: 0,
        })
      : api.get(`/api/v1/maintenance/equipment/${equipmentId}/status`).then(res => res.data),
}

// ============================================================================
// WAREHOUSE API endpoints
// ============================================================================
export const warehouseApi = {
  // Warehouses
  listWarehouses: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockWarehouses.slice((page - 1) * limit, page * limit), total: mockWarehouses.length, page, limit })
      : api.get('/api/v1/warehouses', { params: { page, limit } }).then(res => res.data),

  getWarehouse: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockWarehouses.find(w => w.id === id) || mockWarehouses[0])
      : api.get(`/api/v1/warehouses/${id}`).then(res => res.data),

  createWarehouse: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()) }) : api.post('/api/v1/warehouses', data).then(res => res.data),

  // Zones
  listZones: (warehouseId: string) =>
    MOCK_MODE
      ? mockDelay(mockWarehouses.find(w => w.id === warehouseId)?.zones || [])
      : api.get(`/api/v1/warehouses/${warehouseId}/zones`).then(res => res.data),

  // Racks
  listRacks: (warehouseId: string, zoneId: string) =>
    MOCK_MODE
      ? mockDelay([
          { id: 'r1', zone_id: zoneId, name: 'Rack A1', capacity_kg: 5000, capacity_m3: 50, equipment_count: 3 },
          { id: 'r2', zone_id: zoneId, name: 'Rack A2', capacity_kg: 5000, capacity_m3: 50, equipment_count: 5 },
        ])
      : api.get(`/api/v1/warehouses/${warehouseId}/zones/${zoneId}/racks`).then(res => res.data),

  // Bays
  listBays: (warehouseId: string, zoneId: string, rackId: string) =>
    MOCK_MODE
      ? mockDelay([
          { id: 'b1', rack_id: rackId, name: 'Bay 1', capacity_kg: 500, capacity_m3: 5, equipment_count: 2 },
          { id: 'b2', rack_id: rackId, name: 'Bay 2', capacity_kg: 500, capacity_m3: 5, equipment_count: 1 },
        ])
      : api.get(`/api/v1/warehouses/${warehouseId}/zones/${zoneId}/racks/${rackId}/bays`).then(res => res.data),

  // Stock Locations
  listStockLocations: (warehouseId: string) =>
    MOCK_MODE
      ? mockDelay(mockStockMovements)
      : api.get(`/api/v1/warehouses/${warehouseId}/stock-locations`).then(res => res.data),

  // Movements
  listMovements: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockStockMovements.slice((page - 1) * limit, page * limit), total: mockStockMovements.length, page, limit })
      : api.get('/api/v1/warehouse/movements', { params: { page, limit } }).then(res => res.data),

  // Inventory Checks — matches backend POST /api/v1/inventory-checks etc.
  startInventoryCheck: (name: string, checkType: string, zoneId?: string) =>
    MOCK_MODE
      ? mockDelay({
          id: `inv_${Date.now()}`,
          name,
          status: 'in_progress',
          items: [],
          started_at: new Date().toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        })
      : api.post('/api/v1/inventory-checks', { name, check_type: checkType, zone_id: zoneId }).then(res => res.data),

  listInventoryChecks: (limit = 20, offset = 0) =>
    MOCK_MODE
      ? mockDelay({ data: [], total: 0, limit, offset })
      : api.get('/api/v1/inventory-checks', { params: { limit, offset } }).then(res => res.data),

  getInventoryCheck: (checkId: string) =>
    MOCK_MODE
      ? mockDelay({ id: checkId, name: 'Mock Check', status: 'in_progress', items: [], started_at: new Date().toISOString(), created_at: new Date().toISOString(), updated_at: new Date().toISOString() })
      : api.get(`/api/v1/inventory-checks/${checkId}`).then(res => res.data),

  scanInventoryItem: (checkId: string, equipmentId: string, locationId?: string, notes?: string) =>
    MOCK_MODE
      ? mockDelay({ id: checkId, status: 'in_progress', items: [{ equipment_id: equipmentId, expected_count: 1, actual_count: 1, status: 'found' }] })
      : api.post(`/api/v1/inventory-checks/${checkId}/scan`, { equipment_id: equipmentId, location_id: locationId, notes }).then(res => res.data),

  completeInventoryCheck: (checkId: string, completedBy: string) =>
    MOCK_MODE
      ? mockDelay({ id: checkId, status: 'completed', completed_at: new Date().toISOString() })
      : api.post(`/api/v1/inventory-checks/${checkId}/complete`, { completed_by: completedBy }).then(res => res.data),

  getDiscrepancies: (checkId: string) =>
    MOCK_MODE
      ? mockDelay({ check_id: checkId, discrepancy_count: 0, discrepancies: [] })
      : api.get(`/api/v1/inventory-checks/${checkId}/discrepancies`).then(res => res.data),

  // Location Tree: fetches warehouses with nested zones, racks, bays and equipment items
  getLocationTree: async () => {
    if (MOCK_MODE) {
      return mockDelay([
        {
          id: '1',
          name: 'Lager Muenchen - Hauptstandort',
          code: 'LGR-MUC',
          item_count: 42,
          zones: [
            {
              id: 'z1', name: 'Zone A - Audio', zone_type: 'audio', item_count: 18,
              racks: [
                {
                  id: 'r1', name: 'Regal 1', item_count: 8,
                  bays: [
                    { id: 'b1', name: 'Fach 1-1', item_count: 5, items: [
                      { id: '2', name: 'Shure SM58', quantity: 5, equipment_id: '2' },
                    ]},
                    { id: 'b2', name: 'Fach 1-2', item_count: 3, items: [
                      { id: 'xlr1', name: 'XLR-Kabel 10m', quantity: 3, equipment_id: '2' },
                    ]},
                  ],
                },
                {
                  id: 'r2', name: 'Regal 2', item_count: 10,
                  bays: [
                    { id: 'b3', name: 'Fach 2-1', item_count: 10, items: [
                      { id: 'di1', name: 'DI-Box', quantity: 10, equipment_id: '2' },
                    ]},
                  ],
                },
              ],
            },
            {
              id: 'z2', name: 'Zone B - Lichttechnik', zone_type: 'lighting', item_count: 15,
              racks: [
                {
                  id: 'r3', name: 'Regal 1', item_count: 9,
                  bays: [
                    { id: 'b4', name: 'Fach 1-1', item_count: 4, items: [
                      { id: '4', name: 'Martin MAC Aura XB', quantity: 4, equipment_id: '4' },
                    ]},
                    { id: 'b5', name: 'Fach 1-2', item_count: 5, items: [
                      { id: '10', name: 'Robe MegaPointe', quantity: 5, equipment_id: '10' },
                    ]},
                  ],
                },
                {
                  id: 'r4', name: 'Regal 2', item_count: 6,
                  bays: [
                    { id: 'b6', name: 'Fach 2-1', item_count: 3, items: [
                      { id: '3', name: 'MA Lighting grandMA3', quantity: 1, equipment_id: '3' },
                      { id: 'dmx1', name: 'DMX-Kabel 10m', quantity: 2, equipment_id: '3' },
                    ]},
                    { id: 'b7', name: 'Fach 2-2', item_count: 3, items: [
                      { id: 'led1', name: 'LED PAR 64', quantity: 3, equipment_id: '10' },
                    ]},
                  ],
                },
              ],
            },
            {
              id: 'z3', name: 'Zone C - Buehne & Rigging', zone_type: 'stage', item_count: 9,
              racks: [
                {
                  id: 'r5', name: 'Regal 1', item_count: 6,
                  bays: [
                    { id: 'b8', name: 'Fach 1-1', item_count: 6, items: [
                      { id: '6', name: 'Prolyte X30V Truss 3m', quantity: 6, equipment_id: '6' },
                    ]},
                  ],
                },
                {
                  id: 'r6', name: 'Regal 2', item_count: 3,
                  bays: [
                    { id: 'b9', name: 'Fach 2-1', item_count: 3, items: [
                      { id: '9', name: 'Chainmaster BGV-D8+ 1t', quantity: 3, equipment_id: '9' },
                    ]},
                  ],
                },
              ],
            },
          ],
        },
      ])
    }
    // Real API: fetch warehouses, then zones/racks/bays for each
    const whRes = await api.get('/api/v1/warehouses', { params: { limit: 50 } })
    const warehouses = whRes.data?.data || whRes.data?.items || []
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const tree = await Promise.all(warehouses.map(async (wh: any) => {
      const zonesRes = await api.get('/api/v1/warehouse/zones')
      const zones = zonesRes.data?.zones || []
      return {
        id: wh.id,
        name: wh.name,
        code: wh.code,
        item_count: wh.current_occupancy || 0,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        zones: zones.map((z: any) => ({
          id: z.id,
          name: z.name,
          zone_type: '',
          item_count: 0,
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          racks: (z.shelves || []).map((r: any) => ({
            id: r.id,
            name: r.name,
            item_count: r.expected_count || 0,
            bays: [],
          })),
        })),
      }
    }))
    return tree
  },
}

// ============================================================================
// SCANNER DEVICE MANAGEMENT API endpoints
// ============================================================================
export const scannerDeviceApi = {
  list: () =>
    MOCK_MODE
      ? mockDelay([
          { id: '1', name: 'Lager-Scanner 1', device_id: 'RF-SC-001', type: 'Handheld', is_online: true, last_seen: new Date().toISOString() },
          { id: '2', name: 'Lager-Scanner 2', device_id: 'RF-SC-002', type: 'Handheld', is_online: false, last_seen: '2026-03-23T14:30:00Z' },
          { id: '3', name: 'Tablet Laderampe', device_id: 'RF-TB-001', type: 'Tablet', is_online: true, last_seen: new Date().toISOString() },
        ])
      : api.get('/api/v1/scanner/devices').then(r => r.data),

  ring: (id: string) =>
    MOCK_MODE
      ? mockDelay({ success: true, device_id: id, acknowledged: true }, 1500)
      : api.post(`/api/v1/scanner/devices/${id}/ring`).then(r => r.data),

  generateQRToken: () =>
    MOCK_MODE
      ? mockDelay({ token: `rf-qr-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`, expires_in: 300 })
      : api.post('/api/v1/auth/qr-token').then(r => r.data),
}

// ============================================================================
// SCANNER API endpoints
// ============================================================================
export const scannerApi = {
  // Sessions
  startSession: (deviceId?: string) =>
    MOCK_MODE
      ? mockDelay({ session_id: String(Date.now()), device_id: deviceId, started_at: new Date().toISOString(), scans: [] })
      : api.post('/api/v1/scanner/sessions/start', { device_id: deviceId }).then(res => res.data),

  endSession: (sessionId: string) =>
    MOCK_MODE
      ? mockDelay({ session_id: sessionId, ended_at: new Date().toISOString(), total_scans: 5 })
      : api.put(`/api/v1/scanner/sessions/${sessionId}/end`, {}).then(res => res.data),

  // Scans
  processScan: (sessionId: string, barcode: string, quantity?: number) =>
    MOCK_MODE
      ? mockDelay({
          session_id: sessionId,
          barcode,
          equipment_id: '1',
          equipment_name: 'JBL VTX A12',
          quantity: quantity || 1,
          timestamp: new Date().toISOString(),
        })
      : api.post(`/api/v1/scanner/sessions/${sessionId}/scan`, { barcode, quantity }).then(res => res.data),

  processBatchScan: (sessionId: string, barcodes: Array<{ barcode: string; quantity?: number }>) =>
    MOCK_MODE
      ? mockDelay({
          session_id: sessionId,
          processed: barcodes.length,
          successful: barcodes.length,
          errors: [],
          timestamp: new Date().toISOString(),
        })
      : api.post(`/api/v1/scanner/sessions/${sessionId}/batch-scan`, { barcodes }).then(res => res.data),

  // Offline Sync
  syncOfflineScans: (sessionId: string, offlineData: object) =>
    MOCK_MODE
      ? mockDelay({ session_id: sessionId, synced_items: 10, conflicts: 0 })
      : api.post(`/api/v1/scanner/sessions/${sessionId}/sync-offline`, offlineData).then(res => res.data),

  // Protocol
  getSessionProtocol: (sessionId: string) =>
    MOCK_MODE
      ? mockDelay({
          session_id: sessionId,
          device_id: 'device-1',
          started_at: '2026-03-22T10:00:00Z',
          ended_at: '2026-03-22T10:45:00Z',
          total_scans: 5,
          scans: [
            { barcode: 'RF-SPK-001', equipment_name: 'JBL VTX A12', quantity: 2, timestamp: '2026-03-22T10:05:00Z' },
            { barcode: 'RF-MIC-001', equipment_name: 'Shure SM58', quantity: 10, timestamp: '2026-03-22T10:10:00Z' },
          ],
        })
      : api.get(`/api/v1/scanner/sessions/${sessionId}/protocol`).then(res => res.data),
}

// ============================================================================
// PHASE 4 API ENDPOINTS
// ============================================================================

// AI Service (port 8013)
export const aiApi = {
  chat: (message: string, provider: string) =>
    MOCK_MODE
      ? mockDelay({ id: String(Date.now()), role: 'assistant', content: `Mock-Antwort von ${provider}: "${message}" wurde verarbeitet.`, timestamp: new Date().toISOString() })
      : api.post('/api/v1/ai/complete', { provider_id: provider, request_type: 'general', input_text: message }).then(res => res.data),

  providers: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/ai/providers').then(res => {
          const data = res.data
          return Array.isArray(data) ? data : (data?.providers ?? [])
        }),

  history: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/ai/requests').then(res => {
          const data = res.data
          return Array.isArray(data) ? data : (data?.requests ?? [])
        }),

  feedback: (requestId: string, rating: number) =>
    MOCK_MODE
      ? mockDelay({ success: true, request_id: requestId, rating })
      : api.post('/api/v1/ai/feedback', { request_id: requestId, rating }).then(res => res.data),

  testProvider: (providerName: string, _apiKey?: string, _model?: string, _endpoint?: string) =>
    MOCK_MODE
      ? mockDelay({ success: true, message: `${providerName} antwortet korrekt.` })
      : api.post('/api/v1/ai/complete', {
          provider_id: providerName,
          request_type: 'general',
          input_text: 'Antworte mit genau einem Wort: OK',
        }).then(res => ({ success: true, message: `${providerName} antwortet korrekt.`, data: res.data }))
          .catch((err: any) => ({ success: false, message: err?.response?.data?.error || err?.message || 'Verbindung fehlgeschlagen' })),

  registerProvider: (data: { name: string; model_name: string; api_endpoint?: string; is_active: boolean; priority: number; config?: Record<string, any> }) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/ai/providers', data).then(res => res.data),

  analyzeReceipt: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return MOCK_MODE
      ? mockDelay({
          is_invoice: true,
          document_type: 'invoice',
          confidence: 0.95,
          vendor: 'Mock GmbH',
          gross_amount: 119.00,
          net_amount: 100.00,
          tax_rate: 19.0,
          tax_amount: 19.00,
          currency: 'EUR',
          invoice_date: '2026-03-20',
        })
      : api.post('/api/v1/ai/analyze-receipt', formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
          timeout: 90000,
        }).then(res => res.data)
  },
}

// Workflow Service (port 8014)
export const workflowApi = {
  list: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/workflows').then(res => Array.isArray(res.data) ? res.data : []),

  get: (id: string) =>
    MOCK_MODE
      ? mockDelay({ id, name: 'Workflow', status: 'active' })
      : api.get(`/api/v1/workflows/${id}`).then(res => res.data),

  create: (data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/workflows', data).then(res => res.data),

  execute: (id: string) =>
    MOCK_MODE
      ? mockDelay({ success: true, workflow_id: id, instance_id: String(Date.now()) })
      : api.post(`/api/v1/workflows/${id}/execute`).then(res => res.data),

  instances: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/workflow-instances').then(res => Array.isArray(res.data) ? res.data : []),

  templates: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/workflow-templates').then(res => Array.isArray(res.data) ? res.data : []),
}

// Federation Service (port 8015)
export const federationApi = {
  partners: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/federation/partners').then(res => res.data),

  addPartner: (data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/federation/partners', data).then(res => res.data),

  equipment: (partnerId: string) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get(`/api/v1/federation/partners/${partnerId}/equipment`).then(res => res.data),

  subRentals: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/federation/sub-rentals').then(res => res.data),

  createSubRental: (data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/federation/sub-rentals', data).then(res => res.data),
}

// ============================================================================
// Notification Preferences API (notification-service port 8015)
// ============================================================================
export interface NotificationPreference {
  id?: string
  event_type: string
  channels: string[]
  is_enabled: boolean
  quiet_hours_start?: string | null
  quiet_hours_end?: string | null
  digest_mode?: string
}

// The notification-service returns PascalCase (domain struct without json tags).
// Normalise to snake_case so the frontend can work with a consistent interface.
function normalisePref(raw: any): NotificationPreference {
  return {
    id: raw.ID ?? raw.id,
    event_type: raw.EventType ?? raw.event_type ?? '',
    channels: raw.Channels ?? raw.channels ?? [],
    is_enabled: raw.IsEnabled ?? raw.is_enabled ?? false,
    quiet_hours_start: raw.QuietHoursStart ?? raw.quiet_hours_start ?? null,
    quiet_hours_end: raw.QuietHoursEnd ?? raw.quiet_hours_end ?? null,
    digest_mode: raw.DigestMode ?? raw.digest_mode ?? '',
  }
}

export const notificationPreferencesApi = {
  list: (): Promise<NotificationPreference[]> =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/notifications/preferences').then(res => {
          const raw = Array.isArray(res.data) ? res.data : res.data ?? []
          return raw.map(normalisePref)
        }),

  update: (pref: NotificationPreference): Promise<NotificationPreference> =>
    MOCK_MODE
      ? mockDelay(pref)
      : api.put('/api/v1/notifications/preferences', pref).then(res => normalisePref(res.data)),
}

// ============================================================================
// Time Tracking API (crew-service port 8008)
// ============================================================================
export const timeTrackingApi = {
  listEntries: (params?: any) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/time-entries', { params }).then(res => res.data),

  createEntry: (data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/time-entries', data).then(res => res.data),

  listAbsences: (params?: any) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/absences', { params }).then(res => res.data),

  createAbsence: (data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/absences', data).then(res => res.data),
}

// ============================================================================
// Crew Service (port 8008)
// ============================================================================
export const crewApi = {
  listMembers: (params: { page?: number; per_page?: number } = {}) =>
    MOCK_MODE
      ? mockDelay({ data: [], page: 1, per_page: 20, total: 0, total_pages: 0 })
      : api.get('/api/v1/crew', { params }).then(res => res.data),

  getMember: (id: string) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.get(`/api/v1/crew/${id}`).then(res => res.data),

  createMember: (data: {
    first_name: string
    last_name: string
    email: string
    phone?: string
    role?: string
    status?: string
    hourly_rate?: number
    daily_rate?: number
    emergency_contact?: string
    notes?: string
  }) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/crew', data).then(res => res.data),

  updateMember: (id: string, data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id })
      : api.put(`/api/v1/crew/${id}`, data).then(res => res.data),

  deleteMember: (id: string) =>
    MOCK_MODE
      ? mockDelay(undefined)
      : api.delete(`/api/v1/crew/${id}`).then(() => undefined),

  getQualifications: (memberId: string) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get(`/api/v1/crew/${memberId}/qualifications`).then(res => res.data),

  createQualification: (memberId: string, data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post(`/api/v1/crew/${memberId}/qualifications`, data).then(res => res.data),

  checkAvailability: (memberId: string, start: string, end: string) =>
    MOCK_MODE
      ? mockDelay({ crew_member_id: memberId, is_available: true })
      : api.get(`/api/v1/crew/${memberId}/availability`, { params: { start, end } }).then(res => res.data),

  listAssignments: (params: { page?: number; per_page?: number } = {}) =>
    MOCK_MODE
      ? mockDelay({ data: [], page: 1, per_page: 20, total: 0, total_pages: 0 })
      : api.get('/api/v1/crew/assignments', { params }).then(res => res.data),

  getDrivers: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/crew/drivers').then(res => res.data),

  getDashboard: () =>
    MOCK_MODE
      ? mockDelay(null)
      : api.get('/api/v1/crew/dashboard').then(res => res.data),

  checkConflicts: (memberId: string, start: string, end: string) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/crew/assignments/conflicts', {
          params: { member_id: memberId, start, end },
        }).then(res => {
          const data = res.data
          return Array.isArray(data) ? data : (data?.data ?? data?.items ?? [])
        }),

  createAssignment: (data: {
    project_id: string
    crew_member_id: string
    role: string
    start_date?: string
    end_date?: string
  }) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/crew/assignments', data).then(res => res.data),
}

// ============================================================================
// Booking API (Freelancer Booking System)
// ============================================================================
export const bookingApi = {
  create: (data: { assignment_id: string; message?: string }) =>
    api.post('/api/v1/crew-assignments', data).then(r => r.data),

  list: () =>
    api.get('/api/v1/crew-assignments').then(r => r.data).catch(() => []),

  // Public endpoints - use raw axios (no auth interceptor)
  getDetails: (token: string) =>
    axios.get(`${API_BASE_URL}/api/v1/crew-assignments/${token}/details`).then(r => r.data),

  respond: (token: string, data: { status: string; message?: string }) =>
    axios.post(`${API_BASE_URL}/api/v1/crew-assignments/${token}/respond`, data).then(r => r.data),
}

// ============================================================================
// Document Service (port 8007)
// ============================================================================
export const documentApi = {
  list: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/documents').then(res => res.data),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.get(`/api/v1/documents/${id}`).then(res => res.data),

  create: (data: {
    document_type: string
    reference_id: string
    document_number: string
    title: string
    template_id?: string
    metadata?: Record<string, any>
  }) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post('/api/v1/documents', data).then(res => res.data),

  update: (id: string, data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id })
      : api.put(`/api/v1/documents/${id}`, data).then(res => res.data),

  archive: (id: string) =>
    MOCK_MODE
      ? mockDelay(undefined)
      : api.post(`/api/v1/documents/${id}/archive`).then(() => undefined),

  generate: (id: string, data: { template_id: string; data: Record<string, any> }) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.post(`/api/v1/documents/${id}/generate`, data).then(res => res.data),

  requestSignature: (id: string, data: { signer_name: string; signer_email: string; signer_role: string }) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.post(`/api/v1/documents/${id}/sign`, data).then(res => res.data),

  getSignatures: (id: string) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get(`/api/v1/documents/${id}/signatures`).then(res => res.data),

  getVersions: (id: string) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get(`/api/v1/documents/${id}/versions`).then(res => res.data),

  verifyChecksumChain: () =>
    MOCK_MODE
      ? mockDelay({ status: 'ok', document_count: 0, integrity_valid: true, errors: [] })
      : api.get('/api/v1/documents/verify-chain').then(res => res.data),

  generateDeliveryNote: (projectId: string) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.post(`/api/v1/documents/actions/from-project/${projectId}`).then(res => res.data),

  uploadScan: (formData: FormData) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.post('/api/v1/documents/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        }).then(res => res.data),
}

// Audit Service (port 8017)
export const auditApi = {
  logs: (params?: any) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/audit/logs', { params }).then(res => res.data),

  verify: () =>
    MOCK_MODE
      ? mockDelay({ valid: true, entries_checked: 0, invalid_entries: 0, timestamp: new Date().toISOString(), verification_hash: '' })
      : api.post('/api/v1/audit/verify-chain').then(res => res.data),

  export: (params: any) =>
    MOCK_MODE
      ? mockDelay({ success: true, export_id: String(Date.now()) })
      : api.post('/api/v1/audit/export', params).then(res => res.data),

  exports: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/audit/exports').then(res => res.data),
}

// Admin endpoints
const SERVICE_HEALTH_ENDPOINTS: Record<string, string> = {
  'auth': '/api/health/auth',
  'inventory': '/api/health/inventory',
  'project': '/api/health/project',
  'scanner': '/api/health/scanner',
  'warehouse': '/api/health/warehouse',
  'invoice': '/api/health/invoice',
  'document': '/api/health/document',
  'crew': '/api/health/crew',
  'federation': '/api/health/federation',
  'maintenance': '/api/health/maintenance',
  'transport': '/api/health/transport',
  'insurance': '/api/health/insurance',
  'workflow': '/api/health/workflow',
  'ai': '/api/health/ai',
  'notification': '/api/health/notification',
  'reporting': '/api/health/reporting',
  'audit': '/api/health/audit',
  'expense': '/api/health/expense',
}

export const adminApi = {
  health: async () => {
    if (MOCK_MODE) return mockDelay({ status: 'operational', services: [] })
    const results = await Promise.allSettled(
      Object.entries(SERVICE_HEALTH_ENDPOINTS).map(async ([name, url]) => {
        try {
          const res = await api.get(url, { timeout: 3000 })
          return { name, status: 'healthy', data: res.data }
        } catch {
          return { name, status: 'unhealthy', data: null }
        }
      })
    )
    const services = results.map(r => r.status === 'fulfilled' ? r.value : { name: 'unknown', status: 'error', data: null })
    const allHealthy = services.every(s => s.status === 'healthy')
    return { status: allHealthy ? 'operational' : 'degraded', services }
  },

  settings: () =>
    MOCK_MODE
      ? mockDelay({ company_name: 'JE-Sound&Light', timezone: 'Europe/Berlin', language: 'de-DE' })
      : api.get('/api/v1/admin/settings').then(res => res.data).catch(() => ({ company_name: '', timezone: 'Europe/Berlin', language: 'de-DE' })),

  updateSettings: (data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, updated_at: new Date().toISOString() })
      : api.put('/api/v1/admin/settings', data).then(res => res.data),
}


// ============================================================================
// Setup Service
// ============================================================================
export interface SetupRequest {
  company_name: string
  company_address: string
  company_slug: string
  currency: string
  tax_rate: number
  invoice_prefix: string
  admin_email: string
  admin_password: string
  admin_name: string
  language: string
  setup_token: string
}

export const setupApi = {
  getStatus: () =>
    MOCK_MODE
      ? mockDelay({ is_completed: true, has_setup_token: false })
      : api.get('/api/v1/setup/status').then(res => res.data).catch(() => ({ is_completed: true, has_setup_token: false })),

  complete: (data: SetupRequest) =>
    MOCK_MODE
      ? mockDelay({ success: true, message: 'Setup completed successfully' })
      : api.post('/api/v1/setup/complete', data).then(res => res.data),
}

// ============================================================================
// CONFIG API - Settings/Config key-value store
// ============================================================================
export const configApi = {
  getAll: () =>
    MOCK_MODE
      ? mockDelay({})
      : api.get('/api/v1/config').then(res => res.data),

  get: (key: string) =>
    MOCK_MODE
      ? mockDelay({})
      : api.get(`/api/v1/config/${key}`).then(res => res.data),

  set: (key: string, value: any) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.put(`/api/v1/config/${key}`, value).then(res => res.data),

  testSmtp: () =>
    MOCK_MODE
      ? mockDelay({ success: true, message: 'Test email sent' })
      : api.post('/api/v1/config/smtp-test').then(res => res.data),

  backupNow: () =>
    MOCK_MODE
      ? mockDelay({ success: true, message: 'Backup started' })
      : api.post('/api/v1/system/backup').then(res => res.data),

  backupList: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/system/backups').then(res => res.data),

  backupDownload: (filename: string) =>
    api.get(`/api/v1/system/backups/${encodeURIComponent(filename)}`, {
      responseType: 'blob',
    }).then(res => {
      const url = window.URL.createObjectURL(res.data)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    }),

  backupRestore: (filename: string) =>
    MOCK_MODE
      ? mockDelay({ success: true, message: 'Backup restored' })
      : api.post(`/api/v1/system/backups/${encodeURIComponent(filename)}/restore`).then(res => res.data),

  backupDelete: (filename: string) =>
    MOCK_MODE
      ? mockDelay({ success: true })
      : api.delete(`/api/v1/system/backups/${encodeURIComponent(filename)}`).then(res => res.data),

  backupUpload: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return MOCK_MODE
      ? mockDelay({ success: true, message: 'Backup imported' })
      : api.post('/api/v1/system/backups/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        }).then(res => res.data)
  },
}

// ============================================================================
// SYSTEM API - Version/Update check
// ============================================================================
export const systemApi = {
  getVersion: () =>
    MOCK_MODE
      ? mockDelay({
          current_version: '1.0.0',
          latest_version: '1.0.1',
          update_available: true,
          release_url: 'https://github.com/jeckersberger/CrateDesk/releases/tag/v1.0.1',
          release_notes: 'Bug fixes and new features',
          checked_at: new Date().toISOString(),
        })
      : api.get('/api/v1/system/version').then(r => r.data),

  triggerUpdate: () =>
    MOCK_MODE
      ? mockDelay({ success: true, output: 'Mock update completed' })
      : api.post('/api/v1/system/update').then(r => r.data),
}

// ============================================================================
// Expense API (port 8018)
// ============================================================================
export const expenseApi = {
  list: (params: { page?: number; limit?: number; project_id?: string; category?: string } = {}) =>
    MOCK_MODE
      ? mockDelay({ items: [], total: 0 })
      : api.get('/api/v1/expenses', { params }).then(res => res.data),

  getById: (id: string) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.get(`/api/v1/expenses/${id}`).then(res => res.data),

  create: (data: {
    date: string
    description: string
    amount: number
    vat_rate: number
    category: string
    project_id?: string
    notes?: string
  }) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()), created_at: new Date().toISOString() })
      : api.post('/api/v1/expenses', data).then(res => res.data),

  update: (id: string, data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id })
      : api.put(`/api/v1/expenses/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    MOCK_MODE
      ? mockDelay(undefined)
      : api.delete(`/api/v1/expenses/${id}`).then(() => undefined),

  uploadReceipt: (id: string, formData: FormData) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.post(`/api/v1/expenses/${id}/receipt`, formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        }).then(res => res.data),

  getSummary: (params: { year?: number; quarter?: number } = {}) =>
    MOCK_MODE
      ? mockDelay({ income: 0, expenses: 0, profit: 0, by_month: [] })
      : api.get('/api/v1/expenses/summary', { params }).then(res => res.data),
}
