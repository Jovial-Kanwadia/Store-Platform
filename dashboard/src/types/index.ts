export interface Store {
  name: string
  namespace: string
  // FIX: Flatten these (remove 'spec' nesting)
  engine: string
  plan: string

  status: string // This is the string "Ready" or "Provisioning"
  url?: string
  createdAt: string
}

export interface CreateStoreRequest {
  name: string
  plan: string
  engine: string
  namespace?: string
}
