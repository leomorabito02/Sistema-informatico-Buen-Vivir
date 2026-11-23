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
import { errorInterceptor } from "./core/interceptors/error.interceptor";
import { MessageService } from "primeng/api";

import { provideFirebaseApp, initializeApp } from "@angular/fire/app";
import { provideAuth, getAuth } from "@angular/fire/auth";
import { provideMessaging, getMessaging } from "@angular/fire/messaging";
import { environment } from "../environments/environment";

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({ eventCoalescing: true }),
    MessageService,
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
    provideHttpClient(withInterceptors([authInterceptor, errorInterceptor])),
    // Sin esto, las animaciones de PrimeNG no terminan y el overlay de p-dialog o p-toast
    // queda pegado en pantalla.
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
