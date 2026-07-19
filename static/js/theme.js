const themeToggle = document.querySelector('.theme-toggle');

function syncThemeToggle() {
  if (!themeToggle) return;
  const isDark = document.documentElement.dataset.theme === 'dark';
  themeToggle.setAttribute('aria-pressed', String(isDark));
  themeToggle.setAttribute('aria-label', `Switch to ${isDark ? 'light' : 'dark'} theme`);
}

function persistTheme(theme) {
  try { localStorage.setItem('karthik-theme', theme); } catch {}
  try {
    document.cookie = `karthik-theme=${theme}; Domain=.thisiskarthik.com; Path=/; Max-Age=31536000; SameSite=Lax; Secure`;
  } catch {}
}

themeToggle?.addEventListener('click', () => {
  const nextTheme = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
  document.documentElement.dataset.theme = nextTheme;
  persistTheme(nextTheme);
  syncThemeToggle();
});

persistTheme(document.documentElement.dataset.theme || 'light');
syncThemeToggle();
