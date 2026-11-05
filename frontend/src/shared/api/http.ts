import axios, { AxiosError } from 'axios'
import { VITE_API_BASE_URL } from '../config/env'

export const http = axios.create({
  baseURL: VITE_API_BASE_URL,
  timeout: 15000,
})

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

type QueryParams = Record<string, string | number | undefined>

const etagCache = new Map<string, { etag: string; data: unknown }>()

export async function getWithEtag<T>(url: string, params: QueryParams = {}): Promise<T> {
  const key = `${url}?${JSON.stringify(params)}`
  const cached = etagCache.get(key)
  const headers = cached ? { 'If-None-Match': cached.etag } : undefined

  try {
    const res = await http.get<T>(url, { params, headers })
    const etag = res.headers?.etag
    if (typeof etag === 'string') {
      etagCache.set(key, { etag, data: res.data })
    } else {
      etagCache.delete(key)
    }
    return res.data
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 304 && cached) {
      return cached.data as T
    }
    throw error
  }
}

export function getErrorStatus(error: unknown): number | null {
  if (error instanceof AxiosError && error.response) {
    return error.response.status
  }
  return null
}

export function isSnapshotNotReady(error: unknown): boolean {
  return getErrorStatus(error) === 503
}

export function errorMessage(error: unknown): string {
  if (axios.isAxiosError(error) && typeof error.response?.data?.error === 'string') {
    return error.response.data.error
  }

  return 'Произошла ошибка'
}
