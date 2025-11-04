import { AppBar, Toolbar, Typography, Box, Chip, IconButton, Menu, MenuItem } from '@mui/material'
import TrendingUpIcon from '@mui/icons-material/TrendingUp'
import LanguageIcon from '@mui/icons-material/Language'
import Brightness4Icon from '@mui/icons-material/Brightness4'
import Brightness7Icon from '@mui/icons-material/Brightness7'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useHealthCheck } from '../api/hooks'
import { useThemeMode } from '../contexts/ThemeContext'

function Header() {
  const { t, i18n } = useTranslation()
  const { isError } = useHealthCheck()
  const { mode, toggleTheme } = useThemeMode()
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)

  const handleLanguageClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget)
  }

  const handleLanguageClose = () => {
    setAnchorEl(null)
  }

  const handleLanguageChange = (lang: string) => {
    i18n.changeLanguage(lang)
    handleLanguageClose()
  }

  return (
    <AppBar position="static" elevation={0}>
      <Toolbar>
        <TrendingUpIcon sx={{ mr: 2, fontSize: 32 }} />
        <Typography variant="h6" component="div" sx={{ fontWeight: 700 }}>
          {t('app.title')}
        </Typography>
        <Typography variant="body2" sx={{ ml: 1, opacity: 0.8 }}>
          - {t('app.subtitle')}
        </Typography>
        <Box sx={{ flexGrow: 1 }} />

        {/* Theme Toggle */}
        <IconButton
          color="inherit"
          onClick={toggleTheme}
          sx={{ mr: 1 }}
          title={mode === 'dark' ? 'Light mode' : 'Dark mode'}
        >
          {mode === 'dark' ? <Brightness7Icon /> : <Brightness4Icon />}
        </IconButton>

        {/* Language Selector */}
        <IconButton
          color="inherit"
          onClick={handleLanguageClick}
          sx={{ mr: 1 }}
        >
          <LanguageIcon />
        </IconButton>
        <Menu
          anchorEl={anchorEl}
          open={Boolean(anchorEl)}
          onClose={handleLanguageClose}
        >
          <MenuItem
            onClick={() => handleLanguageChange('en')}
            selected={i18n.language === 'en'}
          >
            English
          </MenuItem>
          <MenuItem
            onClick={() => handleLanguageChange('zh')}
            selected={i18n.language === 'zh'}
          >
            中文
          </MenuItem>
        </Menu>

        {/* Health Status - Only show if there's an error */}
        {isError && (
          <Chip
            label={t('app.offline')}
            color="error"
            size="small"
          />
        )}
      </Toolbar>
    </AppBar>
  )
}

export default Header
