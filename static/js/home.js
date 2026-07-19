const insightButton = document.querySelector('.insight-button');
const insightOutput = document.querySelector('.insight-output');
const warning = document.querySelector('.entry-warning');
const warningMessage = document.querySelector('.warning-message');
const vadivelu = document.querySelector('.hero-banner');

const insights = [
  'MTC buses never arrive when scheduled.',
  'The year of the Linux desktop is here, old man.',
  'Once you watch cricket live, it gives you a different perspective of the game.',
  'Things here may contradict each other and that is not a bug.'
];

let lastInsight = -1;

insightButton?.addEventListener('click', () => {
  let clicks = 0;
  try { clicks = Number(sessionStorage.getItem('karthik-insight-clicks')) || 0; } catch {}
  clicks += 1;
  try { sessionStorage.setItem('karthik-insight-clicks', String(clicks)); } catch {}

  if (clicks >= 5) {
    insightOutput.textContent = 'Again? Go touch grass... or at least a dumbbell.';
    return;
  }

  let next = Math.floor(Math.random() * insights.length);
  if (insights.length > 1 && next === lastInsight) next = (next + 1) % insights.length;
  lastInsight = next;
  insightOutput.textContent = insights[next];
});

let warningTimer;
const originalWarning = warningMessage?.textContent || '';
if (window.matchMedia('(hover: hover)').matches) {
  warning?.addEventListener('pointerenter', () => {
    warningMessage.textContent = 'Please lower your expectations in an orderly manner.';
    window.clearTimeout(warningTimer);
    warningTimer = window.setTimeout(() => { warningMessage.textContent = originalWarning; }, 2000);
  }, {passive: true});
}

let typed = '';
document.addEventListener('keydown', event => {
  if (event.metaKey || event.ctrlKey || event.altKey || event.key.length !== 1) return;
  if (['INPUT', 'TEXTAREA', 'SELECT'].includes(document.activeElement?.tagName)) return;
  typed = `${typed}${event.key.toLowerCase()}`.slice(-8);
  if (typed !== 'vadivelu') return;
  vadivelu?.classList.remove('is-dancing');
  requestAnimationFrame(() => vadivelu?.classList.add('is-dancing'));
  window.setTimeout(() => vadivelu?.classList.remove('is-dancing'), 2400);
  typed = '';
});
