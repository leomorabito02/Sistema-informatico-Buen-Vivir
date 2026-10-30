import { TestBed } from "@angular/core/testing";
import { provideNoopAnimations } from "@angular/platform-browser/animations";
import { SessionExpiredDialogComponent } from "./session-expired-dialog.component";
import { SessionPromptService } from "../../../services/session-prompt.service";

describe("SessionExpiredDialogComponent + SessionPromptService", () => {
  let prompt: SessionPromptService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SessionExpiredDialogComponent],
      providers: [provideNoopAnimations()],
    }).compileComponents();
    prompt = TestBed.inject(SessionPromptService);
  });

  it("muestra el diálogo al preguntar y entrega la respuesta", () => {
    const fixture = TestBed.createComponent(SessionExpiredDialogComponent);
    fixture.detectChanges();
    expect(prompt.visible()).toBeFalse();

    let respuesta: boolean | undefined;
    prompt.ask().subscribe((r) => (respuesta = r));
    fixture.detectChanges();
    expect(prompt.visible()).toBeTrue();
    expect(document.body.textContent).toContain("Tu sesión venció");

    prompt.answer(true);
    expect(respuesta).toBeTrue();
    expect(prompt.visible()).toBeFalse();
  });

  it("varias preguntas a la vez comparten una sola respuesta", () => {
    const respuestas: boolean[] = [];
    prompt.ask().subscribe((r) => respuestas.push(r));
    prompt.ask().subscribe((r) => respuestas.push(r));

    prompt.answer(false);
    expect(respuestas).toEqual([false, false]);
  });

  it("answer() sin pregunta pendiente no falla", () => {
    expect(() => prompt.answer(true)).not.toThrow();
  });
});
