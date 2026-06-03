import { createRouter, createWebHistory, createWebHashHistory } from 'vue-router'
import { routes } from './routes'
import { installGuards } from './guard'

const isDemo = import.meta.env.VITE_DEMO_MOCK === 'true'

export const router = createRouter({
  history: isDemo
    ? createWebHashHistory()
    : createWebHistory('/admin/'),
  routes,
})

installGuards(router)
