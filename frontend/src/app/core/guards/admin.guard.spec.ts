import { TestBed } from "@angular/core/testing";
import { Router } from "@angular/router";
import { adminGuard } from "./admin.guard";
import { AuthService } from "../../services/auth.service";
import { of, throwError } from "rxjs";

describe("AdminGuard", () => {
  let authServiceSpy: jasmine.SpyObj<AuthService>;
  let routerSpy: jasmine.SpyObj<Router>;

  beforeEach(() => {
    authServiceSpy = jasmine.createSpyObj("AuthService", [
      "isAuthenticated",
      "isLocallyAuthenticated",
    ]);
    authServiceSpy.isLocallyAuthenticated.and.returnValue(true);
    routerSpy = jasmine.createSpyObj("Router", ["createUrlTree"]);

    TestBed.configureTestingModule({
      providers: [
        { provide: AuthService, useValue: authServiceSpy },
        { provide: Router, useValue: routerSpy },
      ],
    });
  });

  it("should allow access if role_admin", (done) => {
    authServiceSpy.isAuthenticated.and.returnValue(
      of({ role_id: "role_admin" } as any),
    );

    TestBed.runInInjectionContext(() => {
      const guardResult = adminGuard(null as any, null as any);
      if (typeof guardResult !== "boolean" && "subscribe" in guardResult) {
        guardResult.subscribe((res) => {
          expect(res).toBeTrue();
          done();
        });
      }
    });
  });

  it("should allow access if role_dev", (done) => {
    authServiceSpy.isAuthenticated.and.returnValue(
      of({ role_id: "role_dev" } as any),
    );

    TestBed.runInInjectionContext(() => {
      const guardResult = adminGuard(null as any, null as any);
      if (typeof guardResult !== "boolean" && "subscribe" in guardResult) {
        guardResult.subscribe((res) => {
          expect(res).toBeTrue();
          done();
        });
      }
    });
  });

  it("should redirect to dashboard if not admin", (done) => {
    const mockTree = {} as any;
    routerSpy.createUrlTree.and.returnValue(mockTree);
    authServiceSpy.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );

    TestBed.runInInjectionContext(() => {
      const guardResult = adminGuard(null as any, null as any);
      if (typeof guardResult !== "boolean" && "subscribe" in guardResult) {
        guardResult.subscribe((res) => {
          expect(res).toBe(mockTree);
          expect(routerSpy.createUrlTree).toHaveBeenCalledWith(["/dashboard"]);
          done();
        });
      }
    });
  });

  it("should redirect to login if not authenticated", (done) => {
    const mockTree = {} as any;
    routerSpy.createUrlTree.and.returnValue(mockTree);
    authServiceSpy.isAuthenticated.and.returnValue(
      throwError(() => new Error("Unauth")),
    );

    TestBed.runInInjectionContext(() => {
      const guardResult = adminGuard(null as any, null as any);
      if (typeof guardResult !== "boolean" && "subscribe" in guardResult) {
        guardResult.subscribe((res) => {
          expect(res).toBe(mockTree);
          expect(routerSpy.createUrlTree).toHaveBeenCalledWith(["/login"]);
          done();
        });
      }
    });
  });
});
