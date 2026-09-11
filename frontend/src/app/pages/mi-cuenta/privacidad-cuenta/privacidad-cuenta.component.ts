import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  LucideAngularModule,
  Download,
  Trash2,
  ShieldAlert,
  CircleCheck,
  FileText,
} from "lucide-angular";

@Component({
  selector: "app-privacidad-cuenta",
  standalone: true,
  imports: [CommonModule, LucideAngularModule],
  templateUrl: "./privacidad-cuenta.component.html",
  styleUrl: "./privacidad-cuenta.component.css",
})
export class PrivacidadCuentaComponent {
  readonly DownloadIcon = Download;
  readonly TrashIcon = Trash2;
  readonly ShieldAlertIcon = ShieldAlert;
  readonly CheckCircleIcon = CircleCheck;
  readonly FileTextIcon = FileText;

  isExported = false;

  exportUserData(): void {
    const userData = {
      name: "Juan Pérez",
      dni: "38123456",
      email: "juan.perez@ejemplo.com",
      phone: "+54 9 3547 612118",
      city: "Alta Gracia, Córdoba",
      registeredAt: "2026-01-15",
    };

    const blob = new Blob([JSON.stringify(userData, null, 2)], {
      type: "application/json",
    });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "mis_datos_buen_vivir.json";
    a.click();
    window.URL.revokeObjectURL(url);

    this.isExported = true;
    setTimeout(() => (this.isExported = false), 4000);
  }
}
