const baseUrl = ''

export async function fetchJson<T>(
  path: string,
  options: RequestInit,
  onUnauthorized?: () => void,
): Promise<T> {
  const res = await fetch(`${baseUrl}${path}`, options)
  if (res.status === 401) {
    if (onUnauthorized) {
      onUnauthorized()
    }
    throw new Error('Unauthorized')
  }
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    const message =
      typeof (data as { error?: string })?.error === 'string'
        ? (data as { error?: string }).error
        : res.statusText
    throw new Error(message || 'Request failed')
  }
  return data as T
}

export function apiPost(path: string, token: string | null, onUnauthorized?: () => void): Promise<any> {
  return fetchJson(
    path,
    {
      method: 'POST',
      headers: token ? { 'X-GMM-Token': token } : undefined,
    },
    onUnauthorized,
  )
}

export function buildJsonHeaders(token: string | null): HeadersInit {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) {
    headers['X-GMM-Token'] = token
  }
  return headers
}
