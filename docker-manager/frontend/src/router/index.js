import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import Overview from '../views/Overview.vue'
import AppStore from '../views/AppStore.vue'

const routes = [
  {
    path: '/',
    component: MainLayout,
    children: [
      {
        path: '',
        redirect: '/overview'
      },
      {
        path: 'overview',
        component: Overview
      },
      {
        path: 'app-store',
        component: AppStore
      }
    ]
  }
]

export default createRouter({
  history: createWebHistory(),
  routes
})