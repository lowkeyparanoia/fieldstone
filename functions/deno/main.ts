// Deno edge functions sidecar for Fieldstone
// Run: deno run --allow-net --allow-env --allow-read main.ts

import { serve } from "https://deno.land/std@0.177.0/http/server.ts";

const FIELDSTONE_URL = Deno.env.get("FIELDSTONE_URL") || "http://localhost:8090";
const PORT = parseInt(Deno.env.get("PORT") || "8000");

// Function handlers
const handlers: Record<string, (req: Request) => Promise<Response>> = {
  "video-token": handleVideoToken,
  "score-compute": handleScoreCompute,
  "fitbit-sync": handleFitbitSync,
  "fitbit-oauth": handleFitbitOAuth,
  "booking-noshow": handleBookingNoShow,
  "sms-send": handleSMSSend,
  "razorpay-webhook": handleRazorpayWebhook,
  "account-delete": handleAccountDelete,
  "enrich": handleEnrich,
};

async function handleVideoToken(req: Request): Promise<Response> {
  const body = await req.json().catch(() => ({}));
  const { consultationId } = body;
  
  // In production: lookup consultation, generate Zoom JWT
  return jsonResponse({
    consultationId,
    token: "mock-zoom-token-" + consultationId,
    joinUrl: `https://zoom.us/j/${consultationId}`,
    expiresAt: new Date(Date.now() + 3600 * 1000).toISOString(),
  });
}

async function handleScoreCompute(req: Request): Promise<Response> {
  const body = await req.json().catch(() => ({}));
  const { userId } = body;
  
  // In production: fetch metrics, compute scores, store results
  return jsonResponse({
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

async function handleFitbitSync(req: Request): Promise<Response> {
  const body = await req.json().catch(() => ({}));
  
  // In production: sync Fitbit data for all connected users
  return jsonResponse({
    synced: 0,
    message: "Fitbit sync stub - implement with Fitbit API",
  });
}

async function handleFitbitOAuth(req: Request): Promise<Response> {
  const url = new URL(req.url);
  const code = url.searchParams.get("code");
  
  // In production: exchange code for token, store in device_tokens
  return jsonResponse({
    success: true,
    code,
    message: "Fitbit OAuth callback - implement token exchange",
  });
}

async function handleBookingNoShow(req: Request): Promise<Response> {
  // In production: query consultations, mark past unconfirmed as no_show
  return jsonResponse({
    marked: 0,
    message: "No-show check complete",
  });
}

async function handleSMSSend(req: Request): Promise<Response> {
  const body = await req.json().catch(() => ({}));
  const { phone, message } = body;
  
  // In production: integrate with MSG91/Twilio
  console.log(`SMS to ${phone}: ${message}`);
  
  return jsonResponse({
    success: true,
    phone,
    messageId: "mock-msg-" + Date.now(),
  });
}

async function handleRazorpayWebhook(req: Request): Promise<Response> {
  const body = await req.json().catch(() => ({}));
  
  // In production: verify signature, update payment status
  return jsonResponse({
    received: true,
    event: body.event,
  });
}

async function handleAccountDelete(req: Request): Promise<Response> {
  const body = await req.json().catch(() => ({}));
  const { userId } = body;
  
  // In production: DPDP-compliant deletion
  // 1. Anonymize user data
  // 2. Delete PII
  // 3. Log deletion for audit
  return jsonResponse({
    userId,
    status: "deleted",
    deletedAt: new Date().toISOString(),
  });
}

async function handleEnrich(req: Request): Promise<Response> {
  const body = await req.json().catch(() => ({}));
  const { entityType, entityId, data } = body;
  
  // In production: call enrichment APIs (Clearbit, etc.)
  return jsonResponse({
    entityType,
    entityId,
    enriched: {
      ...data,
      _enrichedAt: new Date().toISOString(),
    },
  });
}

function jsonResponse(data: any, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

// Main server
async function handler(req: Request): Promise<Response> {
  const url = new URL(req.url);
  const path = url.pathname;
  
  // Extract function name from /functions/v1/{name}
  const match = path.match(/^\/functions\/v1\/(.+)$/);
  if (!match) {
    return jsonResponse({ error: "Invalid path" }, 404);
  }
  
  const functionName = match[1];
  const handler = handlers[functionName];
  
  if (!handler) {
    return jsonResponse({ error: `Function '${functionName}' not found` }, 404);
  }
  
  try {
    return await handler(req);
  } catch (err) {
    console.error(`Function ${functionName} error:`, err);
    return jsonResponse({ error: err.message }, 500);
  }
}

console.log(`Edge functions sidecar running on http://localhost:${PORT}`);
await serve(handler, { port: PORT });
