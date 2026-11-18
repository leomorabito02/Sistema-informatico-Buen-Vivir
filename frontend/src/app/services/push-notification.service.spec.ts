import { TestBed } from "@angular/core/testing";
import {
  HttpClientTestingModule,
  HttpTestingController,
} from "@angular/common/http/testing";
import { Messaging } from "@angular/fire/messaging";
import {
  FIREBASE_SW_SCOPE,
  PushNotificationService,
} from "./push-notification.service";

describe("PushNotificationService", () => {
  let service: PushNotificationService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [PushNotificationService],
    });
    service = TestBed.inject(PushNotificationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should be created", () => {
    // Arrange & Act & Assert
    expect(service).toBeTruthy();
  });

  it("should register device token via POST /api/v1/users/me/devices", () => {
    // Arrange
    const dummyToken = "fcm-test-token-123";
    const expectedResponse = { message: "dispositivo registrado" };

    // Act
    service.registerDeviceToken(dummyToken).subscribe((response) => {
      // Assert
      expect(response).toEqual(expectedResponse);
    });

    const req = httpMock.expectOne((r) =>
      r.url.endsWith("/api/v1/users/me/devices"),
    );
    expect(req.request.method).toBe("POST");
    expect(req.request.headers.get("X-Skip-Global-Error")).toBe("true");
    expect(req.request.body.token).toBe(dummyToken);
    req.flush(expectedResponse);
  });
});

describe("PushNotificationService - permiso", () => {
  let service: PushNotificationService;
  let pedirPermiso: jasmine.Spy;

  function permisoActual(valor: NotificationPermission): void {
    spyOnProperty(Notification, "permission", "get").and.returnValue(valor);
  }

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [
        PushNotificationService,
        // Messaging de mentira: alcanza para pasar el chequeo de soporte.
        { provide: Messaging, useValue: {} },
      ],
    });
    service = TestBed.inject(PushNotificationService);
    pedirPermiso = spyOn(Notification, "requestPermission");
  });

  it("sin askPermission no muestra el pedido si el usuario no respondió", async () => {
    permisoActual("default");
    expect(await service.requestPermissionAndGetToken(false)).toBeNull();
    expect(pedirPermiso).not.toHaveBeenCalled();
  });

  it("con askPermission pregunta una vez y respeta la respuesta", async () => {
    permisoActual("default");
    pedirPermiso.and.resolveTo("denied");
    expect(await service.requestPermissionAndGetToken(true)).toBeNull();
    expect(pedirPermiso).toHaveBeenCalledTimes(1);
  });

  it("si el permiso ya fue denegado no vuelve a preguntar", async () => {
    permisoActual("denied");
    expect(await service.requestPermissionAndGetToken(true)).toBeNull();
    expect(pedirPermiso).not.toHaveBeenCalled();
  });

  it("registra el service worker de Firebase en su propio scope, no en /", async () => {
    permisoActual("granted");
    spyOn(console, "error"); // el corte del registro es a propósito
    const registro = jasmine
      .createSpy("register")
      .and.rejectWith(new Error("corte del test"));
    spyOnProperty(navigator, "serviceWorker", "get").and.returnValue({
      register: registro,
    } as unknown as ServiceWorkerContainer);

    expect(await service.requestPermissionAndGetToken(true)).toBeNull();
    expect(registro).toHaveBeenCalledOnceWith("/firebase-messaging-sw.js", {
      scope: FIREBASE_SW_SCOPE,
    });
    expect(pedirPermiso).not.toHaveBeenCalled();
  });
});
