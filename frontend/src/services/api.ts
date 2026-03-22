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
    const res = await api.post('/api/v1/auth/login', { email, password })
    const accessToken = res.data?.data?.access_token || res.data?.token || res.data?.access_token
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
  { id: '1', invoice_number: 'RF-2026-001', client: 'AutoBrand AG', project_name: 'Produktlaunch AutoBrand', status: 'paid', amount: 32000, tax: 6080, total: 38080, issue_date: '2026-03-01', due_date: '2026-03-31', paid_date: '2026-03-15', items: [{ description: 'LED-Wall 6x3m (3 Tage)', quantity: 1, price: 4500, total: 4500 }, { description: 'Line Array System (3 Tage)', quantity: 2, price: 3600, total: 7200 }, { description: 'Lichttechnik Paket', quantity: 1, price: 8500, total: 8500 }, { description: 'Techniker (3 Tage)', quantity: 4, price: 2950, total: 11800 }] },
  { id: '2', invoice_number: 'RF-2026-002', client: 'Familie Weber', project_name: 'Hochzeit Familie Weber', status: 'paid', amount: 8500, tax: 1615, total: 10115, issue_date: '2026-03-09', due_date: '2026-04-09', paid_date: '2026-03-20', items: [{ description: 'DJ-Setup komplett', quantity: 1, price: 3500, total: 3500 }, { description: 'Ambientebeleuchtung', quantity: 1, price: 2800, total: 2800 }, { description: 'Techniker', quantity: 2, price: 1100, total: 2200 }] },
  { id: '3', invoice_number: 'RF-2026-003', client: 'Stadt München', project_name: 'Stadtfest München 2026', status: 'draft', amount: 45000, tax: 8550, total: 53550, issue_date: '2026-03-22', due_date: '2026-04-22', paid_date: null, items: [{ description: 'PA-System Hauptbühne', quantity: 1, price: 12000, total: 12000 }, { description: 'Lichttechnik 3 Bühnen', quantity: 1, price: 15000, total: 15000 }, { description: 'Bühne + Truss', quantity: 1, price: 8000, total: 8000 }, { description: 'Techniker-Team (4 Tage)', quantity: 8, price: 1250, total: 10000 }] },
  { id: '4', invoice_number: 'RF-2026-004', client: 'TechCorp GmbH', project_name: 'Firmen-Gala TechCorp', status: 'sent', amount: 18500, tax: 3515, total: 22015, issue_date: '2026-03-18', due_date: '2026-04-18', paid_date: null, items: [{ description: 'Audio-Paket Gala', quantity: 1, price: 6500, total: 6500 }, { description: 'Licht-Design Gala', quantity: 1, price: 7000, total: 7000 }, { description: 'Video/Streaming', quantity: 1, price: 5000, total: 5000 }] },
  { id: '5', invoice_number: 'RF-2026-005', client: 'Festival GmbH', project_name: 'Open Air Festival Bodensee', status: 'overdue', amount: 25000, tax: 4750, total: 29750, issue_date: '2026-02-01', due_date: '2026-03-01', paid_date: null, items: [{ description: 'Anzahlung Festival-Paket (30%)', quantity: 1, price: 25000, total: 25000 }] },
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
  list: (params: { limit?: number; offset?: number; status?: string; category_id?: string; location_id?: string } = {}) =>
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
}

// Tenant API endpoints
export const tenantApi = {
  getById: (id: string) =>
    api.get(`/api/v1/tenants/${id}`).then(res => res.data),

  update: (id: string, data: object) =>
    api.put(`/api/v1/tenants/${id}`, data).then(res => res.data),
}

// User API endpoints (auth-service)
export const userApi = {
  list: () =>
    api.get('/api/v1/users').then(res => res.data),
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
    api.post(`/api/v1/invoices/actions/from-project/${projectId}`, {}).then(res => res.data),

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

// ============================================================================
// TRANSPORT/FLEET API endpoints
// ============================================================================
export const transportApi = {
  // Vehicle Management
  listVehicles: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockVehicles.slice((page - 1) * limit, page * limit), total: mockVehicles.length, page, limit })
      : api.get('/api/v1/vehicles', { params: { page, limit } }).then(res => res.data),

  getVehicle: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockVehicles.find(v => v.id === id) || mockVehicles[0])
      : api.get(`/api/v1/vehicles/${id}`).then(res => res.data),

  createVehicle: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()), status: 'available' }) : api.post('/api/v1/vehicles', data).then(res => res.data),

  updateVehicle: (id: string, data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id }) : api.put(`/api/v1/vehicles/${id}`, data).then(res => res.data),

  deleteVehicle: (id: string) =>
    MOCK_MODE ? mockDelay({ success: true, id }) : api.delete(`/api/v1/vehicles/${id}`).then(res => res.data),

  // Tour Management
  listTours: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockTours.slice((page - 1) * limit, page * limit), total: mockTours.length, page, limit })
      : api.get('/api/v1/tours', { params: { page, limit } }).then(res => res.data),

  getTour: (id: string) =>
    MOCK_MODE
      ? mockDelay(mockTours.find(t => t.id === id) || mockTours[0])
      : api.get(`/api/v1/tours/${id}`).then(res => res.data),

  createTour: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()), status: 'planned', equipment_items: [] }) : api.post('/api/v1/tours', data).then(res => res.data),

  updateTour: (id: string, data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id }) : api.put(`/api/v1/tours/${id}`, data).then(res => res.data),

  // Tour Equipment Management
  addEquipmentToTour: (tourId: string, equipmentId: string, weight_kg: number, volume_m3: number) =>
    MOCK_MODE
      ? mockDelay({ id: String(Date.now()), tour_id: tourId, equipment_id: equipmentId, weight_kg, volume_m3, loaded_at: new Date().toISOString(), unloaded_at: null })
      : api.post(`/api/v1/tours/${tourId}/equipment`, { equipment_id: equipmentId, weight_kg, volume_m3 }).then(res => res.data),

  removeEquipmentFromTour: (tourId: string, equipmentId: string) =>
    MOCK_MODE
      ? mockDelay({ success: true, tour_id: tourId, equipment_id: equipmentId })
      : api.delete(`/api/v1/tours/${tourId}/equipment/${equipmentId}`).then(res => res.data),

  // Tour Status Management
  startTour: (tourId: string, kmStart: number) =>
    MOCK_MODE
      ? mockDelay({ id: tourId, status: 'in_transit', km_start: kmStart, started_at: new Date().toISOString() })
      : api.post(`/api/v1/tours/${tourId}/start`, { km_start: kmStart }).then(res => res.data),

  completeTour: (tourId: string, kmEnd: number, notes?: string) =>
    MOCK_MODE
      ? mockDelay({ id: tourId, status: 'completed', km_end: kmEnd, completed_at: new Date().toISOString(), notes })
      : api.post(`/api/v1/tours/${tourId}/complete`, { km_end: kmEnd, notes }).then(res => res.data),

  // Tour Capacity
  getTourCapacity: (tourId: string) =>
    MOCK_MODE
      ? mockDelay({ vehicle_capacity_kg: 18000, vehicle_capacity_m3: 52, used_kg: 8500, used_m3: 24, remaining_kg: 9500, remaining_m3: 28 })
      : api.get(`/api/v1/tours/${tourId}/capacity`).then(res => res.data),

  // Driver Logs
  createDriverLog: (tourId: string, driverId: string, startTime: string, endTime: string, breakMinutes: number, kmDriven: number, notes?: string) =>
    MOCK_MODE
      ? mockDelay({ id: String(Date.now()), tour_id: tourId, driver_id: driverId, start_time: startTime, end_time: endTime, break_minutes: breakMinutes, km_driven: kmDriven, notes })
      : api.post(`/api/v1/tours/${tourId}/driver-log`, { driver_id: driverId, start_time: startTime, end_time: endTime, break_minutes: breakMinutes, km_driven: kmDriven, notes }).then(res => res.data),

  getDriverLogs: (tourId: string) =>
    MOCK_MODE
      ? mockDelay([{ id: '1', tour_id: tourId, driver_id: 'user-2', start_time: '2026-03-20T06:00:00Z', end_time: '2026-03-20T14:30:00Z', break_minutes: 45, km_driven: 280, notes: 'Fahrt ohne Besonderheiten' }])
      : api.get(`/api/v1/tours/${tourId}/driver-logs`).then(res => res.data),
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
      : api.post(`/api/v1/maintenance/tasks/${id}/start`, { notes }).then(res => res.data),

  completeTask: (id: string, checklistData?: object, notes?: string) =>
    MOCK_MODE
      ? mockDelay({ id, status: 'completed', completed_at: new Date().toISOString(), checklist_data: checklistData, notes })
      : api.post(`/api/v1/maintenance/tasks/${id}/complete`, { checklist_data: checklistData, notes }).then(res => res.data),

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

  // Electrical Tests (VDE)
  createElectricalTest: (data: object) =>
    MOCK_MODE ? mockDelay({ ...data, id: String(Date.now()) }) : api.post('/api/v1/maintenance/tests/electrical', data).then(res => res.data),

  listElectricalTests: (page = 1, limit = 50) =>
    MOCK_MODE
      ? mockDelay({ items: mockElectricalTests.slice((page - 1) * limit, page * limit), total: mockElectricalTests.length, page, limit })
      : api.get('/api/v1/maintenance/tests/electrical', { params: { page, limit } }).then(res => res.data),

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

  // Inventory Check
  startInventoryCheck: (warehouseId: string, zoneId?: string) =>
    MOCK_MODE
      ? mockDelay({ id: String(Date.now()), warehouse_id: warehouseId, zone_id: zoneId, status: 'in_progress', started_at: new Date().toISOString() })
      : api.post(`/api/v1/warehouses/${warehouseId}/inventory-check/start`, { zone_id: zoneId }).then(res => res.data),

  completeInventoryCheck: (checkId: string, discrepancies?: object) =>
    MOCK_MODE
      ? mockDelay({ id: checkId, status: 'completed', completed_at: new Date().toISOString(), discrepancies })
      : api.post(`/api/v1/warehouse/inventory-check/${checkId}/complete`, { discrepancies }).then(res => res.data),
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
      ? mockDelay({ id: String(Date.now()), role: 'assistant', content: `Response from ${provider}`, timestamp: new Date().toISOString() })
      : api.post('/api/v1/ai/chat', { message, provider }).then(res => res.data),

  providers: () =>
    MOCK_MODE
      ? mockDelay(['Claude', 'GPT-4o', 'Gemini', 'Mistral', 'Ollama'])
      : api.get('/api/v1/ai/providers').then(res => res.data),

  history: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/ai/requests').then(res => res.data),

  feedback: (requestId: string, rating: number) =>
    MOCK_MODE
      ? mockDelay({ success: true, request_id: requestId, rating })
      : api.post('/api/v1/ai/feedback', { request_id: requestId, rating }).then(res => res.data),
}

// Workflow Service (port 8014)
export const workflowApi = {
  list: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/workflows').then(res => res.data),

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
      : api.get('/api/v1/workflow-instances').then(res => res.data),

  templates: () =>
    MOCK_MODE
      ? mockDelay([])
      : api.get('/api/v1/workflow-templates').then(res => res.data),
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
// Crew Service (port 8008)
// ============================================================================
export const crewApi = {
  listMembers: (params: { page?: number; per_page?: number } = {}) =>
    MOCK_MODE
      ? mockDelay({ data: [], page: 1, per_page: 20, total: 0, total_pages: 0 })
      : api.get('/api/v1/crew/members', { params }).then(res => res.data),

  getMember: (id: string) =>
    MOCK_MODE
      ? mockDelay(null)
      : api.get(`/api/v1/crew/members/${id}`).then(res => res.data),

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
      : api.post('/api/v1/crew/members', data).then(res => res.data),

  updateMember: (id: string, data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id })
      : api.put(`/api/v1/crew/members/${id}`, data).then(res => res.data),

  deleteMember: (id: string) =>
    MOCK_MODE
      ? mockDelay(undefined)
      : api.delete(`/api/v1/crew/members/${id}`).then(() => undefined),

  getQualifications: (memberId: string) =>
    MOCK_MODE
      ? mockDelay([])
      : api.get(`/api/v1/crew/members/${memberId}/qualifications`).then(res => res.data),

  createQualification: (memberId: string, data: any) =>
    MOCK_MODE
      ? mockDelay({ ...data, id: String(Date.now()) })
      : api.post(`/api/v1/crew/members/${memberId}/qualifications`, data).then(res => res.data),

  checkAvailability: (memberId: string, start: string, end: string) =>
    MOCK_MODE
      ? mockDelay({ crew_member_id: memberId, is_available: true })
      : api.get(`/api/v1/crew/members/${memberId}/availability`, { params: { start, end } }).then(res => res.data),

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
      ? mockDelay({ company_name: 'RentFlow GmbH', timezone: 'Europe/Berlin', language: 'de-DE' })
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
      ? mockDelay({ is_completed: false, has_setup_token: true })
      : api.get('/api/v1/setup/status').then(res => res.data),

  complete: (data: SetupRequest) =>
    MOCK_MODE
      ? mockDelay({ success: true, message: 'Setup completed successfully' })
      : api.post('/api/v1/setup/complete', data).then(res => res.data),
}
