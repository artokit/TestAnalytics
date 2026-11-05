interface BadgeProps {
  tone: 'ok' | 'warn' | 'err'
  children: string
}

const toneClasses: Record<BadgeProps['tone'], string> = {
  ok: 'bg-ok/15 text-ok',
  warn: 'bg-warn/15 text-warn',
  err: 'bg-danger/15 text-danger',
}

export function Badge({ tone, children }: BadgeProps) {
  return (
    <span className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-semibold ${toneClasses[tone]}`}>
      {children}
    </span>
  )
}
