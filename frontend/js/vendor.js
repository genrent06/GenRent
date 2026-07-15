// GenRent Vendor Dashboard JavaScript

let currentSection = 'overview';
let editingGeneratorId = null;
let vendorProfile = null;

document.addEventListener('DOMContentLoaded', async () => {
  if (!isLoggedIn()) {
    window.location.href = '/login';
    return;
  }

  const user = getUser();
  if (user && user.role !== 'vendor' && user.role !== 'admin') {
    // New vendor — has customer role initially, allow them to setup
  }

  document.getElementById('vendorName').textContent = user ? user.name : 'Vendor';

  await loadVendorProfile();
  loadOverview();
});

function showSection(name) {
  document.querySelectorAll('[id^="section-"]').forEach(el => el.style.display = 'none');
  document.querySelectorAll('.sidebar-nav a').forEach(a => a.classList.remove('active'));

  document.getElementById(`section-${name}`).style.display = 'block';
  const navEl = document.getElementById(`nav-${name}`);
  if (navEl) navEl.classList.add('active');

  currentSection = name;

  if (name === 'generators') {
    if (typeof loadEquipmentStats !== 'undefined') {
      loadEquipmentStats();
    } else {
      loadMyGenerators();
    }
  }
  if (name === 'bookings') loadBookings();
  if (name === 'wallet') loadWallet();
  if (name === 'profile') populateProfileForm();
}

// ---- Vendor Profile ----
async function loadVendorProfile() {
  try {
    vendorProfile = await api.getMyVendorProfile();
  } catch {
    vendorProfile = null;
    document.getElementById('setupBanner').style.display = 'block';
  }
}

function populateProfileForm() {
  if (!vendorProfile) return;
  document.getElementById('companyName').value = vendorProfile.company_name || '';
  document.getElementById('vendorCity').value = vendorProfile.city || '';
  document.getElementById('vendorPhone').value = vendorProfile.phone || '';
  document.getElementById('vendorAddress').value = vendorProfile.address || '';
  document.getElementById('vendorDescription').value = vendorProfile.description || '';
  if (vendorProfile.latitude && vendorProfile.longitude) {
    document.getElementById('vendorLatitude').value = vendorProfile.latitude;
    document.getElementById('vendorLongitude').value = vendorProfile.longitude;
    document.getElementById('vendorLocationDisplay').value = `${vendorProfile.latitude.toFixed(5)}, ${vendorProfile.longitude.toFixed(5)}`;
  }
}

function captureVendorLocation() {
  const btn = document.getElementById('vendorLocBtn');
  const display = document.getElementById('vendorLocationDisplay');
  if (!navigator.geolocation) {
    display.value = 'Geolocation not supported by your browser';
    return;
  }
  btn.textContent = '⏳ Detecting...';
  btn.disabled = true;
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      document.getElementById('vendorLatitude').value = pos.coords.latitude;
      document.getElementById('vendorLongitude').value = pos.coords.longitude;
      display.value = `${pos.coords.latitude.toFixed(5)}, ${pos.coords.longitude.toFixed(5)}`;
      btn.textContent = '✅ Detected';
      btn.disabled = false;
    },
    () => {
      display.value = 'Could not detect location. Please allow location access.';
      btn.textContent = '📍 Detect';
      btn.disabled = false;
    }
  );
}

function captureGenLocation() {
  const btn = document.getElementById('genLocBtn');
  const display = document.getElementById('genLocationDisplay');
  if (!navigator.geolocation) {
    display.value = 'Geolocation not supported by your browser';
    return;
  }
  btn.textContent = '⏳ Detecting...';
  btn.disabled = true;
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      document.getElementById('genLatitude').value = pos.coords.latitude;
      document.getElementById('genLongitude').value = pos.coords.longitude;
      display.value = `${pos.coords.latitude.toFixed(5)}, ${pos.coords.longitude.toFixed(5)}`;
      btn.textContent = '✅ Detected';
      btn.disabled = false;
    },
    () => {
      display.value = 'Could not detect location. Please allow location access.';
      btn.textContent = '📍 Detect';
      btn.disabled = false;
    }
  );
}

let isSavingProfile = false; // Prevent duplicate submissions

async function saveVendorProfile(event) {
  event.preventDefault();

  // Prevent duplicate submissions
  if (isSavingProfile) {
    showAlert('profileAlert', 'Please wait, saving profile...', 'error');
    return;
  }

  clearAlert('profileAlert');

  // Clear previous field errors
  ['companyName', 'vendorCity', 'vendorPhone', 'vendorAddress', 'vendorDescription'].forEach(clearFieldError);

  const companyName = document.getElementById('companyName').value.trim();
  const city = document.getElementById('vendorCity').value;
  const phone = document.getElementById('vendorPhone').value.trim();
  const address = document.getElementById('vendorAddress').value.trim();
  const description = document.getElementById('vendorDescription').value.trim();

  let hasError = false;

  // Validate company name
  if (!companyName) {
    showFieldError('companyName', 'Company name is required');
    hasError = true;
  } else if (companyName.length < 3) {
    showFieldError('companyName', 'Company name must be at least 3 characters');
    hasError = true;
  }

  // Validate city
  if (!city) {
    showFieldError('vendorCity', 'Please select a city');
    hasError = true;
  }

  // Validate phone (optional but if provided, must be valid)
  if (phone && !validatePhoneNumber(phone)) {
    showFieldError('vendorPhone', 'Invalid phone number (use 10-digit format like 9876543210)');
    hasError = true;
  }

  if (hasError) {
    showAlert('profileAlert', 'Please fix the validation errors below', 'error');
    return;
  }

  const lat = parseFloat(document.getElementById('vendorLatitude').value);
  const lng = parseFloat(document.getElementById('vendorLongitude').value);
  const data = {
    company_name: companyName,
    city: city,
    phone: phone,
    address: address,
    description: description,
    latitude: isNaN(lat) ? 0 : lat,
    longitude: isNaN(lng) ? 0 : lng,
  };

  const btn = document.getElementById('saveProfileBtn');
  btn.disabled = true;
  btn.textContent = 'Saving...';
  isSavingProfile = true;

  try {
    if (vendorProfile) {
      vendorProfile = await api.updateVendorProfile(data);
    } else {
      try {
        vendorProfile = await api.createVendor(data);
        // Update user in storage to vendor role
        const user = getUser();
        if (user) { user.role = 'vendor'; setUser(user); }
      } catch (createErr) {
        if (createErr.message && createErr.message.includes('already exists')) {
          // Profile exists but wasn't loaded (e.g. page load failed) — fall back to update
          vendorProfile = await api.updateVendorProfile(data);
        } else {
          throw createErr;
        }
      }
    }
    showAlert('profileAlert', '✅ Profile saved successfully!', 'success');
    document.getElementById('setupBanner').style.display = 'none';
    showToast('Profile saved!', 'success');
  } catch (err) {
    showAlert('profileAlert', err.message, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Save Profile';
    isSavingProfile = false;
  }
}

// ---- Wallet ----
async function loadWallet() {
  const container = document.getElementById('walletTransactions');
  container.innerHTML = '<div style="text-align:center;padding:2rem;color:var(--text-light);">Loading...</div>';
  loadBankAccounts();
  loadWithdrawalHistory();

  try {
    const data = await api.getVendorWallet();
    document.getElementById('walletBalance').textContent = formatCurrency(data.balance || 0);
    document.getElementById('walletHoldBalance').textContent = formatCurrency(data.hold_balance || 0);
    document.getElementById('walletWithdrawalHold').textContent = formatCurrency(data.withdrawal_hold_balance || 0);

    const txns = data.transactions || [];
    if (txns.length === 0) {
      container.innerHTML = `
        <div class="empty-state">
          <div class="empty-icon">💰</div>
          <div class="empty-title">No transactions yet</div>
          <p>Earnings will appear here when customers pay advances</p>
        </div>`;
      return;
    }

    const txnMeta = {
      credit:                { label: 'Credit',              color: '#16a34a', prefix: '+', badge: 'available' },
      escrow_release:        { label: 'Escrow Release',      color: '#16a34a', prefix: '+', badge: 'available' },
      withdrawal_refund:     { label: 'Withdrawal Refund',   color: '#16a34a', prefix: '+', badge: 'available' },
      escrow_hold:           { label: 'Escrow Hold',         color: '#6366f1', prefix: '🔒', badge: 'booked' },
      withdrawal_hold:       { label: 'Payout Hold',         color: '#0ea5e9', prefix: '⏳', badge: 'booked' },
      debit:                 { label: 'Debit',               color: '#dc2626', prefix: '-', badge: 'maintenance' },
      withdrawal_completed:  { label: 'Payout Completed',    color: '#dc2626', prefix: '✅', badge: 'maintenance' },
    };

    container.innerHTML = `
      <table class="data-table">
        <thead><tr><th>Date</th><th>Description</th><th>Type</th><th>Amount</th></tr></thead>
        <tbody>
          ${txns.map(t => {
            const m = txnMeta[t.type] || { label: t.type, color: '#6b7280', prefix: '', badge: '' };
            return `<tr>
              <td>${formatDate(t.created_at)}</td>
              <td style="font-size:0.875rem;">${t.description}</td>
              <td><span class="status-badge status-${m.badge}">${m.label}</span></td>
              <td style="font-weight:700;color:${m.color};">${m.prefix} ${formatCurrency(t.amount)}</td>
            </tr>`;
          }).join('')}
        </tbody>
      </table>`;
  } catch (err) {
    container.innerHTML = `<div class="alert alert-error">${err.message}</div>`;
  }
}

// ---- Bank Accounts ----
async function loadBankAccounts() {
  const container = document.getElementById('bankAccountsList');
  const select = document.getElementById('withdrawBankAccountId');
  try {
    const data = await api.get('/wallet/bank-accounts', true);
    const accounts = data.accounts || [];

    // Populate dropdown
    select.innerHTML = '<option value="">— select account —</option>' +
      accounts.map(a => `<option value="${a.id}">${a.bank_name} · ${a.account_name} · ...${a.account_no.slice(-4)}${a.is_primary ? ' ★' : ''}</option>`).join('');

    if (!accounts.length) {
      container.innerHTML = `<div style="color:var(--text-light);font-size:0.875rem;">No bank accounts added yet</div>`;
      return;
    }
    container.innerHTML = accounts.map(a => `
      <div style="display:flex;justify-content:space-between;align-items:center;padding:0.6rem 0;border-bottom:1px solid var(--border);">
        <div>
          <div style="font-weight:600;font-size:0.875rem;">${a.bank_name} ${a.is_primary ? '<span style="color:#f97316;font-size:0.75rem;">★ Primary</span>' : ''}</div>
          <div style="font-size:0.78rem;color:#6b7280;">${a.account_name} · ···${a.account_no.slice(-4)} · ${a.ifsc}</div>
        </div>
        <button onclick="deleteBankAccount(${a.id})" style="background:none;border:none;cursor:pointer;color:#dc2626;font-size:1rem;" title="Remove">🗑</button>
      </div>`).join('');
  } catch (err) {
    container.innerHTML = `<div style="color:#dc2626;font-size:0.875rem;">${err.message}</div>`;
  }
}

function showAddBankForm() {
  document.getElementById('addBankForm').style.display = 'block';
  // Set up real-time validation listeners
  setupBankFieldValidation();
}
function hideAddBankForm() {
  document.getElementById('addBankForm').style.display = 'none';
  ['newBankName','newAccountName','newAccountNo','newIFSC'].forEach(id => {
    const field = document.getElementById(id);
    field.value = '';
    field.style.borderColor = '';
    // Hide hints
    const hint = field.parentElement.querySelector('.field-hint');
    if (hint) hint.style.display = 'none';
  });
  document.getElementById('newIsPrimary').checked = false;
}

// Real-time validation for bank account fields
function setupBankFieldValidation() {
  const bankNameField = document.getElementById('newBankName');
  const accountNameField = document.getElementById('newAccountName');
  const accountNoField = document.getElementById('newAccountNo');
  const ifscField = document.getElementById('newIFSC');

  if (!bankNameField) return;

  // Helper function to show hint and hide placeholder
  function setupFloatingHint(input, hintSelector) {
    const hint = input.parentElement.querySelector(hintSelector);
    if (!hint) return;

    input.addEventListener('focus', function() {
      if (!this.value) {
        hint.style.display = 'block';
      }
    });

    input.addEventListener('input', function() {
      if (this.value) {
        hint.style.display = 'block';
      }
    });

    input.addEventListener('blur', function() {
      if (!this.value) {
        hint.style.display = 'none';
      }
    });
  }

  // Setup floating hints for all fields
  setupFloatingHint(bankNameField, '.field-hint');
  setupFloatingHint(accountNameField, '.field-hint');
  setupFloatingHint(accountNoField, '.field-hint');
  setupFloatingHint(ifscField, '.field-hint');

  // Bank name validation
  bankNameField.addEventListener('input', function() {
    const value = this.value.trim();
    if (value && validateBankName(value)) {
      this.style.borderColor = '#16a34a'; // Green - Valid
    } else if (value) {
      this.style.borderColor = '#dc2626'; // Red - Invalid
    } else {
      this.style.borderColor = ''; // Empty - default
    }
  });

  // Account holder name validation
  accountNameField.addEventListener('input', function() {
    const value = this.value.trim();
    if (value && validateAccountHolderName(value)) {
      this.style.borderColor = '#16a34a'; // Green - Valid
    } else if (value) {
      this.style.borderColor = '#dc2626'; // Red - Invalid
    } else {
      this.style.borderColor = ''; // Empty - default
    }
  });

  // Account number validation - allow only digits
  accountNoField.addEventListener('input', function() {
    // Remove non-digits
    const digitsOnly = this.value.replace(/\D/g, '');
    this.value = digitsOnly;

    const value = this.value.trim();
    if (value && validateAccountNumber(value)) {
      this.style.borderColor = '#16a34a'; // Green - Valid
    } else if (value && value.length >= 9) {
      this.style.borderColor = '#f59e0b'; // Yellow - Close but not quite
    } else if (value) {
      this.style.borderColor = '#dc2626'; // Red - Invalid
    } else {
      this.style.borderColor = ''; // Empty - default
    }
  });

  // IFSC validation - auto-uppercase
  ifscField.addEventListener('input', function() {
    // Auto-uppercase
    this.value = this.value.toUpperCase();

    const value = this.value.trim();
    if (value && validateIFSC(value)) {
      this.style.borderColor = '#16a34a'; // Green - Valid
    } else if (value && value.length >= 8) {
      this.style.borderColor = '#f59e0b'; // Yellow - Close
    } else if (value) {
      this.style.borderColor = '#dc2626'; // Red - Invalid
    } else {
      this.style.borderColor = ''; // Empty - default
    }
  });
}

// Real-time validation for profile fields
function setupProfileFieldValidation() {
  const companyNameField = document.getElementById('companyName');
  const phoneField = document.getElementById('vendorPhone');

  if (!companyNameField) return;

  // Setup floating hints
  companyNameField.parentElement.querySelectorAll('.field-hint').forEach(hint => {
    companyNameField.addEventListener('focus', function() {
      if (!this.value) hint.style.display = 'block';
    });
    companyNameField.addEventListener('input', function() {
      if (this.value) hint.style.display = 'block';
    });
    companyNameField.addEventListener('blur', function() {
      if (!this.value) hint.style.display = 'none';
    });
  });

  // Company name validation
  companyNameField.addEventListener('input', function() {
    const value = this.value.trim();
    if (value && value.length >= 3) {
      this.style.borderColor = '#16a34a'; // Green - Valid
    } else if (value) {
      this.style.borderColor = '#f59e0b'; // Yellow - Needs more chars
    } else {
      this.style.borderColor = ''; // Empty - default
    }
  });

  // Phone validation with floating hint
  if (phoneField) {
    const phoneHint = phoneField.parentElement.querySelector('.field-hint');
    if (phoneHint) {
      phoneField.addEventListener('focus', function() {
        if (!this.value) phoneHint.style.display = 'block';
      });
      phoneField.addEventListener('input', function() {
        if (this.value) phoneHint.style.display = 'block';
      });
      phoneField.addEventListener('blur', function() {
        if (!this.value) phoneHint.style.display = 'none';
      });
    }

    phoneField.addEventListener('input', function() {
      const value = this.value.trim();
      if (!value) {
        this.style.borderColor = ''; // Empty - default
      } else if (validatePhoneNumber(value)) {
        this.style.borderColor = '#16a34a'; // Green - Valid
      } else {
        this.style.borderColor = '#dc2626'; // Red - Invalid
      }
    });
  }
}

// Call setup when profile section is shown
const originalShowSection = showSection;
showSection = function(name) {
  originalShowSection(name);
  if (name === 'profile') {
    setTimeout(setupProfileFieldValidation, 100);
  }
};

// Validation helper functions
function validateIFSC(ifsc) {
  // IFSC should be 11 characters: 4 letters (bank code) + 7 characters (branch code)
  // Format: XXXX0123456 (e.g., HDFC0001234, SBIN0001234)
  const ifscPattern = /^[A-Z]{4}[0-9A-Z]{7}$/;
  return ifscPattern.test(ifsc.toUpperCase());
}

function validateAccountNumber(accountNo) {
  // Account number should be numeric, 9-18 digits
  const accountPattern = /^\d{9,18}$/;
  return accountPattern.test(accountNo);
}

function validateBankName(bankName) {
  // Bank name should be at least 3 characters, letters and spaces only
  const namePattern = /^[a-zA-Z\s]{3,}$/;
  return namePattern.test(bankName.trim());
}

function validateAccountHolderName(accountName) {
  // Account holder name should be at least 3 characters
  return accountName.trim().length >= 3;
}

function validatePhoneNumber(phone) {
  // Indian phone number: starts with 6-9, total 10 digits
  // Can optionally have +91 prefix or spaces
  const phonePattern = /^(\+91[-\s]?)?[6-9]\d{9}$/;
  return phonePattern.test(phone.replace(/\s/g, ''));
}

// Show field-level error
function showFieldError(fieldId, message) {
  const field = document.getElementById(fieldId);
  if (!field) return;

  // Remove existing error
  const existingError = field.parentElement.querySelector('.field-error');
  if (existingError) existingError.remove();

  // Add error message
  const error = document.createElement('div');
  error.className = 'field-error';
  error.style.cssText = 'color: #dc2626; font-size: 0.75rem; margin-top: 0.25rem; font-weight: 500;';
  error.textContent = message;
  field.parentElement.appendChild(error);

  // Add red border to field
  field.style.borderColor = '#dc2626';
}

// Clear field-level error
function clearFieldError(fieldId) {
  const field = document.getElementById(fieldId);
  if (!field) return;

  const existingError = field.parentElement.querySelector('.field-error');
  if (existingError) existingError.remove();

  // Don't reset border - let the validation state show through
}

async function saveBankAccount() {
  const bankName    = document.getElementById('newBankName').value.trim();
  const accountName = document.getElementById('newAccountName').value.trim();
  const accountNo   = document.getElementById('newAccountNo').value.trim();
  const ifsc        = document.getElementById('newIFSC').value.trim().toUpperCase();
  const isPrimary   = document.getElementById('newIsPrimary').checked;

  // Clear previous field errors
  ['newBankName', 'newAccountName', 'newAccountNo', 'newIFSC'].forEach(clearFieldError);

  let hasError = false;

  // Validate bank name
  if (!bankName) {
    showFieldError('newBankName', 'Bank name is required');
    hasError = true;
  } else if (!validateBankName(bankName)) {
    showFieldError('newBankName', 'Enter a valid bank name (min 3 characters)');
    hasError = true;
  }

  // Validate account holder name
  if (!accountName) {
    showFieldError('newAccountName', 'Account holder name is required');
    hasError = true;
  } else if (!validateAccountHolderName(accountName)) {
    showFieldError('newAccountName', 'Enter a valid name (min 3 characters)');
    hasError = true;
  }

  // Validate account number
  if (!accountNo) {
    showFieldError('newAccountNo', 'Account number is required');
    hasError = true;
  } else if (!validateAccountNumber(accountNo)) {
    showFieldError('newAccountNo', 'Account number must be 9-18 digits');
    hasError = true;
  }

  // Validate IFSC
  if (!ifsc) {
    showFieldError('newIFSC', 'IFSC code is required');
    hasError = true;
  } else if (!validateIFSC(ifsc)) {
    showFieldError('newIFSC', 'Invalid IFSC format (e.g., HDFC0001234)');
    hasError = true;
  }

  if (hasError) {
    showAlert('bankFormAlert', 'Please fix the validation errors below', 'error');
    return;
  }

  try {
    await api.post('/wallet/bank-accounts', { bank_name: bankName, account_name: accountName, account_no: accountNo, ifsc, is_primary: isPrimary }, true);
    hideAddBankForm();
    loadBankAccounts();
    showAlert('bankFormAlert', '', 'success'); // Clear alert on success
  } catch (err) {
    showAlert('bankFormAlert', err.message, 'error');
  }
}

async function deleteBankAccount(id) {
  if (!confirm('Remove this bank account?')) return;
  try {
    await api.delete(`/wallet/bank-accounts/${id}`, true);
    loadBankAccounts();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

let pendingWithdrawalId = null;

async function requestWithdrawal() {
  const bankAccountId = parseInt(document.getElementById('withdrawBankAccountId').value);
  const amount = parseFloat(document.getElementById('withdrawAmount').value);

  if (!bankAccountId) return showAlert('withdrawAlert', 'Please select a bank account', 'error');
  if (!amount || amount < 1000) return showAlert('withdrawAlert', 'Minimum withdrawal is ₹1,000', 'error');

  try {
    const result = await api.post('/wallet/withdraw', { amount, bank_account_id: bankAccountId }, true);
    pendingWithdrawalId = result.withdrawal_id;
    document.getElementById('withdrawStep1').style.display = 'none';
    document.getElementById('withdrawStep2').style.display = 'block';
    showAlert('withdrawAlert', result.message, 'success');
  } catch (err) {
    showAlert('withdrawAlert', err.message, 'error');
  }
}

async function confirmWithdrawalOTP() {
  const otp = document.getElementById('withdrawOTP').value.trim();
  if (otp.length !== 6) return showAlert('withdrawAlert', 'Please enter the 6-digit OTP', 'error');
  if (!pendingWithdrawalId) return showAlert('withdrawAlert', 'No pending withdrawal — please start again', 'error');

  try {
    const result = await api.post(`/wallet/withdraw/${pendingWithdrawalId}/confirm`, { otp }, true);
    showAlert('withdrawAlert', result.message, 'success');
    pendingWithdrawalId = null;
    document.getElementById('withdrawStep2').style.display = 'none';
    document.getElementById('withdrawStep1').style.display = 'block';
    ['withdrawAmount', 'withdrawBankAccountId', 'withdrawOTP'].forEach(id => {
      const el = document.getElementById(id);
      if (el) el.value = '';
    });
    loadWallet();
  } catch (err) {
    showAlert('withdrawAlert', err.message, 'error');
  }
}

function cancelWithdrawalOTP() {
  pendingWithdrawalId = null;
  document.getElementById('withdrawStep2').style.display = 'none';
  document.getElementById('withdrawStep1').style.display = 'block';
  document.getElementById('withdrawOTP').value = '';
}

// Timeline steps per status
const WITHDRAWAL_STEPS = [
  { key: 'otp_sent',   label: 'OTP Requested' },
  { key: 'confirmed',  label: 'OTP Confirmed' },
  { key: 'review',     label: 'Admin Review'  },
  { key: 'done',       label: 'Transfer Done' },
];
const WITHDRAWAL_STEP_MAP = {
  otp_pending: 0,
  expired:     0,
  pending:     2,
  approved:    3,
  paid:        3,
  rejected:    2,
};

function withdrawalTimeline(w) {
  const currentStep = WITHDRAWAL_STEP_MAP[w.status] ?? 0;
  const isRejected  = w.status === 'rejected';
  const isExpired   = w.status === 'expired';
  return `<div style="display:flex;align-items:center;gap:0;margin:0.5rem 0;">
    ${WITHDRAWAL_STEPS.map((s, i) => {
      const done   = i < currentStep || (w.status === 'paid' || w.status === 'approved');
      const active = i === currentStep && !isRejected && !isExpired;
      const fail   = (isRejected && i === 2) || (isExpired && i === 0);
      const color  = fail ? '#dc2626' : done || active ? '#16a34a' : '#d1d5db';
      const icon   = fail ? '✗' : done ? '✓' : active ? '●' : '○';
      return `<span style="display:flex;align-items:center;">
        <span style="width:22px;height:22px;border-radius:50%;background:${color};color:white;
          display:flex;align-items:center;justify-content:center;font-size:0.7rem;font-weight:700;flex-shrink:0;">${icon}</span>
        <span style="font-size:0.68rem;color:${active || done ? '#374151' : '#9ca3af'};margin:0 4px;white-space:nowrap;">${s.label}</span>
        ${i < WITHDRAWAL_STEPS.length - 1 ? `<span style="flex:1;height:2px;background:${done ? '#16a34a' : '#e5e7eb'};min-width:8px;"></span>` : ''}
      </span>`;
    }).join('')}
  </div>`;
}

async function loadWithdrawalHistory() {
  const container = document.getElementById('withdrawalHistory');
  if (!container) return;
  try {
    const data = await api.get('/wallet/withdrawals', true);
    const list = data.withdrawals || [];
    if (!list.length) {
      container.innerHTML = `<div style="text-align:center;padding:2rem;color:var(--text-light);">No withdrawal requests yet</div>`;
      return;
    }
    container.innerHTML = list.map(w => {
      const statusLabel = {
        otp_pending: { text: 'Awaiting OTP',   color: '#6366f1' },
        expired:     { text: 'OTP Expired',    color: '#9ca3af' },
        pending:     { text: 'Admin Review',   color: '#f59e0b' },
        approved:    { text: 'Approved',       color: '#16a34a' },
        rejected:    { text: 'Rejected',       color: '#dc2626' },
        paid:        { text: 'Paid',           color: '#059669' },
      }[w.status] || { text: w.status, color: '#6b7280' };

      return `<div style="border:1px solid var(--border);border-radius:var(--radius);padding:1rem;margin-bottom:0.75rem;">
        <div style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:0.5rem;">
          <div>
            <span style="font-size:1.1rem;font-weight:800;">₹${w.amount.toLocaleString('en-IN')}</span>
            <span style="font-size:0.75rem;color:#6b7280;margin-left:0.5rem;">#${w.id} · ${formatDate(w.created_at)}</span>
          </div>
          <span style="font-size:0.75rem;font-weight:700;color:${statusLabel.color};">${statusLabel.text}</span>
        </div>
        <div style="font-size:0.8rem;color:#374151;margin-bottom:0.5rem;">
          ${w.bank_name} · ${w.account_name} · ···${w.account_no.slice(-4)} · ${w.ifsc}
        </div>
        ${withdrawalTimeline(w)}
        ${w.admin_note ? `<div style="font-size:0.75rem;color:#6b7280;margin-top:0.25rem;">Note: ${w.admin_note}</div>` : ''}
      </div>`;
    }).join('');
  } catch (err) {
    container.innerHTML = `<div class="alert alert-error">${err.message}</div>`;
  }
}

// ---- Overview ----
async function loadOverview() {
  try {
    const bookingsData = await api.getMyBookings();
    const bookings = bookingsData.bookings || [];

    const generatorsData = await api.getMyGenerators();
    const generators = generatorsData.generators || [];

    const available = generators.filter(g => g.availability_status === 'available').length;
    const pending = bookings.filter(b => b.status === 'requested').length;

    document.getElementById('statTotalGen').textContent = generators.length;
    document.getElementById('statAvailGen').textContent = available;
    document.getElementById('statPendingBook').textContent = pending;
    document.getElementById('statTotalBook').textContent = bookings.length;

    renderRecentBookings(bookings.slice(0, 5));
  } catch (err) {
    console.error('Failed to load overview:', err);
  }
}

function renderRecentBookings(bookings) {
  const container = document.getElementById('recentBookings');
  if (!bookings || bookings.length === 0) {
    container.innerHTML = `<div class="empty-state"><div class="empty-icon">📋</div><div class="empty-title">No bookings yet</div><p>Bookings will appear here</p></div>`;
    return;
  }

  container.innerHTML = `
    <table class="data-table">
      <thead>
        <tr>
          <th>#ID</th><th>Customer</th><th>Equipment</th><th>Dates</th><th>Amount</th><th>Status</th>
        </tr>
      </thead>
      <tbody>
        ${bookings.map(b => `
          <tr>
            <td>#${b.id}</td>
            <td>${b.customer ? b.customer.name : 'Customer'}</td>
            <td>${b.generator ? b.generator.name : 'Equipment'}</td>
            <td>${formatDate(b.start_date)} – ${formatDate(b.end_date)}</td>
            <td>${formatCurrency(b.total_price)}</td>
            <td>${statusBadge(b.status)}</td>
          </tr>
        `).join('')}
      </tbody>
    </table>
  `;
}

// ---- Generators ----
async function loadMyGenerators() {
  const container = document.getElementById('generatorsList');
  if (!container) return;
  container.innerHTML = '<div style="text-align:center;padding:3rem;color:var(--text-light);">Loading...</div>';

  try {
    const data = await api.getMyGenerators();
    const generators = data.generators || [];

    if (generators.length === 0) {
      container.innerHTML = `
        <div class="empty-state">
          <div class="empty-icon">⚡</div>
          <div class="empty-title">No equipment yet</div>
          <p>Add your first equipment to start receiving bookings</p>
          <button class="btn btn-primary" style="margin-top:1rem;" onclick="openAddGenerator()">+ Add Equipment</button>
        </div>
      `;
      return;
    }

    container.innerHTML = `
      <div class="generators-grid">
        ${generators.map(gen => `
          <div class="generator-card">
            <div class="card-image">
              ${gen.image_url ? `<img src="${gen.image_url}" alt="${gen.name}" />` : '<span class="gen-icon">⚡</span>'}
              <span class="capacity-badge">${gen.capacity_kva} kVA</span>
            </div>
            <div class="card-body">
              <div class="card-title">${gen.name}</div>
              <div class="card-meta">
                <span class="meta-tag">📍 ${gen.city}</span>
                <span class="meta-tag">⛽ ${gen.fuel_type}</span>
                ${gen.brand ? `<span class="meta-tag">🏷️ ${gen.brand}</span>` : ''}
              </div>
              <div class="card-footer">
                <div class="card-price">${formatCurrency(gen.price_per_day)} <span>/ day</span></div>
                ${statusBadge(gen.availability_status)}
              </div>
              <div style="display:flex;gap:0.5rem;margin-top:1rem;">
                <button class="btn btn-sm btn-primary" onclick="openEditGenerator(${JSON.stringify(gen).replace(/"/g, '&quot;')})">Edit</button>
                <button class="btn btn-sm btn-danger" onclick="deleteGenerator(${gen.id}, '${gen.name}')">Delete</button>
              </div>
            </div>
          </div>
        `).join('')}
      </div>
    `;
  } catch (err) {
    container.innerHTML = `<div class="alert alert-error">${err.message}</div>`;
  }
}

function openAddGenerator() {
  editingGeneratorId = null;
  document.getElementById('modalTitle').textContent = 'Add Equipment';
  document.getElementById('generatorForm').reset();
  document.getElementById('generatorId').value = '';
  document.getElementById('genLatitude').value = '';
  document.getElementById('genLongitude').value = '';
  document.getElementById('genLocationDisplay').value = '';
  document.getElementById('genLocBtn').textContent = '📍 Detect';
  document.getElementById('modalAlert').innerHTML = '';
  document.getElementById('generatorModal').classList.remove('hidden');
}

function openEditGenerator(gen) {
  editingGeneratorId = gen.id;
  document.getElementById('modalTitle').textContent = 'Edit Equipment';
  document.getElementById('generatorId').value = gen.id;
  document.getElementById('genName').value = gen.name;
  document.getElementById('genCapacity').value = gen.capacity_kva;
  document.getElementById('genBrand').value = gen.brand || '';
  document.getElementById('genPriceDay').value = gen.price_per_day;
  document.getElementById('genPriceMonth').value = gen.price_per_month || '';
  document.getElementById('genFuel').value = gen.fuel_type || 'diesel';
  document.getElementById('genCity').value = gen.city;
  document.getElementById('genLocation').value = gen.location;
  document.getElementById('genDesc').value = gen.description || '';
  document.getElementById('genLatitude').value = gen.latitude || '';
  document.getElementById('genLongitude').value = gen.longitude || '';
  document.getElementById('genLocationDisplay').value = (gen.latitude && gen.longitude)
    ? `${Number(gen.latitude).toFixed(5)}, ${Number(gen.longitude).toFixed(5)}` : '';
  document.getElementById('modalAlert').innerHTML = '';
  document.getElementById('generatorModal').classList.remove('hidden');
}

function closeModal() {
  document.getElementById('generatorModal').classList.add('hidden');
}

async function saveGenerator(event) {
  event.preventDefault();
  clearAlert('modalAlert');

  const genLat = parseFloat(document.getElementById('genLatitude').value);
  const genLng = parseFloat(document.getElementById('genLongitude').value);
  const data = {
    name: document.getElementById('genName').value,
    capacity_kva: parseInt(document.getElementById('genCapacity').value),
    brand: document.getElementById('genBrand').value,
    price_per_day: parseFloat(document.getElementById('genPriceDay').value),
    price_per_month: parseFloat(document.getElementById('genPriceMonth').value) || 0,
    fuel_type: document.getElementById('genFuel').value,
    city: document.getElementById('genCity').value,
    location: document.getElementById('genLocation').value,
    description: document.getElementById('genDesc').value,
    latitude: isNaN(genLat) ? 0 : genLat,
    longitude: isNaN(genLng) ? 0 : genLng,
  };

  const btn = document.getElementById('saveGenBtn');
  btn.disabled = true;
  btn.textContent = 'Saving...';

  try {
    if (editingGeneratorId) {
      await api.updateGenerator(editingGeneratorId, data);
      showToast('Generator updated!', 'success');
    } else {
      await api.createGenerator(data);
      showToast('Generator added!', 'success');
    }
    closeModal();
    loadMyGenerators();
  } catch (err) {
    showAlert('modalAlert', err.message, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Save Generator';
  }
}

async function deleteGenerator(id, name) {
  if (!confirm(`Delete generator "${name}"? This cannot be undone.`)) return;
  try {
    await api.deleteGenerator(id);
    showToast('Generator deleted', 'success');
    loadMyGenerators();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// ---- Bookings ----
async function loadBookings() {
  const tbody = document.getElementById('bookingsTableBody');
  tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;padding:2rem;color:var(--text-light);">Loading...</td></tr>';

  try {
    const data = await api.getMyBookings();
    const bookings = data.bookings || [];

    if (bookings.length === 0) {
      tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;padding:2rem;color:var(--text-light);">No bookings yet</td></tr>';
      return;
    }

    tbody.innerHTML = bookings.map(b => `
      <tr>
        <td>#${b.id}</td>
        <td>
          <div style="font-weight:600;">${b.customer ? b.customer.name : 'N/A'}</div>
          <div style="font-size:0.75rem;color:var(--text-light);">${b.customer ? b.customer.phone : ''}</div>
        </td>
        <td>${b.generator ? b.generator.name : 'N/A'}</td>
        <td>
          <div>${formatDate(b.start_date)}</div>
          <div style="font-size:0.75rem;color:var(--text-light);">to ${formatDate(b.end_date)}</div>
        </td>
        <td>
          <div>${formatCurrency(b.total_price)}</div>
          <div style="font-size:0.72rem;color:#f97316;">Advance: ${formatCurrency(b.advance_amount)}</div>
        </td>
        <td>${statusBadge(b.status)}</td>
        <td>
          <div style="display:flex;gap:0.4rem;flex-wrap:wrap;">
            ${b.status === 'requested' ? `
              <button class="btn btn-sm btn-success" onclick="vendorAction(${b.id},'accept')">✅ Accept</button>
              <button class="btn btn-sm btn-danger" onclick="vendorAction(${b.id},'reject')">❌ Reject</button>
            ` : ''}
            ${b.status === 'accepted' ? `
              <span style="font-size:0.75rem;color:#f97316;">⏳ Awaiting customer payment</span>
            ` : ''}
            ${b.status === 'advance_paid' ? `
              <button class="btn btn-sm btn-primary" onclick="vendorAction(${b.id},'dispatch')">🚚 Dispatch</button>
            ` : ''}
            ${b.status === 'dispatched' ? `
              <span style="font-size:0.75rem;color:#2563eb;">📦 Dispatched — waiting for OTP</span>
              <button class="btn btn-sm" style="background:#f59e0b;color:white;" onclick="resendOTP(${b.id})">🔁 Resend OTP</button>
              <button class="btn btn-sm btn-primary" onclick="uploadHandover(${b.id},'delivery')">📷 Upload Handover</button>
            ` : ''}
            ${b.status === 'delivered' && !b.return_initiated_at ? `
              <button class="btn btn-sm" style="background:#7c3aed;color:white;" onclick="uploadHandover(${b.id},'return')">📷 Return Photos</button>
              <button class="btn btn-sm btn-primary" onclick="initiateReturn(${b.id})">🔄 Initiate Return</button>
            ` : ''}
            ${b.status === 'delivered' && b.return_initiated_at && !b.return_otp_verified ? `
              <span style="font-size:0.75rem;color:#7c3aed;">🔄 Return initiated — ask customer for OTP</span>
              <button class="btn btn-sm" style="background:#7c3aed;color:white;" onclick="confirmReturn(${b.id})">✅ Confirm Return OTP</button>
            ` : ''}
            ${b.status === 'completed' ? `
              <span style="font-size:0.75rem;color:#16a34a;">💰 Payment released to wallet</span>
            ` : ''}
          </div>
        </td>
      </tr>
    `).join('');
  } catch (err) {
    tbody.innerHTML = `<tr><td colspan="7" class="alert alert-error">${err.message}</td></tr>`;
  }
}

async function vendorAction(bookingId, action) {
  const messages = { accept: 'Accept this booking?', reject: 'Reject this booking?', dispatch: 'Mark generator as dispatched? An OTP will be generated for the customer.' };
  if (!confirm(messages[action] || 'Confirm?')) return;
  try {
    const result = await api.post(`/bookings/${bookingId}/${action}`, {}, true);
    if (action === 'dispatch' && result.delivery_otp) {
      alert(`Generator dispatched!\n\nDelivery OTP for customer: ${result.delivery_otp}\n\n(In production this is sent via SMS to the customer)`);
    } else {
      showToast(result.message || `Booking ${action}ed`, 'success');
    }
    loadBookings();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function resendOTP(bookingId) {
  if (!confirm('Resend delivery OTP to the customer via email and notification?')) return;
  try {
    await api.post(`/bookings/${bookingId}/resend-otp`, {}, true);
    showToast('OTP resent to customer successfully', 'success');
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// ─── Feature 5: Handover Modal ───────────────────────────────────────────────

let _handoverBookingId = null;
let _handoverType = null;

function uploadHandover(bookingId, type) {
  _handoverBookingId = bookingId;
  _handoverType = type;
  const label = type === 'delivery' ? 'Delivery' : 'Return';
  document.getElementById('handoverModalTitle').textContent = `📷 Upload ${label} Handover`;
  document.getElementById('handoverModal').classList.remove('hidden');
  // clear previous values
  document.getElementById('handoverPhotoUrls').value = '';
  document.getElementById('handoverCondition').value = 'Good';
  document.getElementById('handoverAccessories').value = 'Yes';
  document.getElementById('handoverFuelLevel').value = 'Full';
  document.getElementById('handoverHoursMeter').value = '';
  document.getElementById('handoverNotes').value = '';
}

function closeHandoverModal() {
  document.getElementById('handoverModal').classList.add('hidden');
  _handoverBookingId = null;
  _handoverType = null;
}

async function submitHandover() {
  if (!_handoverBookingId || !_handoverType) return;
  const photosRaw = document.getElementById('handoverPhotoUrls').value.trim();
  if (!photosRaw) {
    showToast('Please provide at least one photo URL', 'error');
    return;
  }
  const photoUrls = photosRaw.split(',').map(p => p.trim()).filter(Boolean);
  const checklist = {
    condition:            document.getElementById('handoverCondition').value,
    accessories_present:  document.getElementById('handoverAccessories').value,
    fuel_level:           document.getElementById('handoverFuelLevel').value,
    hours_meter:          document.getElementById('handoverHoursMeter').value || 'N/A',
  };
  const notes = document.getElementById('handoverNotes').value.trim();
  const btn = document.getElementById('submitHandoverBtn');
  btn.disabled = true; btn.textContent = 'Uploading...';
  try {
    const label = _handoverType === 'delivery' ? 'Delivery' : 'Return';
    await api.post(`/bookings/${_handoverBookingId}/handover?type=${_handoverType}`, { photo_urls: photoUrls, checklist, notes }, true);
    showToast(`✅ ${label} handover uploaded! Customer notified.`, 'success');
    closeHandoverModal();
    loadBookings();
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    btn.disabled = false; btn.textContent = 'Upload Handover';
  }
}

// ─── Feature 5: Return OTP Modal ─────────────────────────────────────────────

let _returnBookingId = null;

function openReturnOTPModal(bookingId) {
  _returnBookingId = bookingId;
  document.getElementById('returnOTPInput').value = '';
  document.getElementById('returnOTPModal').classList.remove('hidden');
}

function closeReturnOTPModal() {
  document.getElementById('returnOTPModal').classList.add('hidden');
  _returnBookingId = null;
}

// Feature 5: Initiate equipment return flow
async function initiateReturn(bookingId) {
  if (!confirm('Initiate equipment return? A 6-digit OTP will be sent to the customer via notification.')) return;
  try {
    const result = await api.post(`/bookings/${bookingId}/initiate-return`, {}, true);
    showToast(result.message || 'Return initiated. Ask the customer for their OTP.', 'success');
    loadBookings();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// Feature 5: Confirm return OTP entered by vendor
function confirmReturn(bookingId) {
  openReturnOTPModal(bookingId);
}

async function submitReturnOTP() {
  const otp = document.getElementById('returnOTPInput').value.trim();
  if (otp.length !== 6) {
    showToast('Please enter a valid 6-digit OTP', 'error');
    return;
  }
  const btn = document.getElementById('submitReturnOTPBtn');
  btn.disabled = true; btn.textContent = 'Confirming...';
  try {
    const result = await api.post(`/bookings/${_returnBookingId}/confirm-return`, { otp }, true);
    showToast(result.message || '🎉 Return confirmed! Booking completed.', 'success');
    closeReturnOTPModal();
    loadBookings();
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    btn.disabled = false; btn.textContent = 'Confirm Return';
  }
}

