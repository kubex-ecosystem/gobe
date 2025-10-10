/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ['class', '[data-theme="dark"]'],
  theme: {
    extend: {
      colors: {
        bg: 'var(--kubex-bg)',
        surface: 'var(--kubex-surface)',
        border: 'var(--kubex-border)',
        text: {
          head: 'var(--kubex-text-head)',
          body: 'var(--kubex-text-body)',
        },
        primary: {
          DEFAULT: 'var(--kubex-primary)',
          hover: 'var(--kubex-primary-hover)',
          50: 'var(--kubex-primary-50)',
        },
        accent: {
          lilac: 'var(--kubex-accent-1)',
          fuchsia: 'var(--kubex-accent-2)',
        },
        success: 'var(--kubex-success)',
        warning: 'var(--kubex-warn)',
        danger: 'var(--kubex-danger)',
      },
      borderRadius: { xl: 'var(--radius)', '2xl': 'calc(var(--radius) + 8px)' },
      boxShadow: { card: 'var(--shadow-card)' },
    }
  },
  plugins: [
    function ({ addComponents, addUtilities, theme }) {
      addComponents({
        '.btn': {
          display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem',
          padding: '.625rem .95rem', borderRadius: theme('borderRadius.xl'),
          background: 'var(--kubex-primary)', color: '#0b0f12', fontWeight: 600,
          transition: 'transform .12s ease, box-shadow .12s ease',
        },
        '.btn:hover': { transform: 'translateY(-1px)', background: 'var(--kubex-primary-hover)' },
        '.btn:focus-visible': { boxShadow: 'var(--ring)' },
        '.btn-ghost': {
          background: 'transparent', color: 'var(--kubex-text-head)',
          border: '1px solid var(--kubex-border)'
        },
        '.btn-danger': { background: 'var(--kubex-danger)', color: '#fff' },
      });

      addComponents({
        '.input': {
          width: '100%', background: 'var(--kubex-surface)', color: 'var(--kubex-text-body)',
          border: '1px solid var(--kubex-border)', borderRadius: theme('borderRadius.xl'),
          padding: '.55rem .8rem', outline: 'none',
          transition: 'box-shadow .12s, border-color .12s'
        },
        '.input::placeholder': { color: '#94a3b8' },
        '.input:focus': { boxShadow: 'var(--ring)', borderColor: 'var(--kubex-primary)' }
      });

      addComponents({
        '.card': {
          background: 'var(--kubex-surface)', border: '1px solid var(--kubex-border)',
          borderRadius: theme('borderRadius.2xl'), boxShadow: theme('boxShadow.card'),
          padding: '1rem', color: 'var(--kubex-text-body)'
        }
      });

      addUtilities({
        '.prose-head': { color: 'var(--kubex-text-head)' },
        '.prose-body': { color: 'var(--kubex-text-body)' },
        '.glow-cyan': { boxShadow: '0 0 0 9999px transparent, 0 0 40px var(--kubex-glow-cyan) inset' },
        '.glow-lilac': { boxShadow: '0 0 0 9999px transparent, 0 0 40px var(--kubex-glow-lilac) inset' },
      });
    }
  ]
};
