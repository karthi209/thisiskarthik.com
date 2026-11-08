// Keep native disclosures as a fallback when modal dialogs are unavailable.
const collectionEntries = document.querySelectorAll('.shelf-entry, .game-entry');
const supportsCollectionDialog = typeof HTMLDialogElement !== 'undefined';
let collectionDialog;
let collectionTrigger;
let previousBodyOverflow;
const viewButtons = document.querySelectorAll('[data-game-view]');
const viewPanels = document.querySelectorAll('[data-game-panel]');

function selectGameView(selectedView) {
  for (const button of viewButtons) {
    const selected = button.dataset.gameView === selectedView;
    button.classList.toggle('is-active', selected);
    button.setAttribute('aria-pressed', String(selected));
  }
  for (const panel of viewPanels) panel.hidden = panel.dataset.gamePanel !== selectedView;
}

for (const button of viewButtons) {
  button.addEventListener('click', () => selectGameView(button.dataset.gameView));
}

function createCollectionDialog() {
  collectionDialog = document.createElement('dialog');
  collectionDialog.className = 'collection-dialog';
  collectionDialog.setAttribute('aria-labelledby', 'collection-dialog-title');
  collectionDialog.innerHTML = `
    <div class="collection-dialog-tools">
      <button type="button" class="collection-dialog-close" aria-label="Close details" autofocus><span aria-hidden="true">×</span></button>
    </div>
    <div class="collection-dialog-body"></div>`;
  document.body.append(collectionDialog);
  collectionDialog.querySelector('button').addEventListener('click', () => collectionDialog.close());
  // Only dismiss a click that both starts and ends on the backdrop.
  let backdropStart = false;
  collectionDialog.addEventListener('pointerdown', event => {
    backdropStart = event.target === collectionDialog;
  });
  collectionDialog.addEventListener('click', event => {
    if (backdropStart && event.target === collectionDialog) collectionDialog.close();
    backdropStart = false;
  });
  collectionDialog.addEventListener('close', () => {
    document.body.style.overflow = previousBodyOverflow;
    collectionTrigger?.focus({ preventScroll: true });
  });
}

function showCollectionEntry(entry) {
  if (!supportsCollectionDialog) {
    entry.open = true;
    return;
  }
  if (!collectionDialog) createCollectionDialog();
  collectionTrigger = entry.querySelector('summary');
  const title = entry.querySelector('.shelf-title, .game-card-title, .game-title');
  const content = entry.querySelector('.shelf-notes, .game-entry-details');
  const body = collectionDialog.querySelector('.collection-dialog-body');
  body.replaceChildren();

  // Ranked game rows use the same reviewed cover as their shelf entry.
  const shelfID = entry.closest('li')?.id.replace(/^playtime-/, 'game-');
  const cover = entry.querySelector('summary img') || document.getElementById(shelfID)?.querySelector('summary img');
  if (cover) {
    const art = document.createElement('div');
    art.className = 'collection-dialog-art';
    const image = cover.cloneNode(true);
    image.loading = 'eager';
    art.append(image);
    body.append(art);
  }
  body.classList.toggle('collection-dialog-body--text', !cover);
  const writing = document.createElement('div');
  writing.className = 'collection-dialog-writing';
  const heading = document.createElement('h2');
  heading.id = 'collection-dialog-title';
  heading.textContent = title?.textContent.trim() || 'Details';
  writing.append(heading);
  if (content) {
    const details = content.cloneNode(true);
    // A saved artifact is useful only when it is different from the main cover.
    for (const artifact of details.querySelectorAll('.game-artifact')) {
      if (cover && artifact.querySelector('img')?.getAttribute('src') === cover.getAttribute('src')) artifact.remove();
    }
    writing.append(details);
  }
  body.append(writing);
  if (!collectionDialog.open) {
    previousBodyOverflow = document.body.style.overflow;
    collectionDialog.showModal();
    document.body.style.overflow = 'hidden';
    collectionDialog.querySelector('button').focus({ preventScroll: true });
  }
  collectionDialog.scrollTop = 0;
}

// One handler per interaction, rather than listeners on every game and book.
for (const entry of collectionEntries) {
  if (supportsCollectionDialog) entry.querySelector('summary').setAttribute('aria-haspopup', 'dialog');
}

document.addEventListener('click', event => {
  if (!supportsCollectionDialog) return;
  const summary = event.target.closest('summary');
  const entry = summary?.parentElement;
  if (!entry?.matches('.shelf-entry, .game-entry')) return;
  event.preventDefault();
  showCollectionEntry(entry);
});

document.addEventListener('keydown', event => {
  if (event.key !== 'Escape') return;
  const entry = event.target.closest('.shelf-entry[open], .game-entry[open]');
  if (!entry) return;
  event.preventDefault();
  entry.open = false;
  entry.querySelector('summary').focus({ preventScroll: true });
});

function openCollectionTarget() {
  let id;
  try { id = decodeURIComponent(location.hash.slice(1)); } catch { return; }
  const target = document.getElementById(id);
  const panel = target?.closest('[data-game-panel]');
  if (panel) selectGameView(panel.dataset.gamePanel);
  const entry = target?.querySelector('.shelf-entry, .game-entry');
  if (!entry) return;
  showCollectionEntry(entry);
  if (!supportsCollectionDialog) requestAnimationFrame(() => target.scrollIntoView({ block: 'start' }));
}

openCollectionTarget();
window.addEventListener('hashchange', openCollectionTarget);
