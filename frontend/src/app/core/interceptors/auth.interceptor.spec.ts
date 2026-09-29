import { TestBed } from "@angular/core/testing";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import {
  HttpClient,
  provideHttpClient,
  withInterceptors,
} from "@angular/common/http";
import { of, throwError } from "rxjs";
import { authInterceptor } from "./auth.interceptor";
import { AuthService } from "../../services/auth.service";
import { LoginResponse } from "../models/auth.model";

describe("authInterceptor", () => {
  let http: HttpClient;
  let httpMock: HttpTestingController;
  let authService: jasmine.SpyObj<AuthService>;

  beforeEach(() => {
    authService = jasmine.createSpyObj<AuthService>("AuthService", [
      "refreshToken",
      "logout",
    ]);
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([authInterceptor])),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authService },
      ],
    });

    http = TestBed.inject(HttpClient);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should clone request and set withCredentials to true", () => {
    http.get("/api/test").subscribe();

    const req = httpMock.expectOne("/api/test");
    expect(req.request.withCredentials).toBe(true);
    expect(req.request.method).toBe("GET");
    expect(req.request.url).toBe("/api/test");
    req.flush({});
  });

  it("should set withCredentials to true for POST requests with body", () => {
    http.post("/api/data", { foo: "bar" }).subscribe();

    const req = httpMock.expectOne("/api/data");
    expect(req.request.withCredentials).toBe(true);
    expect(req.request.method).toBe("POST");
    expect(req.request.body).toEqual({ foo: "bar" });
    req.flush({});
  });

  it("should not add credentials to external URLs", () => {
    http.get("https://example.com/data").subscribe();

    const req = httpMock.expectOne("https://example.com/data");
    expect(req.request.withCredentials).toBeFalse();
    req.flush({});
  });

  it("should refresh the session and retry when the user accepts extending it", () => {
    spyOn(window, "confirm").and.returnValue(true);
    authService.refreshToken.and.returnValue(
      of({ token: "t" } as LoginResponse),
    );
    let body: unknown;
    http.get("/api/cursos").subscribe((b) => (body = b));

    httpMock
      .expectOne("/api/cursos")
      .flush("", { status: 401, statusText: "Unauthorized" });
    httpMock.expectOne("/api/cursos").flush({ ok: true });

    expect(authService.refreshToken).toHaveBeenCalled();
    expect(body).toEqual({ ok: true });
    expect(authService.logout).not.toHaveBeenCalled();
  });

  it("should log out when the user declines extending the session", () => {
    spyOn(window, "confirm").and.returnValue(false);
    let status = 0;
    http.get("/api/cursos").subscribe({ error: (e) => (status = e.status) });

    httpMock
      .expectOne("/api/cursos")
      .flush("", { status: 401, statusText: "Unauthorized" });

    expect(status).toBe(401);
    expect(authService.logout).toHaveBeenCalled();
    expect(authService.refreshToken).not.toHaveBeenCalled();
  });

  it("should log out when the refresh fails", () => {
    spyOn(window, "confirm").and.returnValue(true);
    authService.refreshToken.and.returnValue(
      throwError(() => new Error("refresh vencido")),
    );
    let error: Error | undefined;
    http.get("/api/cursos").subscribe({ error: (e) => (error = e) });

    httpMock
      .expectOne("/api/cursos")
      .flush("", { status: 401, statusText: "Unauthorized" });

    expect(error?.message).toBe("refresh vencido");
    expect(authService.logout).toHaveBeenCalled();
  });

  it("should not try to refresh on auth endpoints or other errors", () => {
    const confirmSpy = spyOn(window, "confirm");
    http.post("/api/v1/auth/login", {}).subscribe({ error: () => undefined });
    http.get("/api/cursos").subscribe({ error: () => undefined });

    httpMock
      .expectOne("/api/v1/auth/login")
      .flush("", { status: 401, statusText: "Unauthorized" });
    httpMock
      .expectOne("/api/cursos")
      .flush("", { status: 500, statusText: "Server Error" });

    expect(confirmSpy).not.toHaveBeenCalled();
    expect(authService.refreshToken).not.toHaveBeenCalled();
  });
});
