// ─────────────────────────────────────────────────
//  FlashCart Frontend — app.js
//  Fixes: response envelope, price format, token key,
//         server-side cart sync, search, error UX
// ─────────────────────────────────────────────────
const API_BASE = '/api/v1';

// ── Helpers ──────────────────────────────────────

/** Extract payload from our standard { success, data } envelope */
function unwrap(json) {
    if (json && json.data !== undefined) return json.data;
    return json;
}

/** Format a dollar-float price → "$129.99" */
const formatPrice = (price) => '$' + Number(price).toFixed(2);

/** Read JWT token from localStorage */
const getToken = () => localStorage.getItem('fc_token');

/** Store JWT token */
const setToken = (t) => localStorage.setItem('fc_token', t);

/** Clear JWT token */
const clearToken = () => localStorage.removeItem('fc_token');

/** Authenticated fetch wrapper */
async function authFetch(url, options = {}) {
    const token = getToken();
    const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) };
    if (token) headers['Authorization'] = `Bearer ${token}`;
    const res = await fetch(url, { ...options, headers });
    const json = await res.json().catch(() => ({}));
    if (!res.ok) {
        const msg = json?.error?.message || json?.error || 'Request failed';
        throw new Error(msg);
    }
    return unwrap(json);
}

// ── Theme ─────────────────────────────────────────

function toggleTheme() {
    const html = document.documentElement;
    const isDark = html.classList.contains('dark');
    const themeIcon = document.getElementById('themeIcon');
    if (isDark) {
        html.classList.remove('dark');
        localStorage.setItem('fc_theme', 'light');
        if (themeIcon) themeIcon.textContent = 'dark_mode';
    } else {
        html.classList.add('dark');
        localStorage.setItem('fc_theme', 'dark');
        if (themeIcon) themeIcon.textContent = 'light_mode';
    }
}

// ── Session ───────────────────────────────────────

function checkSession() {
    const token = getToken();
    const authBtn   = document.getElementById('authBtn');
    const logoutBtn = document.getElementById('logoutBtn');
    if (token) {
        if (authBtn)   authBtn.classList.add('hidden');
        if (logoutBtn) logoutBtn.classList.remove('hidden');
    } else {
        if (authBtn)   authBtn.classList.remove('hidden');
        if (logoutBtn) logoutBtn.classList.add('hidden');
    }
}

function logout() {
    clearToken();
    localStorage.removeItem('fc_cart');
    window.location.href = '/index.html';
}

// ── Product Grid ──────────────────────────────────

let searchTimeout;
function handleSearch(event) {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => fetchProducts(event.target.value.trim()), 300);
}

async function fetchProducts(query = '') {
    const grid = document.getElementById('productGrid');
    if (!grid) return;

    // Show skeleton
    grid.innerHTML = Array(4).fill(`
        <div class="bg-surface-container border border-outline/20 rounded overflow-hidden animate-pulse">
            <div class="h-48 bg-surface"></div>
            <div class="p-6 space-y-3">
                <div class="h-4 bg-surface rounded w-3/4"></div>
                <div class="h-4 bg-surface rounded w-1/3"></div>
            </div>
        </div>`).join('');

    try {
        const url = query
            ? `${API_BASE}/products?search=${encodeURIComponent(query)}&limit=20`
            : `${API_BASE}/products?limit=20`;

        const res = await fetch(url);
        if (!res.ok) throw new Error('Failed to fetch products');
        const json = await res.json();

        // API shape: { success, data: { products: [...], meta: {...} } }
        const payload  = unwrap(json);
        const products = payload?.products ?? payload ?? [];

        if (!Array.isArray(products) || products.length === 0) {
            grid.innerHTML = `
                <div class="col-span-full flex flex-col items-center py-24 text-on-surface-variant gap-4">
                    <span class="material-symbols-outlined text-6xl opacity-30">search_off</span>
                    <p class="font-label-caps text-label-caps tracking-widest opacity-50">
                        ${query ? `No results for "${query}"` : 'No products available'}
                    </p>
                </div>`;
            return;
        }

        grid.innerHTML = products.map(p => `
            <div
                id="product-${p.id}"
                class="bg-surface-container border border-outline/20 rounded relative group overflow-hidden volt-glow transition-shadow cursor-pointer"
                onclick="window.location.href='/product.html?id=${p.id}'"
            >
                <div class="absolute top-0 left-0 w-full h-[2px] bg-primary"></div>
                <div class="h-48 relative overflow-hidden bg-surface flex items-center justify-center p-4">
                    <span class="material-symbols-outlined text-6xl text-outline-variant group-hover:text-primary transition-colors duration-300">inventory_2</span>
                </div>
                <div class="p-6 relative pb-14">
                    <h3 class="font-headline-sm text-lg text-on-surface mb-1 truncate">${escHtml(p.name)}</h3>
                    <div class="font-data-mono-md text-on-surface-variant mb-1 text-sm">${escHtml(p.brand || '')}</div>
                    <div class="font-data-mono-md text-primary font-bold">${formatPrice(p.price)}</div>
                </div>
                <button
                    onclick="event.stopPropagation(); addToCartServer('${p.id}', '${escHtml(p.name)}', ${p.price})"
                    class="reveal-btn absolute bottom-4 left-4 right-4 bg-primary text-on-primary font-bold py-2 rounded opacity-0 translate-y-2 transition-all duration-300 flex items-center justify-center gap-2 uppercase text-sm tracking-wider"
                >
                    <span class="material-symbols-outlined text-sm">add_shopping_cart</span>
                    Add to Cart
                </button>
            </div>
        `).join('');

    } catch (err) {
        console.error('fetchProducts error:', err);
        grid.innerHTML = `
            <div class="col-span-full flex flex-col items-center py-24 text-red-400 gap-4">
                <span class="material-symbols-outlined text-6xl">error</span>
                <p class="font-label-caps text-label-caps tracking-widest">Failed to load products. Is the API running?</p>
            </div>`;
    }
}

// Escape HTML to prevent XSS
function escHtml(str) {
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

// ── Cart (server-synced) ──────────────────────────

// Local cart mirrors the server for fast UI; truth is in Redis.
let localCart = JSON.parse(localStorage.getItem('fc_cart') || '[]');

function saveLocalCart() {
    localStorage.setItem('fc_cart', JSON.stringify(localCart));
    updateCartBadge();
}

function updateCartBadge() {
    const badge = document.getElementById('cartBadge');
    if (!badge) return;
    const count = localCart.reduce((s, i) => s + i.quantity, 0);
    if (count > 0) badge.classList.remove('hidden');
    else           badge.classList.add('hidden');
}

/** Add item to server cart (requires auth), fall back to local-only with a prompt */
async function addToCartServer(productId, productName, price) {
    const token = getToken();
    if (!token) {
        // Store locally and prompt login
        const existing = localCart.find(i => i.product_id === productId);
        if (existing) existing.quantity += 1;
        else localCart.push({ product_id: productId, name: productName, unit_price: price, quantity: 1 });
        saveLocalCart();
        showToast(`"${productName}" saved — log in to checkout`, 'info');
        return;
    }

    try {
        await authFetch(`${API_BASE}/cart/items`, {
            method: 'POST',
            body: JSON.stringify({ product_id: productId, quantity: 1 }),
        });

        // Keep local mirror in sync
        const existing = localCart.find(i => i.product_id === productId);
        if (existing) existing.quantity += 1;
        else localCart.push({ product_id: productId, name: productName, unit_price: price, quantity: 1 });
        saveLocalCart();

        showToast(`"${productName}" added to cart`, 'success');
    } catch (err) {
        showToast(err.message, 'error');
    }
}

/** Toast notification */
function showToast(message, type = 'info') {
    let container = document.getElementById('toastContainer');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toastContainer';
        container.className = 'fixed bottom-6 right-6 flex flex-col gap-2 z-[100]';
        document.body.appendChild(container);
    }

    const colorMap = {
        success: 'bg-green-600',
        error:   'bg-red-600',
        info:    'bg-primary',
    };

    const toast = document.createElement('div');
    toast.className = `${colorMap[type] || 'bg-primary'} text-white text-sm font-label-caps px-sm py-xs rounded shadow-lg transition-all duration-300 opacity-0 translate-y-2`;
    toast.textContent = message;
    container.appendChild(toast);

    // Animate in
    requestAnimationFrame(() => {
        toast.classList.remove('opacity-0', 'translate-y-2');
    });

    // Auto-remove after 3s
    setTimeout(() => {
        toast.classList.add('opacity-0', 'translate-y-2');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// ── Cart Page ─────────────────────────────────────

async function loadCart() {
    const container  = document.getElementById('cartItemsContainer');
    const subtotalEl = document.getElementById('cartSubtotal');
    if (!container) return;

    // Try to fetch server cart if logged in, else use local
    const token = getToken();
    let items = localCart;

    if (token) {
        try {
            const data = await authFetch(`${API_BASE}/cart`);
            // Server cart: { user_id, items: [{product_id, name, unit_price, quantity}] }
            if (data?.items && Array.isArray(data.items)) {
                items = data.items;
                localCart = items; // sync local mirror
                saveLocalCart();
            }
        } catch (e) {
            console.warn('Could not fetch server cart, using local copy:', e);
        }
    }

    if (!items || items.length === 0) {
        container.innerHTML = `
            <div class="flex flex-col items-center py-24 text-on-surface-variant gap-4">
                <span class="material-symbols-outlined text-6xl opacity-30">shopping_cart</span>
                <p class="font-label-caps tracking-widest opacity-50">Your cart is empty</p>
                <a href="/index.html" class="bg-primary text-on-primary font-label-caps px-sm py-xs rounded uppercase tracking-wider mt-2">Browse Products</a>
            </div>`;
        if (subtotalEl) subtotalEl.textContent = '$0.00';
        return;
    }

    let subtotal = 0;
    container.innerHTML = items.map(item => {
        const lineTotal = (item.unit_price || 0) * (item.quantity || 1);
        subtotal += lineTotal;
        return `
            <div class="flex gap-4 bg-surface-container border border-outline/20 p-4 rounded items-center">
                <div class="w-14 h-14 bg-surface rounded flex items-center justify-center flex-shrink-0">
                    <span class="material-symbols-outlined text-outline-variant">inventory_2</span>
                </div>
                <div class="flex-1 min-w-0">
                    <h3 class="font-headline-sm text-on-surface truncate">${escHtml(item.name || item.product_id)}</h3>
                    <p class="text-on-surface-variant text-sm mt-0.5">${formatPrice(item.unit_price || 0)} × ${item.quantity}</p>
                </div>
                <div class="font-data-mono-md text-primary font-bold flex-shrink-0">${formatPrice(lineTotal)}</div>
                <button onclick="removeFromCart('${item.product_id}')" class="text-outline-variant hover:text-red-400 transition-colors ml-2">
                    <span class="material-symbols-outlined">delete</span>
                </button>
            </div>`;
    }).join('');

    if (subtotalEl) subtotalEl.textContent = formatPrice(subtotal);
}

async function removeFromCart(productId) {
    const token = getToken();
    if (token) {
        try {
            await authFetch(`${API_BASE}/cart/items/${productId}`, { method: 'DELETE' });
        } catch (e) {
            console.warn('Server remove failed, removing locally:', e);
        }
    }
    localCart = localCart.filter(i => i.product_id !== productId);
    saveLocalCart();
    loadCart();
}

async function checkout() {
    const token  = getToken();
    const msgEl  = document.getElementById('checkoutMsg');
    const btn    = document.getElementById('checkoutBtn');

    if (!token) {
        window.location.href = '/auth.html';
        return;
    }

    btn.disabled    = true;
    btn.textContent = 'Processing...';

    try {
        const order = await authFetch(`${API_BASE}/orders`, { method: 'POST' });

        if (msgEl) {
            msgEl.className     = 'mt-4 text-sm text-center text-green-400';
            msgEl.textContent   = `✓ Order placed! ID: ${order?.id?.substring(0, 8) ?? 'confirmed'}`;
        }

        // Clear cart
        localCart = [];
        saveLocalCart();
        setTimeout(() => window.location.href = '/index.html', 2500);

    } catch (err) {
        if (msgEl) {
            msgEl.className   = 'mt-4 text-sm text-center text-red-400';
            msgEl.textContent = err.message;
        }
        btn.disabled    = false;
        btn.textContent = 'Proceed to Checkout';
    }
}

// ── Auth Page ─────────────────────────────────────

async function handleLogin(event) {
    event.preventDefault();
    const email    = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value;
    const errorDiv = document.getElementById('authError');

    try {
        // API returns { success, data: { user, access_token } }
        const res  = await fetch(`${API_BASE}/auth/login`, {
            method:  'POST',
            headers: { 'Content-Type': 'application/json' },
            body:    JSON.stringify({ email, password }),
        });
        const json = await res.json();

        if (!res.ok) {
            const msg = json?.error?.message || json?.error || 'Login failed';
            throw new Error(msg);
        }

        const payload = unwrap(json);
        const token   = payload?.access_token;
        if (!token) throw new Error('No token received from server');

        setToken(token);

        // Sync local-only cart items to server after login
        await syncLocalCartToServer();

        window.location.href = '/index.html';
    } catch (err) {
        if (errorDiv) {
            errorDiv.textContent = err.message;
            errorDiv.classList.remove('hidden');
        }
    }
}

async function handleRegister(event) {
    event.preventDefault();
    const email      = document.getElementById('email').value.trim();
    const password   = document.getElementById('password').value;
    const firstName  = (document.getElementById('first_name') || {}).value || '';
    const lastName   = (document.getElementById('last_name')  || {}).value || '';
    const errorDiv   = document.getElementById('authError');

    try {
        const res  = await fetch(`${API_BASE}/auth/register`, {
            method:  'POST',
            headers: { 'Content-Type': 'application/json' },
            body:    JSON.stringify({ email, password, first_name: firstName, last_name: lastName }),
        });
        const json = await res.json();

        if (!res.ok) {
            const msg = json?.error?.message || json?.error || 'Registration failed';
            throw new Error(msg);
        }

        const payload = unwrap(json);
        const token   = payload?.access_token;
        if (!token) throw new Error('No token received from server');

        setToken(token);
        await syncLocalCartToServer();
        window.location.href = '/index.html';
    } catch (err) {
        if (errorDiv) {
            errorDiv.textContent = err.message;
            errorDiv.classList.remove('hidden');
        }
    }
}

async function handleGoogleLogin(response) {
    const errorDiv = document.getElementById('authError');
    try {
        const res  = await fetch(`${API_BASE}/auth/google`, {
            method:  'POST',
            headers: { 'Content-Type': 'application/json' },
            body:    JSON.stringify({ credential: response.credential }),
        });
        const json = await res.json();

        if (!res.ok) {
            const msg = json?.error?.message || json?.error || 'Google login failed';
            throw new Error(msg);
        }

        const payload = unwrap(json);
        const token   = payload?.access_token;
        if (!token) throw new Error('No token received');

        setToken(token);
        await syncLocalCartToServer();
        window.location.href = '/index.html';
    } catch (err) {
        if (errorDiv) {
            errorDiv.textContent = err.message;
            errorDiv.classList.remove('hidden');
        }
    }
}

/** Push any items saved locally before login up to the server cart */
async function syncLocalCartToServer() {
    if (!localCart.length) return;
    for (const item of localCart) {
        try {
            await authFetch(`${API_BASE}/cart/items`, {
                method: 'POST',
                body:   JSON.stringify({ product_id: item.product_id, quantity: item.quantity }),
            });
        } catch (e) {
            console.warn('Failed to sync cart item to server:', item.product_id, e);
        }
    }
}

// ── Product Page ──────────────────────────────────

async function loadProductPage() {
    const params    = new URLSearchParams(window.location.search);
    const productId = params.get('id');
    if (!productId) { window.location.href = '/index.html'; return; }

    try {
        const res     = await fetch(`${API_BASE}/products/${productId}`);
        if (!res.ok) throw new Error('Product not found');
        const json    = await res.json();
        const product = unwrap(json);

        document.getElementById('productName').textContent        = product.name;
        document.getElementById('productBrand').textContent       = product.brand || '';
        document.getElementById('productPrice').textContent       = formatPrice(product.price);
        document.getElementById('productDescription').textContent = product.description || '';

        const addBtn = document.getElementById('addToCartBtn');
        if (addBtn) addBtn.onclick = () => addToCartServer(product.id, product.name, product.price);

        // Recommendations
        try {
            const recJson = await fetch(`${API_BASE}/products?limit=5`).then(r => r.json());
            let recs = (unwrap(recJson)?.products ?? []).filter(p => p.id !== product.id).slice(0, 4);
            const recGrid = document.getElementById('recommendationsGrid');
            if (recGrid && recs.length) {
                recGrid.innerHTML = recs.map(p => `
                    <div class="bg-surface-container border border-outline/20 rounded relative group overflow-hidden volt-glow cursor-pointer" onclick="window.location.href='/product.html?id=${p.id}'">
                        <div class="h-32 bg-surface flex items-center justify-center">
                            <span class="material-symbols-outlined text-4xl text-outline-variant group-hover:text-primary transition-colors">inventory_2</span>
                        </div>
                        <div class="p-4">
                            <h3 class="font-headline-sm text-on-surface truncate">${escHtml(p.name)}</h3>
                            <div class="font-data-mono-md text-primary text-sm">${formatPrice(p.price)}</div>
                        </div>
                    </div>`).join('');
            }
        } catch (_) {}

        await loadReviews(productId);

        const token = getToken();
        if (token) {
            const loginPrompt  = document.getElementById('loginToReviewPrompt');
            const reviewForm   = document.getElementById('reviewFormContainer');
            if (loginPrompt) loginPrompt.classList.add('hidden');
            if (reviewForm)  reviewForm.classList.remove('hidden');
        }

        document.getElementById('loadingState')?.classList.add('hidden');
        document.getElementById('productContent')?.classList.remove('hidden');

    } catch (err) {
        console.error(err);
        const ls = document.getElementById('loadingState');
        if (ls) ls.innerHTML = `<p class="text-red-400">Failed to load product.</p>`;
    }
}

async function loadReviews(productId) {
    try {
        const res     = await fetch(`${API_BASE}/products/${productId}/reviews`);
        if (!res.ok) return;
        const json    = await res.json();
        const reviews = unwrap(json) ?? [];
        const list    = document.getElementById('reviewsList');
        if (!list) return;

        if (!reviews.length) {
            list.innerHTML = '<p class="text-on-surface-variant">No reviews yet. Be the first!</p>';
            return;
        }

        list.innerHTML = reviews.map(r => `
            <div class="bg-surface p-4 rounded border border-outline/20">
                <div class="flex items-center gap-2 mb-2">
                    <span class="font-bold text-on-surface">${escHtml(r.user_first_name || 'User')}</span>
                    <span class="text-yellow-400 text-sm">★ ${r.rating}/5</span>
                </div>
                <p class="text-on-surface-variant text-sm">${escHtml(r.comment || '')}</p>
            </div>`).join('');
    } catch (e) { console.error(e); }
}

async function submitReview(event) {
    event.preventDefault();
    const params    = new URLSearchParams(window.location.search);
    const productId = params.get('id');
    const rating    = parseInt(document.getElementById('reviewRating').value);
    const comment   = document.getElementById('reviewComment').value;
    const errorDiv  = document.getElementById('reviewError');

    try {
        await authFetch(`${API_BASE}/products/${productId}/reviews`, {
            method: 'POST',
            body:   JSON.stringify({ rating, comment }),
        });
        if (errorDiv) errorDiv.classList.add('hidden');
        document.getElementById('reviewComment').value = '';
        await loadReviews(productId);
    } catch (err) {
        if (errorDiv) {
            errorDiv.textContent = err.message;
            errorDiv.classList.remove('hidden');
        }
    }
}

// ── Initialisation ────────────────────────────────

// Restore theme
const savedTheme = localStorage.getItem('fc_theme');
if (savedTheme === 'light') {
    document.documentElement.classList.remove('dark');
    const ti = document.getElementById('themeIcon');
    if (ti) ti.textContent = 'dark_mode';
}

checkSession();
updateCartBadge();

const path = window.location.pathname;
if (path === '/' || path === '/index.html' || path.endsWith('/')) {
    fetchProducts();
} else if (path.includes('product.html')) {
    loadProductPage();
} else if (path.includes('cart.html')) {
    loadCart();
}
