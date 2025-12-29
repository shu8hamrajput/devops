import axios from 'axios'

// In production (Docker), use relative /api path which Nginx will proxy to backend
// In development, use the VITE_API_URL or fallback to localhost:8080
const getApiBaseUrl = () => {
  // Check if we're in production (built app)
  if (import.meta.env.PROD) {
    // In production, use relative /api path (Nginx will proxy)
    return '/api'
  }
  // In development, use environment variable or localhost
  return import.meta.env.VITE_API_URL || 'http://localhost:8080'
}

const API_BASE_URL = getApiBaseUrl()

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// User API
export const userAPI = {
  createUser: (name, email) => 
    api.post('/users', { name, email }),
  
  getAllUsers: () => 
    api.get('/users'),
  
  getUser: (id) => 
    api.get(`/users/${id}`),
}

// Group API
export const groupAPI = {
  createGroup: (name, userIDs) => 
    api.post('/groups', { name, user_ids: userIDs }),
  
  getGroup: (id) => 
    api.get(`/groups/${id}`),
  
  addUserToGroup: (groupId, userId) => 
    api.post(`/groups/${groupId}/users`, { user_id: userId }),
}

// Expense API
export const expenseAPI = {
  createExpense: (description, amount, paidBy, groupId) => 
    api.post('/expenses', {
      description,
      amount,
      paid_by: paidBy,
      group_id: groupId,
    }),
  
  getExpense: (id) => 
    api.get(`/expenses/${id}`),
  
  getExpensesByGroup: (groupId) => 
    api.get(`/groups/${groupId}/expenses`),
}

export default api

