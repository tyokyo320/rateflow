import { Box, FormControl, InputLabel, Select, MenuItem, SelectChangeEvent, IconButton, Tooltip } from '@mui/material'
import SwapHorizIcon from '@mui/icons-material/SwapHoriz'
import { useTranslation } from 'react-i18next'
import type { Currency } from '../../types'

interface CurrencyPairSelectorProps {
  baseCurrency: Currency
  quoteCurrency: Currency
  onBaseCurrencyChange: (currency: Currency) => void
  onQuoteCurrencyChange: (currency: Currency) => void
  onSwap: () => void
}

const currencies: Currency[] = ['CNY', 'JPY', 'USD']

function CurrencyPairSelector({
  baseCurrency,
  quoteCurrency,
  onBaseCurrencyChange,
  onQuoteCurrencyChange,
  onSwap,
}: CurrencyPairSelectorProps) {
  const { t } = useTranslation()

  const handleBaseChange = (event: SelectChangeEvent) => {
    onBaseCurrencyChange(event.target.value as Currency)
  }

  const handleQuoteChange = (event: SelectChangeEvent) => {
    onQuoteCurrencyChange(event.target.value as Currency)
  }

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <FormControl sx={{ minWidth: 100 }} size="small">
        <InputLabel>{t('currency.base')}</InputLabel>
        <Select value={baseCurrency} label={t('currency.base')} onChange={handleBaseChange}>
          {currencies.map((currency) => (
            <MenuItem key={currency} value={currency}>
              {currency}
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      <Tooltip title={t('currency.swap')}>
        <IconButton
          onClick={onSwap}
          sx={{
            bgcolor: '#e3f2fd',
            color: 'primary.main',
            '&:hover': { bgcolor: '#bbdefb' }
          }}
        >
          <SwapHorizIcon />
        </IconButton>
      </Tooltip>

      <FormControl sx={{ minWidth: 100 }} size="small">
        <InputLabel>{t('currency.quote')}</InputLabel>
        <Select value={quoteCurrency} label={t('currency.quote')} onChange={handleQuoteChange}>
          {currencies.map((currency) => (
            <MenuItem key={currency} value={currency}>
              {currency}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    </Box>
  )
}

export default CurrencyPairSelector
