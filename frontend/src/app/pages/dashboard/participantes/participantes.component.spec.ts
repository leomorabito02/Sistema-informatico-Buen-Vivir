import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ParticipantesComponent } from "./participantes.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { provideRouter } from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("ParticipantesComponent", () => {
  let component: ParticipantesComponent;
  let fixture: ComponentFixture<ParticipantesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ParticipantesComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ParticipantesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });
});
