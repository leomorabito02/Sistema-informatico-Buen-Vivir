import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ServerErrorComponent } from "./server-error.component";
import { Router, provideRouter } from "@angular/router";
import { NavigationHistoryService } from "../../services/navigation-history.service";

describe("ServerErrorComponent", () => {
  let component: ServerErrorComponent;
  let fixture: ComponentFixture<ServerErrorComponent>;
  let router: Router;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ServerErrorComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, "navigate");

    fixture = TestBed.createComponent(ServerErrorComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create component", () => {
    expect(component).toBeTruthy();
  });

  it("should navigate to / when goToHome is called", () => {
    component.goToHome();
    expect(router.navigate).toHaveBeenCalledWith(["/"]);
  });

  it("should delegate reloadPage to NavigationHistoryService", () => {
    const navHistory = TestBed.inject(NavigationHistoryService);
    spyOn(navHistory, "reloadPreviousPage");
    component.reloadPage();
    expect(navHistory.reloadPreviousPage).toHaveBeenCalled();
  });
});
