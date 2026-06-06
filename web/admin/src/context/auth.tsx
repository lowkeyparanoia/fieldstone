import React, { createContext, useContext, useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { authApi } from '@/lib/api'

interface User {
  id: string
  email: string
  verified: boolean
}

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  register: (email: string, password: string) => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    // Check for existing token
    const token = localStorage.getItem('fieldstone_token')
    if (token) {
      // Verify token and get user info
      authApi.refresh().then((response) => {
        if (response.data.token) {
          localStorage.setItem('fieldstone_token', response.data.token)
          // In real implementation, decode token or fetch user
          setUser({ id: '1', email: 'admin@fieldstone.io', verified: true })
        }
      }).catch(() => {
        localStorage.removeItem('fieldstone_token')
      }).finally(() => {
        setIsLoading(false)
      })
    } else {
      setIsLoading(false)
    }
  }, [])

  const login = async (email: string, password: string) => {
    try {
      const response = await authApi.login(email, password)
      const { token, user } = response.data
      localStorage.setItem('fieldstone_token', token)
      setUser(user)
      navigate('/')
    } catch (error) {
      throw error
    }
  }

  const register = async (email: string, password: string) => {
    try {
      const response = await authApi.register(email, password)
      const { token, user } = response.data
      localStorage.setItem('fieldstone_token', token)
      setUser(user)
      navigate('/')
    } catch (error) {
      throw error
    }
  }

  const logout = () => {
    localStorage.removeItem('fieldstone_token')
    setUser(null)
    navigate('/login')
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
        register,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
