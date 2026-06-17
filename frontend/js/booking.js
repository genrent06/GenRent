// GenRent Booking Page JavaScript

let generatorData = null;

document.addEventListener('DOMContentLoaded', () => {
  const id = new URLSearchParams(window.location.search).get('id');
  if (!id) {
    window.location.href = '/';
    return;
  }
  loadGeneratorDetail(id);
  updateNavbar();

  // Set minimum date to today
  const today = new Date().toISOString().split('T')[0];
  document.getElementById('startDate').min = today;
  document.getElementById('endDate').min = today;

  document.getElementById('startDate').addEventListener('change', calculatePrice);
  document.getElementById('endDate').addEventListener('change', calculatePrice);
});

function updateNavbar() {
  const navActions = document.getElementById('navActions');
  if (!navActions) return;

  const user = getUser();
  if (user) {
    let dashLink = '/';
    if (user.role === 'vendor') dashLink = '/vendor-dashboard';
    else if (user.role === 'admin') dashLink = '/admin-dashboard';

    navActions.innerHTML = `
      <a href="${dashLink}" class="btn btn-outline btn-sm">Dashboard</a>
      <button class="btn btn-danger btn-sm" onclick="logout()">Logout</button>
    `;
  }
}

async function loadGeneratorDetail(id) {
  try {
    const gen = await api.getGenerator(id);
    generatorData = gen;

    document.getElementById('loadingState').style.display = 'none';
    document.getElementById('detailContent').style.display = 'block';
    document.title = `${gen.name} - GenRent`;

    // Image / Icon
    if (gen.image_url) {
      document.getElementById('genImage').innerHTML = `<img src="${gen.image_url}" alt="${gen.name}" style="width:100%;height:100%;object-fit:cover;border-radius:var(--radius-lg);" />`;
    }

    document.getElementById('genName').textContent = gen.name;

    const vendor = gen.vendor || {};
    document.getElementById('genVendor').innerHTML = `
      🏢 <strong>${vendor.company_name || 'Unknown Vendor'}</strong>
      ${vendor.verified ? `<span class="verified-badge" title="GenRent has verified this business">✔ Verified</span>` : ''}
      &nbsp;·&nbsp; 📍 ${gen.location}, ${gen.city}
    `;

    // Meta tags — use specs for category-specific info
    const specs = gen.specs || {};
    const specTags = Object.entries(specs)
      .map(([k, v]) => `<span class="meta-tag">${formatSpecLabel(k, v)}</span>`)
      .join('');
    document.getElementById('genMeta').innerHTML = `
      ${specTags || `<span class="meta-tag">⚡ ${gen.model || gen.capacity_kva ? (gen.capacity_kva ? gen.capacity_kva + ' kVA' : gen.model) : 'Industrial Equipment'}</span>`}
      ${gen.brand ? `<span class="meta-tag">🏷️ ${gen.brand}</span>` : ''}
      <span class="status-badge status-${gen.availability_status} meta-tag">${gen.availability_status}</span>
      ${gen.available_quantity != null ? `<span class="meta-tag" style="background:#eff6ff;color:#1d4ed8;">📦 ${gen.available_quantity} unit${gen.available_quantity !== 1 ? 's' : ''} available</span>` : ''}
    `;

    if (gen.description) {
      document.getElementById('genDescription').textContent = gen.description;
    }

    // Technical Specs card
    if (Object.keys(specs).length > 0) {
      const specsHtml = Object.entries(specs).map(([k, v]) =>
        `<div style="display:flex;justify-content:space-between;padding:0.4rem 0;border-bottom:1px solid var(--border);font-size:0.875rem;">
          <span style="color:var(--text-light);">${k.replace(/_/g,' ').replace(/\b\w/g,c=>c.toUpperCase())}</span>
          <span style="font-weight:600;">${v}</span>
        </div>`
      ).join('');
      const specsCard = `
        <div style="background:white;border:1px solid var(--border);border-radius:var(--radius-lg);padding:1.25rem;margin-top:1rem;">
          <h4 style="font-size:0.85rem;font-weight:700;color:var(--text-light);text-transform:uppercase;letter-spacing:0.05em;margin-bottom:0.75rem;">⚙️ Technical Specifications</h4>
          ${specsHtml}
        </div>`;
      const descEl = document.getElementById('genDescription');
      descEl.insertAdjacentHTML('afterend', specsCard);
    }

    // Transport Fees
    if ((gen.mobilization_fee || 0) + (gen.demobilization_fee || 0) > 0) {
      const feeHtml = `
        <div style="background:#fefce8;border:1px solid #fde047;border-radius:var(--radius);padding:0.75rem;margin-top:0.75rem;font-size:0.85rem;">
          <div style="font-weight:600;margin-bottom:0.25rem;">🚚 Transport Charges</div>
          ${gen.mobilization_fee ? `<div>Mobilization (delivery): <strong>${formatCurrency(gen.mobilization_fee)}</strong></div>` : ''}
          ${gen.demobilization_fee ? `<div>Demobilization (pickup): <strong>${formatCurrency(gen.demobilization_fee)}</strong></div>` : ''}
        </div>`;
      document.getElementById('vendorInfo').insertAdjacentHTML('afterend', feeHtml);
    }

    // Vendor info
    document.getElementById('vendorInfo').innerHTML = `
      <div style="display:flex;flex-direction:column;gap:0.5rem;">
        <div style="font-weight:600;display:flex;align-items:center;gap:0.5rem;">
          ${vendor.company_name || ''}
          ${vendor.verified ? `<span class="verified-badge" title="GenRent has verified this business">✔ Verified</span>` : ''}
        </div>
        <div style="font-size:0.875rem;color:var(--text-light);">📍 ${vendor.city || ''}</div>
        ${vendor.description ? `<div style="font-size:0.875rem;color:var(--text-light);margin-top:0.5rem;">${vendor.description}</div>` : ''}
      </div>
    `;

    // Pricing
    let pricingHtml = `${formatCurrency(gen.daily_price || gen.price_per_day)} <span>/ day</span>`;
    if (gen.weekly_price) {
      pricingHtml += `<div style="font-size:0.9rem;color:var(--text-light);margin-top:0.25rem;">${formatCurrency(gen.weekly_price)} / week</div>`;
    }
    if (gen.monthly_price || gen.price_per_month) {
      pricingHtml += `<div style="font-size:0.9rem;color:var(--text-light);margin-top:0.25rem;">${formatCurrency(gen.monthly_price || gen.price_per_month)} / month</div>`;
    }
    document.getElementById('priceDisplay').innerHTML = pricingHtml;

    // Show booking form or login prompt
    if (isLoggedIn()) {
      document.getElementById('bookingForm').style.display = 'block';
      document.getElementById('loginPrompt').style.display = 'none';
    } else {
      document.getElementById('bookingForm').style.display = 'none';
      document.getElementById('loginPrompt').style.display = 'block';
    }

    // Disable booking if not available
    if ((gen.available_quantity != null && gen.available_quantity <= 0) || gen.availability_status === 'maintenance') {
      document.getElementById('bookBtn').disabled = true;
      document.getElementById('bookBtn').textContent = 'Not Available';
    }
  } catch (err) {
    document.getElementById('loadingState').innerHTML = `
      <div class="alert alert-error">Failed to load equipment: ${err.message}</div>
      <p style="text-align:center;margin-top:1rem;"><a href="/">← Back to listings</a></p>
    `;
  }
}

function formatSpecLabel(key, value) {
  const labels = {
    capacity_kva: `⚡ ${value} kVA`,
    fuel_type: `⛽ ${value}`,
    phase: `🔌 ${value}`,
    lifting_capacity_tons: `⬆️ ${value}T capacity`,
    operating_weight_tons: `⚖️ ${value}T`,
    working_height_m: `📏 ${value}m height`,
    cfm: `💨 ${value} CFM`,
    pressure_psi: `🔧 ${value} PSI`,
  };
  return labels[key] || `${key.replace(/_/g,' ')}: ${value}`;
}

function calculateTieredPriceDetails(days, daily, weekly, monthly) {
  daily = Number(daily) || 0;
  weekly = Number(weekly) || 0;
  monthly = Number(monthly) || 0;

  if (days <= 0) {
    return { total: 0, breakdown: '', discount: 0 };
  }

  // Fallback if no weekly or monthly rates
  if (weekly <= 0 && monthly <= 0) {
    const total = days * daily;
    return {
      total,
      breakdown: `${days} day${days > 1 ? 's' : ''} × ${formatCurrency(daily)}`,
      discount: 0
    };
  }

  let total = 0;
  let breakdownParts = [];
  const basePriceWithoutDiscounts = days * daily;

  if (monthly > 0) {
    const months = Math.floor(days / 30);
    const remainingDays = days % 30;

    if (months > 0) {
      total += months * monthly;
      breakdownParts.push(`${months} month${months > 1 ? 's' : ''} × ${formatCurrency(monthly)}`);
    }

    let remainingCost = 0;
    let remBreakdown = '';
    if (weekly > 0) {
      const weeks = Math.floor(remainingDays / 7);
      const leftoverDays = remainingDays % 7;

      const weekCost = weeks * weekly;
      let dayCost = leftoverDays * daily;
      let leftoverDaysBreakdown = '';

      if (dayCost > weekly) {
        dayCost = weekly;
        leftoverDaysBreakdown = `leftover capped at weekly rate: ${formatCurrency(weekly)}`;
      } else if (leftoverDays > 0) {
        leftoverDaysBreakdown = `${leftoverDays} day${leftoverDays > 1 ? 's' : ''} × ${formatCurrency(daily)}`;
      }

      remainingCost = weekCost + dayCost;
      
      let parts = [];
      if (weeks > 0) {
        parts.push(`${weeks} week${weeks > 1 ? 's' : ''} × ${formatCurrency(weekly)}`);
      }
      if (leftoverDaysBreakdown) {
        parts.push(leftoverDaysBreakdown);
      }
      remBreakdown = parts.join(' + ');
    } else {
      remainingCost = remainingDays * daily;
      if (remainingDays > 0) {
        remBreakdown = `${remainingDays} day${remainingDays > 1 ? 's' : ''} × ${formatCurrency(daily)}`;
      }
    }

    // Check monthly cap
    if (remainingCost > monthly) {
      remainingCost = monthly;
      breakdownParts.push(`remainder capped at monthly rate: ${formatCurrency(monthly)}`);
    } else if (remainingCost > 0) {
      breakdownParts.push(remBreakdown);
    }

    total += remainingCost;
  } else {
    // Only weekly and daily are available
    const weeks = Math.floor(days / 7);
    const leftoverDays = days % 7;

    if (weeks > 0) {
      total += weeks * weekly;
      breakdownParts.push(`${weeks} week${weeks > 1 ? 's' : ''} × ${formatCurrency(weekly)}`);
    }

    let dayCost = leftoverDays * daily;
    if (dayCost > weekly) {
      dayCost = weekly;
      breakdownParts.push(`leftover capped at weekly rate: ${formatCurrency(weekly)}`);
    } else if (leftoverDays > 0) {
      breakdownParts.push(`${leftoverDays} day${leftoverDays > 1 ? 's' : ''} × ${formatCurrency(daily)}`);
    }
    total += dayCost;
  }

  const discount = Math.max(0, basePriceWithoutDiscounts - total);

  return {
    total,
    breakdown: breakdownParts.join(' + '),
    discount
  };
}

function calculatePrice() {
  if (!generatorData) return;

  const startDate = document.getElementById('startDate').value;
  const endDate = document.getElementById('endDate').value;

  if (!startDate || !endDate) {
    document.getElementById('priceCalc').style.display = 'none';
    return;
  }

  const start = new Date(startDate);
  const end = new Date(endDate);

  if (end <= start) {
    document.getElementById('priceCalc').style.display = 'none';
    return;
  }

  const days = Math.ceil((end - start) / (1000 * 60 * 60 * 24));
  
  const daily = generatorData.daily_price || generatorData.price_per_day || 0;
  const weekly = generatorData.weekly_price || 0;
  const monthly = generatorData.monthly_price || generatorData.price_per_month || 0;
  const mobFee = generatorData.mobilization_fee || 0;
  const demobFee = generatorData.demobilization_fee || 0;

  const { total: rentalTotal, breakdown, discount } = calculateTieredPriceDetails(days, daily, weekly, monthly);
  const transportTotal = mobFee + demobFee;
  const total = rentalTotal + transportTotal;

  const advance = Math.round(total * 0.30 * 100) / 100;
  const vendorShare = Math.round(total * 0.15 * 100) / 100;
  const platformFee = Math.round(total * 0.15 * 100) / 100;
  const remaining = Math.round((total - advance) * 100) / 100;

  document.getElementById('calcDays').textContent = breakdown || `${days} day${days > 1 ? 's' : ''} × ${formatCurrency(daily)}`;
  document.getElementById('calcAmount').textContent = formatCurrency(rentalTotal);
  document.getElementById('calcTotal').textContent = formatCurrency(total);
  document.getElementById('calcAdvance').textContent = formatCurrency(advance);
  document.getElementById('calcVendorShare').textContent = formatCurrency(vendorShare);
  document.getElementById('calcPlatformFee').textContent = formatCurrency(platformFee);
  document.getElementById('calcRemaining').textContent = formatCurrency(remaining);

  // Transport fees line
  let transportEl = document.getElementById('transportFeeRow');
  if (transportTotal > 0) {
    if (!transportEl) {
      transportEl = document.createElement('div');
      transportEl.id = 'transportFeeRow';
      transportEl.style.cssText = 'display:flex;justify-content:space-between;padding:0.5rem 0;border-top:1px dashed var(--border);margin-top:0.25rem;font-size:0.875rem;';
      document.getElementById('priceCalc').appendChild(transportEl);
    }
    transportEl.innerHTML = `<span style="color:var(--text-light);">🚚 Transport (mob+demob)</span><span style="font-weight:600;">${formatCurrency(transportTotal)}</span>`;
    transportEl.style.display = 'flex';
  } else if (transportEl) {
    transportEl.style.display = 'none';
  }

  // Manage discount badge
  let discountBadge = document.getElementById('autoDiscountBadge');
  if (discount > 0) {
    if (!discountBadge) {
      discountBadge = document.createElement('div');
      discountBadge.id = 'autoDiscountBadge';
      discountBadge.style.cssText = `
        background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
        border: 1px solid #bbf7d0;
        border-radius: var(--radius);
        padding: 0.75rem;
        margin-bottom: 0.75rem;
        margin-top: 0.5rem;
        display: flex;
        justify-content: space-between;
        align-items: center;
        box-shadow: 0 2px 4px rgba(22, 163, 74, 0.05);
      `;
      const priceCalc = document.getElementById('priceCalc');
      priceCalc.insertBefore(discountBadge, priceCalc.children[1]); // insert after the breakdown
    }
    discountBadge.innerHTML = `
      <div style="font-size: 0.85rem; color: #16a34a; font-weight: 600; display: flex; align-items: center; gap: 0.25rem;">
        ✨ Auto-Discount Applied!
      </div>
      <div style="font-size: 0.85rem; color: #15803d; font-weight: 700;">
        Saved ${formatCurrency(discount)}
      </div>
    `;
    discountBadge.style.display = 'flex';
  } else if (discountBadge) {
    discountBadge.style.display = 'none';
  }

  document.getElementById('priceCalc').style.display = 'block';
}

async function submitBooking() {
  clearAlert('bookingAlert');

  if (!isLoggedIn()) {
    window.location.href = `/login?redirect=${encodeURIComponent(window.location.href)}`;
    return;
  }

  const address = document.getElementById('bookAddress').value;
  const startDate = document.getElementById('startDate').value;
  const endDate = document.getElementById('endDate').value;
  const notes = document.getElementById('bookNotes').value;

  if (!address || !startDate || !endDate) {
    showAlert('bookingAlert', 'Please fill in all required fields', 'error');
    return;
  }

  const btn = document.getElementById('bookBtn');
  btn.disabled = true;
  btn.innerHTML = '<span class="loading-spinner"></span> Submitting...';

  try {
    await api.createBooking({
      equipment_id: generatorData.id,
      start_date: startDate,
      end_date: endDate,
      address,
      notes,
    });

    showToast('Booking request sent! Waiting for vendor to accept...', 'success');
    btn.textContent = 'Redirecting...';

    // Redirect to my-bookings to track status
    setTimeout(() => {
      window.location.href = `/my-bookings`;
    }, 1000);
  } catch (err) {
    showAlert('bookingAlert', err.message, 'error');
    btn.disabled = false;
    btn.textContent = 'Request Booking →';
  }
}
