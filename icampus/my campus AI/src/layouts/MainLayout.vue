<template>
  <div class="main-layout">
    <div class="app-nav">
      <div class="nav-brand">
        <router-link to="/chat" class="brand-link">
          <h2>校园AI助手</h2>
        </router-link>
      </div>
      <div class="nav-links">
        <router-link to="/chat" class="nav-link">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M21 15C21 15.5304 20.7893 16.0391 20.4142 16.4142C20.0391 16.7893 19.5304 17 19 17H7L3 21V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H19C19.5304 3 20.0391 3.21071 20.4142 3.58579C20.7893 3.96086 21 4.46957 21 5V15Z"
                  stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          AI对话
        </router-link>

        <router-link to="/student-handbook" class="nav-link">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M4 19.5C4 18.837 4.26339 18.2011 4.73223 17.7322C5.20107 17.2634 5.83696 17 6.5 17H20"
                  stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M6.5 2H20V17H6.5C5.83696 17 5.20107 17.2634 4.73223 17.7322C4.26339 18.2011 4 18.837 4 19.5V6.5C4 5.83696 4.26339 5.20107 4.73223 4.73223C5.20107 4.26339 5.83696 4 6.5 4H17.5L20 6.5V2Z"
                  stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          学生手册
        </router-link>

        <router-link to="/school-map" class="nav-link">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M9 20L3 16V6L9 2L15 6V16L9 20Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M9 2V20" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          校园导航
        </router-link>

        <router-link to="/profile" class="nav-link">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M20 21C20 21 19 17 12 17C5 17 4 21 4 21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M12 13C14.2091 13 16 11.2091 16 9C16 6.79086 14.2091 5 12 5C9.79086 5 8 6.79086 8 9C8 11.2091 9.79086 13 12 13Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          个人
        </router-link>
      </div>
    </div>

    <div class="layout-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
</script>

<style scoped>
.main-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.layout-content {
  flex: 1;
  min-height: 0;
}

.app-nav {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  padding: 1rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.brand-link {
  text-decoration: none;
}

.nav-brand h2 {
  margin: 0;
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 700;
  font-size: 1.5rem;
}

.nav-links {
  display: flex;
  gap: 1rem;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  border-radius: 12px;
  text-decoration: none;
  color: #6c757d;
  font-weight: 500;
  transition: all 0.3s ease;
  border: 1px solid transparent;
}

.nav-link:hover {
  background: rgba(39, 174, 96, 0.12);
  color: #27ae60;
  transform: translateY(-1px);
}

.nav-link.router-link-exact-active {
  background: linear-gradient(45deg, #2ecc71, #27ae60);
  color: white;
  box-shadow: 0 4px 15px rgba(39, 174, 96, 0.28);
}

@media (max-width: 900px) {
  .app-nav {
    padding: 0.75rem 1rem;
  }

  .nav-links {
    gap: 0.5rem;
  }

  .nav-link {
    padding: 0.6rem 0.8rem;
    font-size: 0.9rem;
  }

  .nav-link svg {
    width: 18px;
    height: 18px;
  }
}

@media (max-width: 720px) {
  .nav-link {
    padding: 0.55rem 0.65rem;
  }

  .nav-link svg {
    display: none;
  }
}
</style>
