import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'
import Overview from '../views/Overview.vue'
import Docker from '../views/Docker.vue'
import Images from '../views/Images.vue'
import Volumes from '../views/Volumes.vue'
import Networks from '../views/Networks.vue'
import AppStore from '../views/AppStore.vue'
import Navigation from '../views/Navigation.vue'
import Projects from '../views/Projects.vue'
import ProjectDetail from '../views/ProjectDetail.vue'
import DockerDetail from '../views/DockerDetail.vue'

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
          component: Docker
        },
        {
          path: 'containers/:name',
          component: DockerDetail
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
        },
        {
          path: 'navigation',
          component: Navigation
        },
        {
          path: 'projects',
          component: Projects
        },
        {
          path: 'projects/:name',
          component: ProjectDetail
        }
      ]
    }
  ]
})

export default router