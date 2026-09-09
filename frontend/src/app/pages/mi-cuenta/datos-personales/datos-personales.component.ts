import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule, ReactiveFormsModule, FormBuilder, Validators } from "@angular/forms";
import { LucideAngularModule, Save, CheckCircle2, User, Mail, Phone, MapPin } from "lucide-angular";

@Component({
  selector: "app-datos-personales",
  standalone: true,
  imports: [CommonModule, FormsModule, ReactiveFormsModule, LucideAngularModule],
  templateUrl: "./datos-personales.component.html",
  styleUrl: "./datos-personales.component.css",
})
export class DatosPersonalesComponent implements OnInit {
  readonly SaveIcon = Save;
  readonly CheckCircleIcon = CheckCircle2;
  readonly UserIcon = User;
  readonly MailIcon = Mail;
  readonly PhoneIcon = Phone;
  readonly MapPinIcon = MapPin;

  private fb = inject(FormBuilder);

  isSaved = false;

  profileForm = this.fb.group({
    fullName: ["Juan Pérez", [Validators.required, Validators.minLength(3)]],
    dni: ["38123456", [Validators.required, Validators.pattern("^[0-9]{7,8}$")]],
    email: ["juan.perez@ejemplo.com", [Validators.required, Validators.email]],
    phone: ["+54 9 3547 612118", [Validators.required]],
    city: ["Alta Gracia", [Validators.required]],
    address: ["Barrio La Otra Orilla", [Validators.required]],
  });

  ngOnInit(): void {}

  onSubmit(): void {
    if (this.profileForm.valid) {
      this.isSaved = true;
      setTimeout(() => (this.isSaved = false), 4000);
    } else {
      this.profileForm.markAllAsTouched();
    }
  }
}
