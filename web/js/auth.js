// Authentication Module
const auth = {
    currentRole: 'player', // 'player' | 'club'

    init() {
        const savedRole = localStorage.getItem('padel_user_type');
        if (savedRole === 'club' || savedRole === 'player') this.currentRole = savedRole;
        this.applyRoleUI();
        this.checkAuth();
        this.bindEvents();
    },

    setRole(role) {
        if (role !== 'player' && role !== 'club') return;
        this.currentRole = role;
        state.userType = role;
        localStorage.setItem('padel_user_type', role);
        this.applyRoleUI();
    },

    applyRoleUI() {
        const roleHint = document.getElementById('authRoleHint');
        const playerBtn = document.getElementById('rolePlayerBtn');
        const clubBtn = document.getElementById('roleClubBtn');
        if (!roleHint || !playerBtn || !clubBtn) return;

        if (this.currentRole === 'club') {
            roleHint.textContent = 'Modo: Club';
            playerBtn.classList.remove('active');
            clubBtn.classList.add('active');
        } else {
            roleHint.textContent = 'Modo: Jugador';
            playerBtn.classList.add('active');
            clubBtn.classList.remove('active');
        }
    },

    updateLayoutForRole() {
        const isClub = this.currentRole === 'club';
        document.querySelectorAll('.player-only').forEach(el => { el.style.display = isClub ? 'none' : ''; });
        document.querySelectorAll('.club-only').forEach(el => { el.style.display = isClub ? '' : 'none'; });
        router.navigate(isClub ? 'clubDashboard' : 'dashboard');
    },

    loadDashboard() {
        return state.userType === 'club' ? clubDashboard.load() : dashboard.load();
    },

    checkAuth() {
        const urlParams = new URLSearchParams(window.location.search);
        const token = urlParams.get('token');
        const needsProfile = urlParams.get('needs_profile');

        if (token) {
            localStorage.setItem(TOKEN_KEY, token);
            window.history.replaceState({}, document.title, window.location.pathname);
            if (needsProfile === 'true') {
                this.hideAuthModal();
                this.showProfileForm();
            } else {
                this.hideAuthModal();
                this.updateLayoutForRole();
                this.loadDashboard();
            }
            return;
        }

        const storedToken = localStorage.getItem(TOKEN_KEY);
        if (storedToken) {
            this.checkGoogleUserNeedsProfile(storedToken);
        } else {
            this.showAuthModal();
        }
    },

    async checkGoogleUserNeedsProfile(token) {
        try {
            const response = await fetch(`${API_BASE_URL}/profile`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });

            if (response.ok) {
                const profileData = await response.json();
                if (profileData.role) { state.userType = profileData.role; this.setRole(profileData.role); }
                state.user = profileData;

                if (profileData.is_google_user && !profileData.hasProfile) {
                    this.hideAuthModal();
                    this.showProfileForm();
                } else {
                    this.hideAuthModal();
                    this.updateLayoutForRole();
                    this.loadDashboard();
                }
            } else {
                this.hideAuthModal();
                this.showProfileForm();
            }
        } catch (error) {
            console.error('Error checking profile:', error);
            this.hideAuthModal();
            this.showProfileForm();
        }
    },

    showProfileForm() {
        router.navigate('profile');
        profile.load(true);
    },

    async login(email, password) {
        try {
            const response = await api.login(email, password, this.currentRole);
            if (response.token) {
                localStorage.setItem(TOKEN_KEY, response.token);
                state.user = response.user || { name: email, role: this.currentRole };
                state.userType = response.user?.role || this.currentRole;
                this.setRole(state.userType);
                this.hideAuthModal();
                this.updateLayoutForRole();
                this.loadDashboard();
                utils.showToast('¡Bienvenido de nuevo!', 'success');
            }
        } catch (error) {
            utils.showToast('Error al iniciar sesión', 'error');
        }
    },

    async register(name, email, password) {
        try {
            const response = await api.register(name, email, password, this.currentRole);
            if (response.token) {
                localStorage.setItem(TOKEN_KEY, response.token);
                state.user = response.user || { name, email, role: this.currentRole };
                state.userType = response.user?.role || this.currentRole;
                this.setRole(state.userType);
                this.hideAuthModal();
                this.updateLayoutForRole();
                this.loadDashboard();
                utils.showToast('¡Registro exitoso!', 'success');
            }
        } catch (error) {
            utils.showToast('Error al registrarse', 'error');
        }
    },

    googleLogin() {
        window.location.href = `${API_BASE_URL}/auth/google/login`;
    },

    logout() {
        localStorage.removeItem(TOKEN_KEY);
        state.user = null;
        this.showAuthModal();
        utils.showToast('Sesión cerrada', 'info');
    },

    showAuthModal() {
        document.getElementById('authModal').classList.add('active');
        document.getElementById('app').style.display = 'none';
    },

    hideAuthModal() {
        document.getElementById('authModal').classList.remove('active');
        document.getElementById('app').style.display = 'grid';
    },

    bindEvents() {
        document.getElementById('rolePlayerBtn').addEventListener('click', () => this.setRole('player'));
        document.getElementById('roleClubBtn').addEventListener('click', () => this.setRole('club'));
        document.getElementById('googleLoginBtn').addEventListener('click', () => this.googleLogin());

        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            this.login(document.getElementById('loginEmail').value, document.getElementById('loginPassword').value);
        });

        document.getElementById('registerForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            this.register(
                document.getElementById('registerName').value,
                document.getElementById('registerEmail').value,
                document.getElementById('registerPassword').value
            );
        });

        document.getElementById('showRegister').addEventListener('click', (e) => {
            e.preventDefault();
            document.getElementById('loginForm').classList.remove('active');
            document.getElementById('registerForm').classList.add('active');
        });

        document.getElementById('showLogin').addEventListener('click', (e) => {
            e.preventDefault();
            document.getElementById('registerForm').classList.remove('active');
            document.getElementById('loginForm').classList.add('active');
        });
    }
};
