import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HeroComponent } from '../../components/landing/hero/hero.component';
import { PillarsComponent } from '../../components/landing/pillars/pillars.component';
import { ImpactComponent } from '../../components/landing/impact/impact.component';
import { HowItWorksComponent } from '../../components/landing/how-it-works/how-it-works.component';
import { LibraryComponent } from '../../components/landing/library/library.component';
import { LocationComponent } from '../../components/landing/location/location.component';
import { FooterComponent } from '../../components/shared/footer/footer.component';

@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [
    CommonModule,
    HeroComponent,
    PillarsComponent,
    ImpactComponent,
    HowItWorksComponent,
    LibraryComponent,
    LocationComponent,
    FooterComponent
  ],
  templateUrl: './landing.component.html',
  styleUrl: './landing.component.css'
})
export class LandingComponent {

}
