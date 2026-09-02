import { ComponentFixture, TestBed } from "@angular/core/testing";
import { RichTextComponent } from "./rich-text.component";
import { RichTextNode } from "../../core/models/education.model";

describe("RichTextComponent", () => {
  let component: RichTextComponent;
  let fixture: ComponentFixture<RichTextComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RichTextComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(RichTextComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should render null doc as empty string", () => {
    component.doc = null;
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.querySelector(".rich-text")?.innerHTML).toBe("");
  });

  it("should render undefined doc as empty string", () => {
    component.doc = undefined;
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.querySelector(".rich-text")?.innerHTML).toBe("");
  });

  it("should render plain text node with escape characters", () => {
    const doc: RichTextNode = {
      nodeType: "document",
      content: [
        {
          nodeType: "text",
          value: 'Hello <world> & "test"',
        },
      ],
    };
    component.doc = doc;
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Hello <world> & "test"');
  });

  it("should render text node with marks (bold, italic, underline, code)", () => {
    const doc: RichTextNode = {
      nodeType: "document",
      content: [
        {
          nodeType: "text",
          value: "BoldText",
          marks: [{ type: "bold" }],
        },
        {
          nodeType: "text",
          value: "ItalicText",
          marks: [{ type: "italic" }],
        },
        {
          nodeType: "text",
          value: "UnderlineText",
          marks: [{ type: "underline" }],
        },
        {
          nodeType: "text",
          value: "CodeText",
          marks: [{ type: "code" }],
        },
      ],
    };
    component.doc = doc;
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.querySelector("strong")?.textContent).toBe("BoldText");
    expect(compiled.querySelector("em")?.textContent).toBe("ItalicText");
    expect(compiled.querySelector("u")?.textContent).toBe("UnderlineText");
    expect(compiled.querySelector("code")?.textContent).toBe("CodeText");
  });

  it("should render paragraph, headings, lists, blockquote, hr", () => {
    const doc: RichTextNode = {
      nodeType: "document",
      content: [
        { nodeType: "paragraph", content: [{ nodeType: "text", value: "P1" }] },
        { nodeType: "heading-1", content: [{ nodeType: "text", value: "H1" }] },
        { nodeType: "heading-2", content: [{ nodeType: "text", value: "H2" }] },
        { nodeType: "heading-3", content: [{ nodeType: "text", value: "H3" }] },
        { nodeType: "heading-4", content: [{ nodeType: "text", value: "H4" }] },
        { nodeType: "heading-5", content: [{ nodeType: "text", value: "H5" }] },
        { nodeType: "heading-6", content: [{ nodeType: "text", value: "H6" }] },
        {
          nodeType: "unordered-list",
          content: [
            {
              nodeType: "list-item",
              content: [
                {
                  nodeType: "paragraph",
                  content: [{ nodeType: "text", value: "UL Item" }],
                },
              ],
            },
          ],
        },
        {
          nodeType: "ordered-list",
          content: [
            {
              nodeType: "list-item",
              content: [
                {
                  nodeType: "paragraph",
                  content: [{ nodeType: "text", value: "OL Item" }],
                },
              ],
            },
          ],
        },
        {
          nodeType: "blockquote",
          content: [{ nodeType: "text", value: "Quote" }],
        },
        { nodeType: "hr" },
      ],
    };

    component.doc = doc;
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.querySelector("p")?.textContent).toBe("P1");
    expect(compiled.querySelector("h1")?.textContent).toBe("H1");
    expect(compiled.querySelector("h2")?.textContent).toBe("H2");
    expect(compiled.querySelectorAll("h3")).toHaveSize(4);
    expect(compiled.querySelector("ul li")?.textContent).toBe("UL Item");
    expect(compiled.querySelector("ol li")?.textContent).toBe("OL Item");
    expect(compiled.querySelector("blockquote")?.textContent).toBe("Quote");
    expect(compiled.querySelector("hr")).toBeTruthy();
  });

  it("should render hyperlinks with safe and unsafe URIs", () => {
    const doc: RichTextNode = {
      nodeType: "document",
      content: [
        {
          nodeType: "hyperlink",
          data: { uri: "https://example.com" },
          content: [{ nodeType: "text", value: "Safe Link" }],
        },
        {
          nodeType: "hyperlink",
          data: { uri: "javascript:alert(1)" },
          content: [{ nodeType: "text", value: "Unsafe Link" }],
        },
        {
          nodeType: "hyperlink",
          data: {},
          content: [{ nodeType: "text", value: "Empty Link" }],
        },
      ],
    };

    component.doc = doc;
    fixture.detectChanges();

    const links = fixture.nativeElement.querySelectorAll("a");
    expect(links).toHaveSize(3);
    expect(links[0].getAttribute("href")).toBe("https://example.com");
    expect(links[1].getAttribute("href")).toBe("#");
    expect(links[2].getAttribute("href")).toBe("#");
  });

  it("should fallback to rendering child nodes for default/unknown nodeTypes", () => {
    const doc: RichTextNode = {
      nodeType: "document",
      content: [
        {
          nodeType: "custom-unknown-type" as any,
          content: [{ nodeType: "text", value: "Fallback content" }],
        },
      ],
    };

    component.doc = doc;
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain("Fallback content");
  });
});
