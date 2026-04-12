// Leagues Module (includes Groups management)
const leagues = {
    pendingStartRound: null,

    async load() {
        await this.loadLeagues();
        this.bindActions();
    },

    async loadLeagues() {
        try {
            const leaguesList = document.getElementById('leaguesList');
            leaguesList.innerHTML = '';
            const leagues = await api.getLeagues();

            leagues.forEach(league => {
                const leagueName = league.name || league.Name || 'Sin nombre';
                const leagueSeason = league.season || league.Season || '-';
                const leagueId = league.id || league.ID;
                const clubName = league.club?.name || league.club?.Name || 'Desconocido';
                const isActive = league.isActive !== undefined ? league.isActive : (league.IsActive !== undefined ? league.IsActive : true);

                const card = document.createElement('div');
                card.className = 'card league-card';
                card.innerHTML = `
                    <div class="card-header">
                        <div>
                            <h3>${leagueName}</h3>
                            <p class="text-light">Temporada: ${leagueSeason}</p>
                        </div>
                        <div class="league-actions">
                            <button class="btn btn-sm btn-primary" onclick="leagues.viewLeague('${leagueId}')">
                                <i class="fas fa-eye"></i> Ver
                            </button>
                        </div>
                    </div>
                    <div class="card-content">
                        <div class="league-meta">
                            <span><i class="fas fa-trophy"></i> Categorías: ${league.categories?.length || 0}</span>
                            <span><i class="fas fa-users"></i> Club: ${clubName}</span>
                            <span class="status-badge ${isActive ? 'status-active' : 'status-inactive'}">
                                ${isActive ? 'Activa' : 'Inactiva'}
                            </span>
                        </div>
                    </div>
                `;
                leaguesList.appendChild(card);
            });
        } catch (error) {
            console.error('Error loading leagues:', error);
        }
    },

    viewLeague(leagueId) {
        state.currentLeague = leagueId;
        document.getElementById('leagueModal').classList.add('active');
        this.loadLeagueDetails(leagueId);
        this.bindActions();
    },

    async loadLeagueDetails(leagueId) {
        try {
            await this.loadGroups();
            await this.loadStandings(leagueId);
        } catch (error) {
            console.error('Error loading league details:', error);
        }
    },

    async loadStandings(leagueId) {
        try {
            const standings = await api.getLeagueStandings(leagueId);
            document.getElementById('standingsList').innerHTML = `
                <table class="standings-table">
                    <thead>
                        <tr>
                            <th>Posición</th><th>Jugador</th><th>Puntos</th>
                            <th>Ganados</th><th>Perdidos</th><th>Sets a favor</th><th>Sets en contra</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${standings.map((s, i) => `
                            <tr>
                                <td class="position">${i + 1}</td>
                                <td>${s.user.name}</td>
                                <td>${s.points}</td>
                                <td>${s.matchesWon}</td>
                                <td>${s.matchesLost}</td>
                                <td>${s.setsFor}</td>
                                <td>${s.setsAgainst}</td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            `;
        } catch (error) {
            console.error('Error loading standings:', error);
        }
    },

    _memberName(player) {
        if (!player?.groupMember) return 'Desconocido';
        const gm = player.groupMember;
        if (gm.user?.name) return gm.user.name;
        return `${gm.name || ''} ${gm.lastName || ''}`.trim() || 'Desconocido';
    },

    _pairNames(players, pair) {
        return (players || [])
            .filter(p => p.pair === pair)
            .map(p => this._memberName(p))
            .join(' / ') || '-';
    },

    _matchResultText(match) {
        if (match.status !== 'finished') return '-';
        const sets = (match.sets || []).slice().sort((a, b) => a.setNumber - b.setNumber);
        if (sets.length === 0) return `${match.setsPair1}-${match.setsPair2}`;
        return sets.map(s => `${s.gamesPair1}-${s.gamesPair2}`).join('<br>');
    },

    _roundStatusLabel(status) {
        return { pending: 'Pendiente', in_progress: 'En progreso', finished: 'Cerrada' }[status] || status;
    },

    _renderTandaMatches(round, container) {
        container.innerHTML = '';
        const matches = round.matches || [];
        if (matches.length === 0) {
            container.innerHTML = '<p class="text-center text-light">Sin partidos en esta tanda.</p>';
            return;
        }

        // Agrupar por el groupId snapshot guardado en MatchPlayer
        const groupMap = new Map();
        matches.forEach(match => {
            const firstPlayer = (match.players || [])[0];
            const groupId   = firstPlayer?.groupId   || 'sin-grupo';
            const groupName = firstPlayer?.group?.name || 'Sin grupo';
            if (!groupMap.has(groupId)) groupMap.set(groupId, { name: groupName, matches: [] });
            groupMap.get(groupId).matches.push(match);
        });

        groupMap.forEach(({ name, matches: groupMatches }) => {
            const rows = groupMatches.map(match => {
                const p1 = this._pairNames(match.players, 1);
                const p2 = this._pairNames(match.players, 2);
                const result = this._matchResultText(match);
                const isPending = match.status !== 'finished';
                return `<tr>
                    <td>${p1}</td>
                    <td>${p2}</td>
                    <td>${result}</td>
                    <td>${isPending
                        ? `<button class="btn btn-sm btn-success" onclick="leagues.showResultModal('${match.id}')"><i class="fas fa-edit"></i> Resultado</button>`
                        : `<span class="status-badge status-inactive">Terminado</span>`}
                    </td>
                </tr>`;
            }).join('');

            const section = document.createElement('div');
            section.className = 'card';
            section.style.marginBottom = '12px';
            section.innerHTML = `
                <div class="card-header"><h4>${name}</h4></div>
                <div class="card-content">
                    <table class="standings-table">
                        <thead><tr>
                            <th>Pareja 1</th><th>Pareja 2</th><th>Resultado</th><th></th>
                        </tr></thead>
                        <tbody>${rows}</tbody>
                    </table>
                </div>`;
            container.appendChild(section);
        });
    },

    async loadTandas() {
        const tandasList = document.getElementById('tandasList');
        tandasList.innerHTML = '<p class="text-light">Cargando tandas...</p>';
        try {
            const rounds = await api.getLeagueRounds(state.currentLeague);
            tandasList.innerHTML = '';
            if (!rounds || rounds.length === 0) {
                tandasList.innerHTML = '<p class="text-center text-light">No hay tandas todavía. Pulsa "Nueva Tanda" para empezar.</p>';
                return;
            }

            this._rounds = rounds;

            // Desplegable de tandas
            const selectorDiv = document.createElement('div');
            selectorDiv.className = 'form-group';
            selectorDiv.style.marginBottom = '16px';
            const select = document.createElement('select');
            select.id = 'tandaSelector';
            select.className = 'form-control';
            rounds.forEach((round, i) => {
                const opt = document.createElement('option');
                opt.value = i;
                opt.textContent = `Tanda ${round.roundNumber} — ${this._roundStatusLabel(round.status)}`;
                if (i === rounds.length - 1) opt.selected = true;
                select.appendChild(opt);
            });
            const label = document.createElement('label');
            label.innerHTML = '<i class="fas fa-history"></i> Tanda:';
            label.style.marginRight = '8px';
            selectorDiv.appendChild(label);
            selectorDiv.appendChild(select);
            tandasList.appendChild(selectorDiv);

            const matchesContainer = document.createElement('div');
            matchesContainer.id = 'tandaMatchesContainer';
            tandasList.appendChild(matchesContainer);

            this._renderTandaMatches(rounds[rounds.length - 1], matchesContainer);

            select.addEventListener('change', () => {
                this._renderTandaMatches(rounds[parseInt(select.value)], matchesContainer);
            });
        } catch (error) {
            console.error('Error loading tandas:', error);
            tandasList.innerHTML = '<p class="text-error">Error al cargar las tandas.</p>';
        }
    },

    showResultModal(matchId) {
        let match = null;
        for (const round of (this._rounds || [])) {
            match = (round.matches || []).find(m => m.id === matchId);
            if (match) break;
        }
        if (!match) { utils.showToast('Partido no encontrado', 'error'); return; }

        this._pendingResultMatchId = matchId;

        // Rellenar nombres por pareja (uno por línea si son dos jugadores)
        const pair1Players = (match.players || []).filter(p => p.pair === 1);
        const pair2Players = (match.players || []).filter(p => p.pair === 2);
        document.getElementById('resultPair1Names').innerHTML =
            pair1Players.map(p => this._memberName(p)).join('<br>');
        document.getElementById('resultPair2Names').innerHTML =
            pair2Players.map(p => this._memberName(p)).join('<br>');

        // Nombres cortos para las columnas de puntuación (primer nombre)
        const shortName = p => this._memberName(p).split(' ')[0];
        document.getElementById('scoreColP1').textContent =
            pair1Players.map(shortName).join('/') || 'P1';
        document.getElementById('scoreColP2').textContent =
            pair2Players.map(shortName).join('/') || 'P2';

        document.getElementById('resultMatchDate').value = new Date().toISOString().slice(0, 10);
        ['set1p1','set1p2','set2p1','set2p2','set3p1','set3p2'].forEach(id => {
            document.getElementById(id).value = '';
        });
        document.getElementById('resultValidationMsg').style.display = 'none';
        document.getElementById('resultMatchModal').classList.add('active');
    },

    hideResultModal() {
        document.getElementById('resultMatchModal').classList.remove('active');
        this._pendingResultMatchId = null;
    },

    _validateSet(g1, g2) {
        const winner = Math.max(g1, g2);
        const loser  = Math.min(g1, g2);
        if (winner === 6 && loser <= 4) return null;
        if (winner === 7 && (loser === 5 || loser === 6)) return null;
        return `${g1}-${g2} no es un resultado de set válido (válidos: 6-0…6-4, 7-5, 7-6)`;
    },

    _validateSets(sets) {
        for (const s of sets) {
            const err = this._validateSet(s.gamesPair1, s.gamesPair2);
            if (err) return `Set ${s.setNumber}: ${err}`;
        }
        const w1 = sets.filter(s => s.gamesPair1 > s.gamesPair2).length;
        const w2 = sets.filter(s => s.gamesPair2 > s.gamesPair1).length;
        if (w1 === w2) return 'El resultado no puede ser empate';
        if (sets.length === 3) {
            // El 3er set solo se juega si cada pareja ganó 1 de los 2 primeros
            const w1in2 = sets.slice(0,2).filter(s => s.gamesPair1 > s.gamesPair2).length;
            if (w1in2 !== 1) return 'El 3er set solo se juega si los dos primeros los gana cada pareja (1-1)';
        }
        return null;
    },

    async saveResult() {
        const matchId = this._pendingResultMatchId;
        if (!matchId) return;

        const playedAt = document.getElementById('resultMatchDate').value;
        if (!playedAt) { utils.showToast('La fecha es obligatoria', 'error'); return; }

        const setDefs = [
            { n: 1, p1: 'set1p1', p2: 'set1p2' },
            { n: 2, p1: 'set2p1', p2: 'set2p2' },
            { n: 3, p1: 'set3p1', p2: 'set3p2' },
        ];
        const sets = [];
        for (const s of setDefs) {
            const g1 = document.getElementById(s.p1).value;
            const g2 = document.getElementById(s.p2).value;
            if (g1 !== '' && g2 !== '') {
                sets.push({ setNumber: s.n, gamesPair1: parseInt(g1), gamesPair2: parseInt(g2) });
            }
        }

        const validationMsg = document.getElementById('resultValidationMsg');
        if (sets.length < 2) {
            validationMsg.textContent = 'Introduce al menos 2 sets';
            validationMsg.style.display = 'block';
            return;
        }
        const validationErr = this._validateSets(sets);
        if (validationErr) {
            validationMsg.textContent = validationErr;
            validationMsg.style.display = 'block';
            return;
        }
        validationMsg.style.display = 'none';

        try {
            await api.updateMatchResult(matchId, { playedAt, sets });
            utils.showToast('Resultado guardado', 'success');
            this.hideResultModal();
            await this.loadTandas();
        } catch (error) {
            validationMsg.textContent = error.message || 'Error al guardar el resultado';
            validationMsg.style.display = 'block';
        }
    },

    async loadMatches(leagueId) {
        const matchesList = document.getElementById('matchesList');
        matchesList.innerHTML = '<p class="text-light">Cargando partidos...</p>';
        try {
            const rounds = await api.getLeagueRounds(leagueId || state.currentLeague);
            matchesList.innerHTML = '';

            // Agrupa todos los partidos por grupo (snapshot guardado en MatchPlayer)
            const groupMap = new Map(); // groupId → { name, matches[] }
            (rounds || []).forEach(round => {
                (round.matches || []).forEach(match => {
                    const firstPlayer = (match.players || [])[0];
                    const groupId   = firstPlayer?.groupId   || 'sin-grupo';
                    const groupName = firstPlayer?.group?.name || 'Sin grupo';
                    if (!groupMap.has(groupId)) groupMap.set(groupId, { name: groupName, matches: [] });
                    groupMap.get(groupId).matches.push({ round, match });
                });
            });

            if (groupMap.size === 0) {
                matchesList.innerHTML = '<p class="text-center text-light">No hay partidos en esta liga.</p>';
                return;
            }

            groupMap.forEach(({ name, matches }) => {
                const rows = matches.map(({ round, match }) => `
                    <tr>
                        <td>Tanda ${round.roundNumber}</td>
                        <td>${this._pairNames(match.players, 1)}</td>
                        <td>${this._pairNames(match.players, 2)}</td>
                        <td>${this._matchResultText(match)}</td>
                        <td><span class="status-badge ${match.status === 'finished' ? 'status-inactive' : 'status-active'}">${match.status === 'finished' ? 'Terminado' : 'Pendiente'}</span></td>
                    </tr>
                `).join('');
                const section = document.createElement('div');
                section.className = 'card';
                section.style.marginBottom = '16px';
                section.innerHTML = `
                    <div class="card-header"><h4>${name}</h4></div>
                    <div class="card-content">
                        <table class="standings-table">
                            <thead><tr>
                                <th>Tanda</th><th>Pareja 1</th><th>Pareja 2</th><th>Resultado</th><th>Estado</th>
                            </tr></thead>
                            <tbody>${rows}</tbody>
                        </table>
                    </div>
                `;
                matchesList.appendChild(section);
            });
        } catch (error) {
            console.error('Error loading matches:', error);
            matchesList.innerHTML = '<p class="text-light text-error">Error al cargar los partidos.</p>';
        }
    },

    async startTanda() {
        if (!state.currentLeague) { utils.showErrorModal('No hay liga seleccionada'); return; }
        try {
            await api.request(`/leagues/${state.currentLeague}/rounds/start`, { method: 'POST' });
            utils.showToast('Tanda iniciada exitosamente', 'success');
            await this.loadTandas();
        } catch (error) {
            const msg = error.message || '';
            if (msg.includes('tanda anterior no está cerrada') || msg.includes('no se puede iniciar')) {
                // Ofrecer finalizar la tanda actual
                document.getElementById('finishRoundMovements').innerHTML = '';
                document.getElementById('finishRoundResult').style.display = 'none';
                document.getElementById('finishRoundActions').style.display = 'flex';
                document.getElementById('finishRoundConfirmBtn').style.display = 'inline-flex';
                document.getElementById('finishRoundCloseBtn').style.display = 'none';
                document.getElementById('finishRoundModal').classList.add('active');
            } else {
                utils.showErrorModal(msg || 'Error al iniciar la tanda');
            }
        }
    },

    hideFinishRoundModal() {
        document.getElementById('finishRoundModal').classList.remove('active');
    },

    async confirmFinishRound() {
        const actionsDiv = document.getElementById('finishRoundActions');
        const resultDiv  = document.getElementById('finishRoundResult');
        actionsDiv.style.display = 'none';
        resultDiv.innerHTML = '<p class="text-light">Finalizando tanda y recalculando grupos...</p>';
        resultDiv.style.display = 'block';

        try {
            const res = await api.request(`/leagues/${state.currentLeague}/rounds/finish`, { method: 'POST' });
            const movements = res.movements || [];

            const ups   = movements.filter(m => m.direction === 'up');
            const downs = movements.filter(m => m.direction === 'down');
            let html = '';
            if (ups.length) {
                html += `<p><strong>Suben de grupo:</strong></p><ul>` +
                    ups.map(m => `<li>${m.memberName}: ${m.fromGroup} → ${m.toGroup}</li>`).join('') + `</ul>`;
            }
            if (downs.length) {
                html += `<p><strong>Bajan de grupo:</strong></p><ul>` +
                    downs.map(m => `<li>${m.memberName}: ${m.fromGroup} → ${m.toGroup}</li>`).join('') + `</ul>`;
            }
            if (!ups.length && !downs.length) {
                html = '<p>Todos los jugadores se mantienen en sus grupos.</p>';
            }
            document.getElementById('finishRoundMovements').innerHTML = html;

            // Intentar iniciar nueva tanda
            let startMsg = '';
            try {
                await api.request(`/leagues/${state.currentLeague}/rounds/start`, { method: 'POST' });
                startMsg = '<p style="color:var(--success-color, green)">Nueva tanda iniciada correctamente.</p>';
            } catch (startErr) {
                startMsg = `<p style="color:var(--warning-color, orange)">${startErr.message}</p>`;
            }
            resultDiv.innerHTML = startMsg;

            document.getElementById('finishRoundConfirmBtn').style.display = 'none';
            document.getElementById('finishRoundCloseBtn').style.display = 'inline-flex';

            await this.loadTandas();
        } catch (error) {
            resultDiv.innerHTML = `<p style="color:red">${error.message || 'Error al finalizar la tanda'}</p>`;
            actionsDiv.style.display = 'flex';
        }
    },


    // ── Groups ────────────────────────────────────────────────────────────────

    async loadGroups() {
        if (!state.currentLeague) { utils.showToast('No hay liga seleccionada', 'error'); return; }
        try {
            const groups = await api.getLeagueGroups(state.currentLeague);
            const groupsList = document.getElementById('groupsList');
            groupsList.innerHTML = '';

            if (!groups || groups.length === 0) {
                groupsList.innerHTML = '<p class="text-center text-light">No hay grupos en esta liga.</p>';
                return;
            }

            groups.forEach(group => {
                const groupName = group.name || group.Name || 'Sin nombre';
                const minPlayers = group.minPlayers || group.MinPlayers || 4;
                const groupId = group.id || group.ID;

                const card = document.createElement('div');
                card.className = 'card group-card';
                card.innerHTML = `
                    <div class="card-header">
                        <div>
                            <h4>${groupName}</h4>
                            <p>Mínimo ${minPlayers} jugadores</p>
                        </div>
                        <div class="card-actions">
                            <button class="btn btn-sm btn-success" onclick="leagues.showAddMemberModal('${groupId}')" title="Añadir miembro">
                                <i class="fas fa-plus"></i>
                            </button>
                            <button class="btn btn-sm btn-primary" onclick="leagues.viewGroup('${groupId}')">
                                <i class="fas fa-eye"></i> Ver
                            </button>
                            <button class="btn btn-sm btn-secondary" onclick="leagues.editGroup('${groupId}', '${groupName}', ${minPlayers})">
                                <i class="fas fa-edit"></i> Editar
                            </button>
                            <button class="btn btn-sm btn-danger" onclick="leagues.deleteGroup('${groupId}', '${groupName}')">
                                <i class="fas fa-trash"></i> Borrar
                            </button>
                        </div>
                    </div>
                    <div class="card-body">
                        <div class="group-stats">
                            <span class="stat-item"><i class="fas fa-users"></i> ${group.members?.length || 0} miembros</span>
                            <span class="stat-item ${group.members?.length >= minPlayers ? 'success' : 'warning'}">
                                <i class="fas fa-check-circle"></i> ${group.members?.length >= minPlayers ? 'Completo' : 'Incompleto'}
                            </span>
                        </div>
                    </div>
                `;
                groupsList.appendChild(card);
            });
        } catch (error) {
            console.error('Error loading groups:', error);
            utils.showToast('Error al cargar grupos', 'error');
        }
    },

    showCreateGroupModal() {
        document.getElementById('groupModalTitle').textContent = 'Crear Grupo';
        document.getElementById('groupName').value = '';
        document.getElementById('groupMinPlayers').value = '4';
        document.getElementById('groupId').value = '';
        document.getElementById('groupModal').classList.add('active');
    },

    editGroup(groupId, currentName, currentMinPlayers) {
        document.getElementById('groupModalTitle').textContent = 'Editar Grupo';
        document.getElementById('groupName').value = currentName;
        document.getElementById('groupMinPlayers').value = currentMinPlayers;
        document.getElementById('groupId').value = groupId;
        document.getElementById('groupModal').classList.add('active');
    },

    hideGroupModal() {
        document.getElementById('groupModal').classList.remove('active');
    },

    async submitGroup() {
        const groupId = document.getElementById('groupId').value;
        const name = document.getElementById('groupName').value.trim();
        const minPlayers = parseInt(document.getElementById('groupMinPlayers').value);

        if (!name) { utils.showToast('El nombre del grupo es obligatorio', 'error'); return; }
        if (minPlayers < 4) { utils.showToast('El mínimo de jugadores debe ser al menos 4', 'error'); return; }

        try {
            if (groupId) {
                await api.updateGroup(groupId, name, minPlayers);
                utils.showToast('Grupo actualizado correctamente', 'success');
            } else {
                await api.createGroup(state.currentLeague, name, minPlayers);
                utils.showToast('Grupo creado correctamente', 'success');
            }
            this.hideGroupModal();
            this.loadGroups();
        } catch (error) {
            utils.showToast('Error al guardar el grupo', 'error');
        }
    },

    deleteGroup(groupId, groupName) {
        if (confirm(`¿Estás seguro de que quieres borrar el grupo "${groupName}"?`)) {
            this.confirmDeleteGroup(groupId, groupName);
        }
    },

    async confirmDeleteGroup(groupId, groupName) {
        try {
            await api.deleteGroup(groupId);
            utils.showToast(`Grupo "${groupName}" eliminado correctamente`, 'success');
            this.loadGroups();
        } catch (error) {
            utils.showToast('Error al eliminar el grupo', 'error');
        }
    },

    async viewGroup(groupId) {
        state.currentGroup = groupId;
        try {
            const members = await api.getGroupMembers(groupId);
            const groupMembersList = document.getElementById('groupMembersList');
            groupMembersList.innerHTML = '';

            if (!members || members.length === 0) {
                groupMembersList.innerHTML = '<p class="text-center text-light">No hay miembros en este grupo.</p>';
            } else {
                members.forEach(member => {
                    const item = document.createElement('div');
                    item.className = 'list-item';
                    item.innerHTML = `
                        <div>
                            <h4>${member.user ? member.user.name : member.name + ' ' + member.lastName}</h4>
                            <p>${member.user ? member.user.email : member.phone}</p>
                            <small class="text-light">${member.user ? 'Usuario registrado' : 'Usuario no registrado'}</small>
                        </div>
                        <button class="btn btn-sm btn-danger" onclick="leagues.removeMember('${groupId}', '${member.id}', '${member.user ? member.user.name : member.name + ' ' + member.lastName}')">
                            <i class="fas fa-times"></i>
                        </button>
                    `;
                    groupMembersList.appendChild(item);
                });
            }
            document.getElementById('groupDetailsModal').classList.add('active');
        } catch (error) {
            utils.showToast('Error al cargar miembros del grupo', 'error');
        }
    },

    hideGroupDetailsModal() {
        document.getElementById('groupDetailsModal').classList.remove('active');
    },

    showAddMemberModal(groupId) {
        state.currentGroup = groupId;
        document.getElementById('memberSearchResults').innerHTML = '';
        document.getElementById('memberPhone').value = '';
        document.getElementById('memberName').value = '';
        document.getElementById('memberLastName').value = '';
        document.getElementById('memberPhone2').value = '';
        document.getElementById('selectedUserId').value = '';
        document.getElementById('registeredUserSection').style.display = 'none';
        document.getElementById('unregisteredUserSection').style.display = 'block';
        document.getElementById('addMemberModal').classList.add('active');
    },

    hideAddMemberModal() {
        document.getElementById('addMemberModal').classList.remove('active');
    },

    async searchUsers() {
        const phone = document.getElementById('memberPhone').value.trim();
        const resultsDiv = document.getElementById('memberSearchResults');

        if (!phone) { resultsDiv.innerHTML = '<p class="text-light">Introduce un teléfono para buscar usuarios registrados</p>'; return; }
        if (!/^\d{9,}$/.test(phone)) { resultsDiv.innerHTML = '<p class="text-error">El teléfono debe tener al menos 9 dígitos</p>'; return; }

        resultsDiv.innerHTML = '<p class="text-light">Buscando usuarios...</p>';
        try {
            const users = await api.searchUsersByPhone(phone);
            if (!users || users.length === 0) {
                resultsDiv.innerHTML = '<p class="text-light">No se encontraron usuarios registrados con ese teléfono.</p>';
                document.getElementById('unregisteredUserSection').style.display = 'block';
                return;
            }
            resultsDiv.innerHTML = users.map(user => `
                <div class="user-result" onclick="leagues.selectUser('${user.id}', '${user.name}', '${user.email || ''}')">
                    <div>
                        <h4>${user.name} ${user.lastName || ''}</h4>
                        <p>${user.email || 'Sin email'}</p>
                        <small class="text-light">${user.phone || 'Sin teléfono'}</small>
                    </div>
                    <button class="btn btn-sm btn-success">Seleccionar</button>
                </div>
            `).join('');
            document.getElementById('unregisteredUserSection').style.display = 'none';
        } catch (error) {
            resultsDiv.innerHTML = '<p class="text-error">Error al buscar usuarios. Inténtalo de nuevo.</p>';
        }
    },

    selectUser(userId, name, email) {
        document.getElementById('selectedUserId').value = userId;
        document.getElementById('selectedUserInfo').innerHTML = `
            <div class="selected-user-info">
                <h4>${name}</h4>
                <p class="text-light">${email || 'Sin email'}</p>
                <button class="btn btn-sm btn-outline" onclick="leagues.clearUserSelection()">Cambiar selección</button>
            </div>
        `;
        document.getElementById('registeredUserSection').style.display = 'block';
        document.getElementById('unregisteredUserSection').style.display = 'none';
        document.getElementById('memberSearchResults').innerHTML = '';
        document.getElementById('memberPhone').value = '';
    },

    clearUserSelection() {
        document.getElementById('selectedUserId').value = '';
        document.getElementById('selectedUserInfo').innerHTML = '';
        document.getElementById('registeredUserSection').style.display = 'none';
        document.getElementById('unregisteredUserSection').style.display = 'block';
        document.getElementById('memberSearchResults').innerHTML = '<p class="text-light">Introduce un teléfono para buscar usuarios registrados</p>';
    },

    async addMember() {
        const selectedUserId = document.getElementById('selectedUserId').value;
        const name = document.getElementById('memberName').value.trim();
        const lastName = document.getElementById('memberLastName').value.trim();
        const phone = document.getElementById('memberPhone2').value.trim();

        let memberData;
        if (selectedUserId) {
            memberData = { userId: selectedUserId };
        } else {
            if (!name) { utils.showToast('El nombre es obligatorio', 'error'); return; }
            if (!lastName) { utils.showToast('Los apellidos son obligatorios', 'error'); return; }
            if (!phone) { utils.showToast('El teléfono es obligatorio', 'error'); return; }
            if (!/^\d{9,}$/.test(phone)) { utils.showToast('El teléfono debe tener al menos 9 dígitos', 'error'); return; }
            memberData = { name, lastName, phone };
        }

        try {
            await api.addGroupMember(state.currentGroup, memberData);
            utils.showToast('Miembro añadido correctamente', 'success');
            this.hideAddMemberModal();
            this.viewGroup(state.currentGroup);
        } catch (error) {
            utils.showToast('Error al añadir miembro', 'error');
        }
    },

    async removeMember(groupId, memberId, memberName) {
        if (confirm(`¿Estás seguro de que quieres remover a "${memberName}" del grupo?`)) {
            try {
                await api.removeGroupMember(groupId, memberId);
                utils.showToast('Miembro removido correctamente', 'success');
                this.viewGroup(groupId);
            } catch (error) {
                utils.showToast('Error al remover miembro', 'error');
            }
        }
    },

    async loadStandings() {
        try {
            await api.getLeagueStandings(state.currentLeague);
            document.getElementById('standingsList').innerHTML = '<p class="text-center text-light">Funcionalidad de clasificaciones próximamente.</p>';
        } catch (error) {
            console.error('Error loading standings:', error);
        }
    },

    bindActions() {
        // Tab handling
        const tabButtons = document.querySelectorAll('#leagueModal .tab-btn');
        tabButtons.forEach(button => {
            button.addEventListener('click', () => {
                tabButtons.forEach(btn => btn.classList.remove('active'));
                button.classList.add('active');
                document.querySelectorAll('#leagueModal .tab-pane').forEach(pane => pane.classList.remove('active'));
                const tabName = button.dataset.tab;
                document.getElementById(tabName + 'Tab').classList.add('active');

                if (tabName === 'groups') this.loadGroups();
                else if (tabName === 'tandas') this.loadTandas();
                else if (tabName === 'standings') this.loadStandings();
                else if (tabName === 'matches') this.loadMatches(state.currentLeague);
            });
        });

        document.getElementById('closeGroupModal').onclick = () => this.hideGroupModal();
        document.getElementById('memberPhone').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.searchUsers();
        });
    }
};
