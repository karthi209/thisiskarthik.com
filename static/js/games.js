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

function revealGameTarget() {
  let id;
  try { id = decodeURIComponent(location.hash.slice(1)); } catch { return; }
  const target = document.getElementById(id);
  const panel = target?.closest('[data-game-panel]');
  if (panel) selectGameView(panel.dataset.gamePanel);
}

revealGameTarget();
window.addEventListener('hashchange', revealGameTarget);
