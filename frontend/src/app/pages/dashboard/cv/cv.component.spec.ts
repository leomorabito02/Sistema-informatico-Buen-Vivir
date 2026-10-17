import { ComponentFixture, TestBed } from "@angular/core/testing";
import { CvComponent } from "./cv.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { provideRouter } from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";
import { MessageService } from "primeng/api";

describe("CvComponent", () => {
  let component: CvComponent;
  let fixture: ComponentFixture<CvComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CvComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(CvComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  describe("avisos", () => {
    let messages: MessageService;

    beforeEach(() => {
      messages = fixture.debugElement.injector.get(MessageService);
      spyOn(messages, "add");
    });

    it("pide generar el currículum antes de descargarlo", () => {
      component.resume = null;
      component.downloadPdf();
      expect(messages.add).toHaveBeenCalledWith(
        jasmine.objectContaining({ severity: "warn" }),
      );
    });

    it("pide generar el currículum antes de compartirlo", () => {
      component.resume = null;
      component.shareResume();
      expect(messages.add).toHaveBeenCalledWith(
        jasmine.objectContaining({
          detail: "Primero generá el currículum para poder compartirlo.",
        }),
      );
    });
  });

  describe("template and color customization", () => {
    beforeEach(() => {
      component.resume = {
        id: "res-1",
        user_id: "user-1",
        title: "Test CV",
        pdf_url: "",
        public_url: "",
        rxresume_id: "rx-1",
        data: {
          basics: {
            name: "Juan Perez",
            email: "juan@test.com",
            phone: "123456",
            headline: "Developer",
            location: "Buenos Aires",
            website: "",
            summary: "",
            picture: "",
            profiles: [],
          },
          work: [],
          education: [],
          skills: [],
          certifications: [],
          languages: [],
          awards: [],
          volunteer: [],
          projects: [],
          references: [],
          interests: [],
          metadata: {
            template: "onyx",
            design: {
              colors: {
                primary: "#485839",
              },
            },
          },
        },
        created_at: "",
        updated_at: "",
      };
    });

    it("should update selected template correctly", () => {
      component.selectTemplate("azurill");
      expect(component.resume?.data?.metadata?.template).toBe("azurill");
      expect(component.getSelectedTemplateLabel()).toBe("Azurill");
      expect(component.getSelectedTemplateImage()).toContain("azurill.png");
    });

    it("should fallback to defaults when resume has no template", () => {
      if (component.resume?.data?.metadata) {
        component.resume.data.metadata.template = "";
      }
      expect(component.getSelectedTemplateLabel()).toBe("Onyx");
      expect(component.getSelectedTemplateImage()).toContain("onyx.png");
    });

    it("should get and set primary color correctly", () => {
      expect(component.getPrimaryColor()).toBe("#485839");

      component.setPrimaryColor("#9a5b43");
      expect(component.getPrimaryColor()).toBe("#9a5b43");
      expect(component.resume?.data?.metadata?.design?.colors?.primary).toBe("#9a5b43");
    });

    it("should open template modal", () => {
      expect(component.showTemplateModal).toBeFalse();
      component.openTemplateModal();
      expect(component.showTemplateModal).toBeTrue();
    });
  });

  describe("references management", () => {
    beforeEach(() => {
      component.resume = {
        id: "res-1",
        user_id: "user-1",
        title: "Test CV",
        pdf_url: "",
        public_url: "",
        rxresume_id: "rx-1",
        data: {
          basics: {
            name: "Juan Perez",
            email: "juan@test.com",
            phone: "123456",
            headline: "Developer",
            location: "Buenos Aires",
            website: "",
            summary: "",
            picture: "",
            profiles: [],
          },
          work: [],
          education: [],
          skills: [],
          certifications: [],
          languages: [],
          awards: [],
          volunteer: [],
          projects: [],
          references: [],
          interests: [],
        },
        created_at: "",
        updated_at: "",
      };
    });

    it("should add and remove references", () => {
      expect(component.resume?.data?.references?.length).toBe(0);

      component.addReference();
      expect(component.resume?.data?.references?.length).toBe(1);
      const ref = component.resume?.data?.references[0];
      expect(ref?.id).toBeTruthy();

      component.removeReference(0);
      expect(component.resume?.data?.references?.length).toBe(0);
    });
  });

  describe("profile picture synchronization", () => {
    it("should set picture from userProfilePicture on generateResume if basics.picture is empty", () => {
      const rxResumeService = fixture.debugElement.injector.get(
        // @ts-ignore
        component["rxResumeService"].constructor
      );
      spyOn(rxResumeService, "updateMyResume").and.callFake((data: any) => {
        return { subscribe: () => {} } as any;
      });

      component.resume = {
        id: "res-1",
        user_id: "user-1",
        title: "Test CV",
        pdf_url: "",
        public_url: "",
        rxresume_id: "rx-1",
        data: {
          basics: {
            name: "Juan Perez",
            email: "juan@test.com",
            phone: "123456",
            headline: "Developer",
            location: "Buenos Aires",
            website: "",
            summary: "",
            picture: "",
            profiles: [],
          },
          work: [],
          education: [],
          skills: [],
          certifications: [],
          languages: [],
          awards: [],
          volunteer: [],
          projects: [],
          references: [],
          interests: [],
        },
        created_at: "",
        updated_at: "",
      };
      component.userProfilePicture = "https://example.com/avatar.png";

      component.generateResume();

      expect(component.resume.data.basics.picture).toBe(
        "https://example.com/avatar.png"
      );
      expect(rxResumeService.updateMyResume).toHaveBeenCalledWith(
        jasmine.objectContaining({
          basics: jasmine.objectContaining({
            picture: "https://example.com/avatar.png",
          }),
        })
      );
    });
  });
});
