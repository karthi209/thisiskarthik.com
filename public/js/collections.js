// Keep native disclosures as a fallback when modal dialogs are unavailable.
const collectionEntries = document.querySelectorAll('.shelf-entry, .game-entry');
const supportsCollectionDialog = typeof HTMLDialogElement !== 'undefined';
let collectionDialog;
let collectionTrigger;
let previousBodyOverflow;
// Rearrange the original cards: each game keeps one disclosure and one stable ID.
function createGameBrowser() {
  const controls = document.querySelector('.game-browser-controls');
  const results = document.getElementById('game-browser-results');
  if (!controls || !results) return null;
  const arrange = document.getElementById('game-arrange');
  const chips = controls.querySelector('.game-filter-chips');
  const clear = controls.querySelector('.game-filter-clear');
  const count = controls.querySelector('.game-result-count');
  const panels = document.querySelectorAll('[data-game-panel]');
  const records = [...document.querySelectorAll('li[data-game-title]')].map(node => {
    const d = node.dataset;
    const year = node.closest('.game-year')?.querySelector('h3')?.textContent.trim() || '';
    return { node, title: d.gameTitle, genre: d.gameGenre || 'Unknown',
      status: d.gameStatus, backlog: !!node.closest('[data-game-panel="backlog"]'),
      date: d.gameDate, year, yearNumber: Number(d.gameDate.slice(0, 4)) || Number(d.gameYear) || 0,
      rating: d.gameRating === '' ? null : Number(d.gameRating),
      hours: d.gameHours === '' ? null : Number(d.gameHours) };
  });
  let scope = 'history';
  let genre = null;
  let rating = null;
  const byTitle = (a, b) => a.title.localeCompare(b.title);

  function render() {
    const y = window.scrollY;
    let shown = records.filter(g => g.backlog === (scope === 'backlog') &&
      (genre === null || g.genre === genre) && (rating === null || g.rating === rating));
    const mode = arrange.value;
    shown.sort((a, b) => {
      if (mode === 'rating') return (b.rating ?? -1) - (a.rating ?? -1) || byTitle(a, b);
      if (mode === 'hours') return (b.hours ?? -1) - (a.hours ?? -1) || byTitle(a, b);
      if (scope === 'backlog') return byTitle(a, b);
      return Number(b.status === 'playing') - Number(a.status === 'playing') ||
        b.yearNumber - a.yearNumber || b.date.localeCompare(a.date) || byTitle(a, b);
    });
    // Detach hidden records as well: IDs stay unique and filtered deep links remain available.
    for (const g of records) g.node.remove();
    const fragment = document.createDocumentFragment();
    const groups = new Map();
    for (const g of shown) {
      let label;
      if (mode === 'rating') label = g.rating === null ? 'Unrated' : `${g.rating} ★`;
      else if (mode === 'hours') label = g.hours === null ? 'Unrecorded hours' : 'Recorded hours';
      else label = scope === 'backlog' ? 'Backlog' : g.status === 'playing' ? 'Playing now' : g.year || 'Earlier';
      if (!groups.has(label)) {
        const section = document.createElement('section');
        section.className = 'game-browser-group collection-section';
        const heading = document.createElement('h2');
        heading.textContent = label;
        const shelf = document.createElement('ul');
        shelf.className = 'collection-shelf game-chronology-grid';
        section.append(heading, shelf);
        fragment.append(section);
        groups.set(label, shelf);
      }
      const value = g.node.querySelector('.game-arranged-value');
      value.hidden = mode === 'years';
      value.textContent = mode === 'rating' ? (g.rating === null ? 'Unrated' : `${g.rating} / 5`) :
        (g.hours === null ? 'Unrecorded' : `${g.hours}h`);
      groups.get(label).append(g.node);
    }
    if (!shown.length) {
      const empty = document.createElement('p');
      empty.className = 'game-browser-empty';
      empty.textContent = 'No games match these filters.';
      fragment.append(empty);
    }
    results.replaceChildren(fragment);
    count.textContent = `${shown.length} ${shown.length === 1 ? 'game' : 'games'}`;
    chips.replaceChildren();
    for (const [key, value] of [['genre', genre], ['rating', rating]]) {
      if (value === null) continue;
      const chip = document.createElement('button');
      chip.type = 'button';
      chip.textContent = `${value}${key === 'rating' ? '★' : ''} ×`;
      chip.setAttribute('aria-label', `Remove ${key} filter: ${value}`);
      chip.addEventListener('click', () => {
        if (key === 'genre') genre = null; else rating = null;
        render();
        (chips.querySelector('button') || arrange).focus({ preventScroll: true });
      });
      chips.append(chip);
    }
    clear.hidden = genre === null && rating === null;
    for (const button of controls.querySelectorAll('[data-game-scope]')) {
      const active = button.dataset.gameScope === scope;
      button.classList.toggle('is-active', active);
      button.setAttribute('aria-pressed', String(active));
    }
    for (const button of document.querySelectorAll('[data-game-filter-genre]')) {
      button.setAttribute('aria-pressed', String(button.dataset.gameFilterGenre === genre));
    }
    for (const button of document.querySelectorAll('[data-game-filter-rating]')) {
      button.setAttribute('aria-pressed', String(Number(button.dataset.gameFilterRating) === rating));
    }
    // Retain enough document height for the controls to stay where the user left them.
    results.style.minHeight = `${Math.max(240, window.innerHeight - controls.getBoundingClientRect().height)}px`;
    window.scrollTo({ top: y, behavior: 'instant' });
  }
  for (const button of controls.querySelectorAll('[data-game-scope]')) {
    button.addEventListener('click', () => { scope = button.dataset.gameScope; render(); });
  }
  arrange.addEventListener('change', render);
  clear.addEventListener('click', () => {
    genre = null; rating = null; render(); arrange.focus({ preventScroll: true });
  });
  for (const button of document.querySelectorAll('[data-game-filter-genre], [data-game-filter-rating]')) {
    button.disabled = false;
    button.addEventListener('click', () => {
      if (button.dataset.gameFilterGenre !== undefined) {
        const value = button.dataset.gameFilterGenre;
        genre = genre === value ? null : value;
      } else {
        const value = Number(button.dataset.gameFilterRating);
        rating = rating === value ? null : value;
        arrange.value = 'rating';
      }
      render();
    });
  }
  for (const panel of panels) panel.hidden = true;
  controls.hidden = false;
  results.hidden = false;
  render();
  return {
    find(id) {
      const g = records.find(g => g.node.id === id || g.node.id === id.replace(/^playtime-/, 'game-'));
      if (!g) return null;
      if (g.backlog !== (scope === 'backlog') || !g.node.isConnected) {
        scope = g.backlog ? 'backlog' : 'history'; genre = null; rating = null; render();
      }
      return g.node;
    }
  };
}
const gameBrowser = createGameBrowser();

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
  collectionDialog.classList.toggle('collection-dialog--game', !!entry.querySelector('.game-entry-details'));
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
    const catalogue = details.querySelector('.game-catalogue-line');
    if (catalogue) writing.insertBefore(catalogue, heading);
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
  const link = event.target.closest('a[href^="#game-"], a[href^="#playtime-"]');
  if (link && gameBrowser && !event.ctrlKey && !event.metaKey && !event.shiftKey && !event.altKey) {
    const id = link.getAttribute('href').slice(1);
    const target = gameBrowser.find(id);
    if (target) {
      event.preventDefault();
      history.replaceState(null, '', `#${id}`);
      showCollectionEntry(target.querySelector('.shelf-entry'));
      if (!supportsCollectionDialog) target.scrollIntoView({ block: 'start' });
      return;
    }
  }
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
  const target = gameBrowser?.find(id) || document.getElementById(id);
  const entry = target?.querySelector('.shelf-entry, .game-entry');
  if (!entry) return;
  showCollectionEntry(entry);
  if (!supportsCollectionDialog) requestAnimationFrame(() => target.scrollIntoView({ block: 'start' }));
}

openCollectionTarget();
window.addEventListener('hashchange', openCollectionTarget);
