import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { dashboardApi, collectionsApi, usersApi, recordsApi } from '@/lib/api'

// Dashboard Stats Hook
export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: async () => {
      const response = await dashboardApi.getStats()
      return response.data
    },
    refetchInterval: 30000, // Refetch every 30 seconds
  })
}

// Dashboard Activities Hook
export function useDashboardActivities(limit = 10) {
  return useQuery({
    queryKey: ['dashboard', 'activities', limit],
    queryFn: async () => {
      const response = await dashboardApi.getActivities(limit)
      return response.data.items
    },
  })
}

// Health Status Hook
export function useHealthStatus() {
  return useQuery({
    queryKey: ['health'],
    queryFn: async () => {
      const response = await dashboardApi.getHealth()
      return response.data
    },
    refetchInterval: 10000, // Refetch every 10 seconds
  })
}

// Collections Hooks
export function useCollections() {
  return useQuery({
    queryKey: ['collections'],
    queryFn: async () => {
      const response = await collectionsApi.list()
      return response.data.items
    },
  })
}

export function useCollection(id: string) {
  return useQuery({
    queryKey: ['collections', id],
    queryFn: async () => {
      const response = await collectionsApi.get(id)
      return response.data
    },
    enabled: !!id,
  })
}

export function useCreateCollection() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: collectionsApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collections'] })
    },
  })
}

export function useUpdateCollection(id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: Parameters<typeof collectionsApi.update>[1]) =>
      collectionsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collections'] })
      queryClient.invalidateQueries({ queryKey: ['collections', id] })
    },
  })
}

export function useDeleteCollection() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: collectionsApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collections'] })
    },
  })
}

// Records Hooks
export function useRecords(collectionId: string, options?: { page?: number; limit?: number }) {
  return useQuery({
    queryKey: ['records', collectionId, options],
    queryFn: async () => {
      const response = await recordsApi.list(collectionId, options)
      return response.data
    },
    enabled: !!collectionId,
  })
}

export function useCreateRecord(collectionId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: Record<string, unknown>) =>
      recordsApi.create(collectionId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['records', collectionId] })
    },
  })
}

// Users Hooks
export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const response = await usersApi.list()
      return response.data.items
    },
  })
}

export function useDeleteUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: usersApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })
}
