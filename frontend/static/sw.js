// sw.js — Service Worker for Web Push

self.addEventListener('push', e => {
  const data = e.data?.json() ?? {};
  e.waitUntil(
    self.registration.showNotification(data.title ?? 'シフト管理', {
      body:  data.body  ?? '',
      icon:  '/static/icon-192.png',
      badge: '/static/icon-192.png',
      data:  { url: data.url ?? '/' },
      tag:   'shift-push',           // 同じタグは重複表示しない
      renotify: true,
    })
  );
});

self.addEventListener('notificationclick', e => {
  e.notification.close();
  e.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then(list => {
      for (const c of list) {
        if (c.url.startsWith(self.registration.scope) && 'focus' in c) {
          return c.focus();
        }
      }
      return clients.openWindow(e.notification.data?.url ?? '/');
    })
  );
});
