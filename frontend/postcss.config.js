export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
  rules: [
    // import('postcss-import'),
    // import('tailwindcss/nesting'),
    import('tailwindcss'),
    import('autoprefixer'),
    import('postcss-nested'),
  ],
  postcss: {
    plugins: [
      import('postcss-import'),

      import('tailwindcss'),
      import('autoprefixer'),
      import('postcss-nested'),
    ],
  },
}
