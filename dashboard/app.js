(function () {
  const ENDPOINT = '/dashboard/api/payouts';
  const POLL_MS = 5000;

  const el = {
    status: document.getElementById('status'),
    themeSwitch: document.getElementById('theme-switch'),
    searchInput: document.getElementById('search-input'),
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
    // 登录相关
    loginModal: document.getElementById('login-modal'),
    loginForm: document.getElementById('login-form'),
    adminAddress: document.getElementById('admin-address'),
    loginBtn: document.getElementById('login-btn'),
    userInfo: document.getElementById('user-info'),
    userRole: document.getElementById('user-role'),
    userAddress: document.getElementById('user-address'),
    logoutBtn: document.getElementById('logout-btn'),
  };

  let currentList = [];
  let authToken = localStorage.getItem('adminToken');
  let currentUser = null;

  function formatMoney(num) {
    if (!isFinite(num)) return '$0';
    return '$' + Number(num).toLocaleString(undefined, { maximumFractionDigits: 2 });
  }

  function getSearchFiltered(list) {
    const q = (el.searchInput && el.searchInput.value || '').trim().toLowerCase();
    if (!q) return list;
    return list.filter(tx => String(tx.Merchant || '').toLowerCase().includes(q));
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

  async function fetchWithTimeout(url, { timeout = 6000, headers = {} } = {}) {
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), timeout);
    try {
      const defaultHeaders = { 'Accept': 'application/json' };
      // 如果有 authToken，自动添加到请求头
      if (authToken) {
        defaultHeaders['Authorization'] = `Bearer ${authToken}`;
      }
      const allHeaders = { ...defaultHeaders, ...headers };
      const res = await fetch(url, { signal: controller.signal, headers: allHeaders });
      if (!res.ok) {
        if (res.status === 401) {
          // Token 过期或无效，清除并显示登录
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

  // 认证相关函数
  function showLoginModal() {
    el.loginModal.setAttribute('aria-hidden', 'false');
  }

  function hideLoginModal() {
    el.loginModal.setAttribute('aria-hidden', 'true');
  }

  function showUserInfo(user) {
    currentUser = user;
    el.userRole.textContent = 'ADMIN';
    el.userAddress.textContent = short(user.address);
    el.userInfo.style.display = 'flex';
  }

  function hideUserInfo() {
    currentUser = null;
    el.userInfo.style.display = 'none';
  }

  function logout() {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('adminToken');
    hideUserInfo();
    showLoginModal();
    setStatus('Logged out');
  }

  async function login(address) {
    try {
      const response = await fetch('/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ address: address, role: 'admin' }),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
      }

      const data = await response.json();
      authToken = data.token;
      localStorage.setItem('adminToken', authToken);
      
      showUserInfo({ address: data.address, role: data.role });
      hideLoginModal();
      setStatus('Login successful');
      
      // 登录后立即加载数据
      load();
    } catch (error) {
      setStatus('Login failed: ' + error.message);
      throw error;
    }
  }

  async function checkAuth() {
    if (!authToken) {
      showLoginModal();
      return false;
    }

    try {
      const response = await fetch('/auth/me', {
        headers: {
          'Authorization': `Bearer ${authToken}`,
        },
      });

      if (!response.ok) {
        throw new Error('Token expired or invalid');
      }

      const user = await response.json();
      
      // 验证是否为管理员
      if (user.role !== 'admin') {
        throw new Error('Admin access required');
      }
      
      showUserInfo(user);
      return true;
    } catch (error) {
      logout();
      return false;
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
    // Arbitrum Sepolia (common test tokens; adjust if needed)
    '0x9aa7fEc87CA69695Dd1f879567CcF49F3ba417E2': { symbol: 'USDT', chain: 'Arb Sepolia' },
    '0x0f3a3d8e7c8b1e3b5f0b3d3c1a9f4f0a9e3b1c2d': { symbol: 'USDC', chain: 'Arb Sepolia' },
    // Solana Devnet - Common test tokens (更新实际地址后可识别)
    // USDC Devnet: https://explorer.solana.com/address/Gh9ZwEmdLJ8DscKNTkTqPbNwLNNBjuSzaG9Vp2KGtKJr?cluster=devnet
    'gh9zwemldj8dsckntktqpbnwlnnbjuszag9vp2kgtkjr': { symbol: 'USDC', chain: 'Solana Devnet' },
    // USDT Devnet placeholder
    'es9vmfrzacermjfrf4h2fyd4kconky11mccebenwnyb': { symbol: 'USDT', chain: 'Solana Devnet' },
    // Solana Mainnet - Official tokens
    'epjfwdd5aufqssqem2qn1xzybapC8g4wegGkzwytdt1v': { symbol: 'USDC', chain: 'Solana Mainnet' },
    'es9vmfrzacermjfrf4h2fyd4kconky11mccebenwnyb': { symbol: 'USDT', chain: 'Solana Mainnet' },
  };

  // Get chain name by EID
  function getChainNameByEid(eid) {
    switch(Number(eid)) {
      case EID_BASE_SEPOLIA: return 'Base Sepolia';
      case EID_ARB_SEPOLIA: return 'Arbitrum Sepolia';
      case EID_SOLANA_DEVNET: return 'Solana Devnet';
      case EID_SOLANA_MAINNET: return 'Solana Mainnet';
      default: return `Chain ${eid}`;
    }
  }

  function getTokenInfo(addr, dstChain) {
    if (!addr) return null;
    const k = String(addr).toLowerCase();
    
    // 先查找精确匹配
    if (TOKEN_MAP[k]) {
      return TOKEN_MAP[k];
    }
    
    // 如果是 Solana 地址格式（不是 0x 开头），根据目标链推断
    if (!addr.startsWith('0x')) {
      // Solana token 地址推断逻辑
      if (dstChain && dstChain.includes('Solana')) {
        // 根据地址特征或长度推断 token 类型
        // Solana 地址通常是 32-44 字符的 base58 编码
        if (addr.length >= 32) {
          // 默认显示为 USDC/USDT（可以根据实际情况调整）
          return { symbol: 'USDC', chain: dstChain };
        }
      }
    }
    
    return null;
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

  function renderOngoing(list) {
    el.ongoingList.innerHTML = '';

    const head = document.createElement('div');
    head.className = 'tx-row tx-head';
    head.innerHTML = `
      <div class="tx-cell bold">Identity</div>
      <div class="tx-cell bold">Time</div>
      <div class="tx-cell bold">Value</div>
      <div class="tx-cell bold">Destination</div>
      <div class="tx-cell bold hide-md">Tokens</div>
      <div class="tx-cell bold hide-md">Activity</div>
    `;
    el.ongoingList.appendChild(head);

    if (!list.length) {
      el.ongoingEmpty.hidden = false;
      return;
    }
    el.ongoingEmpty.hidden = true;

    for (const tx of list) {
      // 获取目标链名称
      const dstChain = tx.DstChain || getChainNameByEid(tx.DstEid);
      
      // 优先检查 SrcToken（源链代币），传入目标链信息以便智能推断
      const token = getTokenInfo(tx.SrcToken, dstChain) || getTokenInfo(tx.DstToken, dstChain);
      const tokensHTML = token ? `<div class="tokens">${tokenPillHTML(token.symbol, token.chain)}</div>` : `<span class="badge badge-default">N/A</span>`;
      
      // Solana 交易显示绿色标签
      const dstChainBadge = dstChain.includes('Solana') 
        ? `<span class="badge badge-success">${dstChain}</span>` 
        : `<span class="badge badge-default">${dstChain}</span>`;
      
      // 使用 USD 格式化后的值
      const netAmountUSD = tx.NetAmountUSD || '0.00';
      const activity = `<span class="icon inflow"></span>Sent $${netAmountUSD} ${token ? token.symbol : ''} to ${dstChain}`;
      const hue = hueFromString(String(tx.Merchant || ''));
      const row = document.createElement('div');
      row.className = 'tx-row';
      row.innerHTML = `
        <div class="tx-cell" title="${tx.Merchant}"><span class="avatar" style="--h:${hue}"></span><span class="mono">${short(tx.Merchant)}</span></div>
        <div class="tx-cell" title="${tx.Timestamp}">${timeAgo(tx.Timestamp)}</div>
        <div class="tx-cell">$${netAmountUSD}</div>
        <div class="tx-cell">${dstChainBadge}</div>
        <div class="tx-cell hide-md">${tokensHTML}</div>
        <div class="tx-cell hide-md">${activity}</div>
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
  function isNearTop(container, threshold = 24) {
    return container.scrollTop <= threshold;
  }
  function scrollToTop(container) {
    container.scrollTop = 0;
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

  async function load() {
    if (!authToken || !currentUser) {
      setStatus('Please login as admin');
      return;
    }

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

      const stickToTop = isNearTop(el.ongoingList);
      const allTx = [...list].sort((a, b) => (b.__ts || 0) - (a.__ts || 0));
      currentList = allTx;
      renderOngoing(getSearchFiltered(allTx));
      if (stickToTop) scrollToTop(el.ongoingList);

      const ms = Date.now() - start;
      setStatus(`Updated • ${new Date().toLocaleTimeString()} • ${list.length} tx • ${ms}ms`);
    } catch (err) {
      if (err.message === 'Unauthorized') {
        // 认证失败，不显示错误，只等待登录
        return;
      }
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

  function debounce(fn, wait) {
    let t; return function(...args){ clearTimeout(t); t = setTimeout(() => fn.apply(this, args), wait); };
  }
  const onSearch = debounce(() => {
    const stickToTop = isNearTop(el.ongoingList);
    renderOngoing(getSearchFiltered(currentList));
    if (stickToTop) scrollToTop(el.ongoingList);
  }, 200);
  if (el.searchInput) el.searchInput.addEventListener('input', onSearch);

  // 登录表单事件
  el.loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const address = el.adminAddress.value.trim();
    
    if (!address) {
      setStatus('Please enter admin address');
      return;
    }
    
    el.loginBtn.disabled = true;
    el.loginBtn.textContent = 'Logging in...';
    
    try {
      await login(address);
    } catch (error) {
      // 错误已在 login 函数中处理
    } finally {
      el.loginBtn.disabled = false;
      el.loginBtn.textContent = 'Login as Admin';
    }
  });

  el.logoutBtn.addEventListener('click', logout);

  // 初始化
  loadTheme();
  
  // 检查认证状态并加载数据
  async function init() {
    const isAuthenticated = await checkAuth();
    if (isAuthenticated) {
      load();
      setInterval(load, POLL_MS);
    }
  }
  
  init();
})();
