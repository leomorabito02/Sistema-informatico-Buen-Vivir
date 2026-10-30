import { ComponentFixture, TestBed } from "@angular/core/testing";
import { GlobalErrorDialogComponent } from "./global-error-dialog.component";
import { AlertService } from "../../../services/alert.service";
import { MessageService } from "primeng/api";
import { provideAnimations } from "@angular/platform-browser/animations";

describe("GlobalErrorDialogComponent", () => {
  let component: GlobalErrorDialogComponent;
  let fixture: ComponentFixture<GlobalErrorDialogComponent>;
  let alertService: AlertService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [GlobalErrorDialogComponent],
      providers: [AlertService, MessageService, provideAnimations()],
    }).compileComponents();

    fixture = TestBed.createComponent(GlobalErrorDialogComponent);
    component = fixture.componentInstance;
    alertService = TestBed.inject(AlertService);
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should call alertService.confirmDialog when onConfirm is invoked", () => {
    spyOn(alertService, "confirmDialog");
    component.onConfirm();
    expect(alertService.confirmDialog).toHaveBeenCalled();
  });

  it("should call alertService.cancelDialog when onCancel is invoked", () => {
    spyOn(alertService, "cancelDialog");
    component.onCancel();
    expect(alertService.cancelDialog).toHaveBeenCalled();
  });
});
