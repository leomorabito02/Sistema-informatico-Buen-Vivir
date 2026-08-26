import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HeaderComponent } from './header.component';
import { AuthService } from '../../../services/auth.service';
import { provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';
import { Bell, Settings, LogOut } from 'lucide-angular';

describe('HeaderComponent', () => {
  let component: HeaderComponent;
  let fixture: ComponentFixture<HeaderComponent>;
  let mockAuthService: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    mockAuthService = jasmine.createSpyObj('AuthService', ['isAuthenticated', 'logout']);
    mockAuthService.isAuthenticated.and.returnValue(of({ role_id: 'role_participant' } as any));

    await TestBed.configureTestingModule({
      imports: [HeaderComponent],
      providers: [
        { provide: AuthService, useValue: mockAuthService },
        provideRouter([])
      ]
    }).compileComponents();
  });

  it('should create and set roleName on init', () => {
    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component).toBeTruthy();
    expect(component.roleName).toBe('participante');
    expect(component.Bell).toBe(Bell);
    expect(component.Settings).toBe(Settings);
    expect(component.LogOut).toBe(LogOut);
  });

  it('should read role_id from user object if root role_id is missing', () => {
    mockAuthService.isAuthenticated.and.returnValue(of({ user: { role_id: 'role_admin' } } as any));

    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.roleName).toBe('responsable');
  });

  it('should fallback to empty roleName if isAuthenticated fails', () => {
    mockAuthService.isAuthenticated.and.returnValue(throwError(() => new Error('Unauthorized')));

    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    expect(component.roleName).toBe('');
  });

  it('should map roleId to name correctly for all roles', () => {
    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;

    expect(component.mapRoleIdToName('')).toBe('');
    expect(component.mapRoleIdToName('role_participant')).toBe('participante');
    expect(component.mapRoleIdToName('participante')).toBe('participante');
    expect(component.mapRoleIdToName('role_admin')).toBe('responsable');
    expect(component.mapRoleIdToName('responsable')).toBe('responsable');
    expect(component.mapRoleIdToName('role_volunteer')).toBe('voluntario');
    expect(component.mapRoleIdToName('voluntario')).toBe('voluntario');
    expect(component.mapRoleIdToName('role_teacher')).toBe('docente');
    expect(component.mapRoleIdToName('docente')).toBe('docente');
    expect(component.mapRoleIdToName('role_company')).toBe('empresa');
    expect(component.mapRoleIdToName('empresa')).toBe('empresa');
    expect(component.mapRoleIdToName('role_custom')).toBe('custom');
  });

  it('should call authService.logout on logout()', () => {
    fixture = TestBed.createComponent(HeaderComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();

    component.logout();
    expect(mockAuthService.logout).toHaveBeenCalled();
  });
});
