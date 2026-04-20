/* ── State ────────────────────────────────────────────────────────── */
let allEvents   = [];
let filtered    = [];
let currentPage = 1;
const PAGE_SIZE = 15;
let sortCol = 'time';
let sortDir = 'desc';
let currentWindow = '1h';

let timelineChart = null;
let doughnutChart = null;

/* ── Chart colour palette ─────────────────────────────────────────── */
const PALETTE = [
  '#E4003A','#FF6B6B','#FF8E53','#FFC107',
  '#34C759','#0A84FF','#BF5AF2','#00C7BE',
  '#FF9F0A','#30D158','#64D2FF','#FFD60A',
];

/* ── Session storage keys ─────────────────────────────────────────── */
const SS_KEY       = 'f5xc_api_key';
const SS_NAMESPACE = 'f5xc_namespace';
const SS_LB        = 'f5xc_lb';

/* ── Initialise on load ───────────────────────────────────────────── */
document.addEventListener('DOMContentLoaded', async () => {
  await loadServerConfig();   // seed namespace from server before restoring session
  restoreSettings();          // session values win over server defaults
  initCharts();
  updateKeyStatus();

  if (getApiKey()) {
    refresh();
  } else {
    showSetupPrompt(true);
  }
});

// loadServerConfig fetches /api/config and pre-fills the namespace field
// if sessionStorage doesn't already have a user-set value.
async function loadServerConfig() {
  try {
    const resp = await fetch('/api/config');
    if (!resp.ok) return;
    const cfg = await resp.json();
    // Only apply server default if the user hasn't already saved a preference.
    if (cfg.namespace && !sessionStorage.getItem(SS_NAMESPACE)) {
      document.getElementById('namespace-input').value = cfg.namespace;
    }
  } catch (_) {
    // Non-fatal — fallback to whatever is in the input already.
  }
}

/* ── Connection settings ──────────────────────────────────────────── */
function getApiKey() {
  return document.getElementById('api-key-input').value.trim();
}

function getNamespace() {
  return document.getElementById('namespace-input').value.trim();
}

function getLB() {
  return document.getElementById('lb-input').value.trim();
}

function restoreSettings() {
  const key = sessionStorage.getItem(SS_KEY);
  const ns  = sessionStorage.getItem(SS_NAMESPACE);
  const lb  = sessionStorage.getItem(SS_LB);
  if (key) document.getElementById('api-key-input').value = key;
  if (ns)  document.getElementById('namespace-input').value = ns;
  if (lb)  document.getElementById('lb-input').value = lb;
}

function saveSettings() {
  sessionStorage.setItem(SS_KEY,       getApiKey());
  sessionStorage.setItem(SS_NAMESPACE, getNamespace());
  sessionStorage.setItem(SS_LB,        getLB());
}

function onSettingsChange() {
  saveSettings();
  updateKeyStatus();
}

function updateKeyStatus() {
  const el  = document.getElementById('key-status');
  const key = getApiKey();
  if (key) {
    el.textContent = '✓ API key set';
    el.className   = 'key-status key-ok';
    showSetupPrompt(false);
  } else {
    el.textContent = 'No API key set';
    el.className   = 'key-status key-missing';
    showSetupPrompt(true);
  }
}

function showSetupPrompt(show) {
  document.getElementById('setup-prompt').classList.toggle('hidden', !show);
  document.getElementById('dashboard').classList.toggle('hidden', show);
}

function toggleKeyVisibility() {
  const input = document.getElementById('api-key-input');
  const btn   = document.getElementById('show-hide-btn');
  if (input.type === 'password') {
    input.type = 'text';
    btn.title  = 'Hide key';
    btn.style.color = 'var(--accent)';
  } else {
    input.type = 'password';
    btn.title  = 'Show key';
    btn.style.color = '';
  }
}

/* ── Controls ─────────────────────────────────────────────────────── */
function setWindow(w) {
  currentWindow = w;
  document.getElementById('btn-1h').classList.toggle('active', w === '1h');
  document.getElementById('btn-24h').classList.toggle('active', w === '24h');
  if (getApiKey()) refresh();
}

function refresh() {
  saveSettings();
  if (!getApiKey()) {
    showSetupPrompt(true);
    return;
  }
  fetchEvents(currentWindow, getLB());
}

function exportCSV() {
  const key = getApiKey();
  if (!key) { showError('Enter your API key first.'); return; }

  const lb     = getLB();
  const ns     = getNamespace();
  const params = new URLSearchParams({ window: currentWindow });
  if (lb) params.set('lb', lb);

  // For CSV export we need to send the API key — use a fetch + blob download
  // so the key travels as a header (not in the URL).
  showLoading(true);
  fetch(`/api/export?${params}`, {
    headers: buildHeaders(key, ns),
  })
    .then(r => {
      if (!r.ok) return r.text().then(t => { throw new Error(t); });
      return r.blob();
    })
    .then(blob => {
      const url      = URL.createObjectURL(blob);
      const a        = document.createElement('a');
      const ts       = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
      a.href         = url;
      a.download     = `sec_events_${ts}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    })
    .catch(err => showError(err.message))
    .finally(() => showLoading(false));
}

function dismissError() {
  document.getElementById('error-banner').classList.add('hidden');
}

/* ── HTTP helpers ─────────────────────────────────────────────────── */
function buildHeaders(apiKey, namespace) {
  return {
    'X-Api-Key':   apiKey,
    'X-Namespace': namespace,
  };
}

/* ── Fetch ────────────────────────────────────────────────────────── */
async function fetchEvents(window, lb) {
  showLoading(true);
  dismissError();

  const key = getApiKey();
  const ns  = getNamespace();

  try {
    const params = new URLSearchParams({ window });
    if (lb) params.set('lb', lb);

    const resp = await fetch(`/api/events?${params}`, {
      headers: buildHeaders(key, ns),
    });

    if (resp.status === 401) {
      const body = await resp.json().catch(() => ({}));
      throw new Error(body.error || 'API key rejected — check the key and try again.');
    }
    if (!resp.ok) {
      const msg = await resp.text();
      throw new Error(`Server error ${resp.status}: ${msg}`);
    }

    const data = await resp.json();
    allEvents = Array.isArray(data) ? data : [];
    currentPage = 1;
    renderAll();
  } catch (err) {
    showError(err.message);
  } finally {
    showLoading(false);
  }
}

/* ── Render all components ────────────────────────────────────────── */
function renderAll() {
  applySort();
  renderStats();
  renderTimeline();
  renderDoughnut();
  renderTable();
}

/* ── Stats bar ────────────────────────────────────────────────────── */
function renderStats() {
  const total   = allEvents.length;
  const blocked = allEvents.filter(e => (e.waf_action || '').toUpperCase() === 'BLOCK').length;
  const allowed = allEvents.filter(e => (e.waf_action || '').toUpperCase() === 'ALLOW').length;
  const topAtk  = topAttackType(allEvents);

  document.getElementById('stat-total').textContent    = total.toLocaleString();
  document.getElementById('stat-blocked').textContent  = blocked.toLocaleString();
  document.getElementById('stat-allowed').textContent  = allowed.toLocaleString();
  document.getElementById('stat-top-attack').textContent = topAtk || '—';
}

function topAttackType(events) {
  const counts = {};
  events.forEach(e => {
    const t = e.attack_type;
    if (t) counts[t] = (counts[t] || 0) + 1;
  });
  const entries = Object.entries(counts);
  if (!entries.length) return null;
  entries.sort((a, b) => b[1] - a[1]);
  return entries[0][0];
}

/* ── Timeline chart ───────────────────────────────────────────────── */
function initCharts() {
  Chart.defaults.color        = '#7a7f8a';
  Chart.defaults.borderColor  = '#2e3138';
  Chart.defaults.font.family  = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';

  const timelineCtx = document.getElementById('timeline-chart').getContext('2d');
  timelineChart = new Chart(timelineCtx, {
    type: 'bar',
    data: {
      labels: [],
      datasets: [{
        label: 'Events',
        data: [],
        backgroundColor: 'rgba(228,0,58,0.7)',
        borderColor: '#E4003A',
        borderWidth: 1,
        borderRadius: 3,
      }],
    },
    options: {
      responsive: true,
      maintainAspectRatio: true,
      plugins: { legend: { display: false } },
      scales: {
        x: { grid: { color: '#1e2025' }, ticks: { maxRotation: 45, maxTicksLimit: 16 } },
        y: { grid: { color: '#1e2025' }, beginAtZero: true, ticks: { precision: 0 } },
      },
    },
  });

  const doughnutCtx = document.getElementById('doughnut-chart').getContext('2d');
  doughnutChart = new Chart(doughnutCtx, {
    type: 'doughnut',
    data: {
      labels: [],
      datasets: [{
        data: [],
        backgroundColor: PALETTE,
        borderColor: '#1a1c1f',
        borderWidth: 2,
      }],
    },
    options: {
      responsive: true,
      maintainAspectRatio: true,
      plugins: {
        legend: { position: 'bottom', labels: { padding: 12, boxWidth: 12, font: { size: 11 } } },
      },
      cutout: '62%',
    },
  });
}

function renderTimeline() {
  const BUCKET_MS = 5 * 60 * 1000;

  if (!allEvents.length) {
    timelineChart.data.labels = [];
    timelineChart.data.datasets[0].data = [];
    timelineChart.update();
    return;
  }

  const times = allEvents
    .map(e => new Date(e.time).getTime())
    .filter(t => !isNaN(t));

  if (!times.length) { timelineChart.update(); return; }

  const minT = Math.floor(Math.min(...times) / BUCKET_MS) * BUCKET_MS;
  const maxT = Math.ceil(Math.max(...times)  / BUCKET_MS) * BUCKET_MS;

  const buckets = {};
  for (let t = minT; t <= maxT; t += BUCKET_MS) buckets[t] = 0;

  allEvents.forEach(e => {
    const t = new Date(e.time).getTime();
    if (!isNaN(t)) {
      const b = Math.floor(t / BUCKET_MS) * BUCKET_MS;
      buckets[b] = (buckets[b] || 0) + 1;
    }
  });

  const keys   = Object.keys(buckets).map(Number).sort((a, b) => a - b);
  const labels = keys.map(k => new Date(k).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }));
  const data   = keys.map(k => buckets[k]);

  timelineChart.data.labels = labels;
  timelineChart.data.datasets[0].data = data;
  timelineChart.update();
}

/* ── Doughnut chart ───────────────────────────────────────────────── */
function renderDoughnut() {
  const counts = {};
  allEvents.forEach(e => {
    const t = e.attack_type || 'Unknown';
    counts[t] = (counts[t] || 0) + 1;
  });

  const entries = Object.entries(counts).sort((a, b) => b[1] - a[1]);
  doughnutChart.data.labels                       = entries.map(e => e[0]);
  doughnutChart.data.datasets[0].data             = entries.map(e => e[1]);
  doughnutChart.data.datasets[0].backgroundColor  = entries.map((_, i) => PALETTE[i % PALETTE.length]);
  doughnutChart.update();
}

/* ── Sort ─────────────────────────────────────────────────────────── */
function sortBy(col) {
  if (sortCol === col) {
    sortDir = sortDir === 'asc' ? 'desc' : 'asc';
  } else {
    sortCol = col;
    sortDir = 'asc';
  }
  currentPage = 1;
  applySort();
  renderTable();
  updateSortHeaders();
}

function applySort() {
  filtered = [...allEvents].sort((a, b) => {
    let av = a[sortCol] ?? '';
    let bv = b[sortCol] ?? '';
    if (typeof av === 'number' && typeof bv === 'number') {
      return sortDir === 'asc' ? av - bv : bv - av;
    }
    av = String(av).toLowerCase();
    bv = String(bv).toLowerCase();
    if (av < bv) return sortDir === 'asc' ? -1 :  1;
    if (av > bv) return sortDir === 'asc' ?  1 : -1;
    return 0;
  });
}

function updateSortHeaders() {
  document.querySelectorAll('thead th').forEach(th => {
    th.classList.remove('sort-asc', 'sort-desc');
    if (th.dataset.col === sortCol) {
      th.classList.add(sortDir === 'asc' ? 'sort-asc' : 'sort-desc');
    }
  });
}

/* ── Table ────────────────────────────────────────────────────────── */
function renderTable() {
  const tbody  = document.getElementById('events-tbody');
  const infoEl = document.getElementById('table-info');
  const total  = filtered.length;
  const pages  = Math.max(1, Math.ceil(total / PAGE_SIZE));

  if (currentPage > pages) currentPage = pages;

  const start = (currentPage - 1) * PAGE_SIZE;
  const slice = filtered.slice(start, start + PAGE_SIZE);

  infoEl.textContent = total
    ? `${start + 1}–${start + slice.length} of ${total.toLocaleString()}`
    : '';

  if (!slice.length) {
    tbody.innerHTML = `<tr><td colspan="8"><div class="empty-state">No events found. Adjust the time window or load balancer filter and refresh.</div></td></tr>`;
    renderPagination(0, 1);
    return;
  }

  tbody.innerHTML = slice.map(e => {
    const action   = (e.waf_action || '').toUpperCase();
    const rowClass = action === 'BLOCK' ? 'row-block' : action === 'ALLOW' ? 'row-allow' : '';
    const badgeCls = action === 'BLOCK' ? 'badge-block' : action === 'ALLOW' ? 'badge-allow' : 'badge-other';
    const sev      = (e.severity || '').toLowerCase();
    const sevCls   = sev === 'critical' ? 'sev-critical' : sev === 'high' ? 'sev-high' : sev === 'medium' ? 'sev-medium' : 'sev-low';

    return `<tr class="${rowClass}">
      <td title="${escHtml(e.time)}">${formatTime(e.time)}</td>
      <td>${escHtml(e.src_ip   || '—')}</td>
      <td>${escHtml(e.method   || '—')}</td>
      <td title="${escHtml(e.req_path || '')}">${escHtml(e.req_path || '—')}</td>
      <td>${e.response_code || '—'}</td>
      <td><span class="badge ${badgeCls}">${action || '—'}</span></td>
      <td title="${escHtml(e.attack_type || '')}">${escHtml(e.attack_type || '—')}</td>
      <td class="${sevCls}">${escHtml(e.severity || '—')}</td>
    </tr>`;
  }).join('');

  renderPagination(total, pages);
  updateSortHeaders();
}

function renderPagination(total, pages) {
  const el = document.getElementById('pagination');
  if (total <= PAGE_SIZE) { el.innerHTML = ''; return; }

  el.innerHTML = `
    <button class="page-btn" onclick="goPage(${currentPage - 1})" ${currentPage <= 1 ? 'disabled' : ''}>← Prev</button>
    <span class="page-info">Page ${currentPage} of ${pages}</span>
    <button class="page-btn" onclick="goPage(${currentPage + 1})" ${currentPage >= pages ? 'disabled' : ''}>Next →</button>
  `;
}

function goPage(n) {
  const pages = Math.ceil(filtered.length / PAGE_SIZE);
  currentPage = Math.max(1, Math.min(n, pages));
  renderTable();
}

/* ── Helpers ──────────────────────────────────────────────────────── */
function formatTime(raw) {
  if (!raw) return '—';
  const d = new Date(raw);
  if (isNaN(d)) return escHtml(raw);
  return d.toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

function escHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function showLoading(on) {
  document.getElementById('loading').classList.toggle('hidden', !on);
}

function showError(msg) {
  document.getElementById('error-text').textContent = msg;
  document.getElementById('error-banner').classList.remove('hidden');
}
