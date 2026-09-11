import { ComponentFixture, TestBed } from "@angular/core/testing";
import { MiCuentaComponent } from "./mi-cuenta.component";
import { provideRouter } from "@angular/router";
import { AuthService } from "../../services/auth.service";
import { of } from "rxjs";

describe("MiCuentaComponent", () => {
  let component: MiCuentaComponent;
  let fixture: ComponentFixture<MiCuentaComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj("AuthService", [
      "isAuthenticated",
      "logout",
    ]);
    mockAuthService.isAuthenticated.and.returnValue(
      of({ role_id: "role_participant" } as any),
    );

    await TestBed.configureTestingModule({
      imports: [MiCuentaComponent],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        provideRouter([]),
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(MiCuentaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create MiCuentaComponent", () => {
    expect(component).toBeTruthy();
  });
});
