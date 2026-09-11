import { ComponentFixture, TestBed } from "@angular/core/testing";
import { PrivacyPolicyComponent } from "./privacy-policy.component";
import { provideRouter } from "@angular/router";

describe("PrivacyPolicyComponent", () => {
  let component: PrivacyPolicyComponent;
  let fixture: ComponentFixture<PrivacyPolicyComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PrivacyPolicyComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(PrivacyPolicyComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create PrivacyPolicyComponent", () => {
    // Arrange & Act done in beforeEach
    // Assert
    expect(component).toBeTruthy();
  });

  it("should render main title", () => {
    // Arrange
    const compiled = fixture.nativeElement as HTMLElement;

    // Act & Assert
    expect(compiled.querySelector(".legal-title")?.textContent).toContain(
      "Política de Privacidad",
    );
  });
});
