import { ComponentFixture, TestBed } from "@angular/core/testing";
import { AccountNoticeComponent } from "./account-notice.component";
import { Router, provideRouter } from "@angular/router";

describe("AccountNoticeComponent", () => {
  let component: AccountNoticeComponent;
  let fixture: ComponentFixture<AccountNoticeComponent>;
  let router: Router;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AccountNoticeComponent],
      providers: [provideRouter([])],
    }).compileComponents();

    router = TestBed.inject(Router);
    spyOn(router, "navigate");

    fixture = TestBed.createComponent(AccountNoticeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create component", () => {
    expect(component).toBeTruthy();
  });

  it("should navigate to /login when goBackToLogin is called", () => {
    component.goBackToLogin();
    expect(router.navigate).toHaveBeenCalledWith(["/login"]);
  });
});
