import { Component, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";
import { CardModule } from "primeng/card";
import { ButtonModule } from "primeng/button";
import { InputTextModule } from "primeng/inputtext";

@Component({
  selector: "app-cv",
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    CardModule,
    ButtonModule,
    InputTextModule,
  ],
  templateUrl: "./cv.component.html",
  styleUrls: ["./cv.component.css"],
})
export class CvComponent implements OnInit {
  profile = {
    fullName: "",
    email: "",
    phone: "",
    address: "",
    summary: "",
    skills: "",
    experience: "",
  };

  ngOnInit() {}

  downloadPdf() {
    alert("Generando y descargando versión PDF de tu CV...");
  }
}
