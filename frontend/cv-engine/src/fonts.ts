export type FontWeight = "100" | "200" | "300" | "400" | "500" | "600" | "700" | "800" | "900";

const isBrowser = globalThis.window?.document !== undefined;

interface FontMeta {
  family: string;
  category: "sans-serif" | "serif";
  weights: FontWeight[];
  files: Record<string, string>;
}

const SUPPORTED_FONTS: Record<string, FontMeta> = {
  Poppins: {
    family: "Poppins",
    category: "sans-serif",
    weights: ["400", "600", "700"],
    files: {
      "400": "Poppins-Regular.ttf",
      "600": "Poppins-SemiBold.ttf",
      "700": "Poppins-Bold.ttf",
      "400italic": "Poppins-Italic.ttf",
    },
  },
  "IBM Plex Serif": {
    family: "IBM Plex Serif",
    category: "serif",
    weights: ["400", "600", "700"],
    files: {
      "400": "IBMPlexSerif-Regular.ttf",
      "600": "IBMPlexSerif-SemiBold.ttf",
      "700": "IBMPlexSerif-Bold.ttf",
      "400italic": "IBMPlexSerif-Italic.ttf",
    },
  },
  "IBM Plex Sans": {
    family: "IBM Plex Sans",
    category: "sans-serif",
    weights: ["400", "600", "700"],
    files: {
      "400": "IBMPlexSans-Regular.ttf",
      "600": "IBMPlexSans-SemiBold.ttf",
      "700": "IBMPlexSans-Bold.ttf",
      "400italic": "IBMPlexSans-Italic.ttf",
    },
  },
};

const STANDARD_FONTS = new Set(["Helvetica", "Courier", "Times-Roman"]);

export function isStandardPdfFontFamily(family: string): boolean {
  return STANDARD_FONTS.has(family);
}

export function resolveLegacyFontAlias(family: string): string | null {
  if (family === "Arial" || family === "Calibri") return "Poppins";
  if (family === "Times New Roman" || family === "Georgia") return "IBM Plex Serif";
  return null;
}

export function getFont(family: string): FontMeta | undefined {
  if (SUPPORTED_FONTS[family]) return SUPPORTED_FONTS[family];
  const alias = resolveLegacyFontAlias(family);
  if (alias && SUPPORTED_FONTS[alias]) return SUPPORTED_FONTS[alias];

  if (family.toLowerCase().includes("serif")) {
    return SUPPORTED_FONTS["IBM Plex Serif"];
  }
  return SUPPORTED_FONTS["Poppins"];
}

function resolveFontWeight(weight: FontWeight): "400" | "600" | "700" {
  const numericWeight = Number(weight);
  if (numericWeight >= 700) {
    return "700";
  }
  if (numericWeight >= 600) {
    return "600";
  }
  return "400";
}

export function getWebFontSource(
  family: string,
  weight: FontWeight = "400",
  italic = false
): string | null {
  const font = getFont(family) ?? SUPPORTED_FONTS["Poppins"];
  if (!font) return null;

  const resolvedWeight = resolveFontWeight(weight);
  const fileKey = italic && resolvedWeight === "400" ? "400italic" : resolvedWeight;
  const filename = font.files[fileKey] ?? font.files["400"];

  if (!filename) return null;

  if (isBrowser) {
    return `/assets/fonts/${filename}`;
  }

  const proc = (globalThis as { process?: { cwd?: () => string } }).process;
  const baseDir = proc?.cwd ? proc.cwd() : ".";
  return `${baseDir}/fonts/${filename}`;
}

export function resolveBoldFontWeight(
  _family: string | string[],
  _storedWeights: readonly string[]
): FontWeight | null {
  return "700";
}

export function getPdfFallbackFontFamilies(
  _family: string,
  _options: Record<string, unknown> = {}
): string[] {
  return [];
}

export function sortFontWeights<T extends string>(fontWeights: T[]): T[] {
  return [...fontWeights].sort((a, b) => Number(a) - Number(b));
}
