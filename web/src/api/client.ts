import axios from 'axios'
import type { ApiResponse, Rate, RateHistoryData, HealthResponse } from '../types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
apiClient.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
apiClient.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response) {
      // Server responded with error
      console.error('API Error:', error.response.data)
    } else if (error.request) {
      // Request made but no response
      console.error('Network Error:', error.message)
    } else {
      console.error('Error:', error.message)
    }
    return Promise.reject(error)
  }
)

export const rateApi = {
  /**
   * Get the latest exchange rate for a currency pair
   */
  getLatestRate: async (pair: string): Promise<Rate> => {
    const response = await apiClient.get<ApiResponse<Rate>>(
      `/api/v1/rates/latest`,
      {
        params: { pair },
      }
    )
    return response.data.data
  },

  /**
   * Get historical exchange rates
   */
  getHistoricalRates: async (
    pair: string,
    page: number = 1,
    pageSize: number = 30,
    startDate?: string,
    endDate?: string
  ): Promise<RateHistoryData> => {
    const params: any = { pair, page, pageSize }
    if (startDate) params.startDate = startDate
    if (endDate) params.endDate = endDate

    const response = await apiClient.get<any>(
      `/api/v1/rates/list`,
      { params }
    )
    return {
      items: response.data.data || [],
      pagination: response.data.meta || {
        page,
        pageSize,
        total: 0,
        totalPages: 0,
      },
    }
  },

  /**
   * Health check endpoint
   */
  healthCheck: async (): Promise<HealthResponse> => {
    const response = await apiClient.get<HealthResponse>('/health')
    return response.data
  },
}

export default apiClient
