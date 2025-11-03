import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import 'dayjs/locale/en'
import relativeTime from 'dayjs/plugin/relativeTime'
import i18n from '../i18n'

dayjs.extend(relativeTime)

/**
 * Format a number as a rate with specified decimal places
 */
export const formatRate = (rate: number, decimals: number = 6): string => {
  return rate.toFixed(decimals)
}

/**
 * Format a date string to a readable format
 */
export const formatDate = (dateString: string, format: string = 'YYYY-MM-DD'): string => {
  return dayjs(dateString).format(format)
}

/**
 * Format a date string to relative time (e.g., "2 hours ago")
 * Language is determined by i18n current language
 */
export const formatRelativeTime = (dateString: string): string => {
  const locale = i18n.language === 'zh' ? 'zh-cn' : 'en'
  return dayjs(dateString).locale(locale).fromNow()
}

/**
 * Format currency pair for display
 */
export const formatCurrencyPair = (pair: string): string => {
  return pair.replace('/', ' / ')
}

/**
 * Parse currency pair to get base currency, quote currency, and check if inverted
 * For example: 'CNY/JPY' -> { base: 'CNY', quote: 'JPY', isInverted: false, apiPair: 'CNY/JPY' }
 *              'JPY/CNY' -> { base: 'JPY', quote: 'CNY', isInverted: true, apiPair: 'CNY/JPY' }
 */
export const parseCurrencyPair = (pair: string): {
  base: string
  quote: string
  isInverted: boolean
  apiPair: string
} => {
  const [base, quote] = pair.split('/')

  // Define the standard API pairs (what the backend stores)
  const standardPairs = ['CNY/JPY', 'USD/JPY']

  // Check if this is a standard pair
  if (standardPairs.includes(pair)) {
    return { base, quote, isInverted: false, apiPair: pair }
  }

  // If not standard, it's inverted - find the corresponding API pair
  const apiPair = `${quote}/${base}`
  return { base, quote, isInverted: true, apiPair }
}

/**
 * Calculate percentage change
 */
export const calculateChange = (current: number, previous: number): number => {
  if (previous === 0) return 0
  return ((current - previous) / previous) * 100
}

/**
 * Format percentage with sign
 */
export const formatPercentage = (value: number, decimals: number = 2): string => {
  const formatted = Math.abs(value).toFixed(decimals)
  const sign = value >= 0 ? '+' : '-'
  return `${sign}${formatted}%`
}
