(function () {
  const POLL_MS = 5000;
  let currentMerchantAddress = '';
  let currentList = [];

  const el = {
    status: document.getElementById('status'),
    themeSwitch: document.getElementById('theme-switch'),
    searchInput: document.getElementById('search-input'),
    inflow: document.getElementById('kpi-inflow'),
    outflow: document.getElementById('kpi-outflow'),
    transactionList: document.getElementById('transaction-list'),
    transactionEmpty: document.getElementById('transaction-empty'),
    tokenList: document.getElementById('token-list'),
    tokenEmpty: document.getElementById('token-empty'),
    ongoingList: document.getElementById('ongoing-list'),
    ongoingEmpty: document.getElementById('ongoing-empty'),
    refreshBtn: document.getElementById('refresh-btn'),
    modal: document.getElementById('modal'),
    modalBackdrop: document.getElementById('modal-backdrop'),
    modalClose: document.getElementById('modal-close'),
    modalBody: document.getElementById('modal-body'),
    merchantInfo: document.getElementById('merchant-info'),
    merchantAddressDisplay: document.getElementById('merchant-address-display'),
    logoutBtn: document.getElementById('logout-btn'),
    mainContent: document.getElementById('main-content'),
    noDataMessage: document.getElementById('no-data-message'),
    totalTransactions: document.getElementById('total-transactions'),
    totalReceived: document.getElementById('total-received'),
    lastActivity: document.getElementById('last-activity')
  };

  function formatMoney(num) {
    if (!isFinite(num)) return '$0';
    return '$' + Number(num).toLocaleString(undefined, { maximumFractionDigits: 2 });
  }

  function getSearchFiltered(list) {
    const q = (el.searchInput && el.searchInput.value || '').trim().toLowerCase();
    if (!q) return list;
    return list.filter(tx => String(tx.Payer || '').toLowerCase().includes(q));
  }

  function hueFromString(str) {
    let h = 0;
    for (let i = 0; i < str.length; i++) h = (h * 31 + str.charCodeAt(i)) >>> 0;
    return h % 360;
  }

  function tokenPillHTML(symbol, chain) {
    const s = (symbol || 'TOKEN').toUpperCase();
    const symClass = s === 'USDC' ? 'usdc' : (s === 'USDT' ? 'usdt' : '');
    const c = chain ? ` title="${chain}"` : '';
    return `<span class="pill ${symClass}"${c}><span class="mono">${s}</span></span>`;
  }

  function parseList(data) {
    if (Array.isArray(data)) return data;
    if (data && Array.isArray(data.list)) return data.list;
    return [];
  }

  async function fetchWithTimeout(url, { timeout = 6000 } = {}) {
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), timeout);
    try {
      const res = await fetch(url, { signal: controller.signal, headers: { 'Accept': 'application/json' } });
      if (!res.ok) throw new Error('HTTP ' + res.status);
      const data = await res.json();
      return data;
    } finally {
      clearTimeout(id);
    }
  }

  function aggregateTotals(list) {
    let gross = 0;
    let net = 0;
    for (const tx of list) {
      gross += Number(tx.GrossAmount || 0);
      net += Number(tx.NetAmount || 0);
    }
    return { gross, net };
  }

  function groupBy(list, key, amountField) {
    const map = new Map();
    for (const tx of list) {
      const id = tx[key] || '未知';
      const amt = Number(tx[amountField] || 0);
      map.set(id, (map.get(id) || 0) + amt);
    }
    const arr = Array.from(map, ([name, value]) => ({ name, value }));
    arr.sort((a, b) => b.value - a.value);
    return arr;
  }

  function renderSimpleList(container, emptyEl, rows, valuePrefix = '$') {
    container.innerHTML = '';
    if (!rows.length) {
      emptyEl.hidden = false;
      return;
    }
    emptyEl.hidden = true;
    for (const row of rows) {
      const div = document.createElement('div');
      div.className = 'row';
      div.innerHTML = `
        <div class="name" title="${row.name}">${row.name}</div>
        <div class="value">${valuePrefix}${row.value.toLocaleString()}</div>
      `;
      container.appendChild(div);
    }
  }

  // Token recognition map (extend as needed)
  const TOKEN_MAP = {
    // Base mainnet
    '0x833589fcd6edb6e08f4c7c32d4f71b54b7cfb66e': { symbol: 'USDC', chain: 'Base' },
    '0xd9aaec6eab5f9f0a7f0dd7c39c3f1b3aa1c5f6b9': { symbol: 'USDbC', chain: 'Base' },
    // Base Sepolia
    '0x75faf114eafb1bdbe2f0316df893fd58ce46aa4d': { symbol: 'USDC', chain: 'Base Sepolia' },
    '0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673': { symbol: 'USDT', chain: 'Base Sepolia' },
    // Arbitrum Sepolia
    '0x9aa7fEc87CA69695Dd1f879567CcF49F3ba417E2': { symbol: 'USDT', chain: 'Arb Sepolia' },
    '0x0f3a3d8e7c8b1e3b5f0b3d3c1a9f4f0a9e3b1c2d': { symbol: 'USDC', chain: 'Arb Sepolia' },
    '0xdAC17F958D2ee523a2206206994597C13D831ec7': { symbol: 'USDT', chain: 'Arb Sepolia' },
  };

  function getTokenInfo(addr) {
    if (!addr) return null;
    const k = String(addr).toLowerCase();
    return TOKEN_MAP[k] || null;
  }

  function timeAgo(iso) {
    const t = Date.parse(iso);
    if (isNaN(t)) return '-';
    const s = Math.floor((Date.now() - t) / 1000);
    if (s < 60) return `${s}s ago`;
    const m = Math.floor(s / 60);
    if (m < 60) return `${m}m ago`;
    const h = Math.floor(m / 60);
    if (h < 24) return `${h}h ago`;
    const d = Math.floor(h / 24);
    return `${d} days ago`;
  }

  function short(v) {
    if (!v) return '-';
    const s = String(v);
    if (s.length <= 12) return s;
    return s.slice(0, 6) + '…' + s.slice(-4);
  }

  function renderRecentTransactions(list) {
    el.transactionList.innerHTML = '';
    
    if (!list.length) {
      el.transactionEmpty.hidden = false;
      return;
    }
    el.transactionEmpty.hidden = true;

    // Show only the 5 most recent transactions
    const recent = [...list].sort((a, b) => (b.__ts || 0) - (a.__ts || 0)).slice(0, 5);
    
    for (const tx of recent) {
      const token = getTokenInfo(tx.DstToken) || getTokenInfo(tx.SrcToken);
      const tokenHTML = token ? tokenPillHTML(token.symbol, token.chain) : '<span class="badge badge-default">N/A</span>';
      const hue = hueFromString(String(tx.Payer || ''));
      
      const div = document.createElement('div');
      div.className = 'row';
      div.innerHTML = `
        <div class="name" title="${tx.Payer}">
          <span class="avatar" style="--h:${hue}"></span>
          <span class="mono">${short(tx.Payer)}</span>
        </div>
        <div class="value">
          ${formatMoney(Number(tx.NetAmount || 0))}
          ${tokenHTML}
        </div>
      `;
      div.addEventListener('click', () => openModal(tx));
      el.transactionList.appendChild(div);
    }
  }

  function renderTokenSummary(list) {
    const tokenMap = new Map();
    
    for (const tx of list) {
      const token = getTokenInfo(tx.DstToken) || getTokenInfo(tx.SrcToken);
      if (token) {
        const key = `${token.symbol}-${token.chain}`;
        const current = tokenMap.get(key) || { symbol: token.symbol, chain: token.chain, amount: 0, count: 0 };
        current.amount += Number(tx.NetAmount || 0);
        current.count += 1;
        tokenMap.set(key, current);
      }
    }
    
    const tokenSummary = Array.from(tokenMap.values()).sort((a, b) => b.amount - a.amount);
    
    el.tokenList.innerHTML = '';
    if (!tokenSummary.length) {
      el.tokenEmpty.hidden = false;
      return;
    }
    el.tokenEmpty.hidden = true;
    
    for (const token of tokenSummary) {
      const div = document.createElement('div');
      div.className = 'row';
      div.innerHTML = `
        <div class="name">
          ${tokenPillHTML(token.symbol, token.chain)}
          <span style="margin-left: 8px; font-size: 11px; color: var(--muted);">${token.count} tx</span>
        </div>
        <div class="value">${formatMoney(token.amount)}</div>
      `;
      el.tokenList.appendChild(div);
    }
  }

  function renderOngoing(list) {
    el.ongoingList.innerHTML = '';

    const head = document.createElement('div');
    head.className = 'tx-row tx-head';
    head.innerHTML = `
      <div class="tx-cell bold">Payer</div>
      <div class="tx-cell bold">Time</div>
      <div class="tx-cell bold">Value</div>
      <div class="tx-cell bold">Status</div>
      <div class="tx-cell bold hide-md">Token</div>
      <div class="tx-cell bold hide-md">Activity</div>
    `;
    el.ongoingList.appendChild(head);

    if (!list.length) {
      el.ongoingEmpty.hidden = false;
      return;
    }
    el.ongoingEmpty.hidden = true;

    for (const tx of list) {
      const token = getTokenInfo(tx.DstToken) || getTokenInfo(tx.SrcToken);
      const tokensHTML = token ? `<div class="tokens">${tokenPillHTML(token.symbol, token.chain)}</div>` : `<span class="badge badge-default">N/A</span>`;
      const activity = `<span class="icon inflow"></span>Received ${formatMoney(Number(tx.NetAmount || 0)).replace('$','')} ${token ? token.symbol : ''} from ${short(tx.Payer)}`;
      const hue = hueFromString(String(tx.Payer || ''));
      const row = document.createElement('div');
      row.className = 'tx-row';
      row.innerHTML = `
        <div class="tx-cell" title="${tx.Payer}"><span class="avatar" style="--h:${hue}"></span><span class="mono">${short(tx.Payer)}</span></div>
        <div class="tx-cell" title="${tx.Timestamp}">${timeAgo(tx.Timestamp)}</div>
        <div class="tx-cell">${formatMoney(Number(tx.NetAmount || 0))}</div>
        <div class="tx-cell">${statusBadgeHTML(tx.Status)}</div>
        <div class="tx-cell hide-md">${tokensHTML}</div>
        <div class="tx-cell hide-md">${activity}</div>
      `;
      row.addEventListener('click', () => openModal(tx));
      el.ongoingList.appendChild(row);
    }
  }

  function statusBadgeHTML(status) {
    const s = String(status || '').toLowerCase();
    let cls = 'badge-default';
    if (s === 'delivered' || s === 'completed' || s === 'success' || s === 'succeeded') cls = 'badge-success';
    else if (s === 'pending' || s === 'processing' || s === 'in_progress') cls = 'badge-pending';
    else if (s === 'failed' || s === 'error' || s === 'reverted') cls = 'badge-failed';
    const label = status || '-';
    return `<span class="badge ${cls}">${label}</span>`;
  }

  function openModal(obj) {
    el.modalBody.textContent = JSON.stringify(obj, null, 2);
    el.modal.setAttribute('aria-hidden', 'false');
  }

  function closeModal() {
    el.modal.setAttribute('aria-hidden', 'true');
  }

  function updateMerchantInfo(list) {
    const { gross, net } = aggregateTotals(list);
    const totalTx = list.length;
    const lastTx = list.length > 0 ? list.reduce((latest, tx) => {
      const txTime = Date.parse(tx.Timestamp);
      const latestTime = Date.parse(latest.Timestamp);
      return txTime > latestTime ? tx : latest;
    }) : null;

    el.totalTransactions.textContent = totalTx;
    el.totalReceived.textContent = formatMoney(net);
    el.lastActivity.textContent = lastTx ? timeAgo(lastTx.Timestamp) : '-';
  }

  function showNoData() {
    el.merchantInfo.style.display = 'block';
    el.mainContent.style.display = 'none';
    el.noDataMessage.style.display = 'block';
  }

  function showData() {
    el.merchantInfo.style.display = 'block';
    el.mainContent.style.display = 'block';
    el.noDataMessage.style.display = 'none';
  }

  // Event listeners
  el.modalBackdrop.addEventListener('click', closeModal);
  el.modalClose.addEventListener('click', closeModal);
  document.addEventListener('keydown', (e) => { if (e.key === 'Escape') closeModal(); });

  // Theme toggling
  function applyTheme(mode) {
    const body = document.body;
    if (mode === 'dark') body.classList.add('dark'); else body.classList.remove('dark');
  }

  function loadTheme() {
    const saved = localStorage.getItem('theme');
    const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
    const mode = saved || (prefersDark ? 'dark' : 'light');
    applyTheme(mode);
    if (el.themeSwitch) el.themeSwitch.checked = mode === 'dark';
  }

  function toggleTheme() {
    const to = (document.body.classList.contains('dark') ? 'light' : 'dark');
    applyTheme(to);
    localStorage.setItem('theme', to);
  }

  if (el.themeSwitch) el.themeSwitch.addEventListener('change', toggleTheme);

  // Logout functionality
  el.logoutBtn.addEventListener('click', () => {
    localStorage.removeItem('merchantAddress');
    window.location.href = 'login.html';
  });

  async function load() {
    if (!currentMerchantAddress) return;
    
    const start = Date.now();
    setStatus('Loading...');
    
    try {
      const data = await fetchWithTimeout(`/api/merchant/${currentMerchantAddress}/payouts`, { timeout: 8000 });
      const list = normalizeTransactions(parseList(data));

      if (list.length === 0) {
        showNoData();
        setStatus('No transactions found for this merchant');
        return;
      }

      showData();
      updateMerchantInfo(list);

      const { gross, net } = aggregateTotals(list);
      el.inflow.textContent = formatMoney(gross);
      el.outflow.textContent = formatMoney(net);

      // Render recent transactions (top 5)
      renderRecentTransactions(list);

      // Render token summary
      renderTokenSummary(list);

      // Render all transactions
      const stickToTop = isNearTop(el.ongoingList);
      const allTx = [...list].sort((a, b) => (b.__ts || 0) - (a.__ts || 0));
      currentList = allTx;
      renderOngoing(getSearchFiltered(allTx));
      if (stickToTop) scrollToTop(el.ongoingList);

      const ms = Date.now() - start;
      setStatus(`Updated • ${new Date().toLocaleTimeString()} • ${list.length} tx • ${ms}ms`);
    } catch (err) {
      console.error('Load error:', err);
      setStatus('Fetch failed: ' + (err && err.message ? err.message : String(err)) + ' • Ensure merchant address is valid');
      showNoData();
    }
  }

  function setStatus(text) { el.status.textContent = text; }

  function normalizeTransactions(list) {
    for (const tx of list) {
      tx.GrossAmount = Number(tx.GrossAmount || 0);
      tx.NetAmount = Number(tx.NetAmount || 0);
      const t = Date.parse(tx.Timestamp);
      tx.__ts = isNaN(t) ? 0 : t;
    }
    return list;
  }

  function isNearTop(container, threshold = 24) {
    return container.scrollTop <= threshold;
  }

  function scrollToTop(container) {
    container.scrollTop = 0;
  }

  el.refreshBtn.addEventListener('click', () => { load(); });

  function debounce(fn, wait) {
    let t; return function(...args){ clearTimeout(t); t = setTimeout(() => fn.apply(this, args), wait); };
  }

  const onSearch = debounce(() => {
    const stickToTop = isNearTop(el.ongoingList);
    renderOngoing(getSearchFiltered(currentList));
    if (stickToTop) scrollToTop(el.ongoingList);
  }, 200);

  if (el.searchInput) el.searchInput.addEventListener('input', onSearch);

  // Initialize
  function init() {
    // Get merchant address from URL or localStorage
    const urlParams = new URLSearchParams(window.location.search);
    const addressFromUrl = urlParams.get('address');
    const addressFromStorage = localStorage.getItem('merchantAddress');
    
    currentMerchantAddress = addressFromUrl || addressFromStorage;
    
    if (!currentMerchantAddress) {
      window.location.href = 'login.html';
      return;
    }

    // Display merchant address
    el.merchantAddressDisplay.textContent = currentMerchantAddress;
    
    loadTheme();
    load();
    setInterval(load, POLL_MS);
  }

  init();
})();
