import { writable } from 'svelte/store';

type Theme = 'dark' | 'light';

const STORAGE_KEY = 'rentobuy_theme';

function getInitialTheme(): Theme {
  if (typeof window === 'undefined') return 'dark';

  const saved = localStorage.getItem(STORAGE_KEY);
  if (saved === 'light' || saved === 'dark') {
    return saved;
  }

  // Default to dark mode (matching current behavior)
  return 'dark';
}

function createThemeStore() {
  const { subscribe, set, update } = writable<Theme>('dark');

  return {
    subscribe,
    initialize() {
      const theme = getInitialTheme();
      set(theme);
      applyTheme(theme);
    },
    toggle() {
      update(current => {
        const newTheme = current === 'dark' ? 'light' : 'dark';
        localStorage.setItem(STORAGE_KEY, newTheme);
        applyTheme(newTheme);
        return newTheme;
      });
    },
    set(theme: Theme) {
      localStorage.setItem(STORAGE_KEY, theme);
      applyTheme(theme);
      set(theme);
    }
  };
}

function applyTheme(theme: Theme) {
  if (typeof document === 'undefined') return;

  const root = document.documentElement;
  if (theme === 'dark') {
    root.classList.add('dark');
  } else {
    root.classList.remove('dark');
  }
}

export const theme = createThemeStore();
