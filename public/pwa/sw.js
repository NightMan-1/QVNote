/* QVNote service worker: caches the app shell and static assets so the PWA
 * starts instantly and opens offline. Note data still comes from the server;
 * offline writes are not supported. */

const CACHE = 'qvnote-static-v4'

// Cache the app shell and the entry bundles referenced by index.html, so the
// first offline reload can boot the SPA (lazy chunks and note data still need
// the server).
async function precacheShell () {
    const cache = await caches.open(CACHE)
    const response = await fetch('/')
    if (!response.ok) {
        return
    }
    const html = await response.clone().text()
    await cache.put('/', response)

    const assets = new Set()
    for (const match of html.matchAll(/(?:src|href)="(\/static\/[^"]+)"/g)) {
        assets.add(match[1])
    }

    // Lazy chunks (qvApp, qvEditor, ...) are not referenced by index.html, so
    // the server exposes the full list of static assets for precaching.
    try {
        const response = await fetch('/pwa/precache.json')
        if (response.ok) {
            for (const url of await response.json()) {
                assets.add(url)
            }
        }
    } catch (e) {
        // offline install is impossible anyway; index assets are enough
    }

    await Promise.allSettled([...assets].map(url => cache.add(url)))
}

self.addEventListener('install', event => {
    self.skipWaiting()
    event.waitUntil(precacheShell().catch(() => {}))
})

self.addEventListener('activate', event => {
    event.waitUntil(
        caches.keys()
            .then(keys => Promise.all(keys.filter(key => key !== CACHE).map(key => caches.delete(key))))
            .then(() => self.clients.claim())
    )
})

self.addEventListener('fetch', event => {
    const request = event.request
    if (request.method !== 'GET') {
        return
    }

    const url = new URL(request.url)
    if (url.origin !== self.location.origin) {
        return
    }

    // Hashed bundles, fonts and icons: serve from cache, refresh in background.
    if (url.pathname.startsWith('/static/') || url.pathname.startsWith('/pwa/')) {
        event.respondWith(
            caches.open(CACHE).then(cache => cache.match(request, { ignoreVary: true }).then(cached => {
                const network = fetch(request).then(response => {
                    if (response && response.ok) {
                        cache.put(request, response.clone())
                    }
                    return response
                }).catch(() => cached)
                return cached || network
            }))
        )
        return
    }

    // Pages: network first, fall back to the cached shell when offline.
    if (request.mode === 'navigate') {
        event.respondWith(fetch(request).catch(() => caches.match('/', { ignoreVary: true })))
    }
})
