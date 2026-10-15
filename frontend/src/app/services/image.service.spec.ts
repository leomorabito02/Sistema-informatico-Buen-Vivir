import { TestBed } from "@angular/core/testing";
import { ImageService } from "./image.service";
import { environment } from "../../environments/environment";

describe("ImageService", () => {
  let service: ImageService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ImageService);
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should return empty string for null, undefined, boolean, or empty string", () => {
    expect(service.resolveImageUrl(null)).toBe("");
    expect(service.resolveImageUrl(undefined)).toBe("");
    expect(service.resolveImageUrl("")).toBe("");
    expect(service.resolveImageUrl(true)).toBe("");
    expect(service.resolveImageUrl(false)).toBe("");
  });

  it("should return absolute remote URLs unchanged", () => {
    const url = "https://example.com/avatar.jpg";
    expect(service.resolveImageUrl(url)).toBe(url);
  });

  it("should replace localhost:8080 API prefix with environment.apiUrl", () => {
    const localUrl = "http://localhost:8080/api/v1/auth/uploads/user.png";
    const resolved = service.resolveImageUrl(localUrl);
    expect(resolved).toBe(`${environment.apiUrl || ""}/api/v1/auth/uploads/user.png`);
  });

  it("should prepend environment.apiUrl to relative paths", () => {
    const relativeUrl = "/uploads/photo.jpg";
    const resolved = service.resolveImageUrl(relativeUrl);
    expect(resolved).toBe(`${environment.apiUrl || ""}/uploads/photo.jpg`);
  });
});
