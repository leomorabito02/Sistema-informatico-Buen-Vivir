import { TestBed } from "@angular/core/testing";
import { NavigationHistoryService } from "./navigation-history.service";
import { Router, NavigationEnd, provideRouter } from "@angular/router";
import { Subject } from "rxjs";

describe("NavigationHistoryService", () => {
  let service: NavigationHistoryService;
  let router: Router;
  let routerEvents$: Subject<unknown>;

  beforeEach(() => {
    routerEvents$ = new Subject<unknown>();

    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        NavigationHistoryService,
      ],
    });

    router = TestBed.inject(Router);
    // Replace router.events with our test subject
    Object.defineProperty(router, "events", {
      get: () => routerEvents$.asObservable(),
    });

    sessionStorage.clear();
    service = TestBed.inject(NavigationHistoryService);
  });

  afterEach(() => {
    sessionStorage.clear();
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should recognize error routes", () => {
    expect(service.isErrorUrl("/500")).toBeTrue();
    expect(service.isErrorUrl("/404")).toBeTrue();
    expect(service.isErrorUrl("/400")).toBeTrue();
    expect(service.isErrorUrl("/server-error")).toBeTrue();
    expect(service.isErrorUrl("/not-found")).toBeTrue();
    expect(service.isErrorUrl("/dashboard/usuarios")).toBeFalse();
  });

  it("should record valid navigation events and ignore error URLs", () => {
    routerEvents$.next(new NavigationEnd(1, "/dashboard", "/dashboard"));
    routerEvents$.next(new NavigationEnd(2, "/dashboard/cursos", "/dashboard/cursos"));
    routerEvents$.next(new NavigationEnd(3, "/500", "/500"));

    expect(service.getPreviousUrl()).toBe("/dashboard/cursos");
  });

  it("should mark current not found URL and remove from history", () => {
    routerEvents$.next(new NavigationEnd(1, "/dashboard", "/dashboard"));
    routerEvents$.next(new NavigationEnd(2, "/invalid-path", "/invalid-path"));

    service.markCurrentAsNotFound("/invalid-path");
    expect(service.getPreviousUrl()).toBe("/dashboard");
  });

  it("should navigate to previous URL when reloadPreviousPage is called", () => {
    spyOn(router, "navigateByUrl").and.returnValue(Promise.resolve(true));

    routerEvents$.next(new NavigationEnd(1, "/dashboard", "/dashboard"));
    routerEvents$.next(new NavigationEnd(2, "/500", "/500"));

    service.reloadPreviousPage();
    expect(router.navigateByUrl).toHaveBeenCalledWith("/dashboard");
  });

  it("should fall back to / if history is empty", () => {
    spyOn(router, "navigate");
    service.reloadPreviousPage();
    expect(router.navigate).toHaveBeenCalledWith(["/"]);
  });
});
