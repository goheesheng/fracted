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

app.listen(PORT, () => {
  console.log(`[dashboard] Server running on http://127.0.0.1:${PORT}`);
  console.log(`[dashboard] Static dir: ${staticDir}`);
  console.log(`[dashboard] Proxy /api/payouts -> ${UPSTREAM}`);
});
