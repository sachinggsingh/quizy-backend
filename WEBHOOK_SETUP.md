# Stripe Webhook Setup Guide

## Problem: Webhooks Not Reaching Your Server

Your webhook endpoint is working (test endpoint responds), but Stripe webhooks aren't being sent to your local server.

## Solution: Use Stripe CLI for Local Testing

### Step 1: Install Stripe CLI

```bash
# macOS
brew install stripe/stripe-cli/stripe

# Or download from: https://stripe.com/docs/stripe-cli
```

### Step 2: Login to Stripe CLI

```bash
stripe login
```

This will open your browser to authenticate.

### Step 3: Forward Webhooks to Your Local Server

In a **separate terminal** (keep your Go server running), run:

```bash
stripe listen --forward-to localhost:8080/webhook
```

You'll see output like:
```
> Ready! Your webhook signing secret is whsec_xxxxx (^C to quit)
```

### Step 4: Copy the Webhook Secret

Copy the `whsec_xxxxx` value and add it to your `.env` file:

```env
STRIPE_WEBHOOK_SECRET=whsec_xxxxx
```

**Important:** Restart your Go server after updating `.env`!

### Step 5: Test the Webhook

In another terminal, trigger a test event:

```bash
# Test checkout.session.completed
stripe trigger checkout.session.completed

# Test invoice.payment_succeeded  
stripe trigger invoice.payment_succeeded
```

### Step 6: Watch Your Server Logs

You should now see logs like:
```
========== WEBHOOK RECEIVED ==========
Method: POST, Path: /webhook
Payload size: 1234 bytes
[Service] HandleWebhook called...
[Service] Stripe Event constructed successfully - Type: checkout.session.completed
[Webhook] Handling checkout.session.completed...
[TransactionRepo] Inserting transaction...
[TransactionRepo] MongoDB InsertOne Success...
```

## Alternative: Using ngrok for Production-like Testing

If you want to test with real Stripe Dashboard webhooks:

### Step 1: Install ngrok

```bash
brew install ngrok
# Or download from: https://ngrok.com/
```

### Step 2: Expose Your Local Server

```bash
ngrok http 8080
```

This will give you a public URL like: `https://abc123.ngrok.io`

### Step 3: Configure Webhook in Stripe Dashboard

1. Go to: https://dashboard.stripe.com/test/webhooks
2. Click "Add endpoint"
3. Endpoint URL: `https://abc123.ngrok.io/webhook`
4. Select events to listen to:
   - `checkout.session.completed`
   - `invoice.payment_succeeded`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
5. Click "Add endpoint"
6. Copy the "Signing secret" (starts with `whsec_`)
7. Add to your `.env`: `STRIPE_WEBHOOK_SECRET=whsec_xxxxx`
8. Restart your server

### Step 4: Test with Real Checkout

1. Go to your frontend
2. Click "Subscribe to Pro"
3. Complete the payment in Stripe test mode
4. Watch your server logs for webhook events

## Troubleshooting

### No logs at all
- ✅ Webhook endpoint is accessible (test endpoint works)
- ❌ Stripe isn't sending webhooks
- **Fix:** Use `stripe listen` or configure ngrok

### "Error constructing event"
- ❌ Webhook secret mismatch
- **Fix:** Ensure `STRIPE_WEBHOOK_SECRET` matches the secret from `stripe listen` or Stripe Dashboard

### "No local subscription found"
- ❌ Subscription wasn't saved first
- **Fix:** Ensure `checkout.session.completed` saves subscription before `invoice.payment_succeeded` tries to save transaction

### Transactions not saving
- Check MongoDB connection
- Look for MongoDB errors in logs
- Verify `transactions` collection exists

## Quick Test Commands

```bash
# 1. Start your Go server
go run ./cmd/server/main.go

# 2. In another terminal, forward webhooks
stripe listen --forward-to localhost:8080/webhook

# 3. In another terminal, trigger test events
stripe trigger checkout.session.completed
stripe trigger invoice.payment_succeeded

# 4. Check MongoDB
mongosh
use quizdb
db.transactions.find().pretty()
db.subscriptions.find().pretty()
```

## Expected Flow

1. User clicks "Subscribe" → Frontend calls `/create-checkout-session`
2. User completes payment → Stripe redirects to success URL
3. Stripe sends `checkout.session.completed` webhook → Saves subscription + transaction
4. Stripe sends `invoice.payment_succeeded` webhook → Saves transaction (for recurring payments)
