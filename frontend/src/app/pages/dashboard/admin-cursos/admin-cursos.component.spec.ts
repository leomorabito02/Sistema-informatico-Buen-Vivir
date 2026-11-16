import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick,
} from "@angular/core/testing";
import {
  AdminCursosComponent,
  mensajeBorrarMaterial,
} from "./admin-cursos.component";
import { EducationService } from "../../../services/education.service";
import { UsersService } from "../../../services/users.service";
import { MessageService, ConfirmationService } from "primeng/api";
import { of, throwError } from "rxjs";
import { provideHttpClient } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { Auth } from "@angular/fire/auth";

describe("AdminCursosComponent", () => {
  let component: AdminCursosComponent;
  let fixture: ComponentFixture<AdminCursosComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;
  let usersServiceSpy: jasmine.SpyObj<UsersService>;
  let messageServiceSpy: jasmine.SpyObj<MessageService>;
  let confirmationServiceSpy: jasmine.SpyObj<ConfirmationService>;

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj("EducationService", [
      "listCourses",
      "listDraftCourses",
      "listTrainingAreas",
      "listMaterials",
      "createCourse",
      "updateCourse",
      "publishCourse",
      "unpublishCourse",
      "deleteCourse",
      "countActiveEnrollments",
      "createTrainingArea",
      "updateTrainingArea",
      "deleteTrainingArea",
      "createMaterial",
      "updateMaterial",
      "deleteMaterial",
      "uploadMaterial",
      "coursesUsingMaterial",
    ]);
    usersServiceSpy = jasmine.createSpyObj("UsersService", ["listByRole"]);
    messageServiceSpy = jasmine.createSpyObj("MessageService", ["add"]);
    confirmationServiceSpy = jasmine.createSpyObj("ConfirmationService", [
      "confirm",
    ]);

    // Default responses
    educationServiceSpy.listCourses.and.returnValue(of([]));
    educationServiceSpy.listDraftCourses.and.returnValue(of([]));
    educationServiceSpy.listTrainingAreas.and.returnValue(of([]));
    educationServiceSpy.listMaterials.and.returnValue(of([]));
    usersServiceSpy.listByRole.and.returnValue(of([]));

    await TestBed.configureTestingModule({
      imports: [AdminCursosComponent],
      providers: [
        { provide: EducationService, useValue: educationServiceSpy },
        { provide: UsersService, useValue: usersServiceSpy },
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(AdminCursosComponent);
    component = fixture.componentInstance;

    // Get real services injected by the component and spy on them
    const msgService = fixture.debugElement.injector.get(MessageService);
    messageServiceSpy = msgService as jasmine.SpyObj<MessageService>;
    spyOn(msgService, "add");

    const confService = fixture.debugElement.injector.get(ConfirmationService);
    confirmationServiceSpy = confService as jasmine.SpyObj<ConfirmationService>;
    spyOn(confService, "confirm").and.callFake((args: any) => {
      if (args.accept) args.accept();
      return confService;
    });

    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should load data on init", () => {
    expect(educationServiceSpy.listCourses).toHaveBeenCalled();
    expect(educationServiceSpy.listDraftCourses).toHaveBeenCalled();
    expect(educationServiceSpy.listTrainingAreas).toHaveBeenCalled();
    expect(educationServiceSpy.listMaterials).toHaveBeenCalled();
    expect(usersServiceSpy.listByRole).toHaveBeenCalledWith("role_teacher");
    expect(component.loading).toBeFalse();
  });

  it("should handle error loading data", () => {
    educationServiceSpy.listCourses.and.returnValue(
      throwError(() => new Error("Load error")),
    );
    component.reloadAll();
    expect(messageServiceSpy.add).toHaveBeenCalledWith(
      jasmine.objectContaining({
        severity: "error",
        summary: "Error al cargar",
        detail: "Load error",
      }),
    );
    expect(component.loading).toBeFalse();
  });

  it("should format desc preview correctly", () => {
    const c: any = {
      description: { content: [{ content: [{ value: "a".repeat(100) }] }] },
    };
    const preview = component.descPreview(c);
    expect(preview.length).toBeLessThanOrEqual(91); // 90 + 1 for ellipsis
    expect(preview.endsWith("…")).toBeTrue();
  });

  describe("Courses", () => {
    it("should open new course dialog", () => {
      component.newCourse();
      expect(component.editingCourseId).toBeNull();
      expect(component.courseDialog).toBeTrue();
      expect(component.courseForm.title).toBe("");
    });

    it("should calculate form validity", () => {
      component.newCourse();
      expect(component.courseFormValid).toBeFalse();

      component.courseForm.title = "Test Title";
      component.courseForm.mode = "Virtual";
      component.courseForm.trainingAreaId = "area1";
      component.courseForm.responsibleTeacher = "teacher1";
      component.courseForm.duration = "4 semanas";
      // Problema 1 de la auditoría: los materiales son opcionales.
      expect(component.courseFormValid)
        .withContext("sin materiales también es válido")
        .toBeTrue();

      component.courseForm.materialIds = ["m1"];
      expect(component.courseFormValid).toBeTrue();
    });

    it("crea un curso sin materiales y lo publica", () => {
      educationServiceSpy.createCourse.and.returnValue(
        of({ id: "c-nuevo" } as any),
      );
      educationServiceSpy.publishCourse.and.returnValue(of(undefined as any));
      component.newCourse();
      Object.assign(component.courseForm, {
        title: "Huerta",
        description: "Curso",
        mode: "Presencial",
        duration: "4 semanas",
        trainingAreaId: "area1",
        responsibleTeacher: "teacher1",
        publicar: true,
      });

      component.saveCourse();

      expect(educationServiceSpy.createCourse).toHaveBeenCalledWith(
        jasmine.objectContaining({ materialIds: [] }),
      );
      expect(educationServiceSpy.publishCourse).toHaveBeenCalledWith("c-nuevo");
      expect(messageServiceSpy.add).toHaveBeenCalledWith(
        jasmine.objectContaining({ severity: "success" }),
      );
    });

    it("si el curso se crea pero no se puede publicar, avisa que quedó en borrador", () => {
      educationServiceSpy.createCourse.and.returnValue(
        of({ id: "c-nuevo" } as any),
      );
      educationServiceSpy.publishCourse.and.returnValue(
        throwError(() => new Error("faltan campos")),
      );
      component.newCourse();
      Object.assign(component.courseForm, {
        title: "Huerta",
        mode: "Presencial",
        duration: "4 semanas",
        trainingAreaId: "area1",
        responsibleTeacher: "teacher1",
        publicar: true,
      });

      component.saveCourse();

      expect(component.courseDialog).toBeFalse();
      expect(component.savingCourse).toBeFalse();
      expect(messageServiceSpy.add).toHaveBeenCalledWith(
        jasmine.objectContaining({
          severity: "warn",
          summary: "Curso creado en borrador",
        }),
      );
    });

    it("sube un material desde el formulario del curso y lo deja adjuntado", () => {
      educationServiceSpy.uploadMaterial.and.returnValue(
        of({ id: "mat-9", title: "Guía", type: "PDF" } as any),
      );
      component.newCourse();
      component.courseForm.title = "Huerta";
      educationServiceSpy.listCourses.calls.reset();

      component.newMaterialForCourse();
      expect(component.materialDialog).toBeTrue();
      component.materialForm.title = "Guía";
      component.materialFile = new File(["%PDF"], "guia.pdf", {
        type: "application/pdf",
      });
      component.saveMaterial();

      expect(component.materialDialog).toBeFalse();
      expect(component.courseDialog)
        .withContext("el curso sigue abierto")
        .toBeTrue();
      expect(component.courseForm.title).toBe("Huerta");
      expect(component.courseForm.materialIds).toEqual(["mat-9"]);
      expect(component.materials.map((m) => m.id)).toContain("mat-9");
      expect(educationServiceSpy.listCourses)
        .withContext("no recarga todo (perdería lo escrito)")
        .not.toHaveBeenCalled();
    });

    it("un material creado desde la pestaña Materiales no se adjunta a ningún curso", () => {
      educationServiceSpy.createMaterial.and.returnValue(
        of({ id: "mat-10", title: "Video", type: "Video" } as any),
      );
      component.newCourse();
      component.newMaterial();
      component.materialForm = {
        title: "Video",
        type: "Video",
        fileAssetId: "",
        url: "https://example.org/v",
      };
      component.saveMaterial();

      expect(component.courseForm.materialIds).toEqual([]);
      expect(component.materials.map((m) => m.id)).toContain("mat-10");
    });

    it("should delete a course", () => {
      const c: any = { id: "1", title: "Course 1" };
      confirmationServiceSpy.confirm.and.callFake((args: any) => {
        if (args.accept) args.accept();
        return confirmationServiceSpy;
      });
      educationServiceSpy.deleteCourse.and.returnValue(of(undefined));
      educationServiceSpy.countActiveEnrollments.and.returnValue(of(2));

      component.deleteCourse(c);

      expect(educationServiceSpy.countActiveEnrollments).toHaveBeenCalledWith(
        "1",
      );
      expect(confirmationServiceSpy.confirm).toHaveBeenCalledWith(
        jasmine.objectContaining({
          message: jasmine.stringContaining("Tiene 2 participantes inscriptos"),
        }),
      );
      expect(educationServiceSpy.deleteCourse).toHaveBeenCalledWith("1");
      expect(messageServiceSpy.add).toHaveBeenCalledWith(
        jasmine.objectContaining({
          severity: "success",
        }),
      );
    });

    it("asks for the deletion even if counting enrollments fails", () => {
      const c: any = { id: "1", title: "Course 1" };
      educationServiceSpy.deleteCourse.and.returnValue(of(undefined));
      educationServiceSpy.countActiveEnrollments.and.returnValue(
        throwError(() => new Error("caído")),
      );

      component.deleteCourse(c);

      expect(confirmationServiceSpy.confirm).toHaveBeenCalledWith(
        jasmine.objectContaining({
          message: jasmine.stringContaining("No pudimos contar"),
        }),
      );
      expect(educationServiceSpy.deleteCourse).toHaveBeenCalledWith("1");
    });
  });

  describe("Areas", () => {
    it("should open new area dialog", () => {
      component.newArea();
      expect(component.editingAreaId).toBeNull();
      expect(component.areaDialog).toBeTrue();
    });

    it("should delete an area", () => {
      const a: any = { id: "1", name: "Area 1" };
      confirmationServiceSpy.confirm.and.callFake((args: any) => {
        if (args.accept) args.accept();
        return confirmationServiceSpy;
      });
      educationServiceSpy.deleteTrainingArea.and.returnValue(of(undefined));

      component.deleteArea(a);

      expect(educationServiceSpy.deleteTrainingArea).toHaveBeenCalledWith("1");
      expect(messageServiceSpy.add).toHaveBeenCalledWith(
        jasmine.objectContaining({
          severity: "success",
        }),
      );
    });
  });

  describe("Materials", () => {
    it("should validate url", () => {
      component.materialForm.url = "invalid-url";
      expect(component.materialUrlInvalid).toBeTrue();

      component.materialForm.url = "http://example.com";
      expect(component.materialUrlInvalid).toBeFalse();
    });

    it("should delete a material", () => {
      const m: any = { id: "1", title: "Mat 1" };
      confirmationServiceSpy.confirm.and.callFake((args: any) => {
        if (args.accept) args.accept();
        return confirmationServiceSpy;
      });
      educationServiceSpy.deleteMaterial.and.returnValue(of(undefined));
      educationServiceSpy.coursesUsingMaterial.and.returnValue(
        of([{ id: "c1", titulo: "Huerta" }]),
      );

      component.deleteMaterial(m);

      expect(confirmationServiceSpy.confirm).toHaveBeenCalledWith(
        jasmine.objectContaining({ message: jasmine.stringContaining("Huerta") }),
      );

      expect(educationServiceSpy.deleteMaterial).toHaveBeenCalledWith("1");
      expect(messageServiceSpy.add).toHaveBeenCalledWith(
        jasmine.objectContaining({
          severity: "success",
        }),
      );
    });
  });
});

describe("mensajeBorrarMaterial (HU-14)", () => {
  it("nombra los cursos que usan el material", () => {
    expect(mensajeBorrarMaterial("Guía", ["Huerta", "Cocina"])).toContain("está en 2 curso(s): Huerta, Cocina");
  });
  it("avisa cuando no está en cursos o no se pudo comprobar", () => {
    expect(mensajeBorrarMaterial("Guía", [])).toContain("No está en ningún curso");
    expect(mensajeBorrarMaterial("Guía", null)).toContain("No se pudo comprobar");
  });
});
