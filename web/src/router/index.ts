import { createRouter, createWebHistory } from 'vue-router'
import App from '../App.vue'
import SetupGuide from '../components/SetupGuide.vue'

const routes = [
  {
    path: '/',
    name: 'dashboard',
    component: App,
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
