import { Box, Container, Typography, Link, IconButton } from '@mui/material'
import GitHubIcon from '@mui/icons-material/GitHub'
import FavoriteIcon from '@mui/icons-material/Favorite'
import { useTranslation } from 'react-i18next'

function Footer() {
  const { t } = useTranslation()

  return (
    <Box
      component="footer"
      sx={{
        py: 3,
        px: 2,
        mt: 'auto',
        borderTop: '1px solid',
        borderColor: 'divider',
        bgcolor: 'background.paper',
      }}
    >
      <Container maxWidth="xl">
        <Box
          sx={{
            display: 'flex',
            flexDirection: { xs: 'column', sm: 'row' },
            justifyContent: 'space-between',
            alignItems: 'center',
            gap: 2,
          }}
        >
          {/* Left: Copyright and Made with Love */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <Typography variant="body2" color="text.secondary">
              © {new Date().getFullYear()} {t('app.title')}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mx: 1 }}>
              •
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
              {t('footer.madeWith')}
              <FavoriteIcon sx={{ fontSize: 16, color: 'error.main' }} />
              {t('footer.by')} tyokyo320
            </Typography>
          </Box>

          {/* Right: GitHub Link */}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Typography variant="body2" color="text.secondary">
              {t('footer.openSource')}
            </Typography>
            <IconButton
              component={Link}
              href="https://github.com/tyokyo320/rateflow"
              target="_blank"
              rel="noopener noreferrer"
              size="small"
              sx={{
                color: 'text.secondary',
                '&:hover': {
                  color: 'primary.main',
                },
              }}
            >
              <GitHubIcon />
            </IconButton>
          </Box>
        </Box>
      </Container>
    </Box>
  )
}

export default Footer
