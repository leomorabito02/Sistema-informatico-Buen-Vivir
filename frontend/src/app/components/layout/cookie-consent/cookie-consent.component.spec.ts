import { ComponentFixture, TestBed } from "@angular/core/testing";
import { CookieConsentComponent } from "./cookie-consent.component";
import { provideRouter } from "@angular/router";

describe("CookieConsentComponent", () => {
  let component: CookieConsentComponent;
  let fixture: ComponentFixture<CookieConsentComponent>;

  beforeEach(async () => {
    localStorage.clear();

    await TestBed.configureTestingModule({
      imports: [CookieConsentComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(CookieConsentComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  afterEach(() => {
    localStorage.clear();
  });

  it("should create the component", () => {
    // Arrange & Act done in beforeEach
    // Assert
    expect(component).toBeTruthy();
  });

  it("should show banner when consent is not stored in localStorage", () => {
    // Arrange
    localStorage.removeItem("cookie_consent");

    // Act
    component.ngOnInit();

    // Assert
    expect(component.isVisible).toBeTrue();
  });

  it("should store 'accepted' in localStorage when acceptCookies is called", () => {
    // Arrange
    component.isVisible = true;

    // Act
    component.acceptCookies();

    // Assert
    expect(localStorage.getItem("cookie_consent")).toBe("accepted");
    expect(component.isVisible).toBeFalse();
  });

  it("should store 'rejected' in localStorage when rejectCookies is called", () => {
    // Arrange
    component.isVisible = true;

    // Act
    component.rejectCookies();

    // Assert
    expect(localStorage.getItem("cookie_consent")).toBe("rejected");
    expect(component.isVisible).toBeFalse();
  });
});
