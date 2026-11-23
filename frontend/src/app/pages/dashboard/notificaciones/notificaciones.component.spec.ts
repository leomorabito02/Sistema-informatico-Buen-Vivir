import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  flushMicrotasks,
  tick,
} from "@angular/core/testing";
import { NotificacionesComponent } from "./notificaciones.component";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import {
  NotificationsService,
  NotificationItem,
  NotificationDetail,
} from "../../../services/notifications.service";

describe("NotificacionesComponent", () => {
  let component: NotificacionesComponent;
  let fixture: ComponentFixture<NotificacionesComponent>;
  let mockNotificationsService: jasmine.SpyObj<NotificationsService>;

  const mockNotifications: NotificationItem[] = [
    {
      recipient_id: "rcp-1",
      notification_id: "ntf-1",
      title: "Alerta de Seguridad",
      message: "Se ha modificado tu contraseña.",
      category: "security",
      priority: "high",
      source_service: "auth",
      is_critical: true,
      is_read: false,
      created_at: new Date().toISOString(),
    },
    {
      recipient_id: "rcp-2",
      notification_id: "ntf-2",
      title: "Nuevo Curso",
      message: "Ya puedes acceder al curso de Angular.",
      category: "education",
      priority: "normal",
      source_service: "education",
      is_critical: false,
      is_read: true,
      read_at: new Date().toISOString(),
      created_at: new Date().toISOString(),
    },
  ];

  beforeEach(async () => {
    mockNotificationsService = jasmine.createSpyObj<NotificationsService>(
      "NotificationsService",
      [
        "getNotifications",
        "getNotificationDetail",
        "markAsRead",
        "markAllAsRead",
        "deleteNotification",
        "getUnreadCount",
        "refreshUnread",
        "getPreferences",
        "patchPreference",
        "savePreferences",
        "initPreferences",
      ],
    );

    mockNotificationsService.getNotifications.and.callFake(async () =>
      mockNotifications.map((n) => ({ ...n })),
    );
    mockNotificationsService.getNotificationDetail.and.callFake(async (id) => ({
      ...mockNotifications[0],
      recipient_id: id,
      sender_type: "system",
      sender_id: "system",
      data: { test: true },
    }));
    mockNotificationsService.refreshUnread.and.resolveTo(0);
    mockNotificationsService.markAsRead.and.resolveTo();
    mockNotificationsService.markAllAsRead.and.resolveTo();
    mockNotificationsService.deleteNotification.and.resolveTo();
    mockNotificationsService.getPreferences.and.resolveTo([
      {
        category: "security",
        email_enabled: true,
        push_enabled: true,
        in_app_enabled: true,
      },
      {
        category: "education",
        email_enabled: true,
        push_enabled: true,
        in_app_enabled: true,
      },
    ]);
    mockNotificationsService.patchPreference.and.callFake(
      async (cat, patch) => ({
        category: cat,
        email_enabled: patch.email_enabled ?? true,
        push_enabled: patch.push_enabled ?? true,
        in_app_enabled: patch.in_app_enabled ?? true,
      }),
    );
    mockNotificationsService.savePreferences.and.callFake(
      async (prefs) => prefs,
    );
    mockNotificationsService.initPreferences.and.resolveTo([
      {
        category: "security",
        email_enabled: true,
        push_enabled: true,
        in_app_enabled: true,
      },
      {
        category: "education",
        email_enabled: true,
        push_enabled: true,
        in_app_enabled: true,
      },
      {
        category: "labor",
        email_enabled: true,
        push_enabled: true,
        in_app_enabled: true,
      },
    ]);

    await TestBed.configureTestingModule({
      imports: [NotificacionesComponent, BrowserAnimationsModule],
      providers: [
        { provide: NotificationsService, useValue: mockNotificationsService },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(NotificacionesComponent);
    component = fixture.componentInstance;
    await component.loadNotifications();
    fixture.detectChanges();
  });

  it("should create and load notifications on init", async () => {
    await fixture.whenStable();
    expect(component).toBeTruthy();
    expect(mockNotificationsService.getNotifications).toHaveBeenCalled();
    expect(component.notifications).toHaveSize(2);
    expect(component.unreadCount).toBe(1);
  });

  it("should filter notifications by unread status", async () => {
    await fixture.whenStable();
    component.setFilter("unread");
    expect(component.displayedNotifications).toHaveSize(1);
    expect(component.displayedNotifications[0].recipient_id).toBe("rcp-1");

    component.setFilter("all");
    expect(component.displayedNotifications).toHaveSize(2);
  });

  it("should mark notification as read optimistically and call service", async () => {
    await fixture.whenStable();
    const item = component.notifications[0];
    await component.markAsRead(item);

    expect(item.is_read).toBeTrue();
    expect(mockNotificationsService.markAsRead).toHaveBeenCalledWith("rcp-1");
  });

  it("should mark all notifications as read", async () => {
    await fixture.whenStable();
    await component.markAllAsRead();

    expect(component.notifications.every((n) => n.is_read)).toBeTrue();
    expect(mockNotificationsService.markAllAsRead).toHaveBeenCalled();
  });

  it("should delete notification permanently from the list", async () => {
    await fixture.whenStable();
    const itemToDelete = component.notifications[0];
    await component.deleteNotification(itemToDelete);

    expect(component.notifications).toHaveSize(1);
    expect(
      component.notifications.find((n) => n.recipient_id === "rcp-1"),
    ).toBeUndefined();
    expect(mockNotificationsService.deleteNotification).toHaveBeenCalledWith(
      "rcp-1",
    );
  });

  it("should open detail modal and fetch full detail", async () => {
    await fixture.whenStable();
    const item = component.notifications[0];
    const fullDetail: NotificationDetail = {
      ...item,
      sender_type: "system",
      sender_id: "system",
      data: { test: true },
    };
    mockNotificationsService.getNotificationDetail.and.resolveTo(fullDetail);

    await component.viewDetails(item);

    expect(component.displayDetailModal).toBeTrue();
    expect(mockNotificationsService.getNotificationDetail).toHaveBeenCalledWith(
      "rcp-1",
    );
    expect(component.selectedDetail?.data).toEqual({ test: true });
  });

  it("no muestra el servicio emisor ni en la lista ni en el detalle", async () => {
    await fixture.whenStable();
    fixture.detectChanges();
    const lista = (fixture.nativeElement as HTMLElement).textContent ?? "";
    expect(lista).not.toContain("Origen");
    expect(lista.toUpperCase()).not.toContain("EDUCATION");

    const item = component.notifications[0];
    mockNotificationsService.getNotificationDetail.and.resolveTo({
      ...item,
      sender_type: "system",
      sender_id: "system",
      data: {},
    });
    await component.viewDetails(item);
    fixture.detectChanges();
    await fixture.whenStable();
    const todo = document.body.textContent ?? "";
    expect(todo).not.toContain("Servicio emisor");
    // El dato sigue en el modelo para trazabilidad interna.
    expect(component.selectedDetail?.source_service).toBe("auth");
  });

  it("should switch between inbox and preferences tabs", () => {
    expect(component.activeTab).toBe("inbox");
    component.setActiveTab("preferences");
    expect(component.activeTab).toBe("preferences");
    component.setActiveTab("inbox");
    expect(component.activeTab).toBe("inbox");
  });

  it("should load preferences from service", async () => {
    await fixture.whenStable();
    expect(mockNotificationsService.getPreferences).toHaveBeenCalled();
    expect(component.preferences.length).toBeGreaterThan(0);
  });

  it("should detect local preference change without immediate network save", async () => {
    await fixture.whenStable();
    expect(component.hasUnsavedChanges).toBeFalse();

    component.preferences[0].email_enabled = false;
    component.onPreferenceChange();

    expect(component.hasUnsavedChanges).toBeTrue();
    expect(mockNotificationsService.savePreferences).not.toHaveBeenCalled();
  });

  it("should save all preferences in batch when requested", async () => {
    await fixture.whenStable();
    component.preferences[0].email_enabled = false;
    component.onPreferenceChange();

    await component.saveAllPreferences();

    expect(mockNotificationsService.savePreferences).toHaveBeenCalledWith(
      component.preferences,
    );
    expect(component.hasUnsavedChanges).toBeFalse();
    expect(component.preferencesSuccessMessage).toBeTruthy();
  });

  it("should discard changes and revert to initial state", async () => {
    await fixture.whenStable();
    const originalEmail = component.preferences[0].email_enabled;
    component.preferences[0].email_enabled = !originalEmail;
    component.onPreferenceChange();
    expect(component.hasUnsavedChanges).toBeTrue();

    component.cancelPreferenceChanges();
    expect(component.hasUnsavedChanges).toBeFalse();
    expect(component.preferences[0].email_enabled).toBe(originalEmail);
  });

  it("should reset all preferences using backend initPreferences", async () => {
    await fixture.whenStable();
    await component.resetPreferences();

    expect(mockNotificationsService.initPreferences).toHaveBeenCalled();
    expect(
      component.preferences.every((p) => p.email_enabled && p.push_enabled),
    ).toBeTrue();
    expect(component.preferencesSuccessMessage).toContain(
      "restablecieron a su configuración predeterminada",
    );
  });

  // La pantalla se entera sola de las notificaciones nuevas y mantiene al día el contador
  // de la campana.
  describe("actualización automática", () => {
    it("después de marcar como leída refresca el contador", async () => {
      await component.loadNotifications();
      mockNotificationsService.refreshUnread.calls.reset();
      const item = component.notifications.find((n) => !n.is_read)!;
      await component.markAsRead(item);
      await Promise.resolve();
      expect(mockNotificationsService.refreshUnread).toHaveBeenCalled();
    });

    it("si el contador falla, la notificación igual queda leída", async () => {
      mockNotificationsService.refreshUnread.and.rejectWith(new Error("red"));
      await component.loadNotifications();
      const item = component.notifications.find((n) => !n.is_read)!;
      await component.markAsRead(item);
      expect(item.is_read).toBeTrue();
    });

    it("cada minuto vuelve a pedir la lista y al salir deja de hacerlo", fakeAsync(() => {
      component.ngOnDestroy(); // corta el timer real que arrancó el beforeEach
      component.ngOnInit();
      flushMicrotasks();
      mockNotificationsService.getNotifications.calls.reset();
      tick(60_000);
      expect(mockNotificationsService.getNotifications).toHaveBeenCalledTimes(
        1,
      );
      component.ngOnDestroy();
      tick(120_000);
      expect(mockNotificationsService.getNotifications).toHaveBeenCalledTimes(
        1,
      );
    }));
  });
});
