import { TestBed } from "@angular/core/testing";
import {
  HttpClient,
  provideHttpClient,
  withInterceptors,
} from "@angular/common/http";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import { Router } from "@angular/router";
import {
  errorInterceptor,
  SKIP_GLOBAL_ERROR_HEADER,
  withSkipGlobalError,
} from "./error.interceptor";
import { AlertService } from "../../services/alert.service";

describe("errorInterceptor", () => {
  let httpClient: HttpClient;
  let httpMock: HttpTestingController;
  let routerSpy: jasmine.SpyObj<Router>;
  let alertServiceSpy: jasmine.SpyObj<AlertService>;

  beforeEach(() => {
    routerSpy = jasmine.createSpyObj("Router", ["navigate"]);
    alertServiceSpy = jasmine.createSpyObj("AlertService", [
      "showOfflineToast",
      "showForbiddenDialog",
      "showRateLimitDialog",
      "showServerErrorDialog",
      "showError",
    ]);

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([errorInterceptor])),
        provideHttpClientTesting(),
        { provide: Router, useValue: routerSpy },
        { provide: AlertService, useValue: alertServiceSpy },
      ],
    });

    httpClient = TestBed.inject(HttpClient);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should handle status 0 (offline/network) and show offline toast", () => {
    httpClient.get("/test").subscribe({
      next: () => fail("should have failed"),
      error: (err) => {
        expect(err.status).toBe(0);
        expect(alertServiceSpy.showOfflineToast).toHaveBeenCalled();
      },
    });

    const req = httpMock.expectOne("/test");
    req.error(new ProgressEvent("error"), { status: 0 });
  });

  it("should ignore status 401 so authInterceptor can handle session refresh", () => {
    httpClient.get("/test").subscribe({
      next: () => fail("should have failed"),
      error: (err) => {
        expect(err.status).toBe(401);
        expect(alertServiceSpy.showForbiddenDialog).not.toHaveBeenCalled();
        expect(routerSpy.navigate).not.toHaveBeenCalled();
      },
    });

    const req = httpMock.expectOne("/test");
    req.flush(
      { message: "Unauthorized" },
      { status: 401, statusText: "Unauthorized" },
    );
  });

  it("should redirect to /verify-email on 403 EMAIL_NOT_VERIFIED", () => {
    httpClient.get("/test").subscribe({
      next: () => fail("should have failed"),
      error: (err) => {
        expect(err.status).toBe(403);
        expect(routerSpy.navigate).toHaveBeenCalledWith(["/verify-email"], {
          queryParams: { email: "user@test.com" },
        });
      },
    });

    const req = httpMock.expectOne("/test");
    req.flush(
      { error: "EMAIL_NOT_VERIFIED", email: "user@test.com" },
      { status: 403, statusText: "Forbidden" },
    );
  });

  it("should redirect to /pending-approval on 403 ACCOUNT_PENDING_APPROVAL", () => {
    httpClient.get("/test").subscribe({
      next: () => fail("should have failed"),
      error: () => {
        expect(routerSpy.navigate).toHaveBeenCalledWith(["/pending-approval"]);
      },
    });

    const req = httpMock.expectOne("/test");
    req.flush(
      { error: "ACCOUNT_PENDING_APPROVAL" },
      { status: 403, statusText: "Forbidden" },
    );
  });

  it("should redirect to /account-notice on 403 ACCOUNT_BLOCKED", () => {
    httpClient.get("/test").subscribe({
      next: () => fail("should have failed"),
      error: () => {
        expect(routerSpy.navigate).toHaveBeenCalledWith(["/account-notice"]);
      },
    });

    const req = httpMock.expectOne("/test");
    req.flush(
      { error: "ACCOUNT_BLOCKED" },
      { status: 403, statusText: "Forbidden" },
    );
  });

  it("should show forbidden dialog on generic 403", () => {
    httpClient.get("/test").subscribe({
      next: () => fail("should have failed"),
      error: () => {
        expect(alertServiceSpy.showForbiddenDialog).toHaveBeenCalledWith(
          "Permiso denegado",
        );
      },
    });

    const req = httpMock.expectOne("/test");
    req.flush(
      { message: "Permiso denegado" },
      { status: 403, statusText: "Forbidden" },
    );
  });

  it("should redirect to /404 when X-Redirect-On-404 header is present", () => {
    httpClient
      .get("/test", { headers: { "X-Redirect-On-404": "true" } })
      .subscribe({
        next: () => fail("should have failed"),
        error: () => {
          expect(routerSpy.navigate).toHaveBeenCalledWith(["/404"]);
        },
      });

    const req = httpMock.expectOne("/test");
    req.flush(
      { message: "Not Found" },
      { status: 404, statusText: "Not Found" },
    );
  });

  it("should show error toast on standard 404 without redirect", () => {
    httpClient.get("/test").subscribe({
      next: () => fail("should have failed"),
      error: () => {
        expect(alertServiceSpy.showError).toHaveBeenCalled();
        expect(routerSpy.navigate).not.toHaveBeenCalled();
      },
    });

    const req = httpMock.expectOne("/test");
    req.flush(
      { message: "Item no encontrado" },
      { status: 404, statusText: "Not Found" },
    );
  });

  it("should show rate limit dialog on 429", () => {
    httpClient.get("/test").subscribe({
      next: () => fail("should have failed"),
      error: () => {
        expect(alertServiceSpy.showRateLimitDialog).toHaveBeenCalled();
      },
    });

    const req = httpMock.expectOne("/test");
    req.flush(
      { message: "Too many requests" },
      { status: 429, statusText: "Too Many Requests" },
    );
  });

  it("should navigate to /500 on GET request returning 500", () => {
    httpClient.get("/test").subscribe({
      next: () => fail("should have failed"),
      error: () => {
        expect(routerSpy.navigate).toHaveBeenCalledWith(["/500"]);
      },
    });

    const req = httpMock.expectOne("/test");
    req.flush(
      { message: "Internal error" },
      { status: 500, statusText: "Server Error" },
    );
  });

  it("should show server error dialog on POST request returning 500", () => {
    httpClient.post("/test", { data: 123 }).subscribe({
      next: () => fail("should have failed"),
      error: () => {
        expect(alertServiceSpy.showServerErrorDialog).toHaveBeenCalledWith(
          "Fallo en base de datos",
        );
        expect(routerSpy.navigate).not.toHaveBeenCalled();
      },
    });

    const req = httpMock.expectOne("/test");
    req.flush(
      { message: "Fallo en base de datos" },
      { status: 500, statusText: "Server Error" },
    );
  });

  it("should skip handling if X-Skip-Global-Error header is present", () => {
    httpClient
      .get("/test", { headers: { "X-Skip-Global-Error": "true" } })
      .subscribe({
        next: () => fail("should have failed"),
        error: () => {
          expect(alertServiceSpy.showServerErrorDialog).not.toHaveBeenCalled();
          expect(routerSpy.navigate).not.toHaveBeenCalled();
        },
      });

    const req = httpMock.expectOne("/test");
    req.flush(
      { message: "Server error" },
      { status: 500, statusText: "Server Error" },
    );
  });

  it("should create skip-global-error headers using withSkipGlobalError helper", () => {
    const custom = withSkipGlobalError({ "Custom-Header": "abc" });
    expect(custom.headers.get(SKIP_GLOBAL_ERROR_HEADER)).toBe("true");
    expect(custom.headers.get("Custom-Header")).toBe("abc");
  });
});
