// GoBE Service Worker for caching
const CACHE_NAME = 'gobe-dashboard-v1.3.4';
const urlsToCache = [
    '/',
    '/index.html',
    '/style.css',
    '/app.js',
    '/health'
];

// Install event - cache static assets
self.addEventListener('install', function(event) {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then(function(cache) {
                console.log('📦 Cache opened:', CACHE_NAME);
                return cache.addAll(urlsToCache);
            })
            .catch(function(error) {
                console.error('📦 Cache install failed:', error);
            })
    );
});

// Fetch event - serve from cache when offline
self.addEventListener('fetch', function(event) {
    // Only cache GET requests
    if (event.request.method !== 'GET') {
        return;
    }

    // Skip MCP and API requests (always fetch fresh)
    if (event.request.url.includes('/mcp/') ||
        event.request.url.includes('/api/')) {
        return;
    }

    event.respondWith(
        caches.match(event.request)
            .then(function(response) {
                // Return cached version or fetch from network
                return response || fetch(event.request);
            })
            .catch(function(error) {
                console.error('📦 Fetch failed:', error);
                // Return a fallback page or error response
                if (event.request.destination === 'document') {
                    return caches.match('/index.html');
                }
            })
    );
});

// Activate event - clean up old caches
self.addEventListener('activate', function(event) {
    event.waitUntil(
        caches.keys().then(function(cacheNames) {
            return Promise.all(
                cacheNames.map(function(cacheName) {
                    if (cacheName !== CACHE_NAME) {
                        console.log('📦 Deleting old cache:', cacheName);
                        return caches.delete(cacheName);
                    }
                })
            );
        })
    );
});