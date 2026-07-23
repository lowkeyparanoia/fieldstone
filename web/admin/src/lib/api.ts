import axios, { AxiosInstance, AxiosResponse } from 'axios'

// Relative by default, so the UI talks to whatever origin served it. A
// hardcoded host and port breaks on any other port, breaks in production, and
// makes same-origin requests cross-origin, which drags CORS in for no reason.
// VITE_API_BASE_URL is there for running `vite dev` against a remote server.
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

// Create axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000,
})

// Request interceptor to add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('fieldstone_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('fieldstone_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API Types
export interface Collection {
  id: string
  name: string
  schema: SchemaField[]
  createdAt: string
  updatedAt: string
  recordCount: number
}

export interface SchemaField {
  name: string
  type: 'text' | 'number' | 'boolean' | 'date' | 'email' | 'url' | 'json' | 'relation'
  required: boolean
  unique: boolean
  options?: Record<string, unknown>
}

export interface CollectionRecord {
  id: string
  collectionId: string
  data: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface User {
  id: string
  email: string
  verified: boolean
  createdAt: string
  updatedAt: string
  lastLoginAt?: string
}

export interface Tenant {
  id: string
  name: string
  domain?: string
  createdAt: string
  settings: Record<string, unknown>
}

export interface DashboardStats {
  totalRecords: number
  totalUsers: number
  totalCollections: number
  requestsPerMinute: number
  storageUsed: number
  storageLimit: number
  activeUsers: number
}

export interface Activity {
  id: string
  type: 'user_created' | 'collection_modified' | 'record_created' | 'record_updated' | 'record_deleted' | 'backup_completed' | 'webhook_triggered'
  message: string
  userId?: string
  details?: Record<string, unknown>
  createdAt: string
}

export interface HealthStatus {
  status: string
  timestamp: string
  services: {
    database: 'healthy' | 'unhealthy'
    api: 'healthy' | 'unhealthy'
    cache: 'healthy' | 'unhealthy'
  }
}

// Auth API
export const authApi = {
  login: (email: string, password: string): Promise<AxiosResponse<{ token: string; user: User }>> =>
    apiClient.post('/auth/login', { email, password }),
  
  register: (email: string, password: string): Promise<AxiosResponse<{ token: string; user: User }>> =>
    apiClient.post('/auth/register', { email, password }),
  
  refresh: (): Promise<AxiosResponse<{ token: string }>> =>
    apiClient.post('/auth/refresh'),
  
  logout: (): Promise<AxiosResponse<void>> =>
    apiClient.post('/auth/logout'),
}

// Collections API
export const collectionsApi = {
  list: (): Promise<AxiosResponse<{ items: Collection[] }>> =>
    apiClient.get('/collections'),
  
  get: (id: string): Promise<AxiosResponse<Collection>> =>
    apiClient.get(`/collections/${id}`),
  
  create: (data: Partial<Collection>): Promise<AxiosResponse<Collection>> =>
    apiClient.post('/collections', data),
  
  update: (id: string, data: Partial<Collection>): Promise<AxiosResponse<Collection>> =>
    apiClient.put(`/collections/${id}`, data),
  
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete(`/collections/${id}`),
}

// Records API
export const recordsApi = {
  list: (collectionId: string, params?: { page?: number; limit?: number; filter?: string; sort?: string }): Promise<AxiosResponse<{ items: CollectionRecord[]; total: number }>> =>
    apiClient.get(`/collections/${collectionId}/records`, { params }),
  
  get: (collectionId: string, recordId: string): Promise<AxiosResponse<CollectionRecord>> =>
    apiClient.get(`/collections/${collectionId}/records/${recordId}`),
  
  create: (collectionId: string, data: Record<string, unknown>): Promise<AxiosResponse<CollectionRecord>> =>
    apiClient.post(`/collections/${collectionId}/records`, data),
  
  update: (collectionId: string, recordId: string, data: Record<string, unknown>): Promise<AxiosResponse<CollectionRecord>> =>
    apiClient.put(`/collections/${collectionId}/records/${recordId}`, data),
  
  delete: (collectionId: string, recordId: string): Promise<AxiosResponse<void>> =>
    apiClient.delete(`/collections/${collectionId}/records/${recordId}`),
}

// Users API
export const usersApi = {
  list: (): Promise<AxiosResponse<{ items: User[] }>> =>
    apiClient.get('/users'),
  
  get: (id: string): Promise<AxiosResponse<User>> =>
    apiClient.get(`/users/${id}`),
  
  update: (id: string, data: Partial<User>): Promise<AxiosResponse<User>> =>
    apiClient.put(`/users/${id}`, data),
  
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete(`/users/${id}`),
}

// Tenants API
export const tenantsApi = {
  list: (): Promise<AxiosResponse<{ items: Tenant[] }>> =>
    apiClient.get('/tenants'),
  
  get: (id: string): Promise<AxiosResponse<Tenant>> =>
    apiClient.get(`/tenants/${id}`),
  
  create: (data: Partial<Tenant>): Promise<AxiosResponse<Tenant>> =>
    apiClient.post('/tenants', data),
  
  update: (id: string, data: Partial<Tenant>): Promise<AxiosResponse<Tenant>> =>
    apiClient.put(`/tenants/${id}`, data),
  
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete(`/tenants/${id}`),
}

// Dashboard API
export const dashboardApi = {
  getStats: (): Promise<AxiosResponse<DashboardStats>> =>
    apiClient.get('/dashboard/stats'),
  
  getActivities: (limit?: number): Promise<AxiosResponse<{ items: Activity[] }>> =>
    apiClient.get('/dashboard/activities', { params: { limit } }),
  
  getHealth: (): Promise<AxiosResponse<HealthStatus>> =>
    apiClient.get('/health'),
}

export default apiClient
