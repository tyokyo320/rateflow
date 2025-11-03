import { Container, Box } from '@mui/material'
import Header from './components/Header'
import Footer from './components/Footer'
import Dashboard from './features/Dashboard'

function App() {
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
