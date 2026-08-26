import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { authInterceptor } from './auth.interceptor';

describe('authInterceptor', () => {
  let http: HttpClient;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([authInterceptor])),
        provideHttpClientTesting()
      ]
    });

    http = TestBed.inject(HttpClient);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should clone request and set withCredentials to true', () => {
    http.get('/api/test').subscribe();

    const req = httpMock.expectOne('/api/test');
    expect(req.request.withCredentials).toBe(true);
    expect(req.request.method).toBe('GET');
    expect(req.request.url).toBe('/api/test');
    req.flush({});
  });

  it('should set withCredentials to true for POST requests with body', () => {
    http.post('/api/data', { foo: 'bar' }).subscribe();

    const req = httpMock.expectOne('/api/data');
    expect(req.request.withCredentials).toBe(true);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ foo: 'bar' });
    req.flush({});
  });
});
