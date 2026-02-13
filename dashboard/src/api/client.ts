import axios, { AxiosError } from "axios"

function getErrorMessage(err: unknown): string {
  if (!axios.isAxiosError(err)) return "Request failed"

  const axiosErr = err as AxiosError<unknown>
  const data = axiosErr.response?.data as { error?: unknown; message?: unknown } | undefined

  if (typeof data?.error === "string" && data.error.trim()) return data.error
  if (typeof data?.message === "string" && data.message.trim()) return data.message
  if (typeof axiosErr.message === "string" && axiosErr.message.trim()) return axiosErr.message

  return "Request failed"
}

export const apiClient = axios.create({
  baseURL: `${import.meta.env.VITE_API_URL || ""}/api/v1`,
})

apiClient.interceptors.response.use(
  (res) => res,
  (err) => Promise.reject(new Error(getErrorMessage(err)))
)

