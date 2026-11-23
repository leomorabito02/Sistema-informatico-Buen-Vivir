import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick,
} from "@angular/core/testing";
import { of } from "rxjs";
import { ConfirmationService } from "primeng/api";
import { JobOffer, JobOfferService } from "../../../services/job-offer.service";
import {
  OfertasComponent,
  camposFaltantesOferta,
  erroresOferta,
  mensajeErrorOferta,
} from "./ofertas.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import {
  ActivatedRoute,
  convertToParamMap,
  provideRouter,
} from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("OfertasComponent", () => {
  let component: OfertasComponent;
  let fixture: ComponentFixture<OfertasComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OfertasComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(OfertasComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("asks labor-integration for the search text after typing stops", fakeAsync(() => {
    const service = TestBed.inject(JobOfferService);
    const getOffers = spyOn(service, "getOffers").and.returnValue(of([]));

    component.searchTerm = "coc";
    component.buscar();
    component.searchTerm = "cocina";
    component.buscar();
    tick(299);
    expect(getOffers).not.toHaveBeenCalled();

    tick(1);
    expect(getOffers).toHaveBeenCalledOnceWith("cocina");
  }));

  it("shows approve and reject only to the administrator", () => {
    component.jobs = [
      {
        id: "of-1",
        titulo: "Cocinero",
        estado: "Pendiente",
        vacantes: 1,
      } as JobOffer,
    ];
    component.isAdminOrVolunteer = true;
    component.isAdmin = false;
    component.filterJobs();
    fixture.detectChanges();
    const labels = () =>
      Array.from(
        fixture.nativeElement.querySelectorAll(
          "p-button",
        ) as NodeListOf<HTMLElement>,
      ).map((b) => b.getAttribute("label") ?? b.textContent?.trim());
    expect(labels()).not.toContain("Aprobar");

    component.isAdmin = true;
    fixture.detectChanges();
    expect(labels()).toContain("Aprobar");
    expect(labels()).toContain("Rechazar");
  });

  it("marca los obligatorios y muestra el error junto a cada campo", () => {
    const service = TestBed.inject(JobOfferService);
    const crear = spyOn(service, "createOffer").and.returnValue(
      of({} as JobOffer),
    );
    component.openNewJobDialog();
    fixture.detectChanges();

    const dialogo = () =>
      document.body.querySelector(".p-dialog") as HTMLElement;
    expect(dialogo().textContent).toContain(
      "Todos los campos son obligatorios",
    );
    expect(dialogo().querySelectorAll(".req").length).toBeGreaterThanOrEqual(
      10,
    );
    expect(dialogo().querySelectorAll(".field-error").length).toBe(0);

    component.newJob.vacantes = 0;
    component.saveJob();
    fixture.detectChanges();

    expect(crear).not.toHaveBeenCalled();
    const errores = Array.from(dialogo().querySelectorAll(".field-error")).map(
      (e) => e.textContent?.trim(),
    );
    expect(errores).toContain("Ingresá el título del puesto.");
    expect(errores).toContain("Indicá al menos 1 vacante.");
    expect(errores).toContain("Agregá al menos un requisito.");
    expect(errores).toContain("Elegí la fecha límite.");
    expect(dialogo().querySelector("#title")?.classList).toContain("p-invalid");
    expect(dialogo().querySelector("#vacantes")?.classList).toContain(
      "p-invalid",
    );
  });

  it("toma el requisito escrito aunque no se haya confirmado con Enter", () => {
    component.openNewJobDialog();
    component.currentRequisito = "Atención al cliente";
    component.saveJob();
    expect(component.newJob.requisitos).toEqual(["Atención al cliente"]);
    expect(component.errores.requisitos).toBeUndefined();
  });

  it("excludes expired and closed offers from participant Disponibles but keeps applied ones in Mis postulaciones", () => {
    // Arrange
    component.isParticipant = true;
    component.jobs = [
      {
        id: "of-act",
        empresa_id: "c1",
        titulo: "Activa",
        estado: "Activa",
        vacantes: 2,
      } as JobOffer,
      {
        id: "of-ven",
        empresa_id: "c1",
        titulo: "Vencida",
        estado: "Vencida",
        vacantes: 1,
      } as JobOffer,
      {
        id: "of-cer",
        empresa_id: "c1",
        titulo: "Cerrada",
        estado: "Cerrada",
        vacantes: 1,
      } as JobOffer,
    ];
    component.appliedJobs = new Set(["of-cer"]);
    component.myApplicationsMap.set("of-cer", "Aceptado");

    // Act 1: Disponibles
    component.setTab("Disponibles");
    expect(component.filteredJobs.map((j) => j.id)).toEqual(["of-act"]);

    // Act 2: Mis postulaciones
    component.setTab("Postuladas");
    expect(component.filteredJobs.map((j) => j.id)).toEqual(["of-cer"]);
  });

  it("separates vigentes from Finalizadas e historial and loads follow-up details for company", () => {
    // Arrange
    const service = TestBed.inject(JobOfferService);
    const getApps = spyOn(service, "getApplicationsForOffer").and.returnValue(
      of([
        {
          id: "a1",
          oferta_id: "of-ven",
          participante_id: "p1",
          nombre: "Ana Paz",
          estado: "Aceptado",
        },
        {
          id: "a2",
          oferta_id: "of-ven",
          participante_id: "p2",
          nombre: "Luis Gómez",
          estado: "Pendiente",
        },
        {
          id: "a3",
          oferta_id: "of-ven",
          participante_id: "p3",
          nombre: "María Sol",
          estado: "Rechazado",
        },
      ]),
    );
    component.isCompany = true;
    component.jobs = [
      {
        id: "of-act",
        empresa_id: "c1",
        titulo: "Panadero",
        estado: "Activa",
        vacantes: 2,
      } as JobOffer,
      {
        id: "of-ven",
        empresa_id: "c1",
        titulo: "Cajero",
        estado: "Vencida",
        vacantes: 2,
      } as JobOffer,
      {
        id: "of-cer",
        empresa_id: "c1",
        titulo: "Cadete",
        estado: "Cerrada",
        vacantes: 1,
      } as JobOffer,
    ];

    // Act 1: Vigentes section shows only Activa/Pendiente
    component.setSection("vigentes");
    fixture.detectChanges();
    expect(component.filteredJobs.map((j) => j.id)).toEqual(["of-act"]);

    // Act 2: Historial section shows Vencida/Cerrada and fetches applicants
    component.setSection("historial");
    fixture.detectChanges();
    expect(component.filteredJobs.map((j) => j.id)).toEqual([
      "of-ven",
      "of-cer",
    ]);
    expect(getApps).toHaveBeenCalledWith("of-ven");
    expect(getApps).toHaveBeenCalledWith("of-cer");

    // Assert: follow-up highlights who to contact (center) and accepted/pending applicants
    const html = fixture.nativeElement.textContent as string;
    expect(html).toContain("A quién contactar");
    expect(html).toContain("ceibuenvivir7@gmail.com");
    expect(html).toContain("Sobre cuáles postulantes preguntar");
    expect(html).toContain("Ana Paz");
    expect(html).toContain("Luis Gómez");
    expect(html).not.toContain("María Sol");
  });

  it("shows company email and phone with mailto link for staff in Finalizadas e historial", () => {
    // Arrange
    const service = TestBed.inject(JobOfferService);
    spyOn(service, "getApplicationsForOffer").and.returnValue(
      of([
        {
          id: "a1",
          oferta_id: "of-cer",
          participante_id: "p1",
          nombre: "Juan Pérez",
          estado: "Pendiente",
        },
      ]),
    );
    component.isAdminOrVolunteer = true;
    component.isAdmin = false;
    component.jobs = [
      {
        id: "of-cer",
        empresa_id: "c1",
        creador_nombre: "Vivero Sur",
        empresa_email: "rrhh@viverosur.com",
        empresa_telefono: "3547123456",
        titulo: "Jardinero",
        estado: "Cerrada",
        vacantes: 1,
      } as JobOffer,
    ];

    // Act
    component.setSection("historial");
    fixture.detectChanges();

    // Assert
    const html = fixture.nativeElement.textContent as string;
    expect(html).toContain("Vivero Sur");
    expect(html).toContain("rrhh@viverosur.com");
    expect(html).toContain("3547123456");
    expect(html).toContain("Juan Pérez");
    const mailLink = fixture.nativeElement.querySelector(
      'a[href^="mailto:"]',
    ) as HTMLAnchorElement;
    expect(mailLink.href).toContain("rrhh%40viverosur.com");
  });

  it("allows role_company to close an active offer early after confirmation", () => {
    // Arrange
    const service = TestBed.inject(JobOfferService);
    const confirmSvc = fixture.debugElement.injector.get(ConfirmationService);
    spyOn(confirmSvc, "confirm").and.callFake((opts: any) => opts.accept?.());
    const closeSpy = spyOn(service, "closeOffer").and.returnValue(
      of({ id: "of-1", estado: "Cerrada" } as JobOffer),
    );
    const reloadSpy = spyOn(service, "getOffers").and.returnValue(of([]));

    const offer = {
      id: "of-1",
      empresa_id: "c1",
      titulo: "Cocinero",
      estado: "Activa",
      vacantes: 2,
    } as JobOffer;
    component.isCompany = true;
    component.jobs = [offer];
    component.filterJobs();
    fixture.detectChanges();

    // Act
    component.confirmCloseOffer(offer);

    // Assert
    expect(closeSpy).toHaveBeenCalledOnceWith("of-1");
    expect(reloadSpy).toHaveBeenCalled();
  });
});

describe("OfertasComponent queryParams", () => {
  it("opens directly in Finalizadas e historial when seccion=historial", async () => {
    await TestBed.configureTestingModule({
      imports: [OfertasComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
        {
          provide: ActivatedRoute,
          useValue: {
            queryParamMap: of(convertToParamMap({ seccion: "historial" })),
          },
        },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(OfertasComponent);
    const component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.activeSection).toBe("historial");
  });
});

describe("camposFaltantesOferta (HU-39)", () => {
  const completa = {
    titulo: "Cajero",
    descripcion: "Atención al público",
    ubicacion: "Alta Gracia",
    modalidad: "Presencial",
    tipo_empleo: "Medio tiempo",
    salario: "A convenir",
    horario: "9 a 13",
    vacantes: 2,
    requisitos: ["Atención al cliente"],
    fecha_limite: "2026-12-31",
  };

  it("no marca nada con la oferta completa", () => {
    expect(camposFaltantesOferta(completa)).toEqual([]);
  });

  it("lista los campos que faltan", () => {
    expect(
      camposFaltantesOferta({
        ...completa,
        salario: " ",
        vacantes: 0,
        requisitos: [],
        fecha_limite: "",
      }),
    ).toEqual([
      "salario",
      "vacantes (al menos 1)",
      "al menos un requisito",
      "fecha límite",
    ]);
  });

  it("muestra el mensaje del backend en un 400", () => {
    expect(
      mensajeErrorOferta(
        { status: 400, error: "falta el campo obligatorio: horario\n" },
        "x",
      ),
    ).toBe("falta el campo obligatorio: horario");
    expect(mensajeErrorOferta({ status: 500, error: "boom" }, "fallback")).toBe(
      "fallback",
    );
  });
});

describe("erroresOferta (mismos criterios que el backend)", () => {
  const completa = {
    titulo: "Cajero",
    descripcion: "Atención al público",
    ubicacion: "Alta Gracia",
    modalidad: "Presencial",
    tipo_empleo: "Medio tiempo",
    salario: "A convenir",
    horario: "9 a 13",
    vacantes: 1,
    requisitos: ["Atención al cliente"],
    fecha_limite: "2999-12-31",
  };

  it("sin errores con la oferta completa", () => {
    expect(erroresOferta(completa, { exigirFechaFutura: true })).toEqual({});
  });

  it("un error por cada campo vacío", () => {
    const e = erroresOferta({});
    expect(Object.keys(e).sort()).toEqual(
      [
        "descripcion",
        "fecha_limite",
        "horario",
        "modalidad",
        "requisitos",
        "salario",
        "tipo_empleo",
        "titulo",
        "ubicacion",
        "vacantes",
      ].sort(),
    );
  });

  it("al publicar, la fecha límite no puede haber pasado", () => {
    const vieja = { ...completa, fecha_limite: "2020-01-01" };
    expect(erroresOferta(vieja, { exigirFechaFutura: true }).fecha_limite).toBe(
      "La fecha límite ya pasó.",
    );
    expect(erroresOferta(vieja).fecha_limite).toBeUndefined();
  });
});
