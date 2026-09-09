import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ContactoComponent } from "./contacto.component";
import { provideRouter } from "@angular/router";

describe("ContactoComponent", () => {
  let component: ContactoComponent;
  let fixture: ComponentFixture<ContactoComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ContactoComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(ContactoComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create ContactoComponent", () => {
    // Arrange & Act done in beforeEach
    // Assert
    expect(component).toBeTruthy();
  });

  it("should render main title", () => {
    // Arrange
    const compiled = fixture.nativeElement as HTMLElement;

    // Act & Assert
    expect(compiled.querySelector(".legal-title")?.textContent).toContain(
      "Contacto e Información Institucional",
    );
  });
});
