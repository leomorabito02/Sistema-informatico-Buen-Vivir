import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CampanasComponent } from './campanas.component';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { Auth } from '@angular/fire/auth';

describe('CampanasComponent', () => {
  let component: CampanasComponent;
  let fixture: ComponentFixture<CampanasComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CampanasComponent, BrowserAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: Auth, useValue: {} }
      ]
    }).compileComponents();

    fixture = TestBed.createComponent(CampanasComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
