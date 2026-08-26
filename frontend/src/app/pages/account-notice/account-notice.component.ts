import { Component, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { Router, RouterLink } from "@angular/router";
import { LucideAngularModule, ShieldAlert, ArrowLeft } from "lucide-angular";

@Component({
  selector: "app-account-notice",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  templateUrl: "./account-notice.component.html",
  styleUrl: "./account-notice.component.css",
})
export class AccountNoticeComponent {
  readonly ShieldAlertIcon = ShieldAlert;
  readonly ArrowLeftIcon = ArrowLeft;

  private router = inject(Router);

  goBackToLogin(): void {
    this.router.navigate(["/login"]);
  }
}
