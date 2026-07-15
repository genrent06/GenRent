// Vendor Dashboard - Equipment Operations

const EQUIP_API_BASE = '/api/v1';

// Load equipment with operational stats
async function loadEquipmentStats() {
  try {
    const token = getToken();
    const response = await fetch(`${EQUIP_API_BASE}/equipment-stats`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) throw new Error('Failed to load equipment stats');

    const data = await response.json();
    if (!data.stats || data.stats.length === 0) {
      document.getElementById('equipmentStats').style.display = 'none';
      // Show "No equipment" message instead of leaving loading state
      document.getElementById('equipmentTableBody').innerHTML = `
        <tr><td colspan="6" style="text-align:center;padding:2rem;color:var(--text-light);">
          No equipment found. <a href="#" onclick="openAddEquipment()" style="color:var(--primary);font-weight:600;">Add your first equipment →</a>
        </td></tr>
      `;
      return;
    }

    document.getElementById('equipmentStats').style.display = 'grid';

    // Calculate totals
    const total = data.stats.length;
    const active = data.stats.filter(eq => eq.status === 'active').length;
    const totalRevenue = data.stats.reduce((sum, eq) => sum + (eq.total_revenue || 0), 0);

    document.getElementById('statTotalEquip').textContent = total;
    document.getElementById('statActiveEquip').textContent = active;
    document.getElementById('statMonthRevenue').textContent = `₹${totalRevenue.toLocaleString()}`;

    renderEquipmentOperationsTable(data.stats);
  } catch (error) {
    console.error('Error loading equipment stats:', error);
    document.getElementById('equipmentTableBody').innerHTML = `
      <tr><td colspan="6" style="text-align:center;padding:2rem;color:red;">Error loading equipment</td></tr>
    `;
  }
}

// Render equipment operations table
function renderEquipmentOperationsTable(stats) {
  const tbody = document.getElementById('equipmentTableBody');
  
  tbody.innerHTML = stats.map(eq => `
    <tr>
      <td style="font-weight: 500;color:var(--primary);">${eq.name}</td>
      <td>${getCategoryName(eq.category_id)}</td>
      <td>
        <span class="status-badge status-${eq.availability_status || 'available'}">
          ${eq.availability_status ? eq.availability_status.charAt(0).toUpperCase() + eq.availability_status.slice(1) : 'Available'}
        </span>
      </td>
      <td>${eq.total_bookings || 0}</td>
      <td style="font-weight: 600;color:#059669;">₹${(eq.total_revenue || 0).toLocaleString()}</td>
      <td>
        <div class="quick-actions">
          <button class="btn-icon" title="Edit" onclick="editEquipment(${eq.id})">✏️</button>
          <button class="btn-icon" title="Status" onclick="changeStatus(${eq.id})">⚙️</button>
          <button class="btn-icon" title="Delete" onclick="deleteEquipment(${eq.id})" style="color:#ef4444;">🗑️</button>
        </div>
      </td>
    </tr>
  `).join('');
}

// Helper to get category name
function getCategoryName(categoryId) {
  // Convert to integer in case it comes as float from API
  const id = parseInt(categoryId) || 0;
  const categories = {
    1: 'Power Equipment', 2: 'Generators', 3: 'Tower Lights', 4: 'Distribution Panels', 5: 'Cables',
    6: 'Construction Equipment', 7: 'Excavators', 8: 'Backhoe Loaders', 9: 'Concrete Mixers',
    10: 'Compactors', 11: 'Road Rollers', 12: 'Material Handling', 13: 'Forklifts', 14: 'Hydra Cranes',
    15: 'Boom Lifts', 16: 'Scissor Lifts', 17: 'Site Equipment', 18: 'Air Compressors', 19: 'Water Pumps',
    20: 'Welding Machines', 21: 'Cutting Machines'
  };
  return categories[id] || 'Unknown';
}

// Edit equipment
function editEquipment(equipmentId) {
  window.location.href = `/add-equipment?id=${equipmentId}`;
}

// Change status
function changeStatus(equipmentId) {
  const status = prompt('Enter new status (available/maintenance/reserved/booked):');
  if (status) {
    updateEquipmentStatus(equipmentId, status);
  }
}

// Update equipment status
async function updateEquipmentStatus(equipmentId, status) {
  try {
    const token = getToken();
    const response = await fetch(`${EQUIP_API_BASE}/equipment/${equipmentId}/status`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ status })
    });

    if (response.ok) {
      showToast('✅ Equipment status updated', 'success');
      loadEquipmentStats();
    } else {
      showToast('❌ Failed to update status', 'error');
    }
  } catch (error) {
    console.error('Error updating status:', error);
    showToast('❌ Error: ' + error.message, 'error');
  }
}

// Delete equipment
async function deleteEquipment(equipmentId) {
  if (!confirm('Are you sure you want to delete this equipment?')) return;

  try {
    const token = getToken();
    const response = await fetch(`${EQUIP_API_BASE}/equipment/${equipmentId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    if (response.ok) {
      showToast('✅ Equipment deleted', 'success');
      loadEquipmentStats();
    } else {
      showToast('❌ Failed to delete equipment', 'error');
    }
  } catch (error) {
    console.error('Error deleting equipment:', error);
    showToast('❌ Error: ' + error.message, 'error');
  }
}

// Add new equipment
function openAddEquipment() {
  window.location.href = '/add-equipment';
}

// Load equipment on page load
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    if (document.getElementById('equipmentTableBody')) {
      loadEquipmentStats();
    }
  });
} else {
  // DOM already loaded, call immediately if on equipment section
  if (document.getElementById('equipmentTableBody')) {
    loadEquipmentStats();
  }
}
