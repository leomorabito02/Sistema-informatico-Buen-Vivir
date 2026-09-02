import { ComponentFixture, TestBed } from "@angular/core/testing";
import { UsuariosComponent } from "./usuarios.component";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";

describe("UsuariosComponent", () => {
  let component: UsuariosComponent;
  let fixture: ComponentFixture<UsuariosComponent>;

  const mockUsers = [
    {
      name: "Alice",
      email: "alice@example.com",
      role: "Responsable",
      status: "Activo",
    },
    {
      name: "Bob",
      email: "bob@example.com",
      role: "Docente",
      status: "Inactivo",
    },
    {
      name: "Charlie",
      email: "charlie@example.com",
      role: "Empresa",
      status: "Activo",
    },
    {
      name: "David",
      email: "david@example.com",
      role: "Voluntario",
      status: "Activo",
    },
    {
      name: "Eve",
      email: "eve@example.com",
      role: "Participante",
      status: "Activo",
    },
    {
      name: "Frank",
      email: "frank@example.com",
      role: "OtroRol",
      status: "Inactivo",
    },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UsuariosComponent, BrowserAnimationsModule],
    }).compileComponents();

    fixture = TestBed.createComponent(UsuariosComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should filter users by name, email or role", () => {
    component.users = mockUsers;

    component.searchTerm = "alice";
    component.filterUsers();
    expect(component.filteredUsers).toHaveSize(1);
    expect(component.filteredUsers[0].name).toBe("Alice");

    component.searchTerm = "bob@example.com";
    component.filterUsers();
    expect(component.filteredUsers).toHaveSize(1);
    expect(component.filteredUsers[0].name).toBe("Bob");

    component.searchTerm = "Voluntario";
    component.filterUsers();
    expect(component.filteredUsers).toHaveSize(1);
    expect(component.filteredUsers[0].name).toBe("David");
  });

  it("should return correct role severities", () => {
    expect(component.getRoleSeverity("Responsable")).toBe("danger");
    expect(component.getRoleSeverity("Docente")).toBe("warning");
    expect(component.getRoleSeverity("Empresa")).toBe("contrast");
    expect(component.getRoleSeverity("Voluntario")).toBe("info");
    expect(component.getRoleSeverity("Participante")).toBe("success");
    expect(component.getRoleSeverity("Desconocido")).toBeUndefined();
  });

  it("should return correct status severities", () => {
    expect(component.getStatusSeverity("Activo")).toBe("success");
    expect(component.getStatusSeverity("Inactivo")).toBe("warning");
  });
});
