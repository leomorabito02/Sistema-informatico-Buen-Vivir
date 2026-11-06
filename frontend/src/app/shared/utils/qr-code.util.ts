/**
 * Lightweight, self-contained QR Code SVG generator in pure TypeScript.
 * Supports byte mode (UTF-8 / ASCII) up to Version 6 (L/M error correction),
 * suitable for otpauth:// URIs (~60-90 characters).
 */

const GF256_EXP = new Uint8Array(512);
const GF256_LOG = new Uint8Array(256);

(() => {
  let x = 1;
  for (let i = 0; i < 255; i++) {
    GF256_EXP[i] = x;
    GF256_EXP[i + 255] = x;
    GF256_LOG[x] = i;
    x <<= 1;
    if (x & 256) x ^= 0x11d;
  }
})();

function gfMul(x: number, y: number): number {
  if (x === 0 || y === 0) return 0;
  return GF256_EXP[GF256_LOG[x] + GF256_LOG[y]];
}

function rsCompute(data: Uint8Array, eccLen: number): Uint8Array {
  let gen = new Uint8Array([1]);
  for (let i = 0; i < eccLen; i++) {
    const next = new Uint8Array(gen.length + 1);
    const root = GF256_EXP[i];
    for (let j = 0; j < gen.length; j++) {
      next[j] ^= gfMul(gen[j], root);
      next[j + 1] ^= gen[j];
    }
    gen = next;
  }

  const res = new Uint8Array(eccLen);
  for (const byte of data) {
    const factor = byte ^ res[0];
    res.copyWithin(0, 1);
    res[eccLen - 1] = 0;
    for (let j = 0; j < eccLen; j++) {
      res[j] ^= gfMul(gen[j], factor);
    }
  }
  return res;
}

interface VersionConfig {
  version: number;
  totalCodewords: number;
  dataCodewords: number;
  eccCodewords: number;
  alignments: number[];
}

const VERSIONS: VersionConfig[] = [
  { version: 1, totalCodewords: 26, dataCodewords: 16, eccCodewords: 10, alignments: [] },
  { version: 2, totalCodewords: 44, dataCodewords: 28, eccCodewords: 16, alignments: [6, 18] },
  { version: 3, totalCodewords: 70, dataCodewords: 44, eccCodewords: 26, alignments: [6, 22] },
  { version: 4, totalCodewords: 100, dataCodewords: 64, eccCodewords: 36, alignments: [6, 26] },
  { version: 5, totalCodewords: 134, dataCodewords: 86, eccCodewords: 48, alignments: [6, 30] },
  { version: 6, totalCodewords: 172, dataCodewords: 108, eccCodewords: 64, alignments: [6, 34] },
];

function selectVersion(textLen: number): VersionConfig {
  for (const v of VERSIONS) {
    if (v.dataCodewords >= textLen + 3) {
      return v;
    }
  }
  return VERSIONS.at(-1)!;
}

function buildBitStream(enc: Uint8Array, dataCodewordsCount: number): number[] {
  const bits: number[] = [];
  const pushBits = (val: number, len: number) => {
    for (let i = len - 1; i >= 0; i--) {
      bits.push((val >> i) & 1);
    }
  };

  pushBits(0b0100, 4);
  pushBits(enc.length, 8);
  for (const byte of enc) {
    pushBits(byte, 8);
  }

  const maxBits = dataCodewordsCount * 8;
  const termLen = Math.min(4, maxBits - bits.length);
  pushBits(0, termLen);

  while (bits.length % 8 !== 0) {
    bits.push(0);
  }

  const padBytes = [0xec, 0x11];
  let padIdx = 0;
  while (bits.length < maxBits) {
    pushBits(padBytes[padIdx % 2], 8);
    padIdx++;
  }

  return bits;
}

function createCodewords(bits: number[], chosen: VersionConfig): Uint8Array {
  const dataCodewords = new Uint8Array(chosen.dataCodewords);
  for (let i = 0; i < chosen.dataCodewords; i++) {
    let b = 0;
    for (let j = 0; j < 8; j++) {
      b = (b << 1) | bits[i * 8 + j];
    }
    dataCodewords[i] = b;
  }

  const ecc = rsCompute(dataCodewords, chosen.eccCodewords);
  const fullCodewords = new Uint8Array(chosen.totalCodewords);
  fullCodewords.set(dataCodewords, 0);
  fullCodewords.set(ecc, chosen.dataCodewords);
  return fullCodewords;
}

function createEmptyMatrix(dim: number): (boolean | null)[][] {
  return Array.from({ length: dim }, () => new Array(dim).fill(null));
}

function drawFinders(matrix: (boolean | null)[][], dim: number): void {
  const draw = (r: number, c: number) => {
    for (let dr = -1; dr <= 7; dr++) {
      for (let dc = -1; dc <= 7; dc++) {
        const nr = r + dr;
        const nc = c + dc;
        if (nr >= 0 && nr < dim && nc >= 0 && nc < dim) {
          const inFinder = dr >= 0 && dr <= 6 && dc >= 0 && dc <= 6;
          matrix[nr][nc] = inFinder
            ? dr === 0 || dr === 6 || dc === 0 || dc === 6 || (dr >= 2 && dr <= 4 && dc >= 2 && dc <= 4)
            : false;
        }
      }
    }
  };

  draw(0, 0);
  draw(0, dim - 7);
  draw(dim - 7, 0);
}

function drawAlignments(matrix: (boolean | null)[][], coords: number[]): void {
  for (const r of coords) {
    for (const c of coords) {
      if (matrix[r][c] !== null) continue;
      for (let dr = -2; dr <= 2; dr++) {
        for (let dc = -2; dc <= 2; dc++) {
          matrix[r + dr][c + dc] = Math.max(Math.abs(dr), Math.abs(dc)) !== 1;
        }
      }
    }
  }
}

function drawPatternsAndStructure(matrix: (boolean | null)[][], chosen: VersionConfig, dim: number): void {
  drawFinders(matrix, dim);

  for (let i = 8; i < dim - 8; i++) {
    matrix[6][i] ??= i % 2 === 0;
    matrix[i][6] ??= i % 2 === 0;
  }

  if (chosen.alignments.length === 2) {
    drawAlignments(matrix, chosen.alignments);
  }

  matrix[dim - 8][8] = true;

  for (let i = 0; i < 9; i++) {
    matrix[8][i] ??= false;
    matrix[i][8] ??= false;
  }
  for (let i = 0; i < 8; i++) {
    matrix[8][dim - 1 - i] ??= false;
    matrix[dim - 1 - i][8] ??= false;
  }
}

function codewordsToBits(fullCodewords: Uint8Array): number[] {
  const allBits: number[] = [];
  for (const byte of fullCodewords) {
    for (let i = 7; i >= 0; i--) {
      allBits.push((byte >> i) & 1);
    }
  }
  return allBits;
}

function placeColumnPair(
  matrix: (boolean | null)[][],
  cols: number[],
  rows: number[],
  allBits: number[],
  state: { bitIdx: number }
): void {
  for (const r of rows) {
    for (const c of cols) {
      if (matrix[r][c] === null) {
        const bit = state.bitIdx < allBits.length ? allBits[state.bitIdx++] : 0;
        const masked = (r + c) % 2 === 0 ? bit ^ 1 : bit;
        matrix[r][c] = masked === 1;
      }
    }
  }
}

function placeDataBits(matrix: (boolean | null)[][], fullCodewords: Uint8Array, dim: number): void {
  const allBits = codewordsToBits(fullCodewords);
  const state = { bitIdx: 0 };
  let upwards = true;

  for (let right = dim - 1; right > 0; right -= 2) {
    if (right === 6) right--;
    const cols = [right, right - 1];
    const rows = upwards
      ? Array.from({ length: dim }, (_, i) => dim - 1 - i)
      : Array.from({ length: dim }, (_, i) => i);

    placeColumnPair(matrix, cols, rows, allBits, state);
    upwards = !upwards;
  }
}

function applyFormatInfo(matrix: (boolean | null)[][], dim: number): void {
  const formatInfo = 0b101010000010010;
  for (let i = 0; i < 15; i++) {
    const bit = ((formatInfo >> i) & 1) === 1;
    if (i < 6) matrix[8][i] = bit;
    else if (i === 6) matrix[8][7] = bit;
    else if (i <= 8) matrix[8][8 - (i - 7)] = bit;
    else matrix[14 - i][8] = bit;

    if (i < 8) matrix[dim - 1 - i][8] = bit;
    else matrix[8][dim - 8 + (i - 8)] = bit;
  }
}

function renderSVG(matrix: (boolean | null)[][], dim: number, size: number): string {
  let pathD = "";
  const quietZone = 2;
  const viewBoxSize = dim + quietZone * 2;
  for (let r = 0; r < dim; r++) {
    for (let c = 0; c < dim; c++) {
      if (matrix[r][c]) {
        pathD += `M${c + quietZone},${r + quietZone}h1v1h-1z `;
      }
    }
  }

  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${viewBoxSize} ${viewBoxSize}" width="${size}" height="${size}" shape-rendering="crispEdges">
    <rect width="${viewBoxSize}" height="${viewBoxSize}" fill="#ffffff"/>
    <path d="${pathD.trim()}" fill="#0f172a"/>
  </svg>`;
}

export function generateQRCodeSVG(text: string, size = 200): string {
  const enc = new TextEncoder().encode(text);
  const chosen = selectVersion(enc.length);
  const bits = buildBitStream(enc, chosen.dataCodewords);
  const fullCodewords = createCodewords(bits, chosen);

  const dim = chosen.version * 4 + 17;
  const matrix = createEmptyMatrix(dim);

  drawPatternsAndStructure(matrix, chosen, dim);
  placeDataBits(matrix, fullCodewords, dim);
  applyFormatInfo(matrix, dim);

  return renderSVG(matrix, dim, size);
}
