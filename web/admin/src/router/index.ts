import { createRouter, createWebHistory } from 'vue-router'
import { routes } from './routes'
import { installGuards } from './guard'

const baseUrl = import.meta.env.VITE_DEMO_MOCK === 'true'
  ? import.meta.env.BASE_URL
  : '/admin/'

export const router = createRouter({
  history: createWebHistory(baseUrl),
  routes,
})

installGuards(router)
