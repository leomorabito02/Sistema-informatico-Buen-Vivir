import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { AuthService } from './auth.service';
import { Auth } from '@angular/fire/auth';
import { RouterTestingModule } from '@angular/router/testing';
import { environment } from '../../environments/environment';

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;
  let authSpy: any;

  beforeEach(() => {
    authSpy = jasmine.createSpyObj('Auth', ['signOut']);

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule, RouterTestingModule],
      providers: [
        AuthService,
        { provide: Auth, useValue: authSpy }
      ]
    });
    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('checkCredentials should return response on 200', () => {
    const mockResponse = { accounts: [], login_nonce: '123' };

    service.checkCredentials('test@test.com', 'password').subscribe(res => {
      expect(res).toEqual(mockResponse as any);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/auth/check-credentials`);
    expect(req.request.method).toBe('POST');
    req.flush(mockResponse);
  });

  it('checkCredentials should throw on 401', () => {
    service.checkCredentials('test@test.com', 'wrong').subscribe({
      next: () => fail('should have failed with 401'),
      error: (err) => {
        expect(err.message).toBe('No autorizado. Verifique sus credenciales.');
      }
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/auth/check-credentials`);
    req.flush('Unauthorized', { status: 401, statusText: 'Unauthorized' });
  });

  it('loginWithAccount should return token on success', () => {
    const mockResponse = { token: 'tkn', user: {} };

    service.loginWithAccount('usr-1', 'role-1', 'nonce').subscribe(res => {
      expect(res).toEqual(mockResponse as any);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/auth/login`);
    expect(req.request.method).toBe('POST');
    req.flush(mockResponse);
  });

  it('registerWithEmail should handle 409 conflict', () => {
    service.registerWithEmail({ email: 'e', password: 'p', role: 'r', first_name: '', last_name: '', phone: '' }).subscribe({
      next: () => fail('should have failed with 409'),
      error: (err) => {
        expect(err.message).toBe('Email already exists');
      }
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/auth/register`);
    req.flush({ message: 'Email already exists' }, { status: 409, statusText: 'Conflict' });
  });

  it('logout should navigate to /login even on error', () => {
    service.logout();

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/auth/logout`);
    req.flush('Error', { status: 500, statusText: 'Server Error' });

    expect(authSpy.signOut).toHaveBeenCalled();
  });
});
