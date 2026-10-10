import { ComponentFixture, TestBed } from "@angular/core/testing";
import { provideRouter } from "@angular/router";
import { BackLinkComponent } from "./back-link.component";

describe("BackLinkComponent", () => {
  let fixture: ComponentFixture<BackLinkComponent>;
  let component: BackLinkComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [BackLinkComponent],
      providers: [provideRouter([])],
    }).compileComponents();
    fixture = TestBed.createComponent(BackLinkComponent);
    component = fixture.componentInstance;
  });

  it("renders a link to the given route with its label", () => {
    component.to = "/dashboard/promocion-humana";
    component.label = "Volver a la búsqueda";
    fixture.detectChanges();

    const a: HTMLAnchorElement = fixture.nativeElement.querySelector("a");
    expect(a).not.toBeNull();
    expect(a.getAttribute("href")).toBe("/dashboard/promocion-humana");
    expect(a.textContent).toContain("Volver a la búsqueda");
    expect(fixture.nativeElement.querySelector("button")).toBeNull();
  });

  it("renders a button that emits back when there is no route", () => {
    const spy = jasmine.createSpy("back");
    component.back.subscribe(spy);
    fixture.detectChanges();

    const btn: HTMLButtonElement =
      fixture.nativeElement.querySelector("button");
    expect(btn.textContent).toContain("Volver");
    btn.click();
    expect(spy).toHaveBeenCalled();
  });
});
