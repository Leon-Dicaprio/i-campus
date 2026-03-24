import axios from 'axios'

// Toggle between real API and mock API
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true'

// 环境配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

// 创建 axios 实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器：自动添加 username
api.interceptors.request.use(
  (config) => {
    // 排除登录和注册接口，不要干扰身份验证
    const isAuthRequest = config.url.includes('/login') || config.url.includes('/register')
    
    const username = localStorage.getItem('campus_ai_username')
    if (username && !isAuthRequest) {
      if (config.method === 'post') {
        // POST 请求把 username 放在 body
        if (config.data) {
          config.data.username = username
        } else {
          config.data = { username }
        }
      } else {
        // GET/DELETE 请求把 username 放在 params
        config.params = { ...config.params, username }
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器：统一错误处理
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // 未授权，跳转登录
      localStorage.removeItem('campus_ai_username')
      localStorage.removeItem('campus_ai_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Mock API functions
const mockAuthAPI = {
  register: async (username, password) => {
    await new Promise(resolve => setTimeout(resolve, 500))
    return { data: { message: 'User registered successfully' } }
  },

  login: async (username, password) => {
    await new Promise(resolve => setTimeout(resolve, 500))
    return { data: { message: 'Login successful' } }
  }
}

const mockConversationAPI = {
  list: async () => {
    await new Promise(resolve => setTimeout(resolve, 300))
    return {
      data: [
        {
          id: '171234567890',
          user_id: 'test',
          title: '关于食堂的问题',
          created_at: '2024-05-20T10:00:00Z',
          updated_at: '2024-05-20T10:05:00Z'
        },
        {
          id: '171234567891',
          user_id: 'test',
          title: '宿舍相关咨询',
          created_at: '2024-05-19T15:30:00Z',
          updated_at: '2024-05-19T15:35:00Z'
        }
      ]
    }
  },

  create: async (title = 'New Chat') => {
    await new Promise(resolve => setTimeout(resolve, 300))
    return {
      data: {
        id: Date.now().toString(),
        title,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    }
  },

  get: async (conversationId) => {
    await new Promise(resolve => setTimeout(resolve, 300))
    return {
      data: {
        id: conversationId,
        title: 'Mock Conversation',
        messages: [
          {
            id: 'msg_1',
            role: 'user',
            content: '这是测试消息',
            timestamp: new Date().toISOString()
          },
          {
            id: 'msg_2',
            role: 'assistant',
            content: '这是AI的回复',
            timestamp: new Date().toISOString()
          }
        ]
      }
    }
  },

  delete: async (conversationId) => {
    await new Promise(resolve => setTimeout(resolve, 300))
    return { data: { message: 'Conversation deleted' } }
  },

  update: async (conversationId, data) => {
    await new Promise(resolve => setTimeout(resolve, 300))
    return { data: { id: conversationId, ...data } }
  }
}

const mockChatAPI = {
  sendMessage: async (conversationId, content, onToken, onStatus, onError) => {
    // Simulate status updates
    onStatus('思考中...')
    await new Promise(resolve => setTimeout(resolve, 500))

    onStatus('检索知识库...')
    await new Promise(resolve => setTimeout(resolve, 500))

    // Simulate streaming response
    const response = '这是AI的模拟回复。基于您的问题，我为您准备了相关信息。在实际使用中，这里会是真实的AI回复内容。'
    const tokens = response.split('').map(char => char)

    for (const token of tokens) {
      onToken(token)
      await new Promise(resolve => setTimeout(resolve, 50))
    }
  }
}

// Real API functions
const realAuthAPI = {
  register: (username, password) =>
    api.post('/api/register', { username, password }),

  login: (username, password) =>
    api.post('/api/login', { username, password })
}

const realConversationAPI = {
  list: () =>
    api.get('/api/conversations'),

  create: (title = 'New Chat') =>
    api.post('/api/conversations', { title }),

  get: (conversationId) =>
    api.get(`/api/conversations/${conversationId}`),

  delete: (conversationId) =>
    api.delete(`/api/conversations/${conversationId}`),

  update: (conversationId, data) =>
    api.patch(`/api/conversations/${conversationId}`, data)
}

const realChatAPI = {
  sendMessage: async (conversationId, content, onToken, onStatus, onError) => {
    const username = localStorage.getItem('campus_ai_username')

    try {
      const response = await fetch(`${API_BASE_URL}/api/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'text/event-stream'
        },
        body: JSON.stringify({
          username,
          conversation_id: conversationId,
          content
        })
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value)
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.slice(6))

              switch (data.type) {
                case 'token':
                  onToken?.(data.content)
                  break
                case 'status':
                  onStatus?.(data.content)
                  break
                case 'error':
                  onError?.(new Error(data.content))
                  break
              }
            } catch (e) {
              console.warn('Failed to parse SSE data:', line)
            }
          }
        }
      }
    } catch (error) {
      onError?.(error)
    }
  }
}

// Export based on USE_MOCK flag
export const authAPI = USE_MOCK ? mockAuthAPI : realAuthAPI
export const conversationAPI = USE_MOCK ? mockConversationAPI : realConversationAPI
export const chatAPI = USE_MOCK ? mockChatAPI : realChatAPI

export default api
