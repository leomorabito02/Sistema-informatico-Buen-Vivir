import { ComponentFixture, TestBed } from '@angular/core/testing';
import { AsistenciaComponent } from './asistencia.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { EducationService } from '../../../services/education.service';
import { of, throwError } from 'rxjs';
import { Activity, Attendance, Course } from '../../../core/models/education.model';

describe('AsistenciaComponent', () => {
  let component: AsistenciaComponent;
  let fixture: ComponentFixture<AsistenciaComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;

  const mockCourses: Course[] = [
    { id: 'c1', title: 'Course 1', description: null, mode: 'Virtual' },
    { id: 'c2', title: '', description: null, mode: 'Presencial' },
  ];

  const mockActivities: Activity[] = [
    { id: 'a1', cursoId: 'c1', titulo: 'Act 1', fecha: '2026-01-01', createdAt: '2026-01-01' },
  ];

  const mockAttendance: Attendance[] = [
    { id: 'att1', actividadId: 'a1', participanteId: 'p1', presente: true, fecha: '2026-01-01' },
  ];

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj('EducationService', [
      'listMyCourses',
      'listActivities',
      'createActivity',
      'listAttendance',
      'registerAttendance',
    ]);

    educationServiceSpy.listMyCourses.and.returnValue(of(mockCourses));
    educationServiceSpy.listActivities.and.returnValue(of(mockActivities));
    educationServiceSpy.createActivity.and.returnValue(of(mockActivities[0]));
    educationServiceSpy.listAttendance.and.returnValue(of(mockAttendance));
    educationServiceSpy.registerAttendance.and.returnValue(of({ actividadId: 'a1', status: 'ok' }));

    await TestBed.configureTestingModule({
      imports: [AsistenciaComponent, BrowserAnimationsModule],
      providers: [
        { provide: EducationService, useValue: educationServiceSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(AsistenciaComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should load courses on init (>1 courses)', () => {
    fixture.detectChanges();
    expect(educationServiceSpy.listMyCourses).toHaveBeenCalled();
    expect(component.courseOptions).toHaveSize(2);
    expect(component.loadingCourses).toBeFalse();
    expect(component.selectedCourseId).toBeNull();
  });

  it('should auto-select course if only 1 course is returned', () => {
    educationServiceSpy.listMyCourses.and.returnValue(of([mockCourses[0]]));
    fixture.detectChanges();
    expect(component.selectedCourseId).toBe('c1');
    expect(educationServiceSpy.listActivities).toHaveBeenCalledWith('c1');
  });

  it('should handle error when loading courses', () => {
    educationServiceSpy.listMyCourses.and.returnValue(throwError(() => new Error('Error loading courses')));
    fixture.detectChanges();
    expect(component.error).toBe('Error loading courses');
    expect(component.loadingCourses).toBeFalse();
  });

  it('should handle onCourseChange when no course selected', () => {
    component.selectedCourseId = null;
    component.onCourseChange();
    expect(educationServiceSpy.listActivities).not.toHaveBeenCalled();
  });

  it('should handle error in onCourseChange', () => {
    educationServiceSpy.listActivities.and.returnValue(throwError(() => new Error('Error activities')));
    component.selectedCourseId = 'c1';
    component.onCourseChange();
    expect(component.error).toBe('Error activities');
    expect(component.loadingActivities).toBeFalse();
  });

  it('should crearActividad return early if required fields are missing', () => {
    component.selectedCourseId = null;
    component.crearActividad();
    expect(educationServiceSpy.createActivity).not.toHaveBeenCalled();

    component.selectedCourseId = 'c1';
    component.nuevaActividadTitulo = '';
    component.crearActividad();
    expect(educationServiceSpy.createActivity).not.toHaveBeenCalled();
  });

  it('should crearActividad successfully', () => {
    component.selectedCourseId = 'c1';
    component.nuevaActividadTitulo = 'Nueva';
    component.nuevaActividadFecha = new Date('2026-05-01');

    component.crearActividad();

    expect(educationServiceSpy.createActivity).toHaveBeenCalled();
    expect(component.activities).toHaveSize(1);
    expect(component.info).toBe('Actividad creada.');
    expect(component.nuevaActividadTitulo).toBe('');
    expect(component.nuevaActividadFecha).toBeNull();
  });

  it('should handle error in crearActividad', () => {
    educationServiceSpy.createActivity.and.returnValue(throwError(() => new Error('Error creando')));
    component.selectedCourseId = 'c1';
    component.nuevaActividadTitulo = 'Nueva';
    component.nuevaActividadFecha = new Date('2026-05-01');

    component.crearActividad();

    expect(component.error).toBe('Error creando');
    expect(component.creandoActividad).toBeFalse();
  });

  it('should selectActivity successfully', () => {
    component.selectActivity('a1');
    expect(component.selectedActivityId).toBe('a1');
    expect(component.attendance).toEqual(mockAttendance);
    expect(component.draft).toEqual([{ participanteId: 'p1', presente: true }]);
    expect(component.loadingAttendance).toBeFalse();
  });

  it('should handle error in selectActivity', () => {
    educationServiceSpy.listAttendance.and.returnValue(throwError(() => new Error('Error attendance')));
    component.selectActivity('a1');
    expect(component.error).toBe('Error attendance');
    expect(component.loadingAttendance).toBeFalse();
  });

  it('should addRow and removeRow from draft', () => {
    component.draft = [];
    component.addRow();
    expect(component.draft).toHaveSize(1);
    expect(component.draft[0]).toEqual({ participanteId: '', presente: true });

    component.removeRow(0);
    expect(component.draft).toHaveSize(0);
  });

  it('should guard en guardar() when no activity selected or draft empty', () => {
    component.selectedActivityId = null;
    component.guardar();
    expect(educationServiceSpy.registerAttendance).not.toHaveBeenCalled();

    component.selectedActivityId = 'a1';
    component.draft = [{ participanteId: '   ', presente: true }];
    component.guardar();
    expect(educationServiceSpy.registerAttendance).not.toHaveBeenCalled();
  });

  it('should guardar attendance successfully', () => {
    component.selectedActivityId = 'a1';
    component.draft = [{ participanteId: 'p1', presente: true }];

    component.guardar();

    expect(educationServiceSpy.registerAttendance).toHaveBeenCalledWith('a1', [{ participanteId: 'p1', presente: true }]);
    expect(educationServiceSpy.listAttendance).toHaveBeenCalledWith('a1');
  });

  it('should handle error in guardar', () => {
    educationServiceSpy.registerAttendance.and.returnValue(throwError(() => new Error('Error guardar')));
    component.selectedActivityId = 'a1';
    component.draft = [{ participanteId: 'p1', presente: true }];

    component.guardar();

    expect(component.error).toBe('Error guardar');
    expect(component.guardando).toBeFalse();
  });
});
