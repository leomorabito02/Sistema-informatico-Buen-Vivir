import { inject, Injectable, signal } from "@angular/core";
import { MessageService } from "primeng/api";

export interface GlobalDialogConfig {
  visible: boolean;
  title: string;
  message: string;
  severity: "warn" | "error" | "info";
  confirmLabel: string;
  cancelLabel?: string;
  onConfirm?: () => void;
  onCancel?: () => void;
}

const DEFAULT_DIALOG_CONFIG: GlobalDialogConfig = {
  visible: false,
  title: "",
  message: "",
  severity: "info",
  confirmLabel: "Entendido",
};

@Injectable({
  providedIn: "root",
})
export class AlertService {
  private readonly messageService = inject(MessageService);

  readonly dialogConfig = signal<GlobalDialogConfig>({
    ...DEFAULT_DIALOG_CONFIG,
  });

  showSuccess(summary: string, detail?: string): void {
    this.messageService.add({
      severity: "success",
      summary,
      detail: detail || "",
      life: 4000,
    });
  }

  showError(summary: string, detail?: string): void {
    this.messageService.add({
      severity: "error",
      summary,
      detail: detail || "",
      life: 5000,
    });
  }

  showWarn(summary: string, detail?: string): void {
    this.messageService.add({
      severity: "warn",
      summary,
      detail: detail || "",
      life: 4500,
    });
  }

  showInfo(summary: string, detail?: string): void {
    this.messageService.add({
      severity: "info",
      summary,
      detail: detail || "",
      life: 4000,
    });
  }

  showOfflineToast(): void {
    this.messageService.add({
      severity: "error",
      summary: "Sin conexión con el servidor",
      detail:
        "No se puede establecer comunicación con el backend. Por favor, verifica tu conexión a internet.",
      sticky: true,
    });
  }

  showRateLimitDialog(): void {
    this.dialogConfig.set({
      visible: true,
      title: "Límite de solicitudes superado",
      message:
        "Has realizado demasiadas solicitudes en un corto periodo de tiempo. Por favor, aguarda unos instantes antes de volver a intentar.",
      severity: "warn",
      confirmLabel: "Entendido",
    });
  }

  showForbiddenDialog(message?: string): void {
    this.dialogConfig.set({
      visible: true,
      title: "Acceso denegado",
      message:
        message ||
        "No cuentas con los permisos necesarios para realizar esta acción en el sistema.",
      severity: "error",
      confirmLabel: "Entendido",
    });
  }

  showServerErrorDialog(message?: string, onRetry?: () => void): void {
    this.dialogConfig.set({
      visible: true,
      title: "Error inesperado del servidor",
      message:
        message ||
        "Ocurrió un problema temporal al procesar tu solicitud. Intenta nuevamente en unos instantes.",
      severity: "error",
      confirmLabel: onRetry ? "Reintentar" : "Cerrar",
      cancelLabel: onRetry ? "Cancelar" : undefined,
      onConfirm: onRetry,
    });
  }

  closeDialog(): void {
    const current = this.dialogConfig();
    this.dialogConfig.set({ ...current, visible: false });
  }

  confirmDialog(): void {
    const current = this.dialogConfig();
    if (current.onConfirm) {
      current.onConfirm();
    }
    this.closeDialog();
  }

  cancelDialog(): void {
    const current = this.dialogConfig();
    if (current.onCancel) {
      current.onCancel();
    }
    this.closeDialog();
  }
}
