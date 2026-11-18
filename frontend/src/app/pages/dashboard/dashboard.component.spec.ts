import { ComponentFixture, TestBed } from "@angular/core/testing";
import { DashboardComponent } from "./dashboard.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { Router, provideRouter } from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("DashboardComponent", () => {
  let component: DashboardComponent;
  let fixture: ComponentFixture<DashboardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("muestra el aviso 403 cuando un guard redirige con acceso=denegado", async () => {
    const router = TestBed.inject(Router);
    await router.navigateByUrl("/?acceso=denegado");
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(component.accesoDenegado).toBeTrue();
    expect(el.querySelector(".acceso-denegado")?.textContent).toContain("403");

    component.cerrarAviso();
    fixture.detectChanges();
    expect(el.querySelector(".acceso-denegado")).toBeNull();
  });
});
