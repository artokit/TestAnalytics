import type {
  CurrencyRates,
  LtvByPlan,
  LtvSummary,
  Plan,
  Retention,
  SyncStatus,
  UserLtv,
} from '@/entities/metrics/types'
import { getWithEtag, http } from '@/shared/api/http'

export const metricsApi = {
  async summary(from?: string, to?: string): Promise<LtvSummary> {
    return getWithEtag<LtvSummary>('/metrics/ltv/summary', { from: from ?? '', to: to ?? '' })
  },
  async ltvByPlan(from?: string, to?: string): Promise<LtvByPlan> {
    return getWithEtag<LtvByPlan>('/metrics/ltv/by-plan', { from: from ?? '', to: to ?? '' })
  },
  async retention(
    planId: number,
    maxPeriods: number,
    from?: string,
    to?: string,
  ): Promise<Retention> {
    return getWithEtag<Retention>('/metrics/retention', {
      plan_id: planId,
      max_periods: maxPeriods,
      from: from ?? '',
      to: to ?? '',
    })
  },
  async status(): Promise<SyncStatus> {
    const res = await http.get<SyncStatus>('/status')
    return res.data
  },
  async userLtv(telegramId: number): Promise<UserLtv> {
    return getWithEtag<UserLtv>(`/users/by-telegram/${telegramId}/ltv`)
  },
  async currency(): Promise<CurrencyRates> {
    const res = await http.get<CurrencyRates>('/currency')
    return res.data
  },
}

export type { Plan }
