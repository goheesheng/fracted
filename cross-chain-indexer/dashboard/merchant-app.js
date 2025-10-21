(function () {
  const POLL_MS = 5000;
  let currentMerchantAddress = '';
  let merchantToken = '';
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

  function short(v) {
    if (!v) return '-';
    const s = String(v);
    if (s.length <= 12) return s;
    return s.slice(0, 6) + '…' + s.slice(-4);
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
      const headers = { 'Accept': 'application/json' };
      // 如果有 token，添加认证头
      if (merchantToken) {
        headers['Authorization'] = `Bearer ${merchantToken}`;
      }
      const res = await fetch(url, { signal: controller.signal, headers });
      if (!res.ok) {
        if (res.status === 401) {
          // Token 无效，跳转到登录页
          logout();
          throw new Error('Unauthorized');
        }
        throw new Error('HTTP ' + res.status);
      }
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
      // 使用 USD 格式化后的值
      gross += Number(tx.GrossAmountUSD || 0);
      net += Number(tx.NetAmountUSD || 0);
    }
    return { gross, net };
  }

  function groupBy(list, key, amountField) {
    const map = new Map();
    for (const tx of list) {
      const id = tx[key] || '未知';
      // 使用 USD 格式化字段
      let amt = 0;
      if (amountField === 'NetAmount') {
        amt = Number(tx.NetAmountUSD || 0);
      } else if (amountField === 'GrossAmount') {
        amt = Number(tx.GrossAmountUSD || 0);
      } else {
        amt = Number(tx[amountField] || 0);
      }
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

  // LayerZero Endpoint IDs
  const EID_BASE_SEPOLIA = 40245;
  const EID_ARB_SEPOLIA = 40231;
  const EID_SOLANA_DEVNET = 40168;
  const EID_SOLANA_MAINNET = 30168;

  // Token recognition map (extend as needed)
  const TOKEN_MAP = {
    // Base mainnet
    '0x833589fcd6edb6e08f4c7c32d4f71b54b7cfb66e': { symbol: 'USDC', chain: 'Base' },
    '0xd9aaec6eab5f9f0a7f0dd7c39c3f1b3aa1c5f6b9': { symbol: 'USDbC', chain: 'Base' },
    // Base Sepolia
    '0x75faf114eafb1bdbe2f0316df893fd58ce46aa4d': { symbol: 'USDC', chain: 'Base Sepolia' },
    '0x036cbd53842c5426634e7929541ec2318f3dcf7e': { symbol: 'USDC', chain: 'Base Sepolia' },
    '0x323e78f944a9a1fcf3a10efcc5319dbb0bb6e673': { symbol: 'USDT', chain: 'Base Sepolia' },
    // Arbitrum Sepolia
    '0x9aa7fec87ca69695dd1f879567ccf49f3ba417e2': { symbol: 'USDT', chain: 'Arb Sepolia' },
    '0x0f3a3d8e7c8b1e3b5f0b3d3c1a9f4f0a9e3b1c2d': { symbol: 'USDC', chain: 'Arb Sepolia' },
    '0xdac17f958d2ee523a2206206994597c13d831ec7': { symbol: 'USDT', chain: 'Arb Sepolia' },
    '0x75faf114eafb1bdbe2f0316df893fd58ce46aa4d': { symbol: 'USDC', chain: 'Arb Sepolia' },
    '0x4dad09303a773353908f17254b276ee2bd51f0ef': { symbol: 'USDT', chain: 'Arb Sepolia' },
    // Solana Devnet
    'solana:devnet:usdc': { symbol: 'USDC', chain: 'Solana Devnet' },
    'solana:devnet:usdt': { symbol: 'USDT', chain: 'Solana Devnet' },
    // Solana Mainnet
    'solana:mainnet:usdc': { symbol: 'USDC', chain: 'Solana Mainnet' },
    'solana:mainnet:usdt': { symbol: 'USDT', chain: 'Solana Mainnet' },
  };

  function getTokenInfo(addr, dstChain, srcToken) {
    console.log('[DEBUG] getTokenInfo called with:', { addr, dstChain, srcToken });
    if (!addr) return null;
    const k = String(addr).toLowerCase();
    const info = TOKEN_MAP[k];
    if (info) return info;
    
    // 如果 addr 是零地址且目标链是 Solana，从 srcToken 推断
    if (k === '0x0000000000000000000000000000000000000000' && dstChain) {
      if (dstChain.includes('Solana')) {
        // 尝试从源代币推断目标代币符号（通常相同）
        if (srcToken) {
          const srcInfo = TOKEN_MAP[String(srcToken).toLowerCase()];
          if (srcInfo) {
            console.log('[DEBUG] Found srcToken match:', srcToken, '->', srcInfo);
            return { symbol: srcInfo.symbol, chain: dstChain };
          }
        }
        // 默认返回 USDC
        console.log('[DEBUG] Using default USDC for Solana chain');
        return { symbol: 'USDC', chain: dstChain };
      }
    }
    
    // 如果 addr 是零地址，尝试从 srcToken 推断
    if (k === '0x0000000000000000000000000000000000000000' && srcToken) {
      const srcInfo = TOKEN_MAP[String(srcToken).toLowerCase()];
      if (srcInfo) {
        console.log('[DEBUG] Found srcToken match for zero address:', srcToken, '->', srcInfo);
        return { symbol: srcInfo.symbol, chain: dstChain || 'Unknown' };
      }
    }
    
    console.log('[DEBUG] No token info found for:', { addr, dstChain, srcToken, k });
    return null;
  }

  function timeAgo(iso) {
    const t = Date.parse(iso);
    if (isNaN(t)) return '-';
    const s = Math.abs(Math.floor((Date.now() - t) / 1000)); // 取绝对值，未来时间也显示为ago
    
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
      // 优先检查 SrcToken（源链代币），再检查 DstToken（目标链代币）
      const token = getTokenInfo(tx.SrcToken, tx.DstChain) || getTokenInfo(tx.DstToken, tx.DstChain, tx.SrcToken);
      
      // 特殊调试：检查特定的交易 Hash
      if (tx.TxHash === '0xb601d8601253ca455186b10f259398e5c6df96aa5c5684f65533419f210355fb') {
        console.log('[SPECIAL DEBUG] 检查问题交易:', {
          txHash: tx.TxHash,
          srcToken: tx.SrcToken,
          dstToken: tx.DstToken,
          dstChain: tx.DstChain,
          srcTokenLower: tx.SrcToken ? String(tx.SrcToken).toLowerCase() : 'null',
          dstTokenLower: tx.DstToken ? String(tx.DstToken).toLowerCase() : 'null',
          srcTokenInMap: tx.SrcToken ? TOKEN_MAP[String(tx.SrcToken).toLowerCase()] : 'null',
          dstTokenInMap: tx.DstToken ? TOKEN_MAP[String(tx.DstToken).toLowerCase()] : 'null',
          isZeroAddress: tx.DstToken === '0x0000000000000000000000000000000000000000',
          isSolanaChain: tx.DstChain && tx.DstChain.includes('Solana')
        });
      }
      
      // 调试：如果 token 为 null，打印详细信息
      if (!token) {
        console.log('[DEBUG] No token found for tx:', {
          txHash: tx.TxHash,
          srcToken: tx.SrcToken,
          dstToken: tx.DstToken,
          dstChain: tx.DstChain,
          srcTokenLower: tx.SrcToken ? String(tx.SrcToken).toLowerCase() : 'null',
          dstTokenLower: tx.DstToken ? String(tx.DstToken).toLowerCase() : 'null',
          srcTokenInMap: tx.SrcToken ? TOKEN_MAP[String(tx.SrcToken).toLowerCase()] : 'null',
          dstTokenInMap: tx.DstToken ? TOKEN_MAP[String(tx.DstToken).toLowerCase()] : 'null'
        });
      }
      
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
          $${(tx.NetAmountUSD || '0.00')}
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
      // 优先检查 SrcToken（源链代币），再检查 DstToken（目标链代币）
      const token = getTokenInfo(tx.SrcToken, tx.DstChain) || getTokenInfo(tx.DstToken, tx.DstChain, tx.SrcToken);
      if (token) {
        const key = `${token.symbol}-${token.chain}`;
        const current = tokenMap.get(key) || { symbol: token.symbol, chain: token.chain, amount: 0, count: 0 };
        // 使用 USD 格式化后的值
        current.amount += Number(tx.NetAmountUSD || 0);
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
      // 优先检查 SrcToken（源链代币），再检查 DstToken（目标链代币）
      const token = getTokenInfo(tx.SrcToken, tx.DstChain) || getTokenInfo(tx.DstToken, tx.DstChain, tx.SrcToken);
      const tokensHTML = token ? `<div class="tokens">${tokenPillHTML(token.symbol, token.chain)}</div>` : `<span class="badge badge-default">N/A</span>`;
      // 使用 USD 格式化后的值
      const netAmountUSD = tx.NetAmountUSD || '0.00';
      const activity = `<span class="icon inflow"></span>Received $${netAmountUSD} ${token ? token.symbol : ''} from ${short(tx.Payer)}`;
      const hue = hueFromString(String(tx.Payer || ''));
      const row = document.createElement('div');
      row.className = 'tx-row';
      row.innerHTML = `
        <div class="tx-cell" title="${tx.Payer}"><span class="avatar" style="--h:${hue}"></span><span class="mono">${short(tx.Payer)}</span></div>
        <div class="tx-cell" title="${tx.Timestamp}">${timeAgo(tx.Timestamp)}</div>
        <div class="tx-cell">$${netAmountUSD}</div>
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
  function logout() {
    localStorage.removeItem('merchantAddress');
    localStorage.removeItem('merchantToken');
    window.location.href = 'login.html';
  }
  
  el.logoutBtn.addEventListener('click', logout);

  async function load() {
    if (!merchantToken) {
      setStatus('Please login');
      return;
    }
    
    const start = Date.now();
    setStatus('Loading...');
    
    try {
      // 使用认证的 merchant API
      const data = await fetchWithTimeout('/merchant/payouts?limit=100', { timeout: 8000 });
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
    // 从 localStorage 获取 token 和地址
    merchantToken = localStorage.getItem('merchantToken');
    currentMerchantAddress = localStorage.getItem('merchantAddress');
    
    if (!merchantToken || !currentMerchantAddress) {
      // 未登录，跳转到登录页
      window.location.href = 'login.html';
      return;
    }

    // 显示商家地址（支持 EVM 和 Solana 格式，使用简写）
    el.merchantAddressDisplay.textContent = short(currentMerchantAddress);
    el.merchantAddressDisplay.title = currentMerchantAddress; // 完整地址在 tooltip 中
    el.merchantInfo.style.display = 'block';
    
    loadTheme();
    load();
    setInterval(load, POLL_MS);
  }

  init();
})();
