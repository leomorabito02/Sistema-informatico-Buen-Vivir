import { Component, OnInit, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { ImpactService, ImpactMetrics } from "../../../services/impact.service";

@Component({
  selector: "app-impact",
  standalone: true,
  imports: [CommonModule],
  templateUrl: "./impact.component.html",
  styleUrl: "./impact.component.css",
})
export class ImpactComponent implements OnInit {
  private readonly impactService = inject(ImpactService);

  activeParticipants: number | string = 248;
  volunteers: number | string = 32;
  courses: number | string = "Pendiente";

  ngOnInit(): void {
    this.impactService.getImpactMetrics().subscribe((data: ImpactMetrics) => {
      if (data) {
        if (data.activeParticipants !== undefined) {
          this.activeParticipants = data.activeParticipants;
        }
        if (data.volunteers !== undefined) {
          this.volunteers = data.volunteers;
        }
        // User requested explicitly: "cambia el número por 'pendiente' para recordarmelo"
        this.courses = "Pendiente";
      }
    });
  }
}
