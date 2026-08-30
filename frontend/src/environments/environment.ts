declare global {
  interface Window {
    __env?: { apiUrl?: string };
  }
}

export const environment = {
  production: false,
  apiUrl: window.__env?.apiUrl,
  firebase: {
    projectId: "ceibuenvivir",
    appId: "1:429506836254:web:45f5eb50d86f074323bb0f",
    storageBucket: "ceibuenvivir.firebasestorage.app",
    apiKey: "AIzaSyDMdBkILiOSgqTq5ZD2VmrsPtYm7aU5bzc",
    authDomain: "ceibuenvivir.firebaseapp.com",
    messagingSenderId: "429506836254",
    measurementId: "G-4C2KCQ3MNT",
  },
};
