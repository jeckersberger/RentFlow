export type AIProvider = 'Claude' | 'GPT-4o' | 'Gemini' | 'Mistral' | 'Ollama'
export type MessageRole = 'user' | 'assistant'

export interface ChatMessage {
  id: string
  role: MessageRole
  content: string
  timestamp: string
}

export interface ChatThread {
  id: string
  title: string
  messages: ChatMessage[]
  created_at: string
  updated_at: string
}

export interface SmartAssetCreatorRequest {
  category: string
  specifications: Record<string, unknown>
  quantity: number
}

export interface PriceOptimizerResult {
  current_price: number
  suggested_price: number
  market_avg: number
  confidence: number
}

export interface DemandForecast {
  date: string
  demand_level: 'low' | 'medium' | 'high'
  confidence: number
  factors: string[]
}
