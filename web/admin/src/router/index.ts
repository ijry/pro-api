import { createRouter, createWebHistory } from 'vue-router'
import { routes } from './routes'
import { installGuards } from './guard'

export const router = createRouter({
  history: createWebHistory('/admin/'),
  routes,
})

installGuards(router)
