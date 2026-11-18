import { describe, it, expect } from "vitest";
import { generateCvPdfBuffer, mapToResumeData } from "../src/index";
import type { InternalCvData } from "../src/map-to-resume-data";

const sampleCv: InternalCvData = {
  basics: {
    name: "Juan Pérez",
    headline: "Desarrollador Full Stack",
    email: "juan@example.com",
    phone: "+54 11 1234 5678",
    location: "Buenos Aires, Argentina",
    summary: "Profesional apasionado con experiencia en TypeScript, Angular y Node.",
  },
  work: [
    {
      company: "Empresa Buen Vivir",
      position: "Ingeniero de Software",
      startDate: "2022",
      endDate: "Presente",
      description: "Desarrollo de microservicios y aplicaciones web.",
    },
  ],
  skills: [
    { name: "TypeScript", level: "Avanzado" },
    { name: "Go", level: "Intermedio" },
  ],
  education: [
    {
      institution: "Universidad de Buenos Aires",
      degree: "Licenciatura en Informática",
      startDate: "2017",
      endDate: "2021",
    },
  ],
  metadata: {
    template: "azurill",
    primaryColor: "#485839",
  },
};

describe("cv-engine", () => {
  it("maps internal CV correctly", () => {
    const data = mapToResumeData(sampleCv);
    expect(data.basics.name).toBe("Juan Pérez");
    expect(data.metadata.template).toBe("azurill");
    expect(data.metadata.design.colors.primary).toBe("rgba(72, 88, 57, 1)");
    expect(data.metadata.typography.body.fontFamily).toBe("Poppins");
  });

  it("renders sample CV to PDF buffer", async () => {
    const buffer = await generateCvPdfBuffer(sampleCv, {
      template: "azurill",
    });
    expect(buffer).toBeDefined();
    expect(buffer.length).toBeGreaterThan(1000);
    // PDF magic bytes %PDF-
    expect(buffer.subarray(0, 5).toString("utf-8")).toBe("%PDF-");
  }, 30000);

  it("renders sample CV with picture", async () => {
    const png1px = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==";
    const buffer = await generateCvPdfBuffer({
      ...sampleCv,
      basics: {
        ...sampleCv.basics,
        picture: png1px,
      },
    }, {
      template: "onyx",
    });
    expect(buffer).toBeDefined();
    expect(buffer.length).toBeGreaterThan(1000);
    expect(buffer.subarray(0, 5).toString("utf-8")).toBe("%PDF-");
  }, 30000);

  it("renders with custom color and all 15 templates without throwing", async () => {
    const templates = [
      "azurill", "bronzor", "chikorita", "ditgar", "ditto",
      "gengar", "glalie", "kakuna", "lapras", "leafish",
      "meowth", "onyx", "pikachu", "rhyhorn", "scizor"
    ] as const;

    for (const tmpl of templates) {
      const buffer = await generateCvPdfBuffer(sampleCv, {
        template: tmpl,
        primaryColor: "#1e3a8a",
      });
      expect(buffer.subarray(0, 5).toString("utf-8")).toBe("%PDF-");
    }
  }, 120000);
});
