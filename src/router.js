import Vue from 'vue'
import Router from 'vue-router'
const qvApp = () => import('@/components/qvApp')
const qvNotFound = () => import('@/components/qv404')
const qvShutdown = () => import('@/components/qvShutdown')
const qvInstaller = () => import('@/components/qvInstaller')
const qvErrorFatal = () => import('@/components/qvErrorFatal')
const qvSettings = () => import('@/components/qvSettings')

export default new Router({
    mode: 'history', // https://router.vuejs.org/ru/guide/essentials/history-mode.html
    routes: [
        {
            path: '/',
            name: 'qvApp',
            component: qvApp
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
            path: '/settings',
            name: 'qvSettings',
            component: qvSettings
        },
        {
            path: '/editor',
            name: 'qvEditor',
            component: qvApp
        },
        {
            path: '/install',
            name: 'qvInstaller',
            component: qvInstaller
        },
        {
            path: '/error',
            name: 'qvErrorFatal',
            component: qvErrorFatal
        },
        {
            path: '/shutdown',
            name: 'qvShutdown',
            component: qvShutdown
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
