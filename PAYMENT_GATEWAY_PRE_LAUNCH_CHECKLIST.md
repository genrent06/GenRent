# Payment Gateway Pre-Launch Checklist

## Overview
This checklist ensures the payment gateway integration is production-ready and follows best practices for security, reliability, and compliance.

---

## Phase 1: Gateway Configuration

### Razorpay
- [ ] Register production account at [razorpay.com](https://razorpay.com)
- [ ] Generate production API keys (Key ID & Key Secret)
- [ ] Configure webhook endpoints in Razorpay dashboard
- [ ] Enable required payment methods (UPI, Cards, Netbanking, Wallets)
- [ ] Set up refund rules and auto-refund settings
- [ ] Configure settlement schedule (daily/weekly)
- [ ] Verify currency support (INR primarily)
- [ ] Test with production API keys in staging environment

### Stripe
- [ ] Register production account at [stripe.com](https://stripe.com)
- [ ] Generate production API keys (Publishable & Secret)
- [ ] Configure webhook endpoints in Stripe dashboard
- [ ] Set up webhook signing secret for verification
- [ ] Enable required payment methods (Cards, SEPA, etc.)
- [ ] Configure 3D Secure settings for India market
- [ ] Set up Radar fraud protection rules
- [ ] Configure multi-currency support if needed

---

## Phase 2: Environment Setup

### Production Environment
- [ ] Set production environment variables in `.env.production`
  ```env
  # Razorpay
  RAZORPAY_KEY_ID=prod_key_xxxxx
  RAZORPAY_KEY_SECRET=prod_secret_xxxxx
  RAZORPAY_WEBHOOK_SECRET=webhook_secret_xxxxx
  
  # Stripe
  STRIPE_PUBLISHABLE_KEY=pk_prod_xxxxx
  STRIPE_SECRET_KEY=sk_prod_xxxxx
  STRIPE_WEBHOOK_SECRET=whsec_xxxxx
  
  # Payment Gateway
  PAYMENT_GATEWAY=razorpay  # or stripe, or both
  PAYMENT_TIMEOUT=900
  PLATFORM_FEE_RATE=10.0
  ```

- [ ] Ensure `test_mode=false` in production config
- [ ] Configure production database with payment tables
- [ ] Set up SSL/TLS for all payment endpoints
- [ ] Configure CORS for payment gateway domains

### Staging Environment
- [ ] Mirror production configuration with test keys
- [ ] Set up staging database with test data
- [ ] Configure staging webhooks
- [ ] Enable detailed payment logging

---

## Phase 3: Database & Schema

### Required Tables
- [ ] `payments` table with all required fields
  - `id`, `booking_id`, `total_amount`, `advance_amount`
  - `vendor_amount`, `platform_fee`
  - `status`, `payment_method`, `gateway`
  - `gateway_order_id`, `gateway_payment_id`
  - `escrow_held_at`, `escrow_released_at`
  - `refund_id`, `refund_amount`, `refund_status`, `refunded_at`
  - `payment_metadata` (JSONB)
  - `created_at`, `paid_at`

- [ ] `payment_transactions` table
  - `id`, `payment_id`, `booking_id`
  - `transaction_type` (credit/debit/escrow_hold/refund)
  - `amount`, `gateway`, `status`
  - `metadata` (JSONB)
  - `created_at`

- [ ] `vendor_wallets` table
  - `id`, `vendor_id`, `balance`
  - `last credited_at`, `updated_at`

- [ ] `refund_requests` table
  - `id`, `payment_id`, `booking_id`
  - `requested_by`, `amount`, `reason`
  - `status`, `processed_at`
  - `gateway_refund_id`

### Indexes & Constraints
- [ ] Add indexes on `payments.booking_id`
- [ ] Add indexes on `payments.status`
- [ ] Add indexes on `payment_transactions.payment_id`
- [ ] Add foreign key constraints
- [ ] Add check constraints for amounts (> 0)
- [ ] Set up database backups (daily)

---

## Phase 4: API Implementation

### Payment Endpoints
- [ ] `POST /api/payments/create-order` - Create payment order
- [ ] `GET /api/payments/verify?payment_id=xxx` - Verify payment
- [ ] `POST /api/payments/refund` - Process refund
- [ ] `GET /api/payments/methods` - Get payment methods
- [ ] `GET /api/payments/status?booking_id=xxx` - Get payment status
- [ ] `GET /api/payments/history` - Get payment history
- [ ] `POST /webhook/payment` - Handle webhooks

### Escrow Endpoints
- [ ] `POST /api/escrow/hold-funds` - Hold funds in escrow
- [ ] `POST /api/escrow/release-funds` - Release funds to vendor
- [ ] `POST /api/escrow/process-refund` - Process escrow refund
- [ ] `GET /api/escrow/status/:booking_id` - Get escrow status
- [ ] `GET /api/escrow/wallet/:vendor_id` - Get vendor wallet balance
- [ ] `GET /api/escrow/transactions/:vendor_id` - Get vendor transactions

### Refund Endpoints
- [ ] `POST /api/refunds/initiate` - Initiate refund
- [ ] `GET /api/refunds/status/:refund_id` - Get refund status
- [ ] `POST /api/refunds/process-pending` - Process pending refunds
- [ ] `GET /api/refunds/calculate?booking_id=xxx&hours_before=xxx` - Calculate partial refund

### Middleware & Security
- [ ] Implement authentication for all payment endpoints
- [ ] Add rate limiting for payment creation
- [ ] Add request validation for all endpoints
- [ ] Implement idempotency keys for payment creation
- [ ] Add request signing verification for webhooks
- [ ] Implement proper error responses

---

## Phase 5: Testing

### Unit Tests
- [ ] Payment gateway service tests (both Razorpay & Stripe)
- [ ] Escrow service tests (hold, release, refund)
- [ ] Refund service tests (initiate, calculate, process)
- [ ] Handler tests (all endpoints)
- [ ] Error handling tests

### Integration Tests
- [ ] End-to-end payment flow test
- [ ] Escrow flow test (hold → release)
- [ ] Refund flow test (full & partial)
- [ ] Webhook handling tests
- [ ] Database transaction tests
- [ ] Error recovery tests

### Manual Testing
- [ ] Test Razorpay payment flow with test card
- [ ] Test Stripe payment flow with test card
- [ ] Test UPI payment (Razorpay)
- [ ] Test failed payment scenarios
- [ ] Test refund processing
- [ ] Test webhook delivery
- [ ] Test escrow operations
- [ ] Test vendor wallet credits

### Load Testing
- [ ] Test with 100 concurrent payment requests
- [ ] Test with 1000 concurrent payment requests
- [ ] Measure response times (p50, p95, p99)
- [ ] Test webhook handling under load
- [ ] Test database performance under load

---

## Phase 6: Security

### API Security
- [ ] Use HTTPS for all payment endpoints
- [ ] Implement proper authentication (JWT)
- [ ] Add API rate limiting
- [ ] Validate all input parameters
- [ ] Sanitize error messages (don't leak secrets)
- [ ] Implement request size limits
- [ ] Add CORS headers properly

### Webhook Security
- [ ] Verify webhook signatures for all events
- [ ] Use HTTPS webhook endpoints
- [ ] Implement idempotent webhook processing
- [ ] Add webhook retry logic
- [ ] Log all webhook events
- [ ] Monitor for duplicate webhook events

### Data Security
- [ ] Encrypt sensitive data at rest
- [ ] Use TLS for database connections
- [ ] Implement secure key storage (env variables/secret manager)
- [ ] Never log full payment details/card numbers
- [ ] Implement proper access controls
- [ ] Regular security audits

### Compliance
- [ ] GDPR compliance (data handling)
- [ ] PCI DSS compliance (card data handling)
- [ ] RBI compliance for UPI/wallets
- [ ] Privacy policy updated
- [ ] Terms of service updated
- [ ] Refund policy documented

---

## Phase 7: Monitoring & Logging

### Logging
- [ ] Log all payment requests (without sensitive data)
- [ ] Log all payment responses
- [ ] Log all webhook events
- [ ] Log all errors and exceptions
- [ ] Log escrow operations
- [ ] Log refund operations
- [ ] Structured logging (JSON format)
- [ ] Log retention policy (90 days)

### Monitoring
- [ ] Set up payment success rate monitoring
- [ ] Monitor payment failure reasons
- [ ] Monitor webhook delivery failures
- [ ] Monitor refund processing time
- [ ] Monitor escrow release delays
- [ ] Set up alerting for critical failures
- [ ] Dashboard for payment metrics

### Metrics to Track
- [ ] Total payments per day
- [ ] Payment success rate (> 95% target)
- [ ] Average payment processing time
- [ ] Refund rate (< 5% target)
- [ ] Escrow release time
- [ ] Webhook processing time
- [ ] Revenue per payment gateway
- [ ] Platform fee collection

---

## Phase 8: Documentation

### API Documentation
- [ ] Document all payment endpoints
- [ ] Document all escrow endpoints
- [ ] Document webhook events
- [ ] Document error codes
- [ ] Provide code examples
- [ ] Document payment flow
- [ ] Document refund policy

### Internal Documentation
- [ ] Architecture documentation
- [ ] Database schema documentation
- [ ] Configuration guide
- [ ] Troubleshooting guide
- [ ] Runbook for common issues
- [ ] Emergency procedures

### User Documentation
- [ ] Payment methods guide
- [ ] Refund policy page
- [ ] FAQ for payment issues
- [ ] Contact support information

---

## Phase 9: Deployment

### Pre-Deployment
- [ ] Run all tests and ensure 100% pass
- [ ] Review and merge all code changes
- [ ] Update production environment variables
- [ ] Create database backups
- [ ] Notify team about deployment
- [ ] Prepare rollback plan

### Deployment Steps
- [ ] Deploy to production during low-traffic hours
- [ ] Run database migrations
- [ ] Verify webhook endpoints are accessible
- [ ] Test payment flow with small amount
- [ ] Monitor logs for errors
- [ ] Verify webhook delivery
- [ ] Test refund process
- [ ] Scale up servers if needed

### Post-Deployment
- [ ] Monitor payment success rate for first hour
- [ ] Check webhook delivery
- [ ] Verify escrow operations
- [ ] Test refund with small amount
- [ ] Review error logs
- [ ] Monitor server performance
- [ ] Check database performance

---

## Phase 10: Support & Maintenance

### Support Setup
- [ ] Create payment issue escalation flow
- [ ] Set up support email/phone
- [ ] Train support team on payment issues
- [ ] Create troubleshooting guides
- [ ] Set up notification system for failures

### Maintenance Schedule
- [ ] Daily: Review payment failures
- [ ] Daily: Check webhook delivery
- [ ] Weekly: Review refund requests
- [ ] Weekly: Monitor revenue reconciliation
- [ ] Monthly: Review payment gateway fees
- [ ] Monthly: Security audit
- [ ] Quarterly: Performance optimization
- [ ] Quarterly: Compliance review

### Continuous Improvement
- [ ] Collect user feedback on payments
- [ ] Monitor payment gateway performance
- [ ] Optimize payment conversion rate
- [ ] Regular security updates
- [ ] Feature updates based on feedback

---

## Emergency Contacts

### Payment Gateway Support
- **Razorpay Support**: [support@razorpay.com](mailto:support@razorpay.com)
- **Stripe Support**: [support@stripe.com](mailto:support@stripe.com)

### Internal Team
- **Engineering Lead**: [Contact]
- **DevOps Lead**: [Contact]
- **Product Manager**: [Contact]
- **Customer Support Lead**: [Contact]

---

## Test Credentials

### Razorpay Test Mode
- **Test Card**: 4242 4242 4242 4242
- **Test UPI**: razorpay@upi
- **Expiry**: Any future date
- **CVV**: Any 3 digits

### Stripe Test Mode
- **Test Card**: 4242 4242 4242 4242
- **Expiry**: Any future date
- **CVC**: Any 3 digits

---

## Success Criteria

### Launch Day Targets
- [ ] Payment success rate > 95%
- [ ] Webhook delivery rate > 99%
- [ ] Average payment processing time < 5 seconds
- [ ] Zero critical security issues
- [ ] All monitors operational
- [ ] Support team trained and ready

### 30-Day Targets
- [ ] Payment success rate > 98%
- [ ] Refund processing time < 24 hours
- [ ] Escrow release automation working
- [ ] Platform fees collected accurately
- [ ] Customer satisfaction > 4.5/5

---

## Notes

1. **Payment Gateway Selection**: Start with Razorpay for India market, add Stripe for international expansion
2. **Escrow Period**: Hold funds for 24-48 hours after equipment delivery before releasing to vendors
3. **Refund Policy**: Follow the cancellation policy defined in CalculatePartialRefund function
4. **Platform Fee**: Default to 10% but make configurable
5. **Webhook Retry**: Implement exponential backoff for failed webhook deliveries
6. **Database**: Use PostgreSQL for production with proper backup strategy
7. **Monitoring**: Set up alerts for payment failures, webhook delivery failures, and unusual refund patterns

---

**Last Updated**: 2026-07-14
**Version**: 1.0
**Status**: Ready for Production
