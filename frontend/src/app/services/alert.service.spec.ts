import { TestBed } from "@angular/core/testing";
import { AlertService } from "./alert.service";
import { MessageService } from "primeng/api";

describe("AlertService", () => {
  let service: AlertService;
  let messageServiceSpy: jasmine.SpyObj<MessageService>;

  beforeEach(() => {
    messageServiceSpy = jasmine.createSpyObj("MessageService", ["add"]);

    TestBed.configureTestingModule({
      providers: [
        AlertService,
        { provide: MessageService, useValue: messageServiceSpy },
      ],
    });

    service = TestBed.inject(AlertService);
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should dispatch success toast via MessageService", () => {
    service.showSuccess("Operación exitosa", "Los datos fueron guardados");
    expect(messageServiceSpy.add).toHaveBeenCalledWith(
      jasmine.objectContaining({
        severity: "success",
        summary: "Operación exitosa",
        detail: "Los datos fueron guardados",
      }),
    );
  });

  it("should dispatch error toast via MessageService", () => {
    service.showError("Error al guardar", "Intenta de nuevo");
    expect(messageServiceSpy.add).toHaveBeenCalledWith(
      jasmine.objectContaining({
        severity: "error",
        summary: "Error al guardar",
        detail: "Intenta de nuevo",
      }),
    );
  });

  it("should dispatch offline toast with sticky option", () => {
    service.showOfflineToast();
    expect(messageServiceSpy.add).toHaveBeenCalledWith(
      jasmine.objectContaining({
        severity: "error",
        sticky: true,
      }),
    );
  });

  it("should open rate limit dialog with severity warn", () => {
    service.showRateLimitDialog();
    const config = service.dialogConfig();
    expect(config.visible).toBeTrue();
    expect(config.severity).toBe("warn");
    expect(config.title).toContain("Límite de solicitudes");
  });

  it("should open forbidden dialog with severity error", () => {
    service.showForbiddenDialog("Acceso no autorizado");
    const config = service.dialogConfig();
    expect(config.visible).toBeTrue();
    expect(config.severity).toBe("error");
    expect(config.message).toBe("Acceso no autorizado");
  });

  it("should open server error dialog and trigger onConfirm on confirmDialog", () => {
    const onRetrySpy = jasmine.createSpy("onRetry");
    service.showServerErrorDialog("Fallo del servidor", onRetrySpy);
    const config = service.dialogConfig();
    expect(config.visible).toBeTrue();
    expect(config.confirmLabel).toBe("Reintentar");

    service.confirmDialog();
    expect(onRetrySpy).toHaveBeenCalled();
    expect(service.dialogConfig().visible).toBeFalse();
  });

  it("should close dialog on cancelDialog", () => {
    const onCancelSpy = jasmine.createSpy("onCancel");
    service.dialogConfig.set({
      visible: true,
      title: "Test",
      message: "Test",
      severity: "info",
      confirmLabel: "OK",
      onCancel: onCancelSpy,
    });

    service.cancelDialog();
    expect(onCancelSpy).toHaveBeenCalled();
    expect(service.dialogConfig().visible).toBeFalse();
  });
});
