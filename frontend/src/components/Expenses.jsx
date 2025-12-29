import { useState, useEffect } from 'react'
import { expenseAPI, groupAPI, userAPI } from '../services/api'
import './Expenses.css'

function Expenses() {
  const [expenses, setExpenses] = useState([])
  const [groups, setGroups] = useState([])
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [formData, setFormData] = useState({ 
    description: '', 
    amount: '', 
    paid_by: '', 
    group_id: '' 
  })
  const [showForm, setShowForm] = useState(false)
  const [selectedGroup, setSelectedGroup] = useState('')

  useEffect(() => {
    loadUsers()
    loadGroups()
  }, [])

  useEffect(() => {
    if (selectedGroup) {
      loadExpenses(selectedGroup)
    } else {
      setExpenses([])
    }
  }, [selectedGroup])

  const loadExpenses = async (groupId) => {
    setLoading(true)
    setError(null)
    try {
      const response = await expenseAPI.getExpensesByGroup(groupId)
      setExpenses(response.data)
    } catch (err) {
      setError(err.response?.data || 'Failed to load expenses')
      setExpenses([])
    } finally {
      setLoading(false)
    }
  }

  const loadGroups = async () => {
    // Note: Backend doesn't have a GET /groups endpoint
    // Groups will need to be loaded individually or stored locally
    // For now, we'll use an empty array and let users enter group IDs manually
    setGroups([])
  }

  const loadUsers = async () => {
    try {
      const response = await userAPI.getAllUsers()
      setUsers(response.data)
    } catch (err) {
      console.error('Failed to load users:', err)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      await expenseAPI.createExpense(
        formData.description,
        parseFloat(formData.amount),
        formData.paid_by,
        formData.group_id
      )
      setFormData({ description: '', amount: '', paid_by: '', group_id: '' })
      setShowForm(false)
      if (selectedGroup) {
        loadExpenses(selectedGroup)
      }
    } catch (err) {
      setError(err.response?.data || 'Failed to create expense')
    } finally {
      setLoading(false)
    }
  }

  const getUserName = (userId) => {
    const user = users.find(u => u.id === userId)
    return user ? user.name : userId
  }

  const getGroupName = (groupId) => {
    const group = groups.find(g => g.id === groupId)
    return group ? group.name : groupId
  }

  return (
    <div className="expenses-container">
      <div className="section-header">
        <h2>Expenses</h2>
        <button 
          className="btn-primary"
          onClick={() => setShowForm(!showForm)}
        >
          {showForm ? 'Cancel' : '+ Add Expense'}
        </button>
      </div>

      <div className="filter-section">
        <label>Filter by Group ID:</label>
        <input
          type="text"
          placeholder="Enter group ID to view expenses"
          value={selectedGroup}
          onChange={(e) => setSelectedGroup(e.target.value)}
          className="group-input"
        />
        <button 
          className="btn-primary"
          onClick={() => selectedGroup && loadExpenses(selectedGroup)}
          disabled={!selectedGroup || loading}
        >
          Load Expenses
        </button>
      </div>

      {showForm && (
        <form className="expense-form" onSubmit={handleSubmit}>
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
              {users.map(user => (
                <option key={user.id} value={user.id}>
                  {user.name} ({user.email})
                </option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label>Group ID</label>
            <input
              type="text"
              placeholder="Enter group ID"
              value={formData.group_id}
              onChange={(e) => setFormData({ ...formData, group_id: e.target.value })}
              required
            />
            <p className="hint">Enter the ID of the group this expense belongs to</p>
          </div>
          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? 'Creating...' : 'Create Expense'}
          </button>
        </form>
      )}

      {error && <div className="error-message">{error}</div>}

      {loading && !showForm && <div className="loading">Loading expenses...</div>}

      {!selectedGroup && (
        <div className="info-message">
          Enter a group ID above to view its expenses
        </div>
      )}

      <div className="expenses-list">
        {expenses.map((expense) => (
          <div key={expense.id} className="expense-card">
            <div className="expense-header">
              <h3>{expense.description}</h3>
              <span className="expense-amount">${parseFloat(expense.amount).toFixed(2)}</span>
            </div>
            <div className="expense-details">
              <p><strong>Paid by:</strong> {getUserName(expense.paid_by)}</p>
              <p><strong>Group:</strong> {getGroupName(expense.group_id)}</p>
              {expense.created_at && (
                <p><strong>Date:</strong> {new Date(expense.created_at).toLocaleString()}</p>
              )}
            </div>
            <div className="expense-id">ID: {expense.id}</div>
          </div>
        ))}
        {expenses.length === 0 && selectedGroup && !loading && (
          <div className="empty-state">No expenses found for this group.</div>
        )}
      </div>
    </div>
  )
}

export default Expenses

