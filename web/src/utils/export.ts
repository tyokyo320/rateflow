import type { Rate } from '../types'

/**
 * Export rate data to CSV format
 */
export const exportToCSV = (data: Rate[], filename: string = 'exchange-rates.csv') => {
  if (data.length === 0) {
    console.warn('No data to export')
    return
  }

  // CSV headers
  const headers = ['Date', 'Pair', 'Rate', 'Source', 'Updated At']

  // CSV rows
  const rows = data.map(rate => [
    rate.effectiveDate,
    `${rate.baseCurrency}/${rate.quoteCurrency}`,
    rate.rate.toString(),
    rate.source,
    rate.updatedAt || '',
  ])

  // Combine headers and rows
  const csvContent = [
    headers.join(','),
    ...rows.map(row => row.join(','))
  ].join('\n')

  // Create blob and download
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)

  link.setAttribute('href', url)
  link.setAttribute('download', filename)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

/**
 * Export rate data to JSON format
 */
export const exportToJSON = (data: Rate[], filename: string = 'exchange-rates.json') => {
  if (data.length === 0) {
    console.warn('No data to export')
    return
  }

  const jsonContent = JSON.stringify(data, null, 2)

  const blob = new Blob([jsonContent], { type: 'application/json' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)

  link.setAttribute('href', url)
  link.setAttribute('download', filename)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
