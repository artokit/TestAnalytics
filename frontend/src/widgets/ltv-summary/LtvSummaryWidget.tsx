import { useState } from 'react'
import { Card, Loader, StateMessage } from '@/shared/ui'
import { errorMessage, isSnapshotNotReady } from '@/shared/api/http'
import { useSummary } from '@/features/metrics/model/hooks'
import { formatDataAsOf, formatMoney, formatNumber } from '@/entities/metrics/lib/format'

export function LtvSummaryWidget() {
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')

  const { data, isLoading, error, isFetching } = useSummary({
    from: from ? new Date(from).toISOString() : undefined,
    to: to ? new Date(`${to}T23:59:59Z`).toISOString() : undefined,
  })

  return (
    <Card
      title="LTV — общий срез"
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
        </div>
      }
    >
      {isLoading && <Loader />}
      {isFetching && !isLoading && data && null}
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
          className={`grid grid-cols-2 gap-3 transition-opacity ${isFetching ? 'opacity-60' : 'opacity-100'}`}
        >
          <div className="flex flex-col gap-1.5 rounded-[10px] bg-surface2 p-3.5">
            <span className="text-xs text-muted">Средний LTV</span>
            <span className="text-xl font-bold text-ink">{formatMoney(data.metrics.avg_ltv)}</span>
          </div>
          <div className="flex flex-col gap-1.5 rounded-[10px] bg-surface2 p-3.5">
            <span className="text-xs text-muted">Медианный LTV</span>
            <span className="text-xl font-bold text-ink">{formatMoney(data.metrics.median_ltv)}</span>
          </div>
          <div className="flex flex-col gap-1.5 rounded-[10px] bg-surface2 p-3.5">
            <span className="text-xs text-muted">Общая выручка</span>
            <span className="text-xl font-bold text-ink">{formatMoney(data.metrics.total_revenue)}</span>
          </div>
          <div className="flex flex-col gap-1.5 rounded-[10px] bg-surface2 p-3.5">
            <span className="text-xs text-muted">Пользователей</span>
            <span className="text-xl font-bold text-ink">{formatNumber(data.metrics.users)}</span>
          </div>
          <span className="col-span-2 text-xs text-faint">
            данные на: {formatDataAsOf(data.data_as_of)}
          </span>
        </div>
      )}
    </Card>
  )
}
