import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { me } from './api/client'
import App from './App.vue'
import './style.css'

// Apply saved theme, defaulting to dark mode.
const theme = localStorage.getItem('theme') ?? 'dark'
localStorage.setItem('theme', theme)
document.documentElement.classList.toggle('dark', theme === 'dark')

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('./views/Login.vue'), meta: { public: true } },
    { path: '/change-password', component: () => import('./views/ChangePassword.vue'), meta: { public: true } },
    { path: '/', redirect: '/dashboard' },
    { path: '/dashboard', component: () => import('./views/Dashboard.vue') },
    { path: '/campaigns', component: () => import('./views/Campaigns.vue') },
    { path: '/campaigns/new', component: () => import('./views/CampaignWizard.vue') },
    { path: '/campaigns/:id', component: () => import('./views/CampaignDetail.vue') },
    { path: '/targets', component: () => import('./views/Targets.vue') },
    { path: '/targets/:id', component: () => import('./views/TargetListDetail.vue') },
    { path: '/templates', component: () => import('./views/Templates.vue') },
    { path: '/phishlets', component: () => import('./views/Phishlets.vue') },
    { path: '/smtp', component: () => import('./views/SMTP.vue') },
    { path: '/miraged', component: () => import('./views/Miraged.vue') },
    { path: '/miraged/:id', component: () => import('./views/MiragedDetail.vue') },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.public) return true

  try {
    const user = await me()
    if (user.password_change_required && to.path !== '/change-password') {
      return '/change-password'
    }
    return true
  } catch {
    return '/login'
  }
})

createApp(App).use(router).mount('#app')
