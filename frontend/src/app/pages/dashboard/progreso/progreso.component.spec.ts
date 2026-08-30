import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ProgresoComponent } from "./progreso.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { provideRouter } from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("ProgresoComponent", () => {
  let component: ProgresoComponent;
  let fixture: ComponentFixture<ProgresoComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ProgresoComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ProgresoComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });
});
