import { Component, DestroyRef, OnInit, inject } from "@angular/core";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { CommonModule } from "@angular/common";
import { ReactiveFormsModule } from "@angular/forms";
import { ActivatedRoute, Router, RouterOutlet } from "@angular/router";
import { LucideAngularModule } from "lucide-angular";
import { HeaderComponent } from "../../components/layout/header/header.component";
import { FooterComponent } from "../../components/layout/footer/footer.component";

import { ToastModule } from "primeng/toast";
import { CardModule } from "primeng/card";

@Component({
  selector: "app-dashboard",
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterOutlet,
    LucideAngularModule,
    HeaderComponent,
    FooterComponent,
    ToastModule,
    CardModule,
  ],
  templateUrl: "./dashboard.component.html",
  styleUrl: "./dashboard.component.css",
})
export class DashboardComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private destroyRef = inject(DestroyRef);

  /** Llegó redirigido por un guard de rol (`?acceso=denegado`). */
  accesoDenegado = false;

  ngOnInit(): void {
    this.route.queryParamMap
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((q) => {
        if (q.get("acceso") === "denegado") this.accesoDenegado = true;
      });
  }

  cerrarAviso(): void {
    this.accesoDenegado = false;
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { acceso: null },
      queryParamsHandling: "merge",
      replaceUrl: true,
    });
  }
}
