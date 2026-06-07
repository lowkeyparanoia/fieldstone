// Node.js edge functions sidecar for Fieldstone
// Run: npm install && node server.js

const express = require('express');
const app = express();
app.use(express.json());

const PORT = process.env.PORT || 8000;
const FIELDSTONE_URL = process.env.FIELDSTONE_URL || 'http://localhost:8090';

// Function handlers
const handlers = {
  'video-token': handleVideoToken,
  'score-compute': handleScoreCompute,
  'fitbit-sync': handleFitbitSync,
  'fitbit-oauth': handleFitbitOAuth,
  'booking-noshow': handleBookingNoShow,
  'sms-send': handleSMSSend,
  'razorpay-webhook': handleRazorpayWebhook,
  'account-delete': handleAccountDelete,
  'enrich': handleEnrich,
};

async function handleVideoToken(req, res) {
  const { consultationId } = req.body;
  res.json({
    consultationId,
    token: 'mock-zoom-token-' + consultationId,
    joinUrl: `https://zoom.us/j/${consultationId}`,
    expiresAt: new Date(Date.now() + 3600 * 1000).toISOString(),
  });
}

async function handleScoreCompute(req, res) {
  const { userId } = req.body;
  res.json({
    userId,
    scores: {
      recovery: 85,
      strain: 42,
      stress: 30,
      longevity: 78,
    },
    computedAt: new Date().toISOString(),
  });
}

async function handleFitbitSync(req, res) {
  res.json({
    synced: 0,
    message: 'Fitbit sync stub - implement with Fitbit API',
  });
}

async function handleFitbitOAuth(req, res) {
  const { code } = req.query;
  res.json({
    success: true,
    code,
    message: 'Fitbit OAuth callback - implement token exchange',
  });
}

async function handleBookingNoShow(req, res) {
  res.json({
    marked: 0,
    message: 'No-show check complete',
  });
}

async function handleSMSSend(req, res) {
  const { phone, message } = req.body;
  console.log(`SMS to ${phone}: ${message}`);
  res.json({
    success: true,
    phone,
    messageId: 'mock-msg-' + Date.now(),
  });
}

async function handleRazorpayWebhook(req, res) {
  res.json({
    received: true,
    event: req.body.event,
  });
}

async function handleAccountDelete(req, res) {
  const { userId } = req.body;
  res.json({
    userId,
    status: 'deleted',
    deletedAt: new Date().toISOString(),
  });
}

async function handleEnrich(req, res) {
  const { entityType, entityId, data } = req.body;
  res.json({
    entityType,
    entityId,
    enriched: {
      ...data,
      _enrichedAt: new Date().toISOString(),
    },
  });
}

// Route handler
app.all('/functions/v1/:name', async (req, res) => {
  const functionName = req.params.name;
  const handler = handlers[functionName];
  
  if (!handler) {
    return res.status(404).json({ error: `Function '${functionName}' not found` });
  }
  
  try {
    await handler(req, res);
  } catch (err) {
    console.error(`Function ${functionName} error:`, err);
    res.status(500).json({ error: err.message });
  }
});

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

app.listen(PORT, () => {
  console.log(`Edge functions sidecar running on http://localhost:${PORT}`);
});
