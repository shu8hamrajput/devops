import { useState, useEffect } from 'react'
import './App.css'
import Dashboard from './components/Dashboard'
import Users from './components/Users'
import Groups from './components/Groups'
import Expenses from './components/Expenses'
import Login from './components/Login'
import Register from './components/Register'
import ToastContainer from './components/ToastContainer'
import { ToastProvider, useToastContext } from './context/ToastContext'
import { tokenManager, authAPI } from './services/api'

function AppContent() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [showRegister, setShowRegister] = useState(false)
  const [activeTab, setActiveTab] = useState('dashboard')
  const [loading, setLoading] = useState(true)
  const { toasts, removeToast } = useToastContext()

  useEffect(() => {
    // Check if user is authenticated on mount
    setIsAuthenticated(tokenManager.isAuthenticated())
    setLoading(false)
  }, [])

  const handleLoginSuccess = (data) => {
    if (data.access_token && data.refresh_token) {
      tokenManager.setTokens(data.access_token, data.refresh_token)
      setIsAuthenticated(true)
    }
  }

  const handleRegisterSuccess = (data) => {
    // After registration, show login form
    setShowRegister(false)
    // Optionally auto-login if tokens are provided
    if (data.access_token && data.refresh_token) {
      tokenManager.setTokens(data.access_token, data.refresh_token)
      setIsAuthenticated(true)
    }
  }

  const handleLogout = () => {
    authAPI.logout()
    setIsAuthenticated(false)
    setActiveTab('users')
  }

  if (loading) {
    return (
      <div className="app">
        <div className="loading">Loading...</div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <div className="app">
        <header className="app-header">
          <h1>Splitwise</h1>
          <p>Expense Sharing Made Easy</p>
        </header>
        <div className="auth-switch">
          <button
            className={!showRegister ? 'active' : ''}
            onClick={() => setShowRegister(false)}
          >
            Login
          </button>
          <button
            className={showRegister ? 'active' : ''}
            onClick={() => setShowRegister(true)}
          >
            Register
          </button>
        </div>
        {showRegister ? (
          <Register onRegisterSuccess={handleRegisterSuccess} />
        ) : (
          <Login onLoginSuccess={handleLoginSuccess} />
        )}
      </div>
    )
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>Splitwise</h1>
        <p>Expense Sharing Made Easy</p>
        <button className="btn-logout" onClick={handleLogout}>
          Logout
        </button>
      </header>
      
      <nav className="app-nav">
        <button 
          className={activeTab === 'dashboard' ? 'active' : ''}
          onClick={() => setActiveTab('dashboard')}
        >
          Dashboard
        </button>
        <button 
          className={activeTab === 'users' ? 'active' : ''}
          onClick={() => setActiveTab('users')}
        >
          Users
        </button>
        <button 
          className={activeTab === 'groups' ? 'active' : ''}
          onClick={() => setActiveTab('groups')}
        >
          Groups
        </button>
        <button 
          className={activeTab === 'expenses' ? 'active' : ''}
          onClick={() => setActiveTab('expenses')}
        >
          Expenses
        </button>
      </nav>

      <main className="app-main">
        {activeTab === 'dashboard' && <Dashboard />}
        {activeTab === 'users' && <Users />}
        {activeTab === 'groups' && <Groups />}
        {activeTab === 'expenses' && <Expenses />}
      </main>
      <ToastContainer toasts={toasts} onRemove={removeToast} />
    </div>
  )
}

function App() {
  return (
    <ToastProvider>
      <AppContent />
    </ToastProvider>
  )
}

export default App

