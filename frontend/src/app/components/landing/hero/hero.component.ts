import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterLink } from "@angular/router";
import {
  LucideAngularModule,
  Briefcase,
  LogIn,
  Megaphone,
} from "lucide-angular";

import { ButtonModule } from "primeng/button";

@Component({
  selector: "app-hero",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule, ButtonModule],
  templateUrl: "./hero.component.html",
  styleUrl: "./hero.component.css",
})
export class HeroComponent {
  readonly BriefcaseIcon = Briefcase;
  readonly LogInIcon = LogIn;
  readonly MegaphoneIcon = Megaphone;
}
