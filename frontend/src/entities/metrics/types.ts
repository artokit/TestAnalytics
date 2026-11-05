export interface LtvSummaryMetrics {
  avg_ltv: number
  median_ltv: number
  total_revenue: number
  users: number
}

export interface LtvSummary {
  data_as_of: string
  metrics: LtvSummaryMetrics
}

export interface RetentionPeriod {
  period: number
  eligible: number
  renewed: number
  rate: number
  cumulative_rate: number
}

export interface Retention {
  plan_id: number
  data_as_of: string
  cohort_size: number
  periods: RetentionPeriod[]
}

export interface SyncStatus {
  ready: boolean
  last_poll: string
  next_poll: string
  last_sync_err?: string
}

export interface Plan {
  id: number
  name: string
  period_months: number
  price: number
  currency: string
}

export interface PlansResponse {
  items: Plan[]
}
