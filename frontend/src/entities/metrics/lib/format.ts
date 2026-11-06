import type { Currency } from '@/entities/metrics/types'

export function formatMoney(
  minorUnits: number,
  currency: Currency = 'RUB',
  prices?: Record<string, number>,
): string {
  if (currency !== 'RUB' && prices && Number.isFinite(prices[currency]) && prices[currency] > 0) {
    const value = minorUnits / 100 / prices[currency]
    return `${new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 2 }).format(value)} ${currency}`
  }

  const rubles = minorUnits / 100
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 0,
  }).format(rubles)
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat('ru-RU').format(value)
}

export function formatDataAsOf(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleString('ru-RU')
}
