import { TestBed } from "@angular/core/testing";
import {
  HttpClientTestingModule,
  HttpTestingController,
} from "@angular/common/http/testing";
import { ProfileService } from "./profile.service";
import { environment } from "../../environments/environment";

describe("ProfileService", () => {
  let service: ProfileService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    environment.apiUrl = "http://localhost:3000";
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [ProfileService],
    });
    service = TestBed.inject(ProfileService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("getMyProfile should call get and return profile", () => {
    const mockProfile: any = { user: { email: "test@test.com" } };

    service.getMyProfile().subscribe((res) => {
      expect(res).toEqual(mockProfile);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/profile/me`);
    expect(req.request.method).toBe("GET");
    expect(req.request.withCredentials).toBe(true);
    req.flush(mockProfile);
  });

  it("updateMyPassword should call put with correct body", () => {
    const mockResponse: any = { success: true };

    service.updateMyPassword("old", "new").subscribe((res) => {
      expect(res).toEqual(mockResponse);
    });

    const req = httpMock.expectOne(
      `${environment.apiUrl}/api/v1/profile/me/password`,
    );
    expect(req.request.method).toBe("PUT");
    expect(req.request.body).toEqual({
      old_password: "old",
      new_password: "new",
    });
    expect(req.request.withCredentials).toBe(true);
    req.flush(mockResponse);
  });

  it("linkGoogleAccount should post id_token and google_id", () => {
    const mockResponse: any = { success: true };

    service.linkGoogleAccount("token", "gid").subscribe((res) => {
      expect(res).toEqual(mockResponse);
    });

    const req = httpMock.expectOne(
      `${environment.apiUrl}/api/v1/profile/me/link-google`,
    );
    expect(req.request.method).toBe("POST");
    expect(req.request.body).toEqual({ id_token: "token", google_id: "gid" });
    expect(req.request.withCredentials).toBe(true);
    req.flush(mockResponse);
  });

  it("unlinkGoogleAccount should call post endpoint", () => {
    const mockResponse: any = { success: true };

    service.unlinkGoogleAccount().subscribe((res) => {
      expect(res).toEqual(mockResponse);
    });

    const req = httpMock.expectOne(
      `${environment.apiUrl}/api/v1/profile/me/unlink-google`,
    );
    expect(req.request.method).toBe("POST");
    expect(req.request.withCredentials).toBe(true);
    req.flush(mockResponse);
  });

  it("updateProfile should call put with correct data", () => {
    const mockResponse: any = { success: true };
    const mockData: any = { first_name: "Test" };

    service.updateProfile(mockData).subscribe((res) => {
      expect(res).toEqual(mockResponse);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/profile/me`);
    expect(req.request.method).toBe("PUT");
    expect(req.request.body).toEqual(mockData);
    expect(req.request.withCredentials).toBe(true);
    req.flush(mockResponse);
  });

  describe("error handling", () => {
    it("should handle custom error message from server", () => {
      service.getMyProfile().subscribe({
        error: (err) => {
          expect(err.message).toBe("Custom error");
        },
      });
      const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/profile/me`);
      req.flush(
        { message: "Custom error" },
        { status: 400, statusText: "Bad Request" },
      );
    });

    it("should handle 401 error", () => {
      service.getMyProfile().subscribe({
        error: (err) => {
          expect(err.message).toBe(
            "No autorizado. Verifique sus credenciales.",
          );
        },
      });
      const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/profile/me`);
      req.flush(null, { status: 401, statusText: "Unauthorized" });
    });

    it("should handle 403 error", () => {
      service.getMyProfile().subscribe({
        error: (err) => {
          expect(err.message).toBe(
            "No tiene permisos para realizar esta acción.",
          );
        },
      });
      const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/profile/me`);
      req.flush(null, { status: 403, statusText: "Forbidden" });
    });

    it("should handle default error", () => {
      service.getMyProfile().subscribe({
        error: (err) => {
          expect(err.message).toBe(
            "Ocurrió un error inesperado. Por favor, intente nuevamente.",
          );
        },
      });
      const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/profile/me`);
      req.flush(null, { status: 500, statusText: "Server Error" });
    });
  });
});
