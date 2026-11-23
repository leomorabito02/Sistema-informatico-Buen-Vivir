import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick,
} from "@angular/core/testing";
import { of } from "rxjs";
import { JobOfferService } from "../../../services/job-offer.service";
import {
  OfertasComponent,
  camposFaltantesOferta,
  erroresOferta,
  mensajeErrorOferta,
} from "./ofertas.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { provideRouter } from "@angular/router";
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
      { id: "of-1", titulo: "Cocinero", estado: "Pendiente" } as any,
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

  // Los obligatorios se marcan con * y, al intentar publicar, cada campo vacío muestra su
  // error al lado; no se llama al backend.
  it("marca los obligatorios y muestra el error junto a cada campo", () => {
    const service = TestBed.inject(JobOfferService);
    const crear = spyOn(service, "createOffer").and.returnValue(of({} as any));
    component.openNewJobDialog();
    fixture.detectChanges();

    const dialogo = () =>
      document.body.querySelector(".p-dialog") as HTMLElement;
    expect(dialogo().textContent).toContain(
      "Todos los campos son obligatorios",
    );
    expect(dialogo().querySelectorAll(".req").length).toBeGreaterThanOrEqual(9);
    // Recién abierto no hay errores en rojo.
    expect(dialogo().querySelectorAll(".field-error").length).toBe(0);

    component.saveJob();
    fixture.detectChanges();

    expect(crear).not.toHaveBeenCalled();
    const errores = Array.from(dialogo().querySelectorAll(".field-error")).map(
      (e) => e.textContent?.trim(),
    );
    expect(errores).toContain("Ingresá el título del puesto.");
    expect(errores).toContain("Agregá al menos un requisito.");
    expect(errores).toContain("Elegí la fecha límite.");
    expect(dialogo().querySelector("#title")?.classList).toContain("p-invalid");
  });

  it("toma el requisito escrito aunque no se haya confirmado con Enter", () => {
    component.openNewJobDialog();
    component.currentRequisito = "Atención al cliente";
    component.saveJob();
    expect(component.newJob.requisitos).toEqual(["Atención al cliente"]);
    expect(component.errores.requisitos).toBeUndefined();
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
        requisitos: [],
        fecha_limite: "",
      }),
    ).toEqual(["salario", "al menos un requisito", "fecha límite"]);
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
      ].sort(),
    );
  });

  it("al publicar, la fecha límite no puede haber pasado", () => {
    const vieja = { ...completa, fecha_limite: "2020-01-01" };
    expect(erroresOferta(vieja, { exigirFechaFutura: true }).fecha_limite).toBe(
      "La fecha límite ya pasó.",
    );
    // Al editar no se exige (el backend tampoco lo hace).
    expect(erroresOferta(vieja).fecha_limite).toBeUndefined();
  });
});
