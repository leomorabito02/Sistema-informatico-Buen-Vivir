import { Component, OnInit, PLATFORM_ID, inject } from "@angular/core";
import { CommonModule, isPlatformBrowser } from "@angular/common";
import { RouterLink } from "@angular/router";
import { FormsModule, ReactiveFormsModule, FormBuilder, Validators } from "@angular/forms";
import {
  LucideAngularModule,
  Mail,
  Phone,
  MapPin,
  Clock,
  Send,
  CircleCheck,
  ShieldCheck,
  ArrowLeft,
} from "lucide-angular";
import { FooterComponent } from "../../components/layout/footer/footer.component";

@Component({
  selector: "app-contacto",
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    FormsModule,
    ReactiveFormsModule,
    LucideAngularModule,
    FooterComponent,
  ],
  templateUrl: "./contacto.component.html",
  styleUrl: "./contacto.component.css",
})
export class ContactoComponent implements OnInit {
  readonly MailIcon = Mail;
  readonly PhoneIcon = Phone;
  readonly MapPinIcon = MapPin;
  readonly ClockIcon = Clock;
  readonly SendIcon = Send;
  readonly CheckCircleIcon = CircleCheck;
  readonly ShieldCheckIcon = ShieldCheck;
  readonly ArrowLeftIcon = ArrowLeft;

  private readonly fb = inject(FormBuilder);
  private readonly platformId = inject(PLATFORM_ID);

  isSubmitted = false;

  contactForm = this.fb.group({
    name: ["", [Validators.required, Validators.minLength(2)]],
    email: ["", [Validators.required, Validators.email]],
    subject: ["", [Validators.required]],
    message: ["", [Validators.required, Validators.minLength(10)]],
    privacyAccepted: [false, [Validators.requiredTrue]],
  });

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      window.scrollTo(0, 0);
    }
  }

  onSubmit(): void {
    if (this.contactForm.valid) {
      this.isSubmitted = true;
    } else {
      this.contactForm.markAllAsTouched();
    }
  }

  resetForm(): void {
    this.contactForm.reset();
    this.isSubmitted = false;
  }
}
