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
  label?: string
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

export interface StatusDef {
  key: string
  label: string
  initial?: boolean
  submittable?: boolean
}

export interface ActivityTypeConfig {
  key: string
  label: string
  category: 'field' | 'non_field'
  title_field?: string
  fields: FieldConfig[]
  submit_required?: string[]
  blocks_field_activities?: boolean
}

export interface ActivitiesConfig {
  statuses: StatusDef[]
  status_transitions: Record<string, string[]>
  durations: OptionDef[]
  types: ActivityTypeConfig[]
  routing_options: OptionDef[]
  hoisted_fields?: string[] // Populated by backend: field keys that map to top-level DB columns
}

export interface RulesConfig {
  frequency: Record<string, number>
  max_activities_per_day: number
  visit_cadence_days?: number
  default_visit_duration_minutes: Record<string, number>
  visit_duration_step_minutes: number
}

export interface RecoveryConfig {
  weekend_activity_flag: boolean
  recovery_window_days: number
  recovery_type: string
}

export interface TenantConfig {
  tenant: {
    name: string
    locale: string
  }
  accounts: {
    types: AccountTypeConfig[]
  }
  activities: ActivitiesConfig
  options: Record<string, OptionDef[]>
  rules: RulesConfig
  recovery?: RecoveryConfig
}
