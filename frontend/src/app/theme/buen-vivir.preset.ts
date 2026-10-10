import { definePreset } from "@primeuix/themes";
import Aura from "@primeuix/themes/aura";

/**
 * Tema de PrimeNG para Buen Vivir.
 *
 * Extiende Aura, pero cada color, radio, foco y duración apunta a los
 * design tokens de `src/styles.css` (variables CSS `--green-*`,
 * `--neutral-*`, `--radius-*`, etc.). Así hay una sola fuente de verdad:
 * cambiar un token cambia a la vez los estilos propios y los componentes
 * de PrimeNG (botones, tablas, diálogos, toasts, campos).
 *
 * Las paletas primitivas de estado (green, red, orange, sky) también se
 * remapean: Aura las usa para las severidades `success`, `danger`, `warn`
 * e `info` de botones, tags, mensajes y toasts.
 */
const scale = (name: string) => ({
  50: `var(--${name}-50)`,
  100: `var(--${name}-100)`,
  200: `var(--${name}-200)`,
  300: `var(--${name}-300)`,
  400: `var(--${name}-400)`,
  500: `var(--${name}-500)`,
  600: `var(--${name}-600)`,
  700: `var(--${name}-700)`,
  800: `var(--${name}-800)`,
  900: `var(--${name}-900)`,
  950: `var(--${name}-950)`,
});

export const BuenVivirPreset = definePreset(Aura, {
  primitive: {
    borderRadius: {
      none: "0",
      xs: "var(--radius-xs)",
      sm: "var(--radius-sm)",
      md: "var(--radius-md)",
      lg: "var(--radius-lg)",
      xl: "var(--radius-xl)",
    },
    green: scale("green"),
    emerald: scale("green"),
    red: scale("red"),
    orange: scale("ochre"),
    yellow: scale("ochre"),
    amber: scale("ochre"),
    sky: scale("blue"),
    blue: scale("blue"),
  },
  semantic: {
    transitionDuration: "var(--duration-base)",
    typography: {
      fontFamily: "var(--font-family-base)",
      fontSize: "var(--font-size-base)",
      lineHeight: "var(--line-height-base)",
    },
    focusRing: {
      width: "2px",
      style: "solid",
      color: "var(--color-focus)",
      offset: "2px",
      shadow: "none",
    },
    primary: {
      ...scale("green"),
      color: "var(--color-primary)",
      contrastColor: "var(--color-on-primary)",
      hoverColor: "var(--color-primary-hover)",
      activeColor: "var(--color-primary-active)",
    },
    surface: {
      0: "var(--neutral-0)",
      ...scale("neutral"),
    },
    text: {
      color: "var(--color-text)",
      hoverColor: "var(--color-text)",
      mutedColor: "var(--color-text-muted)",
      hoverMutedColor: "var(--neutral-700)",
    },
    highlight: {
      background: "var(--color-primary-soft)",
      focusBackground: "var(--color-primary-soft-strong)",
      color: "var(--green-800)",
      focusColor: "var(--green-900)",
    },
    mask: {
      transitionDuration: "var(--duration-slow)",
      background: "var(--color-overlay)",
      color: "var(--neutral-200)",
    },
    formField: {
      paddingX: "var(--space-3)",
      paddingY: "var(--space-2)",
      borderColor: "var(--color-border-strong)",
      hoverBorderColor: "var(--neutral-700)",
      focusBorderColor: "var(--color-primary)",
      invalidBorderColor: "var(--color-danger)",
      color: "var(--color-text)",
      placeholderColor: "var(--color-text-muted)",
      floatLabelColor: "var(--color-text-muted)",
      floatLabelFocusColor: "var(--color-primary)",
      iconColor: "var(--color-text-muted)",
    },
  },
  // Formas al estilo de las apps de comunidad (Humand): botones en
  // pastilla y tarjetas/diálogos bien redondeados.
  components: {
    button: { root: { borderRadius: "var(--radius-full)" } },
    card: { root: { borderRadius: "var(--radius-xl)" } },
    dialog: { root: { borderRadius: "var(--radius-xl)" } },
  },
});
