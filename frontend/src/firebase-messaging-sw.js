importScripts(
  'https://www.gstatic.com/firebasejs/10.12.0/firebase-app-compat.js',
  'https://www.gstatic.com/firebasejs/10.12.0/firebase-messaging-compat.js'
);

firebase.initializeApp({
  projectId: "ceibuenvivir",
  appId: "1:429506836254:web:45f5eb50d86f074323bb0f",
  storageBucket: "ceibuenvivir.firebasestorage.app",
  apiKey: "AIzaSyDMdBkILiOSgqTq5ZD2VmrsPtYm7aU5bzc",
  authDomain: "ceibuenvivir.firebaseapp.com",
  messagingSenderId: "429506836254",
  measurementId: "G-4C2KCQ3MNT"
});

const messaging = firebase.messaging();

messaging.onBackgroundMessage((payload) => {
  console.log('[firebase-messaging-sw.js] Background message received:', payload);
  const notification = payload.notification || payload.data || {};
  const title = notification.title || 'Buen Vivir';

  self.registration.showNotification(title, {
    body: notification.body || '',
    icon: notification.icon || '/favicon.ico',
    data: payload.data || {}
  });
});

self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if ('focus' in client) {
          return client.focus();
        }
      }
      if (clients.openWindow) {
        return clients.openWindow('/dashboard');
      }
    })
  );
});
