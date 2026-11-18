import { Component, NgZone, OnDestroy, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { CardModule } from "primeng/card";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { BadgeModule } from "primeng/badge";
import { DialogModule } from "primeng/dialog";
import { ToggleSwitchModule } from "primeng/toggleswitch";
import {
  NotificationsService,
  NotificationItem,
  NotificationDetail,
  NotificationPreference,
} from "../../../services/notifications.service";

@Component({
  selector: "app-notificaciones",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    CardModule,
    ButtonModule,
    TagModule,
    BadgeModule,
    DialogModule,
    ToggleSwitchModule,
  ],
  templateUrl: "./notificaciones.component.html",
  styleUrls: ["./notificaciones.component.css"],
})
export class NotificacionesComponent implements OnInit, OnDestroy {
  notifications: NotificationItem[] = [];
  loading = false;
  error: string | null = null;
  currentFilter: "all" | "unread" = "all";
  activeTab: "inbox" | "preferences" = "inbox";

  // Detail Modal state
  selectedDetail: NotificationDetail | null = null;
  displayDetailModal = false;
  detailLoading = false;

  // Preferences state
  preferences: NotificationPreference[] = [];
  initialPreferences: NotificationPreference[] = [];
  hasUnsavedChanges = false;
  preferencesLoading = false;
  preferencesSaving = false;
  preferencesSuccessMessage: string | null = null;
  preferencesErrorMessage: string | null = null;

  constructor(private readonly notificationsService: NotificationsService) {}

  /** Refresco periódico de la lista (notificaciones nuevas sin recargar). */
  private refreshTimer: ReturnType<typeof setInterval> | null = null;
  private readonly zone = inject(NgZone);
  static readonly REFRESH_MS = 60_000;

  ngOnInit(): void {
    void this.loadNotifications();
    void this.loadPreferences();
    // Fuera de la zona de Angular: un intervalo dentro de la zona deja la
    // app siempre "inestable" (whenStable nunca termina). Se vuelve a
    // entrar solo para cargar los datos.
    this.zone.runOutsideAngular(() => {
      this.refreshTimer = setInterval(() => {
        if (typeof document === "undefined" || !document.hidden) {
          this.zone.run(() => void this.loadNotifications(true));
        }
      }, NotificacionesComponent.REFRESH_MS);
    });
  }

  /**
   * Refresca el contador de la campana. Va aparte de la acción principal:
   * si falla, no deshace lo que el usuario ya hizo (marcar, borrar).
   */
  private refreshBadge(): void {
    Promise.resolve()
      .then(() => this.notificationsService.refreshUnread())
      .catch(() => undefined);
  }

  ngOnDestroy(): void {
    if (this.refreshTimer) {
      clearInterval(this.refreshTimer);
      this.refreshTimer = null;
    }
  }

  get unreadCount(): number {
    return this.notifications.filter((n) => !n.is_read).length;
  }

  get displayedNotifications(): NotificationItem[] {
    if (this.currentFilter === "unread") {
      return this.notifications.filter((n) => !n.is_read);
    }
    return this.notifications;
  }

  /** silent: refresco de fondo, sin spinner ni cartel de error. */
  async loadNotifications(silent = false): Promise<void> {
    if (!silent) {
      this.loading = true;
      this.error = null;
    }
    try {
      this.notifications = await this.notificationsService.getNotifications();
      this.refreshBadge();
    } catch (err) {
      if (silent) return;
      console.error(
        "[NotificacionesComponent] Error loading notifications:",
        err,
      );
      this.error =
        "No se pudieron cargar las notificaciones. Intenta nuevamente.";
    } finally {
      this.loading = false;
    }
  }

  setFilter(filter: "all" | "unread"): void {
    this.currentFilter = filter;
  }

  async markAsRead(item: NotificationItem, event?: Event): Promise<void> {
    if (event) {
      event.stopPropagation();
    }
    if (item.is_read) return;

    // Optimistic update
    item.is_read = true;
    item.read_at = new Date().toISOString();

    try {
      await this.notificationsService.markAsRead(item.recipient_id);
      this.refreshBadge();
    } catch (err) {
      console.error(
        "[NotificacionesComponent] Error marking notification as read:",
        err,
      );
      // Rollback on failure
      item.is_read = false;
      item.read_at = undefined;
    }
  }

  async markAllAsRead(): Promise<void> {
    const previousState = this.notifications.map((n) => ({ ...n }));
    this.notifications.forEach((n) => {
      n.is_read = true;
      n.read_at = new Date().toISOString();
    });

    try {
      await this.notificationsService.markAllAsRead();
      this.refreshBadge();
    } catch (err) {
      console.error(
        "[NotificacionesComponent] Error marking all as read:",
        err,
      );
      this.notifications = previousState;
    }
  }

  async deleteNotification(
    item: NotificationItem,
    event?: Event,
  ): Promise<void> {
    if (event) {
      event.stopPropagation();
    }

    const previousList = [...this.notifications];
    this.notifications = this.notifications.filter(
      (n) => n.recipient_id !== item.recipient_id,
    );

    try {
      await this.notificationsService.deleteNotification(item.recipient_id);
      this.refreshBadge();
      if (this.selectedDetail?.recipient_id === item.recipient_id) {
        this.displayDetailModal = false;
        this.selectedDetail = null;
      }
    } catch (err) {
      console.error(
        "[NotificacionesComponent] Error deleting notification:",
        err,
      );
      this.notifications = previousList;
      alert("No se pudo eliminar la notificación. Por favor, reintenta.");
    }
  }

  async viewDetails(item: NotificationItem): Promise<void> {
    this.displayDetailModal = true;
    this.detailLoading = true;
    this.selectedDetail = {
      ...item,
      sender_type: "system",
      sender_id: "system",
    };

    // Auto mark as read when inspecting detail
    if (!item.is_read) {
      void this.markAsRead(item);
    }

    try {
      const fullDetail = await this.notificationsService.getNotificationDetail(
        item.recipient_id,
      );
      this.selectedDetail = fullDetail;
    } catch (err) {
      console.error("[NotificacionesComponent] Error fetching detail:", err);
    } finally {
      this.detailLoading = false;
    }
  }

  closeDetailModal(): void {
    this.displayDetailModal = false;
    this.selectedDetail = null;
  }

  getCategorySeverity(
    category?: string,
  ): "success" | "info" | "warn" | "danger" | "secondary" {
    switch (category) {
      case "security":
        return "danger";
      case "education":
        return "info";
      case "labor":
        return "success";
      case "campaigns":
        return "warn";
      case "human_promotion":
        return "info";
      default:
        return "secondary";
    }
  }

  getCategoryLabel(category?: string): string {
    switch (category) {
      case "security":
        return "Seguridad";
      case "education":
        return "Educación";
      case "labor":
        return "Inserción laboral";
      case "campaigns":
        return "Campañas";
      case "human_promotion":
        return "Promoción humana";
      default:
        return "Sistema";
    }
  }

  formatDate(dateStr?: string): string {
    if (!dateStr) return "";
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString("es-AR", {
        day: "2-digit",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return dateStr;
    }
  }

  setActiveTab(tab: "inbox" | "preferences"): void {
    this.activeTab = tab;
    if (tab === "preferences" && this.preferences.length === 0) {
      void this.loadPreferences();
    }
  }

  async loadPreferences(): Promise<void> {
    this.preferencesLoading = true;
    this.preferencesErrorMessage = null;
    try {
      const loaded = await this.notificationsService.getPreferences();
      this.preferences = loaded.map((p) => ({ ...p }));
      this.initialPreferences = loaded.map((p) => ({ ...p }));
      this.hasUnsavedChanges = false;
    } catch (err) {
      console.error(
        "[NotificacionesComponent] Error loading preferences:",
        err,
      );
      this.preferences = [];
      this.initialPreferences = [];
      this.preferencesErrorMessage =
        "No se pudieron cargar las preferencias de notificaciones.";
    } finally {
      this.preferencesLoading = false;
    }
  }

  onPreferenceChange(): void {
    for (const p of this.preferences) {
      p.in_app_enabled = p.email_enabled || p.push_enabled;
    }
    this.hasUnsavedChanges = this.checkUnsavedChanges();
  }

  isCategoryChanged(pref: NotificationPreference): boolean {
    const init = this.initialPreferences.find(
      (item) => item.category === pref.category,
    );
    if (!init) return false;
    return (
      init.push_enabled !== pref.push_enabled ||
      init.email_enabled !== pref.email_enabled ||
      init.in_app_enabled !== pref.in_app_enabled
    );
  }

  private checkUnsavedChanges(): boolean {
    if (this.preferences.length !== this.initialPreferences.length) {
      return true;
    }
    return this.preferences.some((p) => this.isCategoryChanged(p));
  }

  async saveAllPreferences(): Promise<void> {
    if (this.preferencesSaving) {
      return;
    }

    this.preferencesSaving = true;
    this.preferencesSuccessMessage = null;
    this.preferencesErrorMessage = null;

    try {
      const saved = await this.notificationsService.savePreferences(
        this.preferences,
      );
      this.preferences = saved.map((p) => ({ ...p }));
      this.initialPreferences = saved.map((p) => ({ ...p }));
      this.hasUnsavedChanges = false;
      this.preferencesSuccessMessage =
        "Configuración de notificaciones guardada con éxito.";
      setTimeout(() => {
        this.preferencesSuccessMessage = null;
      }, 3500);
    } catch (err) {
      console.error("[NotificacionesComponent] Error saving preferences:", err);
      this.preferencesErrorMessage =
        "No se pudieron guardar los cambios en las preferencias. Intenta nuevamente.";
    } finally {
      this.preferencesSaving = false;
    }
  }

  cancelPreferenceChanges(): void {
    this.preferences = this.initialPreferences.map((p) => ({ ...p }));
    this.hasUnsavedChanges = false;
    this.preferencesErrorMessage = null;
  }

  async savePreference(pref: NotificationPreference): Promise<void> {
    this.onPreferenceChange();
    await this.saveAllPreferences();
  }

  async resetPreferences(): Promise<void> {
    this.preferencesSaving = true;
    this.preferencesSuccessMessage = null;
    this.preferencesErrorMessage = null;

    try {
      const defaults = await this.notificationsService.initPreferences(true);
      this.preferences = defaults.map((p) => ({ ...p }));
      this.initialPreferences = defaults.map((p) => ({ ...p }));
      this.hasUnsavedChanges = false;
      this.preferencesSuccessMessage =
        "Todas las preferencias se restablecieron a su configuración predeterminada en el servidor.";
      setTimeout(() => {
        this.preferencesSuccessMessage = null;
      }, 3500);
    } catch (err) {
      console.error(
        "[NotificacionesComponent] Error resetting preferences:",
        err,
      );
      this.preferencesErrorMessage =
        "No se pudieron restablecer las preferencias por defecto.";
    } finally {
      this.preferencesSaving = false;
    }
  }

  getChannelModeLabel(pref: NotificationPreference): string {
    if (pref.email_enabled && pref.push_enabled) {
      return "Ambas";
    }
    if (pref.email_enabled && !pref.push_enabled) {
      return "Solo Email";
    }
    if (!pref.email_enabled && pref.push_enabled) {
      return "Solo Push";
    }
    return "Ninguna";
  }

  getChannelModeSeverity(
    pref: NotificationPreference,
  ): "success" | "info" | "warn" | "danger" | "secondary" {
    if (pref.email_enabled && pref.push_enabled) {
      return "success";
    }
    if (pref.email_enabled || pref.push_enabled) {
      return "info";
    }
    return "warn";
  }

  getCategoryDescription(category?: string): string {
    switch (category) {
      case "security":
        return "Alertas de acceso, restablecimiento de contraseña y eventos de seguridad críticos.";
      case "education":
        return "Novedades de cursos, inscripciones, material formativo y recordatorios de clases.";
      case "labor":
        return "Nuevas ofertas laborales, actualización del estado de tus postulaciones y revisión de CV.";
      case "campaigns":
        return "Iniciativas comunitarias, voluntariados, colectas y eventos de la fundación.";
      case "human_promotion":
        return "Actualizaciones de acompañamiento, fichas de seguimiento y nuevos hitos en tu línea de tiempo.";
      default:
        return "Avisos generales del sistema, novedades y mantenimiento de la plataforma.";
    }
  }

  getCategoryIcon(category?: string): string {
    switch (category) {
      case "security":
        return "pi pi-shield";
      case "education":
        return "pi pi-book";
      case "labor":
        return "pi pi-briefcase";
      case "campaigns":
        return "pi pi-heart";
      case "human_promotion":
        return "pi pi-users";
      default:
        return "pi pi-cog";
    }
  }
}
