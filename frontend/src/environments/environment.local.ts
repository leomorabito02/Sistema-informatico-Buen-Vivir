// Entorno para desarrollo local con `ng serve -c local` (frontend en :4300).
// Todas las URLs son relativas: el proxy de Angular (proxy.conf.json) reenvía
//   /api/v1/*  -> auth-service  (Docker, :8082)
//   /api/cursos|areas-formativas|materiales|docente|salud -> education-service (Docker, :8210)
// El browser solo habla con localhost:4300 => sin CORS y las cookies funcionan.

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
  apiUrl: "",
  educationApiUrl: "",
  workerImpactUrl:
    window.__env?.workerImpactUrl ??
    "https://buen-vivir-impact-staging.workers.dev/api/impact",
  firebase: {
    projectId: "ceibuenvivir",
    appId: "1:429506836254:web:45f5eb50d86f074323bb0f",
    storageBucket: "ceibuenvivir.firebasestorage.app",
    apiKey: "AIzaSyDMdBkILiOSgqTq5ZD2VmrsPtYm7aU5bzc",
    authDomain: "ceibuenvivir.firebaseapp.com",
    messagingSenderId: "429506836254",
    measurementId: "G-4C2KCQ3MNT",
    firebaseVapidKey:
      "BIDsxZWpgrdpe84FWgzqLxknJcUaEAoWo47uaH7FjufV-IfWDobgglQDA1w6_2o6gHeddZ8dg5bLH5coAHLZLhs",
  },
};
