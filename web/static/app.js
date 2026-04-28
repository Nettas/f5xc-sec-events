/* ── State ────────────────────────────────────────────────────────── */
let allEvents   = [];
let filtered    = [];
let currentPage = 1;
const PAGE_SIZE = 15;
let sortCol = 'time';
let sortDir = 'desc';
let currentWindow = 1;
let expandedIdx = null;   // index in filtered[] of the currently-expanded row

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
const SS_TENANT    = 'f5xc_tenant';
const SS_NAMESPACE = 'f5xc_namespace';
const SS_LB        = 'f5xc_lb';

/* ── Initialise on load ───────────────────────────────────────────── */
document.addEventListener('DOMContentLoaded', async () => {
  await loadServerConfig();
  restoreSettings();
  initCharts();
  updateKeyStatus();

  if (getApiKey()) {
    refresh();
  } else {
    showSetupPrompt(true);
  }
});

async function loadServerConfig() {
  try {
    const resp = await fetch('/api/config');
    if (!resp.ok) return;
    const cfg = await resp.json();
    if (cfg.tenant && !sessionStorage.getItem(SS_TENANT)) {
      document.getElementById('tenant-input').value = cfg.tenant;
    }
    if (cfg.namespace && !sessionStorage.getItem(SS_NAMESPACE)) {
      document.getElementById('namespace-input').value = cfg.namespace;
    }
  } catch (_) {}
}

/* ── Connection settings ──────────────────────────────────────────── */
function getApiKey()   { return document.getElementById('api-key-input').value.trim(); }
function getTenant()   { return document.getElementById('tenant-input').value.trim(); }
function getNamespace(){ return document.getElementById('namespace-input').value.trim(); }
function getLB()       { return document.getElementById('lb-input').value.trim(); }

function restoreSettings() {
  const key    = sessionStorage.getItem(SS_KEY);
  const tenant = sessionStorage.getItem(SS_TENANT);
  const ns     = sessionStorage.getItem(SS_NAMESPACE);
  const lb     = sessionStorage.getItem(SS_LB);
  if (key)    document.getElementById('api-key-input').value = key;
  if (tenant) document.getElementById('tenant-input').value = tenant;
  if (ns)     document.getElementById('namespace-input').value = ns;
  if (lb)     document.getElementById('lb-input').value = lb;
}

function saveSettings() {
  sessionStorage.setItem(SS_KEY,       getApiKey());
  sessionStorage.setItem(SS_TENANT,    getTenant());
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
function onWindowSlider(val) {
  currentWindow = parseInt(val, 10);
  const label = currentWindow === 1 ? 'Last 1 hour' : `Last ${currentWindow} hours`;
  document.getElementById('window-label').textContent = label;
  if (getApiKey()) refresh();
}

function refresh() {
  saveSettings();
  if (!getApiKey()) { showSetupPrompt(true); return; }
  fetchEvents(currentWindow, getLB());
}

function exportCSV() {
  const key = getApiKey();
  if (!key) { showError('Enter your API key first.'); return; }

  const lb     = getLB();
  const tenant = getTenant();
  const ns     = getNamespace();
  const params = new URLSearchParams({ window: currentWindow });
  if (lb) params.set('lb', lb);

  showLoading(true);
  fetch(`/api/export?${params}`, { headers: buildHeaders(key, tenant, ns) })
    .then(r => {
      if (!r.ok) return r.text().then(t => { throw new Error(t); });
      return r.blob();
    })
    .then(blob => {
      const url = URL.createObjectURL(blob);
      const a   = document.createElement('a');
      const ts  = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
      a.href     = url;
      a.download = `sec_events_${ts}.csv`;
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
function buildHeaders(apiKey, tenant, namespace) {
  return { 'X-Api-Key': apiKey, 'X-Tenant': tenant, 'X-Namespace': namespace };
}

/* ── Fetch ────────────────────────────────────────────────────────── */
async function fetchEvents(window, lb) {
  showLoading(true);
  dismissError();

  const key    = getApiKey();
  const tenant = getTenant();
  const ns     = getNamespace();

  try {
    const params = new URLSearchParams({ window });
    if (lb) params.set('lb', lb);

    const resp = await fetch(`/api/events?${params}`, { headers: buildHeaders(key, tenant, ns) });

    if (resp.status === 401) {
      const body = await resp.json().catch(() => ({}));
      throw new Error(body.error || 'API key rejected — check the key and try again.');
    }
    if (!resp.ok) {
      const msg = await resp.text();
      throw new Error(`Server error ${resp.status}: ${msg}`);
    }

    const data = await resp.json();
    allEvents   = Array.isArray(data) ? data : [];
    currentPage = 1;
    expandedIdx = null;
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
  const blocked = allEvents.filter(e => (e.action || '').toUpperCase() === 'BLOCK').length;
  const allowed = allEvents.filter(e => (e.action || '').toUpperCase() === 'ALLOW').length;
  const topType = topEventType(allEvents);

  document.getElementById('stat-total').textContent    = total.toLocaleString();
  document.getElementById('stat-blocked').textContent  = blocked.toLocaleString();
  document.getElementById('stat-allowed').textContent  = allowed.toLocaleString();
  document.getElementById('stat-top-attack').textContent = topType || '—';
}

function topEventType(events) {
  const counts = {};
  events.forEach(e => {
    const t = e.sec_event_type;
    if (t) counts[t] = (counts[t] || 0) + 1;
  });
  const entries = Object.entries(counts);
  if (!entries.length) return null;
  entries.sort((a, b) => b[1] - a[1]);
  return entries[0][0];
}

/* ── Timeline chart ───────────────────────────────────────────────── */
function initCharts() {
  Chart.defaults.color       = '#7a7f8a';
  Chart.defaults.borderColor = '#2e3138';
  Chart.defaults.font.family = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';

  const timelineCtx = document.getElementById('timeline-chart').getContext('2d');
  timelineChart = new Chart(timelineCtx, {
    type: 'bar',
    data: { labels: [], datasets: [{ label: 'Events', data: [],
      backgroundColor: 'rgba(228,0,58,0.7)', borderColor: '#E4003A',
      borderWidth: 1, borderRadius: 3 }] },
    options: {
      responsive: true, maintainAspectRatio: true,
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
    data: { labels: [], datasets: [{ data: [], backgroundColor: PALETTE,
      borderColor: '#1a1c1f', borderWidth: 2 }] },
    options: {
      responsive: true, maintainAspectRatio: true,
      plugins: { legend: { position: 'bottom', labels: { padding: 12, boxWidth: 12, font: { size: 11 } } } },
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
  const times = allEvents.map(e => new Date(e.time).getTime()).filter(t => !isNaN(t));
  if (!times.length) { timelineChart.update(); return; }

  const minT = Math.floor(Math.min(...times) / BUCKET_MS) * BUCKET_MS;
  const maxT = Math.ceil(Math.max(...times)  / BUCKET_MS) * BUCKET_MS;
  const buckets = {};
  for (let t = minT; t <= maxT; t += BUCKET_MS) buckets[t] = 0;
  allEvents.forEach(e => {
    const t = new Date(e.time).getTime();
    if (!isNaN(t)) { const b = Math.floor(t / BUCKET_MS) * BUCKET_MS; buckets[b] = (buckets[b] || 0) + 1; }
  });
  const keys   = Object.keys(buckets).map(Number).sort((a, b) => a - b);
  timelineChart.data.labels = keys.map(k => new Date(k).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }));
  timelineChart.data.datasets[0].data = keys.map(k => buckets[k]);
  timelineChart.update();
}

/* ── Doughnut chart ───────────────────────────────────────────────── */
function renderDoughnut() {
  const counts = {};
  allEvents.forEach(e => {
    const t = e.sec_event_type || 'Unknown';
    counts[t] = (counts[t] || 0) + 1;
  });
  const entries = Object.entries(counts).sort((a, b) => b[1] - a[1]);
  doughnutChart.data.labels                      = entries.map(e => e[0]);
  doughnutChart.data.datasets[0].data            = entries.map(e => e[1]);
  doughnutChart.data.datasets[0].backgroundColor = entries.map((_, i) => PALETTE[i % PALETTE.length]);
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
  expandedIdx = null;
  applySort();
  renderTable();
  updateSortHeaders();
}

function applySort() {
  filtered = [...allEvents].sort((a, b) => {
    let av = a[sortCol] ?? '';
    let bv = b[sortCol] ?? '';
    if (typeof av === 'number' && typeof bv === 'number') return sortDir === 'asc' ? av - bv : bv - av;
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
    if (th.dataset.col === sortCol) th.classList.add(sortDir === 'asc' ? 'sort-asc' : 'sort-desc');
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
    tbody.innerHTML = `<tr><td colspan="10"><div class="empty-state">No events found. Adjust the time window or load balancer filter and refresh.</div></td></tr>`;
    renderPagination(0, 1);
    return;
  }

  let html = '';
  slice.forEach((e, i) => {
    const globalIdx = start + i;
    const isExpanded = expandedIdx === globalIdx;
    const action   = (e.action || '').toUpperCase();
    const rowClass = action === 'BLOCK' ? 'row-block' : action === 'ALLOW' ? 'row-allow' : '';
    const badgeCls = action === 'BLOCK' ? 'badge-block' : action === 'ALLOW' ? 'badge-allow' : 'badge-other';
    const domain   = e.authority || e.vh_name || '';
    const evType   = (e.sec_event_type || '').replace(/_/g, ' ');

    html += `<tr class="data-row ${rowClass}${isExpanded ? ' row-expanded' : ''}" onclick="toggleRow(${globalIdx})">
      <td title="${escHtml(e.time)}">${isExpanded ? '▼ ' : '▶ '}${formatTime(e.time)}</td>
      <td>${escHtml(e.country   || '—')}</td>
      <td>${escHtml(e.city      || '—')}</td>
      <td>${escHtml(e.src_ip    || '—')}</td>
      <td>${escHtml(e.method    || '—')}</td>
      <td>${e.rsp_code || '—'}</td>
      <td title="${escHtml(e.sec_event_type || '')}">${escHtml(evType || '—')}</td>
      <td><span class="badge ${badgeCls}">${action || '—'}</span></td>
      <td title="${escHtml(domain)}">${escHtml(domain || '—')}</td>
      <td title="${escHtml(e.req_path || '')}">${escHtml(e.req_path || '—')}</td>
    </tr>`;

    if (isExpanded) {
      html += `<tr class="detail-row" onclick="event.stopPropagation()">
        <td colspan="10">${buildDetailPanel(e)}</td>
      </tr>`;
    }
  });

  tbody.innerHTML = html;
  renderPagination(total, pages);
  updateSortHeaders();
}

/* ── Row expand / collapse ────────────────────────────────────────── */
function toggleRow(idx) {
  expandedIdx = (expandedIdx === idx) ? null : idx;
  renderTable();
}

/* ── Detail panel ─────────────────────────────────────────────────── */
function buildDetailPanel(e) {
  const sigs    = parseJsonField(e.signatures);
  const reasons = parseJsonField(e.req_risk_reasons);
  const firstReason = Array.isArray(reasons) && reasons.length
    ? (typeof reasons[0] === 'string' ? reasons[0] : JSON.stringify(reasons[0]))
    : null;

  function row(label, val) {
    const display = (val === null || val === undefined || val === '') ? '—' : escHtml(String(val));
    return `<div class="detail-label">${escHtml(label)}</div><div class="detail-value">${display}</div>`;
  }

  const srcSection = `
    <div class="detail-section">
      <div class="detail-section-title">Source</div>
      <div class="detail-grid">
        ${row('IP',          e.src_ip)}
        ${row('City',        e.city)}
        ${row('Region',      e.region)}
        ${row('Country',     e.country)}
        ${row('ASN',         e.asn)}
        ${row('Browser',     e.browser_type)}
        ${row('Device',      e.device_type)}
        ${row('User Agent',  e.user_agent)}
        ${row('Src Site',    e.src_site)}
        ${row('Src',         e.src)}
        ${row('TLS FP',      e.tls_fingerprint)}
        ${row('JA4 TLS FP',  e.ja4_tls_fingerprint)}
      </div>
    </div>`;

  const reqSection = `
    <div class="detail-section">
      <div class="detail-section-title">Request</div>
      <div class="detail-grid">
        ${row('Req ID',       e.req_id)}
        ${row('Authority',    e.authority)}
        ${row('Path',         e.req_path)}
        ${row('Method',       e.method)}
        ${row('API Endpoint', e.api_endpoint)}
        ${row('Req Size',     e.req_size || e.req_size === 0 ? e.req_size : null)}
        ${row('Rsp Size',     e.rsp_size || e.rsp_size === 0 ? e.rsp_size : null)}
        ${row('Rsp Code',     e.rsp_code || e.rsp_code === 0 ? e.rsp_code : null)}
      </div>
    </div>`;

  const detSection = `
    <div class="detail-section">
      <div class="detail-section-title">Detection</div>
      <div class="detail-grid">
        ${row('Event Type',   e.sec_event_type)}
        ${row('Action',       e.action)}
        ${row('Risk',         e.req_risk)}
        ${row('Risk Reasons', firstReason)}
      </div>
    </div>`;

  let sigsHtml = '';
  if (Array.isArray(sigs) && sigs.length) {
    sigsHtml = sigs.map(sig => {
      const atkTypes = Array.isArray(sig.attack_types)
        ? sig.attack_types.map(a => escHtml(a.name || String(a))).join(', ')
        : escHtml(String(sig.attack_types || '—'));
      const matchInfo = Array.isArray(sig.matching_info)
        ? sig.matching_info.map(m => escHtml(typeof m === 'string' ? m : JSON.stringify(m))).join('; ')
        : escHtml(String(sig.matching_info || '—'));
      const sigState  = (sig.state || '').toLowerCase();
      const sigBadge  = sigState === 'active' ? 'badge-block' : 'badge-other';
      return `<div class="sig-card">
        <div class="sig-header">
          <span class="sig-name">${escHtml(String(sig.name || sig.id || 'Signature'))}</span>
          <span class="badge ${sigBadge}">${escHtml(sig.state || '—')}</span>
        </div>
        <div class="detail-grid">
          <div class="detail-label">ID</div><div class="detail-value">${escHtml(String(sig.id || '—'))}</div>
          <div class="detail-label">Attack Type</div><div class="detail-value">${atkTypes}</div>
          <div class="detail-label">Accuracy</div><div class="detail-value">${escHtml(sig.accuracy || '—')}</div>
          <div class="detail-label">Context</div><div class="detail-value">${escHtml(sig.context || '—')}</div>
          <div class="detail-label">Match Info</div><div class="detail-value">${matchInfo}</div>
          <div class="detail-label">Risk</div><div class="detail-value">${escHtml(sig.risk || '—')}</div>
        </div>
      </div>`;
    }).join('');
  } else {
    sigsHtml = '<div class="detail-empty">No signatures in this event</div>';
  }

  const sigSection = `
    <div class="detail-section detail-signatures">
      <div class="detail-section-title">Signatures</div>
      ${sigsHtml}
    </div>`;

  return `<div class="detail-panel">${srcSection}${reqSection}${detSection}${sigSection}</div>`;
}

/* ── JSON field helper (handles both raw objects and double-encoded strings) */
function parseJsonField(val) {
  if (!val) return null;
  if (typeof val === 'string') {
    try { return JSON.parse(val); } catch { return null; }
  }
  return val;
}

/* ── Pagination ───────────────────────────────────────────────────── */
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
  expandedIdx = null;
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
