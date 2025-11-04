import { useState } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  Typography,
  Box,
  Chip,
  Button,
  ButtonGroup,
  IconButton,
} from '@mui/material'
import DownloadIcon from '@mui/icons-material/Download'
import FirstPageIcon from '@mui/icons-material/FirstPage'
import LastPageIcon from '@mui/icons-material/LastPage'
import KeyboardArrowLeft from '@mui/icons-material/KeyboardArrowLeft'
import KeyboardArrowRight from '@mui/icons-material/KeyboardArrowRight'
import { useTranslation } from 'react-i18next'
import { useHistoricalRates } from '../../api/hooks'
import { formatDate, formatRate, formatRelativeTime, parseCurrencyPair } from '../../utils/formatters'
import { exportToCSV, exportToJSON } from '../../utils/export'
import LoadingSpinner from '../../components/LoadingSpinner'
import ErrorAlert from '../../components/ErrorAlert'

interface RateHistoryTableProps {
  pair: string
  days: number
}

interface TablePaginationActionsProps {
  count: number
  page: number
  rowsPerPage: number
  onPageChange: (event: React.MouseEvent<HTMLButtonElement>, newPage: number) => void
}

function TablePaginationActions(props: TablePaginationActionsProps) {
  const { count, page, rowsPerPage, onPageChange } = props

  const handleFirstPageButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    onPageChange(event, 0)
  }

  const handleBackButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    onPageChange(event, page - 1)
  }

  const handleNextButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    onPageChange(event, page + 1)
  }

  const handleLastPageButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    onPageChange(event, Math.max(0, Math.ceil(count / rowsPerPage) - 1))
  }

  return (
    <Box sx={{ flexShrink: 0, ml: 2.5 }}>
      <IconButton
        onClick={handleFirstPageButtonClick}
        disabled={page === 0}
        aria-label="first page"
      >
        <FirstPageIcon />
      </IconButton>
      <IconButton
        onClick={handleBackButtonClick}
        disabled={page === 0}
        aria-label="previous page"
      >
        <KeyboardArrowLeft />
      </IconButton>
      <IconButton
        onClick={handleNextButtonClick}
        disabled={page >= Math.ceil(count / rowsPerPage) - 1}
        aria-label="next page"
      >
        <KeyboardArrowRight />
      </IconButton>
      <IconButton
        onClick={handleLastPageButtonClick}
        disabled={page >= Math.ceil(count / rowsPerPage) - 1}
        aria-label="last page"
      >
        <LastPageIcon />
      </IconButton>
    </Box>
  )
}

function RateHistoryTable({ pair }: RateHistoryTableProps) {
  const { t } = useTranslation()
  const { isInverted, apiPair } = parseCurrencyPair(pair)
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(10)

  const { data, isLoading, error, refetch } = useHistoricalRates(
    apiPair,
    page + 1,
    rowsPerPage
  )

  const handleChangePage = (_: unknown, newPage: number) => {
    setPage(newPage)
  }

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10))
    setPage(0)
  }

  const handleExportCSV = () => {
    if (data?.items) {
      const filename = `${pair.replace('/', '-')}-rates-${new Date().toISOString().split('T')[0]}.csv`
      exportToCSV(data.items, filename)
    }
  }

  const handleExportJSON = () => {
    if (data?.items) {
      const filename = `${pair.replace('/', '-')}-rates-${new Date().toISOString().split('T')[0]}.json`
      exportToJSON(data.items, filename)
    }
  }

  if (error) {
    return (
      <ErrorAlert
        title={t('error.cannotLoadHistory')}
        message={error.message}
        onRetry={() => refetch()}
      />
    )
  }

  if (isLoading) {
    return <LoadingSpinner message={t('loading.loadingHistory')} />
  }

  const rates = data?.items || []
  const total = data?.pagination?.total || 0

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h6">
          {t('dashboard.historicalData')}
        </Typography>

        {rates.length > 0 && (
          <ButtonGroup size="small" variant="outlined">
            <Button startIcon={<DownloadIcon />} onClick={handleExportCSV}>
              CSV
            </Button>
            <Button startIcon={<DownloadIcon />} onClick={handleExportJSON}>
              JSON
            </Button>
          </ButtonGroup>
        )}
      </Box>

      {rates.length === 0 ? (
        <Box sx={{ textAlign: 'center', py: 4 }}>
          <Typography color="text.secondary">{t('dashboard.noData')}</Typography>
        </Box>
      ) : (
        <>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>{t('table.date')}</TableCell>
                  <TableCell>{t('table.pair')}</TableCell>
                  <TableCell align="right">{t('table.rate')}</TableCell>
                  <TableCell>{t('table.source')}</TableCell>
                  <TableCell>{t('table.updatedAt')}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {rates.map((rate) => {
                  const displayRate = isInverted ? 1 / rate.rate : rate.rate

                  return (
                    <TableRow
                      key={rate.id}
                      sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                    >
                      <TableCell>
                        {formatDate(rate.effectiveDate, 'YYYY-MM-DD')}
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={pair}
                          size="small"
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2" fontWeight={600}>
                          {formatRate(displayRate, isInverted ? 4 : 6)}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={rate.source}
                          size="small"
                          color="primary"
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {formatRelativeTime(rate.updatedAt)}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </TableContainer>

          <TablePagination
            component="div"
            count={total}
            page={page}
            onPageChange={handleChangePage}
            rowsPerPage={rowsPerPage}
            onRowsPerPageChange={handleChangeRowsPerPage}
            rowsPerPageOptions={[5, 10, 25, 50, 100]}
            labelRowsPerPage={t('table.rowsPerPage')}
            labelDisplayedRows={({ from, to, count }) =>
              t('table.displayedRows', { from, to, count: typeof count === 'number' ? count : Number(count) })
            }
            ActionsComponent={TablePaginationActions}
          />
        </>
      )}
    </Box>
  )
}

export default RateHistoryTable
