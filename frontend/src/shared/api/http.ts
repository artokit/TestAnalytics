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
