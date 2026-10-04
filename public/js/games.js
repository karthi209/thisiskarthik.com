const viewButtons = document.querySelectorAll('[data-game-view]');
const viewPanels = document.querySelectorAll('[data-game-panel]');

function selectGameView(selectedView) {
  for (const candidate of viewButtons) {
    const selected = candidate.dataset.gameView === selectedView;
    candidate.classList.toggle('is-active', selected);
    candidate.setAttribute('aria-pressed', String(selected));
  }
  for (const link of document.querySelectorAll('[data-game-section]')) {
    const shelf = selectedView === 'shelf';
    link.hash = link.dataset.gameSection === 'history'
      ? (shelf ? 'collection-heading' : 'history-heading')
      : (shelf ? 'shelf-backlog' : 'backlog-heading');
  }
  for (const panel of viewPanels) {
    panel.hidden = panel.dataset.gamePanel !== selectedView;
  }
}

for (const button of viewButtons) {
  button.addEventListener('click', () => {
    selectGameView(button.dataset.gameView);
  });
}

const rediscoveryButton = document.querySelector('[data-random-game]');
const rediscoveryMessage = document.querySelector('[data-random-game-message]');
const rememberedGames = document.querySelectorAll('[data-memory-game]');

if (rediscoveryButton && rememberedGames.length === 0) {
  rediscoveryButton.hidden = true;
}

rediscoveryButton?.addEventListener('click', () => {
  const game = rememberedGames[Math.floor(Math.random() * rememberedGames.length)];
  const details = game.querySelector('details');
  selectGameView('years');
  details.open = true;
  details.querySelector('summary').focus({ preventScroll: true });
  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  game.scrollIntoView({ behavior: reducedMotion ? 'instant' : 'smooth', block: 'center' });

  const when = game.dataset.gameDate ? ` · ${game.dataset.gameDate}` : '';
  rediscoveryMessage.textContent = `${game.dataset.gameTitle}${when}`;
});

// Historical machine links open the journal where their annotations live.
for (const link of document.querySelectorAll('[data-game-section]')) {
  link.addEventListener('click', () => {
    if (!document.querySelector('[data-game-panel="playtime"]').hidden) {
      selectGameView('shelf');
    }
  });
}

const initialTarget = document.getElementById(location.hash.slice(1));
if (initialTarget?.closest('[data-game-panel="years"]')) {
  selectGameView('years');
} else if (initialTarget?.closest('[data-game-panel="playtime"]')) {
  selectGameView('playtime');
}

for (const link of document.querySelectorAll('[data-game-jump]')) {
  link.addEventListener('click', () => selectGameView('years'));
}
