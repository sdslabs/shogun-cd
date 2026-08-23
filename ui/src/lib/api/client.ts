import { getSessionToken } from "@/features/auth/session"
import type { ApiResponse } from "@/lib/api/types"

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || "/api").replace(/\/$/, "")

export function absoluteApiUrl(path: string) {
  const base = /^https?:\/\//.test(API_BASE_URL) ? API_BASE_URL : `${window.location.origin}${API_BASE_URL}`
  return `${base}${path.startsWith("/") ? path : `/${path}`}`
}

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly detail?: string,
  ) {
    super(message)
    this.name = "ApiError"
  }
}

interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown
  auth?: boolean
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set("Accept", "application/json")

  if (options.body !== undefined) headers.set("Content-Type", "application/json")

  if (options.auth !== false) {
    const token = getSessionToken()
    if (token) headers.set("Authorization", `Bearer ${token}`)
  }

  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers,
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
    })
  } catch (error) {
    throw new ApiError(
      "Shogun is unreachable",
      0,
      error instanceof Error ? error.message : undefined,
    )
  }

  const payload = (await response.json().catch(() => null)) as ApiResponse<T> | null

  if (!response.ok || !payload?.success) {
    if (response.status === 401 && options.auth !== false) {
      window.dispatchEvent(new CustomEvent("shogun:unauthorized"))
    }
    throw new ApiError(
      payload?.message || `Request failed with status ${response.status}`,
      response.status,
      payload?.error,
    )
  }

  return payload.data as T
}
