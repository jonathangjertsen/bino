function filterDashboard() {
  const q = document.getElementById('dashboard-search').value.trim().toLowerCase();
  const boxes = document.querySelectorAll('.filter-box');
  const container = document.querySelector('.search-container');
  const searchBox = document.getElementById('dashboard-search');
  const noneMsg = document.getElementById('filter-none');
  const active = q.length > 0;
  [container, searchBox].forEach(el => el.classList.toggle('active-search', active));
  if (!active) {
    boxes.forEach(b => b.hidden = false);
    noneMsg.hidden = true;
    return;
  }
  let anyMatch = false;
  boxes.forEach(box => {
    const items = box.querySelectorAll('.filter-content');
    let match = false;
    for (const el of items) {
      if (el.textContent.toLowerCase().includes(q)) { match = true; break; }
    }
    box.hidden = !match;
    if (match) anyMatch = true;
  });
  noneMsg.hidden = anyMatch;
}
document.addEventListener('input', e => {
  if (e.target && e.target.id === 'dashboard-search') filterDashboard();
});
document.addEventListener('DOMContentLoaded', filterDashboard);
