import {
  REFRESH_AT_KEY,
  REFRESH_REUSE_MS,
  conCandadoDeRefresh,
} from "./auth.service";

describe("conCandadoDeRefresh (D10: varias pestañas)", () => {
  beforeEach(() => localStorage.removeItem(REFRESH_AT_KEY));
  afterEach(() => localStorage.removeItem(REFRESH_AT_KEY));

  it("pide el refresh y anota la hora", async () => {
    const pedir = jasmine.createSpy("pedir").and.resolveTo({ ok: true });
    const res = await conCandadoDeRefresh(pedir);
    expect(pedir).toHaveBeenCalledTimes(1);
    expect(res).toEqual({ ok: true });
    expect(Number(localStorage.getItem(REFRESH_AT_KEY))).toBeGreaterThan(0);
  });

  it("si otra pestaña acaba de renovar, no vuelve a pedir", async () => {
    localStorage.setItem(REFRESH_AT_KEY, String(Date.now()));
    const pedir = jasmine.createSpy("pedir").and.resolveTo({ ok: true });
    await conCandadoDeRefresh(pedir);
    expect(pedir).not.toHaveBeenCalled();
  });

  it("si la última renovación es vieja, pide de nuevo", async () => {
    localStorage.setItem(
      REFRESH_AT_KEY,
      String(Date.now() - REFRESH_REUSE_MS - 1000),
    );
    const pedir = jasmine.createSpy("pedir").and.resolveTo({ ok: true });
    await conCandadoDeRefresh(pedir);
    expect(pedir).toHaveBeenCalledTimes(1);
  });

  it("dos renovaciones simultáneas hacen un solo pedido", async () => {
    let llamadas = 0;
    const pedir = () =>
      new Promise((r) => setTimeout(() => r({ n: ++llamadas }), 20));
    await Promise.all([conCandadoDeRefresh(pedir), conCandadoDeRefresh(pedir)]);
    expect(llamadas).toBe(1);
  });

  it("si el refresh falla no anota la hora", async () => {
    const pedir = jasmine.createSpy("pedir").and.rejectWith(new Error("401"));
    await expectAsync(conCandadoDeRefresh(pedir)).toBeRejected();
    expect(localStorage.getItem(REFRESH_AT_KEY)).toBeNull();
  });
});
