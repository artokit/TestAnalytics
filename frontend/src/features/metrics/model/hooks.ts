import { useQuery } from '@tanstack/react-query'
import { metricsApi } from '../api/metrics.api'

const POLL_INTERVAL = 10_000

export type DateRange = { from?: string; to?: string }

export function useSummary(range?: DateRange) {
  return useQuery({
    queryKey: ['metrics', 'ltv-summary', range?.from ?? '', range?.to ?? ''],
    queryFn: () => metricsApi.summary(range?.from, range?.to),
    refetchInterval: POLL_INTERVAL,
    retry: 1,
  })
}

export function useRetention(planId: number, maxPeriods: number, range?: DateRange) {
  return useQuery({
    queryKey: [
      'metrics',
      'retention',
      planId,
      maxPeriods,
      range?.from ?? '',
      range?.to ?? '',
    ],
    queryFn: () => metricsApi.retention(planId, maxPeriods, range?.from, range?.to),
    refetchInterval: POLL_INTERVAL,
    retry: 1,
    enabled: planId > 0,
  })
}

export function useStatus() {
  return useQuery({
    queryKey: ['metrics', 'status'],
    queryFn: metricsApi.status,
    refetchInterval: POLL_INTERVAL,
    retry: 1,
  })
}
