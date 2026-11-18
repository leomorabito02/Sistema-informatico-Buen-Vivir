import { ComponentFixture, TestBed } from "@angular/core/testing";
import { CookiePolicyComponent } from "./cookie-policy.component";
import { provideRouter } from "@angular/router";

describe("CookiePolicyComponent", () => {
  let component: CookiePolicyComponent;
  let fixture: ComponentFixture<CookiePolicyComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CookiePolicyComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(CookiePolicyComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create CookiePolicyComponent", () => {
    // Arrange & Act done in beforeEach
    // Assert
    expect(component).toBeTruthy();
  });

  it("should render main title", () => {
    // Arrange
    const compiled = fixture.nativeElement as HTMLElement;

    // Act & Assert
    expect(compiled.querySelector(".legal-title")?.textContent).toContain(
      "Política de cookies",
    );
  });
});
