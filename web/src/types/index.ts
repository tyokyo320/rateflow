export interface Rate {
  id: string
  pair: string
  baseCurrency: string
  quoteCurrency: string
  rate: number
  effectiveDate: string
  source: string
  createdAt: string
  updatedAt: string
}

export interface ApiResponse<T> {
  success: boolean
  data: T
  timestamp: string
}

export interface PaginationMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export interface RateHistoryData {
  items: Rate[]
  pagination: PaginationMeta
}

export interface HealthResponse {
  status: string
}

export interface ChartDataPoint {
  date: string
  rate: number
}

export type Currency = 'CNY' | 'JPY' | 'USD'
export type CurrencyPair = 'CNY/JPY' | 'USD/JPY'

export interface DateRange {
  start: Date
  end: Date
}
