const API_BASE = '/api/v1';

// Theme Logic
function toggleTheme() {
    const html = document.documentElement;
    const isDark = html.classList.contains('dark');
    const themeIcon = document.getElementById('themeIcon');
    
    if (isDark) {
        html.classList.remove('dark');
        localStorage.setItem('theme', 'light');
        if (themeIcon) themeIcon.textContent = 'dark_mode';
    } else {
        html.classList.add('dark');
        localStorage.setItem('theme', 'dark');
        if (themeIcon) themeIcon.textContent = 'light_mode';
    }
}

// Session Check
function checkSession() {
    const token = localStorage.getItem('token');
    const authBtn = document.getElementById('authBtn');
    const logoutBtn = document.getElementById('logoutBtn');
    
    if (token) {
        if (authBtn) authBtn.classList.add('hidden');
        if (logoutBtn) logoutBtn.classList.remove('hidden');
    } else {
        if (authBtn) authBtn.classList.remove('hidden');
        if (logoutBtn) logoutBtn.classList.add('hidden');
    }
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('cart');
    window.location.href = '/index.html';
}

// Format currency
const formatPrice = (cents) => {
    return '$' + (cents / 100).toFixed(2);
};

// Search Logic
let searchTimeout;
function handleSearch(event) {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
        fetchProducts(event.target.value);
    }, 300);
}

// Fetch products from Go Backend
async function fetchProducts(query = '') {
    try {
        const url = query ? `${API_BASE}/products?search=${encodeURIComponent(query)}` : `${API_BASE}/products`;
        const response = await fetch(url);
        if (!response.ok) throw new Error('Failed to fetch products');
        
        const products = await response.json();
        const grid = document.getElementById('productGrid');
        if (!grid) return;

        grid.innerHTML = products.map(p => `
            <div class="bg-surface-container border border-outline/20 rounded relative group overflow-hidden volt-glow transition-shadow cursor-pointer" onclick="window.location.href='/product.html?id=${p.id}'">
                <div class="absolute top-0 left-0 w-full h-[2px] bg-primary"></div>
                <div class="h-48 relative overflow-hidden bg-surface flex items-center justify-center p-4">
                    <span class="material-symbols-outlined text-6xl text-gray-500 group-hover:text-primary transition-colors">inventory_2</span>
                </div>
                <div class="p-6 relative">
                    <h3 class="font-headline-sm text-lg text-on-surface mb-2 truncate">${p.name}</h3>
                    <div class="font-data-mono-md text-on-surface-variant mb-6">${formatPrice(p.price_cents)}</div>
                    <button onclick="event.stopPropagation(); addToCart('${p.id}')" class="reveal-btn absolute bottom-4 left-4 right-4 bg-primary text-on-primary font-bold py-2 rounded opacity-0 translate-y-2 transition-all duration-300 flex items-center justify-center gap-2 uppercase">
                        <span class="material-symbols-outlined text-sm">add_shopping_cart</span>
                        Add to Cart
                    </button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error fetching products:', error);
    }
}

// Authentication Logic
async function handleLogin(event) {
    event.preventDefault();
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const errorDiv = document.getElementById('authError');
    
    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || 'Login failed');
        
        localStorage.setItem('token', data.token);
        window.location.href = '/index.html';
    } catch (error) {
        errorDiv.textContent = error.message;
        errorDiv.classList.remove('hidden');
    }
}

async function handleGoogleLogin(response) {
    const errorDiv = document.getElementById('authError');
    try {
        const res = await fetch(`${API_BASE}/auth/google`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ credential: response.credential })
        });
        
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Google Login failed');
        
        localStorage.setItem('token', data.token);
        window.location.href = '/index.html';
    } catch (error) {
        errorDiv.textContent = error.message;
        errorDiv.classList.remove('hidden');
    }
}

// Cart Logic
let cart = JSON.parse(localStorage.getItem('cart')) || [];

function addToCart(productId) {
    const existing = cart.find(item => item.product_id === productId);
    if (existing) {
        existing.quantity += 1;
    } else {
        cart.push({ product_id: productId, quantity: 1 });
    }
    localStorage.setItem('cart', JSON.stringify(cart));
    
    // Show badge
    const badge = document.getElementById('cartBadge');
    if (badge) badge.classList.remove('hidden');
    
    alert('Item added to cart!');
}

async function loadCart() {
    const container = document.getElementById('cartItemsContainer');
    const subtotalEl = document.getElementById('cartSubtotal');
    if (!container) return;

    if (cart.length === 0) {
        container.innerHTML = '<p class="text-gray-500 text-center mt-10">Your cart is empty.</p>';
        return;
    }

    // For simplicity, we just show Product IDs in this demo.
    // In a full app, we'd fetch product details by ID or hydrate the cart.
    container.innerHTML = cart.map(item => `
        <div class="flex gap-4 bg-surface-container/50 p-4 rounded-lg border border-outline/20">
            <div class="w-16 h-16 bg-gray-800 rounded flex items-center justify-center">
                <span class="material-symbols-outlined text-gray-500">inventory_2</span>
            </div>
            <div class="flex-1">
                <h3 class="font-semibold text-sm">Product ${item.product_id.substring(0, 8)}...</h3>
                <p class="text-gray-400 text-xs mt-1">Qty: ${item.quantity}</p>
            </div>
            <button onclick="removeFromCart('${item.product_id}')" class="text-gray-500 hover:text-red-400">
                <span class="material-symbols-outlined">delete</span>
            </button>
        </div>
    `).join('');
    
    // Mock subtotal calculation
    subtotalEl.textContent = 'Calculating...';
}

function removeFromCart(productId) {
    cart = cart.filter(item => item.product_id !== productId);
    localStorage.setItem('cart', JSON.stringify(cart));
    loadCart();
}

async function checkout() {
    const token = localStorage.getItem('token');
    const msgEl = document.getElementById('checkoutMsg');
    const btn = document.getElementById('checkoutBtn');
    
    if (!token) {
        window.location.href = '/auth.html';
        return;
    }

    btn.disabled = true;
    btn.textContent = 'Processing...';

    try {
        const response = await fetch(`${API_BASE}/orders`, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (!response.ok) {
            const err = await response.json();
            throw new Error(err.error || 'Checkout failed');
        }
        
        msgEl.className = 'mt-4 text-sm text-center text-green-400';
        msgEl.textContent = 'Order Placed Successfully!';
        localStorage.removeItem('cart');
        cart = [];
        setTimeout(() => window.location.href = '/index.html', 2000);
    } catch (error) {
        msgEl.className = 'mt-4 text-sm text-center text-red-400';
        msgEl.textContent = error.message;
        btn.disabled = false;
        btn.textContent = 'Proceed to Checkout';
    }
}

// Initialization
const savedTheme = localStorage.getItem('theme');
if (savedTheme === 'light') {
    document.documentElement.classList.remove('dark');
    const themeIcon = document.getElementById('themeIcon');
    if (themeIcon) themeIcon.textContent = 'dark_mode';
}

checkSession();

if (window.location.pathname === '/' || window.location.pathname === '/index.html') {
    fetchProducts();
    if (cart.length > 0) {
        const badge = document.getElementById('cartBadge');
        if (badge) badge.classList.remove('hidden');
    }
} else if (window.location.pathname === '/product.html') {
    loadProductPage();
    if (cart.length > 0) {
        const badge = document.getElementById('cartBadge');
        if (badge) badge.classList.remove('hidden');
    }
}

// Product Page Logic
async function loadProductPage() {
    const params = new URLSearchParams(window.location.search);
    const productId = params.get('id');
    if (!productId) {
        window.location.href = '/index.html';
        return;
    }

    try {
        // Fetch Product Details
        const res = await fetch(`${API_BASE}/products/${productId}`);
        if (!res.ok) throw new Error('Product not found');
        const product = await res.json();

        document.getElementById('productName').textContent = product.name;
        document.getElementById('productBrand').textContent = product.brand;
        document.getElementById('productPrice').textContent = formatPrice(product.price_cents);
        document.getElementById('productDescription').textContent = product.description;
        
        const addBtn = document.getElementById('addToCartBtn');
        addBtn.onclick = () => addToCart(product.id);

        // Fetch Recommendations (just a simple limit fetch)
        const recRes = await fetch(`${API_BASE}/products?limit=4`);
        if (recRes.ok) {
            let recs = await recRes.json();
            recs = recs.filter(p => p.id !== product.id).slice(0, 4);
            const recGrid = document.getElementById('recommendationsGrid');
            if (recGrid) {
                recGrid.innerHTML = recs.map(p => `
                    <div class="bg-surface-container border border-outline/20 rounded relative group overflow-hidden volt-glow transition-shadow cursor-pointer" onclick="window.location.href='/product.html?id=${p.id}'">
                        <div class="h-32 relative overflow-hidden bg-surface flex items-center justify-center p-4">
                            <span class="material-symbols-outlined text-4xl text-gray-500 group-hover:text-primary transition-colors">inventory_2</span>
                        </div>
                        <div class="p-4 relative">
                            <h3 class="font-headline-sm text-md text-on-surface mb-1 truncate">${p.name}</h3>
                            <div class="font-data-mono-md text-on-surface-variant text-sm">${formatPrice(p.price_cents)}</div>
                        </div>
                    </div>
                `).join('');
            }
        }

        // Fetch Reviews
        await loadReviews(productId);

        // Handle Review Form Visibility
        const token = localStorage.getItem('token');
        if (token) {
            document.getElementById('loginToReviewPrompt').classList.add('hidden');
            document.getElementById('reviewFormContainer').classList.remove('hidden');
        }

        document.getElementById('loadingState').classList.add('hidden');
        document.getElementById('productContent').classList.remove('hidden');

    } catch (error) {
        console.error(error);
        document.getElementById('loadingState').innerHTML = '<p class="text-red-400">Failed to load product.</p>';
    }
}

async function loadReviews(productId) {
    try {
        const res = await fetch(`${API_BASE}/products/${productId}/reviews`);
        if (!res.ok) throw new Error('Failed to fetch reviews');
        const reviews = await res.json();
        
        const list = document.getElementById('reviewsList');
        if (!reviews || reviews.length === 0) {
            list.innerHTML = '<p class="text-on-surface-variant">No reviews yet. Be the first!</p>';
            return;
        }

        list.innerHTML = reviews.map(r => `
            <div class="bg-surface p-4 rounded border border-outline/20">
                <div class="flex items-center gap-2 mb-2">
                    <span class="font-bold text-on-surface">${r.user_first_name || 'User'}</span>
                    <span class="text-yellow-400 text-sm">★ ${r.rating}/5</span>
                </div>
                <p class="text-on-surface-variant text-sm">${r.comment}</p>
            </div>
        `).join('');
    } catch (error) {
        console.error(error);
    }
}

async function submitReview(event) {
    event.preventDefault();
    const token = localStorage.getItem('token');
    const params = new URLSearchParams(window.location.search);
    const productId = params.get('id');
    const rating = parseInt(document.getElementById('reviewRating').value);
    const comment = document.getElementById('reviewComment').value;
    const errorDiv = document.getElementById('reviewError');

    try {
        const res = await fetch(`${API_BASE}/products/${productId}/reviews`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ rating, comment })
        });

        if (!res.ok) {
            const data = await res.json();
            throw new Error(data.error || 'Failed to submit review');
        }

        errorDiv.classList.add('hidden');
        document.getElementById('reviewComment').value = '';
        await loadReviews(productId);
        
    } catch (error) {
        errorDiv.textContent = error.message;
        errorDiv.classList.remove('hidden');
    }
}
