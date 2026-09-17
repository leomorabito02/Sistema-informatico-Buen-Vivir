declare global {
  interface Window {
    __env?: {
      apiUrl?: string;
      educationApiUrl?: string;
      workerImpactUrl?: string;
    };
  }
}

export const environment = {
  production: false,
  apiUrl: window.__env?.apiUrl,
  // Base del microservicio de Educación. Vacío = mismo origen (útil detrás
  // del api-gateway o del proxy de `ng serve -c local`).
  educationApiUrl: window.__env?.educationApiUrl ?? "",
  workerImpactUrl:
    window.__env?.workerImpactUrl ??
    "https://buen-vivir-impact-staging.leonardomorabito02.workers.dev/api/impact",
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
