import { Component } from "@angular/core";
import { RouterOutlet } from "@angular/router";
import { CookieConsentComponent } from "./components/layout/cookie-consent/cookie-consent.component";
import { SessionExpiredDialogComponent } from "./components/layout/session-expired-dialog/session-expired-dialog.component";

@Component({
  selector: "app-root",
  standalone: true,
  imports: [
    RouterOutlet,
    CookieConsentComponent,
    SessionExpiredDialogComponent,
  ],
  templateUrl: "./app.component.html",
  styleUrl: "./app.component.css",
})
export class AppComponent {
  title = "frontend-buen-vivir";
}
