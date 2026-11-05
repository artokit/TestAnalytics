interface StateMessageProps {
  title: string
  description?: string
  error?: boolean
}

export function StateMessage({ title, description, error = false }: StateMessageProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-1.5 px-3 py-8 text-center text-muted">
      <span
        className={`text-[15px] font-semibold ${error ? 'text-danger' : 'text-ink'}`}
      >
        {title}
      </span>
      {description && <span className="text-[13px]">{description}</span>}
    </div>
  )
}
