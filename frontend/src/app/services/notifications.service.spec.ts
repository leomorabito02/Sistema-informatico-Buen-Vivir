import { TestBed } from "@angular/core/testing";
import { NotificationsService } from "./notifications.service";

describe("NotificationsService", () => {
  let service: NotificationsService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(NotificationsService);
  });

  afterEach(() => {
    if (jasmine.isSpy(window.fetch)) {
      (window.fetch as jasmine.Spy).and.callThrough();
    }
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should get notifications using fetch", async () => {
    const mockData = [
      {
        recipient_id: "rcp-1",
        notification_id: "ntf-1",
        title: "Test",
        message: "Message",
        category: "system",
        priority: "normal",
        source_service: "system",
        is_critical: false,
        is_read: false,
        created_at: new Date().toISOString(),
      },
    ];

    spyOn(window, "fetch").and.resolveTo(
      new Response(JSON.stringify(mockData), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const result = await service.getNotifications();
    expect(result).toEqual(mockData);
    expect(window.fetch).toHaveBeenCalledWith(
      jasmine.stringMatching(/\/api\/v1\/notifications\?limit=50&offset=0/),
      jasmine.objectContaining({
        method: "GET",
        credentials: "include",
      }),
    );
  });

  it("should mark notification as read using fetch PATCH", async () => {
    spyOn(window, "fetch").and.resolveTo(
      new Response(JSON.stringify({ success: true }), {
        status: 200,
      }),
    );

    await service.markAsRead("rcp-1");
    expect(window.fetch).toHaveBeenCalledWith(
      jasmine.stringMatching(/\/api\/v1\/notifications\/rcp-1\/read/),
      jasmine.objectContaining({
        method: "PATCH",
        credentials: "include",
      }),
    );
  });

  it("should delete notification using fetch DELETE", async () => {
    spyOn(window, "fetch").and.resolveTo(new Response(null, { status: 204 }));

    await service.deleteNotification("rcp-1");
    expect(window.fetch).toHaveBeenCalledWith(
      jasmine.stringMatching(/\/api\/v1\/notifications\/rcp-1/),
      jasmine.objectContaining({
        method: "DELETE",
        credentials: "include",
      }),
    );
  });

  it("should get preferences using fetch GET", async () => {
    const mockPrefs = [
      {
        category: "education",
        email_enabled: true,
        push_enabled: false,
        in_app_enabled: true,
      },
    ];
    spyOn(window, "fetch").and.resolveTo(
      new Response(JSON.stringify(mockPrefs), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const result = await service.getPreferences();
    expect(result).toEqual(mockPrefs);
    expect(window.fetch).toHaveBeenCalledWith(
      jasmine.stringMatching(/\/api\/v1\/notifications\/preferences/),
      jasmine.objectContaining({
        method: "GET",
        credentials: "include",
      }),
    );
  });

  it("should init preferences using fetch POST", async () => {
    const mockDefaults = [
      {
        category: "security",
        email_enabled: true,
        push_enabled: true,
        in_app_enabled: true,
      },
    ];
    spyOn(window, "fetch").and.resolveTo(
      new Response(JSON.stringify(mockDefaults), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const result = await service.initPreferences();
    expect(result).toEqual(mockDefaults);
    expect(window.fetch).toHaveBeenCalledWith(
      jasmine.stringMatching(/\/api\/v1\/notifications\/preferences\/init/),
      jasmine.objectContaining({
        method: "POST",
        credentials: "include",
      }),
    );
  });

  it("should patch preference using fetch PATCH", async () => {
    const updated = {
      category: "education",
      email_enabled: false,
      push_enabled: false,
      in_app_enabled: true,
    };
    spyOn(window, "fetch").and.resolveTo(
      new Response(JSON.stringify(updated), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const result = await service.patchPreference("education", {
      email_enabled: false,
    });
    expect(result).toEqual(updated);
    expect(window.fetch).toHaveBeenCalledWith(
      jasmine.stringMatching(
        /\/api\/v1\/notifications\/preferences\?category=education/,
      ),
      jasmine.objectContaining({
        method: "PATCH",
        credentials: "include",
      }),
    );
  });

  it("should save all preferences using fetch POST", async () => {
    const mockList = [
      {
        category: "education",
        email_enabled: true,
        push_enabled: true,
        in_app_enabled: true,
      },
    ];
    spyOn(window, "fetch").and.resolveTo(
      new Response(JSON.stringify(mockList), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    const result = await service.savePreferences(mockList);
    expect(result).toEqual(mockList);
    expect(window.fetch).toHaveBeenCalledWith(
      jasmine.stringMatching(/\/api\/v1\/notifications\/preferences/),
      jasmine.objectContaining({
        method: "POST",
        credentials: "include",
      }),
    );
  });
});
