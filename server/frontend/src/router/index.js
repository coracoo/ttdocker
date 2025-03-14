import { createRouter, createWebHistory } from 'vue-router'
import TemplateList from '../views/TemplateList.vue'

const routes = [
  {
    path: '/',
    name: 'TemplateList',
    component: TemplateList
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router