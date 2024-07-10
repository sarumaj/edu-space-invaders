self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open("v1").then((cache) => {
      // Pre-cache some static assets
      return cache.addAll([
        "/",
        "/favicon.ico",
        "/index.html",
        "/manifest.json",
        "/style.css",
        "/wasm.js",
        "/icons/icon-192x192.png",
        "/icons/icon-512-512.png",
      ]);
    })
  );
});

self.addEventListener("fetch", (event) => {
  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      if (cachedResponse) {
        // Extract the ETag from the cached response
        const etag = cachedResponse.headers.get("ETag");

        // Make a conditional request to the server using the ETag
        const headers = new Headers();
        if (etag) {
          headers.append("If-None-Match", etag);
        }

        return fetch(event.request, { headers }).then((networkResponse) => {
          // If the response status is 304 (Not Modified), return the cached response
          if (networkResponse.status === 304) {
            return cachedResponse;
          }

          // Otherwise, update the cache with the new response and return it
          return caches.open("v1").then((cache) => {
            cache.put(event.request, networkResponse.clone());
            return networkResponse;
          });
        });
      }

      // If there's no cached response, fetch from the network and cache it
      return fetch(event.request).then((networkResponse) => {
        return caches.open("v1").then((cache) => {
          cache.put(event.request, networkResponse.clone());
          return networkResponse;
        });
      });
    })
  );
});
