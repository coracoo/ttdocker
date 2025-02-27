import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import Overview from '../views/Overview.vue'
import Docker from '../views/Docker.vue'  // 修改这里
import Images from '../views/Images.vue'
import Volumes from '../views/Volumes.vue'
import Networks from '../views/Networks.vue'
import AppStore from '../views/AppStore.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
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
          path: 'containers',
          component: Docker  // 修改这里
        },
        {
          path: 'app-store',
          component: AppStore
        },
        {
          path: 'images',
          component: Images
        },
        {
          path: 'volumes',
          component: Volumes
        },
        {
          path: 'networks',
          component: Networks
        }
      ]
    }
  ]
})

export default router