// GenRent Payment Page

let selectedMethod = null;
let bookingDetails = null;

document.addEventListener('DOMContentLoaded', async () => {
  if (!isLoggedIn()) {
    window.location.href = '/login';
    return;
  }

  const bookingId = new URLSearchParams(window.location.search).get('booking_id');
  if (!bookingId) {
    window.location.href = '/';
    return;
  }

  await loadPaymentDetails(bookingId);
});

async function loadPaymentDetails(bookingId) {
  try {
    const data = await api.getPaymentDetails(bookingId);
    bookingDetails = data;

    const total = data.total_amount;
    const advance = data.advance_amount;
    const remaining = Math.round((total - advance) * 100) / 100;

    const gen = data.booking.generator || {};
    const vendor = gen.vendor || {};

    document.getElementById('bookingSummary').textContent =
      `${gen.name || 'Equipment'} · ${vendor.company_name || ''}`;

    document.getElementById('advanceDisplay').textContent = formatCurrency(advance);
    document.getElementById('breakTotal').textContent = formatCurrency(total);
    document.getElementById('breakAdvance').textContent = formatCurrency(advance);
    document.getElementById('breakRemaining').textContent = formatCurrency(remaining);

    document.getElementById('loadingState').style.display = 'none';
    document.getElementById('paymentContent').style.display = 'block';
  } catch (err) {
    document.getElementById('loadingState').innerHTML =
      `<div class="alert alert-error">${err.message}</div>
       <p style="text-align:center;margin-top:1rem;"><a href="/">← Back to home</a></p>`;
  }
}

function selectMethod(method, el) {
  selectedMethod = method;

  // Reset all selections
  document.querySelectorAll('.method-option').forEach(m => m.classList.remove('selected'));
  el.classList.add('selected');

  // Hide all extra inputs
  ['upiSection', 'cardSection', 'netbankingSection'].forEach(id => {
    document.getElementById(id).classList.remove('show');
  });

  // Show relevant input
  if (method === 'upi') document.getElementById('upiSection').classList.add('show');
  if (method === 'card') document.getElementById('cardSection').classList.add('show');
  if (method === 'netbanking') document.getElementById('netbankingSection').classList.add('show');

  const payBtn = document.getElementById('payBtn');
  payBtn.disabled = false;
  payBtn.textContent = method === 'cash'
    ? 'Confirm Cash Payment'
    : `Pay ${document.getElementById('advanceDisplay').textContent} via ${method.toUpperCase()}`;
}

function formatCard(input) {
  let v = input.value.replace(/\D/g, '').substring(0, 16);
  input.value = v.replace(/(.{4})/g, '$1 ').trim();
}

async function processPayment() {
  if (!selectedMethod) {
    showAlert('paymentAlert', 'Please select a payment method', 'error');
    return;
  }

  const bookingId = new URLSearchParams(window.location.search).get('booking_id');
  const btn = document.getElementById('payBtn');
  btn.disabled = true;
  btn.innerHTML = '<span class="loading-spinner"></span> Processing...';
  clearAlert('paymentAlert');

  try {
    const result = await api.processPayment({
      booking_id: parseInt(bookingId),
      method: selectedMethod,
    });

    // Show success screen
    document.getElementById('paymentContent').style.display = 'none';
    document.getElementById('successState').style.display = 'block';
    document.getElementById('txnId').textContent = result.transaction_id;
    document.getElementById('paidAmount').textContent = formatCurrency(result.advance_paid);
    document.getElementById('successMsg').textContent =
      `${formatCurrency(result.escrow_hold)} held in escrow · released to vendor after delivery confirmation`;

    showToast('Payment successful!', 'success');
  } catch (err) {
    showAlert('paymentAlert', err.message, 'error');
    btn.disabled = false;
    btn.textContent = 'Pay Now';
  }
}
