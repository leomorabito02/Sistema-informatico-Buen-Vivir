import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick,
} from "@angular/core/testing";
import { of } from "rxjs";
import { JobOfferService } from "../../../services/job-offer.service";
import { OfertasComponent } from "./ofertas.component";
import { provideHttpClientTesting } from "@angular/common/http/testing";
import { provideHttpClient } from "@angular/common/http";
import { provideRouter } from "@angular/router";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { Auth } from "@angular/fire/auth";

describe("OfertasComponent", () => {
  let component: OfertasComponent;
  let fixture: ComponentFixture<OfertasComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OfertasComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(OfertasComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("asks labor-integration for the search text after typing stops", fakeAsync(() => {
    const service = TestBed.inject(JobOfferService);
    const getOffers = spyOn(service, "getOffers").and.returnValue(of([]));

    component.searchTerm = "coc";
    component.buscar();
    component.searchTerm = "cocina";
    component.buscar();
    tick(299);
    expect(getOffers).not.toHaveBeenCalled();

    tick(1);
    expect(getOffers).toHaveBeenCalledOnceWith("cocina");
  }));

  it("shows approve and reject only to the administrator", () => {
    component.jobs = [
      { id: "of-1", titulo: "Cocinero", estado: "Pendiente" } as any,
    ];
    component.isAdminOrVolunteer = true;
    component.isAdmin = false;
    component.filterJobs();
    fixture.detectChanges();
    const labels = () =>
      Array.from(
        fixture.nativeElement.querySelectorAll(
          "p-button",
        ) as NodeListOf<HTMLElement>,
      ).map((b) => b.getAttribute("label") ?? b.textContent?.trim());
    expect(labels()).not.toContain("Aprobar");

    component.isAdmin = true;
    fixture.detectChanges();
    expect(labels()).toContain("Aprobar");
    expect(labels()).toContain("Rechazar");
  });
});
