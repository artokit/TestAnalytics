import type { Plan } from '@/entities/metrics/types'

export const PLANS: Plan[] = [
  {
    id: 1,
    name: 'Месячная подписка',
    period_months: 1,
    price: 50000,
    currency: 'RUB',
  },
  {
    id: 2,
    name: 'Подписка на 3 месяца',
    period_months: 3,
    price: 135000,
    currency: 'RUB',
  },
  {
    id: 3,
    name: 'Подписка на 6 месяцев',
    period_months: 6,
    price: 240000,
    currency: 'RUB',
  },
  {
    id: 4,
    name: 'Годовая подписка',
    period_months: 12,
    price: 420000,
    currency: 'RUB',
  },
]
