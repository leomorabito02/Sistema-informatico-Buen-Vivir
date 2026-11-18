import { Component, OnInit, PLATFORM_ID, inject } from "@angular/core";
import { CommonModule, isPlatformBrowser } from "@angular/common";
import { RouterLink } from "@angular/router";
import {
  FormsModule,
  ReactiveFormsModule,
  FormBuilder,
  Validators,
} from "@angular/forms";
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
import { ContactoService } from "../../services/contacto.service";
import { rutaInicio } from "../../shared/utils/ruta-inicio";

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
  /** "Volver al inicio": panel si hay sesión, landing si no. */
  readonly inicio = rutaInicio();
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
  private readonly contacto = inject(ContactoService);

  /** Motivos del formulario (el texto es lo que llega en el email). */
  readonly asuntos = [
    { value: "cursos", label: "Consulta sobre cursos" },
    { value: "empleo", label: "Bolsa de empleo e inserción laboral" },
    { value: "voluntariado", label: "Voluntariado y acompañamiento" },
    { value: "datos", label: "Derechos sobre datos personales (ARCO)" },
    { value: "otro", label: "Otro motivo" },
  ];

  isSubmitted = false;
  enviando = false;
  errorEnvio = "";

  contactForm = this.fb.group({
    name: [
      "",
      [Validators.required, Validators.minLength(2), Validators.maxLength(100)],
    ],
    email: ["", [Validators.required, Validators.email]],
    subject: ["", [Validators.required]],
    message: [
      "",
      [
        Validators.required,
        Validators.minLength(10),
        Validators.maxLength(3000),
      ],
    ],
    privacyAccepted: [false, [Validators.requiredTrue]],
  });

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      window.scrollTo(0, 0);
    }
  }

  /**
   * Envía la consulta al centro (antes solo mostraba "enviado" sin mandar
   * nada). El mensaje de éxito aparece recién cuando el backend confirma.
   */
  onSubmit(): void {
    if (this.contactForm.invalid || this.enviando) {
      this.contactForm.markAllAsTouched();
      return;
    }
    const v = this.contactForm.getRawValue();
    const asunto =
      this.asuntos.find((a) => a.value === v.subject)?.label ?? v.subject ?? "";
    this.enviando = true;
    this.errorEnvio = "";
    this.contacto
      .enviar({
        nombre: (v.name ?? "").trim(),
        email: (v.email ?? "").trim(),
        asunto,
        mensaje: (v.message ?? "").trim(),
        acepta_privacidad: !!v.privacyAccepted,
      })
      .subscribe({
        next: () => {
          this.enviando = false;
          this.isSubmitted = true;
        },
        error: (err: Error & { errores?: Record<string, string> }) => {
          this.enviando = false;
          const porCampo = Object.values(err.errores ?? {});
          this.errorEnvio = porCampo.length ? porCampo.join(" ") : err.message;
        },
      });
  }

  resetForm(): void {
    this.contactForm.reset({ privacyAccepted: false });
    this.isSubmitted = false;
    this.errorEnvio = "";
  }
}
