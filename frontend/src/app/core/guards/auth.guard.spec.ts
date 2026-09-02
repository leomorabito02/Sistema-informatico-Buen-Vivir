import { TestBed } from "@angular/core/testing";
import { CanActivateFn, Router, UrlTree } from "@angular/router";
import { authGuard } from "./auth.guard";
import { AuthService } from "../../services/auth.service";
import { of, throwError } from "rxjs";

describe("authGuard", () => {
  const executeGuard: CanActivateFn = (...guardParameters) =>
    TestBed.runInInjectionContext(() => authGuard(...guardParameters));

  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockRouter: jasmine.SpyObj<Router>;
  const mockUrlTree = {} as UrlTree;

  beforeEach(() => {
    mockAuthService = jasmine.createSpyObj("AuthService", ["isAuthenticated", "isLocallyAuthenticated"]);
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

  it("should redirect to /account-notice if user_status is blocked", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ user_status: "blocked" } as any),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(mockUrlTree);
      expect(mockRouter.createUrlTree).toHaveBeenCalledWith([
        "/account-notice",
      ]);
      done();
    });
  });

  it("should redirect to /account-notice if user.status is blocked", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ user: { status: "blocked" } } as any),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(mockUrlTree);
      expect(mockRouter.createUrlTree).toHaveBeenCalledWith([
        "/account-notice",
      ]);
      done();
    });
  });

  it("should redirect to /account-notice when both user_status and user.status are blocked", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ user_status: "blocked", user: { status: "blocked" } } as any),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(mockUrlTree);
      expect(mockRouter.createUrlTree).toHaveBeenCalledWith([
        "/account-notice",
      ]);
      done();
    });
  });

  it("should prioritize /account-notice over /pending-approval when blocked and pending", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ user_status: "blocked", approval_status: "pending" } as any),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(mockUrlTree);
      expect(mockRouter.createUrlTree).toHaveBeenCalledWith([
        "/account-notice",
      ]);
      done();
    });
  });

  it("should redirect to /pending-approval if approval_status is pending and user is active", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({
        approval_status: "pending",
        user_status: "active",
        user: { status: "active" },
      } as any),
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

  it("should return true if user is active and approved", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({
        user_status: "active",
        approval_status: "approved",
        user: { status: "active" },
      } as any),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(true);
      expect(mockRouter.createUrlTree).not.toHaveBeenCalled();
      done();
    });
  });

  it("should return true if status properties are empty string or undefined", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      of({ user_status: "", approval_status: "" } as any),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(true);
      expect(mockRouter.createUrlTree).not.toHaveBeenCalled();
      done();
    });
  });

  it("should redirect to /login if isAuthenticated throws an error", (done) => {
    mockAuthService.isAuthenticated.and.returnValue(
      throwError(() => new Error("Unauthorized")),
    );

    const result = executeGuard({} as any, {} as any);
    (result as any).subscribe((res: any) => {
      expect(res).toBe(mockUrlTree);
      expect(mockRouter.createUrlTree).toHaveBeenCalledWith(["/login"]);
      done();
    });
  });
});
