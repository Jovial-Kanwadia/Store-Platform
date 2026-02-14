export interface Store {
  name: string
  namespace: string
  engine: string
  plan: string

  status: string
  url?: string
  createdAt: string
}

export interface CreateStoreRequest {
  name: string
  plan: string
  engine: string
  namespace?: string
}
