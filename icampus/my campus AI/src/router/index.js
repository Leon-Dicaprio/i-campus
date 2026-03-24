import { createRouter, createWebHashHistory } from 'vue-router'
import LoginView from '@/views/LoginView.vue'
import MainLayout from '@/layouts/MainLayout.vue'
import ChatView from '@/views/ChatView.vue'
import ProfileView from '@/views/ProfileView.vue'
import StudentHandbookView from '@/views/StudentHandbookView.vue'
import SchoolMapView from '@/views/SchoolMapView.vue'

const routes = [
  { path: '/login', component: LoginView },
  {
    path: '/',
    component: MainLayout,
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: 'chat' },
      { path: 'chat', component: ChatView },
      { path: 'profile', component: ProfileView },
      { path: 'student-handbook', component: StudentHandbookView },
      { path: 'school-map', component: SchoolMapView }
    ]
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, _, next) => {
  const token = localStorage.getItem('campus_ai_token')
  const username = localStorage.getItem('campus_ai_username')
  const requiresAuth = to.matched.some(r => r.meta?.requiresAuth)
  if (requiresAuth && (!token || !username)) {
    next('/login')
  } else if (to.path === '/login' && token && username) {
    next('/')
  } else {
    next()
  }
})

export default router
