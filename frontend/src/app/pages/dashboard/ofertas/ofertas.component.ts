import { Component, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { CardModule } from "primeng/card";
import { ButtonModule } from "primeng/button";
import { TagModule } from "primeng/tag";
import { InputTextModule } from "primeng/inputtext";
import { DialogModule } from "primeng/dialog";

interface JobOffer {
  id: number;
  title: string;
  company: string;
  location: string;
  type: string;
  date: string;
  status: string;
  description?: string;
}

@Component({
  selector: "app-ofertas",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    CardModule,
    ButtonModule,
    TagModule,
    InputTextModule,
    DialogModule,
  ],
  templateUrl: "./ofertas.component.html",
  styleUrls: ["./ofertas.component.css"],
})
export class OfertasComponent implements OnInit {
  searchTerm: string = "";
  showDialog = false;

  newJob = {
    title: "",
    company: "",
    location: "",
    type: "Tiempo Completo",
    description: "",
  };

  jobs: JobOffer[] = [];
  filteredJobs: JobOffer[] = [];

  ngOnInit() {}

  filterJobs() {
    const term = this.searchTerm.toLowerCase().trim();
    this.filteredJobs = this.jobs.filter(
      (j) =>
        j.title.toLowerCase().includes(term) ||
        j.company.toLowerCase().includes(term) ||
        j.location.toLowerCase().includes(term),
    );
  }

  openNewJobDialog() {
    this.showDialog = true;
  }

  saveJob() {
    if (!this.newJob.title || !this.newJob.company) return;
    this.jobs.unshift({
      id: this.jobs.length + 1,
      title: this.newJob.title,
      company: this.newJob.company,
      location: this.newJob.location || "No especificada",
      type: this.newJob.type,
      date: "Hoy",
      status: "Activa",
    });
    this.filterJobs();
    this.showDialog = false;
    this.newJob = {
      title: "",
      company: "",
      location: "",
      type: "Tiempo Completo",
      description: "",
    };
  }
}
