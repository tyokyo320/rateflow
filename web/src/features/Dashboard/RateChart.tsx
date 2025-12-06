import { useMemo, useState } from 'react'
import {
  Box,
  Typography,
  ToggleButtonGroup,
  ToggleButton,
  Button,
  Popover,
  Stack,
} from '@mui/material'
import DateRangeIcon from '@mui/icons-material/DateRange'
import { DatePicker } from '@mui/x-date-pickers/DatePicker'
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider'
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs'
import dayjs, { Dayjs } from 'dayjs'
import 'dayjs/locale/zh-cn'
import 'dayjs/locale/en'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import { useTranslation } from 'react-i18next'
import { useHistoricalRates } from '../../api/hooks'
import { formatDate, formatRate, parseCurrencyPair } from '../../utils/formatters'
import LoadingSpinner from '../../components/LoadingSpinner'
import ErrorAlert from '../../components/ErrorAlert'
import type { ChartDataPoint } from '../../types'

interface RateChartProps {
  pair: string
  days: number
  onDaysChange: (days: number) => void
}

const dayOptions = [7, 14, 30, 60, 90]

function RateChart({ pair, days, onDaysChange }: RateChartProps) {
  const { t, i18n } = useTranslation()
  const { isInverted, apiPair } = parseCurrencyPair(pair)
  const [anchorEl, setAnchorEl] = useState<HTMLButtonElement | null>(null)
  const [startDate, setStartDate] = useState<Dayjs | null>(dayjs().subtract(days, 'day'))
  const [endDate, setEndDate] = useState<Dayjs | null>(dayjs())
  const [customDateRange, setCustomDateRange] = useState<[string, string] | null>(null)

  // Use custom date range if set, otherwise use days parameter
  const { data, isLoading, error, refetch } = useHistoricalRates(
    apiPair,
    1,
    days + 1,
    customDateRange?.[0],
    customDateRange?.[1]
  )

  const chartData = useMemo<ChartDataPoint[]>(() => {
    if (!data?.items) return []

    // For long date ranges (> 90 days), show year in date label
    const dateFormat = days > 90 ? 'YY-MM-DD' : 'MM-DD'

    // Backend returns data in descending order (newest first)
    // We need to reverse it to show oldest to newest (left to right on chart)
    return data.items
      .map((rate) => ({
        date: formatDate(rate.effectiveDate, dateFormat),
        rate: isInverted ? 1 / rate.rate : rate.rate,
        fullDate: rate.effectiveDate,
      }))
      .reverse()
  }, [data, isInverted, days])

  // Calculate Y-axis domain with appropriate scale based on rate range
  const yAxisConfig = useMemo(() => {
    if (chartData.length === 0) return { domain: [0, 100] as [number, number], decimals: 2, allowDecimals: true }

    const rates = chartData.map(d => d.rate)
    const min = Math.min(...rates)
    const max = Math.max(...rates)
    const range = max - min

    // For small values (< 1), use decimal precision
    if (max < 1) {
      // Calculate appropriate decimal places based on the value magnitude
      // Clamp decimals between 4 and 10 to avoid toFixed() errors
      const decimals = range > 0
        ? Math.min(10, Math.max(4, Math.ceil(-Math.log10(range)) + 2))
        : 6
      const padding = range * 0.15 || 0.001 // 15% padding

      return {
        domain: [
          Math.max(0, min - padding),
          max + padding
        ] as [number, number],
        decimals,
        allowDecimals: true
      }
    }

    // For medium values (1-10), use 1-2 decimal places
    if (max < 10) {
      const padding = range * 0.1 || 0.1
      return {
        domain: [
          Math.floor((min - padding) * 10) / 10,
          Math.ceil((max + padding) * 10) / 10
        ] as [number, number],
        decimals: 2,
        allowDecimals: true
      }
    }

    // For large values (>= 10), use integers
    const minFloor = Math.floor(min)
    const maxCeil = Math.ceil(max)
    return {
      domain: [minFloor - 0.5, maxCeil + 0.5] as [number, number],
      decimals: 0,
      allowDecimals: false
    }
  }, [chartData])

  // Calculate smart tick interval for X-axis based on data length
  const xAxisInterval = useMemo(() => {
    const length = chartData.length
    if (length <= 7) return 0 // Show all for 7 days
    if (length <= 14) return 1 // Show every other for 14 days
    if (length <= 30) return Math.floor(length / 7) // Show ~7 ticks for 30 days
    if (length <= 60) return Math.floor(length / 8) // Show ~8 ticks for 60 days
    if (length <= 90) return Math.floor(length / 10) // Show ~10 ticks for 90 days
    if (length <= 180) return Math.floor(length / 12) // Show ~12 ticks for 180 days
    if (length <= 365) return Math.floor(length / 12) // Show ~12 ticks for 1 year
    return Math.floor(length / 15) // Show ~15 ticks for longer ranges
  }, [chartData.length])

  // Determine dot display based on data density
  const showDots = useMemo(() => {
    return chartData.length <= 30 // Only show dots for 30 days or less
  }, [chartData.length])

  const handleDaysChange = (_: React.MouseEvent<HTMLElement>, newDays: number | null) => {
    if (newDays !== null) {
      // Clear custom date range when switching to preset days
      setCustomDateRange(null)
      onDaysChange(newDays)
    }
  }

  const handleCustomRangeClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget)
  }

  const handleCustomRangeClose = () => {
    setAnchorEl(null)
    // Reset to current days range
    setStartDate(dayjs().subtract(days, 'day'))
    setEndDate(dayjs())
  }

  const handleDateRangeApply = () => {
    if (startDate && endDate && startDate.isBefore(endDate)) {
      const numDays = endDate.diff(startDate, 'day')
      // Set custom date range for API call
      setCustomDateRange([
        startDate.format('YYYY-MM-DD'),
        endDate.format('YYYY-MM-DD')
      ])
      onDaysChange(numDays)
      setAnchorEl(null)
    }
  }

  const isCustomRange = !dayOptions.includes(days)
  const open = Boolean(anchorEl)

  if (error) {
    return (
      <ErrorAlert
        title={t('error.cannotLoadChart')}
        message={error.message}
        onRetry={() => refetch()}
      />
    )
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2, flexWrap: 'wrap', gap: 1 }}>
        <Typography variant="h6">{t('dashboard.rateTrend')}</Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <ToggleButtonGroup
            value={isCustomRange ? null : days}
            exclusive
            onChange={handleDaysChange}
            size="small"
          >
            {dayOptions.map((option) => (
              <ToggleButton key={option} value={option}>
                {t('chart.days', { count: option })}
              </ToggleButton>
            ))}
          </ToggleButtonGroup>

          <Button
            variant={isCustomRange ? 'contained' : 'outlined'}
            size="small"
            startIcon={<DateRangeIcon />}
            onClick={handleCustomRangeClick}
          >
            {isCustomRange ? `${days}${t('chart.daysUnit')}` : t('chart.custom')}
          </Button>

          <Popover
            open={open}
            anchorEl={anchorEl}
            onClose={handleCustomRangeClose}
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'right',
            }}
            transformOrigin={{
              vertical: 'top',
              horizontal: 'right',
            }}
          >
            <LocalizationProvider dateAdapter={AdapterDayjs} adapterLocale={i18n.language === 'zh' ? 'zh-cn' : 'en'}>
              <Box sx={{ p: 2, display: 'flex', flexDirection: 'column', gap: 2, minWidth: 300 }}>
                <Typography variant="subtitle2">{t('chart.customRange')}</Typography>
                <Stack spacing={2}>
                  <DatePicker
                    label={t('chart.startDate')}
                    value={startDate}
                    onChange={(newValue) => setStartDate(newValue)}
                    maxDate={endDate || dayjs()}
                    slotProps={{ textField: { size: 'small' } }}
                  />
                  <DatePicker
                    label={t('chart.endDate')}
                    value={endDate}
                    onChange={(newValue) => setEndDate(newValue)}
                    minDate={startDate || undefined}
                    maxDate={dayjs()}
                    slotProps={{ textField: { size: 'small' } }}
                  />
                </Stack>
                <Box sx={{ display: 'flex', gap: 1, justifyContent: 'flex-end' }}>
                  <Button size="small" onClick={handleCustomRangeClose}>
                    {t('chart.cancel')}
                  </Button>
                  <Button
                    size="small"
                    variant="contained"
                    onClick={handleDateRangeApply}
                    disabled={!startDate || !endDate || !startDate.isBefore(endDate)}
                  >
                    {t('chart.apply')}
                  </Button>
                </Box>
              </Box>
            </LocalizationProvider>
          </Popover>
        </Box>
      </Box>

      {isLoading ? (
        <LoadingSpinner message={t('loading.loadingChart')} />
      ) : chartData.length === 0 ? (
        <Box sx={{ textAlign: 'center', py: 4 }}>
          <Typography color="text.secondary">{t('dashboard.noData')}</Typography>
        </Box>
      ) : (
        <ResponsiveContainer width="100%" height={350}>
          <LineChart data={chartData} margin={{ top: 10, right: 20, left: 5, bottom: chartData.length > 30 ? 35 : 15 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
            <XAxis
              dataKey="date"
              tick={{ fontSize: 11 }}
              tickMargin={8}
              interval={xAxisInterval}
              angle={chartData.length > 30 ? -45 : 0}
              textAnchor={chartData.length > 30 ? 'end' : 'middle'}
              height={chartData.length > 30 ? 55 : 30}
            />
            <YAxis
              domain={yAxisConfig.domain}
              tick={{ fontSize: 11 }}
              tickFormatter={(value) => value.toFixed(yAxisConfig.decimals)}
              tickMargin={5}
              width={yAxisConfig.decimals > 2 ? 65 : 55}
              allowDecimals={yAxisConfig.allowDecimals}
            />
            <Tooltip
              formatter={(value: number) => formatRate(value, 6)}
              labelFormatter={(label, payload) => {
                if (payload && payload.length > 0) {
                  const fullDate = payload[0].payload.fullDate
                  return `${t('table.date')}: ${formatDate(fullDate, 'YYYY-MM-DD')}`
                }
                return `${t('table.date')}: ${label}`
              }}
              contentStyle={{
                backgroundColor: 'rgba(255, 255, 255, 0.95)',
                border: '1px solid #ccc',
                borderRadius: 8,
              }}
            />
            <Legend
              wrapperStyle={{ paddingTop: 10 }}
              iconType="line"
            />
            <Line
              type="monotone"
              dataKey="rate"
              stroke="#5e92f3"
              strokeWidth={2}
              dot={showDots ? { r: 3, fill: '#5e92f3' } : false}
              activeDot={{ r: 6, fill: '#1976d2' }}
              name={t('chart.rate')}
            />
          </LineChart>
        </ResponsiveContainer>
      )}
    </Box>
  )
}

export default RateChart
