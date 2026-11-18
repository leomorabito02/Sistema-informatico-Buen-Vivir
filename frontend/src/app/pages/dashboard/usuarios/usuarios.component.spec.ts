import {
  ComponentFixture,
  TestBed,
  fakeAsync,
  tick,
} from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { Subject, of, throwError } from "rxjs";
import { ConfirmationService, MessageService } from "primeng/api";
import {
  DashboardUser,
  UsuariosComponent,
  validateNewParticipant,
} from "./usuarios.component";
import { AuthService } from "../../../services/auth.service";
import { UsersService } from "../../../services/users.service";
import { EducationService } from "../../../services/education.service";
import { RxResumeService } from "../../../services/rxresume.service";
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

function row(overrides: Partial<DashboardUser>): DashboardUser {
  return {
    id: "u",
    name: "",
    email: "",
    phone: "",
    role: "Participante",
    status: "Activo",
    coursesStatus: "",
    ...overrides,
  };
}

describe("UsuariosComponent", () => {
  let component: UsuariosComponent;
  let fixture: ComponentFixture<UsuariosComponent>;
  let authService: jasmine.SpyObj<AuthService>;
  let usersService: jasmine.SpyObj<UsersService>;
  let education: jasmine.SpyObj<EducationService>;
  let rxResumeService: jasmine.SpyObj<RxResumeService>;

  const admin: MeResponse = {
    user: apiUser({ user_id: "admin-1" }),
    role_id: "role_admin",
  };

  const mockUsers: DashboardUser[] = [
    row({
      id: "1",
      name: "Alice",
      email: "alice@example.com",
      phone: "111",
      role: "Responsable",
    }),
    row({
      id: "2",
      name: "Bob",
      email: "bob@example.com",
      phone: "222",
      role: "Docente",
    }),
    row({
      id: "3",
      name: "David",
      email: "david@example.com",
      phone: "444",
      role: "Voluntario",
    }),
    row({
      id: "4",
      name: "Eve",
      email: "eve@example.com",
      phone: "555",
      role: "Participante",
    }),
  ];

  function setup(
    me: MeResponse | null,
    users: User[] | null,
    bulkFails = false,
  ) {
    authService = jasmine.createSpyObj<AuthService>("AuthService", [
      "isAuthenticated",
    ]);
    usersService = jasmine.createSpyObj<UsersService>("UsersService", [
      "listUsers",
      "createUser",
      "getUserProfile",
      "updateUser",
      "updateRoleData",
      "deleteUser",
      "getParticipantDocuments",
      "uploadParticipantDocument",
      "updateParticipantDocument",
      "deleteParticipantDocument",
      "downloadParticipantDocument",
    ]);
    usersService.getParticipantDocuments.and.returnValue(of([]));
    usersService.uploadParticipantDocument.and.returnValue(
      of({ id: "pdoc-1", name: "doc1.pdf" }),
    );
    usersService.updateParticipantDocument.and.returnValue(
      of({ id: "pdoc-1", name: "doc1_edit.pdf" }),
    );
    usersService.deleteParticipantDocument.and.returnValue(
      of({ success: true }),
    );
    education = jasmine.createSpyObj<EducationService>("EducationService", [
      "getBulkUserCourses",
      "getParticipantEnrollments",
      "getTeacherCourses",
    ]);
    rxResumeService = jasmine.createSpyObj<RxResumeService>("RxResumeService", [
      "getUserResumeByAdmin",
      "uploadUserPictureByAdmin",
    ]);
    rxResumeService.getUserResumeByAdmin.and.returnValue(
      of({ data: { basics: { picture: "" } } } as any),
    );
    rxResumeService.uploadUserPictureByAdmin.and.returnValue(
      of({ data: { basics: { picture: "" } } } as any),
    );
    authService.isAuthenticated.and.returnValue(
      me ? of(me) : throwError(() => new Error("sin sesión")),
    );
    usersService.listUsers.and.returnValue(
      users ? of(users) : throwError(() => new Error("auth caído")),
    );
    education.getBulkUserCourses.and.returnValue(
      bulkFails
        ? throwError(() => new Error("education caído"))
        : of([{ user_id: "u1", courses: ["Huerta", "Carpintería"] }]),
    );
    education.getParticipantEnrollments.and.returnValue(of([]));
    education.getTeacherCourses.and.returnValue(of([]));
    spyOn(console, "error");

    TestBed.configureTestingModule({
      imports: [UsuariosComponent, NoopAnimationsModule],
      providers: [
        { provide: AuthService, useValue: authService },
        { provide: UsersService, useValue: usersService },
        { provide: EducationService, useValue: education },
        { provide: RxResumeService, useValue: rxResumeService },
      ],
    });

    fixture = TestBed.createComponent(UsuariosComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  /** Espía de los toasts (MessageService propio del componente). */
  function spyToasts(): jasmine.Spy {
    return spyOn(fixture.debugElement.injector.get(MessageService), "add");
  }

  /** Acepta en el acto la confirmación que pida el componente. */
  function autoConfirm(): jasmine.Spy {
    const confirm = fixture.debugElement.injector.get(ConfirmationService);
    return spyOn(confirm, "confirm").and.callFake((c: any) => {
      c.accept();
      return confirm;
    });
  }

  it("maps API users to rows and loads their courses in one batch", () => {
    setup(admin, [
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
        status: "blocked",
      }),
      apiUser({ user_id: "u3", first_name: "Vol", role_id: "role_volunteer" }),
      apiUser({ user_id: "u4", first_name: "Emp", role_id: "role_company" }),
      apiUser({ user_id: "u5", first_name: "Par", role_id: "" }),
    ]);

    expect(component.currentUserId).toBe("admin-1");
    expect(component.isVolunteerOnly).toBeFalse();
    expect(education.getBulkUserCourses).toHaveBeenCalledWith([
      "u1",
      "u2",
      "u3",
      "u4",
      "u5",
    ]);
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
        phone: "123",
        status: "Activo",
        coursesStatus: "Huerta, Carpintería",
      }),
    );
    expect(component.users[1].name).toBe("solo@mail.com");
    expect(component.users[1].status).toBe("Bloqueado");
    expect(component.users[1].coursesStatus).toBe("Sin cursos");
    expect(component.users[2].phone).toBe("-");
    expect(component.filteredUsers).toHaveSize(5);
  });

  it("maps participant enrolled courses directly by user_id alongside teacher courses", () => {
    setup(admin, []);
    usersService.listUsers.and.returnValue(
      of([
        apiUser({
          user_id: "u-part",
          first_name: "Lucía",
          last_name: "Gómez",
          role_id: "role_participant",
        }),
        apiUser({
          user_id: "u-teach",
          first_name: "Carlos",
          last_name: "Ruiz",
          role_id: "role_teacher",
        }),
      ]),
    );
    education.getBulkUserCourses.and.returnValue(
      of([
        { user_id: "u-part", courses: ["Panadería", "Costura"] },
        { user_id: "u-teach", courses: ["Huerta"] },
      ]),
    );

    component.reload();

    expect(education.getBulkUserCourses).toHaveBeenCalledWith([
      "u-part",
      "u-teach",
    ]);
    expect(component.users[0].coursesStatus).toBe("Panadería, Costura");
    expect(component.users[1].coursesStatus).toBe("Huerta");
  });

  it("shows a table-styled skeleton while loading users and their courses", () => {
    setup(admin, []);
    const pendingUsers$ = new Subject<User[]>();
    const pendingCourses$ = new Subject<
      { user_id: string; courses: string[] }[]
    >();
    usersService.listUsers.and.returnValue(pendingUsers$);
    education.getBulkUserCourses.and.returnValue(pendingCourses$);

    component.reload();
    fixture.detectChanges();

    expect(component.loading).toBeTrue();
    expect(
      fixture.nativeElement.querySelectorAll(".table-skeleton tbody tr"),
    ).toHaveSize(6);

    pendingUsers$.next([apiUser({ user_id: "u1", first_name: "Ana" })]);
    pendingUsers$.complete();
    fixture.detectChanges();

    expect(component.loading).toBeTrue();
    expect(
      fixture.nativeElement.querySelector(".table-skeleton"),
    ).not.toBeNull();

    pendingCourses$.next([{ user_id: "u1", courses: ["Huerta"] }]);
    pendingCourses$.complete();
    fixture.detectChanges();

    expect(component.loading).toBeFalse();
    expect(fixture.nativeElement.querySelector(".table-skeleton")).toBeNull();
  });

  it("keeps the list when the courses batch fails", () => {
    setup(admin, [apiUser({ user_id: "u1" })], true);
    expect(component.filteredUsers).toHaveSize(1);
    expect(component.filteredUsers[0].coursesStatus).toBe("—");
  });

  it("shows an empty list when the API returns nothing", () => {
    setup(admin, []);
    expect(component.users).toEqual([]);
    expect(education.getBulkUserCourses).not.toHaveBeenCalled();
  });

  it("shows the error state when the API fails, and retries", () => {
    setup(admin, null);
    expect(component.users).toEqual([]);
    expect(component.loading).toBeFalse();
    expect(component.loadError).toBe("auth caído");
    expect(
      fixture.nativeElement.querySelector("app-error-state"),
    ).not.toBeNull();

    usersService.listUsers.and.returnValue(of([apiUser({ user_id: "u1" })]));
    component.reload();
    expect(component.loadError).toBe("");
    expect(component.users).toHaveSize(1);
  });

  it("still loads users when the session check fails", () => {
    setup(null, [apiUser({ user_id: "u1", first_name: "Ana" })]);
    expect(usersService.listUsers).toHaveBeenCalled();
    expect(component.users).toHaveSize(1);
  });

  it("asks the backend for the search text after typing stops", fakeAsync(() => {
    setup(admin, []);
    usersService.listUsers.calls.reset();

    component.searchTerm = "ali";
    component.buscar();
    component.searchTerm = "alice";
    component.buscar();
    tick(299);
    expect(usersService.listUsers).not.toHaveBeenCalled();

    tick(1);
    expect(usersService.listUsers).toHaveBeenCalledOnceWith({ q: "alice" });
  }));

  it("filters by the selected role", () => {
    setup(admin, []);
    component.users = mockUsers;

    component.selectedRole = "Docente";
    component.filterUsers();
    expect(component.filteredUsers.map((u) => u.name)).toEqual(["Bob"]);
  });

  it("only shows participants to a volunteer", () => {
    setup({ user: apiUser({ role_id: "role_volunteer" }) }, []);
    expect(component.isVolunteerOnly).toBeTrue();

    component.users = mockUsers;
    component.filterUsers();
    expect(component.filteredUsers.map((u) => u.name)).toEqual(["Eve"]);
  });

  it("detects pending role approvals", () => {
    setup(admin, []);
    expect(component.isRolePending("pending")).toBeTrue();
    expect(component.isRolePending("Pendiente")).toBeTrue();
    expect(component.isRolePending("approved")).toBeFalse();
    expect(component.isRolePending(undefined)).toBeFalse();
  });

  it("offers the admin role only to admins when creating a user", () => {
    setup(admin, []);
    component.openNewUserModal();
    expect(component.showNewUserModal).toBeTrue();
    expect(component.availableNewRoles).toContain("role_admin");

    component.isVolunteerOnly = true;
    component.openNewUserModal();
    expect(component.availableNewRoles).not.toContain("role_admin");
  });

  it("creates a user and reloads the list", () => {
    setup(admin, []);
    const toasts = spyToasts();
    usersService.createUser.and.returnValue(of({}));

    component.createUser();
    expect(usersService.createUser).not.toHaveBeenCalled(); // faltan datos
    expect(toasts).toHaveBeenCalledWith(
      jasmine.objectContaining({
        severity: "warn",
        detail: "Completar todos los campos obligatorios para continuar",
      }),
    );

    // HU-07 y HU-01: participante sin Gmail, DNI ni dirección no se crea.
    component.newUserForm = {
      first_name: "N",
      last_name: "U",
      email: "n@u.com",
      phone: "1",
      password: "Clave123!",
      role_id: "role_participant",
    };
    component.createUser();
    expect(usersService.createUser).not.toHaveBeenCalled();

    component.newUserForm = {
      ...component.newUserForm,
      email: "n@gmail.com",
      dni: "30111222",
      address: "San Martín 100",
    };
    component.showNewUserModal = true;
    component.createUser();

    expect(usersService.createUser).toHaveBeenCalledWith(
      jasmine.objectContaining({
        email: "n@gmail.com",
        role_id: "role_participant",
        dni: "30111222",
        address: "San Martín 100",
      }),
    );
    expect(component.showNewUserModal).toBeFalse();
    expect(usersService.listUsers).toHaveBeenCalledTimes(2);
    expect(toasts).toHaveBeenCalledWith(
      jasmine.objectContaining({
        severity: "success",
        detail: "Usuario creado correctamente",
      }),
    );
  });

  it("shows the backend message when creating a user fails", () => {
    setup(admin, []);
    const toasts = spyToasts();
    usersService.createUser.and.returnValue(
      throwError(() => new Error("ya existe una cuenta")),
    );
    component.newUserForm = {
      email: "n@u.com",
      password: "x",
      role_id: "role_admin",
    };

    component.createUser();

    expect(toasts).toHaveBeenCalledWith(
      jasmine.objectContaining({
        severity: "error",
        detail: "ya existe una cuenta",
      }),
    );
  });

  it("opens the detail with profile and enrollments of a participant", () => {
    setup(admin, []);
    usersService.getUserProfile.and.returnValue(
      of({ user: {}, role_data: { participant_id: "part-1" } }),
    );
    education.getParticipantEnrollments.and.returnValue(
      of([
        {
          id: "e1",
          cursoId: "c1",
          courseTitle: "Panadería",
          estado: "activa",
        } as never,
        { id: "e2", cursoId: "c2", estado: "baja" } as never,
      ]),
    );
    const user = row({
      id: "u1",
      originalUser: apiUser({ user_id: "u1", role_id: "role_participant" }),
    });

    component.openUserDetail(user);

    expect(component.showDetailModal).toBeTrue();
    expect(user.profileData.role_data.participant_id).toBe("part-1");
    expect(education.getParticipantEnrollments).toHaveBeenCalledWith("part-1");
    expect(user.coursesList).toEqual([
      jasmine.objectContaining({
        id: "e1",
        cursoId: "c1",
        courseTitle: "Panadería",
      }),
    ]);
  });

  it("loads the courses of a teacher in the detail", () => {
    setup(admin, []);
    usersService.getUserProfile.and.returnValue(
      throwError(() => new Error("x")),
    );
    education.getTeacherCourses.and.returnValue(of([{ id: "c1" } as never]));
    const user = row({
      id: "t1",
      originalUser: apiUser({ user_id: "t1", role_id: "role_teacher" }),
    });

    component.openUserDetail(user);

    expect(component.showDetailModal).toBeTrue();
    expect(education.getTeacherCourses).toHaveBeenCalledWith("t1");
    expect(user.coursesList?.length).toBe(1);
  });

  it("does not let a volunteer edit non-participants", () => {
    setup({ user: apiUser({ role_id: "role_volunteer" }) }, []);
    const toasts = spyToasts();
    const teacher = row({
      originalUser: apiUser({ user_id: "t1", role_id: "role_teacher" }),
    });

    component.openEditUser(teacher, new Event("click"));

    expect(toasts).toHaveBeenCalledWith(
      jasmine.objectContaining({ severity: "info" }),
    );
    expect(component.showEditModal).toBeFalse();
  });

  it("loads the edit form normalizing the legacy 'accepted' status", () => {
    setup(admin, []);
    usersService.getUserProfile.and.returnValue(
      of({
        user: {},
        role_data: {
          participant_id: "part-1",
          approval_status: "accepted",
          birthdate: "2000-01-01T00:00:00Z",
        },
      }),
    );
    const user = row({
      originalUser: apiUser({
        user_id: "u1",
        first_name: "Ana",
        role_id: "role_participant",
      }),
    });

    component.openEditUser(user, new Event("click"));

    expect(component.showEditModal).toBeTrue();
    expect(component.editForm).toEqual(
      jasmine.objectContaining({
        first_name: "Ana",
        approval_status: "approved",
        birthdate: "2000-01-01",
        participant_id: "part-1",
      }),
    );
  });

  it("saves the user and its role data, recording who approved it", () => {
    setup(admin, []);
    usersService.updateUser.and.returnValue(of({}));
    usersService.updateRoleData.and.returnValue(of({}));
    component.selectedUser = row({
      originalUser: apiUser({ user_id: "u1", role_id: "role_participant" }),
    });
    component.editForm = {
      first_name: "Ana",
      last_name: "P",
      phone: "1",
      status: "active",
      participant_id: "part-1",
      approval_status: "approved",
      birthdate: "2000-01-01",
    };
    component.showEditModal = true;

    component.saveEdit();

    expect(usersService.updateUser).toHaveBeenCalledWith("u1", {
      first_name: "Ana",
      last_name: "P",
      phone: "1",
      status: "active",
    });
    const [table, id, payload] =
      usersService.updateRoleData.calls.mostRecent().args;
    expect(table).toBe("participants");
    expect(id).toBe("part-1");
    expect(payload).toEqual(
      jasmine.objectContaining({
        approval_status: "approved",
        approved_by: "admin-1",
      }),
    );
    expect(payload.first_name).toBeUndefined();
    expect(component.showEditModal).toBeFalse();
  });

  it("skips the role update for admins and for users without a role profile", () => {
    setup(admin, []);
    usersService.updateUser.and.returnValue(of({}));

    component.selectedUser = row({
      originalUser: apiUser({ user_id: "a1", role_id: "role_admin" }),
    });
    component.editForm = { first_name: "A" };
    component.saveEdit();

    component.selectedUser = row({
      originalUser: apiUser({ user_id: "v1", role_id: "role_volunteer" }),
    });
    component.editForm = { first_name: "V" }; // sin volunteer_id: no existe el perfil
    component.saveEdit();

    expect(usersService.updateUser).toHaveBeenCalledTimes(2);
    expect(usersService.updateRoleData).not.toHaveBeenCalled();
  });

  it("shows an error toast when saving fails", () => {
    setup(admin, []);
    const toasts = spyToasts();
    usersService.updateUser.and.returnValue(throwError(() => new Error("403")));
    component.selectedUser = row({ originalUser: apiUser({ user_id: "u1" }) });

    component.saveEdit();

    expect(toasts).toHaveBeenCalledWith(
      jasmine.objectContaining({ severity: "error", detail: "403" }),
    );
  });

  describe("eliminar usuario", () => {
    const ana = () =>
      row({ name: "Ana", originalUser: apiUser({ user_id: "u9" }) });

    it("pide confirmación, borra, avisa y recarga la lista", () => {
      setup(admin, []);
      const toasts = spyToasts();
      const confirm = autoConfirm();
      usersService.deleteUser.and.returnValue(of(undefined));

      component.deleteUser(ana(), new Event("click"));

      expect(confirm).toHaveBeenCalledWith(
        jasmine.objectContaining({ header: "Eliminar usuario" }),
      );
      expect(usersService.deleteUser).toHaveBeenCalledWith("u9");
      expect(toasts).toHaveBeenCalledWith(
        jasmine.objectContaining({ severity: "success" }),
      );
      expect(usersService.listUsers).toHaveBeenCalledTimes(2);
    });

    it("no borra si se cancela la confirmación", () => {
      setup(admin, []);
      const confirm = fixture.debugElement.injector.get(ConfirmationService);
      spyOn(confirm, "confirm").and.returnValue(confirm);

      component.deleteUser(ana(), new Event("click"));

      expect(usersService.deleteUser).not.toHaveBeenCalled();
    });

    it("muestra el error del backend si no se pudo borrar", () => {
      setup(admin, []);
      const toasts = spyToasts();
      autoConfirm();
      usersService.deleteUser.and.returnValue(
        throwError(() => new Error("forbidden")),
      );

      component.deleteUser(ana(), new Event("click"));

      expect(toasts).toHaveBeenCalledWith(
        jasmine.objectContaining({ severity: "error", detail: "forbidden" }),
      );
      expect(usersService.listUsers).toHaveBeenCalledTimes(1);
    });

    it("ignora filas sin id de usuario", () => {
      setup(admin, []);
      const confirm = autoConfirm();
      component.deleteUser(row({}), new Event("click"));
      expect(confirm).not.toHaveBeenCalled();
    });
  });

  it("checks canEditPicture correctly for staff and participants", () => {
    setup(admin, []);
    component.currentUserRoleId = "role_admin";
    component.selectedUser = row({
      role: "Participante",
      originalUser: apiUser({ user_id: "p1", role_id: "role_participant" }),
    });
    expect(component.canEditPicture()).toBeTrue();

    component.selectedUser = row({
      role: "Docente",
      originalUser: apiUser({ user_id: "t1", role_id: "role_teacher" }),
    });
    expect(component.canEditPicture()).toBeFalse();
  });

  it("uploads and synchronizes participant detail picture", () => {
    setup(admin, []);
    usersService.updateRoleData.and.returnValue(of({}));
    const fakeResume = {
      data: { basics: { picture: "https://gcp.bucket/pictures/p1/pic.jpg" } },
    } as any;
    rxResumeService.uploadUserPictureByAdmin.and.returnValue(of(fakeResume));

    const participantRow = row({
      id: "p1",
      role: "Participante",
      originalUser: apiUser({ user_id: "p1", role_id: "role_participant" }),
      profileData: { role_data: { participant_id: "part-1", image_url: "" } },
    });
    component.users = [participantRow];
    component.selectedUser = participantRow;

    const file = new File(["dummy"], "photo.png", { type: "image/png" });
    const event = { target: { files: [file] } };

    component.onDetailPictureSelected(event);

    expect(rxResumeService.uploadUserPictureByAdmin).toHaveBeenCalledWith(
      "p1",
      file,
    );
    expect(component.selectedUser.image_url).toBe(
      "https://gcp.bucket/pictures/p1/pic.jpg",
    );
    expect(usersService.updateRoleData).toHaveBeenCalledWith(
      "participants",
      "part-1",
      {
        image_url: "https://gcp.bucket/pictures/p1/pic.jpg",
      },
    );
  });

  it("uploads a document for a participant", () => {
    setup(admin, []);
    const participantRow = row({
      id: "p1",
      role: "Participante",
      originalUser: apiUser({ user_id: "p1", role_id: "role_participant" }),
      profileData: { role_data: { participant_id: "part-1" } },
    });
    component.selectedUser = participantRow;
    component.participantDocuments = [];

    const file = new File(["dummy"], "dni.pdf", { type: "application/pdf" });
    const event = { target: { files: [file], value: "C:\\fakepath\\dni.pdf" } };

    component.onDocumentSelected(event);

    expect(usersService.uploadParticipantDocument).toHaveBeenCalledWith(
      "part-1",
      file,
      "dni.pdf",
    );
    expect(component.participantDocuments).toHaveSize(1);
    expect(component.participantDocuments[0].id).toBe("pdoc-1");
  });

  it("manages document edit state and updates document", () => {
    setup(admin, []);
    const participantRow = row({
      id: "p1",
      role: "Participante",
      originalUser: apiUser({ user_id: "p1", role_id: "role_participant" }),
      profileData: { role_data: { participant_id: "part-1" } },
    });
    component.selectedUser = participantRow;
    component.participantDocuments = [{ id: "pdoc-1", name: "doc1.pdf" }];

    component.startEditDocument(component.participantDocuments[0]);
    expect(component.editingDocId).toBe("pdoc-1");
    expect(component.editingDocName).toBe("doc1.pdf");

    component.editingDocName = "doc1_edit.pdf";
    const replaceFile = new File(["updated"], "doc1_new.pdf", {
      type: "application/pdf",
    });
    component.onEditDocFileSelected({ target: { files: [replaceFile] } });

    component.saveEditDocument("pdoc-1");

    expect(usersService.updateParticipantDocument).toHaveBeenCalledWith(
      "part-1",
      "pdoc-1",
      replaceFile,
      "doc1_edit.pdf",
    );
    expect(component.participantDocuments[0].name).toBe("doc1_edit.pdf");
    expect(component.editingDocId).toBeNull();
  });

  it("deletes a document for a participant", () => {
    setup(admin, []);
    spyOn(window, "confirm").and.returnValue(true);
    const participantRow = row({
      id: "p1",
      role: "Participante",
      originalUser: apiUser({ user_id: "p1", role_id: "role_participant" }),
      profileData: { role_data: { participant_id: "part-1" } },
    });
    component.selectedUser = participantRow;
    component.participantDocuments = [
      { id: "pdoc-1", name: "doc1.pdf" },
      { id: "pdoc-2", name: "doc2.pdf" },
    ];

    component.deleteDocument("pdoc-1");

    expect(usersService.deleteParticipantDocument).toHaveBeenCalledWith(
      "part-1",
      "pdoc-1",
    );
    expect(component.participantDocuments).toHaveSize(1);
    expect(component.participantDocuments[0].id).toBe("pdoc-2");
  });
});

describe("validateNewParticipant (HU-07, HU-01)", () => {
  const ok = {
    role_id: "role_participant",
    email: "ana@gmail.com",
    dni: "30111222",
    address: "Calle 1",
    phone: "351",
  };
  it("acepta un participante completo y no valida otros roles", () => {
    expect(validateNewParticipant(ok)).toBe("");
    expect(
      validateNewParticipant({
        role_id: "role_company",
        email: "x@empresa.com",
      }),
    ).toBe("");
  });
  it("pide Gmail, DNI de 7 u 8 números y dirección", () => {
    expect(validateNewParticipant({ ...ok, email: "a@hotmail.com" })).toContain(
      "Gmail",
    );
    expect(validateNewParticipant({ ...ok, dni: "30.111.222" })).toContain(
      "DNI",
    );
    expect(validateNewParticipant({ ...ok, address: "" })).toContain(
      "dirección",
    );
  });
});

describe("UsuariosComponent.approvalDecision (HU-12)", () => {
  it("solo decide cuando la cuenta estaba pendiente", () => {
    const c = Object.create(UsuariosComponent.prototype) as UsuariosComponent;
    c.originalApprovalStatus = "pending";
    c.editForm = { approval_status: "rejected" };
    expect(c.approvalDecision()).toBe("rejected");
    c.editForm = { approval_status: "approved" };
    expect(c.approvalDecision()).toBe("approved");
    c.originalApprovalStatus = "approved";
    expect(c.approvalDecision()).toBeNull();
  });
});

describe("UsuariosComponent: aprobar desde la lista (HU-12, HU-38)", () => {
  const pendiente = (role: string, role_id: string) =>
    ({
      id: "u1",
      name: "Ana Paz",
      email: "a@gmail.com",
      phone: "",
      role,
      status: "Activo",
      roleApprovalStatus: "pending",
      coursesStatus: "",
      originalUser: { user_id: "u1", role_id },
    }) as unknown as DashboardUser;

  function crear(voluntario = false) {
    const c = Object.create(UsuariosComponent.prototype) as UsuariosComponent;
    c.users = [
      pendiente("Voluntario", "role_volunteer"),
      pendiente("Participante", "role_participant"),
    ];
    c.isVolunteerOnly = voluntario;
    c.selectedRole = "Todos los roles";
    return c;
  }

  it("el admin ve todas las pendientes; el voluntario solo participantes", () => {
    expect(crear().pendientes).toHaveSize(2);
    expect(crear(true).pendientes).toHaveSize(1);
  });

  it("el filtro deja solo las pendientes", () => {
    const c = crear();
    c.users.push({
      ...pendiente("Docente", "role_teacher"),
      roleApprovalStatus: "approved",
    });
    c.togglePendientes();
    expect(c.filteredUsers).toHaveSize(2);
    c.togglePendientes();
    expect(c.filteredUsers).toHaveSize(3);
  });

  it("aprobar pide confirmación y llama a setApproval", () => {
    const c = crear();
    const usersService = jasmine.createSpyObj("UsersService", ["setApproval"]);
    usersService.setApproval.and.returnValue(of({}));
    const recarga = spyOn(c as any, "loadUsers");
    const confirmation = { confirm: (o: { accept: () => void }) => o.accept() };
    const messages = jasmine.createSpyObj("MessageService", ["add"]);
    Object.assign(c, { usersService, confirmation, messages });
    c.decidirAprobacion(c.users[0], "approved", new Event("click"));
    expect(usersService.setApproval).toHaveBeenCalledWith("u1", "approved");
    expect(messages.add).toHaveBeenCalledWith(
      jasmine.objectContaining({ summary: "Cuenta aprobada" }),
    );
    expect(recarga).toHaveBeenCalled();
  });
});
