import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './style.css'

// Apply saved theme, defaulting to dark mode.
const theme = localStorage.getItem('theme') ?? 'dark'
localStorage.setItem('theme', theme)
document.documentElement.classList.toggle('dark', theme === 'dark')

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/campaigns' },
    { path: '/campaigns', component: () => import('./views/Campaigns.vue') },
    { path: '/campaigns/:id', component: () => import('./views/CampaignDetail.vue') },
    { path: '/targets', component: () => import('./views/Targets.vue') },
    { path: '/targets/:id', component: () => import('./views/TargetListDetail.vue') },
    { path: '/templates', component: () => import('./views/Templates.vue') },
    { path: '/phishlets', component: () => import('./views/Phishlets.vue') },
    { path: '/smtp', component: () => import('./views/SMTP.vue') },
    { path: '/settings', component: () => import('./views/Settings.vue') },
  ],
})

createApp(App).use(router).mount('#app')
