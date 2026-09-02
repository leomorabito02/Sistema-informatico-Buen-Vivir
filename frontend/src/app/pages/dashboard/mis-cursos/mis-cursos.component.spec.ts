import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MisCursosComponent } from './mis-cursos.component';
import { EducationService } from '../../../services/education.service';
import { MessageService } from 'primeng/api';
import { of, throwError } from 'rxjs';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { Auth } from '@angular/fire/auth';

describe('MisCursosComponent', () => {
  let component: MisCursosComponent;
  let fixture: ComponentFixture<MisCursosComponent>;
  let educationServiceSpy: jasmine.SpyObj<EducationService>;
  let messageServiceSpy: jasmine.SpyObj<MessageService>;

  beforeEach(async () => {
    educationServiceSpy = jasmine.createSpyObj('EducationService', ['listMyCourses', 'listMyDraftCourses', 'createMaterial', 'uploadMaterial', 'updateMyCourse']);
    messageServiceSpy = jasmine.createSpyObj('MessageService', ['add']);
    
    educationServiceSpy.listMyCourses.and.returnValue(of([]));
    educationServiceSpy.listMyDraftCourses.and.returnValue(of([]));

    await TestBed.configureTestingModule({
      imports: [MisCursosComponent, BrowserAnimationsModule],
      providers: [
        { provide: EducationService, useValue: educationServiceSpy },
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(MisCursosComponent);
    component = fixture.componentInstance;
    
    const msgService = fixture.debugElement.injector.get(MessageService);
    messageServiceSpy = msgService as jasmine.SpyObj<MessageService>;
    spyOn(msgService, 'add');
    
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should load my courses on init', () => {
    const published: any[] = [{ id: '1', title: 'Course 1' }];
    const drafts: any[] = [{ id: '2', title: 'Draft 1' }];
    educationServiceSpy.listMyCourses.and.returnValue(of(published));
    educationServiceSpy.listMyDraftCourses.and.returnValue(of(drafts));
    
    component.ngOnInit();
    
    expect(educationServiceSpy.listMyCourses).toHaveBeenCalled();
    expect(educationServiceSpy.listMyDraftCourses).toHaveBeenCalled();
    expect(component.courses).toHaveSize(2);
    expect(component.courses[0].publicado).toBeTrue();
    expect(component.courses[1].publicado).toBeFalse();
    expect(component.loading).toBeFalse();
  });

  it('should handle error when loading courses', () => {
    educationServiceSpy.listMyCourses.and.returnValue(throwError(() => new Error('Load error')));
    
    component.ngOnInit();
    
    expect(component.error).toBe('Load error');
    expect(component.loading).toBeFalse();
  });

  describe('Materials', () => {
    it('should open materials dialog', () => {
      const c: any = { id: '1' };
      component.openMaterials(c);
      
      expect(component.matCourse).toBe(c);
      expect(component.matDialog).toBeTrue();
    });

    it('should validate url', () => {
      component.matForm.url = 'invalid';
      expect(component.matUrlInvalid).toBeTrue();
      component.matForm.url = 'http://test.com';
      expect(component.matUrlInvalid).toBeFalse();
    });

    it('should add material via url', () => {
      const c: any = { id: '1', title: 'Course 1', raw: { id: '1' } };
      const newMat: any = { id: 'mat1', title: 'New Mat' };
      
      component.matCourse = c;
      component.matForm.title = 'Test';
      component.matForm.url = 'http://test.com';
      
      educationServiceSpy.createMaterial.and.returnValue(of(newMat));
      educationServiceSpy.updateMyCourse.and.returnValue(of({ ...c.raw, material: [newMat] }));
      
      component.addMaterial();
      
      expect(educationServiceSpy.createMaterial).toHaveBeenCalled();
      expect(educationServiceSpy.updateMyCourse).toHaveBeenCalled();
      expect(messageServiceSpy.add).toHaveBeenCalledWith(jasmine.objectContaining({ severity: 'success' }));
    });
  });
});
