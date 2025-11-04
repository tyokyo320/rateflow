import { useEffect } from 'react'
import { Container, Box } from '@mui/material'
import { useTranslation } from 'react-i18next'
import Header from './components/Header'
import Footer from './components/Footer'
import Dashboard from './features/Dashboard'

function App() {
  const { t, i18n } = useTranslation()

  // Update document title and lang attribute when language changes
  useEffect(() => {
    document.title = `${t('app.title')} - ${t('app.subtitle')}`
    document.documentElement.lang = i18n.language
  }, [t, i18n.language])

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <Header />
      <Container maxWidth="xl" sx={{ mt: 4, mb: 4, flex: 1 }}>
        <Dashboard />
      </Container>
      <Footer />
    </Box>
  )
}

export default App
