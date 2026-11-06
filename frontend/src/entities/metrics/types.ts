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

export interface PlanLtv {
  plan_id: number
  users: number
  avg_ltv: number
  total_revenue: number
}

export interface LtvByPlan {
  data_as_of: string
  plans: PlanLtv[]
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

export type Currency = 'RUB' | 'USDT' | 'SOL'

export interface CurrencyRates {
  base: string
  prices: Record<string, number>
}

export type SubscriptionStatus = 'active' | 'expired' | 'canceled' | 'none'

export interface UserLtv {
  user_id: string
  telegram_id: number
  first_plan_id: number
  ltv: number
  payments_count: number
  current_status: string
}
