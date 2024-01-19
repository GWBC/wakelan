import { createRouter, createWebHistory } from 'vue-router'
import WakeLan from '../views/WakeLan.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: WakeLan
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue')
    },
    {
      path: '/log',
      name: 'log',
      component: () => import('../views/Log.vue')
    },
    {
      path: '/config',
      name: 'config',
      component: () => import('../views/Config.vue')
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/Login.vue')
    },
    {
      path: '/filetransfer',
      name: 'filetransfer',
      component: () => import('../views/FileTransfer.vue')
    }
  ]
})

export default router
