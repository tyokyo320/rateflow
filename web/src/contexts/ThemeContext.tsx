import { createContext, useContext, useState, useMemo, useEffect, ReactNode } from 'react'
import { ThemeProvider as MuiThemeProvider, createTheme } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'

type ThemeMode = 'light' | 'dark'

interface ThemeContextType {
  mode: ThemeMode
  toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined)

export const useThemeMode = () => {
  const context = useContext(ThemeContext)
  if (!context) {
    throw new Error('useThemeMode must be used within ThemeProvider')
  }
  return context
}

interface ThemeProviderProps {
  children: ReactNode
}

export function ThemeProvider({ children }: ThemeProviderProps) {
  const [mode, setMode] = useState<ThemeMode>(() => {
    const saved = localStorage.getItem('theme-mode')
    return (saved as ThemeMode) || 'light'
  })

  useEffect(() => {
    localStorage.setItem('theme-mode', mode)
  }, [mode])

  const toggleTheme = () => {
    setMode((prevMode) => (prevMode === 'light' ? 'dark' : 'light'))
  }

  const theme = useMemo(
    () =>
      createTheme({
        palette: {
          mode,
          ...(mode === 'light'
            ? {
                // Light mode colors
                primary: {
                  main: '#5e92f3',
                  light: '#90caf9',
                  dark: '#1976d2',
                },
                secondary: {
                  main: '#ab47bc',
                  light: '#ce93d8',
                  dark: '#8e24aa',
                },
                background: {
                  default: '#fafafa',
                  paper: '#ffffff',
                },
                success: {
                  main: '#4caf50',
                },
                error: {
                  main: '#f44336',
                },
              }
            : {
                // Dark mode colors
                primary: {
                  main: '#90caf9',
                  light: '#bbdefb',
                  dark: '#5e92f3',
                },
                secondary: {
                  main: '#ce93d8',
                  light: '#f3e5f5',
                  dark: '#ab47bc',
                },
                background: {
                  default: '#121212',
                  paper: '#1e1e1e',
                },
                success: {
                  main: '#66bb6a',
                },
                error: {
                  main: '#ef5350',
                },
              }),
        },
        typography: {
          fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
          h4: {
            fontWeight: 700,
          },
          h6: {
            fontWeight: 600,
          },
        },
        shape: {
          borderRadius: 12,
        },
        components: {
          MuiCard: {
            styleOverrides: {
              root: {
                boxShadow: mode === 'light'
                  ? '0 2px 8px rgba(0,0,0,0.08)'
                  : '0 2px 8px rgba(0,0,0,0.4)',
              },
            },
          },
          MuiPaper: {
            styleOverrides: {
              root: {
                backgroundImage: 'none',
              },
            },
          },
        },
      }),
    [mode]
  )

  const contextValue = useMemo(() => ({ mode, toggleTheme }), [mode])

  return (
    <ThemeContext.Provider value={contextValue}>
      <MuiThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </MuiThemeProvider>
    </ThemeContext.Provider>
  )
}
