import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './App'
import { initAuth } from './services/auth'
import './i18n'
import './styles/global.css'
import './styles/layout.css'

async function bootstrap() {
  const rootEl = document.getElementById('root')
  if (!rootEl) {
    throw new Error('Root element #root not found in document.')
  }

  await initAuth({
    tenantId: '',
    clientId: '',
    redirectUri: window.location.origin,
    apiScope: '',
  })

  createRoot(rootEl).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
}

bootstrap()

