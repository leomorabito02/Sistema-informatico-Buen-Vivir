import type { ResumeData as RrResumeData } from "../vendor/schema/resume/data";
import { defaultResumeData } from "../vendor/schema/resume/default";
import type { Template } from "../vendor/schema/templates";

export interface InternalBasics {
  name?: string;
  email?: string;
  phone?: string;
  headline?: string;
  location?: string;
  website?: string;
  summary?: string;
  picture?: string;
  profiles?: Array<{
    id?: string;
    network?: string;
    username?: string;
    url?: string;
  }>;
}

export interface InternalWork {
  id?: string;
  company?: string;
  position?: string;
  startDate?: string;
  endDate?: string;
  description?: string;
}

export interface InternalEducation {
  id?: string;
  institution?: string;
  degree?: string;
  field?: string;
  startDate?: string;
  endDate?: string;
  grade?: string;
  status?: string;
}

export interface InternalSkill {
  id?: string;
  name?: string;
  level?: string;
}

export interface InternalCertification {
  id?: string;
  name?: string;
  issuer?: string;
  date?: string;
}

export interface InternalLanguage {
  id?: string;
  name?: string;
  level?: string;
}

export interface InternalAward {
  id?: string;
  title?: string;
  awarder?: string;
  date?: string;
  summary?: string;
}

export interface InternalVolunteer {
  id?: string;
  organization?: string;
  position?: string;
  startDate?: string;
  endDate?: string;
  summary?: string;
}

export interface InternalProject {
  id?: string;
  name?: string;
  description?: string;
  startDate?: string;
  endDate?: string;
  url?: string;
  keywords?: string[];
}

export interface InternalReference {
  id?: string;
  name?: string;
  description?: string;
}

export interface InternalInterest {
  id?: string;
  name?: string;
  keywords?: string[];
}

export interface InternalDesignColors {
  primary?: string;
  [key: string]: unknown;
}

export interface InternalDesign {
  colors?: InternalDesignColors;
  [key: string]: unknown;
}

export interface InternalMetadata {
  template?: string;
  primaryColor?: string;
  design?: InternalDesign;
  [key: string]: unknown;
}

export interface InternalCvData {
  basics?: InternalBasics;
  work?: InternalWork[];
  education?: InternalEducation[];
  skills?: InternalSkill[];
  certifications?: InternalCertification[];
  languages?: InternalLanguage[];
  awards?: InternalAward[];
  volunteer?: InternalVolunteer[];
  projects?: InternalProject[];
  references?: InternalReference[];
  interests?: InternalInterest[];
  metadata?: InternalMetadata;
}

export interface InternalResume {
  id?: string;
  user_id?: string;
  data?: InternalCvData;
}

const VALID_TEMPLATES = new Set<Template>([
  "azurill",
  "bronzor",
  "chikorita",
  "ditgar",
  "ditto",
  "gengar",
  "glalie",
  "kakuna",
  "lapras",
  "leafish",
  "meowth",
  "onyx",
  "pikachu",
  "rhyhorn",
  "scizor",
]);

export function hexToRgba(hexOrColor: string, alpha = 1): string {
  if (!hexOrColor) return "rgba(72, 88, 57, 1)";
  const trimmed = hexOrColor.trim();
  if (trimmed.startsWith("rgb")) return trimmed;
  let hex = trimmed.replace(/^#/, "");
  if (hex.length === 3) {
    hex = hex
      .split("")
      .map((c) => c + c)
      .join("");
  }
  if (hex.length === 6) {
    const num = Number.parseInt(hex, 16);
    const r = (num >> 16) & 255;
    const g = (num >> 8) & 255;
    const b = num & 255;
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
  }
  return trimmed;
}

function isValidPdfPictureUrl(url: string): boolean {
  if (!url) return false;
  if (url.startsWith("data:image/") || url.startsWith("blob:")) return true;
  try {
    const parsed = new URL(url, "http://localhost");
    const path = parsed.pathname.toLowerCase();
    return (
      path.endsWith(".jpg") ||
      path.endsWith(".jpeg") ||
      path.endsWith(".png") ||
      path.endsWith(".webp")
    );
  } catch {
    return false;
  }
}

function mapBasics(base: RrResumeData, basics?: InternalBasics, pictureUrl?: string): void {
  if (basics) {
    base.basics.name = basics.name || "";
    base.basics.headline = basics.headline || "";
    base.basics.email = basics.email || "";
    base.basics.phone = basics.phone || "";
    base.basics.location = basics.location || "";
    base.basics.website = { url: basics.website || "", label: basics.website || "" };

    if (basics.profiles && basics.profiles.length > 0) {
      base.basics.customFields = basics.profiles.map((p, idx) => ({
        id: p.id || `profile-${idx}`,
        icon: p.network || "globe",
        text: p.username || "",
        link: p.url || "",
      }));
    }

    if (basics.summary) {
      base.summary.content = basics.summary;
      base.summary.hidden = false;
    }
  }

  const resolvedPic = (pictureUrl ?? basics?.picture ?? "").trim();
  if (resolvedPic && isValidPdfPictureUrl(resolvedPic)) {
    base.picture = { ...base.picture, url: resolvedPic, hidden: false };
  } else {
    base.picture.hidden = true;
  }
}

function mapWork(base: RrResumeData, work?: InternalWork[]): void {
  if (!work || work.length === 0) return;
  base.sections.experience.items = work.map((w, idx) => ({
    id: w.id || `work-${idx}`,
    hidden: false,
    company: w.company || "",
    position: w.position || "",
    location: "",
    period: [w.startDate, w.endDate].filter(Boolean).join(" - "),
    website: { url: "", label: "", inlineLink: false },
    description: w.description || "",
    roles: [],
  }));
  base.sections.experience.hidden = false;
}

function mapEducation(base: RrResumeData, education?: InternalEducation[]): void {
  if (!education || education.length === 0) return;
  base.sections.education.items = education.map((e, idx) => ({
    id: e.id || `edu-${idx}`,
    hidden: false,
    school: e.institution || "",
    degree: e.degree || "",
    area: e.field || "",
    grade: e.grade || "",
    location: "",
    period: [e.startDate, e.endDate].filter(Boolean).join(" - "),
    website: { url: "", label: "", inlineLink: false },
    description: e.status || "",
  }));
  base.sections.education.hidden = false;
}

function mapSkills(base: RrResumeData, skills?: InternalSkill[]): void {
  if (!skills || skills.length === 0) return;
  base.sections.skills.items = skills.map((s, idx) => ({
    id: s.id || `skill-${idx}`,
    hidden: false,
    name: s.name || "",
    proficiency: s.level || "",
    level: 0,
    icon: "",
    iconColor: "",
    keywords: [],
  }));
  base.sections.skills.hidden = false;
}

function mapLanguages(base: RrResumeData, languages?: InternalLanguage[]): void {
  if (!languages || languages.length === 0) return;
  base.sections.languages.items = languages.map((l, idx) => ({
    id: l.id || `lang-${idx}`,
    hidden: false,
    language: l.name || "",
    fluency: l.level || "",
    level: 0,
  }));
  base.sections.languages.hidden = false;
}

function mapCertifications(base: RrResumeData, certifications?: InternalCertification[]): void {
  if (!certifications || certifications.length === 0) return;
  base.sections.certifications.items = certifications.map((c, idx) => ({
    id: c.id || `cert-${idx}`,
    hidden: false,
    title: c.name || "",
    issuer: c.issuer || "",
    date: c.date || "",
    website: { url: "", label: "", inlineLink: false },
    description: "",
  }));
  base.sections.certifications.hidden = false;
}

function mapAwards(base: RrResumeData, awards?: InternalAward[]): void {
  if (!awards || awards.length === 0) return;
  base.sections.awards.items = awards.map((a, idx) => ({
    id: a.id || `award-${idx}`,
    hidden: false,
    title: a.title || "",
    awarder: a.awarder || "",
    date: a.date || "",
    website: { url: "", label: "", inlineLink: false },
    description: a.summary || "",
  }));
  base.sections.awards.hidden = false;
}

function mapVolunteer(base: RrResumeData, volunteer?: InternalVolunteer[]): void {
  if (!volunteer || volunteer.length === 0) return;
  base.sections.volunteer.items = volunteer.map((v, idx) => ({
    id: v.id || `vol-${idx}`,
    hidden: false,
    organization: v.organization || "",
    location: "",
    period: [v.startDate, v.endDate].filter(Boolean).join(" - "),
    website: { url: "", label: "", inlineLink: false },
    description: v.summary || "",
  }));
  base.sections.volunteer.hidden = false;
}

function mapProjects(base: RrResumeData, projects?: InternalProject[]): void {
  if (!projects || projects.length === 0) return;
  base.sections.projects.items = projects.map((p, idx) => ({
    id: p.id || `proj-${idx}`,
    hidden: false,
    name: p.name || "",
    description: p.description || "",
    period: [p.startDate, p.endDate].filter(Boolean).join(" - "),
    website: { url: p.url || "", label: p.url || "", inlineLink: false },
    keywords: p.keywords || [],
  }));
  base.sections.projects.hidden = false;
}

function mapReferences(base: RrResumeData, references?: InternalReference[]): void {
  if (!references || references.length === 0) return;
  base.sections.references.items = references.map((r, idx) => ({
    id: r.id || `ref-${idx}`,
    hidden: false,
    name: r.name || "",
    position: "",
    website: { url: "", label: "", inlineLink: false },
    phone: "",
    description: r.description || "",
  }));
  base.sections.references.hidden = false;
}

function mapInterests(base: RrResumeData, interests?: InternalInterest[]): void {
  if (!interests || interests.length === 0) return;
  base.sections.interests.items = interests.map((i, idx) => ({
    id: i.id || `int-${idx}`,
    hidden: false,
    name: i.name || "",
    icon: "",
    iconColor: "",
    keywords: i.keywords || [],
  }));
  base.sections.interests.hidden = false;
}

/**
 * Maps an internal Buen Vivir CV to Reactive Resume's ResumeData schema.
 */
export function mapToResumeData(
  input: InternalResume | InternalCvData,
  opts: { pictureUrl?: string; overridePrimaryColor?: string } = {}
): RrResumeData {
  const cv: InternalCvData = "data" in input && input.data ? input.data : (input as InternalCvData);
  const base = structuredClone(defaultResumeData);

  // 1. Template
  const rawTemplate = cv.metadata?.template;
  base.metadata.template =
    typeof rawTemplate === "string" && VALID_TEMPLATES.has(rawTemplate as Template)
      ? (rawTemplate as Template)
      : "azurill";

  // 2. Primary Color
  const rawPrimary =
    opts.overridePrimaryColor ||
    cv.metadata?.design?.colors?.primary ||
    cv.metadata?.primaryColor ||
    "#485839";
  base.metadata.design.colors.primary = hexToRgba(rawPrimary);

  // 3. Typography (Poppins default)
  base.metadata.typography = {
    heading: {
      fontFamily: "Poppins",
      fontWeights: ["600", "700"],
      fontSize: 14,
      lineHeight: 1.5,
    },
    body: {
      fontFamily: "Poppins",
      fontWeights: ["400", "600"],
      fontSize: 10,
      lineHeight: 1.5,
    },
  };

  // 4. Sections & Fields
  mapBasics(base, cv.basics, opts.pictureUrl);
  mapWork(base, cv.work);
  mapEducation(base, cv.education);
  mapSkills(base, cv.skills);
  mapLanguages(base, cv.languages);
  mapCertifications(base, cv.certifications);
  mapAwards(base, cv.awards);
  mapVolunteer(base, cv.volunteer);
  mapProjects(base, cv.projects);
  mapReferences(base, cv.references);
  mapInterests(base, cv.interests);

  return base;
}
