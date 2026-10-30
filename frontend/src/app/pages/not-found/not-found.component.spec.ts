import { ComponentFixture, TestBed } from "@angular/core/testing";
import { NotFoundComponent } from "./not-found.component";
import { Router, provideRouter } from "@angular/router";

describe("NotFoundComponent", () => {
  let component: NotFoundComponent;
  let fixture: ComponentFixture<NotFoundComponent>;
  let router: Router;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NotFoundComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, "navigate");

    fixture = TestBed.createComponent(NotFoundComponent);
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

  it("should navigate to / or call history.back on goBack", () => {
    spyOn(window.history, "back");
    component.goBack();
    if (window.history.length > 1) {
      expect(window.history.back).toHaveBeenCalled();
    } else {
      expect(router.navigate).toHaveBeenCalledWith(["/"]);
    }
  });
});
