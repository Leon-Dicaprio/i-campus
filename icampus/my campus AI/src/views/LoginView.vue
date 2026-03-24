<template>
  <div class="login-page">
    <div class="card">
      <h2>🏫 莞工校园AI助手</h2>
      <p>{{ isRegister ? '注册新账号' : '学号/手机号登录' }}</p>
      <form @submit.prevent="handleSubmit">
        <input v-model="user" placeholder="用户名" required />
        <input v-model="pwd" type="password" placeholder="密码" required />
        <button type="submit" :disabled="loading">
          {{ loading ? (isRegister ? '注册中...' : '登录中...') : (isRegister ? '注册' : '登录') }}
        </button>
      </form>
      <div class="switch-mode">
        <button type="button" @click="isRegister = !isRegister" class="link-btn">
          {{ isRegister ? '已有账号？去登录' : '没有账号？去注册' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '@/services/api'

const user = ref('')
const pwd = ref('')
const loading = ref(false)
const router = useRouter()
const isRegister = ref(false)

const handleSubmit = async () => {
  if (!user.value || !pwd.value) return
  loading.value = true
  
  try {
    if (isRegister.value) {
      await authAPI.register(user.value, pwd.value)
      alert('注册成功，请登录')
      isRegister.value = false
    } else {
      await authAPI.login(user.value, pwd.value)
      localStorage.setItem('campus_ai_username', user.value)
      localStorage.setItem('campus_ai_token', 'logged_in')
      router.push('/chat')
    }
  } catch (error) {
    const message = error.response?.data?.error || error.message || '操作失败'
    alert(message)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  background: 
    linear-gradient(rgba(255,255,255,0.6), rgba(255,255,255,0.6)),
    url('https://lh3.googleusercontent.com/gps-cs-s/AHVAwepOTCdNwCa2Yo8fjzZ4KVmshbZw_XGpD94RrtL07HJKrpWhHJv-ncvUnOpsLZwoZJCPRrGqWWgSNdPyBZtn4aqOdUsPurj5T4wqMiZ6CL99ZU5rp7X6LpW7r68WYYb-Uw4vD3k5Pg=s680-w680-h510-rw');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
}
.card {
  background: white;
  padding: 32px;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.1);
  width: 90%;
  max-width: 360px;
  text-align: center;
}
.card h2 {
  color: #1a5fb4;
  margin-bottom: 8px;
}
.card p {
  color: #666;
  margin-bottom: 24px;
}
input {
  width: 100%;
  padding: 12px;
  margin: 8px 0;
  border: 1px solid #ccc;
  border-radius: 6px;
  font-size: 16px;
}
button {
  width: 100%;
  padding: 12px;
  background: #1a5fb4;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  margin-top: 16px;
  cursor: pointer;
}
button:disabled {
  background: #b0c4db;
}
.switch-mode {
  margin-top: 16px;
}
.link-btn {
  background: none;
  border: none;
  color: #1a5fb4;
  text-decoration: underline;
  cursor: pointer;
  font-size: 14px;
  padding: 0;
  margin: 0;
}
.link-btn:hover {
  color: #0d47a1;
}
</style>
