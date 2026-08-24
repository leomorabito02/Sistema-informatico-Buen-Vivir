import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SeguimientoComponent } from './seguimiento.component';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { Auth } from '@angular/fire/auth';

describe('SeguimientoComponent', () => {
  let component: SeguimientoComponent;
  let fixture: ComponentFixture<SeguimientoComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SeguimientoComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(SeguimientoComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
