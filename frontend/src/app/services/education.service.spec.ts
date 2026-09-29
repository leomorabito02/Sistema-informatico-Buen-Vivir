import { TestBed } from "@angular/core/testing";
import { EducationService } from "./education.service";
import {
  HttpTestingController,
  provideHttpClientTesting,
} from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";

describe("EducationService", () => {
  let service: EducationService;
  let httpMock: HttpTestingController;
  const getBaseUrl = () => (service as any)["base"] || (service as any).base;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        EducationService,
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });
    service = TestBed.inject(EducationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should fetch health status", () => {
    const mockHealth = {
      status: "ok",
      service: "education",
      appEnv: "test",
      contentfulEnvironment: "master",
    };

    service.health().subscribe((res) => {
      expect(res).toEqual(mockHealth);
    });

    const req = httpMock.expectOne(`${getBaseUrl()}/api/salud`);
    expect(req.request.method).toBe("GET");
    req.flush(mockHealth);
  });

  it("should list courses and get single course", () => {
    const mockCourses: any[] = [{ id: "1", title: "Course 1" }];

    service.listCourses().subscribe((res) => {
      expect(res).toEqual(mockCourses);
    });
    const req1 = httpMock.expectOne(`${getBaseUrl()}/api/cursos`);
    expect(req1.request.method).toBe("GET");
    req1.flush(mockCourses);

    service.getCourse("1").subscribe((res) => {
      expect(res).toEqual(mockCourses[0]);
    });
    const req2 = httpMock.expectOne(`${getBaseUrl()}/api/cursos/1`);
    expect(req2.request.method).toBe("GET");
    req2.flush(mockCourses[0]);
  });

  it("should list draft courses", () => {
    const mockDrafts: any[] = [{ id: "2", title: "Draft 1" }];
    service.listDraftCourses().subscribe((res) => {
      expect(res).toEqual(mockDrafts);
    });
    const req = httpMock.expectOne(`${getBaseUrl()}/api/cursos/borradores`);
    expect(req.request.method).toBe("GET");
    req.flush(mockDrafts);
  });

  it("should list training areas and materials", () => {
    service.listTrainingAreas().subscribe((res) => {
      expect(res).toEqual([{ id: "a1", name: "Area 1" }]);
    });
    const req1 = httpMock.expectOne(`${getBaseUrl()}/api/areas-formativas`);
    req1.flush([{ id: "a1", name: "Area 1" }]);

    service.listMaterials().subscribe((res) => {
      expect(res).toEqual([{ id: "m1", title: "Mat 1" }]);
    });
    const req2 = httpMock.expectOne(`${getBaseUrl()}/api/materiales`);
    req2.flush([{ id: "m1", title: "Mat 1" }]);
  });

  it("should create, update, delete, publish, unpublish course", () => {
    const input: any = { title: "Test Course" };

    service.createCourse(input).subscribe();
    const req1 = httpMock.expectOne(`${getBaseUrl()}/api/cursos`);
    expect(req1.request.method).toBe("POST");
    req1.flush({ id: "1", ...input });

    service.updateCourse("1", input).subscribe();
    const req2 = httpMock.expectOne(`${getBaseUrl()}/api/cursos/1`);
    expect(req2.request.method).toBe("PUT");
    req2.flush({ id: "1", ...input });

    service.deleteCourse("1").subscribe();
    const req3 = httpMock.expectOne(`${getBaseUrl()}/api/cursos/1`);
    expect(req3.request.method).toBe("DELETE");
    req3.flush(null);

    service.publishCourse("1").subscribe();
    const req4 = httpMock.expectOne(`${getBaseUrl()}/api/cursos/1/publicar`);
    expect(req4.request.method).toBe("POST");
    req4.flush({ id: "1", status: "published" });

    service.unpublishCourse("1").subscribe();
    const req5 = httpMock.expectOne(`${getBaseUrl()}/api/cursos/1/despublicar`);
    expect(req5.request.method).toBe("POST");
    req5.flush({ id: "1", status: "draft" });
  });

  it("should handle training areas CRUD", () => {
    const areaInput: any = { name: "Area" };

    service.createTrainingArea(areaInput).subscribe();
    const req1 = httpMock.expectOne(`${getBaseUrl()}/api/areas-formativas`);
    expect(req1.request.method).toBe("POST");
    req1.flush({ id: "a1", ...areaInput });

    service.updateTrainingArea("a1", areaInput).subscribe();
    const req2 = httpMock.expectOne(`${getBaseUrl()}/api/areas-formativas/a1`);
    expect(req2.request.method).toBe("PUT");
    req2.flush({ id: "a1", ...areaInput });

    service.deleteTrainingArea("a1").subscribe();
    const req3 = httpMock.expectOne(`${getBaseUrl()}/api/areas-formativas/a1`);
    expect(req3.request.method).toBe("DELETE");
    req3.flush(null);
  });

  it("should handle materials CRUD and upload", () => {
    const matInput: any = { title: "Mat" };

    service.createMaterial(matInput).subscribe();
    const req1 = httpMock.expectOne(`${getBaseUrl()}/api/materiales`);
    expect(req1.request.method).toBe("POST");
    req1.flush({ id: "m1", ...matInput });

    const mockFile = new File(["content"], "test.pdf", {
      type: "application/pdf",
    });
    service.uploadMaterial(mockFile, "test", "PDF").subscribe();
    const req2 = httpMock.expectOne(`${getBaseUrl()}/api/materiales/upload`);
    expect(req2.request.method).toBe("POST");
    expect(req2.request.body instanceof FormData).toBeTrue();
    req2.flush({ id: "m1", title: "test", type: "PDF" });

    service.updateMaterial("m1", matInput).subscribe();
    const req3 = httpMock.expectOne(`${getBaseUrl()}/api/materiales/m1`);
    expect(req3.request.method).toBe("PUT");
    req3.flush({ id: "m1", ...matInput });

    service.deleteMaterial("m1").subscribe();
    const req4 = httpMock.expectOne(`${getBaseUrl()}/api/materiales/m1`);
    expect(req4.request.method).toBe("DELETE");
    req4.flush(null);
  });

  it("should handle teacher portal endpoints", () => {
    service.listMyCourses().subscribe();
    const req1 = httpMock.expectOne(`${getBaseUrl()}/api/docente/cursos`);
    expect(req1.request.method).toBe("GET");
    req1.flush([]);

    service.listMyDraftCourses().subscribe();
    const req2 = httpMock.expectOne(
      `${getBaseUrl()}/api/docente/cursos/borradores`,
    );
    expect(req2.request.method).toBe("GET");
    req2.flush([]);

    service.updateMyCourse("c1", { title: "My Course" } as any).subscribe();
    const req3 = httpMock.expectOne(`${getBaseUrl()}/api/docente/cursos/c1`);
    expect(req3.request.method).toBe("PUT");
    req3.flush({});

    service.listActivities("c1").subscribe();
    const req4 = httpMock.expectOne(
      `${getBaseUrl()}/api/docente/cursos/c1/actividades`,
    );
    expect(req4.request.method).toBe("GET");
    req4.flush([]);

    service
      .createActivity("c1", {
        titulo: "Act 1",
        consigna: "Consigna",
        fecha: "2026-01-01T00:00:00Z",
      })
      .subscribe();
    const req5 = httpMock.expectOne(
      `${getBaseUrl()}/api/docente/cursos/c1/actividades`,
    );
    expect(req5.request.method).toBe("POST");
    req5.flush({});

    service.listAttendance("a1").subscribe();
    const req6 = httpMock.expectOne(
      `${getBaseUrl()}/api/docente/actividades/a1/asistencia`,
    );
    expect(req6.request.method).toBe("GET");
    req6.flush([]);

    service
      .updateActivity("a1", {
        titulo: "Act 1",
        consigna: "Consigna",
        fecha: "2026-01-02T00:00:00Z",
      })
      .subscribe();
    const req7 = httpMock.expectOne(
      `${getBaseUrl()}/api/docente/actividades/a1`,
    );
    expect(req7.request.method).toBe("PUT");
    req7.flush({});

    service.deleteActivity("a1").subscribe();
    const req7b = httpMock.expectOne(
      `${getBaseUrl()}/api/docente/actividades/a1`,
    );
    expect(req7b.request.method).toBe("DELETE");
    req7b.flush(null);

    service.listActivitySubmissions("a1").subscribe();
    const req7c = httpMock.expectOne(
      `${getBaseUrl()}/api/docente/actividades/a1/entregas`,
    );
    expect(req7c.request.method).toBe("GET");
    req7c.flush([]);

    service.listEnrollments("c1").subscribe();
    const req8 = httpMock.expectOne(
      `${getBaseUrl()}/api/docente/cursos/c1/inscriptos`,
    );
    expect(req8.request.method).toBe("GET");
    req8.flush([]);

    service.getCourseProgress("c1").subscribe();
    const req9 = httpMock.expectOne(
      `${getBaseUrl()}/api/docente/cursos/c1/progreso`,
    );
    expect(req9.request.method).toBe("GET");
    req9.flush([]);
  });

  describe("Error Handling", () => {
    it("should handle 401 Unauthorized", () => {
      service.listCourses().subscribe({
        next: () => fail("expected error"),
        error: (err: Error) => {
          expect(err.message).toBe(
            "Tu sesión expiró o no es válida. Volvé a iniciar sesión.",
          );
        },
      });
      const req = httpMock.expectOne(`${getBaseUrl()}/api/cursos`);
      req.flush(null, { status: 401, statusText: "Unauthorized" });
    });

    it("should handle 403 Forbidden", () => {
      service.listCourses().subscribe({
        next: () => fail("expected error"),
        error: (err: Error) => {
          expect(err.message).toBe(
            "No tenés permisos para esta acción en Educación.",
          );
        },
      });
      const req = httpMock.expectOne(`${getBaseUrl()}/api/cursos`);
      req.flush(null, { status: 403, statusText: "Forbidden" });
    });

    it("should handle 404 Not Found", () => {
      service.listCourses().subscribe({
        next: () => fail("expected error"),
        error: (err: Error) => {
          expect(err.message).toBe(
            "El recurso solicitado no existe o no está publicado.",
          );
        },
      });
      const req = httpMock.expectOne(`${getBaseUrl()}/api/cursos`);
      req.flush(null, { status: 404, statusText: "Not Found" });
    });

    it("should handle 502 Bad Gateway", () => {
      service.listCourses().subscribe({
        next: () => fail("expected error"),
        error: (err: Error) => {
          expect(err.message).toBe(
            "El módulo de Educación no pudo responder (Contentful/Firestore).",
          );
        },
      });
      const req = httpMock.expectOne(`${getBaseUrl()}/api/cursos`);
      req.flush(null, { status: 502, statusText: "Bad Gateway" });
    });

    it("should handle default generic error", () => {
      service.listCourses().subscribe({
        next: () => fail("expected error"),
        error: (err: Error) => {
          expect(err.message).toBe(
            "Ocurrió un error al comunicarse con el módulo de Educación.",
          );
        },
      });
      const req = httpMock.expectOne(`${getBaseUrl()}/api/cursos`);
      req.flush(null, { status: 500, statusText: "Internal Error" });
    });

    it("should handle JSON error body", () => {
      service.listCourses().subscribe({
        next: () => fail("expected error"),
        error: (err: Error) => {
          expect(err.message).toBe("Custom error message");
        },
      });
      const req = httpMock.expectOne(`${getBaseUrl()}/api/cursos`);
      req.flush(
        { error: "Custom error message" },
        { status: 400, statusText: "Bad Request" },
      );
    });
  });

  it("should call the admin/volunteer activity and submission endpoints", () => {
    const input = {
      titulo: "Act",
      consigna: "Hacer",
      fecha: "2026-10-05T00:00:00Z",
    };

    service.updateCourseActivity("a1", input).subscribe();
    const put = httpMock.expectOne(`${getBaseUrl()}/api/actividades/a1`);
    expect(put.request.method).toBe("PUT");
    expect(put.request.body).toEqual(input);
    put.flush({});

    service.deleteCourseActivity("a1").subscribe();
    const del = httpMock.expectOne(`${getBaseUrl()}/api/actividades/a1`);
    expect(del.request.method).toBe("DELETE");
    del.flush(null);

    service.listCourseActivitySubmissions("a1").subscribe();
    const entregas = httpMock.expectOne(
      `${getBaseUrl()}/api/actividades/a1/entregas`,
    );
    expect(entregas.request.method).toBe("GET");
    entregas.flush([]);
  });

  it("should call the participant endpoints", () => {
    service.listMyEnrolledCourses().subscribe();
    httpMock.expectOne(`${getBaseUrl()}/api/participante/cursos`).flush([]);

    service.getMyCourse("c1").subscribe();
    httpMock
      .expectOne(`${getBaseUrl()}/api/participante/cursos/c1`)
      .flush({ curso: {}, actividades: [], progreso: {} });
  });

  it("should submit an activity as multipart with text and files", () => {
    const file = new File(["contenido"], "tp.pdf", { type: "application/pdf" });
    service.submitActivity("a1", "mi respuesta", [file]).subscribe();

    const req = httpMock.expectOne(
      `${getBaseUrl()}/api/participante/actividades/a1/entrega`,
    );
    expect(req.request.method).toBe("POST");
    const body = req.request.body as FormData;
    expect(body.get("texto")).toBe("mi respuesta");
    expect((body.getAll("archivos")[0] as File).name).toBe("tp.pdf");
    req.flush({ id: "a1_p1" });
  });

  it("should download a submission file as a blob", () => {
    let blob: Blob | undefined;
    service.downloadSubmissionFile("a1_p1", 0).subscribe((b) => (blob = b));

    const req = httpMock.expectOne(
      `${getBaseUrl()}/api/entregas/a1_p1/archivos/0`,
    );
    expect(req.request.responseType).toBe("blob");
    req.flush(new Blob(["%PDF"]));
    expect(blob).toBeDefined();
  });

  it("should explain a 413 as a size problem", () => {
    let message = "";
    service
      .submitActivity("a1", "", [])
      .subscribe({ error: (e: Error) => (message = e.message) });
    httpMock
      .expectOne(`${getBaseUrl()}/api/participante/actividades/a1/entrega`)
      .flush(null, { status: 413, statusText: "Payload Too Large" });
    expect(message).toContain("tamaño");
  });
});
