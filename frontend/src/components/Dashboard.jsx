import { useState, useEffect } from 'react'
import { groupAPI, expenseAPI, userAPI, balanceAPI, tokenManager } from '../services/api'
import './Dashboard.css'

function Dashboard() {
  const [groups, setGroups] = useState([])
  const [users, setUsers] = useState([])
  const [allUsers, setAllUsers] = useState([])
  const [userBalance, setUserBalance] = useState(null)
  const [groupBalances, setGroupBalances] = useState({})
  const [userToUserBalances, setUserToUserBalances] = useState({})
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
      const [groupsRes, expensesRes, balanceRes] = await Promise.all([
        groupAPI.getUserGroups(userId),
        expenseAPI.getUserExpenses(userId),
        balanceAPI.getUserBalance(userId).catch(() => null)
      ])
      
      setGroups(groupsRes.data || [])
      
      // Load user balance
      if (balanceRes) {
        setUserBalance(balanceRes.data)
      }
      
      // Load group balances
      const groupBalancePromises = (groupsRes.data || []).map(group =>
        balanceAPI.getGroupBalance(group.id)
          .then(res => ({ groupId: group.id, balance: res.data }))
          .catch(() => null)
      )
      const groupBalanceResults = await Promise.all(groupBalancePromises)
      const groupBalanceMap = {}
      groupBalanceResults.forEach(result => {
        if (result) {
          groupBalanceMap[result.groupId] = result.balance
        }
      })
      setGroupBalances(groupBalanceMap)
      
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
      
      // Get user details and balances for each user ID
      const userPromises = Array.from(userSet).map(otherUserId => 
        Promise.all([
          userAPI.getUser(otherUserId).catch(() => null),
          balanceAPI.getUserToUserBalance(userId, otherUserId).catch(() => null)
        ]).then(([userRes, balanceRes]) => ({
          user: userRes?.data,
          balance: balanceRes?.data?.balance || 0
        }))
      )
      const userResults = await Promise.all(userPromises)
      setUsers(userResults.filter(u => u.user !== null).map(u => u.user))
      
      const userBalanceMap = {}
      userResults.forEach(result => {
        if (result.user) {
          userBalanceMap[result.user.id] = result.balance
        }
      })
      setUserToUserBalances(userBalanceMap)
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

      {userBalance && (
        <div className="balance-cards">
          <div className="balance-card">
            <h3>Your Balance Summary</h3>
            <div className="balance-details">
              <div className="balance-item">
                <span className="balance-label">You Owe:</span>
                <span className="balance-value negative">${(userBalance.total_owed || 0).toFixed(2)}</span>
              </div>
              <div className="balance-item">
                <span className="balance-label">You Are Owed:</span>
                <span className="balance-value positive">${(userBalance.total_owed_to || 0).toFixed(2)}</span>
              </div>
              <div className="balance-item net">
                <span className="balance-label">Net Balance:</span>
                <span className={`balance-value ${(userBalance.net_balance || 0) >= 0 ? 'positive' : 'negative'}`}>
                  ${Math.abs(userBalance.net_balance || 0).toFixed(2)}
                  {(userBalance.net_balance || 0) >= 0 ? ' owed to you' : ' you owe'}
                </span>
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="dashboard-content">
        <div className="dashboard-section">
          <h3>My Groups</h3>
          {groups.length === 0 ? (
            <div className="empty-state">No groups found. Create a group to get started!</div>
          ) : (
            <div className="groups-list">
              {groups.map((group) => {
                const groupBalance = groupBalances[group.id]
                return (
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
                    {groupBalance && groupBalance.balances && groupBalance.balances.length > 0 && (
                      <div className="group-balance">
                        <strong>Balances:</strong>
                        <ul>
                          {groupBalance.balances.slice(0, 3).map((balance, idx) => {
                            const fromUser = allUsers.find(u => u.id === balance.from_user_id)
                            const toUser = allUsers.find(u => u.id === balance.to_user_id)
                            const amount = balance.amount || 0
                            return (
                              <li key={idx}>
                                {fromUser?.name || balance.from_user_id} owes {toUser?.name || balance.to_user_id} ${amount.toFixed(2)}
                              </li>
                            )
                          })}
                          {groupBalance.balances.length > 3 && (
                            <li className="more-balances">+{groupBalance.balances.length - 3} more</li>
                          )}
                        </ul>
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>

        <div className="dashboard-section">
          <h3>Users I Have Expenses With</h3>
          {users.length === 0 ? (
            <div className="empty-state">No expenses with other users yet.</div>
          ) : (
            <div className="users-list">
              {users.map((user) => {
                const balance = userToUserBalances[user.id] || 0
                return (
                  <div key={user.id} className="user-card">
                    <div className="user-info">
                      <h4>{user.name}</h4>
                      <p>{user.email}</p>
                    </div>
                    {balance !== 0 && balance !== undefined && balance !== null && (
                      <div className="user-balance">
                        {balance > 0 ? (
                          <span className="balance-positive">You owe ${(balance || 0).toFixed(2)}</span>
                        ) : (
                          <span className="balance-negative">Owes you ${Math.abs(balance || 0).toFixed(2)}</span>
                        )}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default Dashboard
