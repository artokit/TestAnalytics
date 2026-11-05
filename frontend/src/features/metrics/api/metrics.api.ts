import type {
  LtvSummary,
  Plan,
  Retention,
  SyncStatus,
} from '@/entities/metrics/types'
import { http } from '@/shared/api/http'

export const metricsApi = {
  async summary(from?: string, to?: string): Promise<LtvSummary> {
    const res = await http.get<LtvSummary>('/metrics/ltv/summary', {
      params: { from: from ?? '', to: to ?? '' },
    })
    return res.data
  },
  async retention(
    planId: number,
    maxPeriods: number,
    from?: string,
    to?: string,
  ): Promise<Retention> {
    const res = await http.get<Retention>('/metrics/retention', {
      params: { plan_id: planId, max_periods: maxPeriods, from: from ?? '', to: to ?? '' },
    })
    return res.data
  },
  async status(): Promise<SyncStatus> {
    const res = await http.get<SyncStatus>('/status')
    return res.data
  },
}

export type { Plan }
