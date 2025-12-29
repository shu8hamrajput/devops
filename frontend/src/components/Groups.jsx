import { useState, useEffect } from 'react'
import { groupAPI, userAPI } from '../services/api'
import './Groups.css'

function Groups() {
  const [groups, setGroups] = useState([])
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [formData, setFormData] = useState({ name: '', user_ids: [] })
  const [showForm, setShowForm] = useState(false)
  const [groupLookupId, setGroupLookupId] = useState('')
  const [lookedUpGroup, setLookedUpGroup] = useState(null)

  useEffect(() => {
    loadUsers()
  }, [])

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
    if (formData.user_ids.length === 0) {
      setError('Please select at least one user')
      return
    }
    setLoading(true)
    setError(null)
    try {
      const response = await groupAPI.createGroup(formData.name, formData.user_ids)
      const newGroup = response.data
      setGroups([...groups, newGroup])
      setFormData({ name: '', user_ids: [] })
      setShowForm(false)
    } catch (err) {
      setError(err.response?.data || 'Failed to create group')
    } finally {
      setLoading(false)
    }
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
      setError(err.response?.data || 'Failed to find group')
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
        <button 
          className="btn-primary"
          onClick={() => setShowForm(!showForm)}
        >
          {showForm ? 'Cancel' : '+ Create Group'}
        </button>
      </div>

      {showForm && (
        <form className="group-form" onSubmit={handleSubmit}>
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
          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? 'Creating...' : 'Create Group'}
          </button>
        </form>
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
                        <li key={idx}>
                          {user ? `${user.name} (${user.email})` : userId}
                        </li>
                      )
                    })}
                  </ul>
                </div>
              )}
            </div>
            <div className="group-id">ID: {group.id}</div>
          </div>
        ))}
        {groups.length === 0 && !loading && (
          <div className="empty-state">
            No groups loaded. Create a new group or lookup an existing one by ID!
          </div>
        )}
      </div>
    </div>
  )
}

export default Groups

