import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterLink, RouterLinkActive, RouterOutlet } from "@angular/router";
import {
  LucideAngularModule,
  User,
  Shield,
  Laptop,
  Lock,
  ArrowLeft,
} from "lucide-angular";
import { HeaderComponent } from "../../components/layout/header/header.component";
import { FooterComponent } from "../../components/layout/footer/footer.component";

@Component({
  selector: "app-mi-cuenta",
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    RouterLinkActive,
    RouterOutlet,
    LucideAngularModule,
    HeaderComponent,
    FooterComponent,
  ],
  templateUrl: "./mi-cuenta.component.html",
  styleUrl: "./mi-cuenta.component.css",
})
export class MiCuentaComponent {
  readonly UserIcon = User;
  readonly ShieldIcon = Shield;
  readonly LaptopIcon = Laptop;
  readonly LockIcon = Lock;
  readonly ArrowLeftIcon = ArrowLeft;
}
