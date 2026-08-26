import { ComponentFixture, TestBed } from "@angular/core/testing";
import { InsercionComponent } from "./insercion.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { provideRouter } from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("InsercionComponent", () => {
  let component: InsercionComponent;
  let fixture: ComponentFixture<InsercionComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [InsercionComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(InsercionComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });
});
