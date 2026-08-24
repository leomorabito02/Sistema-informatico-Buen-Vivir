import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CursosComponent } from './cursos.component';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { Auth } from '@angular/fire/auth';

describe('CursosComponent', () => {
  let component: CursosComponent;
  let fixture: ComponentFixture<CursosComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CursosComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(CursosComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
