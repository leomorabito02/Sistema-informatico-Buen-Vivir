import { ComponentFixture, TestBed } from "@angular/core/testing";
import {
  BookOpen,
  Building2,
  Code,
  Handshake,
  ShieldCheck,
  Sprout,
} from "lucide-angular";
import { SelectRoleComponent } from "./select-role.component";
import { Router, provideRouter } from "@angular/router";
import { AuthService } from "../../services/auth.service";
import { of, throwError, Subject } from "rxjs";

describe("SelectRoleComponent", () => {
  let component: SelectRoleComponent;
  let fixture: ComponentFixture<SelectRoleComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockRouter: any;

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj("AuthService", ["loginWithAccount"]);

    await TestBed.configureTestingModule({
      imports: [SelectRoleComponent],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        provideRouter([]),
      ],
    }).compileComponents();

    mockRouter = TestBed.inject(Router);
    spyOn(mockRouter, "navigate");
    spyOn(mockRouter, "getCurrentNavigation").and.returnValue(null as any);
  });

  describe("ngOnInit and Initial State", () => {
    it("should have initial default states and roleIcons defined", () => {
      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      expect(component.loginError).toBe("");
      expect(component.loadingAccountId).toBe("");
      expect(component.roleIcons["role_admin"]).toBe(ShieldCheck);
      expect(component.roleIcons["role_dev"]).toBe(Code);
      expect(component.roleIcons["role_participant"]).toBe(Sprout);
      expect(component.roleIcons["role_company"]).toBe(Building2);
      expect(component.roleIcons["role_teacher"]).toBe(BookOpen);
      expect(component.roleIcons["role_volunteer"]).toBe(Handshake);
    });

    it("should redirect to login if no accounts in state", () => {
      (mockRouter.getCurrentNavigation as jasmine.Spy).and.returnValue(null);
      history.replaceState({}, "");

      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      fixture.detectChanges();

      expect(mockRouter.navigate).toHaveBeenCalledWith(["/login"]);
    });

    it("should redirect to login if accounts is empty array []", () => {
      (mockRouter.getCurrentNavigation as jasmine.Spy).and.returnValue(null);
      history.replaceState({ accounts: [], loginNonce: "nonce" }, "");

      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      fixture.detectChanges();

      expect(mockRouter.navigate).toHaveBeenCalledWith(["/login"]);
    });

    it("should redirect to login if loginNonce is missing", () => {
      (mockRouter.getCurrentNavigation as jasmine.Spy).and.returnValue(null);
      history.replaceState({ accounts: [{ user_id: "1", role_id: "r1" }] }, "");

      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      fixture.detectChanges();

      expect(mockRouter.navigate).toHaveBeenCalledWith(["/login"]);
    });

    it("should load accounts from router getCurrentNavigation state", () => {
      const accounts = [
        { user_id: "1", role_id: "r1", role_name: "Role 1", status: "active" },
      ];
      (mockRouter.getCurrentNavigation as jasmine.Spy).and.returnValue({
        extras: { state: { accounts, loginNonce: "nav-nonce" } },
      });
      history.replaceState({}, "");

      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      fixture.detectChanges();

      expect(component.accounts).toEqual(accounts);
      expect(component.loginNonce).toBe("nav-nonce");
      expect(mockRouter.navigate).not.toHaveBeenCalled();
    });

    it("should load accounts from history state if navigation state is empty", () => {
      const accounts = [
        { user_id: "1", role_id: "r1", role_name: "Role 1", status: "active" },
      ];
      history.pushState({ accounts, loginNonce: "nonce" }, "");

      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      fixture.detectChanges();

      expect(component.accounts).toEqual(accounts);
      expect(component.loginNonce).toBe("nonce");
      expect(mockRouter.navigate).not.toHaveBeenCalled();
    });
  });

  describe("selectAccount", () => {
    beforeEach(() => {
      const accounts = [
        { user_id: "1", role_id: "r1", role_name: "Role 1", status: "active" },
      ];
      history.pushState({ accounts, loginNonce: "nonce" }, "");

      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      fixture.detectChanges();

      mockRouter.navigate.calls.reset();
    });

    it("should navigate to account-notice if account is not active", () => {
      const account = {
        user_id: "1",
        role_id: "r1",
        role_name: "Role 1",
        status: "inactive",
      };
      component.selectAccount(account);
      expect(mockRouter.navigate).toHaveBeenCalledWith(["/account-notice"]);
    });

    it("should login and navigate to dashboard", () => {
      const account = {
        user_id: "1",
        role_id: "r1",
        role_name: "Role 1",
        status: "active",
      };
      const subject = new Subject<any>();
      mockAuthService.loginWithAccount.and.returnValue(subject.asObservable());

      component.loginError = "previous error";

      component.selectAccount(account);

      expect(mockAuthService.loginWithAccount).toHaveBeenCalledWith(
        "1",
        "r1",
        "nonce",
      );
      expect(component.loginError).toBe("");
      expect(component.loadingAccountId).toBe("1");

      subject.next({ token: "jwt", user: {} as any });

      expect(mockRouter.navigate).toHaveBeenCalledWith(["/dashboard"]);
      expect(component.loadingAccountId).toBe("");
    });

    it("should navigate to account-notice if login response returns blocked", () => {
      const account = {
        user_id: "1",
        role_id: "r1",
        role_name: "Role 1",
        status: "active",
      };
      mockAuthService.loginWithAccount.and.returnValue(
        of({ token: "jwt", user: {} as any, user_status: "blocked" }),
      );

      component.selectAccount(account);

      expect(mockRouter.navigate).toHaveBeenCalledWith(["/account-notice"]);
    });

    it("should navigate to pending-approval if login response returns pending", () => {
      const account = {
        user_id: "1",
        role_id: "r1",
        role_name: "Role 1",
        status: "active",
      };
      mockAuthService.loginWithAccount.and.returnValue(
        of({ token: "jwt", user: {} as any, approval_status: "pending" }),
      );

      component.selectAccount(account);

      expect(mockRouter.navigate).toHaveBeenCalledWith(["/pending-approval"]);
    });

    it("should navigate to account-notice on error containing activa", () => {
      const account = {
        user_id: "1",
        role_id: "r1",
        role_name: "Role 1",
        status: "active",
      };
      mockAuthService.loginWithAccount.and.returnValue(
        throwError(() => ({ error: { error: "cuenta no activa" } })),
      );

      component.selectAccount(account);

      expect(mockRouter.navigate).toHaveBeenCalledWith(["/account-notice"]);
    });

    it("should navigate to account-notice on error containing bloque", () => {
      const account = {
        user_id: "1",
        role_id: "r1",
        role_name: "Role 1",
        status: "active",
      };
      mockAuthService.loginWithAccount.and.returnValue(
        throwError(() => ({ error: { error: "usuario bloqueado" } })),
      );

      component.selectAccount(account);

      expect(mockRouter.navigate).toHaveBeenCalledWith(["/account-notice"]);
    });

    it("should navigate to verify-email on error containing EMAIL_NOT_VERIFIED", () => {
      const account = {
        user_id: "1",
        role_id: "r1",
        role_name: "Role 1",
        status: "active",
      };
      mockAuthService.loginWithAccount.and.returnValue(
        throwError(() => ({ error: { error: "EMAIL_NOT_VERIFIED" } })),
      );

      component.selectAccount(account);

      expect(mockRouter.navigate).toHaveBeenCalledWith(["/verify-email"]);
    });

    it("should show generic error message on other errors", () => {
      const account = {
        user_id: "1",
        role_id: "r1",
        role_name: "Role 1",
        status: "active",
      };
      const subject = new Subject<any>();
      mockAuthService.loginWithAccount.and.returnValue(subject.asObservable());

      component.selectAccount(account);

      subject.error({ error: { error: "internal error" } });

      expect(component.loginError).toBe(
        "Ocurrió un error al iniciar sesión. Intentá nuevamente.",
      );
      expect(mockRouter.navigate).not.toHaveBeenCalled();
      expect(component.loadingAccountId).toBe("");
    });

    it("should show generic error message if err.error.error is missing", () => {
      const account = {
        user_id: "1",
        role_id: "r1",
        role_name: "Role 1",
        status: "active",
      };
      mockAuthService.loginWithAccount.and.returnValue(
        throwError(() => ({ error: {} })),
      );

      component.selectAccount(account);

      expect(component.loginError).toBe(
        "Ocurrió un error al iniciar sesión. Intentá nuevamente.",
      );
    });
  });
});
