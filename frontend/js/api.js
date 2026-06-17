// GenRent API Client

const API_BASE = '/api/v1';

// ---- Token Management ----
function getToken() {
  return localStorage.getItem('genrent_token');
}

function setToken(token) {
  localStorage.setItem('genrent_token', token);
}

function setUser(user) {
  localStorage.setItem('genrent_user', JSON.stringify(user));
}

function getUser() {
  try {
    return JSON.parse(localStorage.getItem('genrent_user'));
  } catch {
    return null;
  }
}

function clearAuth() {
  localStorage.removeItem('genrent_token');
  localStorage.removeItem('genrent_user');
}

function isLoggedIn() {
  return !!getToken();
}

// ---- HTTP Helpers ----
async function apiRequest(method, path, body = null, auth = true) {
  const headers = { 'Content-Type': 'application/json' };
  if (auth && getToken()) {
    headers['Authorization'] = `Bearer ${getToken()}`;
  }

  const options = { method, headers };
  if (body) {
    options.body = JSON.stringify(body);
  }

  const response = await fetch(API_BASE + path, options);

  let data;
  const contentType = response.headers.get('content-type') || '';
  if (contentType.includes('application/json')) {
    data = await response.json();
  } else {
    const text = await response.text();
    data = { error: text || `HTTP ${response.status}` };
  }

  if (!response.ok) {
    throw new Error(data.error || `Request failed (${response.status})`);
  }

  return data;
}

const api = {
  get: (path, auth = true) => apiRequest('GET', path, null, auth),
  post: (path, body, auth = false) => apiRequest('POST', path, body, auth),
  put: (path, body) => apiRequest('PUT', path, body, true),
  delete: (path) => apiRequest('DELETE', path, null, true),

  // Auth
  register: (data) => api.post('/auth/register', data, false),
  login: (data) => api.post('/auth/login', data, false),
  getProfile: () => api.get('/auth/profile'),

  // Generators
  searchGenerators: (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/generators${query ? '?' + query : ''}`, false);
  },
  getGenerator: async (id) => {
    const eq = await api.get(`/equipment/${id}`, false);
    if (eq) {
      eq.price_per_day = eq.daily_price;
      eq.price_per_month = eq.monthly_price;
      eq.capacity_kva = eq.model ? parseInt(eq.model) || 0 : 0;
      eq.fuel_type = (eq.specs && eq.specs.fuel) || "diesel";
    }
    return eq;
  },
  createGenerator: (data) => api.post('/generators', data, true),
  updateGenerator: (id, data) => api.put(`/generators/${id}`, data),
  deleteGenerator: (id) => api.delete(`/generators/${id}`),
  getMyGenerators: async () => {
    try {
      const res = await api.get('/equipment/mine');
      if (res && res.equipment) {
        res.generators = res.equipment.map(eq => ({
          id: eq.id,
          name: eq.name,
          brand: eq.brand,
          capacity_kva: eq.model ? parseInt(eq.model) || 0 : 0,
          daily_price: eq.daily_price,
          monthly_price: eq.monthly_price,
          availability_status: eq.availability_status,
          city: eq.city,
          location: eq.location,
          description: eq.description,
          image_url: eq.image_url,
        }));
      }
      return res;
    } catch (err) {
      return api.get('/generators/mine');
    }
  },

  // Vendors
  listVendors: (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/vendors${query ? '?' + query : ''}`, false);
  },
  getVendor: (id) => api.get(`/vendors/${id}`, false),
  createVendor: (data) => api.post('/vendors', data, true),
  getMyVendorProfile: () => api.get('/vendors/me'),
  updateVendorProfile: (data) => api.put('/vendors/me', data),

  // Bookings
  createBooking: (data) => api.post('/bookings', data, true),
  getMyBookings: () => api.get('/bookings'),
  getBooking: (id) => api.get(`/bookings/${id}`),
  getBookingStatus: (id) => api.get(`/bookings/${id}/status`),
  updateBookingStatus: (id, status) => api.put(`/bookings/${id}/status`, { status }),

  // Payments
  getPaymentDetails: (bookingId) => api.get(`/payments/booking/${bookingId}`),
  processPayment: (data) => api.post('/payments', data, true),

  // Vendor Wallet
  getVendorWallet: () => api.get('/wallet'),

  // Notifications
  getNotifications: () => api.get('/notifications'),
  markNotificationRead: (id) => api.post(`/notifications/${id}/read`, {}, true),
  markAllRead: () => api.post('/notifications/read-all', {}, true),

  // Admin
  adminStats: () => api.get('/admin/stats'),
  adminVendors: (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/admin/vendors${query ? '?' + query : ''}`);
  },
  adminVerifyVendor: (id) => api.put(`/admin/vendors/${id}/verify`, {}),
  adminRejectVendor: (id) => api.put(`/admin/vendors/${id}/reject`, {}),
  adminGenerators: (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/admin/generators${query ? '?' + query : ''}`);
  },
  adminUpdateGeneratorStatus: (id, status) => api.put(`/admin/generators/${id}/status`, { status }),
  adminBookings: (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return api.get(`/admin/bookings${query ? '?' + query : ''}`);
  },
};

// ---- Toast Notifications ----
function showToast(message, type = 'info') {
  const container = document.getElementById('toastContainer');
  if (!container) return;

  const icons = { success: '✅', error: '❌', info: 'ℹ️' };
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.innerHTML = `<span>${icons[type] || ''}</span><span class="toast-message">${message}</span>`;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(100%)';
    toast.style.transition = 'all 0.3s ease';
    setTimeout(() => toast.remove(), 300);
  }, 3500);
}

// ---- Alert helper ----
function showAlert(containerId, message, type = 'error') {
  const el = document.getElementById(containerId);
  if (el) {
    el.innerHTML = `<div class="alert alert-${type}">${message}</div>`;
  }
}

function clearAlert(containerId) {
  const el = document.getElementById(containerId);
  if (el) el.innerHTML = '';
}

// ---- Helpers ----
function formatCurrency(amount) {
  return '₹' + Number(amount).toLocaleString('en-IN');
}

function formatDate(dateStr) {
  return new Date(dateStr).toLocaleDateString('en-IN', {
    day: 'numeric', month: 'short', year: 'numeric'
  });
}

function statusBadge(status) {
  return `<span class="status-badge status-${status}">${status}</span>`;
}

function logout() {
  clearAuth();
  window.location.href = '/login';
}
