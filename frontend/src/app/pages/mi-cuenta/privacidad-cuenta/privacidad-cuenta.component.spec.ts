import { ComponentFixture, TestBed } from "@angular/core/testing";
import { PrivacidadCuentaComponent } from "./privacidad-cuenta.component";

describe("PrivacidadCuentaComponent", () => {
  let component: PrivacidadCuentaComponent;
  let fixture: ComponentFixture<PrivacidadCuentaComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PrivacidadCuentaComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(PrivacidadCuentaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create PrivacidadCuentaComponent", () => {
    expect(component).toBeTruthy();
  });
});
