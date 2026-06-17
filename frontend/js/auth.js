// GenRent Auth Page JavaScript

// ---- Login ----
async function handleLogin(event) {
  event.preventDefault();
  clearAlert('alertBox');

  const btn = document.getElementById('loginBtn');
  const email = document.getElementById('email').value;
  const password = document.getElementById('password').value;

  btn.disabled = true;
  btn.innerHTML = '<span class="loading-spinner"></span> Logging in...';

  try {
    const data = await api.login({ email, password });
    setToken(data.token);
    setUser(data.user);

    showAlert('alertBox', 'Login successful! Redirecting...', 'success');

    // Redirect based on role
    setTimeout(() => {
      const role = data.user.role;
      if (role === 'admin') {
        window.location.href = '/admin-dashboard';
      } else if (role === 'vendor') {
        window.location.href = '/vendor-dashboard';
      } else {
        const redirect = new URLSearchParams(window.location.search).get('redirect') || '/';
        window.location.href = redirect;
      }
    }, 500);
  } catch (err) {
    showAlert('alertBox', err.message, 'error');
    btn.disabled = false;
    btn.textContent = 'Login';
  }
}

// ---- Register ----
let selectedRole = 'customer';

function setRole(role) {
  selectedRole = role;
  document.getElementById('selectedRole').value = role;

  document.getElementById('roleCustomer').classList.toggle('active', role === 'customer');
  document.getElementById('roleVendor').classList.toggle('active', role === 'vendor');
}

async function handleRegister(event) {
  event.preventDefault();
  clearAlert('alertBox');

  const btn = document.getElementById('registerBtn');
  const name = document.getElementById('name').value;
  const email = document.getElementById('email').value;
  const phone = document.getElementById('phone').value;
  const password = document.getElementById('password').value;
  const role = document.getElementById('selectedRole').value;

  btn.disabled = true;
  btn.innerHTML = '<span class="loading-spinner"></span> Creating account...';

  try {
    await api.register({ name, email, phone, password, role });

    // Auto login after register
    const loginData = await api.login({ email, password });
    setToken(loginData.token);
    setUser(loginData.user);

    showAlert('alertBox', 'Account created! Redirecting...', 'success');

    setTimeout(() => {
      if (role === 'vendor') {
        window.location.href = '/vendor-dashboard';
      } else {
        window.location.href = '/';
      }
    }, 700);
  } catch (err) {
    showAlert('alertBox', err.message, 'error');
    btn.disabled = false;
    btn.textContent = 'Create Account';
  }
}

// ---- Redirect if already logged in ----
document.addEventListener('DOMContentLoaded', () => {
  const user = getUser();
  if (isLoggedIn() && user) {
    // Only redirect if on login/register page
    const path = window.location.pathname;
    if (path === '/login' || path === '/register') {
      if (user.role === 'admin') window.location.href = '/admin-dashboard';
      else if (user.role === 'vendor') window.location.href = '/vendor-dashboard';
      else window.location.href = '/';
    }
  }
});
