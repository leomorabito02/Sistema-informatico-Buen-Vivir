import { Component, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { DialogModule } from "primeng/dialog";
import { ButtonModule } from "primeng/button";
import {
  LucideAngularModule,
  TriangleAlert,
  ShieldAlert,
  CircleAlert,
} from "lucide-angular";
import { AlertService } from "../../../services/alert.service";

@Component({
  selector: "app-global-error-dialog",
  standalone: true,
  imports: [CommonModule, DialogModule, ButtonModule, LucideAngularModule],
  templateUrl: "./global-error-dialog.component.html",
  styleUrl: "./global-error-dialog.component.css",
})
export class GlobalErrorDialogComponent {
  readonly alertService = inject(AlertService);

  readonly TriangleAlertIcon = TriangleAlert;
  readonly ShieldAlertIcon = ShieldAlert;
  readonly CircleAlertIcon = CircleAlert;

  onConfirm(): void {
    this.alertService.confirmDialog();
  }

  onCancel(): void {
    this.alertService.cancelDialog();
  }
}
