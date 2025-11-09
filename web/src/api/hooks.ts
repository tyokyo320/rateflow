import { useQuery, UseQueryResult } from '@tanstack/react-query'
import { rateApi } from './client'
import type { Rate, RateHistoryData, HealthResponse } from '../types'

/**
 * Hook to fetch the latest exchange rate
 */
export const useLatestRate = (
  pair: string
): UseQueryResult<Rate, Error> => {
  return useQuery({
    queryKey: ['latestRate', pair],
    queryFn: () => rateApi.getLatestRate(pair),
    enabled: !!pair,
    refetchInterval: 5 * 60 * 1000, // Refetch every 5 minutes
  })
}

/**
 * Hook to fetch historical exchange rates
 */
export const useHistoricalRates = (
  pair: string,
  page: number = 1,
  pageSize: number = 30,
  startDate?: string,
  endDate?: string
): UseQueryResult<RateHistoryData, Error> => {
  return useQuery({
    queryKey: ['historicalRates', pair, page, pageSize, startDate, endDate],
    queryFn: () => rateApi.getHistoricalRates(pair, page, pageSize, startDate, endDate),
    enabled: !!pair,
  })
}

/**
 * Hook for health check
 */
export const useHealthCheck = (): UseQueryResult<HealthResponse, Error> => {
  return useQuery({
    queryKey: ['health'],
    queryFn: () => rateApi.healthCheck(),
    refetchInterval: 30000, // Check every 30 seconds
  })
}
