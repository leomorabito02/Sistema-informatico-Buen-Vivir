export const environment = {
  production: false,
  apiUrl: (window as any).__env?.apiUrl,
  firebase: {
    projectId: 'your_project_id',
    appId: 'your_app_id',
    storageBucket: 'your_storage_bucket',
    apiKey: 'your_api_key',
    authDomain: 'your_auth_domain',
    messagingSenderId: 'your_messaging_sender_id',
  }
};