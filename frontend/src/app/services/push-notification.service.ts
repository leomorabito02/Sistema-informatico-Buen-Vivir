import { Injectable, inject } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Messaging, getToken, onMessage } from "@angular/fire/messaging";
import { environment } from "../../environments/environment";
import { Observable, of } from "rxjs";
import { catchError } from "rxjs/operators";

export interface DeviceRegistrationResponse {
  message?: string;
}

@Injectable({
  providedIn: "root",
})
export class PushNotificationService {
  private readonly messaging = inject(Messaging, { optional: true });
  private readonly http = inject(HttpClient);
  private readonly apiUrl = environment.apiUrl ?? "";

  async requestPermissionAndGetToken(): Promise<string | null> {
    if (
      typeof window === "undefined" ||
      !("Notification" in window) ||
      !("serviceWorker" in navigator) ||
      !this.messaging
    ) {
      console.warn("[PushNotificationService] Push notifications or service workers not supported in this environment.");
      return null;
    }

    try {
      const permission = await Notification.requestPermission();
      if (permission !== "granted") {
        console.warn("[PushNotificationService] Notification permission was not granted:", permission);
        return null;
      }

      const vapidKey =
        environment.firebase?.firebaseVapidKey ||
        (environment as { firebaseVapidKey?: string }).firebaseVapidKey;

      if (!vapidKey) {
        console.error("[PushNotificationService] firebaseVapidKey is not defined in environment.");
        return null;
      }

      await navigator.serviceWorker.register("/firebase-messaging-sw.js", { scope: "/" });
      const swRegistration = await navigator.serviceWorker.ready;

      const token = await getToken(this.messaging, {
        vapidKey,
        serviceWorkerRegistration: swRegistration,
      });

      console.info("[PushNotificationService] FCM token obtained:", token);
      return token;
    } catch (error) {
      console.error("[PushNotificationService] Error obtaining FCM token:", error);
      return null;
    }
  }

  listenForegroundMessages(callback?: (payload: unknown) => void): void {
    if (!this.messaging) {
      return;
    }

    onMessage(this.messaging, (payload) => {
      console.info("[PushNotificationService] Foreground notification received:", payload);
      if (callback) {
        callback(payload);
      }

      // Automatically display native browser notification if in foreground
      const notif = (payload as { notification?: { title?: string; body?: string; icon?: string } })?.notification;
      if (notif && typeof Notification !== "undefined" && Notification.permission === "granted") {
        new Notification(notif.title || "Buen Vivir", {
          body: notif.body || "",
          icon: notif.icon || "/favicon.ico",
        });
      }
    });
  }

  registerDeviceToken(token: string): Observable<DeviceRegistrationResponse | null> {
    return this.http
      .post<DeviceRegistrationResponse>(
        `${this.apiUrl}/api/v1/users/me/devices`,
        {
          token,
          userAgent: typeof navigator !== "undefined" ? navigator.userAgent : "",
        },
        { withCredentials: true },
      )
      .pipe(catchError(() => of(null)));
  }
}
