import { useMemo } from 'react'
import {
  Card,
  CardContent,
  Typography,
  Box,
  Chip,
  Skeleton,
  alpha,
  useTheme,
} from '@mui/material'
import ShowChartIcon from '@mui/icons-material/ShowChart'
import TrendingUpIcon from '@mui/icons-material/TrendingUp'
import TrendingDownIcon from '@mui/icons-material/TrendingDown'
import TrendingFlatIcon from '@mui/icons-material/TrendingFlat'
import { LineChart, Line, ResponsiveContainer } from 'recharts'
import { useTranslation } from 'react-i18next'
import { useLatestRate, useHistoricalRates } from '../../api/hooks'
import { formatRate, formatRelativeTime, formatCurrencyPair, parseCurrencyPair } from '../../utils/formatters'
import ErrorAlert from '../../components/ErrorAlert'

interface CurrentRateCardProps {
  pair: string
}

function CurrentRateCard({ pair }: CurrentRateCardProps) {
  const { t } = useTranslation()
  const theme = useTheme()
  const { isInverted, apiPair } = parseCurrencyPair(pair)
  const { data: rate, isLoading, error, refetch } = useLatestRate(apiPair)
  const { data: historyData } = useHistoricalRates(apiPair, 1, 8) // Last 7 days for sparkline

  // Calculate display rate based on whether the pair is inverted
  const displayRate = rate ? (isInverted ? 1 / rate.rate : rate.rate) : 0

  // Prepare sparkline data
  const sparklineData = useMemo(() => {
    if (!historyData?.items) return []
    return historyData.items
      .map((item) => ({
        rate: isInverted ? 1 / item.rate : item.rate,
      }))
      .reverse()
      .slice(0, 7)
  }, [historyData, isInverted])

  // Calculate 24h change
  const rateChange = useMemo(() => {
    if (!sparklineData || sparklineData.length < 2) return null
    const oldRate = sparklineData[0].rate
    const newRate = sparklineData[sparklineData.length - 1].rate
    const change = newRate - oldRate
    const changePercent = (change / oldRate) * 100
    return { change, changePercent, trend: change > 0 ? 'up' : change < 0 ? 'down' : 'flat' }
  }, [sparklineData])

  if (error) {
    return (
      <Card>
        <CardContent>
          <ErrorAlert
            title={t('error.cannotGetRate')}
            message={error.message}
            onRetry={() => refetch()}
          />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card
      sx={{
        height: '100%',
        background: theme.palette.mode === 'dark'
          ? 'linear-gradient(135deg, #1e3a5f 0%, #2a5298 100%)'
          : 'linear-gradient(135deg, #e3f2fd 0%, #bbdefb 100%)',
        position: 'relative',
        overflow: 'hidden',
        border: '1px solid',
        borderColor: 'primary.light',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: -50,
          right: -50,
          width: 150,
          height: 150,
          borderRadius: '50%',
          background: alpha(theme.palette.primary.main, 0.08),
        }
      }}
    >
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
          <ShowChartIcon sx={{ color: 'primary.main' }} />
          <Typography variant="h6" color="primary.dark">
            {t('dashboard.currentRate')}
          </Typography>
        </Box>

        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', mb: 2 }}>
          <Box sx={{ flex: 1 }}>
            {isLoading ? (
              <Skeleton variant="text" width="60%" height={48} />
            ) : (
              <>
                <Typography variant="h3" component="div" fontWeight={700} color="primary.dark">
                  {formatRate(displayRate, isInverted ? 4 : 6)}
                </Typography>
                {rateChange && (
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mt: 0.5 }}>
                    {rateChange.trend === 'up' && <TrendingUpIcon sx={{ color: 'success.main', fontSize: 18 }} />}
                    {rateChange.trend === 'down' && <TrendingDownIcon sx={{ color: 'error.main', fontSize: 18 }} />}
                    {rateChange.trend === 'flat' && <TrendingFlatIcon sx={{ color: 'text.secondary', fontSize: 18 }} />}
                    <Typography
                      variant="body2"
                      sx={{
                        color: rateChange.trend === 'up' ? 'success.main' : rateChange.trend === 'down' ? 'error.main' : 'text.secondary',
                        fontWeight: 600
                      }}
                    >
                      {rateChange.changePercent >= 0 ? '+' : ''}
                      {rateChange.changePercent.toFixed(2)}% (7d)
                    </Typography>
                  </Box>
                )}
              </>
            )}
          </Box>

          {/* Mini Sparkline */}
          {!isLoading && sparklineData.length > 0 && (
            <Box sx={{ width: 100, height: 50, ml: 2 }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={sparklineData}>
                  <Line
                    type="monotone"
                    dataKey="rate"
                    stroke={rateChange?.trend === 'up' ? '#4caf50' : rateChange?.trend === 'down' ? '#f44336' : '#9e9e9e'}
                    strokeWidth={2}
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            </Box>
          )}
        </Box>

        <Box sx={{ mb: 2 }}>
          {isLoading ? (
            <Skeleton variant="rectangular" width={100} height={24} />
          ) : (
            <Chip
              label={formatCurrencyPair(pair)}
              color="primary"
              variant="outlined"
              size="small"
            />
          )}
        </Box>

        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
          {isLoading ? (
            <>
              <Skeleton variant="text" width="80%" />
              <Skeleton variant="text" width="60%" />
            </>
          ) : (
            <>
              <Typography variant="body2" color="text.secondary">
                {t('dashboard.dataSource')}: {rate?.source}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {t('dashboard.updatedAt')}: {rate?.updatedAt ? formatRelativeTime(rate.updatedAt) : '-'}
              </Typography>
              {isInverted && (
                <Typography variant="caption" color="primary.main" sx={{ fontStyle: 'italic' }}>
                  {t('dashboard.invertRate')}: 1 / {formatRate(rate?.rate || 0, 6)}
                </Typography>
              )}
            </>
          )}
        </Box>
      </CardContent>
    </Card>
  )
}

export default CurrentRateCard
