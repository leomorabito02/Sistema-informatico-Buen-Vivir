import { ComponentFixture, TestBed } from '@angular/core/testing';
import { AsistenciaComponent } from './asistencia.component';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { Auth } from '@angular/fire/auth';

describe('AsistenciaComponent', () => {
  let component: AsistenciaComponent;
  let fixture: ComponentFixture<AsistenciaComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AsistenciaComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(AsistenciaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
