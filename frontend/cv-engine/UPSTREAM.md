# Upstream Vendor Information

- **Upstream Repository**: [Reactive Resume](https://github.com/AmruthPillai/Reactive-Resume)
- **Version**: 5.3.1 (Commit: `b953435f2c66134ec9f99c747d5b14a122c846f5`)
- **License**: MIT (see [LICENSE-REACTIVE-RESUME](./LICENSE-REACTIVE-RESUME))
- **Vendored Packages**:
  - `packages/schema/src` -> `vendor/schema`
  - `packages/utils/src` -> `vendor/utils`
  - `packages/resume/src` -> `vendor/resume`
  - `packages/pdf/src` -> `vendor/pdf`
- **Modifications**:
  - Excluded test files and fixtures
  - Decoupled from `@reactive-resume/fonts` to use offline, self-hosted TTF fonts (Poppins, IBM Plex Serif, IBM Plex Sans)
  - Decoupled from CJK dependencies (`cjk-regex`, Noto fonts) to keep bundle small and deterministic
