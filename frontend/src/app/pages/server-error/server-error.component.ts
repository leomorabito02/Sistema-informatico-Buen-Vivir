import { Component, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { Router, RouterLink } from "@angular/router";
import { LucideAngularModule, TriangleAlert, RotateCw, House } from "lucide-angular";

@Component({
  selector: "app-server-error",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  templateUrl: "./server-error.component.html",
  styleUrl: "./server-error.component.css",
})
export class ServerErrorComponent {
  readonly TriangleAlertIcon = TriangleAlert;
  readonly RotateCwIcon = RotateCw;
  readonly HomeIcon = House;

  private readonly router = inject(Router);

  reloadPage(): void {
    window.location.reload();
  }

  goToHome(): void {
    void this.router.navigate(["/"]);
  }
}
