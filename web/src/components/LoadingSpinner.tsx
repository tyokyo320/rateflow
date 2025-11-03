import { Box, CircularProgress, Typography } from '@mui/material'
import { useTranslation } from 'react-i18next'

interface LoadingSpinnerProps {
  message?: string
}

function LoadingSpinner({ message }: LoadingSpinnerProps) {
  const { t } = useTranslation()
  const displayMessage = message || t('loading.loading')

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: 200,
        gap: 2,
      }}
    >
      <CircularProgress />
      <Typography variant="body2" color="text.secondary">
        {displayMessage}
      </Typography>
    </Box>
  )
}

export default LoadingSpinner
