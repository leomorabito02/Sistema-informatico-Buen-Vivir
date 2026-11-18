import { ComponentFixture, TestBed } from "@angular/core/testing";
import { TermsConditionsComponent } from "./terms-conditions.component";
import { provideRouter } from "@angular/router";

describe("TermsConditionsComponent", () => {
  let component: TermsConditionsComponent;
  let fixture: ComponentFixture<TermsConditionsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TermsConditionsComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(TermsConditionsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create TermsConditionsComponent", () => {
    // Arrange & Act done in beforeEach
    // Assert
    expect(component).toBeTruthy();
  });

  it("should render main title", () => {
    // Arrange
    const compiled = fixture.nativeElement as HTMLElement;

    // Act & Assert
    expect(compiled.querySelector(".legal-title")?.textContent).toContain(
      "Términos y condiciones de uso",
    );
  });
});
