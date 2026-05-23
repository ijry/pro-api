import { get } from './http'

export interface ModelInfo {
  id: string
  name: string
  provider: string
  description?: string
  context_length?: number
  input_price_per_1k?: number
  output_price_per_1k?: number
}

export interface ModelsResponse {
  models: ModelInfo[]
}

export const modelApi = {
  list: () => get<ModelsResponse>('/api/public/models'),
}
