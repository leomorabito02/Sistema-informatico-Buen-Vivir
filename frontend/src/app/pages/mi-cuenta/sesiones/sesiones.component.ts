import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import { LucideAngularModule, Laptop, Smartphone, LogOut, ShieldCheck, CircleCheck } from "lucide-angular";

interface SessionItem {
  id: string;
  device: string;
  browser: string;
  ip: string;
  location: string;
  lastActive: string;
  isCurrent: boolean;
}

@Component({
  selector: "app-sesiones",
  standalone: true,
  imports: [CommonModule, LucideAngularModule],
  templateUrl: "./sesiones.component.html",
  styleUrl: "./sesiones.component.css",
})
export class SesionesComponent {
  readonly LaptopIcon = Laptop;
  readonly SmartphoneIcon = Smartphone;
  readonly LogOutIcon = LogOut;
  readonly ShieldCheckIcon = ShieldCheck;
  readonly CheckCircleIcon = CircleCheck;

  isOtherRevoked = false;

  sessions: SessionItem[] = [
    {
      id: "s1",
      device: "Navegador Web (Linux / Chrome)",
      browser: "Chrome 150.0",
      ip: "192.0.2.42",
      location: "Alta Gracia, Córdoba, Argentina",
      lastActive: "Sesión Actual (Activa)",
      isCurrent: true,
    },
    {
      id: "s2",
      device: "Dispositivo Móvil (Android)",
      browser: "Chrome Mobile 149.0",
      ip: "192.0.2.99",
      location: "Córdoba Capital, Argentina",
      lastActive: "Hace 2 días",
      isCurrent: false,
    },
  ];

  revokeOtherSessions(): void {
    this.sessions = this.sessions.filter((s) => s.isCurrent);
    this.isOtherRevoked = true;
    setTimeout(() => (this.isOtherRevoked = false), 4000);
  }
}
