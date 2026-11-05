import { LtvSummaryWidget } from '@/widgets/ltv-summary/LtvSummaryWidget'
import { RetentionChartWidget } from '@/widgets/retention-chart/RetentionChartWidget'
import { SyncStatusWidget } from '@/widgets/sync-status/SyncStatusWidget'

export function DashboardPage() {
  return (
    <main className="grid grid-cols-1 items-start gap-4 md:grid-cols-2">
      <SyncStatusWidget />
      <LtvSummaryWidget />
      <div className="md:col-span-2">
        <RetentionChartWidget />
      </div>
    </main>
  )
}
