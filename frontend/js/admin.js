// GenRent Admin Dashboard JavaScript

document.addEventListener('DOMContentLoaded', async () => {
  if (!isLoggedIn()) {
    window.location.href = '/login';
    return;
  }

  const user = getUser();
  if (!user || user.role !== 'admin') {
    showToast('Admin access required', 'error');
    setTimeout(() => window.location.href = '/', 1500);
    return;
  }

  loadStats();
});

function showSection(name) {
  document.querySelectorAll('[id^="section-"]').forEach(el => el.style.display = 'none');
  document.querySelectorAll('.sidebar-nav a').forEach(a => a.classList.remove('active'));

  document.getElementById(`section-${name}`).style.display = 'block';
  const navEl = document.getElementById(`nav-${name}`);
  if (navEl) navEl.classList.add('active');

  if (name === 'vendors') loadVendors();
  if (name === 'generators') loadAdminGenerators();
  if (name === 'bookings') loadAdminBookings();
  if (name === 'withdrawals') loadAdminWithdrawals();
  if (name === 'disputes') loadAdminDisputes();
}

// ---- Stats ----
async function loadStats() {
  try {
    const data = await api.adminStats();

    const statsGrid = document.getElementById('statsGrid');
    statsGrid.innerHTML = `
      <div class="stat-card orange">
        <div class="stat-card-label">Total Vendors</div>
        <div class="stat-card-value">${data.vendors.total}</div>
        <div class="stat-card-sub">${data.vendors.pending} pending verification</div>
      </div>
      <div class="stat-card blue">
        <div class="stat-card-label">Equipment Listed</div>
        <div class="stat-card-value">${data.generators.total}</div>
        <div class="stat-card-sub">${data.generators.available} available</div>
      </div>
      <div class="stat-card green">
        <div class="stat-card-label">Total Bookings</div>
        <div class="stat-card-value">${data.bookings.total}</div>
        <div class="stat-card-sub">${data.bookings.pending} pending</div>
      </div>
      <div class="stat-card">
        <div class="stat-card-label">Total Revenue</div>
        <div class="stat-card-value" style="font-size:1.5rem;">${formatCurrency(data.total_revenue)}</div>
        <div class="stat-card-sub">From completed bookings</div>
      </div>
    `;

    document.getElementById('vendorStats').innerHTML = `
      <div style="display:flex;flex-direction:column;gap:0.75rem;">
        <div style="display:flex;justify-content:space-between;">
          <span>Total Vendors</span><strong>${data.vendors.total}</strong>
        </div>
        <div style="display:flex;justify-content:space-between;">
          <span>Verified</span><strong style="color:var(--success)">${data.vendors.verified}</strong>
        </div>
        <div style="display:flex;justify-content:space-between;">
          <span>Pending</span><strong style="color:var(--warning)">${data.vendors.pending}</strong>
        </div>
        <div style="display:flex;justify-content:space-between;">
          <span>Customers</span><strong>${data.customers}</strong>
        </div>
      </div>
    `;

    document.getElementById('bookingStats').innerHTML = `
      <div style="display:flex;flex-direction:column;gap:0.75rem;">
        <div style="display:flex;justify-content:space-between;">
          <span>Total Bookings</span><strong>${data.bookings.total}</strong>
        </div>
        <div style="display:flex;justify-content:space-between;">
          <span>Pending</span><strong style="color:var(--warning)">${data.bookings.pending}</strong>
        </div>
        <div style="display:flex;justify-content:space-between;">
          <span>Confirmed</span><strong style="color:var(--info)">${data.bookings.confirmed}</strong>
        </div>
        <div style="display:flex;justify-content:space-between;">
          <span>Completed</span><strong style="color:var(--success)">${data.bookings.completed}</strong>
        </div>
      </div>
    `;
  } catch (err) {
    showToast('Failed to load stats: ' + err.message, 'error');
  }
}

// ---- Vendors ----
async function loadVendors() {
  const tbody = document.getElementById('vendorsTableBody');
  tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:2rem;color:var(--text-light);">Loading...</td></tr>';

  try {
    const verified = document.getElementById('vendorVerifiedFilter').value;
    const params = verified !== '' ? { verified } : {};
    const data = await api.adminVendors(params);
    const vendors = data.vendors || [];

    if (vendors.length === 0) {
      tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:2rem;color:var(--text-light);">No vendors found</td></tr>';
      return;
    }

    tbody.innerHTML = vendors.map(v => `
      <tr>
        <td>#${v.id}</td>
        <td>
          <div style="font-weight:600;">${v.company_name}</div>
          <div style="font-size:0.75rem;color:var(--text-light);">${v.address || ''}</div>
        </td>
        <td>
          <div>${v.user ? v.user.name : 'N/A'}</div>
          <div style="font-size:0.75rem;color:var(--text-light);">${v.user ? v.user.email : ''}</div>
        </td>
        <td>${v.city}</td>
        <td>${v.verified
          ? '<span class="status-badge status-available">Verified</span>'
          : '<span class="status-badge status-pending">Pending</span>'
        }</td>
        <td>
          ${!v.verified ? `
            <button class="btn btn-sm btn-success" onclick="verifyVendor(${v.id})">Verify</button>
          ` : `
            <button class="btn btn-sm btn-danger" onclick="rejectVendor(${v.id})">Revoke</button>
          `}
        </td>
      </tr>
    `).join('');
  } catch (err) {
    tbody.innerHTML = `<tr><td colspan="6" class="alert alert-error">${err.message}</td></tr>`;
  }
}

async function verifyVendor(id) {
  try {
    await api.adminVerifyVendor(id);
    showToast('Vendor verified successfully', 'success');
    loadVendors();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function rejectVendor(id) {
  if (!confirm('Revoke vendor verification?')) return;
  try {
    await api.adminRejectVendor(id);
    showToast('Vendor verification revoked', 'success');
    loadVendors();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// ---- Generators ----
async function loadAdminGenerators() {
  const tbody = document.getElementById('generatorsTableBody');
  tbody.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:2rem;color:var(--text-light);">Loading...</td></tr>';

  try {
    const data = await api.adminGenerators();
    const generators = data.generators || [];

    if (generators.length === 0) {
      tbody.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:2rem;color:var(--text-light);">No equipment found</td></tr>';
      return;
    }

    tbody.innerHTML = generators.map(g => `
      <tr>
        <td>#${g.id}</td>
        <td>
          <div style="font-weight:600;">${g.name}</div>
          ${g.brand ? `<div style="font-size:0.75rem;color:var(--text-light);">${g.brand}</div>` : ''}
        </td>
        <td>${g.vendor ? g.vendor.company_name : 'N/A'}</td>
        <td>${g.capacity_kva} kVA</td>
        <td>${formatCurrency(g.price_per_day)}</td>
        <td>${g.city}</td>
        <td>${statusBadge(g.availability_status)}</td>
        <td>
          <select class="btn btn-sm" style="border:1px solid var(--border);cursor:pointer;"
            onchange="updateGeneratorStatus(${g.id}, this.value)">
            <option value="available" ${g.availability_status === 'available' ? 'selected' : ''}>Available</option>
            <option value="maintenance" ${g.availability_status === 'maintenance' ? 'selected' : ''}>Maintenance</option>
            <option value="booked" ${g.availability_status === 'booked' ? 'selected' : ''}>Booked</option>
          </select>
        </td>
      </tr>
    `).join('');
  } catch (err) {
    tbody.innerHTML = `<tr><td colspan="8" class="alert alert-error">${err.message}</td></tr>`;
  }
}

async function updateGeneratorStatus(id, status) {
  try {
    await api.adminUpdateGeneratorStatus(id, status);
    showToast('Equipment status updated', 'success');
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// ---- Bookings ----
async function loadAdminBookings() {
  const tbody = document.getElementById('adminBookingsBody');
  tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;padding:2rem;color:var(--text-light);">Loading...</td></tr>';

  try {
    const status = document.getElementById('adminBookingFilter').value;
    const params = status ? { status } : {};
    const data = await api.adminBookings(params);
    const bookings = data.bookings || [];

    if (bookings.length === 0) {
      tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;padding:2rem;color:var(--text-light);">No bookings found</td></tr>';
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
        <td>${b.generator && b.generator.vendor ? b.generator.vendor.company_name : 'N/A'}</td>
        <td>
          <div>${formatDate(b.start_date)}</div>
          <div style="font-size:0.75rem;color:var(--text-light);">to ${formatDate(b.end_date)}</div>
        </td>
        <td>${formatCurrency(b.total_price)}</td>
        <td>${statusBadge(b.status)}</td>
      </tr>
    `).join('');
  } catch (err) {
    tbody.innerHTML = `<tr><td colspan="7" class="alert alert-error">${err.message}</td></tr>`;
  }
}

// ---- Withdrawals ----
async function loadAdminWithdrawals() {
  const tbody = document.getElementById('withdrawalsBody');
  tbody.innerHTML = '<tr><td colspan="9" style="text-align:center;padding:2rem;color:var(--text-light);">Loading...</td></tr>';

  try {
    const status = document.getElementById('withdrawalFilter').value || 'pending';
    const data = await api.get(`/admin/withdrawals?status=${status}`, true);
    const list = data.withdrawals || [];

    if (!list.length) {
      tbody.innerHTML = `<tr><td colspan="9" style="text-align:center;padding:2rem;color:var(--text-light);">No ${status} withdrawal requests</td></tr>`;
      return;
    }

    const statusColor = { pending: '#f59e0b', approved: '#16a34a', rejected: '#dc2626' };
    tbody.innerHTML = list.map(w => `
      <tr>
        <td>#${w.id}</td>
        <td>
          <div style="font-weight:600;">${w.vendor ? w.vendor.company_name : 'N/A'}</div>
          <div style="font-size:0.75rem;color:var(--text-light);">${w.vendor && w.vendor.user ? w.vendor.user.email : ''}</div>
        </td>
        <td style="font-weight:700;">₹${w.amount.toLocaleString('en-IN')}</td>
        <td>${w.bank_name}</td>
        <td>
          <div>${w.account_name}</div>
          <div style="font-size:0.75rem;color:#6b7280;">${w.account_no}</div>
        </td>
        <td>${w.ifsc}</td>
        <td>${formatDate(w.created_at)}</td>
        <td><span style="font-weight:600;color:${statusColor[w.status] || '#6b7280'};">${w.status.toUpperCase()}</span></td>
        <td>
          ${w.status === 'pending' ? `
            <button class="btn btn-sm btn-primary" onclick="approveWithdrawal(${w.id})" style="margin-right:0.25rem;">Approve</button>
            <button class="btn btn-sm" style="background:#fef2f2;color:#dc2626;border:1px solid #fecaca;" onclick="rejectWithdrawal(${w.id})">Reject</button>
          ` : `<span style="font-size:0.8rem;color:#6b7280;">${w.admin_note || '—'}</span>`}
        </td>
      </tr>
    `).join('');
  } catch (err) {
    tbody.innerHTML = `<tr><td colspan="9" class="alert alert-error">${err.message}</td></tr>`;
  }
}

async function approveWithdrawal(id) {
  if (!confirm(`Approve withdrawal #${id}? This confirms the bank transfer was made.`)) return;
  try {
    const result = await api.post(`/admin/withdrawals/${id}/approve`, {}, true);
    showToast(result.message, 'success');
    loadAdminWithdrawals();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function rejectWithdrawal(id) {
  const note = prompt('Reason for rejection (required):');
  if (!note) return;
  try {
    const result = await api.post(`/admin/withdrawals/${id}/reject`, { note }, true);
    showToast(result.message, 'success');
    loadAdminWithdrawals();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// ---- Feature 5: Admin Disputes ----
async function loadAdminDisputes() {
  const tbody = document.getElementById('disputesBody');
  if (!tbody) return;
  const status = document.getElementById('disputeStatusFilter')?.value || '';
  try {
    const data = await api.get(`/admin/disputes${status ? '?status=' + status : ''}`, true);
    const disputes = data.disputes || [];
    if (disputes.length === 0) {
      tbody.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:2rem;color:var(--text-light);">No disputes found</td></tr>';
      return;
    }
    tbody.innerHTML = disputes.map(d => {
      const statusColor = { open: '#dc2626', resolved: '#16a34a', rejected: '#6b7280' }[d.status] || '#6b7280';
      return `<tr>
        <td>#${d.id}</td>
        <td><a href="#" onclick="loadAdminBookings()">Booking #${d.booking_id}</a></td>
        <td>${d.user?.name || d.raised_by}</td>
        <td title="${d.description}">${d.description.substring(0, 60)}${d.description.length > 60 ? '...' : ''}</td>
        <td>₹${d.claimed_amount.toLocaleString('en-IN')}</td>
        <td>${new Date(d.created_at).toLocaleDateString('en-IN')}</td>
        <td><span style="color:${statusColor};font-weight:600;">${d.status.toUpperCase()}</span></td>
        <td>
          ${d.status === 'open' ? `
            <button class="btn btn-sm btn-success" onclick="resolveDispute(${d.id},'resolved')">✅ Resolve</button>
            <button class="btn btn-sm btn-danger" onclick="resolveDispute(${d.id},'rejected')">❌ Reject</button>
          ` : `<span style="font-size:0.75rem;color:var(--text-light);">${d.admin_notes || '—'}</span>`}
        </td>
      </tr>`;
    }).join('');
  } catch (err) {
    tbody.innerHTML = `<tr><td colspan="8" class="alert alert-error">${err.message}</td></tr>`;
  }
}

async function resolveDispute(disputeId, status) {
  const adminNotes = prompt(`Enter admin notes for ${status === 'resolved' ? 'resolution' : 'rejection'} (required):`);
  if (!adminNotes || adminNotes.trim().length < 5) {
    showToast('Admin notes are required (min 5 chars)', 'error');
    return;
  }
  try {
    const result = await api.put(`/admin/disputes/${disputeId}/resolve`, { status, admin_notes: adminNotes }, true);
    showToast(result.message || `Dispute ${status}`, 'success');
    loadAdminDisputes();
  } catch (err) {
    showToast(err.message, 'error');
  }
}
