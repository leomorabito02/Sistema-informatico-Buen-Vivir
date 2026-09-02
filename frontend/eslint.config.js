// @ts-check
const tseslint = require("@typescript-eslint/eslint-plugin");
const tsParser = require("@typescript-eslint/parser");
const angularPlugin = require("@angular-eslint/eslint-plugin");
const angularTemplatePlugin = require("@angular-eslint/eslint-plugin-template");
const templateParser = require("@angular-eslint/template-parser");

/** @type {import('eslint').Linter.Config[]} */
module.exports = [
  // ─── TypeScript source files ───────────────────────────────────────────────
  {
    files: ["src/**/*.ts"],
    plugins: {
      "@typescript-eslint": tseslint,
      "@angular-eslint": angularPlugin,
    },
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        project: "./tsconfig.json",
        createDefaultProgram: true,
        experimentalDecorators: true,
      },
    },
    rules: {
      // Angular-specific rules
      "@angular-eslint/component-class-suffix": [
        "error",
        { suffixes: ["Component"] },
      ],
      "@angular-eslint/directive-class-suffix": [
        "error",
        { suffixes: ["Directive"] },
      ],
      "@angular-eslint/no-input-rename": "error",
      "@angular-eslint/no-output-rename": "error",
      "@angular-eslint/use-lifecycle-interface": "warn",

      // TypeScript rules
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/no-unused-vars": [
        "warn",
        { argsIgnorePattern: "^_" },
      ],
    },
  },

  // ─── TypeScript spec/test files (relaxed rules) ────────────────────────────
  {
    files: ["src/**/*.spec.ts"],
    plugins: {
      "@typescript-eslint": tseslint,
    },
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        project: "./tsconfig.spec.json",
        createDefaultProgram: true,
      },
    },
    rules: {
      // Test files intentionally use `as any` for partial mocks
      "@typescript-eslint/no-explicit-any": "off",
    },
  },

  // ─── Angular HTML templates ─────────────────────────────────────────────────
  {
    files: ["src/**/*.html"],
    plugins: {
      "@angular-eslint/template": angularTemplatePlugin,
    },
    languageOptions: {
      parser: templateParser,
    },
    rules: {
      "@angular-eslint/template/no-negated-async": "warn",
    },
  },

  // ─── Ignore build output and test configs ───────────────────────────────────
  {
    ignores: [
      "dist/**",
      "node_modules/**",
      "coverage/**",
      "reports/**",
      ".angular/**",
      ".stryker-tmp/**",
      "cypress/**",
    ],
  },
];
