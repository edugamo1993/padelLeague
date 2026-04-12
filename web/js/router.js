// Router
const router = {
    init() {
        this.bindEvents();
        this.checkRoute();
    },

    bindEvents() {
        // Navigation links
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                this.navigate(link.getAttribute('data-page'));
                document.getElementById('sidebar').classList.remove('active');
                document.getElementById('sidebarOverlay').classList.remove('active');
            });
        });

        // Close modals on backdrop click
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) modal.classList.remove('active');
            });
        });

        document.querySelectorAll('.close-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                btn.closest('.modal').classList.remove('active');
            });
        });

        // Logout
        document.getElementById('logoutBtn').addEventListener('click', () => auth.logout());

        // Sidebar toggle (inside sidebar — mobile close)
        document.getElementById('sidebarToggle').addEventListener('click', () => {
            document.getElementById('sidebar').classList.remove('active');
            document.getElementById('sidebarOverlay').classList.remove('active');
        });

        // Hamburger button (mobile top bar)
        document.getElementById('hamburgerBtn').addEventListener('click', () => {
            document.getElementById('sidebar').classList.toggle('active');
            document.getElementById('sidebarOverlay').classList.toggle('active');
        });

        // Sidebar overlay click
        document.getElementById('sidebarOverlay').addEventListener('click', () => {
            document.getElementById('sidebar').classList.remove('active');
            document.getElementById('sidebarOverlay').classList.remove('active');
        });

        // Match result form
        document.getElementById('submitResultBtn').addEventListener('click', () => matches.submitResult());
    },

    navigate(page) {
        // Update active nav link
        document.querySelectorAll('.nav-link').forEach(link => link.classList.remove('active'));
        const activeLink = document.querySelector(`.nav-link[data-page="${page}"]`);
        if (activeLink) activeLink.classList.add('active');

        // Show page
        document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
        const pageEl = document.getElementById(`${page}Page`);
        if (pageEl) pageEl.classList.add('active');

        // Load page content
        switch (page) {
            case 'dashboard':
                state.userType === 'club' ? clubDashboard.load() : dashboard.load();
                break;
            case 'clubDashboard':
                clubDashboard.load();
                break;
            case 'clubs':
                clubs.load();
                break;
            case 'leagues':
                leagues.load();
                break;
            case 'matches':
                state.userType === 'club' ? this.navigate('clubDashboard') : matches.load();
                break;
            case 'profile':
                state.userType === 'club' ? this.navigate('clubDashboard') : profile.load();
                break;
        }
    },

    checkRoute() {
        const hash = window.location.hash.replace('#', '') || 'dashboard';
        this.navigate(hash);

        window.addEventListener('hashchange', () => {
            const hash = window.location.hash.replace('#', '') || 'dashboard';
            this.navigate(hash);
        });
    }
};
