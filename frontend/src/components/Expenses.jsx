import { useState, useEffect } from 'react'
import { expenseAPI, groupAPI, userAPI } from '../services/api'
import { useToastContext } from '../context/ToastContext'
import './Expenses.css'

const SHARE_TYPES = {
  EQUAL: 'EQUAL',
  PERCENTAGE: 'PERCENTAGE',
  EXACT_AMOUNT: 'EXACT_AMOUNT',
  SHARES: 'SHARES'
}

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
  const [splitConfig, setSplitConfig] = useState({
    shareType: SHARE_TYPES.EQUAL,
    userShares: {}
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

  const getSelectedGroupMembers = () => {
    if (!formData.group_id) return []
    const group = groups.find(g => g.id === formData.group_id)
    return group ? (group.user_ids || []) : []
  }

  const initializeSplitConfig = () => {
    const members = getSelectedGroupMembers()
    const newUserShares = {}
    members.forEach(userId => {
      if (splitConfig.shareType === SHARE_TYPES.EQUAL) {
        newUserShares[userId] = 1.0
      } else {
        newUserShares[userId] = 0
      }
    })
    setSplitConfig({ ...splitConfig, userShares: newUserShares })
  }

  useEffect(() => {
    if (formData.group_id && !editingExpense) {
      initializeSplitConfig()
    }
  }, [formData.group_id])

  const handleSplitTypeChange = (shareType) => {
    const members = getSelectedGroupMembers()
    const newUserShares = {}
    
    if (shareType === SHARE_TYPES.EQUAL) {
      members.forEach(userId => {
        newUserShares[userId] = 1.0
      })
    } else {
      members.forEach(userId => {
        newUserShares[userId] = 0
      })
    }
    
    setSplitConfig({ shareType, userShares: newUserShares })
  }

  const handleUserShareChange = (userId, value) => {
    setSplitConfig({
      ...splitConfig,
      userShares: {
        ...splitConfig.userShares,
        [userId]: parseFloat(value) || 0
      }
    })
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      if (editingExpense) {
        // Update existing expense (splits not updated in edit mode for now)
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
        // Create new expense with split configuration
        const useCustomSplit = splitConfig.shareType !== SHARE_TYPES.EQUAL || 
          Object.values(splitConfig.userShares).some(v => v !== 1.0 && v !== 0)
        
        if (useCustomSplit && Object.keys(splitConfig.userShares).length > 0) {
          await expenseAPI.createExpenseWithSplit(
            formData.description,
            parseFloat(formData.amount),
            formData.paid_by,
            formData.group_id,
            splitConfig.shareType,
            splitConfig.userShares
          )
        } else {
          await expenseAPI.createExpense(
            formData.description,
            parseFloat(formData.amount),
            formData.paid_by,
            formData.group_id
          )
        }
      }
      setFormData({ description: '', amount: '', paid_by: '', group_id: '' })
      setSplitConfig({ shareType: SHARE_TYPES.EQUAL, userShares: {} })
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
    setSplitConfig({ shareType: SHARE_TYPES.EQUAL, userShares: {} })
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

          {!editingExpense && formData.group_id && (
            <div className="form-group">
              <label>Split Type</label>
              <select
                value={splitConfig.shareType}
                onChange={(e) => handleSplitTypeChange(e.target.value)}
              >
                <option value={SHARE_TYPES.EQUAL}>Equal</option>
                <option value={SHARE_TYPES.PERCENTAGE}>Percentage</option>
                <option value={SHARE_TYPES.EXACT_AMOUNT}>Exact Amount</option>
                <option value={SHARE_TYPES.SHARES}>Shares</option>
              </select>
              <p className="hint">
                {splitConfig.shareType === SHARE_TYPES.EQUAL && 'Split equally among all group members'}
                {splitConfig.shareType === SHARE_TYPES.PERCENTAGE && 'Split by percentage (must sum to 100%)'}
                {splitConfig.shareType === SHARE_TYPES.EXACT_AMOUNT && 'Split by exact amounts (must sum to expense amount)'}
                {splitConfig.shareType === SHARE_TYPES.SHARES && 'Split by shares (e.g., 2:1:1 means first person gets 2/4, others get 1/4 each)'}
              </p>
            </div>
          )}

          {!editingExpense && formData.group_id && splitConfig.shareType !== SHARE_TYPES.EQUAL && (
            <div className="form-group">
              <label>Split Configuration</label>
              <div className="split-config">
                {getSelectedGroupMembers().map(userId => {
                  const user = users.find(u => u.id === userId)
                  if (!user) return null
                  
                  return (
                    <div key={userId} className="split-item">
                      <label>{user.name}</label>
                      <input
                        type="number"
                        step={splitConfig.shareType === SHARE_TYPES.PERCENTAGE ? "0.01" : "0.01"}
                        min="0"
                        value={splitConfig.userShares[userId] || 0}
                        onChange={(e) => handleUserShareChange(userId, e.target.value)}
                        placeholder={
                          splitConfig.shareType === SHARE_TYPES.PERCENTAGE ? "0-100" :
                          splitConfig.shareType === SHARE_TYPES.EXACT_AMOUNT ? "Amount" :
                          "Shares"
                        }
                      />
                      <span className="split-unit">
                        {splitConfig.shareType === SHARE_TYPES.PERCENTAGE ? "%" :
                         splitConfig.shareType === SHARE_TYPES.EXACT_AMOUNT ? "$" : ""}
                      </span>
                    </div>
                  )
                })}
              </div>
              {splitConfig.shareType === SHARE_TYPES.PERCENTAGE && (
                <p className="hint">
                  Total: {Object.values(splitConfig.userShares).reduce((a, b) => a + (parseFloat(b) || 0), 0).toFixed(2)}%
                </p>
              )}
              {splitConfig.shareType === SHARE_TYPES.EXACT_AMOUNT && formData.amount && (
                <p className="hint">
                  Total: ${Object.values(splitConfig.userShares).reduce((a, b) => a + (parseFloat(b) || 0), 0).toFixed(2)} / ${parseFloat(formData.amount).toFixed(2)}
                </p>
              )}
            </div>
          )}

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

