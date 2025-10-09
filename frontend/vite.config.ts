import * as path from 'path';
import { defineConfig, loadEnv } from 'vite';

const getEnvResilient = (mode: string, envDir: string) => {
  // Estratégia de fallback para diferentes locais de .env
  const envPaths = [
    envDir,                    // diretório raiz
    path.join(envDir, 'config'), // ./config/
    path.join(envDir, '..'),   // diretório pai
  ];

  let finalEnv = {};

  for (const envPath of envPaths) {
    try {
      const env = loadEnv(mode, envPath, ['GEMINI_', 'GITHUB_', 'JIRA_', 'API_']);
      finalEnv = { ...finalEnv, ...env };
    } catch (error) {
      // Silently continue to next path
      continue;
    }
  }

  return finalEnv;
};

export default defineConfig(({ mode }) => {
  const env: Record<string, string> = getEnvResilient(mode, process.cwd());
  return {
    root: '.',
    base: './',
    publicDir: 'public',
    cacheDir: 'node_modules/.vite',
    mode: mode,
    define: {
      'process.env.API_KEY': JSON.stringify(env.VITE_GEMINI_API_KEY || env.GEMINI_API_KEY || ""),
      'process.env.GEMINI_API_KEY': JSON.stringify(env.VITE_GEMINI_API_KEY || env.GEMINI_API_KEY || ""),
      'process.env.GITHUB_PAT': JSON.stringify(env.VITE_GITHUB_PAT || env.GITHUB_PAT || ""),
      'process.env.JIRA_API_TOKEN': JSON.stringify(env.VITE_JIRA_API_TOKEN || env.JIRA_API_TOKEN || ""),
      'process.env.JIRA_INSTANCE_URL': JSON.stringify(env.VITE_JIRA_INSTANCE_URL || env.JIRA_INSTANCE_URL || ""),
      'process.env.JIRA_USER_EMAIL': JSON.stringify(env.VITE_JIRA_USER_EMAIL || env.JIRA_USER_EMAIL || "")
    },
    build: {
      rollupOptions: {
        external: [
          'buffer', 'stream', 'util', 'events', 'http', 'https', 'url', 'zlib', 'crypto',
          './src/components/layout/Footer', './src/components/layout/Header', './src/components/layout/Navbar',
        ],
        input: {
          main: path.resolve(__dirname, 'index.html')
        },
        output: {
          manualChunks: {
            vendor: [
              'react',
              'react-dom',
              'framer-motion',
              'react-markdown'
            ]
          },
          plugins: []
        }
      },
      outDir: 'dist',
      sourcemap: false,
      chunkSizeWarningLimit: 900,
    },
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: `@import "@/src/styles/index.scss";`
        }
      }
    },
    optimizeDeps: {
      include: ['react', 'react-dom', 'framer-motion', 'react-markdown'],
    },
    esbuild: {
      drop: ['console', 'debugger'],
    },
  };
});
