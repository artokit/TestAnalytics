interface SelectProps<T extends string | number> {
  value: T
  options: { value: T; label: string }[]
  onChange: (value: T) => void
}

export function Select<T extends string | number>({ value, options, onChange }: SelectProps<T>) {
  return (
    <select
      className="rounded-lg border border-line bg-surface2 px-2.5 py-1.5 text-[13px] text-ink"
      value={value}
      onChange={(e) => onChange((typeof value === 'number' ? Number(e.target.value) : e.target.value) as T)}
    >
      {options.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  )
}
