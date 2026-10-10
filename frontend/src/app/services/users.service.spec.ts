import { TestBed } from "@angular/core/testing";
import {
  HttpClientTestingModule,
  HttpTestingController,
} from "@angular/common/http/testing";
import { UsersService } from "./users.service";
import { User } from "../core/models/auth.model";
import { environment } from "../../environments/environment";

describe("UsersService", () => {
  let service: UsersService;
  let httpMock: HttpTestingController;

  const mockUsers: User[] = [
    {
      user_id: "1",
      first_name: "Admin",
      last_name: "User",
      email: "admin@example.com",
      phone: "123456",
      status: "active",
      role_id: "role_admin",
      twofa_enabled: false,
      created_at: "2026-01-01",
      updated_at: "2026-01-01",
    },
    {
      user_id: "2",
      first_name: "Teacher",
      last_name: "User",
      email: "teacher@example.com",
      phone: "123456",
      status: "active",
      role_id: "role_teacher",
      twofa_enabled: false,
      created_at: "2026-01-01",
      updated_at: "2026-01-01",
    },
  ];

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [UsersService],
    });

    service = TestBed.inject(UsersService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should list users successfully", () => {
    service.listUsers().subscribe((users) => {
      expect(users).toHaveSize(2);
      expect(users).toEqual(mockUsers);
    });

    const req = httpMock.expectOne(`${environment.apiUrl ?? ""}/api/v1/users`);
    expect(req.request.method).toBe("GET");
    expect(req.request.withCredentials).toBeTrue();
    req.flush({ success: true, data: mockUsers });
  });

  it("should return empty array if data is missing in response", () => {
    service.listUsers().subscribe((users) => {
      expect(users).toHaveSize(0);
    });

    const req = httpMock.expectOne(`${environment.apiUrl ?? ""}/api/v1/users`);
    req.flush({ success: true } as any);
  });

  it("should list users by role", () => {
    service.listByRole("role_admin").subscribe((users) => {
      expect(users).toHaveSize(1);
      expect(users[0].first_name).toBe("Admin");
    });

    const req = httpMock.expectOne(`${environment.apiUrl ?? ""}/api/v1/users`);
    req.flush({ success: true, data: mockUsers });
  });

  it("should handle error with body.errors", () => {
    service.listUsers().subscribe({
      next: () => fail("should have failed"),
      error: (err: Error) => {
        expect(err.message).toBe("Error desde errors");
      },
    });

    const req = httpMock.expectOne(`${environment.apiUrl ?? ""}/api/v1/users`);
    req.flush(
      { errors: "Error desde errors" },
      { status: 400, statusText: "Bad Request" },
    );
  });

  it("should handle error with body.error", () => {
    service.listUsers().subscribe({
      next: () => fail("should have failed"),
      error: (err: Error) => {
        expect(err.message).toBe("Error desde error");
      },
    });

    const req = httpMock.expectOne(`${environment.apiUrl ?? ""}/api/v1/users`);
    req.flush(
      { error: "Error desde error" },
      { status: 400, statusText: "Bad Request" },
    );
  });

  it("should handle 401 error", () => {
    service.listUsers().subscribe({
      next: () => fail("should have failed"),
      error: (err: Error) => {
        expect(err.message).toBe("Sesión inválida. Volvé a iniciar sesión.");
      },
    });

    const req = httpMock.expectOne(`${environment.apiUrl ?? ""}/api/v1/users`);
    req.flush({}, { status: 401, statusText: "Unauthorized" });
  });

  it("should handle 403 error", () => {
    service.listUsers().subscribe({
      next: () => fail("should have failed"),
      error: (err: Error) => {
        expect(err.message).toBe(
          "Solo un administrador puede listar usuarios.",
        );
      },
    });

    const req = httpMock.expectOne(`${environment.apiUrl ?? ""}/api/v1/users`);
    req.flush({}, { status: 403, statusText: "Forbidden" });
  });

  it("should handle 500 error", () => {
    service.listUsers().subscribe({
      next: () => fail("should have failed"),
      error: (err: Error) => {
        expect(err.message).toBe(
          "Solo un administrador puede listar usuarios.",
        );
      },
    });

    const req = httpMock.expectOne(`${environment.apiUrl ?? ""}/api/v1/users`);
    req.flush({}, { status: 500, statusText: "Internal Server Error" });
  });

  it("should handle generic error", () => {
    service.listUsers().subscribe({
      next: () => fail("should have failed"),
      error: (err: Error) => {
        expect(err.message).toBe("No se pudo obtener el listado de usuarios.");
      },
    });

    const req = httpMock.expectOne(`${environment.apiUrl ?? ""}/api/v1/users`);
    req.flush({}, { status: 404, statusText: "Not Found" });
  });

  describe("deleteUser", () => {
    const url = () => `${environment.apiUrl ?? ""}/api/v1/users/u9`;

    it("hace DELETE con credenciales", () => {
      let done = false;
      service.deleteUser("u9").subscribe(() => (done = true));

      const req = httpMock.expectOne(url());
      expect(req.request.method).toBe("DELETE");
      expect(req.request.withCredentials).toBeTrue();
      req.flush({ success: true, data: { message: "ok" } });
      expect(done).toBeTrue();
    });

    it("devuelve el error del backend", () => {
      let message = "";
      service.deleteUser("u9").subscribe({
        error: (err: Error) => (message = err.message),
      });

      httpMock
        .expectOne(url())
        .flush(
          { error: "forbidden: access denied" },
          { status: 500, statusText: "Internal Server Error" },
        );
      expect(message).toBe("forbidden: access denied");
    });

    it("explica cuando la sesión venció", () => {
      let message = "";
      service.deleteUser("u9").subscribe({
        error: (err: Error) => (message = err.message),
      });

      httpMock
        .expectOne(url())
        .flush({}, { status: 401, statusText: "Unauthorized" });
      expect(message).toContain("Tu sesión expiró");
    });
  });
});
