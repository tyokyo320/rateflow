import { useState, useMemo } from 'react'
import { Grid, Paper, Typography, Box, Alert } from '@mui/material'
import { useTranslation } from 'react-i18next'
import CurrentRateCard from './CurrentRateCard'
import RateChart from './RateChart'
import RateHistoryTable from './RateHistoryTable'
import RateStatsCard from './RateStatsCard'
import CurrencyPairSelector from './CurrencyPairSelector'
import CurrencyConverter from './CurrencyConverter'
import type { Currency } from '../../types'

function Dashboard() {
  const { t } = useTranslation()
  const [baseCurrency, setBaseCurrency] = useState<Currency>('CNY')
  const [quoteCurrency, setQuoteCurrency] = useState<Currency>('JPY')
  const [days, setDays] = useState(30)

  const displayPair = `${baseCurrency}/${quoteCurrency}`
  const isSameCurrency = baseCurrency === quoteCurrency

  const handleSwap = () => {
    const temp = baseCurrency
    setBaseCurrency(quoteCurrency)
    setQuoteCurrency(temp)
  }

  const handleBaseCurrencyChange = (currency: Currency) => {
    setBaseCurrency(currency)
  }

  const handleQuoteCurrencyChange = (currency: Currency) => {
    setQuoteCurrency(currency)
  }

  return (
    <Box>
      <Box sx={{ mb: 3, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Typography variant="h4" component="h1" fontWeight={700}>
          {t('dashboard.title')}
        </Typography>
        <CurrencyPairSelector
          baseCurrency={baseCurrency}
          quoteCurrency={quoteCurrency}
          onBaseCurrencyChange={handleBaseCurrencyChange}
          onQuoteCurrencyChange={handleQuoteCurrencyChange}
          onSwap={handleSwap}
        />
      </Box>

      {isSameCurrency ? (
        <Alert severity="warning" sx={{ mb: 3 }}>
          {t('dashboard.sameCurrencyWarning', { pair: displayPair })}
        </Alert>
      ) : (
        <Grid container spacing={3}>
          {/* Top Row: Current Rate and Stats Cards */}
          <Grid item xs={12} md={6} lg={4}>
            <CurrentRateCard pair={displayPair} />
          </Grid>

          <Grid item xs={12} md={6} lg={4}>
            <RateStatsCard pair={displayPair} />
          </Grid>

          {/* Currency Converter */}
          <Grid item xs={12} md={12} lg={4}>
            <CurrencyConverter pair={displayPair} />
          </Grid>

          {/* Rate Chart - Full width on its own row */}
          <Grid item xs={12}>
            <Paper sx={{ p: 3 }}>
              <RateChart
                pair={displayPair}
                days={days}
                onDaysChange={setDays}
              />
            </Paper>
          </Grid>

          {/* Historical Data Table */}
          <Grid item xs={12}>
            <Paper sx={{ p: 3 }}>
              <RateHistoryTable pair={displayPair} days={days} />
            </Paper>
          </Grid>
        </Grid>
      )}
    </Box>
  )
}

export default Dashboard
