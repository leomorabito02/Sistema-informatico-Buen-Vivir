import { TestBed } from "@angular/core/testing";
import { CanActivateFn, Router, UrlTree } from "@angular/router";
import { Observable, firstValueFrom, of, throwError } from "rxjs";
import { adminOrVolunteerGuard } from "./admin-volunteer.guard";
import { teacherGuard } from "./teacher.guard";
import { AuthService } from "../../services/auth.service";
import { MeResponse } from "../models/auth.model";

describe("adminOrVolunteerGuard / teacherGuard", () => {
  let authService: jasmine.SpyObj<AuthService>;
  let router: jasmine.SpyObj<Router>;
  const loginTree = { to: "login" } as unknown as UrlTree;
  const dashboardTree = { to: "dashboard" } as unknown as UrlTree;

  beforeEach(() => {
    authService = jasmine.createSpyObj("AuthService", [
      "isAuthenticated",
      "isLocallyAuthenticated",
    ]);
    authService.isLocallyAuthenticated.and.returnValue(true);
    router = jasmine.createSpyObj("Router", ["createUrlTree"]);
    router.createUrlTree.and.callFake((cmds: unknown[]) =>
      cmds[0] === "/login" ? loginTree : dashboardTree,
    );
    TestBed.configureTestingModule({
      providers: [
        { provide: AuthService, useValue: authService },
        { provide: Router, useValue: router },
      ],
    });
  });

  function run(guard: CanActivateFn): Promise<boolean | UrlTree> {
    return TestBed.runInInjectionContext(() => {
      const result = guard(null as never, null as never);
      return result instanceof Observable
        ? firstValueFrom(result as Observable<boolean | UrlTree>)
        : Promise.resolve(result as boolean | UrlTree);
    });
  }

  function withRole(me: Partial<MeResponse>) {
    authService.isAuthenticated.and.returnValue(of(me as MeResponse));
  }

  const cases: [string, CanActivateFn, string[], string[]][] = [
    [
      "adminOrVolunteerGuard",
      adminOrVolunteerGuard,
      ["role_admin", "role_volunteer", "role_dev"],
      ["role_teacher", "role_participant", "role_company"],
    ],
    [
      "teacherGuard",
      teacherGuard,
      ["role_teacher", "role_dev"],
      ["role_admin", "role_volunteer", "role_participant"],
    ],
  ];

  for (const [name, guard, allowed, denied] of cases) {
    describe(name, () => {
      for (const role of allowed) {
        it(`lets ${role} in`, async () => {
          withRole({ role_id: role });
          expect(await run(guard)).toBeTrue();
        });
      }

      for (const role of denied) {
        it(`sends ${role} to the dashboard`, async () => {
          withRole({ role_id: role });
          expect(await run(guard)).toBe(dashboardTree);
        });
      }

      it("reads the role from the user when role_id is missing", async () => {
        withRole({ user: { role_id: allowed[0] } as MeResponse["user"] });
        expect(await run(guard)).toBeTrue();
      });

      it("sends to login without a local session", async () => {
        authService.isLocallyAuthenticated.and.returnValue(false);
        expect(await run(guard)).toBe(loginTree);
        expect(authService.isAuthenticated).not.toHaveBeenCalled();
      });

      it("sends to login when the session is invalid", async () => {
        authService.isAuthenticated.and.returnValue(
          throwError(() => new Error("401")),
        );
        expect(await run(guard)).toBe(loginTree);
      });
    });
  }
});
