import { useState, useEffect } from 'react'
import { userAPI } from '../services/api'
import './Users.css'

function Users() {
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [formData, setFormData] = useState({ name: '', email: '' })
  const [showForm, setShowForm] = useState(false)

  useEffect(() => {
    loadUsers()
  }, [])

  const loadUsers = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await userAPI.getAllUsers()
      setUsers(response.data)
    } catch (err) {
      setError(err.response?.data || 'Failed to load users')
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      await userAPI.createUser(formData.name, formData.email)
      setFormData({ name: '', email: '' })
      setShowForm(false)
      loadUsers()
    } catch (err) {
      setError(err.response?.data || 'Failed to create user')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="users-container">
      <div className="section-header">
        <h2>Users</h2>
        <button 
          className="btn-primary"
          onClick={() => setShowForm(!showForm)}
        >
          {showForm ? 'Cancel' : '+ Add User'}
        </button>
      </div>

      {showForm && (
        <form className="user-form" onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Name</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>Email</label>
            <input
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              required
            />
          </div>
          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? 'Creating...' : 'Create User'}
          </button>
        </form>
      )}

      {error && <div className="error-message">{error}</div>}

      {loading && !showForm && <div className="loading">Loading users...</div>}

      <div className="users-list">
        {users.map((user) => (
          <div key={user.id} className="user-card">
            <div className="user-info">
              <h3>{user.name}</h3>
              <p>{user.email}</p>
            </div>
            <div className="user-id">ID: {user.id}</div>
          </div>
        ))}
        {users.length === 0 && !loading && (
          <div className="empty-state">No users found. Create your first user!</div>
        )}
      </div>
    </div>
  )
}

export default Users

