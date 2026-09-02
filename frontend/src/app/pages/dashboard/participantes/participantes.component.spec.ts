import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ParticipantesComponent } from './participantes.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { EducationService } from '../../../services/education.service';
import { of, throwError } from 'rxjs';
import { Course, Enrollment } from '../../../core/models/education.model';

describe('ParticipantesComponent', () => {
  let component: ParticipantesComponent;
  let fixture: ComponentFixture<ParticipantesComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;

  const mockCourses: Course[] = [
    { id: 'c1', title: 'Course 1', description: null, mode: 'Virtual' },
    { id: 'c2', title: 'Course 2', description: null, mode: 'Presencial' },
  ];

  const mockEnrollments: Enrollment[] = [
    { id: 'e1', cursoId: 'c1', participanteId: 'user_juan', estado: 'activa', fechaInscripcion: '2026-01-01' },
    { id: 'e2', cursoId: 'c1', participanteId: 'user_maria', estado: 'finalizada', fechaInscripcion: '2026-01-01' },
    { id: 'e3', cursoId: 'c1', participanteId: 'user_pedro', estado: 'baja', fechaInscripcion: '2026-01-01' },
  ];

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj('EducationService', [
      'listMyCourses',
      'listEnrollments',
    ]);

    educationServiceSpy.listMyCourses.and.returnValue(of(mockCourses));
    educationServiceSpy.listEnrollments.and.returnValue(of(mockEnrollments));

    await TestBed.configureTestingModule({
      imports: [ParticipantesComponent, BrowserAnimationsModule],
      providers: [{ provide: EducationService, useValue: educationServiceSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(ParticipantesComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should load courses on init (>1 courses)', () => {
    fixture.detectChanges();
    expect(component.courseOptions).toHaveSize(2);
    expect(component.loadingCourses).toBeFalse();
  });

  it('should auto-select course if only 1 course is returned', () => {
    educationServiceSpy.listMyCourses.and.returnValue(of([mockCourses[0]]));
    fixture.detectChanges();
    expect(component.selectedCourseId).toBe('c1');
    expect(educationServiceSpy.listEnrollments).toHaveBeenCalledWith('c1');
  });

  it('should handle error when loading courses', () => {
    educationServiceSpy.listMyCourses.and.returnValue(throwError(() => new Error('Error loading courses')));
    fixture.detectChanges();
    expect(component.error).toBe('Error loading courses');
    expect(component.loadingCourses).toBeFalse();
  });

  it('should return early onCourseChange if no course selected', () => {
    component.selectedCourseId = null;
    component.onCourseChange();
    expect(educationServiceSpy.listEnrollments).not.toHaveBeenCalled();
  });

  it('should fetch enrollments onCourseChange', () => {
    component.selectedCourseId = 'c1';
    component.onCourseChange();

    expect(component.enrollments).toHaveSize(3);
    expect(component.filtered).toHaveSize(3);
    expect(component.loadingEnrollments).toBeFalse();
  });

  it('should handle error in onCourseChange', () => {
    educationServiceSpy.listEnrollments.and.returnValue(throwError(() => new Error('Error enrollments')));
    component.selectedCourseId = 'c1';
    component.onCourseChange();

    expect(component.error).toBe('Error enrollments');
    expect(component.loadingEnrollments).toBeFalse();
  });

  it('should filter enrollments by participanteId or estado', () => {
    component.enrollments = mockEnrollments;

    component.searchTerm = 'juan';
    component.filter();
    expect(component.filtered).toHaveSize(1);
    expect(component.filtered[0].participanteId).toBe('user_juan');

    component.searchTerm = 'baja';
    component.filter();
    expect(component.filtered).toHaveSize(1);
    expect(component.filtered[0].participanteId).toBe('user_pedro');
  });

  it('should return correct status severities', () => {
    expect(component.estadoSeverity('activa')).toBe('success');
    expect(component.estadoSeverity('finalizada')).toBe('info');
    expect(component.estadoSeverity('baja')).toBe('danger');
    expect(component.estadoSeverity('desconocido')).toBe('secondary');
  });
});
