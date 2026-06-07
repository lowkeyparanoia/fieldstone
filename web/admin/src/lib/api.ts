import axios, { AxiosInstance, AxiosResponse } from 'axios'

const API_BASE_URL = 'http://localhost:8090/api'

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

export interface StorageBucket {
  name: string
  public: boolean
  createdAt: string
}

export interface StorageObject {
  name: string
  bucket: string
  size: number
  contentType: string
  createdAt: string
  updatedAt: string
}

export interface Webhook {
  id: string
  url: string
  secret: string
  events: string[]
  active: boolean
  retries: number
  createdAt: string
  updatedAt: string
}

export interface GraphQLResponse {
  data?: unknown
  errors?: Array<{ message: string }>
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
  
  sendOTP: (phone: string): Promise<AxiosResponse<{ message: string }>> =>
    apiClient.post('/auth/otp/send', { phone }),
  
  verifyOTP: (phone: string, token: string): Promise<AxiosResponse<{ token: string; user: User }>> =>
    apiClient.post('/auth/otp/verify', { phone, token }),
  
  magicLink: (email: string): Promise<AxiosResponse<{ message: string }>> =>
    apiClient.post('/auth/magiclink', { email }),
  
  resetPassword: (email: string): Promise<AxiosResponse<{ message: string }>> =>
    apiClient.post('/auth/recover', { email }),
}

// Collections API (legacy)
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

// Records API (legacy)
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

// Gateway API (new PostgREST-like API)
export const gatewayApi = {
  query: (table: string, params?: Record<string, string>): Promise<AxiosResponse<{ data: unknown[] }>> =>
    apiClient.get(`/v1/${table}`, { params }),
  
  getById: (table: string, id: string, params?: Record<string, string>): Promise<AxiosResponse<{ data: unknown }>> =>
    apiClient.get(`/v1/${table}/${id}`, { params }),
  
  insert: (table: string, data: Record<string, unknown>): Promise<AxiosResponse<{ data: unknown }>> =>
    apiClient.post(`/v1/${table}`, data),
  
  update: (table: string, id: string, data: Record<string, unknown>): Promise<AxiosResponse<{ data: unknown }>> =>
    apiClient.patch(`/v1/${table}`, { ...data, id }),
  
  delete: (table: string, id: string): Promise<AxiosResponse<{ success: boolean }>> =>
    apiClient.delete(`/v1/${table}`, { params: { id } }),
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

// Storage API
export const storageApi = {
  listBuckets: (): Promise<AxiosResponse<{ buckets: StorageBucket[] }>> =>
    apiClient.get('/storage/buckets'),
  
  createBucket: (name: string, isPublic: boolean = false): Promise<AxiosResponse<{ name: string }>> =>
    apiClient.post('/storage/buckets', { name, public: isPublic }),
  
  deleteBucket: (name: string): Promise<AxiosResponse<{ success: boolean }>> =>
    apiClient.delete(`/storage/buckets/${name}`),
  
  listObjects: (bucket: string, prefix?: string): Promise<AxiosResponse<{ objects: StorageObject[] }>> =>
    apiClient.get(`/storage/buckets/${bucket}/objects`, { params: { prefix } }),
  
  upload: (bucket: string, path: string, file: File): Promise<AxiosResponse<StorageObject>> => {
    const formData = new FormData()
    formData.append('file', file)
    return apiClient.post(`/storage/buckets/${bucket}/objects?path=${encodeURIComponent(path)}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  
  deleteObject: (bucket: string, path: string): Promise<AxiosResponse<{ success: boolean }>> =>
    apiClient.delete(`/storage/buckets/${bucket}/objects/${path}`),
}

// Webhooks API
export const webhooksApi = {
  list: (): Promise<AxiosResponse<{ items: Webhook[] }>> =>
    apiClient.get('/webhooks'),
  
  get: (id: string): Promise<AxiosResponse<Webhook>> =>
    apiClient.get(`/webhooks/${id}`),
  
  create: (data: Partial<Webhook>): Promise<AxiosResponse<Webhook>> =>
    apiClient.post('/webhooks', data),
  
  update: (id: string, data: Partial<Webhook>): Promise<AxiosResponse<Webhook>> =>
    apiClient.put(`/webhooks/${id}`, data),
  
  delete: (id: string): Promise<AxiosResponse<void>> =>
    apiClient.delete(`/webhooks/${id}`),
}

// GraphQL API
export const graphqlApi = {
  query: (query: string, variables?: Record<string, unknown>): Promise<AxiosResponse<GraphQLResponse>> =>
    apiClient.post('/graphql', { query, variables }),
}

// Functions API
export const functionsApi = {
  invoke: (name: string, body?: Record<string, unknown>): Promise<AxiosResponse<unknown>> =>
    apiClient.post(`/functions/v1/${name}`, body),
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

// Vector Search API
export const vectorApi = {
  search: (table: string, query: string, limit: number = 10): Promise<AxiosResponse<{ data: unknown[] }>> =>
    apiClient.post('/v1/semantic_search', { table, query, limit }),
}

export default apiClient
