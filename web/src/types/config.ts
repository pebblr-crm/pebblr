/**
 * Tenant configuration types — mirror the Go backend config model.
 * Loaded from GET /api/v1/config and used to drive dynamic UI rendering.
 */

export interface OptionDef {
  key: string
  label: string
}

export interface FieldConfig {
  key: string
  type: 'text' | 'select' | 'multi_select' | 'relation' | 'date'
  required: boolean
  editable?: boolean
  options?: string[]
  options_ref?: string
}

export interface AccountTypeConfig {
  key: string
  label: string
  fields: FieldConfig[]
}

export interface TenantConfig {
  tenant: {
    name: string
    locale: string
  }
  accounts: {
    types: AccountTypeConfig[]
  }
  options: Record<string, OptionDef[]>
}
