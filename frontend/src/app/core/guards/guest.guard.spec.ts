import { TestBed } from "@angular/core/testing";
import { CanActivateFn, Router, UrlTree } from "@angular/router";
import { guestGuard } from "./guest.guard";
import { AuthService } from "../../services/auth.service";
import { of, throwError } from "rxjs";

describe("guestGuard", () => {
  const executeGuard: CanActivateFn = (...guardParameters) =>
    TestBed.runInInjectionContext(() => guestGuard(...guardParameters));

  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockRouter: jasmine.SpyObj<Router>;
  const mockUrlTree = {} as UrlTree;

  beforeEach(() => {
    mockAuthService = jasmine.createSpyObj("AuthService", [
      "isAuthenticated",
      "isLocallyAuthenticated",
    ]);
    mockAuthService.isLocallyAuthenticated.and.returnValue(true);
    mockRouter = jasmine.createSpyObj("Router", ["createUrlTree"]);
    mockRouter.createUrlTree.and.returnValue(mockUrlTree);

    TestBed.configureTestingModule({
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        { provide: Router, useValue: mockRouter },
      ],
    });
  });

  it("should redirect to /pending-approval if approval_status is pending", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ approval_status: "pending" } as any),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(mockUrlTree);
      expect(mockRouter.createUrlTree).toHaveBeenCalledWith([
        "/pending-approval",
      ]);
      done();
    });
  });

  it("should redirect to /dashboard if authenticated and not pending", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ approval_status: "approved", user: { status: "active" } } as any),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(mockUrlTree);
      expect(mockRouter.createUrlTree).toHaveBeenCalledWith(["/dashboard"]);
      done();
    });
  });

  it("should redirect to /dashboard even if approval_status is empty or unknown", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ approval_status: "" } as any),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(mockUrlTree);
      expect(mockRouter.createUrlTree).toHaveBeenCalledWith(["/dashboard"]);
      done();
    });
  });

  it("should return true if not authenticated (throws error)", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      throwError(() => new Error("Unauthorized")),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(true);
      expect(mockRouter.createUrlTree).not.toHaveBeenCalled();
      done();
    });
  });
});
