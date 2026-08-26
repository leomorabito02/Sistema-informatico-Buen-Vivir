import { Component, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { CardModule } from "primeng/card";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { ProgressBarModule } from "primeng/progressbar";

interface Campaign {
  name: string;
  goal: number;
  raised: number;
  status: string;
  startDate?: string;
  endDate?: string;
  volunteers?: number;
}

@Component({
  selector: "app-campanas",
  standalone: true,
  imports: [
    CommonModule,
    CardModule,
    ButtonModule,
    TagModule,
    ProgressBarModule,
  ],
  templateUrl: "./campanas.component.html",
  styleUrls: ["./campanas.component.css"],
})
export class CampanasComponent implements OnInit {
  campaigns: Campaign[] = [];

  ngOnInit() {}
}
