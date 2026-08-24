import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MiPerfilComponent } from './mi-perfil.component';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { Auth } from '@angular/fire/auth';

describe('MiPerfilComponent', () => {
  let component: MiPerfilComponent;
  let fixture: ComponentFixture<MiPerfilComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MiPerfilComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(MiPerfilComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
