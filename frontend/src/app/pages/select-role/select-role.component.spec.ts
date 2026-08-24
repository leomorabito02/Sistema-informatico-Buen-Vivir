import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SelectRoleComponent } from './select-role.component';
import { Router, provideRouter } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { of, throwError } from 'rxjs';

describe('SelectRoleComponent', () => {
  let component: SelectRoleComponent;
  let fixture: ComponentFixture<SelectRoleComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockRouter: any;

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj('AuthService', ['loginWithAccount']);

    await TestBed.configureTestingModule({
      imports: [SelectRoleComponent],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        provideRouter([])
      ]
    }).compileComponents();
    
    mockRouter = TestBed.inject(Router);
    spyOn(mockRouter, 'navigate');
    spyOn(mockRouter, 'getCurrentNavigation').and.returnValue(null as any);
  });

  describe('ngOnInit', () => {
    it('should redirect to login if no accounts in state', () => {
      (mockRouter.getCurrentNavigation as jasmine.Spy).and.returnValue(null);
      history.replaceState({}, '');
      
      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      fixture.detectChanges(); // This will trigger ngOnInit
      
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/login']);
    });

    it('should load accounts from history state', () => {
      const accounts = [{ user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' }];
      history.pushState({ accounts, loginNonce: 'nonce' }, '');
      
      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      fixture.detectChanges();
      
      expect(component.accounts).toEqual(accounts);
      expect(component.loginNonce).toBe('nonce');
      expect(mockRouter.navigate).not.toHaveBeenCalled();
    });
  });

  describe('selectAccount', () => {
    beforeEach(() => {
      const accounts = [{ user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' }];
      history.pushState({ accounts, loginNonce: 'nonce' }, '');
      
      fixture = TestBed.createComponent(SelectRoleComponent);
      component = fixture.componentInstance;
      fixture.detectChanges();
      
      mockRouter.navigate.calls.reset();
    });

    it('should navigate to account-notice if account is not active', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'inactive' };
      component.selectAccount(account);
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/account-notice']);
    });

    it('should login and navigate to dashboard', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' };
      mockAuthService.loginWithAccount.and.returnValue(of({ token: 'jwt', user: {} as any }));
      
      component.selectAccount(account);
      
      expect(mockAuthService.loginWithAccount).toHaveBeenCalledWith('1', 'r1', 'nonce');
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/dashboard']);
      expect(component.loadingAccountId).toBe('');
    });

    it('should navigate to account-notice if login response returns blocked', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' };
      mockAuthService.loginWithAccount.and.returnValue(of({ token: 'jwt', user: {} as any, user_status: 'blocked' }));
      
      component.selectAccount(account);
      
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/account-notice']);
    });

    it('should navigate to pending-approval if login response returns pending', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' };
      mockAuthService.loginWithAccount.and.returnValue(of({ token: 'jwt', user: {} as any, approval_status: 'pending' }));
      
      component.selectAccount(account);
      
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/pending-approval']);
    });

    it('should navigate to account-notice on specific error message', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' };
      mockAuthService.loginWithAccount.and.returnValue(throwError(() => ({ error: { error: 'cuenta no está activa' } })));
      
      component.selectAccount(account);
      
      expect(mockRouter.navigate).toHaveBeenCalledWith(['/account-notice']);
    });

    it('should show generic error message on other errors', () => {
      const account = { user_id: '1', role_id: 'r1', role_name: 'Role 1', status: 'active' };
      mockAuthService.loginWithAccount.and.returnValue(throwError(() => ({ error: { error: 'internal error' } })));
      
      component.selectAccount(account);
      
      expect(component.loginError).toBe('Ocurrió un error al iniciar sesión. Intentá nuevamente.');
      expect(mockRouter.navigate).not.toHaveBeenCalled();
    });
  });
});
