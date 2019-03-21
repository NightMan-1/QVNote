import Vue from 'vue'
import Router from 'vue-router'
import qvIndex from '@/components/qvIndex'
import qvApp from '@/components/qvApp'
const qvInstaller = () => import('@/components/qvInstaller')
const qvErrorFatal = () => import('@/components/qvErrorFatal')
const qvNotFound = () => import('@/components/qv404')
const qvShutdown = () => import('@/components/qvShutdown')

export default new Router({
  // mode: 'history', // https://router.vuejs.org/ru/guide/essentials/history-mode.html
  routes: [
    {
      path: '/',
      name: 'qvIndex',
      component: qvIndex
    },
    {
      path: '/app',
      name: 'qvApp',
      component: qvApp
    },
    {
      path: '/settings',
      name: 'qvSettings',
      component: qvApp
    },
    {
      path: '/shutdown',
      name: 'qvShutdown',
      component: qvShutdown
    },
    {
      path: '/notes',
      name: 'qvNotes',
      component: qvApp,
      children: [
        {
          path: ':nbUUID',
          name: 'qvNotebooks',
          component: qvApp,
          children: [
            {
              path: ':noteUUID',
              name: 'qvNote',
              component: qvApp
            }
          ]
        }
      ]
    },
    {
      path: '/tags',
      name: 'qvTags',
      component: qvApp,
      children: [
        {
          path: ':nbUUID',
          name: 'qvTagsList',
          component: qvApp,
          children: [
            {
              path: ':noteUUID',
              name: 'qvTag',
              component: qvApp
            }
          ]
        }
      ]
    },
    {
      path: '/install',
      name: 'qvInstaller',
      component: qvInstaller
    },
    {
      path: '/editor',
      name: 'qvEditor',
      component: qvApp
    },
    {
      path: '/error',
      name: 'qvErrorFatal',
      component: qvErrorFatal
    },
    {
      path: '/404',
      name: '404',
      component: qvNotFound
    },
    {
      path: '*',
      redirect: '/404'
    }
  ]
})

Vue.use(Router)
