import { Component, inject, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { Router, RouterLink } from "@angular/router";
import {
  LucideAngularModule,
  Search,
  ArrowLeft,
  House,
  RotateCw,
} from "lucide-angular";
import { NavigationHistoryService } from "../../services/navigation-history.service";

@Component({
  selector: "app-not-found",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  templateUrl: "./not-found.component.html",
  styleUrl: "./not-found.component.css",
})
export class NotFoundComponent implements OnInit {
  readonly SearchIcon = Search;
  readonly ArrowLeftIcon = ArrowLeft;
  readonly HomeIcon = House;
  readonly RotateCwIcon = RotateCw;

  private readonly router = inject(Router);
  private readonly navHistory = inject(NavigationHistoryService);

  ngOnInit(): void {
    this.navHistory.markCurrentAsNotFound(this.router.url);
    if (typeof window !== "undefined") {
      this.navHistory.markCurrentAsNotFound(window.location.pathname);
    }
  }

  reloadPage(): void {
    this.navHistory.reloadPreviousPage();
  }

  goBack(): void {
    if (typeof window !== "undefined" && window.history.length > 1) {
      window.history.back();
    } else {
      void this.router.navigate(["/"]);
    }
  }

  goToHome(): void {
    void this.router.navigate(["/"]);
  }
}
