import {
  ApplicationConfig,
  provideZoneChangeDetection,
  isDevMode,
} from "@angular/core";
import { provideRouter, withInMemoryScrolling } from "@angular/router";
import { provideHttpClient, withInterceptors } from "@angular/common/http";
import { provideServiceWorker } from "@angular/service-worker";
import { providePrimeNG } from "primeng/config";
import Aura from "@primeuix/themes/aura";
import { routes } from "./app.routes";
import { authInterceptor } from "./core/interceptors/auth.interceptor";

import { provideFirebaseApp, initializeApp } from "@angular/fire/app";
import { provideAuth, getAuth } from "@angular/fire/auth";
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
    ),
    provideHttpClient(withInterceptors([authInterceptor])),
    provideServiceWorker("ngsw-worker.js", {
      enabled: !isDevMode(),
      registrationStrategy: "registerWhenStable:30000",
    }),
    provideFirebaseApp(() => initializeApp(environment.firebase)),
    provideAuth(() => getAuth()),
    providePrimeNG({
      theme: {
        preset: Aura,
        options: {
          darkModeSelector: "none",
        },
      },
    }),
  ],
};
