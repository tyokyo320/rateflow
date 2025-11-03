import { Alert, AlertTitle, Button, Box } from '@mui/material'
import RefreshIcon from '@mui/icons-material/Refresh'
import { useTranslation } from 'react-i18next'

interface ErrorAlertProps {
  title?: string
  message: string
  onRetry?: () => void
}

function ErrorAlert({
  title,
  message,
  onRetry
}: ErrorAlertProps) {
  const { t } = useTranslation()
  const displayTitle = title || t('error.loadFailed')

  return (
    <Box sx={{ width: '100%' }}>
      <Alert
        severity="error"
        action={
          onRetry && (
            <Button
              color="inherit"
              size="small"
              onClick={onRetry}
              startIcon={<RefreshIcon />}
            >
              {t('error.retry')}
            </Button>
          )
        }
      >
        <AlertTitle>{displayTitle}</AlertTitle>
        {message}
      </Alert>
    </Box>
  )
}

export default ErrorAlert
