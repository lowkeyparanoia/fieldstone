// Without this the @tailwind directives in src/index.css are emitted verbatim,
// the browser discards them as invalid CSS, and every utility class in the app
// resolves to nothing. That is why the admin UI rendered as unstyled HTML.
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
