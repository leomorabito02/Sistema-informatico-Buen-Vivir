import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MisCursosComponent } from "./mis-cursos.component";
import { EducationService } from "../../../services/education.service";
import { MessageService } from "primeng/api";
import { of, throwError } from "rxjs";
import { provideHttpClient } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideRouter } from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("MisCursosComponent", () => {
  let component: MisCursosComponent;
  let fixture: ComponentFixture<MisCursosComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;
  let messageServiceSpy: jasmine.SpyObj<MessageService>;

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj("EducationService", [
      "listMyCourses",
      "listMyDraftCourses",
      "createMaterial",
      "uploadMaterial",
      "updateMyCourse",
      "listTrainingAreas",
      "createMyCourse",
      "deleteMyCourse",
    ]);
    educationServiceSpy.listTrainingAreas.and.returnValue(
      of([{ id: "a1", name: "Huerta" }]),
    );
    messageServiceSpy = jasmine.createSpyObj("MessageService", ["add"]);

    educationServiceSpy.listMyCourses.and.returnValue(of([]));
    educationServiceSpy.listMyDraftCourses.and.returnValue(of([]));

    await TestBed.configureTestingModule({
      imports: [MisCursosComponent, BrowserAnimationsModule],
      providers: [
        { provide: EducationService, useValue: educationServiceSpy },
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(MisCursosComponent);
    component = fixture.componentInstance;

    const msgService = fixture.debugElement.injector.get(MessageService);
    messageServiceSpy = msgService as jasmine.SpyObj<MessageService>;
    spyOn(msgService, "add");

    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should load my courses on init", () => {
    const published: any[] = [{ id: "1", title: "Course 1" }];
    const drafts: any[] = [{ id: "2", title: "Draft 1" }];
    educationServiceSpy.listMyCourses.and.returnValue(of(published));
    educationServiceSpy.listMyDraftCourses.and.returnValue(of(drafts));

    component.ngOnInit();

    expect(educationServiceSpy.listMyCourses).toHaveBeenCalled();
    expect(educationServiceSpy.listMyDraftCourses).toHaveBeenCalled();
    expect(component.courses).toHaveSize(2);
    expect(component.courses[0].publicado).toBeTrue();
    expect(component.courses[1].publicado).toBeFalse();
    expect(component.loading).toBeFalse();
  });

  it("should handle error when loading courses", () => {
    educationServiceSpy.listMyCourses.and.returnValue(
      throwError(() => new Error("Load error")),
    );

    component.ngOnInit();

    expect(component.error).toBe("Load error");
    expect(component.loading).toBeFalse();
  });

  describe("Materials", () => {
    it("should open materials dialog", () => {
      const c: any = { id: "1" };
      component.openMaterials(c);

      expect(component.matCourse).toBe(c);
      expect(component.matDialog).toBeTrue();
    });

    it("should validate url", () => {
      component.matForm.url = "invalid";
      expect(component.matUrlInvalid).toBeTrue();
      component.matForm.url = "http://test.com";
      expect(component.matUrlInvalid).toBeFalse();
    });

    it("should add material via url", () => {
      const c: any = { id: "1", title: "Course 1", raw: { id: "1" } };
      const newMat: any = { id: "mat1", title: "New Mat" };

      component.matCourse = c;
      component.matForm.title = "Test";
      component.matForm.url = "http://test.com";

      educationServiceSpy.createMaterial.and.returnValue(of(newMat));
      educationServiceSpy.updateMyCourse.and.returnValue(
        of({ ...c.raw, material: [newMat] }),
      );

      component.addMaterial();

      expect(educationServiceSpy.createMaterial).toHaveBeenCalled();
      expect(educationServiceSpy.updateMyCourse).toHaveBeenCalled();
      expect(messageServiceSpy.add).toHaveBeenCalledWith(
        jasmine.objectContaining({ severity: "success" }),
      );
    });
  });

  describe("Curso propio (22/11/2026)", () => {
    it("crea el curso y avisa que queda pendiente de aprobación", () => {
      educationServiceSpy.createMyCourse.and.returnValue(
        of({ id: "n1", title: "Nuevo" } as any),
      );
      component.openCourseForm();
      component.courseForm = {
        title: "Huerta urbana",
        description: "Aprender a cultivar",
        trainingAreaId: "a1",
        mode: "Presencial",
        duration: "8 semanas",
      };

      component.saveCourse();

      expect(educationServiceSpy.createMyCourse).toHaveBeenCalledWith(
        jasmine.objectContaining({
          title: "Huerta urbana",
          trainingAreaId: "a1",
          mode: "Presencial",
          duration: "8 semanas",
        }),
      );
      expect(messageServiceSpy.add).toHaveBeenCalledWith(
        jasmine.objectContaining({
          severity: "success",
          detail: "Quedó pendiente de aprobación del administrador.",
        }),
      );
    });

    it("no guarda si faltan datos obligatorios", () => {
      component.openCourseForm();
      component.courseForm.title = "Solo título";
      expect(component.courseFormValid).toBeFalse();
      component.saveCourse();
      expect(educationServiceSpy.createMyCourse).not.toHaveBeenCalled();
    });

    it("al editar conserva los materiales del curso", () => {
      const vm: any = {
        id: "c1",
        title: "Curso",
        publicado: false,
        raw: {
          id: "c1",
          title: "Curso",
          mode: "Virtual",
          duration: "4 semanas",
          trainingArea: { id: "a1", name: "Huerta" },
          material: [{ id: "m1" }],
        },
      };
      educationServiceSpy.updateMyCourse.and.returnValue(of(vm.raw));
      component.openCourseForm(vm);
      component.courseForm.description = "Descripción";

      component.saveCourse();

      expect(educationServiceSpy.updateMyCourse).toHaveBeenCalledWith(
        "c1",
        jasmine.objectContaining({ materialIds: ["m1"], mode: "Virtual" }),
      );
    });

    it("elimina un curso pendiente", () => {
      const vm: any = { id: "c1", title: "Curso", publicado: false, raw: {} };
      component.courses = [vm];
      educationServiceSpy.deleteMyCourse.and.returnValue(of(void 0));

      component.askDelete(vm);
      component.confirmDelete();

      expect(educationServiceSpy.deleteMyCourse).toHaveBeenCalledWith("c1");
      expect(component.courses.length).toBe(0);
    });
  });
});
