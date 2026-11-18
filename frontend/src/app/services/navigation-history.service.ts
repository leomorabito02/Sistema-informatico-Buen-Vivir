import { inject, Injectable } from "@angular/core";
import { NavigationEnd, Router } from "@angular/router";
import { filter } from "rxjs";

const STORAGE_KEY = "buenvivir_nav_history";
const ERROR_ROUTES = new Set([
  "/500",
  "/404",
  "/400",
  "/server-error",
  "/not-found",
  "/error",
]);

@Injectable({
  providedIn: "root",
})
export class NavigationHistoryService {
  private readonly router = inject(Router);
  private history: string[] = [];

  constructor() {
    this.loadFromStorage();
    this.router.events
      .pipe(
        filter(
          (event): event is NavigationEnd => event instanceof NavigationEnd,
        ),
      )
      .subscribe((event: NavigationEnd) => {
        const url = event.urlAfterRedirects || event.url;
        if (!this.isErrorUrl(url)) {
          if (this.history[this.history.length - 1] !== url) {
            this.history.push(url);
            if (this.history.length > 20) {
              this.history.shift();
            }
            this.saveToStorage();
          }
        }
      });
  }

  isErrorUrl(url: string): boolean {
    if (!url) return true;
    const clean = url.split("?")[0].split("#")[0];
    return ERROR_ROUTES.has(clean);
  }

  markCurrentAsNotFound(url: string): void {
    if (!url) return;
    this.history = this.history.filter((h) => h !== url && !this.isErrorUrl(h));
    this.saveToStorage();
  }

  getPreviousUrl(): string {
    const nav = this.router.getCurrentNavigation();
    const stateUrl = (nav?.extras?.state as { returnUrl?: string })?.returnUrl;
    if (stateUrl && !this.isErrorUrl(stateUrl)) {
      return stateUrl;
    }

    if (typeof window !== "undefined" && window.history?.state) {
      const historyStateUrl = (window.history.state as { returnUrl?: string })
        ?.returnUrl;
      if (historyStateUrl && !this.isErrorUrl(historyStateUrl)) {
        return historyStateUrl;
      }
    }

    for (let i = this.history.length - 1; i >= 0; i--) {
      const candidate = this.history[i];
      if (
        candidate &&
        !this.isErrorUrl(candidate) &&
        candidate !== this.router.url
      ) {
        return candidate;
      }
    }

    const last = this.history[this.history.length - 1];
    if (last && !this.isErrorUrl(last)) {
      return last;
    }

    return "/";
  }

  reloadPreviousPage(): void {
    const target = this.getPreviousUrl();
    if (target && target !== "/" && !this.isErrorUrl(target)) {
      void this.router.navigateByUrl(target).then((navigated) => {
        if (!navigated && typeof window !== "undefined") {
          window.location.href = target;
        }
      });
    } else {
      void this.router.navigate(["/"]);
    }
  }

  private loadFromStorage(): void {
    if (typeof sessionStorage === "undefined") return;
    try {
      const raw = sessionStorage.getItem(STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw);
        if (Array.isArray(parsed)) {
          this.history = parsed.filter(
            (u) => typeof u === "string" && !this.isErrorUrl(u),
          );
        }
      }
    } catch {
      this.history = [];
    }
  }

  private saveToStorage(): void {
    if (typeof sessionStorage === "undefined") return;
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(this.history));
    } catch {
      // Ignore storage errors in restrictive environments
    }
  }
}
