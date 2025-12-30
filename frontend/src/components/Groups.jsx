import { useState, useEffect } from 'react'
import { groupAPI, userAPI, balanceAPI } from '../services/api'
import { useToastContext } from '../context/ToastContext'
import './Groups.css'

function Groups() {
  const [groups, setGroups] = useState([])
  const [users, setUsers] = useState([])
  const [groupBalances, setGroupBalances] = useState({})
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const { success, error: showError } = useToastContext()
  const [formData, setFormData] = useState({ name: '', user_ids: [] })
  const [showForm, setShowForm] = useState(false)
  const [editingGroup, setEditingGroup] = useState(null)
  const [deletingGroup, setDeletingGroup] = useState(null)
  const [removingUser, setRemovingUser] = useState(null)
  const [groupLookupId, setGroupLookupId] = useState('')
  const [lookedUpGroup, setLookedUpGroup] = useState(null)

  useEffect(() => {
    loadUsers()
    loadGroups()
  }, [])

  const loadUsers = async () => {
    try {
      const response = await userAPI.getAllUsers()
      setUsers(response.data)
    } catch (err) {
      console.error('Failed to load users:', err)
    }
  }

  const loadGroups = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await groupAPI.getAllGroups()
      const groupsData = response.data || []
      setGroups(groupsData)
      
      // Load balances for all groups
      const balancePromises = groupsData.map(group =>
        balanceAPI.getGroupBalance(group.id)
          .then(res => ({ groupId: group.id, balance: res.data }))
          .catch(() => null)
      )
      const balanceResults = await Promise.all(balancePromises)
      const balanceMap = {}
      balanceResults.forEach(result => {
        if (result) {
          balanceMap[result.groupId] = result.balance
        }
      })
      setGroupBalances(balanceMap)
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to load groups';
      setError(typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage))
      setGroups([])
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (formData.user_ids.length === 0) {
      setError('Please select at least one user')
      return
    }
    setLoading(true)
    setError(null)
    try {
      if (editingGroup) {
        // Update existing group
        await groupAPI.updateGroup(editingGroup.id, formData.name, formData.user_ids)
        setEditingGroup(null)
      } else {
        // Create new group
        await groupAPI.createGroup(formData.name, formData.user_ids)
      }
      setFormData({ name: '', user_ids: [] })
      setShowForm(false)
      success(editingGroup ? 'Group updated successfully' : 'Group created successfully')
      // Reload groups to get the latest list
      await loadGroups()
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || (editingGroup ? 'Failed to update group' : 'Failed to create group');
      const message = typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage)
      setError(message)
      showError(message)
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (group) => {
    setEditingGroup(group)
    setFormData({
      name: group.name,
      user_ids: group.user_ids || []
    })
    setShowForm(true)
  }

  const handleCancelEdit = () => {
    setEditingGroup(null)
    setFormData({ name: '', user_ids: [] })
    setShowForm(false)
  }

  const handleDeleteClick = (group) => {
    setDeletingGroup(group)
  }

  const handleDeleteConfirm = async () => {
    if (!deletingGroup) return
    
    setLoading(true)
    setError(null)
    try {
      await groupAPI.deleteGroup(deletingGroup.id)
      setDeletingGroup(null)
      success('Group deleted successfully')
      await loadGroups()
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to delete group';
      const message = typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage)
      setError(message)
      showError(message)
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteCancel = () => {
    setDeletingGroup(null)
  }

  const handleRemoveUserClick = (group, userId) => {
    setRemovingUser({ group, userId })
  }

  const handleRemoveUserConfirm = async () => {
    if (!removingUser) return
    
    setLoading(true)
    setError(null)
    try {
      await groupAPI.removeUserFromGroup(removingUser.group.id, removingUser.userId)
      setRemovingUser(null)
      success('User removed from group successfully')
      await loadGroups()
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to remove user from group';
      const message = typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage)
      setError(message)
      showError(message)
    } finally {
      setLoading(false)
    }
  }

  const handleRemoveUserCancel = () => {
    setRemovingUser(null)
  }

  const handleLookupGroup = async () => {
    if (!groupLookupId.trim()) {
      setError('Please enter a group ID')
      return
    }
    setLoading(true)
    setError(null)
    try {
      const response = await groupAPI.getGroup(groupLookupId.trim())
      const group = response.data
      setLookedUpGroup(group)
      // Add to groups list if not already there
      if (!groups.find(g => g.id === group.id)) {
        setGroups([...groups, group])
      }
    } catch (err) {
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to find group';
      setError(typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage))
      setLookedUpGroup(null)
    } finally {
      setLoading(false)
    }
  }

  const toggleUser = (userId) => {
    setFormData({
      ...formData,
      user_ids: formData.user_ids.includes(userId)
        ? formData.user_ids.filter(id => id !== userId)
        : [...formData.user_ids, userId]
    })
  }

  return (
    <div className="groups-container">
      <div className="section-header">
        <h2>Groups</h2>
        {!editingGroup && (
          <button 
            className="btn-primary"
            onClick={() => setShowForm(!showForm)}
          >
            {showForm ? 'Cancel' : '+ Create Group'}
          </button>
        )}
      </div>

      {showForm && (
        <form className="group-form" onSubmit={handleSubmit}>
          <div className="form-header">
            <h3>{editingGroup ? 'Edit Group' : 'Create Group'}</h3>
            {editingGroup && (
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
            <label>Group Name</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>Select Users</label>
            <div className="user-checkboxes">
              {users.map((user) => (
                <label key={user.id} className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={formData.user_ids.includes(user.id)}
                    onChange={() => toggleUser(user.id)}
                  />
                  <span>{user.name} ({user.email})</span>
                </label>
              ))}
            </div>
            {users.length === 0 && (
              <p className="hint">Create users first before creating a group</p>
            )}
          </div>
          <div className="form-actions">
            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? (editingGroup ? 'Updating...' : 'Creating...') : (editingGroup ? 'Update Group' : 'Create Group')}
            </button>
          </div>
        </form>
      )}

      {deletingGroup && (
        <div className="modal-overlay" onClick={handleDeleteCancel}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Delete Group</h3>
            <p>Are you sure you want to delete the group "{deletingGroup.name}"?</p>
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

      {removingUser && (
        <div className="modal-overlay" onClick={handleRemoveUserCancel}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Remove User from Group</h3>
            <p>Are you sure you want to remove {users.find(u => u.id === removingUser.userId)?.name || removingUser.userId} from "{removingUser.group.name}"?</p>
            <div className="modal-actions">
              <button 
                className="btn-secondary"
                onClick={handleRemoveUserCancel}
                disabled={loading}
              >
                Cancel
              </button>
              <button 
                className="btn-danger"
                onClick={handleRemoveUserConfirm}
                disabled={loading}
              >
                {loading ? 'Removing...' : 'Remove'}
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="lookup-section">
        <h3>Lookup Group by ID</h3>
        <div className="lookup-controls">
          <input
            type="text"
            placeholder="Enter group ID"
            value={groupLookupId}
            onChange={(e) => setGroupLookupId(e.target.value)}
            className="lookup-input"
          />
          <button 
            className="btn-primary"
            onClick={handleLookupGroup}
            disabled={loading}
          >
            Lookup
          </button>
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}

      {loading && !showForm && <div className="loading">Loading...</div>}

      <div className="groups-list">
        {groups.map((group) => (
          <div key={group.id} className="group-card">
            <div className="group-info">
              <h3>{group.name}</h3>
              <p>{group.user_ids?.length || 0} member(s)</p>
              {group.user_ids && group.user_ids.length > 0 && (
                <div className="group-members">
                  <strong>Members:</strong>
                  <ul>
                    {group.user_ids.map((userId, idx) => {
                      const user = users.find(u => u.id === userId)
                      return (
                        <li key={idx} className="member-item">
                          <span>{user ? `${user.name} (${user.email})` : userId}</span>
                          <button
                            className="btn-remove-member"
                            onClick={() => handleRemoveUserClick(group, userId)}
                            title="Remove user from group"
                          >
                            ×
                          </button>
                        </li>
                      )
                    })}
                  </ul>
                </div>
              )}
              {groupBalances[group.id] && groupBalances[group.id].balances && groupBalances[group.id].balances.length > 0 && (
                <div className="group-balance-summary">
                  <strong>Group Balances:</strong>
                  <ul>
                    {groupBalances[group.id].balances.slice(0, 3).map((balance, idx) => {
                      const fromUser = users.find(u => u.id === balance.from_user_id)
                      const toUser = users.find(u => u.id === balance.to_user_id)
                      const amount = balance.amount || 0
                      return (
                        <li key={idx}>
                          {fromUser?.name || balance.from_user_id} owes {toUser?.name || balance.to_user_id} ${amount.toFixed(2)}
                        </li>
                      )
                    })}
                    {groupBalances[group.id].balances.length > 3 && (
                      <li className="more-balances">+{groupBalances[group.id].balances.length - 3} more</li>
                    )}
                  </ul>
                </div>
              )}
            </div>
            <div className="group-actions">
              <button 
                className="btn-edit"
                onClick={() => handleEdit(group)}
                title="Edit group"
              >
                Edit
              </button>
              <button 
                className="btn-delete"
                onClick={() => handleDeleteClick(group)}
                title="Delete group"
              >
                Delete
              </button>
            </div>
            <div className="group-id">ID: {group.id}</div>
          </div>
        ))}
        {groups.length === 0 && !loading && (
          <div className="empty-state">
            No groups found. Create a new group to get started!
          </div>
        )}
      </div>
    </div>
  )
}

export default Groups

