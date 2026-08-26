import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  LucideAngularModule,
  BookOpen,
  Briefcase,
  Heart,
  Megaphone,
} from "lucide-angular";

import { CardModule } from "primeng/card";

@Component({
  selector: "app-pillars",
  standalone: true,
  imports: [CommonModule, LucideAngularModule, CardModule],
  templateUrl: "./pillars.component.html",
  styleUrl: "./pillars.component.css",
})
export class PillarsComponent {
  readonly BookOpenIcon = BookOpen;
  readonly BriefcaseIcon = Briefcase;
  readonly HeartIcon = Heart;
  readonly MegaphoneIcon = Megaphone;
}
