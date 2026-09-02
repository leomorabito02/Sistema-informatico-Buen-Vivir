import { Component, Input, SecurityContext } from "@angular/core";
import { CommonModule } from "@angular/common";
import { DomSanitizer, SafeHtml } from "@angular/platform-browser";
import { RichTextNode } from "../../core/models/education.model";

/**
 * Renderiza un documento Rich Text de Contentful (el campo `description`
 * de un curso) respetando el formato: párrafos, títulos, listas,
 * negrita/cursiva/subrayado/código, citas, separadores y enlaces.
 *
 * Uso: <app-rich-text [doc]="curso.description"></app-rich-text>
 *
 * Convierte el árbol a HTML acotado y lo pasa por DomSanitizer
 * (SecurityContext.HTML) antes de bindearlo, así que no ejecuta scripts
 * ni atributos peligrosos aunque el contenido venga manipulado.
 */
@Component({
  selector: "app-rich-text",
  standalone: true,
  imports: [CommonModule],
  template: `<div class="rich-text" [innerHTML]="html"></div>`,
  styles: [
    `
      .rich-text {
        line-height: 1.6;
        color: var(--text-main, #1d1d1b);
      }
      .rich-text :first-child {
        margin-top: 0;
      }
      .rich-text :last-child {
        margin-bottom: 0;
      }
      .rich-text h1,
      .rich-text h2,
      .rich-text h3 {
        font-family: "Palo Seco", "Poppins", sans-serif;
        margin: 1.2rem 0 0.5rem;
        line-height: 1.25;
      }
      .rich-text p {
        margin: 0 0 0.75rem;
      }
      .rich-text ul,
      .rich-text ol {
        margin: 0 0 0.75rem 1.25rem;
      }
      .rich-text li {
        margin: 0.2rem 0;
      }
      .rich-text blockquote {
        margin: 0 0 0.75rem;
        padding: 0.25rem 0 0.25rem 0.9rem;
        border-left: 3px solid var(--primary, #485839);
        color: #555;
      }
      .rich-text a {
        color: var(--primary, #485839);
        text-decoration: underline;
      }
      .rich-text code {
        background: rgba(0, 0, 0, 0.06);
        padding: 0.1rem 0.3rem;
        border-radius: 4px;
        font-size: 0.9em;
      }
      .rich-text hr {
        border: none;
        border-top: 1px solid var(--border, #e2e2d8);
        margin: 1rem 0;
      }
    `,
  ],
})
export class RichTextComponent {
  private _doc: RichTextNode | null = null;
  html: SafeHtml = "";

  @Input() set doc(value: RichTextNode | null | undefined) {
    this._doc = value ?? null;
    const raw = this._doc ? this.renderNodes(this._doc.content ?? []) : "";
    // Sanitiza y re-marca como seguro: si el sanitizer recorta algo, ya no
    // vuelve a bindearse crudo.
    const clean = this.sanitizer.sanitize(SecurityContext.HTML, raw) ?? "";
    this.html = this.sanitizer.bypassSecurityTrustHtml(clean);
  }

  constructor(private sanitizer: DomSanitizer) {}

  private esc(s: string): string {
    return (s ?? "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  private renderNodes(nodes: RichTextNode[]): string {
    return nodes.map((n) => this.renderNode(n)).join("");
  }

  private renderNode(node: RichTextNode): string {
    switch (node.nodeType) {
      case "text": {
        let t = this.esc(node.value ?? "");
        for (const m of node.marks ?? []) {
          if (m.type === "bold") t = `<strong>${t}</strong>`;
          else if (m.type === "italic") t = `<em>${t}</em>`;
          else if (m.type === "underline") t = `<u>${t}</u>`;
          else if (m.type === "code") t = `<code>${t}</code>`;
        }
        return t;
      }
      case "paragraph":
        return `<p>${this.renderNodes(node.content ?? [])}</p>`;
      case "heading-1":
        return `<h1>${this.renderNodes(node.content ?? [])}</h1>`;
      case "heading-2":
        return `<h2>${this.renderNodes(node.content ?? [])}</h2>`;
      case "heading-3":
      case "heading-4":
      case "heading-5":
      case "heading-6":
        return `<h3>${this.renderNodes(node.content ?? [])}</h3>`;
      case "unordered-list":
        return `<ul>${this.renderNodes(node.content ?? [])}</ul>`;
      case "ordered-list":
        return `<ol>${this.renderNodes(node.content ?? [])}</ol>`;
      case "list-item":
        return `<li>${this.renderNodes(node.content ?? [])}</li>`;
      case "blockquote":
        return `<blockquote>${this.renderNodes(node.content ?? [])}</blockquote>`;
      case "hr":
        return `<hr />`;
      case "hyperlink": {
        const uri = String(
          (node.data as { uri?: string } | undefined)?.uri ?? "#",
        );
        const safeUri = /^(https?:|mailto:|\/)/i.test(uri) ? uri : "#";
        return `<a href="${this.esc(safeUri)}" target="_blank" rel="noopener noreferrer">${this.renderNodes(node.content ?? [])}</a>`;
      }
      default:
        // nodos no soportados (assets embebidos, etc.): renderizar su texto
        return this.renderNodes(node.content ?? []);
    }
  }
}
