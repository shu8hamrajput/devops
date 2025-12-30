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

// Token management
const TOKEN_KEY = 'access_token'
const REFRESH_TOKEN_KEY = 'refresh_token'

export const tokenManager = {
  setTokens: (accessToken, refreshToken) => {
    if (accessToken) localStorage.setItem(TOKEN_KEY, accessToken)
    if (refreshToken) localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  },
  
  getAccessToken: () => localStorage.getItem(TOKEN_KEY),
  getRefreshToken: () => localStorage.getItem(REFRESH_TOKEN_KEY),
  
  clearTokens: () => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  },
  
  isAuthenticated: () => !!localStorage.getItem(TOKEN_KEY),
}

// Add token to requests
api.interceptors.request.use(
  (config) => {
    const token = tokenManager.getAccessToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Handle token refresh on 401 errors
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      const refreshToken = tokenManager.getRefreshToken()
      if (!refreshToken) {
        tokenManager.clearTokens()
        window.location.href = '/login'
        return Promise.reject(error)
      }

      try {
        const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        })

        const { access_token, refresh_token } = response.data
        tokenManager.setTokens(access_token, refresh_token)

        originalRequest.headers.Authorization = `Bearer ${access_token}`
        return api(originalRequest)
      } catch (refreshError) {
        tokenManager.clearTokens()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

// Auth API
export const authAPI = {
  register: (name, email, password) =>
    api.post('/auth/register', { name, email, password }),
  
  login: (email, password) =>
    api.post('/auth/login', { email, password }),
  
  refreshToken: (refreshToken) =>
    api.post('/auth/refresh', { refresh_token: refreshToken }),
  
  logout: () => {
    tokenManager.clearTokens()
  },
}

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
  
  getAllGroups: () => 
    api.get('/groups'),
  
  getGroup: (id) => 
    api.get(`/groups/${id}`),
  
  updateGroup: (id, name, userIDs) =>
    api.put(`/groups/${id}`, { name, user_ids: userIDs }),
  
  deleteGroup: (id) =>
    api.delete(`/groups/${id}`),
  
  getUserGroups: (userId) =>
    api.get(`/users/${userId}/groups`),
  
  addUserToGroup: (groupId, userId) => 
    api.post(`/groups/${groupId}/users`, { user_id: userId }),
  
  removeUserFromGroup: (groupId, userId) =>
    api.delete(`/groups/${groupId}/users/${userId}`),
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
  
  createUserToUserExpense: (description, amount, paidBy, owedBy) =>
    api.post('/expenses', {
      description,
      amount,
      paid_by: paidBy,
      owed_by: owedBy,
    }),
  
  getExpense: (id) => 
    api.get(`/expenses/${id}`),
  
  updateExpense: (id, description, amount, paidBy, groupId, owedBy) =>
    api.put(`/expenses/${id}`, {
      description,
      amount,
      paid_by: paidBy,
      group_id: groupId,
      owed_by: owedBy,
    }),
  
  deleteExpense: (id) =>
    api.delete(`/expenses/${id}`),
  
  getUserExpenses: (userId) =>
    api.get(`/users/${userId}/expenses`),
  
  getExpensesByGroup: (groupId) => 
    api.get(`/groups/${groupId}/expenses`),
}

export default api

