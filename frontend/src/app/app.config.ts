import {
  ApplicationConfig,
  provideZoneChangeDetection,
  isDevMode,
} from "@angular/core";
import {
  provideRouter,
  withInMemoryScrolling,
  withViewTransitions,
} from "@angular/router";
import { provideHttpClient, withInterceptors } from "@angular/common/http";
import { provideServiceWorker } from "@angular/service-worker";
import { provideAnimations } from "@angular/platform-browser/animations";
import { providePrimeNG } from "primeng/config";
import { BuenVivirPreset } from "./theme/buen-vivir.preset";
import { routes } from "./app.routes";
import { authInterceptor } from "./core/interceptors/auth.interceptor";

import { provideFirebaseApp, initializeApp } from "@angular/fire/app";
import { provideAuth, getAuth } from "@angular/fire/auth";
import { provideMessaging, getMessaging } from "@angular/fire/messaging";
import { environment } from "../environments/environment";

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideRouter(
      routes,
      withInMemoryScrolling({
        scrollPositionRestoration: "top",
        anchorScrolling: "enabled",
      }),
      // Fundido corto al cambiar de pantalla (ver ::view-transition en
      // styles.css). Sin soporte del navegador, navega igual que antes.
      withViewTransitions(),
    ),
    provideHttpClient(withInterceptors([authInterceptor])),
    // Sin esto, las animaciones de PrimeNG (p-dialog, p-toast, p-select…)
    // no disparan el callback de fin de transición, que es lo que sacan
    // el overlay/mask del DOM al cerrar: sin provider de animaciones el
    // mask queda pegado en pantalla (se ve "trabada"/borrosa) para siempre.
    provideAnimations(),
    provideServiceWorker("ngsw-worker.js", {
      enabled: !isDevMode(),
      registrationStrategy: "registerWhenStable:30000",
    }),
    provideFirebaseApp(() => initializeApp(environment.firebase)),
    provideAuth(() => getAuth()),
    provideMessaging(() => getMessaging()),
    providePrimeNG({
      theme: {
        preset: BuenVivirPreset,
        options: {
          darkModeSelector: "none",
        },
      },
    }),
  ],
};
