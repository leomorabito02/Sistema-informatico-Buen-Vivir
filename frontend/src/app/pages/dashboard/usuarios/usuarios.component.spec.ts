import { ComponentFixture, TestBed } from "@angular/core/testing";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { of, throwError } from "rxjs";
import { DashboardUser, UsuariosComponent } from "./usuarios.component";
import { AuthService } from "../../../services/auth.service";
import { UsersService } from "../../../services/users.service";
import { MeResponse, User } from "../../../core/models/auth.model";

function apiUser(overrides: Partial<User>): User {
  return {
    user_id: "u",
    first_name: "",
    last_name: "",
    email: "",
    phone: "",
    status: "active",
    role_id: "role_participant",
    twofa_enabled: false,
    created_at: "",
    updated_at: "",
    ...overrides,
  };
}

describe("UsuariosComponent", () => {
  let component: UsuariosComponent;
  let fixture: ComponentFixture<UsuariosComponent>;
  let authService: jasmine.SpyObj<AuthService>;
  let usersService: jasmine.SpyObj<UsersService>;

  const mockUsers: DashboardUser[] = [
    {
      name: "Alice",
      email: "alice@example.com",
      phone: "111",
      role: "Responsable",
      status: "Activo",
    },
    {
      name: "Bob",
      email: "bob@example.com",
      phone: "222",
      role: "Docente",
      status: "Pendiente",
    },
    {
      name: "Charlie",
      email: "charlie@example.com",
      phone: "333",
      role: "Empresa",
      status: "Activo",
    },
    {
      name: "David",
      email: "david@example.com",
      phone: "444",
      role: "Voluntario",
      status: "Activo",
    },
    {
      name: "Eve",
      email: "eve@example.com",
      phone: "555",
      role: "Participante",
      status: "Activo",
    },
  ];

  function setup(me: MeResponse | null, users: User[] | null) {
    authService = jasmine.createSpyObj<AuthService>("AuthService", [
      "isAuthenticated",
    ]);
    usersService = jasmine.createSpyObj<UsersService>("UsersService", [
      "listUsers",
    ]);
    authService.isAuthenticated.and.returnValue(
      me ? of(me) : throwError(() => new Error("sin sesión")),
    );
    usersService.listUsers.and.returnValue(
      users ? of(users) : throwError(() => new Error("auth caído")),
    );

    TestBed.configureTestingModule({
      imports: [UsuariosComponent, BrowserAnimationsModule],
      providers: [
        { provide: AuthService, useValue: authService },
        { provide: UsersService, useValue: usersService },
      ],
    });

    fixture = TestBed.createComponent(UsuariosComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  it("should create and map API users to dashboard rows", () => {
    setup({ user: apiUser({}), role_id: "role_admin" }, [
      apiUser({
        user_id: "u1",
        first_name: "Ana",
        last_name: "Pérez",
        email: "ana@x.com",
        phone: "123",
        role_id: "role_admin",
      }),
      apiUser({
        user_id: "u2",
        email: "solo@mail.com",
        role_id: "role_teacher",
        status: "pending",
      }),
      apiUser({ user_id: "u3", first_name: "Vol", role_id: "role_volunteer" }),
      apiUser({ user_id: "u4", first_name: "Emp", role_id: "role_company" }),
      apiUser({ user_id: "u5", first_name: "Par", role_id: "" }),
    ]);

    expect(component).toBeTruthy();
    expect(component.isVolunteerOnly).toBeFalse();
    expect(component.users.map((u) => u.role)).toEqual([
      "Responsable",
      "Docente",
      "Voluntario",
      "Empresa",
      "Participante",
    ]);
    expect(component.users[0]).toEqual(
      jasmine.objectContaining({
        id: "u1",
        name: "Ana Pérez",
        email: "ana@x.com",
        phone: "123",
        status: "Activo",
      }),
    );
    // sin nombre se muestra el email; "pending" se muestra como Pendiente
    expect(component.users[1].name).toBe("solo@mail.com");
    expect(component.users[1].status).toBe("Pendiente");
    // teléfono vacío se muestra como "-"
    expect(component.users[2].phone).toBe("-");
    expect(component.filteredUsers.length).toBe(5);
  });

  it("should fall back to mock users when the API returns nothing or fails", () => {
    setup({ user: apiUser({}), role_id: "role_admin" }, []);
    expect(component.users).toEqual(component.initialMockUsers);

    TestBed.resetTestingModule();
    setup({ user: apiUser({}), role_id: "role_admin" }, null);
    expect(component.users).toEqual(component.initialMockUsers);
  });

  it("should still load users when the session check fails", () => {
    setup(null, [apiUser({ first_name: "Ana" })]);
    expect(usersService.listUsers).toHaveBeenCalled();
    expect(component.users.length).toBe(1);
  });

  it("should filter users by name, email, phone or role", () => {
    setup({ user: apiUser({}), role_id: "role_admin" }, []);
    component.users = mockUsers;

    component.searchTerm = "alice";
    component.filterUsers();
    expect(component.filteredUsers.map((u) => u.name)).toEqual(["Alice"]);

    component.searchTerm = "bob@example.com";
    component.filterUsers();
    expect(component.filteredUsers.map((u) => u.name)).toEqual(["Bob"]);

    component.searchTerm = "444";
    component.filterUsers();
    expect(component.filteredUsers.map((u) => u.name)).toEqual(["David"]);

    component.searchTerm = "Voluntario";
    component.filterUsers();
    expect(component.filteredUsers.map((u) => u.name)).toEqual(["David"]);
  });

  it("should filter by the selected role", () => {
    setup({ user: apiUser({}), role_id: "role_admin" }, []);
    component.users = mockUsers;

    component.selectedRole = "Docente";
    component.filterUsers();
    expect(component.filteredUsers.map((u) => u.name)).toEqual(["Bob"]);

    component.selectedRole = "Todos los roles";
    component.filterUsers();
    expect(component.filteredUsers.length).toBe(mockUsers.length);
  });

  it("should only show participants to a volunteer", () => {
    setup({ user: apiUser({ role_id: "role_volunteer" }) }, []);
    expect(component.isVolunteerOnly).toBeTrue();

    component.users = mockUsers;
    component.filterUsers();
    expect(component.filteredUsers.map((u) => u.name)).toEqual(["Eve"]);
  });
});
