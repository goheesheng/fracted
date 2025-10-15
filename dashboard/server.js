const express = require('express');
const path = require('path');
const axios = require('axios');

const app = express();
const PORT = process.env.PORT || 8081;

// Serve static dashboard
const staticDir = path.join(__dirname);
app.use(express.static(staticDir));

// Proxy API to avoid CORS and keep the frontend simple
const UPSTREAM = process.env.UPSTREAM_PAYOUTS || 'http://85.211.176.154:8080/payouts';
app.get('/api/payouts', async (req, res) => {
  try {
    const response = await axios.get(UPSTREAM, { timeout: 10000 });
    res.set('Cache-Control', 'no-store');
    res.json(response.data);
  } catch (err) {
    const code = err.response?.status || 502;
    res.status(code).json({ error: 'Upstream error', detail: err.message });
  }
});

// Root -> index.html
app.get('/', (req, res) => {
  res.sendFile(path.join(staticDir, 'index.html'));
});

// Login page
app.get('/login.html', (req, res) => {
  res.sendFile(path.join(staticDir, 'login.html'));
});

// Merchant dashboard page
app.get('/merchant-dashboard.html', (req, res) => {
  res.sendFile(path.join(staticDir, 'merchant-dashboard.html'));
});

// Merchant-specific API endpoint
app.get('/api/merchant/:address/payouts', async (req, res) => {
  try {
    const merchantAddress = req.params.address.toLowerCase();
    const response = await axios.get(UPSTREAM, { timeout: 10000 });
    
    // Filter transactions for this specific merchant
    const allData = response.data;
    
    // Handle both array and object formats
    let transactions = [];
    if (Array.isArray(allData)) {
      transactions = allData;
    } else if (allData && Array.isArray(allData.list)) {
      transactions = allData.list;
    }
    
    const filteredTransactions = transactions.filter(tx => 
      tx.Merchant && tx.Merchant.toLowerCase() === merchantAddress
    );
    
    // Return in the same format as the original API
    const filteredData = Array.isArray(allData) ? filteredTransactions : {
      ...allData,
      list: filteredTransactions
    };
    
    res.set('Cache-Control', 'no-store');
    res.json(filteredData);
  } catch (err) {
    const code = err.response?.status || 502;
    res.status(code).json({ error: 'Upstream error', detail: err.message });
  }
});

app.listen(PORT, () => {
  console.log(`[dashboard] Server running on http://127.0.0.1:${PORT}`);
  console.log(`[dashboard] Static dir: ${staticDir}`);
  console.log(`[dashboard] Proxy /api/payouts -> ${UPSTREAM}`);
  console.log(`[dashboard] Merchant API: /api/merchant/:address/payouts`);
});
