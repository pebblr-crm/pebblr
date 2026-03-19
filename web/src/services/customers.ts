import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryResult,
  type UseMutationResult,
} from '@tanstack/react-query'
import { api } from './api'
import type {
  Customer,
  CreateCustomerInput,
  UpdateCustomerInput,
  CustomerListParams,
} from '@/types/customer'
import type { PaginatedResponse } from '@/types/api'

// ── Query keys ────────────────────────────────────────────────────────────────

export const customerKeys = {
  all: ['customers'] as const,
  lists: () => [...customerKeys.all, 'list'] as const,
  list: (params: CustomerListParams) => [...customerKeys.lists(), params] as const,
  details: () => [...customerKeys.all, 'detail'] as const,
  detail: (id: string) => [...customerKeys.details(), id] as const,
}

// ── API functions ─────────────────────────────────────────────────────────────

function buildCustomerPath(params: CustomerListParams): string {
  const qs = new URLSearchParams()
  if (params.page !== undefined) qs.set('page', String(params.page))
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.type) qs.set('type', params.type)
  const query = qs.toString()
  return query ? `/customers?${query}` : '/customers'
}

export function fetchCustomers(
  params: CustomerListParams = {},
): Promise<PaginatedResponse<Customer>> {
  return api.get<PaginatedResponse<Customer>>(buildCustomerPath(params))
}

export function fetchCustomer(id: string): Promise<Customer> {
  return api.get<Customer>(`/customers/${id}`)
}

export function createCustomer(input: CreateCustomerInput): Promise<Customer> {
  return api.post<Customer>('/customers', input)
}

export function updateCustomer({ id, ...input }: UpdateCustomerInput): Promise<Customer> {
  return api.patch<Customer>(`/customers/${id}`, input)
}

export function deleteCustomer(id: string): Promise<void> {
  return api.delete(`/customers/${id}`)
}

// ── TanStack Query hooks ──────────────────────────────────────────────────────

export function useCustomers(
  params: CustomerListParams = {},
): UseQueryResult<PaginatedResponse<Customer>> {
  return useQuery({
    queryKey: customerKeys.list(params),
    queryFn: () => fetchCustomers(params),
  })
}

export function useCustomer(id: string): UseQueryResult<Customer> {
  return useQuery({
    queryKey: customerKeys.detail(id),
    queryFn: () => fetchCustomer(id),
    enabled: Boolean(id),
  })
}

export function useCreateCustomer(): UseMutationResult<Customer, Error, CreateCustomerInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createCustomer,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: customerKeys.lists() })
    },
  })
}

export function useUpdateCustomer(): UseMutationResult<Customer, Error, UpdateCustomerInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateCustomer,
    onSuccess: (updated) => {
      queryClient.setQueryData(customerKeys.detail(updated.id), updated)
      void queryClient.invalidateQueries({ queryKey: customerKeys.lists() })
    },
  })
}

export function useDeleteCustomer(): UseMutationResult<void, Error, string> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteCustomer,
    onSuccess: (_data, id) => {
      queryClient.removeQueries({ queryKey: customerKeys.detail(id) })
      void queryClient.invalidateQueries({ queryKey: customerKeys.lists() })
    },
  })
}
