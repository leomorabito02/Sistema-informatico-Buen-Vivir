import { MAX_FOTO_BYTES, validarFotoPerfil } from "./foto-perfil";

function archivo(nombre: string, tipo: string, bytes = 10): File {
  return new File([new Uint8Array(bytes)], nombre, { type: tipo });
}

describe("validarFotoPerfil", () => {
  it("acepta JPG, PNG, WEBP y GIF", () => {
    expect(validarFotoPerfil(archivo("a.jpg", "image/jpeg"))).toBeNull();
    expect(validarFotoPerfil(archivo("a.PNG", "image/png"))).toBeNull();
    expect(validarFotoPerfil(archivo("a.webp", "image/webp"))).toBeNull();
    expect(validarFotoPerfil(archivo("a.gif", "image/gif"))).toBeNull();
  });

  it("rechaza otros tipos aunque tengan extensión de imagen", () => {
    expect(validarFotoPerfil(archivo("a.png", "text/html"))).toContain("JPG");
    expect(validarFotoPerfil(archivo("a.pdf", "application/pdf"))).toContain(
      "JPG",
    );
    expect(validarFotoPerfil(archivo("a.svg", "image/svg+xml"))).toContain(
      "JPG",
    );
  });

  it("rechaza imágenes de más de 5 MB", () => {
    const grande = archivo("a.jpg", "image/jpeg", MAX_FOTO_BYTES + 1);
    expect(validarFotoPerfil(grande)).toContain("5 MB");
  });

  it("sin archivo pide elegir uno", () => {
    expect(validarFotoPerfil(null)).toBe("Elegí una imagen.");
  });
});
