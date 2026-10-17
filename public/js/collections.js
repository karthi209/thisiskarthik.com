// Native disclosures still work without JavaScript.
for (const entry of document.querySelectorAll('.shelf-entry, .game-entry')) {
  entry.addEventListener('keydown', event => {
    if (event.key !== 'Escape' || !entry.open) return;
    event.preventDefault();
    entry.open = false;
    entry.querySelector('summary').focus({ preventScroll: true });
  });
}

function openCollectionTarget() {
  let id;
  try { id = decodeURIComponent(location.hash.slice(1)); } catch { return; }
  const target = document.getElementById(id);
  const entry = target?.querySelector('.shelf-entry, .game-entry');
  if (!entry) return;
  entry.open = true;
  // Wait for the expanded shelf and selected game view to finish layout.
  requestAnimationFrame(() => target.scrollIntoView({ block: 'start' }));
}

openCollectionTarget();
window.addEventListener('hashchange', openCollectionTarget);
