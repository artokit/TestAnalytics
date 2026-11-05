import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { DashboardPage } from '@/pages/dashboard/DashboardPage'
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
  return (
    <QueryClientProvider client={queryClient}>
      <div className="mx-auto max-w-[1080px] px-5 pb-10 pt-6">
        <header className="mb-5 flex flex-col gap-1">
          <h1 className="text-[22px] font-semibold text-ink">Аналитика подписок</h1>
          <span className="text-[13px] text-faint">users · subscriptions · payments → metrics</span>
        </header>
        <DashboardPage />
      </div>
    </QueryClientProvider>
  )
}
