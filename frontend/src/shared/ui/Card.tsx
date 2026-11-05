import { ReactNode } from 'react'

interface CardProps {
  title?: string
  extra?: ReactNode
  children: ReactNode
}

export function Card({ title, extra, children }: CardProps) {
  return (
    <div className="flex flex-col gap-3 rounded-xl border border-line bg-surface p-4">
      {(title || extra) && (
        <div className="flex items-center justify-between gap-2">
          {title && (
            <span className="text-xs font-semibold uppercase tracking-wider text-muted">
              {title}
            </span>
          )}
          {extra}
        </div>
      )}
      <div className="min-h-0 flex-1">{children}</div>
    </div>
  )
}
