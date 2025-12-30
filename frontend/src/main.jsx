import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.jsx'
import './index.css'
import { initTracing } from './services/tracing'

// Initialize OpenTelemetry tracing before rendering the app
// Wrap in try-catch to prevent tracing errors from blocking app startup
try {
  initTracing()
} catch (error) {
  console.error('Failed to initialize tracing:', error)
  // Continue with app startup even if tracing fails
}

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)

