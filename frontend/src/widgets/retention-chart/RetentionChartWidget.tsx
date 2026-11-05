import { useMemo, useState } from 'react'
import { CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { Card, Loader, Select, StateMessage } from '@/shared/ui'
import { errorMessage, isSnapshotNotReady } from '@/shared/api/http'
import { useRetention } from '@/features/metrics/model/hooks'
import { formatDataAsOf, formatNumber } from '@/entities/metrics/lib/format'
import { PLANS } from '@/entities/metrics/lib/plans'

export function RetentionChartWidget() {
  const [planId, setPlanId] = useState(1)
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')

  const planOptions = useMemo(
    () =>
      PLANS.map((plan) => ({
        value: plan.id,
        label: `${plan.name} (${plan.period_months} мес.)`,
      })),
    [],
  )

  const selectedPlanId = planOptions.some((opt) => opt.value === planId)
    ? planId
    : planOptions[0]?.value ?? 0

  const retention = useRetention(selectedPlanId, 12, {
    from: from ? new Date(from).toISOString() : undefined,
    to: to ? new Date(`${to}T23:59:59Z`).toISOString() : undefined,
  })

  return (
    <Card
      title="Продления по периодам (retention)"
      extra={
        <div className="flex items-center gap-1.5">
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
          <Select value={selectedPlanId} options={planOptions} onChange={setPlanId} />
        </div>
      }
    >
      {retention.isLoading && <Loader />}
      {retention.error !== null && (
        isSnapshotNotReady(retention.error) ? (
          <StateMessage
            title="Снимок ещё не собран"
            description="Метрики появятся после первого успешного опроса"
          />
        ) : (
          <StateMessage error title="Не удалось получить данные" description={errorMessage(retention.error)} />
        )
      )}
      {!retention.isLoading && retention.data && retention.data.periods.length === 0 && (
        <StateMessage title="Нет данных по выбранному тарифу" />
      )}
      {!retention.isLoading && retention.data && retention.data.periods.length > 0 && (
        <>
          <ResponsiveContainer width="100%" height={260}>
            <LineChart data={retention.data.periods} margin={{ top: 8, right: 12, bottom: 4, left: -12 }}>
              <CartesianGrid stroke="#2a2e3f" strokeDasharray="3 3" />
              <XAxis dataKey="period" tick={{ fill: '#8b90a7', fontSize: 12 }} />
              <YAxis tick={{ fill: '#8b90a7', fontSize: 12 }} domain={[0, 1]} ticks={[0, 0.25, 0.5, 0.75, 1]} />
              <Tooltip
                contentStyle={{ background: '#1a1d29', border: '1px solid #2a2e3f' }}
                labelStyle={{ color: '#8b90a7' }}
                formatter={(value: number | string, name: string) => [Number(value).toFixed(4), name]}
              />
              <Line
                type="monotone"
                dataKey="rate"
                name="От дошедших"
                stroke="#6c7bff"
                strokeWidth={2}
                dot={{ r: 3 }}
              />
              <Line
                type="monotone"
                dataKey="cumulative_rate"
                name="От когорты"
                stroke="#4fd48e"
                strokeWidth={2}
                dot={{ r: 3 }}
              />
            </LineChart>
          </ResponsiveContainer>

          <div className="flex justify-between gap-2 text-xs text-muted">
            <span>
              Когорта: <strong className="text-ink">{formatNumber(retention.data.cohort_size)}</strong> пользователей
            </span>
            <span>
              данные на: <strong className="text-ink">{formatDataAsOf(retention.data.data_as_of)}</strong>
            </span>
          </div>
        </>
      )}
    </Card>
  )
}
