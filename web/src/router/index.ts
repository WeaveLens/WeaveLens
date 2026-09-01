import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../components/Dashboard.vue'
import SetupGuide from '../components/SetupGuide.vue'

const routes = [
  {
    path: '/',
    name: 'dashboard',
    component: Dashboard,
  },
  {
    path: '/setup/aws',
    name: 'setup-aws',
    component: SetupGuide,
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
