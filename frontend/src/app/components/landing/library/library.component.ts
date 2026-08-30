import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterLink } from "@angular/router";
import {
  LucideAngularModule,
  UtensilsCrossed,
  Hammer,
  Sprout,
  Laptop,
  Scissors,
  HeartPulse,
  Briefcase,
  Mic,
  Wheat,
  Leaf,
  Salad,
} from "lucide-angular";

@Component({
  selector: "app-library",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  templateUrl: "./library.component.html",
  styleUrl: "./library.component.css",
})
export class LibraryComponent {
  readonly UtensilsCrossedIcon = UtensilsCrossed;
  readonly HammerIcon = Hammer;
  readonly SproutIcon = Sprout;
  readonly LaptopIcon = Laptop;
  readonly ScissorsIcon = Scissors;
  readonly HeartPulseIcon = HeartPulse;
  readonly BriefcaseIcon = Briefcase;
  readonly MicIcon = Mic;
  readonly WheatIcon = Wheat;
  readonly LeafIcon = Leaf;
  readonly SaladIcon = Salad;
}
