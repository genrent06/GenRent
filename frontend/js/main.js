// GenRent Main Page JavaScript

let currentPage = 1;
let currentFilters = {};

// ---- On Page Load ----
document.addEventListener('DOMContentLoaded', () => {
  updateNavbar();
  loadGenerators();
});

function updateNavbar() {
  const navActions = document.getElementById('navActions');
  if (!navActions) return;

  const user = getUser();
  if (user) {
    let extraLink = `<a href="/my-bookings" class="btn btn-outline btn-sm">My Bookings</a>`;
    if (user.role === 'vendor') {
      extraLink = `<a href="/vendor-dashboard" class="btn btn-outline btn-sm">Dashboard</a>`;
    } else if (user.role === 'admin') {
      extraLink = `<a href="/admin-dashboard" class="btn btn-outline btn-sm">Dashboard</a>`;
    }

    navActions.innerHTML = `
      ${extraLink}
      <button class="btn btn-danger btn-sm" onclick="logout()">Logout</button>
    `;
  }
}

// ---- Search ----
function searchGenerators() {
  const city = document.getElementById('searchCity').value;
  const capacity = document.getElementById('searchCapacity').value;

  currentFilters = {};
  if (city) currentFilters.city = city;
  if (capacity) currentFilters.capacity = capacity;
  currentPage = 1;

  document.getElementById('generators').scrollIntoView({ behavior: 'smooth' });
  loadGenerators();
}

function filterByCapacity(capacity) {
  document.getElementById('searchCapacity').value = capacity;
  currentFilters.capacity = capacity;
  currentPage = 1;
  loadGenerators();
}

async function loadGenerators() {
  const loading = document.getElementById('generatorsLoading');
  const grid = document.getElementById('generatorsGrid');
  const empty = document.getElementById('generatorsEmpty');
  const pagination = document.getElementById('pagination');

  loading.style.display = 'block';
  grid.style.display = 'none';
  empty.style.display = 'none';
  pagination.style.display = 'none';

  try {
    const params = { ...currentFilters, page: currentPage, limit: 9 };
    const data = await api.searchGenerators(params);

    loading.style.display = 'none';

    if (!data.generators || data.generators.length === 0) {
      empty.style.display = 'block';
      return;
    }

    grid.innerHTML = data.generators.map(renderGeneratorCard).join('');
    grid.style.display = 'grid';

    // Pagination
    const totalPages = Math.ceil(data.total / 9);
    if (totalPages > 1) {
      renderPagination(data.page, totalPages);
      pagination.style.display = 'flex';
    }
  } catch (err) {
    loading.innerHTML = `<div class="alert alert-error">Failed to load generators: ${err.message}</div>`;
  }
}

function renderGeneratorCard(gen) {
  const vendor = gen.vendor || {};

  return `
    <div class="generator-card" onclick="window.location='/booking?id=${gen.id}'">
      <div class="card-image">
        ${gen.image_url
          ? `<img src="${gen.image_url}" alt="${gen.name}" />`
          : `<span class="gen-icon">⚡</span>`
        }
        <span class="capacity-badge">${gen.capacity_kva} kVA</span>
      </div>
      <div class="card-body">
        <div class="card-title">${gen.name}</div>
        <div class="card-vendor">
          🏢 ${vendor.company_name || 'Unknown Vendor'}
          ${vendor.verified ? `<span class="verified-badge" title="GenRent has verified this business">✔ Verified</span>` : ''}
        </div>
        <div class="card-meta">
          <span class="meta-tag">📍 ${gen.city}</span>
          <span class="meta-tag">⛽ ${gen.fuel_type || 'Diesel'}</span>
          ${gen.brand ? `<span class="meta-tag">🏷️ ${gen.brand}</span>` : ''}
        </div>
        <div class="card-footer">
          <div class="card-price">
            ${formatCurrency(gen.price_per_day)} <span>/ day</span>
          </div>
          <span class="status-badge status-${gen.availability_status}">${gen.availability_status}</span>
        </div>
      </div>
    </div>
  `;
}

function renderPagination(page, totalPages) {
  const pagination = document.getElementById('pagination');
  let html = '';

  if (page > 1) {
    html += `<button class="page-btn" onclick="goToPage(${page - 1})">‹</button>`;
  }

  for (let i = 1; i <= totalPages; i++) {
    if (i === page || Math.abs(i - page) <= 2 || i === 1 || i === totalPages) {
      html += `<button class="page-btn ${i === page ? 'active' : ''}" onclick="goToPage(${i})">${i}</button>`;
    } else if (Math.abs(i - page) === 3) {
      html += `<span style="padding:0 0.5rem;color:var(--text-light);">...</span>`;
    }
  }

  if (page < totalPages) {
    html += `<button class="page-btn" onclick="goToPage(${page + 1})">›</button>`;
  }

  pagination.innerHTML = html;
}

function goToPage(page) {
  currentPage = page;
  loadGenerators();
  document.getElementById('generators').scrollIntoView({ behavior: 'smooth' });
}
