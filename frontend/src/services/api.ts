import axios from 'axios';

const api = axios.create({
  baseURL: '',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor: attach Bearer token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('cd_access_token');
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: unwrap envelope {data: ..., meta: ...}, handle 401
api.interceptors.response.use(
  (response) => {
    const body = response.data;
    return body?.data !== undefined ? body.data : body;
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('cd_access_token');
      localStorage.removeItem('cd_user');
      window.location.href = '/login';
    }
    return Promise.reject(error.response?.data ?? error);
  }
);

export default api;
