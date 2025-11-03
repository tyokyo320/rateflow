import { useState, useEffect } from 'react'
import {
  Card,
  CardContent,
  Typography,
  Box,
  TextField,
  IconButton,
  InputAdornment,
} from '@mui/material'
import SwapVertIcon from '@mui/icons-material/SwapVert'
import CalculateIcon from '@mui/icons-material/Calculate'
import { useTranslation } from 'react-i18next'
import { useLatestRate } from '../../api/hooks'
import { parseCurrencyPair } from '../../utils/formatters'

interface CurrencyConverterProps {
  pair: string
}

function CurrencyConverter({ pair }: CurrencyConverterProps) {
  const { t } = useTranslation()
  const { isInverted, apiPair } = parseCurrencyPair(pair)
  const { data: rate } = useLatestRate(apiPair)

  const [baseCurrency, quoteCurrency] = pair.split('/')
  const [baseAmount, setBaseAmount] = useState<string>('1000')
  const [quoteAmount, setQuoteAmount] = useState<string>('0')
  const [lastEdited, setLastEdited] = useState<'base' | 'quote'>('base')

  const exchangeRate = rate ? (isInverted ? 1 / rate.rate : rate.rate) : 0

  useEffect(() => {
    if (exchangeRate === 0) return

    if (lastEdited === 'base') {
      const base = parseFloat(baseAmount) || 0
      const quote = base * exchangeRate
      setQuoteAmount(quote.toFixed(2))
    } else {
      const quote = parseFloat(quoteAmount) || 0
      const base = quote / exchangeRate
      setBaseAmount(base.toFixed(2))
    }
  }, [baseAmount, quoteAmount, exchangeRate, lastEdited])

  const handleBaseAmountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    if (value === '' || /^\d*\.?\d*$/.test(value)) {
      setBaseAmount(value)
      setLastEdited('base')
    }
  }

  const handleQuoteAmountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    if (value === '' || /^\d*\.?\d*$/.test(value)) {
      setQuoteAmount(value)
      setLastEdited('quote')
    }
  }

  const handleSwap = () => {
    setBaseAmount(quoteAmount)
    setQuoteAmount(baseAmount)
    setLastEdited(lastEdited === 'base' ? 'quote' : 'base')
  }

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
          <CalculateIcon sx={{ color: 'primary.main' }} />
          <Typography variant="h6" color="primary.dark">
            {t('converter.title')}
          </Typography>
        </Box>

        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          {/* Base Currency Input */}
          <TextField
            label={baseCurrency}
            value={baseAmount}
            onChange={handleBaseAmountChange}
            type="text"
            inputMode="decimal"
            fullWidth
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <Typography variant="body2" color="text.secondary">
                    {baseCurrency}
                  </Typography>
                </InputAdornment>
              ),
            }}
            sx={{
              '& .MuiOutlinedInput-root': {
                fontSize: '1.1rem',
                fontWeight: 500,
              }
            }}
          />

          {/* Swap Button */}
          <Box sx={{ display: 'flex', justifyContent: 'center', my: -1 }}>
            <IconButton
              onClick={handleSwap}
              sx={{
                bgcolor: 'primary.light',
                color: 'white',
                '&:hover': { bgcolor: 'primary.main' },
                width: 40,
                height: 40,
              }}
              size="small"
            >
              <SwapVertIcon />
            </IconButton>
          </Box>

          {/* Quote Currency Input */}
          <TextField
            label={quoteCurrency}
            value={quoteAmount}
            onChange={handleQuoteAmountChange}
            type="text"
            inputMode="decimal"
            fullWidth
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <Typography variant="body2" color="text.secondary">
                    {quoteCurrency}
                  </Typography>
                </InputAdornment>
              ),
            }}
            sx={{
              '& .MuiOutlinedInput-root': {
                fontSize: '1.1rem',
                fontWeight: 500,
              }
            }}
          />

          {/* Exchange Rate Info */}
          <Box
            sx={{
              bgcolor: 'primary.light',
              p: 1.5,
              borderRadius: 1,
              mt: 1,
            }}
          >
            <Typography variant="body2" color="white" textAlign="center">
              1 {baseCurrency} = {exchangeRate.toFixed(6)} {quoteCurrency}
            </Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  )
}

export default CurrencyConverter
