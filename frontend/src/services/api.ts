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

// Auth API endpoints
export const authApi = {
  login: (email: string, password: string) =>
    api.post('/api/v1/auth/login', { email, password }).then(res => res.data),

  register: (email: string, password: string, name: string) =>
    api.post('/api/v1/auth/register', { email, password, name }).then(res => res.data),

  logout: () =>
    api.post('/api/v1/auth/logout').then(res => res.data),

  getCurrentUser: () =>
    api.get('/api/v1/auth/me').then(res => res.data),

  refreshToken: () =>
    api.post('/api/v1/auth/refresh').then(res => res.data),
}

// Equipment API endpoints
export const equipmentApi = {
  list: (page = 1, limit = 50) =>
    api.get('/api/v1/equipment', { params: { page, limit } }).then(res => res.data),

  getById: (id: string) =>
    api.get(`/api/v1/equipment/${id}`).then(res => res.data),

  create: (data: unknown) =>
    api.post('/api/v1/equipment', data).then(res => res.data),

  update: (id: string, data: unknown) =>
    api.put(`/api/v1/equipment/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    api.delete(`/api/v1/equipment/${id}`).then(res => res.data),

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
    api.get(`/api/v1/equipment/barcode/${barcode}`).then(res => res.data),

  // Equipment per Suchbegriff finden
  search: (query: string, limit = 20) =>
    api.get('/api/v1/equipment/search', { params: { q: query, limit } }).then(res => res.data),

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
    api.get('/api/v1/projects', { params: { page, limit } }).then(res => res.data),

  getById: (id: string) =>
    api.get(`/api/v1/projects/${id}`).then(res => res.data),

  create: (data: unknown) =>
    api.post('/api/v1/projects', data).then(res => res.data),

  update: (id: string, data: unknown) =>
    api.put(`/api/v1/projects/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    api.delete(`/api/v1/projects/${id}`).then(res => res.data),

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
    api.get('/api/v1/invoices', { params: { page, limit } }).then(res => res.data),

  getById: (id: string) =>
    api.get(`/api/v1/invoices/${id}`).then(res => res.data),

  create: (data: unknown) =>
    api.post('/api/v1/invoices', data).then(res => res.data),

  update: (id: string, data: unknown) =>
    api.put(`/api/v1/invoices/${id}`, data).then(res => res.data),

  delete: (id: string) =>
    api.delete(`/api/v1/invoices/${id}`).then(res => res.data),

  getOpen: () =>
    api.get('/api/v1/invoices/open', {}).then(res => res.data),

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
