import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterLink } from "@angular/router";
import {
  LucideAngularModule,
  Mail,
  Phone,
  MapPin,
  Heart,
  Linkedin,
} from "lucide-angular";

@Component({
  selector: "app-footer",
  standalone: true,
  imports: [CommonModule, RouterLink, LucideAngularModule],
  templateUrl: "./footer.component.html",
  styleUrl: "./footer.component.css",
})
export class FooterComponent {
  readonly MailIcon = Mail;
  readonly PhoneIcon = Phone;
  readonly MapPinIcon = MapPin;
  readonly HeartIcon = Heart;
  readonly LinkedinIcon = Linkedin;
}
