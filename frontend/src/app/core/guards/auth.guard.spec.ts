import { TestBed } from '@angular/core/testing';
import { CanActivateFn, Router, UrlTree } from '@angular/router';
import { authGuard } from './auth.guard';
import { AuthService } from '../../services/auth.service';
import { of, throwError } from 'rxjs';

describe('authGuard', () => {
  const executeGuard: CanActivateFn = (...guardParameters) => 
      TestBed.runInInjectionContext(() => authGuard(...guardParameters));

  let mockAuthService: jasmine.SpyObj<AuthService>;
  let mockRouter: jasmine.SpyObj<Router>;

  beforeEach(() => {
    mockAuthService = jasmine.createSpyObj('AuthService', ['isAuthenticated']);
    mockRouter = jasmine.createSpyObj('Router', ['createUrlTree']);

    TestBed.configureTestingModule({
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        { provide: Router, useValue: mockRouter }
      ]
    });
  });

  it('should redirect to /account-notice if user_status is blocked', (done) => {
    mockAuthService.isAuthenticated.and.returnValue(of({ user_status: 'blocked' } as any));
    const urlTree = {} as UrlTree;
    mockRouter.createUrlTree.and.returnValue(urlTree);

    const result = executeGuard({} as any, {} as any);
    if (typeof result !== 'boolean' && 'subscribe' in (result as any)) {
      (result as any).subscribe((res: any) => {
        expect(res).toBe(urlTree);
        expect(mockRouter.createUrlTree).toHaveBeenCalledWith(['/account-notice']);
        done();
      });
    }
  });

  it('should redirect to /account-notice if user.status is blocked', (done) => {
    mockAuthService.isAuthenticated.and.returnValue(of({ user: { status: 'blocked' } } as any));
    const urlTree = {} as UrlTree;
    mockRouter.createUrlTree.and.returnValue(urlTree);

    const result = executeGuard({} as any, {} as any);
    if (typeof result !== 'boolean' && 'subscribe' in (result as any)) {
      (result as any).subscribe((res: any) => {
        expect(res).toBe(urlTree);
        expect(mockRouter.createUrlTree).toHaveBeenCalledWith(['/account-notice']);
        done();
      });
    }
  });

  it('should redirect to /pending-approval if approval_status is pending', (done) => {
    mockAuthService.isAuthenticated.and.returnValue(of({ approval_status: 'pending' } as any));
    const urlTree = {} as UrlTree;
    mockRouter.createUrlTree.and.returnValue(urlTree);

    const result = executeGuard({} as any, {} as any);
    if (typeof result !== 'boolean' && 'subscribe' in (result as any)) {
      (result as any).subscribe((res: any) => {
        expect(res).toBe(urlTree);
        expect(mockRouter.createUrlTree).toHaveBeenCalledWith(['/pending-approval']);
        done();
      });
    }
  });

  it('should return true if authenticated and not blocked/pending', (done) => {
    mockAuthService.isAuthenticated.and.returnValue(of({ user: { status: 'active' } } as any));

    const result = executeGuard({} as any, {} as any);
    if (typeof result !== 'boolean' && 'subscribe' in (result as any)) {
      (result as any).subscribe((res: any) => {
        expect(res).toBe(true);
        done();
      });
    }
  });

  it('should redirect to /login if isAuthenticated throws an error', (done) => {
    mockAuthService.isAuthenticated.and.returnValue(throwError(() => new Error('Unauthorized')));
    const urlTree = {} as UrlTree;
    mockRouter.createUrlTree.and.returnValue(urlTree);

    const result = executeGuard({} as any, {} as any);
    if (typeof result !== 'boolean' && 'subscribe' in (result as any)) {
      (result as any).subscribe((res: any) => {
        expect(res).toBe(urlTree);
        expect(mockRouter.createUrlTree).toHaveBeenCalledWith(['/login']);
        done();
      });
    }
  });
});
