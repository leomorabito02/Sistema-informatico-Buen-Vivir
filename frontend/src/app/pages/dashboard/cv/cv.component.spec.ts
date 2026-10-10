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

    it("rechaza imágenes de más de 10 MB", () => {
      const file = { size: 11 * 1024 * 1024 };
      component.onPictureSelected({ target: { files: [file] } });
      expect(messages.add).toHaveBeenCalledWith(
        jasmine.objectContaining({
          detail: "La imagen no debe superar los 10 MB.",
        }),
      );
      expect(component.uploadingPicture).toBeFalse();
    });
  });
});
