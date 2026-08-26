import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { AuthService } from './auth.service';
import { provideHttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Auth } from '@angular/fire/auth';
import { Router } from '@angular/router';

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;
  let mockRouter: jasmine.SpyObj<Router>;

  beforeEach(() => {
    environment.apiUrl = 'http://localhost:3000';
    mockRouter = jasmine.createSpyObj('Router', ['navigate']);

    TestBed.configureTestingModule({
      providers: [
        AuthService,
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: Auth, useValue: { signOut: jasmine.createSpy('signOut') } },
        { provide: Router, useValue: mockRouter }
      ]
    });

    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should send POST to checkCredentials with credentials', () => {
    service.checkCredentials('test@test.com', 'password123').subscribe();

    const req = httpMock.expectOne('http://localhost:3000/api/v1/auth/check-credentials');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ email: 'test@test.com', password: 'password123' });
    expect(req.request.withCredentials).toBe(true);
    req.flush({ accounts: [], login_nonce: 'nonce' });
  });

  it('should send POST to loginWithAccount with credentials', () => {
    service.loginWithAccount('usr1', 'r1', 'nonce123').subscribe();

    const req = httpMock.expectOne('http://localhost:3000/api/v1/auth/login');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ user_id: 'usr1', role_id: 'r1', login_nonce: 'nonce123' });
    expect(req.request.withCredentials).toBe(true);
    req.flush({ token: 'jwt' });
  });

  it('should send POST to registerWithEmail with credentials', () => {
    const formData = {
      first_name: 'John',
      last_name: 'Doe',
      email: 'john@doe.com',
      phone: '123456',
      password: 'Password1!',
      role: 'volunteer'
    };

    service.registerWithEmail(formData).subscribe();

    const req = httpMock.expectOne('http://localhost:3000/api/v1/auth/register');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(formData);
    expect(req.request.withCredentials).toBe(true);
    req.flush({ message: 'Success' });
  });

  it('should send GET to isAuthenticated with credentials', () => {
    service.isAuthenticated().subscribe();

    const req = httpMock.expectOne('http://localhost:3000/api/v1/auth/me');
    expect(req.request.method).toBe('GET');
    expect(req.request.withCredentials).toBe(true);
    req.flush({ user_id: 'usr1' });
  });

  it('should send PUT to updateProfile with credentials', () => {
    service.updateProfile({ dni: '123' }).subscribe();

    const req = httpMock.expectOne('http://localhost:3000/api/v1/profile/me');
    expect(req.request.method).toBe('PUT');
    expect(req.request.body).toEqual({ dni: '123' });
    expect(req.request.withCredentials).toBe(true);
    req.flush({ message: 'Updated' });
  });

  it('should send POST to logout, sign out from Auth and navigate to login on success', () => {
    service.logout();

    const req = httpMock.expectOne('http://localhost:3000/api/v1/auth/logout');
    expect(req.request.method).toBe('POST');
    expect(req.request.withCredentials).toBe(true);
    req.flush({});

    expect(mockRouter.navigate).toHaveBeenCalledWith(['/login']);
  });

  it('should sign out from Auth and navigate to login even if logout request fails', () => {
    service.logout();

    const req = httpMock.expectOne('http://localhost:3000/api/v1/auth/logout');
    req.flush('Error', { status: 500, statusText: 'Internal Server Error' });

    expect(mockRouter.navigate).toHaveBeenCalledWith(['/login']);
  });
});
