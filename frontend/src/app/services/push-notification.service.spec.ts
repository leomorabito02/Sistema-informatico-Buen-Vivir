import { TestBed } from "@angular/core/testing";
import { HttpClientTestingModule, HttpTestingController } from "@angular/common/http/testing";
import { PushNotificationService } from "./push-notification.service";

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

    const req = httpMock.expectOne((r) => r.url.endsWith("/api/v1/users/me/devices"));
    expect(req.request.method).toBe("POST");
    expect(req.request.body.token).toBe(dummyToken);
    req.flush(expectedResponse);
  });
});
