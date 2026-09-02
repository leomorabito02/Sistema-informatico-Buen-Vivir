import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick,
} from "@angular/core/testing";
import { AdminCursosComponent } from "./admin-cursos.component";
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
      "createTrainingArea",
      "updateTrainingArea",
      "deleteTrainingArea",
      "createMaterial",
      "updateMaterial",
      "deleteMaterial",
      "uploadMaterial",
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

      expect(component.courseFormValid).toBeTrue();
    });

    it("should delete a course", () => {
      const c: any = { id: "1", title: "Course 1" };
      confirmationServiceSpy.confirm.and.callFake((args: any) => {
        if (args.accept) args.accept();
        return confirmationServiceSpy;
      });
      educationServiceSpy.deleteCourse.and.returnValue(of(undefined));

      component.deleteCourse(c);

      expect(educationServiceSpy.deleteCourse).toHaveBeenCalledWith("1");
      expect(messageServiceSpy.add).toHaveBeenCalledWith(
        jasmine.objectContaining({
          severity: "success",
        }),
      );
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

      component.deleteMaterial(m);

      expect(educationServiceSpy.deleteMaterial).toHaveBeenCalledWith("1");
      expect(messageServiceSpy.add).toHaveBeenCalledWith(
        jasmine.objectContaining({
          severity: "success",
        }),
      );
    });
  });
});
