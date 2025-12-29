import { useState } from 'react'
import './App.css'
import Users from './components/Users'
import Groups from './components/Groups'
import Expenses from './components/Expenses'

function App() {
  const [activeTab, setActiveTab] = useState('users')

  return (
    <div className="app">
      <header className="app-header">
        <h1>Splitwise</h1>
        <p>Expense Sharing Made Easy</p>
      </header>
      
      <nav className="app-nav">
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
        {activeTab === 'users' && <Users />}
        {activeTab === 'groups' && <Groups />}
        {activeTab === 'expenses' && <Expenses />}
      </main>
    </div>
  )
}

export default App

