import { TestBed } from '@angular/core/testing';
import { CanActivateFn, Router, UrlTree } from '@angular/router';
import { guestGuard } from './guest.guard';
import { AuthService } from '../../services/auth.service';
import { of, throwError } from 'rxjs';

describe('guestGuard', () => {
  const executeGuard: CanActivateFn = (...guardParameters) => 
      TestBed.runInInjectionContext(() => guestGuard(...guardParameters));

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

  it('should redirect to /dashboard if authenticated', (done) => {
    mockAuthService.isAuthenticated.and.returnValue(of({ user: { status: 'active' } } as any));
    const urlTree = {} as UrlTree;
    mockRouter.createUrlTree.and.returnValue(urlTree);

    const result = executeGuard({} as any, {} as any);
    if (typeof result !== 'boolean' && 'subscribe' in (result as any)) {
      (result as any).subscribe((res: any) => {
        expect(res).toBe(urlTree);
        expect(mockRouter.createUrlTree).toHaveBeenCalledWith(['/dashboard']);
        done();
      });
    }
  });

  it('should return true if not authenticated', (done) => {
    mockAuthService.isAuthenticated.and.returnValue(throwError(() => new Error('Unauthorized')));

    const result = executeGuard({} as any, {} as any);
    if (typeof result !== 'boolean' && 'subscribe' in (result as any)) {
      (result as any).subscribe((res: any) => {
        expect(res).toBe(true);
        done();
      });
    }
  });
});
