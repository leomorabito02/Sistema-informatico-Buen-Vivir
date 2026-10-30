import { Component, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { Router, RouterLink } from "@angular/router";
import { LucideAngularModule, Search, ArrowLeft, House } from "lucide-angular";

@Component({
  selector: "app-not-found",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  templateUrl: "./not-found.component.html",
  styleUrl: "./not-found.component.css",
})
export class NotFoundComponent {
  readonly SearchIcon = Search;
  readonly ArrowLeftIcon = ArrowLeft;
  readonly HomeIcon = House;

  private readonly router = inject(Router);

  goBack(): void {
    if (window.history.length > 1) {
      window.history.back();
    } else {
      void this.router.navigate(["/"]);
    }
  }

  goToHome(): void {
    void this.router.navigate(["/"]);
  }
}
