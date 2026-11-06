import { useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { DashboardPage } from '@/pages/dashboard/DashboardPage'
import { UserLtvDialog } from '@/widgets/user-ltv/UserLtvDialog'
import './styles/global.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      staleTime: 5000,
    },
  },
})

export function App() {
  const [ltvDialogOpen, setLtvDialogOpen] = useState(false)

  return (
    <QueryClientProvider client={queryClient}>
      <div className="mx-auto max-w-[1080px] px-5 pb-10 pt-6">
        <header className="mb-5 flex items-end justify-between gap-2">
          <div className="flex flex-col gap-1">
            <h1 className="text-[22px] font-semibold text-ink">Аналитика подписок</h1>
            <span className="text-[13px] text-faint">users · subscriptions · payments → metrics</span>
          </div>
          <button
            type="button"
            onClick={() => setLtvDialogOpen(true)}
            className="rounded-lg border border-line bg-surface2 px-3 py-1.5 text-[13px] font-medium text-ink transition-opacity hover:opacity-80"
          >
            LTV по Telegram
          </button>
        </header>
        <DashboardPage />
      </div>
      <UserLtvDialog open={ltvDialogOpen} onClose={() => setLtvDialogOpen(false)} />
    </QueryClientProvider>
  )
}
