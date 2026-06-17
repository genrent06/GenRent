// GenRent - My Bookings (Customer)

const STATUS_STEPS = ['requested', 'accepted', 'advance_paid', 'dispatched', 'delivered', 'completed'];
const STATUS_LABELS = {
  requested:    'Requested',
  accepted:     'Accepted',
  advance_paid: 'Advance Paid',
  dispatched:   'Dispatched',
  delivered:    'Delivered',
  completed:    'Completed',
  cancelled:    'Cancelled',
};
const STATUS_ICONS = {
  requested:    '📋',
  accepted:     '✅',
  advance_paid: '💳',
  dispatched:   '🚚',
  delivered:    '📦',
  completed:    '🎉',
  cancelled:    '❌',
};

let ratingSelected = {};
let pollingInterval = null;
let currentNotifications = [];

document.addEventListener('DOMContentLoaded', () => {
  if (!isLoggedIn()) {
    window.location.href = '/login?redirect=/my-bookings';
    return;
  }
  loadBookings();
  window.addEventListener('beforeunload', stopPolling);
});

function startPolling(active) {
  stopPolling();
  if (active) {
    pollingInterval = setInterval(() => {
      silentRefresh();
    }, 10000); // poll every 10 seconds
  }
}

function stopPolling() {
  if (pollingInterval) {
    clearInterval(pollingInterval);
    pollingInterval = null;
  }
}

async function silentRefresh() {
  try {
    const data = await api.getMyBookings();
    const bookings = Array.isArray(data) ? data : (data.bookings || []);
    if (!bookings.length) return;

    // Fetch notifications to get Return OTPs if any bookings are delivered
    if (bookings.some(b => b.status === 'delivered')) {
      try {
        const notifRes = await api.getNotifications();
        currentNotifications = notifRes.notifications || [];
      } catch (e) {
        console.error("Failed to load notifications in refresh:", e);
      }
    }

    const activeStatuses = ['requested', 'accepted', 'advance_paid', 'dispatched', 'delivered'];
    const hasActive = bookings.some(b => activeStatuses.includes(b.status));

    // Preserve user-typed values before wiping the DOM
    const savedInputs = {};
    document.querySelectorAll('[id^="otp-"], [id^="review-"]').forEach(el => {
      if (el.value) savedInputs[el.id] = el.value;
    });

    document.getElementById('bookingsList').innerHTML = bookings.map(renderBookingCard).join('');
    loadAllHandovers(bookings);

    // Restore typed values after re-render
    Object.entries(savedInputs).forEach(([id, val]) => {
      const el = document.getElementById(id);
      if (el) el.value = val;
    });

    // Re-apply star highlight for any in-progress ratings
    Object.entries(ratingSelected).forEach(([bookingId, rating]) => {
      const stars = document.querySelectorAll(`#stars-${bookingId} span`);
      stars.forEach(s => {
        s.classList.toggle('active', parseInt(s.dataset.val) <= rating);
      });
    });

    if (!hasActive) stopPolling();
  } catch {
    // silent fail on background poll
  }
}

async function loadBookings() {
  try {
    const data = await api.getMyBookings();
    document.getElementById('loadingState').style.display = 'none';

    const bookings = Array.isArray(data) ? data : (data.bookings || []);

    if (!bookings.length) {
      document.getElementById('bookingsList').innerHTML = `
        <div class="empty-state">
          <div class="empty-icon">📋</div>
          <h3>No bookings yet</h3>
          <p style="margin-top:0.5rem;">Browse equipment and make your first booking!</p>
          <a href="/" class="btn btn-primary" style="margin-top:1rem;">Browse Equipment</a>
        </div>`;
      return;
    }

    // Fetch notifications to get Return OTPs if any bookings are delivered
    if (bookings.some(b => b.status === 'delivered')) {
      try {
        const notifRes = await api.getNotifications();
        currentNotifications = notifRes.notifications || [];
      } catch (e) {
        console.error("Failed to load notifications:", e);
      }
    }

    document.getElementById('bookingsList').innerHTML = bookings.map(renderBookingCard).join('');
    loadAllHandovers(bookings);

    // Auto-poll every 10s if there are active bookings (not completed/cancelled)
    const activeStatuses = ['requested', 'accepted', 'advance_paid', 'dispatched', 'delivered'];
    const hasActive = bookings.some(b => activeStatuses.includes(b.status));
    startPolling(hasActive);
  } catch (err) {
    document.getElementById('loadingState').innerHTML =
      `<div class="alert alert-error">Failed to load bookings: ${err.message}</div>`;
  }
}

function renderTimeline(b) {
  const events = [
    { label: 'Booking Requested', time: b.created_at, icon: '📋' },
    { label: 'Vendor Accepted', time: b.accepted_at, icon: '✅' },
    { label: 'Advance Paid', time: b.advance_paid && b.accepted_at ? b.accepted_at : null, icon: '💳', show: b.advance_paid },
    { label: 'Equipment Dispatched', time: b.dispatched_at, icon: '🚚' },
    { label: 'Delivery Confirmed', time: b.delivered_at, icon: '📦' },
    { label: 'Booking Completed', time: b.completed_at, icon: '🎉' },
  ];

  const completed = events.filter(e => e.time || e.show);
  if (completed.length <= 1) return ''; // Only show if there are at least 2 events

  return `<div style="margin:1rem 0;padding:1rem;background:#f8fafc;border-radius:var(--radius);border-left:3px solid var(--primary);">
    <div style="font-size:0.75rem;font-weight:600;color:var(--text-light);text-transform:uppercase;letter-spacing:0.05em;margin-bottom:0.75rem;">Booking Timeline</div>
    ${events.filter(e => e.time).map(e => `
      <div style="display:flex;align-items:flex-start;gap:0.75rem;margin-bottom:0.5rem;">
        <span style="font-size:1rem;min-width:1.5rem;">${e.icon}</span>
        <div>
          <div style="font-size:0.8rem;font-weight:600;">${e.label}</div>
          <div style="font-size:0.72rem;color:var(--text-light);">${formatDate(e.time)}</div>
        </div>
      </div>
    `).join('')}
  </div>`;
}

function renderBookingCard(b) {
  const gen = b.generator || {};
  const status = b.status;
  const isCancelled = status === 'cancelled';

  const progressSteps = STATUS_STEPS.map((s, i) => {
    let cls = '';
    const currentIdx = STATUS_STEPS.indexOf(status);
    if (isCancelled) {
      cls = i === 0 ? 'done' : (i <= currentIdx ? 'done' : '');
    } else {
      if (i < currentIdx) cls = 'done';
      else if (i === currentIdx) cls = 'active';
    }
    return `<div class="progress-step ${cls}">
      <div class="step-dot">${cls === 'done' ? '✓' : (i + 1)}</div>
      <div>${STATUS_LABELS[s]}</div>
    </div>`;
  });

  if (isCancelled) {
    progressSteps.push(`<div class="progress-step cancelled">
      <div class="step-dot">✕</div>
      <div>Cancelled</div>
    </div>`);
  }

  const statusInfo = buildStatusInfo(b);

  return `
  <div class="booking-card-full" id="booking-${b.id}">
    <div class="booking-header">
      <div>
        <strong>Booking #${b.id}</strong>
        <span style="color:var(--text-light);font-size:0.8rem;margin-left:0.75rem;">
          ${formatDate(b.created_at)}
        </span>
      </div>
      <span class="status-badge status-${status}">
        ${STATUS_ICONS[status] || ''} ${STATUS_LABELS[status] || status}
      </span>
    </div>
    <div class="booking-body">
      <div class="booking-meta">
        <div class="booking-meta-item">
          <div class="label">Equipment</div>
          <div class="value">${gen.name || 'N/A'} ${gen.capacity_kva ? `(${gen.capacity_kva} kVA)` : ''}</div>
        </div>
        <div class="booking-meta-item">
          <div class="label">Rental Period</div>
          <div class="value">${formatDate(b.start_date)} → ${formatDate(b.end_date)}</div>
        </div>
        <div class="booking-meta-item">
          <div class="label">Total Amount</div>
          <div class="value">${formatCurrency(b.total_price)}</div>
        </div>
        <div class="booking-meta-item">
          <div class="label">Advance (30%)</div>
          <div class="value" style="color:${b.advance_paid ? '#16a34a' : '#c2410c'};">
            ${formatCurrency(b.advance_amount)} ${b.advance_paid ? '✓ Paid' : '● Pending'}
          </div>
        </div>
        ${b.address ? `<div class="booking-meta-item">
          <div class="label">Delivery Address</div>
          <div class="value">${b.address}</div>
        </div>` : ''}
      </div>

      <div class="progress-steps">${progressSteps.join('')}</div>

      ${renderTimeline(b)}

      <div id="handover-photos-${b.id}"></div>

      ${statusInfo}

      ${b.cancel_reason ? `<div style="margin-top:1rem;padding:0.75rem;background:#fef2f2;border:1px solid #fecaca;border-radius:var(--radius);font-size:0.875rem;color:#991b1b;">
        <strong>Cancellation reason:</strong> ${b.cancel_reason}
      </div>` : ''}

      ${b.customer_rating > 0 ? `<div style="margin-top:1rem;padding:0.75rem;background:#fefce8;border:1px solid #fde047;border-radius:var(--radius);font-size:0.875rem;">
        <strong>Your rating:</strong> ${'⭐'.repeat(b.customer_rating)} &nbsp;
        ${b.customer_review ? `<em>"${b.customer_review}"</em>` : ''}
      </div>` : ''}
    </div>
  </div>`;
}

function buildStatusInfo(b) {
  switch (b.status) {
    case 'requested':
      return `<div class="info-box">
        ⏳ <strong>Waiting for vendor to accept</strong> — Your booking request has been sent.
        The vendor will accept or reject it. You'll be notified once they respond.
        <div class="booking-actions">
          <button class="btn btn-danger btn-sm" onclick="cancelBooking(${b.id})">Cancel Request</button>
        </div>
      </div>`;

    case 'accepted':
      return `<div class="info-box" style="background:#f0fdf4;border-color:#86efac;color:#166534;">
        ✅ <strong>Vendor accepted your booking!</strong> — Now pay the 30% advance to confirm.
        Your payment goes to escrow and is only released after delivery confirmation.
        <div class="booking-actions">
          <a href="/payment?booking_id=${b.id}" class="btn btn-primary btn-sm">Pay Advance ${formatCurrency(b.advance_amount)} →</a>
          <button class="btn btn-danger btn-sm" onclick="cancelBooking(${b.id})">Cancel</button>
        </div>
      </div>`;

    case 'advance_paid':
      return `<div class="info-box">
        💳 <strong>Advance paid!</strong> — The vendor will now prepare and dispatch the equipment.
        You'll receive a 6-digit OTP once the equipment is dispatched to your location.
      </div>`;

    case 'dispatched':
      return `<div class="otp-section">
        <h4>🚚 Equipment Dispatched — Confirm Delivery</h4>
        <p style="font-size:0.875rem;margin-bottom:1rem;color:#15803d;">
          The equipment is on its way! Enter the 6-digit OTP you received to confirm delivery.
          This releases the payment to the vendor.
        </p>
        <div class="otp-input">
          <input type="text" id="otp-${b.id}" maxlength="6" placeholder="000000"
            oninput="this.value=this.value.replace(/[^0-9]/g,'')" />
          <button class="btn btn-primary" onclick="confirmDelivery(${b.id})">Confirm Delivery</button>
        </div>
        <p style="font-size:0.75rem;color:#6b7280;margin-top:0.5rem;">
          The OTP was shared with the vendor on dispatch. Ask the delivery person for the OTP.
        </p>
      </div>`;

    case 'delivered':
      if (b.return_initiated_at) {
        // Find OTP notification
        const otpNotif = currentNotifications.find(n => n.booking_id === b.id && n.type === 'return_otp');
        let otpStr = '______';
        if (otpNotif) {
          const match = otpNotif.message.match(/\b\d{6}\b/);
          if (match) otpStr = match[0];
        }
        return `<div class="info-box" style="background:#f0fdf4;border-color:#86efac;color:#166534;">
          🔄 <strong>Return Initiated by Vendor</strong>
          <p style="margin-top:0.5rem;font-size:0.875rem;">
            Please share the following OTP with the vendor to confirm return and complete the booking:
          </p>
          <div style="font-size:2rem;font-weight:800;letter-spacing:0.2em;color:#15803d;margin:0.75rem 0;text-align:center;">
            ${otpStr}
          </div>
          ${otpNotif ? `<p style="font-size:0.75rem;color:#166534;margin-top:0.5rem;">${otpNotif.message}</p>` : ''}
        </div>`;
      } else {
        return `<div class="info-box" style="background:#fff7ed;border-color:#fed7aa;color:#c2410c;">
          ⏳ <strong>Equipment In Use</strong>
          <p style="margin-top:0.5rem;font-size:0.875rem;">
            The equipment is currently in your possession. When you are ready to return the equipment, the vendor will initiate the return flow and ask you for a Return OTP.
          </p>
        </div>`;
      }

    case 'completed':
      const hasReviewed = b.customer_rating > 0;
      return `
        <div style="margin-top:1rem;padding:0.75rem;background:#f0fdf4;border:1px solid #86efac;border-radius:var(--radius);font-size:0.875rem;color:#166534;">
          🎉 <strong>Booking Completed!</strong> — Thanks for using GenRent.
          <div style="margin-top:0.5rem;">
            <button class="btn btn-sm" style="background:#ef4444;color:white;font-size:0.75rem;" onclick="openDisputeModal(${b.id})">
              ⚠️ Report Damage
            </button>
            <small style="color:#9ca3af;margin-left:0.5rem;">Available within 48 hours</small>
          </div>
        </div>
        ${!hasReviewed ? `
          <div class="review-section" style="margin-top:1rem;">
            <h4 style="color:#92400e;">⭐ Leave a review for the vendor</h4>
            <p style="font-size:0.875rem;color:#78350f;margin-bottom:0.75rem;">
              Please rate your experience with the equipment and vendor.
            </p>
            <div>
              <label style="font-size:0.875rem;font-weight:600;">Rating</label>
              <div class="star-rating" id="stars-${b.id}">
                ${[1,2,3,4,5].map(n => `<span onclick="setRating(${b.id}, ${n})" data-val="${n}">★</span>`).join('')}
              </div>
            </div>
            <div class="form-group" style="margin-top:0.5rem;">
              <textarea id="review-${b.id}" placeholder="Write a review (optional)..." rows="2"
                style="width:100%;padding:0.5rem;border:1px solid var(--border);border-radius:var(--radius);font-family:inherit;"></textarea>
            </div>
            <div class="booking-actions">
              <button class="btn btn-primary" onclick="submitBookingReview(${b.id})">Submit Review</button>
            </div>
          </div>
        ` : ''}
      `;

    case 'cancelled':
      return '';

    default:
      return '';
  }
}

function setRating(bookingId, value) {
  ratingSelected[bookingId] = value;
  const stars = document.querySelectorAll(`#stars-${bookingId} span`);
  stars.forEach(s => {
    s.classList.toggle('active', parseInt(s.dataset.val) <= value);
  });
}

async function confirmDelivery(bookingId) {
  const otp = document.getElementById(`otp-${bookingId}`).value.trim();
  if (otp.length !== 6) {
    showToast('Please enter the 6-digit OTP', 'error');
    return;
  }
  try {
    const result = await api.post(`/bookings/${bookingId}/confirm-delivery`, { otp }, true);
    showToast(result.message || 'Delivery confirmed!', 'success');
    setTimeout(loadBookings, 800);
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function completeBooking(bookingId) {
  const rating = ratingSelected[bookingId] || 0;
  const review = document.getElementById(`review-${bookingId}`)?.value.trim() || '';

  try {
    const result = await api.post(`/bookings/${bookingId}/complete`, {}, true);
    showToast(result.message || 'Booking completed!', 'success');

    if (rating > 0) {
      try {
        await api.post(`/bookings/${bookingId}/review`, { rating, review }, true);
      } catch {
        // review submission failure is non-critical
      }
    }
    setTimeout(loadBookings, 800);
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function cancelBooking(bookingId) {
  const reason = prompt('Reason for cancellation (optional):');
  if (reason === null) return; // user pressed Cancel on the prompt
  try {
    const result = await apiRequest('POST', `/api/bookings/${bookingId}/cancel`, { reason }, true);
    showToast(result.message || 'Booking cancelled', 'info');
    setTimeout(loadBookings, 800);
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// Feature 5: Customer raises damage dispute modal control
let _disputeBookingId = null;

function openDisputeModal(bookingId) {
  _disputeBookingId = bookingId;
  document.getElementById('disputeBookingId').value = bookingId;
  document.getElementById('disputeDescription').value = '';
  document.getElementById('disputeAmount').value = '0';
  document.getElementById('disputePhotoUrls').value = '';
  document.getElementById('disputeModal').classList.remove('hidden');
}

function closeDisputeModal() {
  document.getElementById('disputeModal').classList.add('hidden');
  _disputeBookingId = null;
}

async function submitDamageDisputeForm() {
  const description = document.getElementById('disputeDescription').value.trim();
  if (description.length < 10) {
    showToast('Please provide a detailed description (min 10 characters)', 'error');
    return;
  }
  const amountStr = document.getElementById('disputeAmount').value.trim();
  const claimedAmount = parseFloat(amountStr) || 0;
  const photoUrls = document.getElementById('disputePhotoUrls').value.trim();
  const photos = photoUrls.split(',').map(p => p.trim()).filter(Boolean);

  const btn = document.getElementById('submitDisputeBtn');
  btn.disabled = true; btn.textContent = 'Raising Dispute...';

  try {
    await api.post(`/bookings/${_disputeBookingId}/dispute`, {
      description,
      claimed_amount: claimedAmount,
      photo_urls: photos,
    }, true);
    showToast('Damage dispute raised. Our team will review within 48 hours.', 'success');
    closeDisputeModal();
    loadBookings();
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    btn.disabled = false; btn.textContent = 'Raise Dispute';
  }
}

async function submitBookingReview(bookingId) {
  const rating = ratingSelected[bookingId] || 0;
  const review = document.getElementById(`review-${bookingId}`)?.value.trim() || '';
  if (rating === 0) {
    showToast('Please select a rating star', 'error');
    return;
  }
  try {
    await api.post(`/bookings/${bookingId}/review`, { rating, review }, true);
    showToast('Review submitted successfully!', 'success');
    setTimeout(loadBookings, 800);
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// Feature 5: Fetch and display handover photos for customer
function loadAllHandovers(bookings) {
  bookings.forEach(b => {
    if (b.status === 'dispatched' || b.status === 'delivered' || b.status === 'completed') {
      fetchHandoverData(b.id);
    }
  });
}

async function fetchHandoverData(bookingId) {
  try {
    const data = await api.get(`/bookings/${bookingId}/handover`, true);
    const handovers = data.handovers || [];
    const container = document.getElementById(`handover-photos-${bookingId}`);
    if (!container) return;
    if (handovers.length === 0) {
      container.innerHTML = '';
      return;
    }
    
    let html = `<div style="margin-top:0.75rem;padding:0.75rem;background:#f8fafc;border:1px solid var(--border);border-radius:var(--radius);">
      <div style="font-weight:600;font-size:0.8rem;margin-bottom:0.5rem;color:var(--text);">📷 Handover Evidence & Checklist</div>`;
      
    handovers.forEach(h => {
      const photosHtml = (h.photo_urls || []).map(url => `
        <a href="${url}" target="_blank" style="margin-right:0.5rem;display:inline-block;">
          <img src="${url}" style="width:60px;height:60px;object-fit:cover;border-radius:4px;border:1px solid var(--border);" onerror="this.src='https://placehold.co/60x60?text=Photo'"/>
        </a>
      `).join('');
      
      let checklistHtml = '';
      if (h.checklist && Object.keys(h.checklist).length > 0) {
        checklistHtml = `<div style="font-size:0.75rem;margin-top:0.4rem;color:var(--text-light);">
          <strong>Checklist:</strong> ${Object.entries(h.checklist).map(([k, v]) => `${k}: ${v}`).join(', ')}
        </div>`;
      }
      
      html += `
        <div style="margin-top:0.5rem;font-size:0.78rem;border-top:1px dashed var(--border);padding-top:0.5rem;">
          <strong style="text-transform:capitalize;">${h.type} handover:</strong>
          <div style="margin-top:0.3rem;display:flex;flex-wrap:wrap;">${photosHtml}</div>
          ${checklistHtml}
          ${h.notes ? `<div style="font-size:0.75rem;margin-top:0.25rem;color:var(--text-light);"><em>Notes: ${h.notes}</em></div>` : ''}
        </div>
      `;
    });
    
    html += `</div>`;
    container.innerHTML = html;
  } catch (err) {
    console.error('Error fetching handover photos:', err);
  }
}
