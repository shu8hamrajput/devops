import { useState, useEffect } from 'react'
import { expenseAPI, groupAPI, userAPI } from '../services/api'
import { useToastContext } from '../context/ToastContext'
import './Expenses.css'

function Expenses() {
  const [expenses, setExpenses] = useState([])
  const [groups, setGroups] = useState([])
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const { success, error: showError } = useToastContext()
  const [formData, setFormData] = useState({ 
    description: '', 
    amount: '', 
    paid_by: '', 
    group_id: '' 
  })
  const [showForm, setShowForm] = useState(false)
  const [editingExpense, setEditingExpense] = useState(null)
  const [deletingExpense, setDeletingExpense] = useState(null)
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
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to load expenses';
      setError(typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage))
      setExpenses([])
    } finally {
      setLoading(false)
    }
  }

  const loadGroups = async () => {
    try {
      const response = await groupAPI.getAllGroups()
      setGroups(response.data || [])
    } catch (err) {
      console.error('Failed to load groups:', err)
      setGroups([])
    }
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
      if (editingExpense) {
        // Update existing expense
        await expenseAPI.updateExpense(
          editingExpense.id,
          formData.description,
          parseFloat(formData.amount),
          formData.paid_by,
          formData.group_id || null,
          null
        )
        setEditingExpense(null)
      } else {
        // Create new expense
        await expenseAPI.createExpense(
          formData.description,
          parseFloat(formData.amount),
          formData.paid_by,
          formData.group_id
        )
      }
      setFormData({ description: '', amount: '', paid_by: '', group_id: '' })
      setShowForm(false)
      success(editingExpense ? 'Expense updated successfully' : 'Expense created successfully')
      if (selectedGroup) {
        loadExpenses(selectedGroup)
      }
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || (editingExpense ? 'Failed to update expense' : 'Failed to create expense');
      const message = typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage)
      setError(message)
      showError(message)
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (expense) => {
    setEditingExpense(expense)
    setFormData({
      description: expense.description,
      amount: expense.amount.toString(),
      paid_by: expense.paid_by,
      group_id: expense.group_id || ''
    })
    setShowForm(true)
  }

  const handleCancelEdit = () => {
    setEditingExpense(null)
    setFormData({ description: '', amount: '', paid_by: '', group_id: '' })
    setShowForm(false)
  }

  const handleDeleteClick = (expense) => {
    setDeletingExpense(expense)
  }

  const handleDeleteConfirm = async () => {
    if (!deletingExpense) return
    
    setLoading(true)
    setError(null)
    try {
      await expenseAPI.deleteExpense(deletingExpense.id)
      setDeletingExpense(null)
      success('Expense deleted successfully')
      if (selectedGroup) {
        loadExpenses(selectedGroup)
      }
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to delete expense';
      const message = typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage)
      setError(message)
      showError(message)
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteCancel = () => {
    setDeletingExpense(null)
  }

  const getUserName = (userId) => {
    const user = users.find(u => u.id === userId)
    return user ? user.name : userId
  }

  const getGroupName = (groupId) => {
    if (!groupId) return 'N/A'
    const group = groups.find(g => g.id === groupId)
    return group ? group.name : groupId
  }

  return (
    <div className="expenses-container">
      <div className="section-header">
        <h2>Expenses</h2>
        {!editingExpense && (
          <button 
            className="btn-primary"
            onClick={() => setShowForm(!showForm)}
          >
            {showForm ? 'Cancel' : '+ Add Expense'}
          </button>
        )}
      </div>

      <div className="filter-section">
        <label>Filter by Group:</label>
        <select
          value={selectedGroup}
          onChange={(e) => setSelectedGroup(e.target.value)}
          className="group-select"
        >
          <option value="">Select a group</option>
          {groups.map(group => (
            <option key={group.id} value={group.id}>
              {group.name}
            </option>
          ))}
        </select>
        {selectedGroup && (
          <button 
            className="btn-secondary"
            onClick={() => {
              setSelectedGroup('')
              setExpenses([])
            }}
          >
            Clear Filter
          </button>
        )}
      </div>

      {showForm && (
        <form className="expense-form" onSubmit={handleSubmit}>
          <div className="form-header">
            <h3>{editingExpense ? 'Edit Expense' : 'Add Expense'}</h3>
            {editingExpense && (
              <button 
                type="button"
                className="btn-secondary"
                onClick={handleCancelEdit}
              >
                Cancel
              </button>
            )}
          </div>
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
            {groups.length === 0 && (
              <p className="hint">No groups available. Create a group first.</p>
            )}
          </div>
          <div className="form-actions">
            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? (editingExpense ? 'Updating...' : 'Creating...') : (editingExpense ? 'Update Expense' : 'Create Expense')}
            </button>
          </div>
        </form>
      )}

      {deletingExpense && (
        <div className="modal-overlay" onClick={handleDeleteCancel}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Delete Expense</h3>
            <p>Are you sure you want to delete the expense "{deletingExpense.description}"?</p>
            <p className="warning-text">This action cannot be undone.</p>
            <div className="modal-actions">
              <button 
                className="btn-secondary"
                onClick={handleDeleteCancel}
                disabled={loading}
              >
                Cancel
              </button>
              <button 
                className="btn-danger"
                onClick={handleDeleteConfirm}
                disabled={loading}
              >
                {loading ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}

      {error && <div className="error-message">{error}</div>}

      {loading && !showForm && <div className="loading">Loading expenses...</div>}

      {!selectedGroup && (
        <div className="info-message">
          Select a group above to view its expenses
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
              {expense.group_id && (
                <p><strong>Group:</strong> {getGroupName(expense.group_id)}</p>
              )}
              {expense.owed_by && (
                <p><strong>Owed by:</strong> {getUserName(expense.owed_by)}</p>
              )}
              {expense.created_at && (
                <p><strong>Date:</strong> {new Date(expense.created_at).toLocaleString()}</p>
              )}
            </div>
            <div className="expense-actions">
              <button 
                className="btn-edit"
                onClick={() => handleEdit(expense)}
                title="Edit expense"
              >
                Edit
              </button>
              <button 
                className="btn-delete"
                onClick={() => handleDeleteClick(expense)}
                title="Delete expense"
              >
                Delete
              </button>
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

