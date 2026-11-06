import { useState } from 'react'
import { Card, Loader, Select, StateMessage } from '@/shared/ui'
import { errorMessage, isSnapshotNotReady } from '@/shared/api/http'
import { useCurrencyRates, useLtvByPlan } from '@/features/metrics/model/hooks'
import { formatDataAsOf, formatMoney, formatNumber } from '@/entities/metrics/lib/format'
import { PLANS } from '@/entities/metrics/lib/plans'
import type { Currency, PlanLtv } from '@/entities/metrics/types'

const CURRENCY_OPTIONS: { value: Currency; label: string }[] = [
  { value: 'RUB', label: 'RUB' },
  { value: 'USDT', label: 'USDT' },
  { value: 'SOL', label: 'SOL' },
]

function planName(planId: number): string {
  return PLANS.find((p) => p.id === planId)?.name ?? `Тариф ${planId}`
}

function planMonths(planId: number): number {
  return PLANS.find((p) => p.id === planId)?.period_months ?? 0
}

export function LtvByPlanWidget() {
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [currency, setCurrency] = useState<Currency>('RUB')

  const { data: rates, isLoading: ratesLoading } = useCurrencyRates()

  const { data, isLoading, error, isFetching } = useLtvByPlan({
    from: from ? new Date(from).toISOString() : undefined,
    to: to ? new Date(`${to}T23:59:59Z`).toISOString() : undefined,
  })

  const plans: PlanLtv[] = data
    ? [...data.plans].sort((a, b) => planMonths(a.plan_id) - planMonths(b.plan_id))
    : []

  return (
    <Card
      title="LTV по тарифу первой подписки"
      extra={
        <div className="flex gap-1.5">
          <input
            type="date"
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            className="rounded-lg border border-line bg-surface2 px-2 py-1 text-[12px] text-ink"
          />
          <input
            type="date"
            value={to}
            onChange={(e) => setTo(e.target.value)}
            className="rounded-lg border border-line bg-surface2 px-2 py-1 text-[12px] text-ink"
          />
          <Select
            value={currency}
            options={CURRENCY_OPTIONS}
            onChange={setCurrency}
          />
        </div>
      }
    >
      {isLoading && <Loader />}
      {!isLoading && error !== null && (
        isSnapshotNotReady(error) ? (
          <StateMessage
            title="Снимок ещё не собран"
            description="Метрики появятся после первого успешного опроса источников"
          />
        ) : (
          <StateMessage error title="Не удалось получить данные" description={errorMessage(error)} />
        )
      )}
      {!isLoading && error === null && data && (
        <div
          className={`grid grid-cols-1 gap-3 transition-opacity sm:grid-cols-2 ${isFetching ? 'opacity-60' : 'opacity-100'}`}
        >
          {plans.map((plan) => (
            <div key={plan.plan_id} className="flex flex-col gap-1.5 rounded-[10px] bg-surface2 p-3.5">
              <span className="text-xs font-semibold text-ink">{planName(plan.plan_id)}</span>
              <div className="grid grid-cols-3 gap-2">
                <div className="flex flex-col gap-1">
                  <span className="text-[11px] text-muted">Средний LTV</span>
                  <span className="text-sm font-bold text-ink">
                    {formatMoney(plan.avg_ltv, currency, ratesLoading ? undefined : rates?.prices)}
                  </span>
                </div>
                <div className="flex flex-col gap-1">
                  <span className="text-[11px] text-muted">Выручка</span>
                  <span className="text-sm font-bold text-ink">
                    {formatMoney(plan.total_revenue, currency, ratesLoading ? undefined : rates?.prices)}
                  </span>
                </div>
                <div className="flex flex-col gap-1">
                  <span className="text-[11px] text-muted">Пользователей</span>
                  <span className="text-sm font-bold text-ink">{formatNumber(plan.users)}</span>
                </div>
              </div>
            </div>
          ))}
          <span className="text-xs text-faint sm:col-span-2">
            данные на: {formatDataAsOf(data.data_as_of)}
          </span>
        </div>
      )}
    </Card>
  )
}
