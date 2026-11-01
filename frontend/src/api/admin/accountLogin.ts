import apiClient from '../client'

export interface AccountLoginCredentials {
  login_username?: string
  login_password?: string
  login_totp_secret?: string
}

export async function getAccountLoginCredentials(id: string, signal?: AbortSignal) {
  const { data } = await apiClient.get<AccountLoginCredentials>(`/admin/accounts/${id}/login-credentials`, { signal })
  return data
}
