import { apiClient } from "@/api/client"
import type { CreateStoreRequest, Store } from "@/types"

export async function getStores(): Promise<Store[]> {
  const res = await apiClient.get<Store[]>("/stores")
  return res.data
}

export async function createStore(payload: CreateStoreRequest): Promise<Store> {
  const res = await apiClient.post<Store>("/stores", payload)
  return res.data
}

export async function deleteStore(name: string): Promise<void> {
  await apiClient.delete(`/stores/${encodeURIComponent(name)}`)
}

