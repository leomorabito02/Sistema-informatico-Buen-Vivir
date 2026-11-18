import { Injectable, signal } from "@angular/core";
import { environment } from "../../environments/environment";

export interface NotificationItem {
  recipient_id: string;
  notification_id: string;
  title: string;
  message: string;
  category: string;
  priority: string;
  source_service: string;
  action_url?: string;
  is_critical: boolean;
  is_read: boolean;
  read_at?: string;
  created_at: string;
}

export interface NotificationDetail extends NotificationItem {
  sender_type: string;
  sender_id: string;
  data?: Record<string, unknown>;
}

export interface UnreadCountResponse {
  unread_count: number;
}

export interface NotificationPreference {
  category: string;
  email_enabled: boolean;
  push_enabled: boolean;
  in_app_enabled: boolean;
}

@Injectable({
  providedIn: "root",
})
export class NotificationsService {
  private get baseUrl(): string {
    const base = environment.apiUrl ?? "";
    return `${base}/api/v1/notifications`;
  }

  async getNotifications(limit = 50, offset = 0): Promise<NotificationItem[]> {
    try {
      const url = `${this.baseUrl}?limit=${encodeURIComponent(limit)}&offset=${encodeURIComponent(offset)}`;
      const res = await fetch(url, {
        method: "GET",
        headers: {
          Accept: "application/json",
        },
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error(`Failed to load notifications: HTTP ${res.status}`);
      }

      return (await res.json()) as NotificationItem[];
    } catch (err) {
      console.error("[NotificationsService] Error loading notifications:", err);
      throw err;
    }
  }

  async getNotificationDetail(id: string): Promise<NotificationDetail> {
    try {
      const url = `${this.baseUrl}/${encodeURIComponent(id)}`;
      const res = await fetch(url, {
        method: "GET",
        headers: {
          Accept: "application/json",
        },
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error(
          `Failed to fetch notification detail: HTTP ${res.status}`,
        );
      }

      return (await res.json()) as NotificationDetail;
    } catch (err) {
      console.error(
        `[NotificationsService] Error fetching detail for ${id}:`,
        err,
      );
      throw err;
    }
  }

  async markAsRead(id: string): Promise<void> {
    try {
      const url = `${this.baseUrl}/${encodeURIComponent(id)}/read`;
      const res = await fetch(url, {
        method: "PATCH",
        headers: {
          Accept: "application/json",
        },
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error(
          `Failed to mark notification as read: HTTP ${res.status}`,
        );
      }
    } catch (err) {
      console.error(`[NotificationsService] Error marking as read ${id}:`, err);
      throw err;
    }
  }

  async markAllAsRead(): Promise<void> {
    try {
      const url = `${this.baseUrl}/read-all`;
      const res = await fetch(url, {
        method: "PATCH",
        headers: {
          Accept: "application/json",
        },
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error(`Failed to mark all as read: HTTP ${res.status}`);
      }
    } catch (err) {
      console.error("[NotificationsService] Error marking all as read:", err);
      throw err;
    }
  }

  async deleteNotification(id: string): Promise<void> {
    try {
      const url = `${this.baseUrl}/${encodeURIComponent(id)}`;
      const res = await fetch(url, {
        method: "DELETE",
        credentials: "include",
      });

      if (!res.ok && res.status !== 204) {
        throw new Error(`Failed to delete notification: HTTP ${res.status}`);
      }
    } catch (err) {
      console.error(
        `[NotificationsService] Error deleting notification ${id}:`,
        err,
      );
      throw err;
    }
  }

  /**
   * Cantidad de no leídas que muestra la campana del encabezado. La
   * refrescan el encabezado (cada minuto y al cambiar de pantalla) y la
   * pantalla de notificaciones después de cada acción (auditoría
   * 2026-11-11, Problema 3: antes no había contador y la interfaz no se
   * enteraba de las notificaciones nuevas).
   */
  readonly unread = signal(0);

  async refreshUnread(): Promise<number> {
    const n = await this.getUnreadCount();
    this.unread.set(n);
    return n;
  }

  async getUnreadCount(): Promise<number> {
    try {
      const url = `${this.baseUrl}/unread-count`;
      const res = await fetch(url, {
        method: "GET",
        headers: {
          Accept: "application/json",
        },
        credentials: "include",
      });

      if (!res.ok) {
        return 0;
      }

      const body = (await res.json()) as UnreadCountResponse;
      return body.unread_count ?? 0;
    } catch (err) {
      console.error("[NotificationsService] Error fetching unread count:", err);
      return 0;
    }
  }

  async getPreferences(): Promise<NotificationPreference[]> {
    try {
      const url = `${this.baseUrl}/preferences`;
      const res = await fetch(url, {
        method: "GET",
        headers: {
          Accept: "application/json",
        },
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error(`Failed to load preferences: HTTP ${res.status}`);
      }

      const data = (await res.json()) as NotificationPreference[];
      return Array.isArray(data) ? data : [];
    } catch (err) {
      console.error("[NotificationsService] Error fetching preferences:", err);
      throw err;
    }
  }

  async initPreferences(reset = false): Promise<NotificationPreference[]> {
    try {
      const url = `${this.baseUrl}/preferences/init${reset ? "?reset=true" : ""}`;
      const res = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({ reset }),
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error(
          `Failed to initialize default preferences: HTTP ${res.status}`,
        );
      }

      const data = (await res.json()) as NotificationPreference[];
      return Array.isArray(data) ? data : [];
    } catch (err) {
      console.error(
        "[NotificationsService] Error initializing preferences:",
        err,
      );
      throw err;
    }
  }

  async patchPreference(
    category: string,
    patch: {
      email_enabled?: boolean;
      push_enabled?: boolean;
      in_app_enabled?: boolean;
    },
  ): Promise<NotificationPreference> {
    try {
      const url = `${this.baseUrl}/preferences?category=${encodeURIComponent(category)}`;
      const res = await fetch(url, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(patch),
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error(`Failed to update preference: HTTP ${res.status}`);
      }

      return (await res.json()) as NotificationPreference;
    } catch (err) {
      console.error(
        `[NotificationsService] Error updating preference for ${category}:`,
        err,
      );
      throw err;
    }
  }

  async savePreferences(
    prefs: NotificationPreference[],
  ): Promise<NotificationPreference[]> {
    try {
      const url = `${this.baseUrl}/preferences`;
      const res = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(prefs),
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error(`Failed to save preferences: HTTP ${res.status}`);
      }

      return (await res.json()) as NotificationPreference[];
    } catch (err) {
      console.error("[NotificationsService] Error saving preferences:", err);
      throw err;
    }
  }
}
