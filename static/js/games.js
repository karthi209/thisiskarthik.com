const viewButtons = document.querySelectorAll('[data-game-view]');
const viewPanels = document.querySelectorAll('[data-game-panel]');

function selectGameView(selectedView) {
  for (const candidate of viewButtons) {
    const selected = candidate.dataset.gameView === selectedView;
    candidate.classList.toggle('is-active', selected);
    candidate.setAttribute('aria-pressed', String(selected));
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
  game.scrollIntoView({ behavior: 'smooth', block: 'center' });

  const when = game.dataset.gameDate ? ` · ${game.dataset.gameDate}` : '';
  rediscoveryMessage.textContent = `${game.dataset.gameTitle}${when}`;
});
