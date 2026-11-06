import { useState } from 'react'
import { Badge, Loader, StateMessage } from '@/shared/ui'
import { errorMessage, isSnapshotNotReady } from '@/shared/api/http'
import { useUserLtv } from '@/features/metrics/model/hooks'
import { formatMoney, formatNumber } from '@/entities/metrics/lib/format'
import type { SubscriptionStatus } from '@/entities/metrics/types'

interface UserLtvDialogProps {
  open: boolean
  onClose: () => void
}

function parseTelegramId(raw: string): number {
  const trimmed = raw.trim()
  if (!trimmed || !/^\d+$/.test(trimmed)) {
    return -1
  }
  const value = Number(trimmed)
  if (!Number.isSafeInteger(value) || value <= 0) {
    return -1
  }
  return value
}

function statusTone(status: string): { tone: 'ok' | 'warn' | 'err'; label: string } {
  const normalized = status as SubscriptionStatus
  switch (normalized) {
    case 'active':
      return { tone: 'ok', label: 'Активна' }
    case 'expired':
      return { tone: 'warn', label: 'Истекла' }
    case 'canceled':
      return { tone: 'err', label: 'Отменена' }
    default:
      return { tone: 'warn', label: '—' }
  }
}

export function UserLtvDialog({ open, onClose }: UserLtvDialogProps) {
  const [raw, setRaw] = useState('')
  const [telegramId, setTelegramId] = useState(0)
  const [invalid, setInvalid] = useState(false)

  const { data, isLoading, error } = useUserLtv(telegramId)

  if (!open) {
    return null
  }

  const submit = () => {
    const parsed = parseTelegramId(raw)
    if (parsed === -1) {
      setInvalid(true)
      setTelegramId(0)
      return
    }
    setInvalid(false)
    setTelegramId(parsed)
  }

  const badge = data ? statusTone(data.current_status) : null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-ink/40 p-4"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[440px] flex-col gap-3 rounded-xl border border-line bg-surface p-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold uppercase tracking-wider text-muted">
            LTV по Telegram
          </span>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg px-2 py-0.5 text-[15px] leading-none text-muted transition-colors hover:text-ink"
          >
            ✕
          </button>
        </div>

        <form
          className="flex gap-2"
          onSubmit={(e) => {
            e.preventDefault()
            submit()
          }}
        >
          <input
            type="text"
            inputMode="numeric"
            placeholder="telegram_id"
            value={raw}
            onChange={(e) => setRaw(e.target.value)}
            className="flex-1 rounded-lg border border-line bg-surface2 px-2.5 py-1.5 text-[13px] text-ink outline-none focus:border-muted"
          />
          <button
            type="submit"
            className="rounded-lg bg-ink px-3 py-1.5 text-[13px] font-medium text-surface transition-opacity hover:opacity-85"
          >
            Найти
          </button>
        </form>

        {invalid && (
          <StateMessage
            error
            title="Неверный telegram_id"
            description="Введите целое положительное число"
          />
        )}

        {telegramId > 0 && isLoading && <Loader />}

        {telegramId > 0 && !isLoading && error !== null && (
          isSnapshotNotReady(error) ? (
            <StateMessage
              title="Снимок ещё не собран"
              description="Метрики появятся после первого успешного опроса источников"
            />
          ) : (
            <StateMessage
              error
              title="Пользователь не найден"
              description={errorMessage(error)}
            />
          )
        )}

        {telegramId > 0 && !isLoading && error === null && data && (
          <div className="grid grid-cols-2 gap-2.5">
            <div className="flex flex-col gap-1 rounded-[10px] bg-surface2 p-3">
              <span className="text-[11px] text-muted">LTV</span>
              <span className="text-lg font-bold text-ink">{formatMoney(data.ltv)}</span>
            </div>
            <div className="flex flex-col gap-1 rounded-[10px] bg-surface2 p-3">
              <span className="text-[11px] text-muted">Платежей</span>
              <span className="text-lg font-bold text-ink">{formatNumber(data.payments_count)}</span>
            </div>
            <div className="flex flex-col gap-1 rounded-[10px] bg-surface2 p-3">
              <span className="text-[11px] text-muted">Тариф первой подписки</span>
              <span className="text-sm font-semibold text-ink">
                {data.first_plan_id > 0 ? `план №${data.first_plan_id}` : '—'}
              </span>
            </div>
            <div className="flex flex-col gap-1 rounded-[10px] bg-surface2 p-3">
              <span className="text-[11px] text-muted">Статус подписки</span>
              <span>
                <Badge tone={badge?.tone ?? 'warn'}>{badge?.label ?? '—'}</Badge>
              </span>
            </div>
            <div className="col-span-2 flex flex-col gap-0.5 rounded-[10px] bg-surface2 p-3">
              <span className="text-[11px] text-muted">user_id</span>
              <span className="font-mono text-[12px] text-ink">{data.user_id}</span>
            </div>
            <span className="col-span-2 text-xs text-faint">
              telegram_id: {formatNumber(data.telegram_id)}
            </span>
          </div>
        )}
      </div>
    </div>
  )
}
