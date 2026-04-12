// Clubs Module
const clubs = {
    async load() {
        await this.loadClubs();
    },

    async loadClubs() {
        try {
            const userClubs = await api.getUserClubs();
            const clubsGrid = document.getElementById('clubsGrid');
            clubsGrid.innerHTML = '';

            if (userClubs.length === 0) {
                clubsGrid.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-building"></i>
                        <h3>No estás en ningún club</h3>
                        <p>Para ver clubs y sus ligas, debes estar registrado en una liga de algún club.</p>
                    </div>
                `;
                return;
            }

            userClubs.forEach(club => {
                const card = document.createElement('div');
                card.className = 'card league-card';
                card.innerHTML = `
                    <div class="card-header">
                        <div>
                            <h3>${club.club_name}</h3>
                            <p class="text-light">${club.location || 'Sin ubicación'}</p>
                        </div>
                        <div class="club-actions">
                            <button class="btn btn-sm btn-primary" onclick="clubs.viewClubDetails('${club.club_id}')">
                                <i class="fas fa-eye"></i> Ver Liga
                            </button>
                        </div>
                    </div>
                    <div class="card-content">
                        <div class="league-meta">
                            <span class="status-badge ${club.role === 'admin' ? 'status-active' : 'status-inactive'}">
                                ${club.role === 'admin' ? 'Administrador' : 'Jugador'}
                            </span>
                        </div>
                    </div>
                `;
                clubsGrid.appendChild(card);
            });
        } catch (error) {
            console.error('Error loading user clubs:', error);
            document.getElementById('clubsGrid').innerHTML = `
                <div class="error-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <h3>Error al cargar clubs</h3>
                    <p>No se pudieron cargar los clubs a los que perteneces.</p>
                </div>
            `;
        }
    },

    viewClubDetails(clubId) {
        state.currentClub = clubId;
        document.getElementById('clubModal').classList.add('active');
        this.loadClubDetails(clubId);
    },

    async loadClubDetails(clubId) {
        try {
            const club = await api.request(`/clubs/${clubId}`);
            document.getElementById('clubModalTitle').textContent = club.name;
            document.getElementById('clubDetails').innerHTML = `
                <div class="league-meta">
                    <span><i class="fas fa-map-marker-alt"></i> ${club.location || 'Sin ubicación'}</span>
                    <span><i class="fas fa-calendar"></i> Creado: ${utils.formatDate(club.createdAt)}</span>
                </div>
                <div class="mt-20">
                    <h4>Miembros</h4>
                    <div class="list-container">
                        ${club.userClubs?.map(uc => `
                            <div class="list-item">
                                <div>
                                    <h4>${uc.user.name}</h4>
                                    <p>${uc.user.email}</p>
                                </div>
                                <span class="status-badge">${uc.role}</span>
                            </div>
                        `).join('') || '<p class="text-light">No hay miembros</p>'}
                    </div>
                </div>
            `;
        } catch (error) {
            console.error('Error loading club details:', error);
        }
    },

    async createLeague() {
        const name = prompt('Nombre de la liga:');
        const season = prompt('Temporada (ej: 2025-Primavera):');
        if (name && season) {
            try {
                await api.createLeague(state.currentClub, name, season);
                utils.showToast('Liga creada exitosamente', 'success');
                this.loadClubDetails(state.currentClub);
            } catch (error) {
                utils.showToast('Error al crear la liga', 'error');
            }
        }
    }
};
