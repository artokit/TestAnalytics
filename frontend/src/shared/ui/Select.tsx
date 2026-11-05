interface SelectProps {
  value: number
  options: { value: number; label: string }[]
  onChange: (value: number) => void
}

export function Select({ value, options, onChange }: SelectProps) {
  return (
    <select
      className="rounded-lg border border-line bg-surface2 px-2.5 py-1.5 text-[13px] text-ink"
      value={value}
      onChange={(e) => onChange(Number(e.target.value))}
    >
      {options.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  )
}
