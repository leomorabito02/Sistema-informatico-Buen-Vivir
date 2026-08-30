import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import { LucideAngularModule, MapPin, Clock, Phone } from "lucide-angular";

@Component({
  selector: "app-location",
  standalone: true,
  imports: [CommonModule, LucideAngularModule],
  templateUrl: "./location.component.html",
  styleUrl: "./location.component.css",
})
export class LocationComponent {
  readonly MapPinIcon = MapPin;
  readonly ClockIcon = Clock;
  readonly PhoneIcon = Phone;
}
