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
}

export default api
