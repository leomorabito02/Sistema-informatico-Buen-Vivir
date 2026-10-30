import { TestBed } from "@angular/core/testing";
import { provideHttpClient } from "@angular/common/http";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { CvPdfService } from "./cv-pdf.service";

describe("CvPdfService", () => {
  let service: CvPdfService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        CvPdfService,
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });
    service = TestBed.inject(CvPdfService);
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should return empty string if picture URL is empty or null", async () => {
    expect(await service.resolvePictureAsDataUrl("")).toBe("");
    expect(await service.resolvePictureAsDataUrl(null)).toBe("");
  });

  it("should preserve already valid data URLs", async () => {
    const dataUrl =
      "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==";
    expect(await service.resolvePictureAsDataUrl(dataUrl)).toBe(dataUrl);
  });

  it("should generate PDF blob using mock engine", async () => {
    const mockBlob = new Blob(["%PDF-mock"], { type: "application/pdf" });
    const mockEngine = {
      generateCvPdfBlob: jasmine
        .createSpy("generateCvPdfBlob")
        .and.resolveTo(mockBlob),
    };
    (service as any).enginePromise = Promise.resolve(mockEngine);

    const blob = await service.generatePdfBlob({
      basics: { name: "Test User", email: "test@example.com" } as any,
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
    });

    expect(blob).toBe(mockBlob);
    expect(mockEngine.generateCvPdfBlob).toHaveBeenCalled();
  });
});
