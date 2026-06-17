// Equipment API functions
const EQUIPMENT_API_BASE = '/api';

let currentEquipmentPage = 1;

// Load popular categories on page load
async function loadPopularCategories() {
  try {
    const response = await fetch(`${EQUIPMENT_API_BASE}/categories/popular?limit=6`);
    const data = await response.json();
    
    const container = document.getElementById('popularCategories');
    if (!data.categories || data.categories.length === 0) {
      container.innerHTML = '<div style="grid-column:1/-1;text-align:center;padding:2rem;color:#6b7280;">No categories available</div>';
      return;
    }

    const icons = {
      'Generators': '⚡',
      'Tower Lights': '💡',
      'Forklifts': '🚜',
      'Excavators': '🏗️',
      'Compressors': '🔧',
      'Cranes': '🏭',
      'Concrete Mixers': '🏭',
      'Water Pumps': '💧',
      'Welding Machines': '🔩',
      'Cutting Machines': '✂️',
    };

    container.innerHTML = data.categories.map(cat => `
      <a href="#generators" class="category-card" onclick="filterByCategory(${cat.id})">
        <div class="category-icon">${icons[cat.name] || '🔧'}</div>
        <div class="category-label">${cat.name}</div>
        <div class="category-sub">${cat.equipment_count} available</div>
      </a>
    `).join('');
  } catch (error) {
    console.error('Error loading categories:', error);
  }
}

// Load categories dropdown
async function loadCategories() {
  try {
    const response = await fetch(`${EQUIPMENT_API_BASE}/categories/hierarchy`);
    const data = await response.json();
    
    const select = document.getElementById('searchCategory');
    let options = '<option value="">All Categories</option>';

    if (data.categories) {
      data.categories.forEach(cat => {
        if (cat.subcategories && cat.subcategories.length > 0) {
          cat.subcategories.forEach(sub => {
            options += `<option value="${sub.id}">${sub.name}</option>`;
          });
        } else {
          options += `<option value="${cat.id}">${cat.name}</option>`;
        }
      });
    }

    select.innerHTML = options;
  } catch (error) {
    console.error('Error loading categories:', error);
  }
}

// PRIMARY_SPECS maps categories to their primary numerical spec for search filtering
const PRIMARY_SPECS = {
  'Generators': { key: 'capacity_kva', label: 'Capacity (kVA)', min: 10, max: 1000, step: 10 },
  'Forklifts': { key: 'lifting_capacity_tons', label: 'Lifting Capacity (Tons)', min: 1, max: 20, step: 0.5 },
  'Excavators': { key: 'operating_weight_tons', label: 'Operating Weight (Tons)', min: 1, max: 100, step: 1 },
  'Backhoe Loaders': { key: 'bucket_capacity_m3', label: 'Bucket Capacity (m³)', min: 0.1, max: 2.0, step: 0.05 },
  'Hydra Cranes': { key: 'lifting_capacity_tons', label: 'Lifting Capacity (Tons)', min: 5, max: 50, step: 1 },
  'Boom Lifts': { key: 'working_height_m', label: 'Working Height (m)', min: 5, max: 50, step: 1 },
  'Scissor Lifts': { key: 'working_height_m', label: 'Working Height (m)', min: 5, max: 30, step: 1 },
  'Air Compressors': { key: 'cfm', label: 'Capacity (CFM)', min: 50, max: 1500, step: 10 },
  'Water Pumps': { key: 'flow_rate_lpm', label: 'Flow Rate (LPM)', min: 100, max: 5000, step: 50 },
  'Welding Machines': { key: 'amperage', label: 'Max Amperage (A)', min: 100, max: 600, step: 50 },
  'Tower Lights': { key: 'wattage', label: 'Wattage per Lamp (W)', min: 100, max: 2000, step: 50 },
  'Concrete Mixers': { key: 'drum_capacity_litres', label: 'Drum Capacity (Litres)', min: 100, max: 1000, step: 50 }
};

// Handle category selection change in search
function onSearchCategoryChange() {
  const select = document.getElementById('searchCategory');
  const catName = select.options[select.selectedIndex]?.text;
  const container = document.getElementById('dynamicSpecFilters');
  const grid = document.getElementById('specFiltersGrid');

  if (!catName || !PRIMARY_SPECS[catName]) {
    container.style.display = 'none';
    grid.innerHTML = '';
    searchEquipment(1);
    return;
  }

  const spec = PRIMARY_SPECS[catName];
  container.style.display = 'block';
  grid.innerHTML = `
    <div class="search-field" style="width: 100%;">
      <label style="color:#d1d5db; display:flex; justify-content:space-between; width: 100%;">
        <span>Minimum ${spec.label}: <strong id="specMinVal">${spec.min}</strong></span>
      </label>
      <input type="range" id="searchSpecMin" min="${spec.min}" max="${spec.max}" step="${spec.step}" value="${spec.min}" 
        style="width: 100%; margin-top: 0.5rem;"
        oninput="document.getElementById('specMinVal').innerText = this.value" onchange="searchEquipment(1)">
    </div>
  `;
  searchEquipment(1);
}

// Search equipment with table view
async function searchEquipment(page = 1) {
  currentEquipmentPage = page;
  const city = document.getElementById('searchCity').value;
  const category = document.getElementById('searchCategory').value;

  let query = `${EQUIPMENT_API_BASE}/equipment/search?page=${currentEquipmentPage}&limit=10&`;
  if (city) query += `city=${city}&`;
  if (category) query += `category=${category}&`;

  const specMinInput = document.getElementById('searchSpecMin');
  if (specMinInput) {
    const select = document.getElementById('searchCategory');
    const catName = select.options[select.selectedIndex]?.text;
    const spec = PRIMARY_SPECS[catName];
    if (spec) {
      query += `spec_key=${spec.key}&spec_min=${specMinInput.value}&`;
    }
  }

  const loading = document.getElementById('generatorsLoading');
  const table = document.getElementById('generatorsGrid');
  const empty = document.getElementById('generatorsEmpty');
  const pagination = document.getElementById('pagination');

  loading.style.display = 'block';
  table.style.display = 'none';
  empty.style.display = 'none';
  pagination.style.display = 'none';

  try {
    const response = await fetch(query);
    const data = await response.json();

    if (!data.equipment || data.equipment.length === 0) {
      loading.style.display = 'none';
      empty.style.display = 'block';
      return;
    }

    renderEquipmentTable(data.equipment);
    loading.style.display = 'none';
    table.style.display = 'block';

    if (data.total > data.limit) {
      renderEquipmentPagination(data.page, Math.ceil(data.total / data.limit));
      pagination.style.display = 'flex';
    }
  } catch (error) {
    console.error('Error searching equipment:', error);
    loading.style.display = 'none';
    empty.style.display = 'block';
  }
}

// Render equipment table
function renderEquipmentTable(equipment) {
  const tbody = document.getElementById('equipmentBody');
  tbody.innerHTML = equipment.map(eq => `
    <tr>
      <td>${eq.name}</td>
      <td>${eq.category || 'N/A'}</td>
      <td>${eq.city}</td>
      <td>₹${eq.monthly_price ? eq.monthly_price.toLocaleString() : eq.daily_price.toLocaleString()}</td>
      <td>${eq.vendor}</td>
      <td><a href="/booking?id=${eq.id}" class="btn btn-sm btn-primary">View</a></td>
    </tr>
  `).join('');
}

// Filter by category
function filterByCategory(categoryId) {
  document.getElementById('searchCategory').value = categoryId;
  document.getElementById('searchCity').value = 'Mumbai';
  onSearchCategoryChange();
}

// Helper function to render pagination
function renderEquipmentPagination(currentPage, totalPages) {
  const pagination = document.getElementById('pagination');
  let html = '';

  if (currentPage > 1) {
    html += `<button class="pagination-btn" onclick="goToEquipmentPage(${currentPage - 1})">← Previous</button>`;
  }

  for (let i = Math.max(1, currentPage - 2); i <= Math.min(totalPages, currentPage + 2); i++) {
    if (i === currentPage) {
      html += `<button class="pagination-btn active">${i}</button>`;
    } else {
      html += `<button class="pagination-btn" onclick="goToEquipmentPage(${i})">${i}</button>`;
    }
  }

  if (currentPage < totalPages) {
    html += `<button class="pagination-btn" onclick="goToEquipmentPage(${currentPage + 1})">Next →</button>`;
  }

  pagination.innerHTML = html;
}

// Go to specific page
function goToEquipmentPage(page) {
  searchEquipment(page);
  document.getElementById('generators').scrollIntoView({ behavior: 'smooth' });
}

// Update navbar actions based on auth state
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

// Load equipment on page init
document.addEventListener('DOMContentLoaded', () => {
  updateNavbar();
  loadCategories();
  loadPopularCategories();
  searchEquipment(1);
});
