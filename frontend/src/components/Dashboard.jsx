import { useState, useEffect } from 'react'
import { groupAPI, expenseAPI, userAPI, tokenManager } from '../services/api'
import './Dashboard.css'

function Dashboard() {
  const [groups, setGroups] = useState([])
  const [users, setUsers] = useState([])
  const [allUsers, setAllUsers] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [showExpenseForm, setShowExpenseForm] = useState(false)
  const [expenseType, setExpenseType] = useState('group') // 'group' or 'user'
  const [formData, setFormData] = useState({
    description: '',
    amount: '',
    paid_by: '',
    group_id: '',
    owed_by: ''
  })

  useEffect(() => {
    loadDashboard()
    loadAllUsers()
  }, [])

  const getCurrentUserId = () => {
    // Get user ID from token (we'll need to decode JWT or get it from login response)
    // For now, we'll get it from the first expense or group
    // In a real app, you'd store the user ID after login
    const token = tokenManager.getAccessToken()
    if (!token) return null
    
    // Simple JWT decode (just for getting user ID, not for validation)
    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      return payload.sub
    } catch (e) {
      return null
    }
  }

  const loadDashboard = async () => {
    const userId = getCurrentUserId()
    if (!userId) {
      setError('User not authenticated')
      return
    }

    setLoading(true)
    setError(null)
    try {
      const [groupsRes, expensesRes] = await Promise.all([
        groupAPI.getUserGroups(userId),
        expenseAPI.getUserExpenses(userId)
      ])
      
      setGroups(groupsRes.data || [])
      
      // Extract unique users from expenses
      const userSet = new Set()
      expensesRes.data?.forEach(expense => {
        if (expense.paid_by && expense.paid_by !== userId) {
          userSet.add(expense.paid_by)
        }
        if (expense.owed_by && expense.owed_by !== userId) {
          userSet.add(expense.owed_by)
        }
      })
      
      // Get user details for each user ID
      const userPromises = Array.from(userSet).map(userId => 
        userAPI.getUser(userId).catch(() => null)
      )
      const userResults = await Promise.all(userPromises)
      setUsers(userResults.filter(u => u !== null).map(u => u.data))
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to load dashboard'
      setError(typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage))
    } finally {
      setLoading(false)
    }
  }

  const loadAllUsers = async () => {
    try {
      const response = await userAPI.getAllUsers()
      setAllUsers(response.data || [])
    } catch (err) {
      console.error('Failed to load users:', err)
    }
  }

  const handleExpenseSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    
    try {
      if (expenseType === 'group') {
        await expenseAPI.createExpense(
          formData.description,
          parseFloat(formData.amount),
          formData.paid_by,
          formData.group_id
        )
      } else {
        await expenseAPI.createUserToUserExpense(
          formData.description,
          parseFloat(formData.amount),
          formData.paid_by,
          formData.owed_by
        )
      }
      
      setFormData({ description: '', amount: '', paid_by: '', group_id: '', owed_by: '' })
      setShowExpenseForm(false)
      loadDashboard()
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to create expense'
      setError(typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage))
    } finally {
      setLoading(false)
    }
  }

  const getUserName = (userId) => {
    const user = allUsers.find(u => u.id === userId)
    return user ? user.name : userId
  }

  return (
    <div className="dashboard-container">
      <div className="dashboard-header">
        <h2>Dashboard</h2>
        <button 
          className="btn-primary"
          onClick={() => setShowExpenseForm(!showExpenseForm)}
        >
          {showExpenseForm ? 'Cancel' : '+ Add Expense'}
        </button>
      </div>

      {showExpenseForm && (
        <div className="expense-form-card">
          <h3>Add Expense</h3>
          <div className="expense-type-selector">
            <button
              className={expenseType === 'group' ? 'active' : ''}
              onClick={() => setExpenseType('group')}
            >
              Group Expense
            </button>
            <button
              className={expenseType === 'user' ? 'active' : ''}
              onClick={() => setExpenseType('user')}
            >
              User-to-User Expense
            </button>
          </div>
          <form onSubmit={handleExpenseSubmit}>
            <div className="form-group">
              <label>Description</label>
              <input
                type="text"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                required
              />
            </div>
            <div className="form-group">
              <label>Amount</label>
              <input
                type="number"
                step="0.01"
                min="0"
                value={formData.amount}
                onChange={(e) => setFormData({ ...formData, amount: e.target.value })}
                required
              />
            </div>
            <div className="form-group">
              <label>Paid By</label>
              <select
                value={formData.paid_by}
                onChange={(e) => setFormData({ ...formData, paid_by: e.target.value })}
                required
              >
                <option value="">Select a user</option>
                {allUsers.map(user => (
                  <option key={user.id} value={user.id}>
                    {user.name} ({user.email})
                  </option>
                ))}
              </select>
            </div>
            {expenseType === 'group' ? (
              <div className="form-group">
                <label>Group</label>
                <select
                  value={formData.group_id}
                  onChange={(e) => setFormData({ ...formData, group_id: e.target.value })}
                  required
                >
                  <option value="">Select a group</option>
                  {groups.map(group => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
              </div>
            ) : (
              <div className="form-group">
                <label>Owed By</label>
                <select
                  value={formData.owed_by}
                  onChange={(e) => setFormData({ ...formData, owed_by: e.target.value })}
                  required
                >
                  <option value="">Select a user</option>
                  {allUsers.map(user => (
                    <option key={user.id} value={user.id}>
                      {user.name} ({user.email})
                    </option>
                  ))}
                </select>
              </div>
            )}
            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Creating...' : 'Create Expense'}
            </button>
          </form>
        </div>
      )}

      {error && <div className="error-message">{error}</div>}

      {loading && !showExpenseForm && <div className="loading">Loading dashboard...</div>}

      <div className="dashboard-content">
        <div className="dashboard-section">
          <h3>My Groups</h3>
          {groups.length === 0 ? (
            <div className="empty-state">No groups found. Create a group to get started!</div>
          ) : (
            <div className="groups-list">
              {groups.map((group) => (
                <div key={group.id} className="group-card">
                  <div className="group-header">
                    <h4>{group.name}</h4>
                    <span className="group-meta">
                      {new Date(group.updated_at).toLocaleDateString()}
                    </span>
                  </div>
                  <div className="group-members">
                    <strong>Members:</strong> {group.user_ids.length}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="dashboard-section">
          <h3>Users I Have Expenses With</h3>
          {users.length === 0 ? (
            <div className="empty-state">No expenses with other users yet.</div>
          ) : (
            <div className="users-list">
              {users.map((user) => (
                <div key={user.id} className="user-card">
                  <div className="user-info">
                    <h4>{user.name}</h4>
                    <p>{user.email}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default Dashboard
