import { useMemo } from 'react'
import {
  Card,
  CardContent,
  Typography,
  Box,
  Grid,
  Skeleton,
  useTheme,
} from '@mui/material'
import TrendingUpIcon from '@mui/icons-material/TrendingUp'
import TrendingDownIcon from '@mui/icons-material/TrendingDown'
import { useTranslation } from 'react-i18next'
import { useHistoricalRates } from '../../api/hooks'
import { formatRate, parseCurrencyPair } from '../../utils/formatters'
import ErrorAlert from '../../components/ErrorAlert'

interface RateStatsCardProps {
  pair: string
}

function RateStatsCard({ pair }: RateStatsCardProps) {
  const { t } = useTranslation()
  const theme = useTheme()
  const { isInverted, apiPair } = parseCurrencyPair(pair)
  const { data, isLoading, error, refetch } = useHistoricalRates(apiPair, 1, 30)

  const stats = useMemo(() => {
    if (!data?.items || data.items.length === 0) {
      return { high: 0, low: 0, average: 0, change: 0, changePercent: 0 }
    }

    const rates = data.items.map(item => isInverted ? 1 / item.rate : item.rate)
    const high = Math.max(...rates)
    const low = Math.min(...rates)
    const average = rates.reduce((sum, rate) => sum + rate, 0) / rates.length

    // Calculate 24h change (compare latest with oldest in 30-day period)
    const latestRate = rates[0]
    const oldestRate = rates[rates.length - 1]
    const change = latestRate - oldestRate
    const changePercent = oldestRate !== 0 ? (change / oldestRate) * 100 : 0

    return { high, low, average, change, changePercent }
  }, [data, isInverted])

  if (error) {
    return (
      <Card>
        <CardContent>
          <ErrorAlert
            title={t('error.loadFailed')}
            message={error.message}
            onRetry={() => refetch()}
          />
        </CardContent>
      </Card>
    )
  }

  const StatItem = ({ label, value, isChange = false }: { label: string; value: string; isChange?: boolean }) => (
    <Box>
      <Typography variant="caption" color="text.secondary" display="block">
        {label}
      </Typography>
      {isLoading ? (
        <Skeleton variant="text" width={80} height={32} />
      ) : (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Typography variant="h6" fontWeight={600} color="secondary.dark">
            {value}
          </Typography>
          {isChange && stats.change !== 0 && (
            <>
              {stats.change > 0 ? (
                <TrendingUpIcon sx={{ color: '#4caf50', fontSize: 20 }} />
              ) : (
                <TrendingDownIcon sx={{ color: '#f44336', fontSize: 20 }} />
              )}
              <Typography
                variant="caption"
                sx={{
                  color: stats.change > 0 ? '#4caf50' : '#f44336',
                  fontWeight: 600
                }}
              >
                ({stats.changePercent > 0 ? '+' : ''}{stats.changePercent.toFixed(2)}%)
              </Typography>
            </>
          )}
        </Box>
      )}
    </Box>
  )

  return (
    <Card
      sx={{
        height: '100%',
        background: theme.palette.mode === 'dark'
          ? 'linear-gradient(135deg, #3a1e5f 0%, #5a2a98 100%)'
          : 'linear-gradient(135deg, #f3e5f5 0%, #e1bee7 100%)',
        border: '1px solid',
        borderColor: 'secondary.light',
      }}
    >
      <CardContent>
        <Typography variant="h6" gutterBottom sx={{ color: 'secondary.dark', mb: 3 }}>
          {t('stats.title')}
        </Typography>
        <Grid container spacing={2}>
          <Grid item xs={6}>
            <StatItem
              label={t('stats.high')}
              value={formatRate(stats.high, isInverted ? 4 : 6)}
            />
          </Grid>
          <Grid item xs={6}>
            <StatItem
              label={t('stats.low')}
              value={formatRate(stats.low, isInverted ? 4 : 6)}
            />
          </Grid>
          <Grid item xs={6}>
            <StatItem
              label={t('stats.average')}
              value={formatRate(stats.average, isInverted ? 4 : 6)}
            />
          </Grid>
          <Grid item xs={6}>
            <StatItem
              label={t('stats.change24h')}
              value={formatRate(Math.abs(stats.change), isInverted ? 4 : 6)}
              isChange
            />
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  )
}

export default RateStatsCard
