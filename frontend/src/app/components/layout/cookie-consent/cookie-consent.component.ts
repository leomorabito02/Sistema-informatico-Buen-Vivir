import { Component, OnInit, PLATFORM_ID, inject } from "@angular/core";
import { CommonModule, isPlatformBrowser } from "@angular/common";
import { RouterLink } from "@angular/router";
import { LucideAngularModule, Cookie, ShieldCheck, X } from "lucide-angular";

@Component({
  selector: "app-cookie-consent",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  templateUrl: "./cookie-consent.component.html",
  styleUrl: "./cookie-consent.component.css",
})
export class CookieConsentComponent implements OnInit {
  readonly CookieIcon = Cookie;
  readonly ShieldCheckIcon = ShieldCheck;
  readonly CloseIcon = X;

  isVisible = false;
  private platformId = inject(PLATFORM_ID);

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      const consent = localStorage.getItem("cookie_consent");
      if (!consent) {
        this.isVisible = true;
      }
    }
  }

  acceptCookies(): void {
    if (isPlatformBrowser(this.platformId)) {
      localStorage.setItem("cookie_consent", "accepted");
    }
    this.isVisible = false;
  }

  rejectCookies(): void {
    if (isPlatformBrowser(this.platformId)) {
      localStorage.setItem("cookie_consent", "rejected");
    }
    this.isVisible = false;
  }
}
