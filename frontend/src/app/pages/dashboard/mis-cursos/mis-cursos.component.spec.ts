import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MisCursosComponent } from "./mis-cursos.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { provideRouter } from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("MisCursosComponent", () => {
  let component: MisCursosComponent;
  let fixture: ComponentFixture<MisCursosComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MisCursosComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(MisCursosComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });
});
