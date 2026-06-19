import { createRouter, createWebHistory } from 'vue-router'

const qvApp = () => import('@/components/qvApp.vue')
const qvNotFound = () => import('@/components/qv404.vue')
const qvShutdown = () => import('@/components/qvShutdown.vue')
const qvInstaller = () => import('@/components/qvInstaller.vue')
const qvErrorFatal = () => import('@/components/qvErrorFatal.vue')
const qvSettings = () => import('@/components/qvSettings.vue')

const guardDirectAccess = (_to, from) => {
    if (!from.name) {
        return '/'
    }
}

const router = createRouter({
    history: createWebHistory(), // https://router.vuejs.org/ru/guide/essentials/history-mode.html
    routes: [
        {
            path: '/',
            name: 'qvApp',
            component: qvApp
        },
        {
            path: '/notes/',
            name: 'qvNotes',
            component: qvApp,
            redirect: '/',
            children: [
                {
                    path: ':nbUUID/',
                    name: 'qvNotebooks',
                    component: qvApp,
                    children: [
                        {
                            path: ':noteUUID/',
                            name: 'qvNote',
                            component: qvApp
                        }
                    ]
                }
            ]
        },
        {
            path: '/tags/',
            name: 'qvTags',
            component: qvApp,
            redirect: '/',
            children: [
                {
                    path: ':nbUUID/',
                    name: 'qvTagsList',
                    component: qvApp,
                    children: [
                        {
                            path: ':noteUUID/',
                            name: 'qvTag',
                            component: qvApp
                        }
                    ]
                }
            ]
        },
        {
            path: '/settings/',
            name: 'qvSettings',
            component: qvSettings,
            beforeEnter: guardDirectAccess
        },
        {
            path: '/editor/',
            name: 'qvEditor',
            component: qvApp,
            beforeEnter: guardDirectAccess
        },
        {
            path: '/install/',
            name: 'qvInstaller',
            component: qvInstaller
        },
        {
            path: '/error/',
            name: 'qvErrorFatal',
            component: qvErrorFatal,
            beforeEnter: guardDirectAccess
        },
        {
            path: '/shutdown/',
            name: 'qvShutdown',
            component: qvShutdown,
            beforeEnter: guardDirectAccess
        },
        {
            path: '/error404/',
            name: 'qvError404',
            component: qvNotFound
        },
        {
            path: '/:pathMatch(.*)*',
            name: 'qvNotFound',
            component: qvNotFound
        }

    ]
})

export default router
