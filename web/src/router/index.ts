import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/',         redirect: '/lines' },
    { path: '/lines',    component: () => import('@/views/Lines.vue')    },
    { path: '/signals',  component: () => import('@/views/Signals.vue')  },
    { path: '/monitor',  component: () => import('@/views/Monitor.vue')  },
    { path: '/commands', component: () => import('@/views/Commands.vue') },
  ]
})

export default router
