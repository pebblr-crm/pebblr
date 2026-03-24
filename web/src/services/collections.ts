import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryResult,
  type UseMutationResult,
} from '@tanstack/react-query'
import { api } from './api'
import type {
  Collection,
  CreateCollectionInput,
  UpdateCollectionInput,
} from '@/types/collection'

// ── Query keys ────────────────────────────────────────────────────────────────

export const collectionKeys = {
  all: ['collections'] as const,
  lists: () => [...collectionKeys.all, 'list'] as const,
  details: () => [...collectionKeys.all, 'detail'] as const,
  detail: (id: string) => [...collectionKeys.details(), id] as const,
}

// ── API functions ─────────────────────────────────────────────────────────────

export function fetchCollections(): Promise<Collection[]> {
  return api.get<{ items: Collection[] }>('/collections').then((r) => r.items)
}

export function fetchCollection(id: string): Promise<Collection> {
  return api.get<Collection>(`/collections/${id}`)
}

export function createCollection(input: CreateCollectionInput): Promise<Collection> {
  return api.post<Collection>('/collections', input)
}

export function updateCollection({ id, ...input }: UpdateCollectionInput): Promise<Collection> {
  return api.put<Collection>(`/collections/${id}`, input)
}

export function deleteCollection(id: string): Promise<void> {
  return api.delete<void>(`/collections/${id}`)
}

// ── TanStack Query hooks ──────────────────────────────────────────────────────

export function useCollections(): UseQueryResult<Collection[]> {
  return useQuery({
    queryKey: collectionKeys.lists(),
    queryFn: fetchCollections,
  })
}

export function useCollection(id: string): UseQueryResult<Collection> {
  return useQuery({
    queryKey: collectionKeys.detail(id),
    queryFn: () => fetchCollection(id),
    enabled: Boolean(id),
  })
}

export function useCreateCollection(): UseMutationResult<Collection, Error, CreateCollectionInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createCollection,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: collectionKeys.lists() }).catch(() => {})
    },
  })
}

export function useUpdateCollection(): UseMutationResult<Collection, Error, UpdateCollectionInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateCollection,
    onSuccess: (updated) => {
      queryClient.setQueryData(collectionKeys.detail(updated.id), updated)
      queryClient.invalidateQueries({ queryKey: collectionKeys.lists() }).catch(() => {})
    },
  })
}

export function useDeleteCollection(): UseMutationResult<void, Error, string> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteCollection,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: collectionKeys.lists() }).catch(() => {})
    },
  })
}
