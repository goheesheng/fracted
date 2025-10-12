(function () {
  const ENDPOINT = '/api/payouts';
  const POLL_MS = 5000;

  const el = {
    status: document.getElementById('status'),
    inflow: document.getElementById('kpi-inflow'),
    outflow: document.getElementById('kpi-outflow'),
    merchantList: document.getElementById('merchant-list'),
    merchantEmpty: document.getElementById('merchant-empty'),
    payerList: document.getElementById('payer-list'),
    payerEmpty: document.getElementById('payer-empty'),
    ongoingList: document.getElementById('ongoing-list'),
    ongoingEmpty: document.getElementById('ongoing-empty'),
    refreshBtn: document.getElementById('refresh-btn'),
    modal: document.getElementById('modal'),
    modalBackdrop: document.getElementById('modal-backdrop'),
    modalClose: document.getElementById('modal-close'),
    modalBody: document.getElementById('modal-body'),
  };

  function formatMoney(num) {
    if (!isFinite(num)) return '$0';
    return '$' + Number(num).toLocaleString(undefined, { maximumFractionDigits: 2 });
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

  function renderOngoing(list) {
    el.ongoingList.innerHTML = '';

    const head = document.createElement('div');
    head.className = 'tx-row tx-head';
    head.innerHTML = `
      <div class="tx-cell bold">TxHash</div>
      <div class="tx-cell bold">Merchant</div>
      <div class="tx-cell bold">Payer</div>
      <div class="tx-cell bold">Gross</div>
      <div class="tx-cell bold hide-md">Net</div>
      <div class="tx-cell bold hide-md">Status</div>
    `;
    el.ongoingList.appendChild(head);

    if (!list.length) {
      el.ongoingEmpty.hidden = false;
      return;
    }
    el.ongoingEmpty.hidden = true;

    for (const tx of list) {
      const row = document.createElement('div');
      row.className = 'tx-row';
      row.innerHTML = `
        <div class="tx-cell mono" title="${tx.TxHash}">${short(tx.TxHash)}</div>
        <div class="tx-cell mono" title="${tx.Merchant}">${short(tx.Merchant)}</div>
        <div class="tx-cell mono" title="${tx.Payer}">${short(tx.Payer)}</div>
        <div class="tx-cell">${formatMoney(Number(tx.GrossAmount || 0))}</div>
        <div class="tx-cell hide-md">${formatMoney(Number(tx.NetAmount || 0))}</div>
        <div class="tx-cell hide-md">${statusBadgeHTML(tx.Status)}</div>
      `;
      row.addEventListener('click', () => openModal(tx));
      el.ongoingList.appendChild(row);
    }
  }

  function short(v) {
    if (!v) return '-';
    const s = String(v);
    if (s.length <= 12) return s;
    return s.slice(0, 6) + '…' + s.slice(-4);
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

  function isNearBottom(container, threshold = 24) {
    return (container.scrollHeight - container.scrollTop - container.clientHeight) <= threshold;
  }
  function scrollToBottom(container) {
    container.scrollTop = container.scrollHeight;
  }

  function openModal(obj) {
    el.modalBody.textContent = JSON.stringify(obj, null, 2);
    el.modal.setAttribute('aria-hidden', 'false');
  }
  function closeModal() {
    el.modal.setAttribute('aria-hidden', 'true');
  }

  el.modalBackdrop.addEventListener('click', closeModal);
  el.modalClose.addEventListener('click', closeModal);
  document.addEventListener('keydown', (e) => { if (e.key === 'Escape') closeModal(); });

  async function load() {
    const start = Date.now();
    setStatus('Loading...');
    try {
      const data = await fetchWithTimeout(ENDPOINT, { timeout: 8000 });
      const list = normalizeTransactions(parseList(data));

      const { gross, net } = aggregateTotals(list);
      el.inflow.textContent = formatMoney(gross);
      el.outflow.textContent = formatMoney(net);

      const merchantAgg = groupBy(list, 'Merchant', 'NetAmount');
      renderSimpleList(el.merchantList, el.merchantEmpty, merchantAgg);

      const payerAgg = groupBy(list, 'Payer', 'GrossAmount');
      renderSimpleList(el.payerList, el.payerEmpty, payerAgg);

      const stickToBottom = isNearBottom(el.ongoingList);
      const allTx = [...list].sort((a, b) => (a.__ts || 0) - (b.__ts || 0));
      renderOngoing(allTx);
      if (stickToBottom) scrollToBottom(el.ongoingList);

      const ms = Date.now() - start;
      setStatus(`Updated • ${new Date().toLocaleTimeString()} • ${list.length} tx • ${ms}ms`);
    } catch (err) {
      setStatus('Fetch failed: ' + (err && err.message ? err.message : String(err)) + ' • Ensure upstream/proxy is reachable');
      // Keep old UI, show empties if first load
      if (!el.merchantList.children.length) el.merchantEmpty.hidden = false;
      if (!el.payerList.children.length) el.payerEmpty.hidden = false;
      if (!el.ongoingList.children.length) el.ongoingEmpty.hidden = false;
    }
  }

  function setStatus(text) { el.status.textContent = text; }

  function isFinalStatus(status) {
    const s = String(status || '').toLowerCase();
    return s === 'delivered' || s === 'completed' || s === 'success' || s === 'succeeded';
  }

  function normalizeTransactions(list) {
    // Ensure numeric fields are numbers and timestamp ordering is possible
    for (const tx of list) {
      tx.GrossAmount = Number(tx.GrossAmount || 0);
      tx.NetAmount = Number(tx.NetAmount || 0);
      const t = Date.parse(tx.Timestamp);
      tx.__ts = isNaN(t) ? 0 : t;
    }
    return list;
  }

  el.refreshBtn.addEventListener('click', () => { load(); });

  load();
  setInterval(load, POLL_MS);
})();
