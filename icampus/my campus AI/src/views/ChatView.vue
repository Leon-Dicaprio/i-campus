<template>
  <div class="chat-container">
    <div class="chat-content">
      <Sidebar ref="sidebarRef" :current-chat-id="currentChatId" @new-chat="startNewChat" @load-chat="loadChat" />
      <div class="main-content">
        <div v-if="!currentChatId" class="welcome-screen">
          <h1>校园AI助手</h1>
          <p>开始一个新的对话或从侧边栏选择历史记录</p>
          <button @click="startNewChat" class="new-chat-btn">新对话</button>
        </div>
        <div v-else class="chat-interface">
          <div ref="chatBox" class="chat-messages">
            <div v-for="(msg, index) in messages" :key="index" class="message" :class="{ 'user-msg': msg.role === 'user', 'ai-msg': msg.role === 'assistant' }">
              <div class="message-avatar">
                <img :src="msg.role === 'user' ? userAvatar : '/images/ai-avatar.png'" alt="avatar" />
              </div>
              <div class="message-content">
                <div class="message-text" v-html="formatMessage(msg.content)"></div>
                <div v-if="msg.role === 'assistant' && msg.relatedQuestions?.length" class="suggested-questions">
                  <div v-for="(question, qIndex) in msg.relatedQuestions" 
                       :key="qIndex"
                       class="suggested-question"
                       @click="selectQuestion(question)">
                    {{ question }}
                  </div>
                </div>
              </div>
            </div>
            <div v-if="loading" class="typing-indicator">
              <span></span><span></span><span></span>
            </div>
          </div>
          <div class="input-container">
            <div class="input-wrapper">
              <textarea
                v-model="userInput"
                @keydown.enter.exact.prevent="!loading && userInput.trim() && sendMessage()"
                @keydown.shift.enter.exact.prevent="userInput += '\n'"
                placeholder="输入消息..."
                :disabled="loading"
                rows="1"
                ref="messageInput"
              ></textarea>
              <button 
                @click="sendMessage" 
                :disabled="!userInput.trim() || loading"
                class="send-button"
                :class="{ 'loading': loading }"
              >
                <svg v-if="!loading" width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M7 11L12 6L17 11M12 18V7" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <span v-else class="loading-dots">
                  <span>.</span><span>.</span><span>.</span>
                </span>
              </button>
            </div>
            <p class="disclaimer">校园AI助手可能会产生不准确或过时的信息</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import Sidebar from '@/components/Sidebar.vue'
import { conversationAPI, chatAPI } from '@/services/api'
import { marked } from 'marked'

// Configure marked options
marked.use({
  breaks: true,
  gfm: true
})

const currentChatId = ref('')
const sidebarRef = ref(null)
const messages = ref([])
const userInput = ref('')
const loading = ref(false)
const userAvatar = ref('/images/user-avatar.png')
const currentStatus = ref('')

const startNewChat = async () => {
  try {
    const response = await conversationAPI.create()
    const newChat = response.data
    currentChatId.value = newChat.id
    messages.value = []
    userInput.value = ''
    loading.value = false
    sidebarRef.value?.loadChatHistory?.()
    
    // Update URL
    const url = new URL(window.location)
    url.searchParams.set('chat', newChat.id)
    window.history.pushState({}, '', url)
  } catch (error) {
    alert('创建对话失败：' + (error.response?.data?.error || error.message))
  }
}

const loadChat = async (chatId) => {
  if (!chatId) {
    currentChatId.value = ''
    messages.value = []
    return
  }
  
  try {
    const response = await conversationAPI.get(chatId)
    const conversation = response.data
    currentChatId.value = conversation.id
    messages.value = conversation.messages || []
    
    // Update URL
    const url = new URL(window.location)
    url.searchParams.set('chat', conversation.id)
    window.history.pushState({}, '', url)
  } catch (error) {
    alert('加载对话失败：' + (error.response?.data?.error || error.message))
  }
}

const sendMessage = async () => {
  if (!userInput.value.trim() || loading.value || !currentChatId.value) return
  
  const userMessage = userInput.value.trim()
  const isFirstMessage = messages.value.length === 0
  if (isFirstMessage) {
    const title = userMessage.split('\n')[0].trim()
    const updated = sidebarRef.value?.updateChatTitle?.(currentChatId.value, title)
    if (!updated) {
      await sidebarRef.value?.loadChatHistory?.()
      sidebarRef.value?.updateChatTitle?.(currentChatId.value, title)
    }
    localStorage.setItem(`chat_title_${currentChatId.value}`, title)
    conversationAPI.update(currentChatId.value, { title }).catch(() => {})
  }
  
  // Add user message immediately
  messages.value.push({
    id: 'user-' + Date.now(),
    role: 'user',
    content: userMessage,
    timestamp: new Date().toISOString()
  })
  
  userInput.value = ''
  loading.value = true
  currentStatus.value = '思考中...'
  
  // Scroll to bottom
  nextTick(() => {
    const chatBox = document.querySelector('.chat-messages')
    if (chatBox) chatBox.scrollTop = chatBox.scrollHeight
  })
  
  try {
    let aiMessageContent = ''
    const aiMessageId = 'ai-' + Date.now()
    
    await chatAPI.sendMessage(
      currentChatId.value,
      userMessage,
      // onToken
      (token) => {
        aiMessageContent += token
        
        // Find or create AI message
        let aiMessage = messages.value.find(m => m.id === aiMessageId)
        if (!aiMessage) {
          aiMessage = {
            id: aiMessageId,
            role: 'assistant',
            content: '',
            timestamp: new Date().toISOString()
          }
          messages.value.push(aiMessage)
        }
        
        aiMessage.content = aiMessageContent
        
        // Auto scroll
        nextTick(() => {
          const chatBox = document.querySelector('.chat-messages')
          if (chatBox) chatBox.scrollTop = chatBox.scrollHeight
        })
      },
      // onStatus
      (status) => {
        currentStatus.value = status
      },
      // onError
      (error) => {
        loading.value = false
        currentStatus.value = ''
        alert('发送失败：' + error.message)
      }
    )
    
    loading.value = false
    currentStatus.value = ''
    
    // Add suggested questions (mock for now)
    const aiMessage = messages.value.find(m => m.id === aiMessageId)
    if (aiMessage) {
      aiMessage.relatedQuestions = generateSuggestedQuestions(userMessage)
    }
    
    sidebarRef.value?.loadChatHistory?.()
    
  } catch (error) {
    loading.value = false
    currentStatus.value = ''
    alert('发送失败：' + error.message)
  }
}

const formatMessage = (content) => {
  if (!content) return ''
  try {
    // Render markdown to HTML
    return marked.parse(content)
  } catch (e) {
    console.error('Markdown parsing error:', e)
    return content
  }
}

const selectQuestion = (question) => {
  userInput.value = question
  nextTick(() => {
    if (!loading.value && userInput.value.trim()) {
      sendMessage()
    }
  })
}

const generateSuggestedQuestions = (userMessage) => {
  const lowerMessage = userMessage.toLowerCase()
  
  // Canteen questions
if (lowerMessage.includes('食堂') || lowerMessage.includes('吃饭') || lowerMessage.includes('餐厅')) {
  return [
    '食堂的营业时间是什么时候？',
    '哪个食堂的菜品最好吃？',
    '食堂的收费标准是怎样的？',
    '校园内还有其他餐饮选择吗？'
  ]
}

// Dormitory questions
if (lowerMessage.includes('宿舍') || lowerMessage.includes('住宿')) {
  return [
    '宿舍的入住流程是怎样的？',
    '宿舍有哪些设施和规定？',
    '如何申请调换宿舍？',
    '宿舍的网络怎么样？'
  ]
}

// Academic questions
if (lowerMessage.includes('课程') || lowerMessage.includes('选课') || lowerMessage.includes('学习')) {
  return [
    '如何进行选课操作？',
    '本学期的课程安排是怎样的？',
    '如何查询成绩？',
    '有哪些推荐的选修课程？'
  ]
}

// 🗺️ Navigation & Location questions (高德MCP)
if (lowerMessage.includes('地图') || lowerMessage.includes('导航') || lowerMessage.includes('位置') || 
    lowerMessage.includes('路线') || lowerMessage.includes('怎么走') || lowerMessage.includes('距离')) {
  return [
    '从宿舍到教学楼怎么走最近？',
    '校园内有哪些主要建筑的定位？',
    '如何查看实时公交/地铁信息？',
    '附近有哪些推荐的地点？'
  ]
}

// 🚌 Transportation questions (高德MCP)
if (lowerMessage.includes('交通') || lowerMessage.includes('公交') || lowerMessage.includes('地铁') || 
    lowerMessage.includes('打车') || lowerMessage.includes('出行')) {
  return [
    '校门口有哪些公交线路？',
    '到市中心最快的交通方式是什么？',
    '地铁首末班车时间是多少？',
    '如何查询实时路况？'
  ]
}

// 🏫 Campus Facilities questions
if (lowerMessage.includes('图书馆') || lowerMessage.includes('体育馆') || lowerMessage.includes('设施') || 
    lowerMessage.includes('操场') || lowerMessage.includes('实验室')) {
  return [
    '图书馆的开放时间是什么时候？',
    '体育馆如何预约使用？',
    '实验室的使用规定有哪些？',
    '校园设施分布在哪里？'
  ]
}

// 🎉 Activities & Clubs questions
if (lowerMessage.includes('活动') || lowerMessage.includes('社团') || lowerMessage.includes('比赛') || 
    lowerMessage.includes('演出') || lowerMessage.includes('讲座')) {
  return [
    '近期有哪些校园活动？',
    '如何加入感兴趣的社团？',
    '活动报名流程是怎样的？',
    '讲座信息在哪里查看？'
  ]
}

// 📋 Administrative questions
if (lowerMessage.includes('证明') || lowerMessage.includes('手续') || lowerMessage.includes('办理') || 
    lowerMessage.includes('申请') || lowerMessage.includes('盖章')) {
  return [
    '学生证明如何办理？',
    '相关部门的办公时间是什么？',
    '需要携带哪些材料？',
    '办理进度如何查询？'
  ]
}

// 🏥 Health & Safety questions
if (lowerMessage.includes('医院') || lowerMessage.includes('医疗') || lowerMessage.includes('安全') || 
    lowerMessage.includes('急诊') || lowerMessage.includes('校医')) {
  return [
    '校医院的位置在哪里？',
    '急诊电话是多少？',
    '医保报销流程是怎样的？',
    '校园安全联系方式是什么？'
  ]
}

// 💳 Card & Payment questions
if (lowerMessage.includes('校园卡') || lowerMessage.includes('一卡通') || lowerMessage.includes('充值') || 
    lowerMessage.includes('支付') || lowerMessage.includes('缴费')) {
  return [
    '校园卡如何充值？',
    '一卡通丢失怎么办？',
    '学费缴纳方式有哪些？',
    '校园卡使用范围是什么？'
  ]
}

// 🌐 Network & IT questions
if (lowerMessage.includes('网络') || lowerMessage.includes('wifi') || lowerMessage.includes('账号') || 
    lowerMessage.includes('密码') || lowerMessage.includes('系统')) {
  return [
    '校园wifi如何连接？',
    '账号密码忘记怎么办？',
    '网络故障如何报修？',
    '有哪些常用的校园系统？'
  ]
}

// Default General campus questions
return [
  '校园地图在哪里可以查看？',
  '图书馆的开放时间是什么时候？',
  '如何联系相关部门？',
  '校园有哪些重要通知？'
]
}

const loadUserAvatar = () => {
  const user = JSON.parse(localStorage.getItem('user') || '{}')
  if (user.avatar) {
    userAvatar.value = user.avatar
  }
}

const listenForAvatarChanges = () => {
  window.addEventListener('storage', (e) => {
    if (e.key === 'user') {
      const user = JSON.parse(e.newValue || '{}')
      if (user.avatar) {
        userAvatar.value = user.avatar
      }
    }
  })
}

// Initialize the component
onMounted(() => {
  // Check if there's a chat ID in the URL
  const urlParams = new URLSearchParams(window.location.search)
  const chatId = urlParams.get('chat')
  if (chatId) {
    loadChat(chatId)
  }
  
  loadUserAvatar()
  listenForAvatarChanges()
  
  // Focus the input field when the component mounts
  document.querySelector('textarea')?.focus()
})
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(135deg, #e8fff2 0%, #c9f7d6 45%, #b5e8c2 100%);
  color: #2c3e50;
  overflow: hidden;
}

.chat-content {
  flex: 1;
  display: flex;
  height: 100%;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.welcome-screen {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 0;
  text-align: center;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 0;
  margin: 0;
  box-shadow: none;
}

.welcome-screen h1 {
  font-size: 2.5rem;
  margin-bottom: 1rem;
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 700;
}

.welcome-screen p {
  color: #7f8c8d;
  margin-bottom: 2rem;
  font-size: 1.1rem;
}

.new-chat-btn {
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  color: white;
  border: none;
  padding: 1rem 2rem;
  border-radius: 25px;
  cursor: pointer;
  font-size: 1.1rem;
  font-weight: 600;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(39, 174, 96, 0.28);
}

.new-chat-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(39, 174, 96, 0.36);
}

.chat-interface {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 0;
  margin: 0;
  box-shadow: none;
  overflow: hidden;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 2rem;
  scroll-behavior: smooth;
  background: transparent;
}

.message {
  display: flex;
  padding: 1rem;
  max-width: 48rem;
  margin: 0 auto 1rem auto;
  width: 100%;
  border-radius: 15px;
  transition: all 0.3s ease;
  align-items: flex-start;
}

.message-avatar {
  margin-right: 1rem;
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid #e9ecef;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
}

.message-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.message-content {
  flex: 1;
  padding-top: 0.25rem;
  min-width: 0;
}

.user-msg {
  background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
  color: white;
  margin-left: auto;
  max-width: 40rem;
  flex-direction: row-reverse;
}

.user-msg .message-avatar {
  margin-right: 0;
  margin-left: 1rem;
  border-color: rgba(255, 255, 255, 0.3);
}

.ai-msg {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  color: #2c3e50;
  margin-right: auto;
  max-width: 40rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.message-text {
  line-height: 1.6;
  white-space: pre-wrap;
  color: inherit;
}

.suggested-questions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 1rem;
}

.suggested-question {
  background: linear-gradient(45deg, #f8f9fa, #e9ecef);
  border: 1px solid #dee2e6;
  border-radius: 20px;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.3s ease;
  color: #495057;
  font-weight: 500;
}

.suggested-question:hover {
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  color: white;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(39, 174, 96, 0.26);
}

.input-container {
  padding: 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(20px);
}

.input-wrapper {
  position: relative;
  max-width: 48rem;
  margin: 0 auto;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 25px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s ease;
}

textarea {
  width: 100%;
  min-height: 3rem;
  max-height: 200px;
  padding: 1rem 3.5rem 1rem 1.5rem;
  border: none;
  border-radius: 25px;
  background: transparent;
  color: #2c3e50;
  font-size: 1rem;
  line-height: 1.5;
  resize: none;
  outline: none;
  transition: all 0.3s ease;
}

.input-wrapper:focus-within {
  border-color: rgba(39, 174, 96, 0.55);
  box-shadow: 0 0 0 3px rgba(39, 174, 96, 0.14), 0 4px 20px rgba(0, 0, 0, 0.15);
}

.send-button {
  position: absolute;
  right: 0.5rem;
  bottom: 0.5rem;
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  color: white;
  border: none;
  border-radius: 50%;
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(39, 174, 96, 0.26);
}

.send-button:not(:disabled):hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(39, 174, 96, 0.34);
}

.send-button:disabled {
  background: rgba(255, 255, 255, 0.5);
  cursor: not-allowed;
  opacity: 0.6;
  box-shadow: none;
}

.disclaimer {
  color: rgba(44, 62, 80, 0.65);
  font-size: 0.8rem;
  text-align: center;
  margin-top: 1rem;
  font-style: italic;
}

.typing-indicator {
  display: flex;
  justify-content: center;
  padding: 1rem 0;
  background: #f8f9fa;
  border-radius: 15px;
  margin: 1rem auto;
  max-width: 100px;
}

.typing-indicator span {
  display: inline-block;
  width: 8px;
  height: 8px;
  margin: 0 2px;
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  border-radius: 50%;
  animation: typing 1s infinite ease-in-out;
}

.typing-indicator span:nth-child(2) {
  animation-delay: 0.2s;
}

.typing-indicator span:last-child {
  animation-delay: 0.4s;
}

@keyframes typing {
  0%, 60%, 100% {
    transform: translateY(0);
  }
  30% {
    transform: translateY(-4px);
  }
}

.loading-dots {
  display: inline-flex;
  align-items: center;
  height: 100%;
  font-size: 1.2rem;
  line-height: 1;
  letter-spacing: 0.2em;
  color: white;
}

/* Custom scrollbar */
::-webkit-scrollbar {
  width: 8px;
}

::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}

::-webkit-scrollbar-thumb {
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  border-radius: 10px;
}
</style>
