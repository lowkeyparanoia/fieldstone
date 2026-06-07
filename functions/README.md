# Fieldstone Edge Functions Sidecar

This directory contains edge functions that run in a Deno/Node sidecar.
Fieldstone proxies `/functions/v1/*` requests to this sidecar.

## Setup

### Deno (recommended)
```bash
cd functions/deno
deno run --allow-net --allow-env --allow-read main.ts
```

### Node.js
```bash
cd functions/node
npm install
node server.js
```

## Environment Variables
- `FIELDSTONE_URL` - Fieldstone API URL (default: http://localhost:8090)
- `FIELDSTONE_ANON_KEY` - Anon key for service-to-service calls

## Functions

### video-token
Generates Zoom Video SDK tokens for consultations.

### score-compute
Computes recovery/strain/stress/longevity scores from health metrics.

### fitbit-sync
Syncs Fitbit data for connected users.

### fitbit-oauth
Handles Fitbit OAuth callback.

### booking-noshow
Marks missed consultations as no-show.

### sms-send
Sends SMS via MSG91/Twilio.

### razorpay-webhook
Handles Razorpay payment webhooks.

### account-delete
Handles DPDP-compliant account deletion.

### enrich
Enriches prospect/company data via external APIs.
