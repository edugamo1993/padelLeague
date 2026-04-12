// Player Dashboard Module
const dashboard = {
    async load() {
        document.getElementById('currentUser').textContent = state.user?.name || 'Invitado';
        document.getElementById('userRole').textContent = state.userType === 'club' ? 'Club' : 'Jugador';
        await this.loadUserStats();
        await this.loadUserClubs();
        await this.loadUpcomingMatches();
    },

    async loadUserStats() {
        try {
            const history = await api.getUserHistory(state.user.id);
            document.getElementById('statMatches').textContent = history.matches.length;
            document.getElementById('statWins').textContent = history.matches.filter(m => m.won).length;
            document.getElementById('statLosses').textContent = history.matches.filter(m => !m.won).length;
        } catch (error) {
            console.error('Error loading stats:', error);
        }
    },

    async loadUserClubs() {
        try {
            const clubs = await api.getClubs();
            const clubsList = document.getElementById('clubsList');
            clubsList.innerHTML = '';

            if (clubs.length === 0) {
                clubsList.innerHTML = '<p class="text-center text-light">No estás en ningún club. Únete a uno o crea tu propio club.</p>';
                return;
            }

            clubs.forEach(club => {
                const item = document.createElement('div');
                item.className = 'list-item';
                item.innerHTML = `
                    <div>
                        <h4>${club.name}</h4>
                        <p>${club.location || 'Sin ubicación'}</p>
                    </div>
                    <div class="club-actions">
                        <button class="btn btn-sm btn-primary" onclick="dashboard.viewClub('${club.id}')">
                            <i class="fas fa-eye"></i> Ver
                        </button>
                    </div>
                `;
                clubsList.appendChild(item);
            });
        } catch (error) {
            console.error('Error loading clubs:', error);
        }
    },

    async loadUpcomingMatches() {
        try {
            const matches = await api.getMatches();
            const upcomingMatches = document.getElementById('upcomingMatches');
            upcomingMatches.innerHTML = '';

            const pending = matches.filter(m => m.status === 'pending');
            if (pending.length === 0) {
                upcomingMatches.innerHTML = '<p class="text-center text-light">No hay partidos pendientes.</p>';
                return;
            }

            pending.slice(0, 3).forEach(match => {
                const item = document.createElement('div');
                item.className = 'list-item';
                item.innerHTML = `
                    <div>
                        <h4>Partido ${match.id}</h4>
                        <p>${utils.formatDate(match.createdAt)}</p>
                    </div>
                    <div class="match-actions">
                        <button class="btn btn-sm btn-primary" onclick="matches.viewMatch('${match.id}')">
                            <i class="fas fa-eye"></i> Ver
                        </button>
                    </div>
                `;
                upcomingMatches.appendChild(item);
            });
        } catch (error) {
            console.error('Error loading matches:', error);
        }
    },

    viewClub(clubId) {
        state.currentClub = clubId;
        router.navigate('clubs');
    }
};

// Club Dashboard Module
const clubDashboard = {
    pendingDelete: null,

    async load() {
        this.renderHeader();
        await this.loadAdminClubs();
        await this.loadLeagues();
        this.bindActions();
    },

    renderHeader() {
        document.getElementById('currentClubUser').textContent = state.user?.name || 'Club';
        document.getElementById('userRoleClub').textContent = 'Club';
    },

    async loadAdminClubs() {
        try {
            const clubs = await api.getUserClubs();
            const list = document.getElementById('clubAdminClubsList');
            list.innerHTML = '';

            if (!clubs || clubs.length === 0) {
                list.innerHTML = '<p class="text-center text-light">No gestionas ningún club aún.</p>';
                return;
            }

            if (!state.currentClub && clubs.length > 0) {
                state.currentClub = clubs[0].club_id || clubs[0].id;
            }

            clubs.forEach(club => {
                const item = document.createElement('div');
                item.className = 'list-item';
                item.innerHTML = `
                    <div>
                        <h4>${club.club_name || club.name || 'Club'}</h4>
                        <p>${club.location || 'Sin ubicación'}</p>
                    </div>
                    <div>
                        <span class="status-badge">${club.role || 'admin'}</span>
                    </div>
                `;
                list.appendChild(item);
            });
        } catch (error) {
            console.error('Error loading club admin clubs:', error);
        }
    },

    async loadLeagues() {
        try {
            const leagues = await api.getLeagues();
            const list = document.getElementById('clubLeaguesList');
            list.innerHTML = '';

            if (!leagues || leagues.length === 0) {
                list.innerHTML = '<p class="text-center text-light">No hay ligas disponibles.</p>';
                return;
            }

            leagues.forEach(league => {
                const leagueName = league.name || league.Name || 'Sin nombre';
                const leagueSeason = league.season || league.Season || '-';
                const leagueId = league.id || league.ID;
                const clubName = league.club?.name || league.club?.Name || 'Desconocido';

                const item = document.createElement('div');
                item.className = 'list-item';
                item.innerHTML = `
                    <div>
                        <h4>${leagueName}</h4>
                        <p>Temporada: ${leagueSeason} - Club: ${clubName}</p>
                    </div>
                    <div>
                        <button class="btn btn-sm btn-primary" onclick="leagues.viewLeague('${leagueId}')">Ver</button>
                        <button class="btn btn-sm btn-danger" onclick="clubDashboard.deleteLeague('${leagueId}', '${leagueName}')">
                            <i class="fas fa-trash"></i> Borrar
                        </button>
                    </div>
                `;
                list.appendChild(item);
            });
        } catch (error) {
            console.error('Error loading club leagues:', error);
        }
    },

    showCreateLeagueModal() {
        document.getElementById('newLeagueName').value = '';
        document.getElementById('newLeagueSeason').value = '';
        document.getElementById('createLeagueModal').classList.add('active');
    },

    hideCreateLeagueModal() {
        document.getElementById('createLeagueModal').classList.remove('active');
    },

    async submitCreateLeague() {
        const clubId = state.currentClub || '';
        if (!clubId) { utils.showToast('Selecciona un club primero.', 'error'); return; }

        const name = document.getElementById('newLeagueName').value.trim();
        const season = document.getElementById('newLeagueSeason').value.trim();
        if (!name || !season) { utils.showToast('Completa nombre y temporada.', 'error'); return; }

        try {
            await api.createLeague(clubId, name, season);
            utils.showToast('Liga creada', 'success');
            this.hideCreateLeagueModal();
            this.load();
        } catch (error) {
            utils.showToast('No se pudo crear la liga', 'error');
        }
    },

    deleteLeague(leagueId, leagueName) {
        this.showDeleteConfirmModal(leagueId, leagueName);
    },

    showDeleteConfirmModal(leagueId, leagueName) {
        document.getElementById('deleteConfirmTitle').textContent = `¿Eliminar "${leagueName}"?`;
        document.getElementById('deleteConfirmMessage').textContent =
            'Esta acción no se puede deshacer. Se eliminarán todas las categorías, rondas y partidos asociados.';
        document.getElementById('deleteDetails').innerHTML = `
            <p><strong>Liga:</strong> ${leagueName}</p>
            <p><strong>ID:</strong> ${leagueId}</p>
        `;
        this.pendingDelete = { leagueId, leagueName };
        document.getElementById('deleteConfirmModal').classList.add('active');
    },

    hideDeleteConfirmModal() {
        document.getElementById('deleteConfirmModal').classList.remove('active');
        this.pendingDelete = null;
    },

    async confirmDeleteLeague() {
        if (!this.pendingDelete) return;
        const { leagueId, leagueName } = this.pendingDelete;
        try {
            await api.deleteLeague(leagueId);
            utils.showToast(`Liga "${leagueName}" eliminada correctamente`, 'success');
            this.hideDeleteConfirmModal();
            this.load();
        } catch (error) {
            utils.showToast('No se pudo eliminar la liga', 'error');
        }
    },

    bindActions() {
        document.getElementById('btnCreateClubLeague').onclick = () => this.showCreateLeagueModal();
        document.getElementById('submitCreateLeague').onclick = () => this.submitCreateLeague();
        document.getElementById('cancelCreateLeague').onclick = () => this.hideCreateLeagueModal();
        document.getElementById('closeCreateLeagueModal').onclick = () => this.hideCreateLeagueModal();
        document.getElementById('closeDeleteConfirmModal').onclick = () => this.hideDeleteConfirmModal();
        document.getElementById('cancelDeleteBtn').onclick = () => this.hideDeleteConfirmModal();
        document.getElementById('confirmDeleteBtn').onclick = () => this.confirmDeleteLeague();
        document.getElementById('btnManageClubMembers').onclick = () => router.navigate('clubs');
    }
};
