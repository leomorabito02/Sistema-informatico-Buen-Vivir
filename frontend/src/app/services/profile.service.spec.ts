import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { ProfileService } from './profile.service';
import { environment } from '../../environments/environment';

describe('ProfileService', () => {
  let service: ProfileService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [ProfileService]
    });
    service = TestBed.inject(ProfileService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('updateMyPassword should call put with correct body', () => {
    const mockResponse = { success: true };

    service.updateMyPassword('old', 'new').subscribe(res => {
      expect(res).toEqual(mockResponse);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/profile/me/password`);
    expect(req.request.method).toBe('PUT');
    expect(req.request.body).toEqual({ old_password: 'old', new_password: 'new' });
    req.flush(mockResponse);
  });

  it('linkGoogleAccount should post id_token and google_id', () => {
    const mockResponse = { success: true };

    service.linkGoogleAccount('token', 'gid').subscribe(res => {
      expect(res).toEqual(mockResponse);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/profile/me/link-google`);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ id_token: 'token', google_id: 'gid' });
    req.flush(mockResponse);
  });

  it('unlinkGoogleAccount should call post endpoint', () => {
    const mockResponse = { success: true };

    service.unlinkGoogleAccount().subscribe(res => {
      expect(res).toEqual(mockResponse);
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/api/v1/profile/me/unlink-google`);
    expect(req.request.method).toBe('POST');
    req.flush(mockResponse);
  });
});
