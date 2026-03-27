import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './style.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/campaigns' },
    { path: '/campaigns', component: () => import('./views/Campaigns.vue') },
    { path: '/campaigns/:id', component: () => import('./views/CampaignDetail.vue') },
    { path: '/targets', component: () => import('./views/Targets.vue') },
    { path: '/templates', component: () => import('./views/Templates.vue') },
    { path: '/smtp', component: () => import('./views/SMTP.vue') },
    { path: '/settings', component: () => import('./views/Settings.vue') },
  ],
})

createApp(App).use(router).mount('#app')
