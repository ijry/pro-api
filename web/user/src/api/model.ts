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

type RawModelsResponse = ModelsResponse | ModelInfo[]

function normalizeModels(resp: RawModelsResponse): ModelsResponse {
  if (Array.isArray(resp)) return { models: resp }
  return { models: Array.isArray(resp.models) ? resp.models : [] }
}

export const modelApi = {
  list: async () => normalizeModels(await get<RawModelsResponse>('/api/public/models')),
}
