import { ComponentFixture, TestBed } from "@angular/core/testing";
import {
  ActivatedRoute,
  convertToParamMap,
  provideRouter,
} from "@angular/router";
import { provideHttpClient } from "@angular/common/http";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import { ContactoComponent } from "./contacto.component";

describe("ContactoComponent", () => {
  let component: ContactoComponent;
  let fixture: ComponentFixture<ContactoComponent>;
  let http: HttpTestingController;

  function completar() {
    component.contactForm.setValue({
      name: "Ana Pérez",
      email: "ana@example.com",
      subject: "cursos",
      message: "Quiero saber cuándo empieza el taller.",
      privacyAccepted: true,
    });
  }

  beforeEach(async () => {
    localStorage.removeItem("isAuthenticated");
    await TestBed.configureTestingModule({
      imports: [ContactoComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ContactoComponent);
    component = fixture.componentInstance;
    http = TestBed.inject(HttpTestingController);
    fixture.detectChanges();
  });

  afterEach(() => {
    http.verify();
    localStorage.removeItem("isAuthenticated");
  });

  it("muestra el título", () => {
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.querySelector(".legal-title")?.textContent).toContain(
      "Contacto e información institucional",
    );
  });

  it("envía la consulta al backend y recién ahí muestra el éxito", () => {
    completar();
    component.onSubmit();
    expect(component.isSubmitted).toBeFalse();
    expect(component.enviando).toBeTrue();

    const req = http.expectOne((r) =>
      r.url.endsWith("/api/v1/notifications/contacto"),
    );
    expect(req.request.method).toBe("POST");
    expect(req.request.body).toEqual({
      nombre: "Ana Pérez",
      email: "ana@example.com",
      asunto: "Consulta sobre cursos",
      mensaje: "Quiero saber cuándo empieza el taller.",
      acepta_privacidad: true,
    });
    req.flush({ enviada: true }, { status: 202, statusText: "Accepted" });

    expect(component.isSubmitted).toBeTrue();
    expect(component.enviando).toBeFalse();
  });

  it("si el backend falla no muestra éxito y explica el error", () => {
    completar();
    component.onSubmit();
    http
      .expectOne((r) => r.url.endsWith("/api/v1/notifications/contacto"))
      .flush(
        { error: "Enviaste varias consultas seguidas." },
        { status: 429, statusText: "Too Many Requests" },
      );
    expect(component.isSubmitted).toBeFalse();
    expect(component.errorEnvio).toBe("Enviaste varias consultas seguidas.");
  });

  it("con datos inválidos no envía nada", () => {
    component.onSubmit();
    http.expectNone((r) => r.url.endsWith("/api/v1/notifications/contacto"));
    expect(component.contactForm.touched).toBeTrue();
  });

  it("'Volver al inicio' lleva al panel si hay sesión y a la landing si no", () => {
    expect(component.inicio).toBe("/");
    localStorage.setItem("isAuthenticated", "true");
    const otra = TestBed.createComponent(ContactoComponent).componentInstance;
    expect(otra.inicio).toBe("/dashboard");
  });

  it("preloads valid subject and message from queryParams", async () => {
    // Arrange
    TestBed.resetTestingModule();
    await TestBed.configureTestingModule({
      imports: [ContactoComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              queryParamMap: convertToParamMap({
                asunto: "empleo",
                mensaje: "Consulta sobre la oferta Panadero.",
              }),
            },
          },
        },
      ],
    }).compileComponents();

    // Act
    const prefilledFixture = TestBed.createComponent(ContactoComponent);
    prefilledFixture.detectChanges();

    // Assert
    expect(prefilledFixture.componentInstance.contactForm.value.subject).toBe(
      "empleo",
    );
    expect(prefilledFixture.componentInstance.contactForm.value.message).toBe(
      "Consulta sobre la oferta Panadero.",
    );
  });
});
