import { TestBed } from "@angular/core/testing";
import { provideHttpClient } from "@angular/common/http";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import {
  HumanPromotionError,
  HumanPromotionService,
} from "./human-promotion.service";
import {
  EntradaActividad,
  HitoInput,
} from "../core/models/human-promotion.model";

const base = "/api/v1/human-promotion";

const actividad: EntradaActividad = {
  id: "i1",
  tipo: "inscripcion",
  fecha: "2026-09-02",
  descripcion: "Se inscribió en un curso",
  curso_id: "c1",
};

describe("HumanPromotionService", () => {
  let service: HumanPromotionService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(HumanPromotionService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it("busca participantes con la cookie de sesión", () => {
    service.buscarParticipantes("ana").subscribe();
    const req = http.expectOne((r) => r.url.endsWith(`${base}/participantes`));
    expect(req.request.params.get("q")).toBe("ana");
    expect(req.request.withCredentials).toBeTrue();
    req.flush([]);
  });

  it("lee el log de auditoría y lo pasa a la forma común (HU-35)", () => {
    service
      .listarAuditoria({ desde: "2026-10-01", accion: "" })
      .subscribe((rows) => {
        expect(rows).toEqual([
          {
            id: "1",
            usuarioId: "u1",
            usuarioNombre: "Ana Paz",
            accion: "actualizar",
            recurso: "ficha",
            recursoId: "f1",
            fecha: "2026-10-20T15:00:00Z",
            ip: undefined,
          },
        ]);
      });
    const req = http.expectOne((r) => r.url.endsWith(`${base}/auditoria`));
    expect(req.request.params.get("desde")).toBe("2026-10-01");
    expect(req.request.params.has("accion")).toBeFalse();
    expect(req.request.withCredentials).toBeTrue();
    req.flush([
      {
        id: "1",
        usuario_id: "u1",
        usuario_nombre: "Ana Paz",
        accion: "actualizar",
        recurso: "ficha",
        recurso_id: "f1",
        fecha: "2026-10-20T15:00:00Z",
      },
    ]);
  });

  it("pide la ficha por participante y distingue el 404 (sin ficha)", () => {
    let error: HumanPromotionError | undefined;
    service.getFichaPorParticipante("p1").subscribe({
      error: (e: HumanPromotionError) => (error = e),
    });
    const req = http.expectOne((r) => r.url.endsWith(`${base}/fichas`));
    expect(req.request.params.get("participante_id")).toBe("p1");
    req.flush(
      { error: "recurso no encontrado" },
      { status: 404, statusText: "Not Found" },
    );
    expect(error?.status).toBe(404);
    expect(error?.message).toBe("recurso no encontrado");
  });

  it("crea y edita hitos en la ficha", () => {
    const input: HitoInput = { fecha: "2026-09-10", descripcion: "Taller" };
    service.crearHito("f1", input).subscribe();
    const alta = http.expectOne(`${base}/fichas/f1/hitos`);
    expect(alta.request.method).toBe("POST");
    expect(alta.request.body).toEqual(input);
    alta.flush({});

    service.editarHito("f1", "h 1", input).subscribe();
    const edicion = http.expectOne(`${base}/fichas/f1/hitos/h%201`);
    expect(edicion.request.method).toBe("PUT");
    edicion.flush({});
  });

  it("lista la actividad en el centro de la ficha", () => {
    service
      .listarActividad("f1")
      .subscribe((a) => expect(a).toEqual([actividad]));
    const req = http.expectOne(`${base}/fichas/f1/actividad`);
    expect(req.request.method).toBe("GET");
    expect(req.request.withCredentials).toBeTrue();
    req.flush([actividad]);
  });

  it("mi línea de tiempo trae hitos y actividad", () => {
    service.miLineaDeTiempo().subscribe((linea) => {
      expect(linea.hitos.length).toBe(1);
      expect(linea.actividad).toEqual([actividad]);
    });
    http
      .expectOne(`${base}/mi-linea-de-tiempo`)
      .flush({
        hitos: [{ id: "h1", fecha: "2026-09-01", descripcion: "x" }],
        actividad: [actividad],
      });
  });

  it("mi línea de tiempo tolera una respuesta sin actividad (backend anterior)", () => {
    service.miLineaDeTiempo().subscribe((linea) => {
      expect(linea.hitos).toEqual([]);
      expect(linea.actividad).toEqual([]);
    });
    http.expectOne(`${base}/mi-linea-de-tiempo`).flush({ hitos: null });
  });

  it("descarga el PDF como blob", () => {
    service
      .descargarPdf("f1")
      .subscribe((b) => expect(b instanceof Blob).toBeTrue());
    const req = http.expectOne(`${base}/fichas/f1/pdf`);
    expect(req.request.responseType).toBe("blob");
    req.flush(new Blob(["%PDF"]));
  });

  it("traduce los errores sin cuerpo a mensajes en español", () => {
    const mensajes: Record<number, string> = {
      0: "No se pudo conectar",
      401: "Tu sesión expiró",
      403: "No tenés permisos",
    };
    for (const [status, texto] of Object.entries(mensajes)) {
      let error: HumanPromotionError | undefined;
      service.listarHitos("f1").subscribe({ error: (e) => (error = e) });
      const req = http.expectOne(`${base}/fichas/f1/hitos`);
      if (Number(status) === 0) {
        req.error(new ProgressEvent("error"), { status: 0 });
      } else {
        req.flush(null, { status: Number(status), statusText: "x" });
      }
      expect(error?.message).toContain(texto);
    }
  });
});
