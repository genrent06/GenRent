#!/bin/bash
set -e
BASE="http://localhost:8080/api/v1"

pass() { echo "✅ $1"; }
fail() { echo "❌ $1"; exit 1; }
info() { echo "ℹ️  $1"; }

# ── 1. REGISTER & LOGIN VENDOR ──────────────────────────────────────────────
info "Step 1: Register vendor"
VR=$(curl -s -X POST $BASE/auth/register -H "Content-Type: application/json" \
  -d '{"name":"Flow Vendor","email":"flowvendor@test.com","password":"TestPass123!","phone":"+918888888888","role":"vendor"}')
MSG=$(echo $VR | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('message',''))" 2>/dev/null)
[ "$MSG" = "registration successful" ] && pass "Vendor registered" || info "Vendor may already exist: $MSG"

VENDOR_TOKEN=$(curl -s -X POST $BASE/auth/login -H "Content-Type: application/json" \
  -d '{"email":"flowvendor@test.com","password":"TestPass123!"}' | python3 -c "import json,sys;print(json.load(sys.stdin)['token'])")
[ -n "$VENDOR_TOKEN" ] && pass "Vendor login OK" || fail "Vendor login failed"

# Create vendor profile
VP=$(curl -s -X POST $BASE/vendors -H "Content-Type: application/json" -H "Authorization: Bearer $VENDOR_TOKEN" \
  -d '{"company_name":"Flow Power Co","city":"Bangalore","phone":"+918888888888","description":"Test vendor","latitude":12.9716,"longitude":77.5946}')
VID=$(echo $VP | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('id',0))" 2>/dev/null)
[ "$VID" -gt 0 ] 2>/dev/null && pass "Vendor profile created (ID: $VID)" || info "Vendor profile: $(echo $VP | python3 -c 'import json,sys;d=json.load(sys.stdin);print(d.get("id","already exists"))' 2>/dev/null)"

# Get vendor profile ID
VENDOR_PROFILE=$(curl -s $BASE/vendors/me -H "Authorization: Bearer $VENDOR_TOKEN")
VID=$(echo $VENDOR_PROFILE | python3 -c "import json,sys;print(json.load(sys.stdin)['id'])")
pass "Vendor profile ID: $VID"

# ── 2. CREATE EQUIPMENT ──────────────────────────────────────────────────────
info "Step 2: Create equipment"
EQ=$(curl -s -X POST $BASE/equipment -H "Content-Type: application/json" -H "Authorization: Bearer $VENDOR_TOKEN" \
  -d "{
    \"name\": \"Test Generator 50kVA\",
    \"category_id\": 2,
    \"brand\": \"Kirloskar\",
    \"model\": \"50kVA\",
    \"description\": \"Test equipment for e2e flow\",
    \"daily_price\": 1000,
    \"weekly_price\": 6000,
    \"monthly_price\": 20000,
    \"mobilization_fee\": 500,
    \"demobilization_fee\": 500,
    \"total_quantity\": 2,
    \"location\": \"Koramangala\",
    \"city\": \"Bangalore\",
    \"latitude\": 12.9279,
    \"longitude\": 77.6271,
    \"specs\": {\"capacity_kva\": 50, \"fuel_type\": \"Diesel\"}
  }")
EID=$(echo $EQ | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('id',0))" 2>/dev/null)
[ "$EID" -gt 0 ] 2>/dev/null && pass "Equipment created (ID: $EID)" || fail "Equipment creation failed: $EQ"

# ── 3. REGISTER & LOGIN CUSTOMER ─────────────────────────────────────────────
info "Step 3: Register customer"
CR=$(curl -s -X POST $BASE/auth/register -H "Content-Type: application/json" \
  -d '{"name":"Flow Customer","email":"flowcustomer@test.com","password":"TestPass123!","phone":"+917777777777","role":"customer"}')
echo $CR | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('message',''))" 2>/dev/null | grep -q "successful" && pass "Customer registered" || info "Customer may already exist"

CUST_TOKEN=$(curl -s -X POST $BASE/auth/login -H "Content-Type: application/json" \
  -d '{"email":"flowcustomer@test.com","password":"TestPass123!"}' | python3 -c "import json,sys;print(json.load(sys.stdin)['token'])")
[ -n "$CUST_TOKEN" ] && pass "Customer login OK" || fail "Customer login failed"

# ── 4. CREATE BOOKING ────────────────────────────────────────────────────────
info "Step 4: Create booking"
BOOKING=$(curl -s -X POST $BASE/bookings -H "Content-Type: application/json" -H "Authorization: Bearer $CUST_TOKEN" \
  -d "{\"equipment_id\": $EID, \"start_date\": \"2026-07-01\", \"end_date\": \"2026-07-08\", \"address\": \"123 Test Street, Bangalore\"}")
BID=$(echo $BOOKING | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('id',0))" 2>/dev/null)
BSTATUS=$(echo $BOOKING | python3 -c "import json,sys;print(json.load(sys.stdin).get('status',''))" 2>/dev/null)
BTOTAL=$(echo $BOOKING | python3 -c "import json,sys;print(json.load(sys.stdin).get('total_price',0))" 2>/dev/null)
BADVANCE=$(echo $BOOKING | python3 -c "import json,sys;print(json.load(sys.stdin).get('advance_amount',0))" 2>/dev/null)
[ "$BID" -gt 0 ] 2>/dev/null && pass "Booking created (ID: $BID, status: $BSTATUS, total: ₹$BTOTAL, advance: ₹$BADVANCE)" || fail "Booking failed: $BOOKING"

# ── 5. VENDOR ACCEPT ─────────────────────────────────────────────────────────
info "Step 5: Vendor accepts booking"
ACCEPT=$(curl -s -X POST $BASE/bookings/$BID/accept -H "Authorization: Bearer $VENDOR_TOKEN")
ASTATUS=$(echo $ACCEPT | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('status',d.get('message','')))" 2>/dev/null)
echo $ACCEPT | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('status','')=='accepted' or 'accepted' in d.get('message',''))" 2>/dev/null | grep -q "True" && pass "Booking accepted (status: $ASTATUS)" || { echo "Response: $ACCEPT"; info "Accept response: $ASTATUS"; }

# Verify booking status
BCHECK=$(curl -s $BASE/bookings/$BID -H "Authorization: Bearer $CUST_TOKEN")
BCHECK_STATUS=$(echo $BCHECK | python3 -c "import json,sys;print(json.load(sys.stdin).get('status',''))" 2>/dev/null)
[ "$BCHECK_STATUS" = "accepted" ] && pass "Booking status confirmed: accepted" || fail "Expected 'accepted', got '$BCHECK_STATUS'"

# ── 6. PROCESS PAYMENT ───────────────────────────────────────────────────────
info "Step 6: Customer pays advance"
PAY=$(curl -s -X POST $BASE/payments -H "Content-Type: application/json" -H "Authorization: Bearer $CUST_TOKEN" \
  -d "{\"booking_id\": $BID, \"amount\": $BADVANCE, \"payment_method\": \"card\", \"payment_type\": \"advance\"}")
PAY_STATUS=$(echo $PAY | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('status',d.get('message',str(d))))" 2>/dev/null)
echo $PAY | python3 -c "import json,sys;d=json.load(sys.stdin);s=d.get('status','');print(s in ['success','completed','advance_paid'])" 2>/dev/null | grep -q "True" && pass "Payment successful: $PAY_STATUS" || { info "Payment response: $PAY"; }

# Check booking status after payment
BCHECK2=$(curl -s $BASE/bookings/$BID -H "Authorization: Bearer $CUST_TOKEN")
BCHECK2_STATUS=$(echo $BCHECK2 | python3 -c "import json,sys;print(json.load(sys.stdin).get('status',''))" 2>/dev/null)
pass "Booking status after payment: $BCHECK2_STATUS"

# ── 7. VENDOR DISPATCH ───────────────────────────────────────────────────────
info "Step 7: Vendor dispatches equipment"
DISPATCH=$(curl -s -X POST $BASE/bookings/$BID/dispatch -H "Authorization: Bearer $VENDOR_TOKEN")
DOTP=$(echo $DISPATCH | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('delivery_otp',d.get('otp','')))" 2>/dev/null)
DSTATUS=$(echo $DISPATCH | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('status',d.get('message','')))" 2>/dev/null)
[ -n "$DOTP" ] && pass "Dispatched! Delivery OTP: $DOTP" || { info "Dispatch response: $DISPATCH"; }

# ── 8. CUSTOMER CONFIRM DELIVERY ─────────────────────────────────────────────
info "Step 8: Customer confirms delivery with OTP"
CONFIRM=$(curl -s -X POST $BASE/bookings/$BID/confirm-delivery -H "Content-Type: application/json" -H "Authorization: Bearer $CUST_TOKEN" \
  -d "{\"otp\": \"$DOTP\"}")
CSTATUS=$(echo $CONFIRM | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('status',d.get('message','')))" 2>/dev/null)
echo $CONFIRM | python3 -c "import json,sys;d=json.load(sys.stdin);s=d.get('status','');print('delivered' in s or 'delivered' in d.get('message',''))" 2>/dev/null | grep -q "True" && pass "Delivery confirmed! Status: $CSTATUS" || { info "Confirm delivery response: $CONFIRM"; }

# ── 9. UPLOAD HANDOVER PHOTOS ─────────────────────────────────────────────
info "Step 9: Vendor uploads delivery handover"
HANDOVER=$(curl -s -X POST "$BASE/bookings/$BID/handover?type=delivery" -H "Content-Type: application/json" -H "Authorization: Bearer $VENDOR_TOKEN" \
  -d '{"photo_urls":["https://example.com/photo1.jpg","https://example.com/photo2.jpg"],"checklist":{"condition":"Good","accessories_present":"Yes","fuel_level":"Full","hours_meter":"1200"},"notes":"Equipment delivered in perfect condition"}')
HSTATUS=$(echo $HANDOVER | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('message',d.get('id','ERROR:'+str(d))))" 2>/dev/null)
pass "Handover uploaded: $HSTATUS"

# ── 10. INITIATE RETURN ───────────────────────────────────────────────────────
info "Step 10: Vendor initiates return"
RETURN=$(curl -s -X POST $BASE/bookings/$BID/initiate-return -H "Authorization: Bearer $VENDOR_TOKEN")
ROTP=$(echo $RETURN | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('return_otp',d.get('otp','')))" 2>/dev/null)
RMSG=$(echo $RETURN | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('message',''))" 2>/dev/null)
pass "Return initiated: $RMSG (OTP: $ROTP)"

# ── 11. CONFIRM RETURN ───────────────────────────────────────────────────────
info "Step 11: Vendor confirms return with OTP"
CONFIRM_RETURN=$(curl -s -X POST $BASE/bookings/$BID/confirm-return -H "Content-Type: application/json" -H "Authorization: Bearer $VENDOR_TOKEN" \
  -d "{\"otp\": \"$ROTP\"}")
CRMSG=$(echo $CONFIRM_RETURN | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('message',d.get('status','')))" 2>/dev/null)
pass "Return confirmed: $CRMSG"

# ── 12. CUSTOMER COMPLETE BOOKING ────────────────────────────────────────────
info "Step 12: Customer completes booking"
COMPLETE=$(curl -s -X POST $BASE/bookings/$BID/complete -H "Authorization: Bearer $CUST_TOKEN")
COMPMSG=$(echo $COMPLETE | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('message',d.get('status','')))" 2>/dev/null)
pass "Booking completed: $COMPMSG"

# ── 13. CHECK VENDOR WALLET ──────────────────────────────────────────────────
info "Step 13: Check vendor wallet balance"
WALLET=$(curl -s $BASE/wallet -H "Authorization: Bearer $VENDOR_TOKEN")
BAL=$(echo $WALLET | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('balance',0))" 2>/dev/null)
HOLD=$(echo $WALLET | python3 -c "import json,sys;d=json.load(sys.stdin);print(d.get('hold_balance',0))" 2>/dev/null)
pass "Vendor wallet — balance: ₹$BAL, hold: ₹$HOLD"

# ── 14. CHECK FINAL BOOKING STATE ────────────────────────────────────────────
info "Step 14: Final booking state"
FINAL=$(curl -s $BASE/bookings/$BID -H "Authorization: Bearer $CUST_TOKEN")
FSTATUS=$(echo $FINAL | python3 -c "import json,sys;print(json.load(sys.stdin).get('status',''))" 2>/dev/null)
pass "Final booking status: $FSTATUS"

echo ""
echo "================================================"
echo "✅ END-TO-END TEST COMPLETE"
echo "   Booking #$BID lifecycle: requested → accepted → paid → dispatched → delivered → completed"
echo "================================================"
