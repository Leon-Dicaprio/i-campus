<template>
  <div class="profile-view">
    <div class="profile-content">
      <div class="profile-toolbar">
        <button type="button" class="back-btn" @click="goBack">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
            <path d="M19 12H5M12 19L5 12L12 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          返回聊天
        </button>
        <h1>个人设置</h1>
      </div>

      <div class="profile-section">
        <h2>用户头像</h2>
        <div class="avatar-section">
          <div class="current-avatar">
            <img :src="formData.avatar" alt="用户头像" />
          </div>
          <div class="avatar-options">
            <div class="file-upload">
              <input 
                type="file" 
                id="avatar-upload" 
                @change="handleAvatarChange" 
                accept="image/*"
                style="display: none"
              >
              <label for="avatar-upload" class="upload-btn">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                  <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                上传头像
              </label>
            </div>
          </div>
        </div>
      </div>

      <div class="profile-section">
        <h2>基本信息</h2>
        <div class="form-group">
          <label for="username">用户昵称</label>
          <input 
            type="text" 
            id="username" 
            v-model="formData.username"
            placeholder="请输入用户昵称"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label for="email">邮箱地址</label>
          <input 
            type="email" 
            id="email" 
            v-model="formData.email"
            placeholder="请输入邮箱地址"
            class="form-input"
          >
        </div>
      </div>

      <div class="profile-section">
        <h2>修改密码</h2>
        <div class="form-group">
          <label for="current-password">当前密码</label>
          <input 
            type="password" 
            id="current-password" 
            v-model="formData.currentPassword"
            placeholder="请输入当前密码"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label for="new-password">新密码</label>
          <input 
            type="password" 
            id="new-password" 
            v-model="formData.newPassword"
            placeholder="请输入新密码"
            class="form-input"
          >
        </div>
        <div class="form-group">
          <label for="confirm-password">确认新密码</label>
          <input 
            type="password" 
            id="confirm-password" 
            v-model="formData.confirmPassword"
            placeholder="请再次输入新密码"
            class="form-input"
          >
        </div>
      </div>

      <div class="action-buttons">
        <button @click="saveProfile" class="save-btn" :disabled="isSaving">
          <span v-if="!isSaving">保存设置</span>
          <span v-else>保存中...</span>
        </button>
        <button @click="logout" class="logout-btn">退出登录</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const goBack = () => {
  router.push('/chat')
}

const formData = ref({
  username: '',
  email: '',
  avatar: '/images/user-avatar.png',
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const isSaving = ref(false)

const handleAvatarChange = (event) => {
  const file = event.target.files[0]
  if (file && file.type.startsWith('image/')) {
    const reader = new FileReader()
    reader.onload = (e) => {
      formData.value.avatar = e.target.result
    }
    reader.readAsDataURL(file)
  }
}

const saveProfile = async () => {
  if (isSaving.value) return
  
  // 验证密码修改
  if (formData.value.newPassword) {
    if (formData.value.newPassword !== formData.value.confirmPassword) {
      alert('新密码和确认密码不一致')
      return
    }
    if (!formData.value.currentPassword) {
      alert('请输入当前密码')
      return
    }
  }
  
  isSaving.value = true
  
  try {
    // 保存用户信息到localStorage
    const userData = {
      name: formData.value.username,
      email: formData.value.email,
      avatar: formData.value.avatar,
      updatedAt: new Date().toISOString()
    }
    
    localStorage.setItem('user', JSON.stringify(userData))
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    // 清空密码字段
    formData.value.currentPassword = ''
    formData.value.newPassword = ''
    formData.value.confirmPassword = ''
    
    alert('设置保存成功！')
  } catch (error) {
    console.error('保存失败:', error)
    alert('保存失败，请重试')
  } finally {
    isSaving.value = false
  }
}

const logout = () => {
  localStorage.clear()
  router.push('/login')
}

const loadUserProfile = () => {
  const user = JSON.parse(localStorage.getItem('user') || '{}')
  if (user) {
    formData.value.username = user.name || '用户'
    formData.value.email = user.email || 'user@example.com'
    formData.value.avatar = user.avatar || '/images/user-avatar.png'
  }
}

onMounted(() => {
  loadUserProfile()
})
</script>

<style scoped>
.profile-view {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #e8fff2 0%, #c9f7d6 45%, #b5e8c2 100%);
  padding: 24px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.profile-content {
  width: 100%;
  max-width: 720px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.profile-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(10px);
  box-shadow: 0 6px 20px rgba(15, 81, 50, 0.08);
  border: 1px solid rgba(15, 81, 50, 0.08);
}

.profile-toolbar h1 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 800;
  color: #0f5132;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 0.9rem;
  border-radius: 12px;
  border: 1px solid rgba(15, 81, 50, 0.18);
  background: rgba(255, 255, 255, 0.85);
  color: #0f5132;
  cursor: pointer;
  font-weight: 800;
  transition: transform 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
}

.back-btn:hover {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 6px 16px rgba(15, 81, 50, 0.12);
}

.profile-section {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 15px;
  padding: 22px;
  box-shadow: 0 10px 30px rgba(15, 81, 50, 0.10);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(15, 81, 50, 0.08);
}

.profile-section h2 {
  margin: 0 0 1.5rem 0;
  color: #2c3e50;
  font-size: 1.25rem;
  font-weight: 600;
  border-bottom: 2px solid rgba(39, 174, 96, 0.35);
  padding-bottom: 0.5rem;
}

.avatar-section {
  display: flex;
  gap: 2rem;
  align-items: flex-start;
}

.current-avatar {
  flex-shrink: 0;
}

.current-avatar img {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  object-fit: cover;
  border: 4px solid rgba(39, 174, 96, 0.35);
  box-shadow: 0 10px 30px rgba(15, 81, 50, 0.18);
}

.avatar-options {
  flex: 1;
}

.upload-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 10px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.3s ease;
  box-shadow: 0 6px 18px rgba(39, 174, 96, 0.26);
  margin-bottom: 1rem;
}

.upload-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 22px rgba(39, 174, 96, 0.34);
}

@media (max-width: 640px) {
  .avatar-section {
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }

  .avatar-options {
    width: 100%;
  }

  .upload-btn {
    width: 100%;
    justify-content: center;
  }

  .current-avatar img {
    width: 96px;
    height: 96px;
  }
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #2c3e50;
  font-weight: 600;
  font-size: 0.9rem;
}

.form-input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 2px solid #e9ecef;
  border-radius: 10px;
  font-size: 1rem;
  transition: all 0.3s ease;
  background: white;
  color: #2c3e50;
}

.form-input:focus {
  outline: none;
  border-color: rgba(39, 174, 96, 0.65);
  box-shadow: 0 0 0 3px rgba(39, 174, 96, 0.14);
}

.action-buttons {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
}

.save-btn {
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  color: white;
  border: none;
  padding: 0.75rem 2rem;
  border-radius: 10px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.3s ease;
  box-shadow: 0 10px 22px rgba(39, 174, 96, 0.28);
}

.save-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 14px 26px rgba(39, 174, 96, 0.34);
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.logout-btn {
  background: #e74c3c;
  color: white;
  border: none;
  padding: 0.75rem 2rem;
  border-radius: 10px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.3s ease;
}

.logout-btn:hover {
  background: #c0392b;
  transform: translateY(-1px);
}

@media (max-width: 640px) {
  .profile-view {
    padding: 16px;
  }

  .profile-toolbar {
    padding: 12px 12px;
  }

  .profile-toolbar h1 {
    font-size: 1.05rem;
  }

  .profile-section {
    padding: 16px;
  }

  .action-buttons {
    flex-direction: column;
    align-items: stretch;
  }

  .save-btn,
  .logout-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
