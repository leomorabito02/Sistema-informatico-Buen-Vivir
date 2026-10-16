import { ComponentFixture, TestBed } from "@angular/core/testing";
import { CvPublicComponent } from "./cv-public.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { ActivatedRoute } from "@angular/router";
import { of, throwError } from "rxjs";
import { RxResumeService, Resume } from "../../services/rxresume.service";
import { CvPdfService } from "../../services/cv-pdf.service";

describe("CvPublicComponent", () => {
  let component: CvPublicComponent;
  let fixture: ComponentFixture<CvPublicComponent>;
  let rxResumeService: RxResumeService;

  const mockResume: Resume = {
    id: "pub-1",
    user_id: "user-123",
    title: "Currículum Vitae",
    pdf_url: "https://storage.googleapis.com/test.pdf",
    public_url: "https://buenvivir.org/cv-public/user-123",
    rxresume_id: "rx-123",
    data: {
      basics: {
        name: "María Gómez",
        email: "maria@example.com",
        phone: "11223344",
        headline: "Diseñadora UX/UI",
        location: "Córdoba, Argentina",
        website: "",
        summary: "Profesional apasionada por el diseño de interfaces.",
        picture: "uploads/pic.jpg",
        profiles: [],
      },
      work: [
        {
          id: "w1",
          company: "Estudio Creativo",
          position: "Diseñadora Senior",
          startDate: "2021",
          endDate: "2024",
          description: "Liderazgo de proyectos UI.",
        },
      ],
      education: [
        {
          id: "e1",
          institution: "Universidad Nacional",
          degree: "Licenciatura en Diseño",
          field: "Diseño Gráfico",
          startDate: "2016",
          endDate: "2020",
          grade: "Distinguido",
        },
      ],
      skills: [{ id: "s1", name: "Figma", level: "Avanzado" }],
      certifications: [],
      languages: [{ id: "l1", name: "Inglés", level: "Intermedio" }],
      awards: [],
      volunteer: [],
      projects: [],
      references: [
        {
          id: "r1",
          name: "Lic. Pedro Álvarez",
          description: "Director en Estudio Creativo - pedro@estudio.com",
        },
      ],
      interests: [],
      metadata: {
        template: "onyx",
        design: {
          colors: {
            primary: "#9a5b43",
          },
        },
      },
    },
    created_at: "2024-01-01",
    updated_at: "2024-01-02",
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CvPublicComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: {
                get: (key: string) => (key === "id" ? "user-123" : null),
              },
            },
          },
        },
      ],
    }).compileComponents();

    rxResumeService = TestBed.inject(RxResumeService);
    fixture = TestBed.createComponent(CvPublicComponent);
    component = fixture.componentInstance;
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should load public resume on init", () => {
    spyOn(rxResumeService, "getPublicResume").and.returnValue(of(mockResume));

    component.ngOnInit();

    expect(rxResumeService.getPublicResume).toHaveBeenCalledWith("user-123");
    expect(component.resume).toEqual(mockResume);
    expect(component.loading).toBeFalse();
    expect(component.error).toBeFalse();
    expect(document.title).toBe("CV - María Gómez");
  });

  it("should set error state if loading fails", () => {
    spyOn(rxResumeService, "getPublicResume").and.returnValue(
      throwError(() => new Error("Not found")),
    );

    component.ngOnInit();

    expect(component.error).toBeTrue();
    expect(component.loading).toBeFalse();
    expect(component.resume).toBeNull();
  });

  it("should get primary color from metadata or return default", () => {
    component.resume = mockResume;
    expect(component.getPrimaryColor()).toBe("#9a5b43");

    component.resume = {
      ...mockResume,
      data: {
        ...mockResume.data,
        metadata: undefined,
      },
    };
    expect(component.getPrimaryColor()).toBe("#485839");
  });

  it("should trigger PDF download via cvPdfService or fallback", async () => {
    const cvPdfService = TestBed.inject(CvPdfService);
    spyOn(cvPdfService, "downloadPdf").and.resolveTo();
    spyOn(rxResumeService, "downloadPublicPDF");
    component.userId = "user-123";
    component.resume = mockResume;

    await component.downloadPDF();

    expect(cvPdfService.downloadPdf).toHaveBeenCalled();

    component.resume = null;
    await component.downloadPDF();
    expect(rxResumeService.downloadPublicPDF).toHaveBeenCalledWith("user-123");
  });
});
