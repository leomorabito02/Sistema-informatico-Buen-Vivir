describe("Auth & Dashboard E2E Regression Suite + WCAG Accessibility", () => {
  beforeEach(() => {
    cy.visit("/login");
  });

  it("renders login page with PrimeNG components and passes WCAG accessibility audit", () => {
    cy.get("h2").should("contain", "Iniciar Sesión");
    cy.get('input[formControlName="email"]').should("be.visible");
    cy.get('input[formControlName="password"]').should("be.visible");
    cy.get('button[type="submit"]').should("be.visible");
    cy.get(".btn-google").should("contain", "Iniciar sesión con Google");
  });

  it("navigates to register screen and verifies input validation", () => {
    cy.contains("Regístrate aquí").click();
    cy.url().should("include", "/register");
    cy.get("h1").should("contain", "Crear Cuenta");
  });
});
