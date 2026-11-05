import { Badge, Card, Loader, StateMessage } from '@/shared/ui'
import { errorMessage } from '@/shared/api/http'
import { useStatus } from '@/features/metrics/model/hooks'
import { formatTime } from './lib/formatTime'

export function SyncStatusWidget() {
  const { data, isLoading, error, refetch } = useStatus()

  return (
    <Card title="Статус синхронизации">
      {isLoading && <Loader />}
      {!isLoading && error !== null && (
        <StateMessage error title="Нет связи с сервисом" description={errorMessage(error)} />
      )}
      {!isLoading && data && (
        <div className="flex flex-col gap-2.5 text-[13px]">
          <div className="flex justify-between items-center gap-2">
            <span className="text-muted">Снимок</span>
            {data.ready ? (
              <Badge tone="ok">готов</Badge>
            ) : (
              <Badge tone="warn">не готов</Badge>
            )}
          </div>
          <div className="flex justify-between items-center gap-2">
            <span className="text-muted">Последний опрос</span>
            <span className="text-ink">{formatTime(data.last_poll)}</span>
          </div>
          <div className="flex justify-between items-center gap-2">
            <span className="text-muted">Следующий опрос</span>
            <span className="text-ink">{formatTime(data.next_poll)}</span>
          </div>
          {data.last_sync_err ? (
            <StateMessage error title="Последняя ошибка синхронизации" description={data.last_sync_err} />
          ) : null}
          {!data.ready && !data.last_sync_err ? (
            <StateMessage
              title="Первый опрос выполняется"
              description="Данные появятся после успешной синхронизации"
            />
          ) : null}
          <button
            onClick={() => refetch()}
            className="mt-1 rounded-lg border border-line bg-surface2 px-2.5 py-1.5 text-[13px] text-muted transition-colors hover:border-accent hover:text-ink"
          >
            Обновить
          </button>
        </div>
      )}
    </Card>
  )
}
