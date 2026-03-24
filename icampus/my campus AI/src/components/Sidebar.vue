<template>
  <div class="sidebar">
    <div class="user-section">
      <router-link to="/profile" class="user-menu">
        <div class="user-avatar">
          <img :src="userAvatar" alt="User" @error="handleImageError" />
          <div v-if="imageError" class="default-avatar">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
              <path d="M20 21C20 21 19 20 12 20C5 20 4 21 4 21" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M12 11C14.2091 11 16 9.20914 16 7C16 4.79086 14.2091 3 12 3C9.79086 3 8 4.79086 8 7C8 9.20914 9.79086 11 12 11Z" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
        </div>
        <div class="user-info">
          <div class="username">{{ username }}</div>
          <div class="user-email">{{ userEmail }}</div>
        </div>
        <div class="settings-icon">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M12 15C13.6569 15 15 13.6569 15 12C15 10.3431 13.6569 9 12 9C10.3431 9 9 10.3431 9 12C9 13.6569 10.3431 15 12 15Z" 
                  stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M12 2C15.866 2 19.2859 3.30324 21.8284 5.39018C19.7415 7.93271 18.4382 11.3526 18.4382 15.2196C18.4382 19.0866 19.7415 22.5065 21.8284 25.049C19.2859 27.1359 15.866 28.4392 12 28.4392C8.13401 28.4392 4.71412 27.1359 2.17157 25.049C4.25851 22.5065 5.56176 19.0866 5.56176 15.2196C5.56176 11.3526 4.25851 7.93271 2.17157 5.39018C4.71412 3.30324 8.13401 2 12 2Z" 
                  stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
      </router-link>
    </div>
    
    <div class="sidebar-header">
      <button @click="startNewChat" class="new-chat-btn">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 4V20M4 12H20" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
        新对话
      </button>
    </div>
    
    <div class="chat-history">
      <div v-if="chatHistory.length === 0" class="empty-state">
        <p>没有历史对话</p>
      </div>
      <div v-else>
        <div 
          v-for="chat in chatHistory" 
          :key="chat.id"
          class="chat-item"
          :class="{ active: currentChatId === chat.id }"
          @click="$emit('load-chat', chat.id)"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M21 15C21 15.5304 20.7893 16.0391 20.4142 16.4142C20.0391 16.7893 19.5304 17 19 17H7L3 21V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H19C19.5304 3 20.0391 3.21071 20.4142 3.58579C20.7893 3.96086 21 4.46957 21 5V15Z" 
                  stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <span class="truncate">{{ chat.title || '新对话' }}</span>
          <button 
            class="delete-btn" 
            @click.stop="deleteChat(chat.id)"
            title="删除对话"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
              <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import { conversationAPI } from '@/services/api'

  const props = defineProps({
    currentChatId: {
      type: String,
      default: ''
    }
  })

  const emit = defineEmits(['new-chat', 'load-chat'])

  const router = useRouter()
  const chatHistory = ref([])
  const username = ref('')
  const userEmail = ref('')
  const userAvatar = ref('/images/user-avatar.png')
  const imageError = ref(false)

  const handleImageError = () => {
    imageError.value = true
  }

  const deleteChat = async (chatId) => {
    if (confirm('确定要删除这个对话吗？')) {
      try {
        await conversationAPI.delete(chatId)
        localStorage.removeItem(`chat_title_${chatId}`)
        await loadChatHistory()
        
        if (props.currentChatId === chatId) {
          emit('load-chat', '') // 清空当前聊天
        }
      } catch (error) {
        alert('删除失败：' + (error.response?.data?.error || error.message))
      }
    }
  }

  const startNewChat = () => {
    emit('new-chat')
  }

  const loadChatHistory = async () => {
    try {
      const conversations = await conversationAPI.list()
      const list = conversations.data || []
      for (const c of list) {
        const localTitle = localStorage.getItem(`chat_title_${c.id}`)
        if (localTitle && (!c.title || c.title === 'New Chat' || c.title === '新对话')) {
          c.title = localTitle
        }
      }
      chatHistory.value = list
    } catch (error) {
      console.error('Failed to load chat history:', error)
      chatHistory.value = []
    }
  }

  const loadUserProfile = () => {
    const storedUsername = localStorage.getItem('campus_ai_username')
    if (storedUsername) {
      username.value = storedUsername
      userEmail.value = `${storedUsername}@example.com`
    }
  }

  onMounted(() => {
    loadChatHistory()
    loadUserProfile()
  })
  
  defineExpose({
    loadChatHistory,
    updateChatTitle: (chatId, title) => {
      const chat = chatHistory.value.find(c => c.id === chatId)
      if (chat) {
        chat.title = title
        return true
      }
      return false
    }
  })
</script>

<style scoped>
.sidebar {
  width: 280px;
  background: rgba(255, 255, 255, 0.95);
  border-right: 1px solid rgba(255, 255, 255, 0.2);
  display: flex;
  flex-direction: column;
  height: 100vh;
  backdrop-filter: blur(10px);
  box-shadow: 4px 0 20px rgba(0, 0, 0, 0.1);
}

.sidebar-header {
  padding: 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.new-chat-btn {
  width: 100%;
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  color: white;
  border: none;
  padding: 0.875rem 1rem;
  border-radius: 12px;
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 600;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  box-shadow: 0 4px 15px rgba(39, 174, 96, 0.28);
}

.new-chat-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(39, 174, 96, 0.36);
}

.chat-history {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
}

.empty-state {
  text-align: center;
  color: #7f8c8d;
  padding: 2rem;
  font-style: italic;
}

.chat-item {
  padding: 0.875rem;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  margin-bottom: 0.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: #2c3e50;
  border: 1px solid transparent;
  height: 60px;
  min-height: 60px;
  max-height: 60px;
}

.chat-item:hover {
  background: rgba(39, 174, 96, 0.12);
  border-color: rgba(39, 174, 96, 0.22);
  transform: translateX(2px);
}

.chat-item.active {
  background: linear-gradient(45deg, rgba(46, 204, 113, 0.18), rgba(39, 174, 96, 0.18));
  border-color: #27ae60;
  font-weight: 600;
}

.chat-item svg {
  flex-shrink: 0;
  color: #27ae60;
}

.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  color: inherit;
}

.delete-btn {
  display: none;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: rgba(231, 76, 60, 0.1);
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.3s ease;
  color: #e74c3c;
  opacity: 0.7;
}

.delete-btn:hover {
  background: rgba(231, 76, 60, 0.2);
  opacity: 1;
  transform: scale(1.1);
}

.chat-item:hover .delete-btn {
  display: flex;
}

.user-section {
  padding: 0.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.5);
  flex-shrink: 0;
}

.user-menu {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  justify-content: space-between;
  height: 40px;
  text-decoration: none;
  color: inherit;
  transition: all 0.3s ease;
  padding: 0.25rem;
  border-radius: 8px;
  width: 100%;
}

.user-menu:hover {
  background: rgba(39, 174, 96, 0.12);
  transform: translateY(-1px);
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid #2ecc71;
  box-shadow: 0 2px 8px rgba(39, 174, 96, 0.26);
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.default-avatar {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(45deg, #2ecc71, #27ae60);
}

.user-info {
  flex: 1;
  min-width: 0;
}

.username {
  font-weight: 600;
  color: #2c3e50;
  font-size: 0.9rem;
}

.user-email {
  font-size: 0.75rem;
  color: #7f8c8d;
  margin-top: 2px;
}

.settings-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #7f8c8d;
  transition: all 0.3s ease;
}

.user-menu:hover .settings-icon {
  color: #27ae60;
}

/* Custom scrollbar */
::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}

::-webkit-scrollbar-thumb {
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  border-radius: 10px;
}

::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(45deg, #27ae60, #1f8f4d);
}
</style>
