import { Component, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  FormsModule,
  ReactiveFormsModule,
  FormBuilder,
  Validators,
} from "@angular/forms";
import {
  LucideAngularModule,
  Lock,
  ShieldCheck,
  Key,
  AlertCircle,
  CheckCircle2,
} from "lucide-angular";

@Component({
  selector: "app-seguridad",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    LucideAngularModule,
  ],
  templateUrl: "./seguridad.component.html",
  styleUrl: "./seguridad.component.css",
})
export class SeguridadComponent {
  readonly LockIcon = Lock;
  readonly ShieldCheckIcon = ShieldCheck;
  readonly KeyIcon = Key;
  readonly AlertCircleIcon = AlertCircle;
  readonly CheckCircleIcon = CheckCircle2;

  private fb = inject(FormBuilder);

  isPasswordUpdated = false;

  passwordForm = this.fb.group({
    currentPassword: ["", [Validators.required]],
    newPassword: ["", [Validators.required, Validators.minLength(8)]],
    confirmPassword: ["", [Validators.required]],
  });

  onSubmit(): void {
    if (this.passwordForm.valid) {
      if (
        this.passwordForm.value.newPassword ===
        this.passwordForm.value.confirmPassword
      ) {
        this.isPasswordUpdated = true;
        this.passwordForm.reset();
        setTimeout(() => (this.isPasswordUpdated = false), 4000);
      } else {
        this.passwordForm.get("confirmPassword")?.setErrors({ mismatch: true });
      }
    } else {
      this.passwordForm.markAllAsTouched();
    }
  }
}
