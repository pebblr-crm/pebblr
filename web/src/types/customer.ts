/**
 * Customer domain types — mirror the Go backend domain model.
 * Field names use camelCase (frontend convention), mapping to the
 * snake_case json tags on the Go domain.Customer struct.
 */

export type { CustomerType } from './lead'

import type { CustomerType } from './lead'

export interface Address {
  street: string
  city: string
  state: string
  country: string
  zip: string
}

export interface Customer {
  id: string
  name: string
  type: CustomerType
  address: Address
  phone: string
  email: string
  notes: string
  createdAt: string
  updatedAt: string
}

export interface CreateCustomerInput {
  name: string
  type: CustomerType
  address?: Address
  phone?: string
  email?: string
  notes?: string
}

export interface UpdateCustomerInput {
  id: string
  name?: string
  type?: CustomerType
  address?: Address
  phone?: string
  email?: string
  notes?: string
}

export interface CustomerListParams {
  page?: number
  limit?: number
  type?: CustomerType
}
