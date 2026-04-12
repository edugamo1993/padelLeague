// Matches Module
const matches = {
    async load() {
        await this.loadUserMatches();
    },

    async loadUserMatches() {
        try {
            const history = await api.getUserHistory(state.user.id);
            const userMatchesList = document.getElementById('userMatchesList');
            userMatchesList.innerHTML = '';

            if (history.matches.length === 0) {
                userMatchesList.innerHTML = '<p class="text-center text-light">No has participado en partidos.</p>';
                return;
            }

            history.matches.forEach(match => {
                const status = utils.getMatchStatus(match);
                const card = document.createElement('div');
                card.className = 'match-card';
                card.innerHTML = `
                    <div class="match-header">
                        <div>
                            <h4>Partido ${match.id}</h4>
                            <p>${utils.formatDate(match.createdAt)}</p>
                            <span class="match-status ${status.class}">${status.text}</span>
                        </div>
                        <div class="match-actions">
                            <button class="btn btn-sm btn-primary" onclick="matches.viewMatch('${match.id}')">
                                <i class="fas fa-eye"></i> Ver
                            </button>
                            ${match.status === 'pending' ? `
                                <button class="btn btn-sm btn-success" onclick="matches.recordResult('${match.id}')">
                                    <i class="fas fa-edit"></i> Resultado
                                </button>` : ''}
                        </div>
                    </div>
                    <div class="match-players">
                        <div class="player-pair">
                            <h4>Pareja 1</h4>
                            <div class="player-list">
                                ${match.players?.filter(p => p.pair === 1).map(p => `<div class="player-item">${p.user.name}</div>`).join('') || '<div class="player-item">Sin jugadores</div>'}
                            </div>
                        </div>
                        <div class="player-pair">
                            <h4>Pareja 2</h4>
                            <div class="player-list">
                                ${match.players?.filter(p => p.pair === 2).map(p => `<div class="player-item">${p.user.name}</div>`).join('') || '<div class="player-item">Sin jugadores</div>'}
                            </div>
                        </div>
                    </div>
                    <div class="match-actions">
                        <span class="text-light">Resultado: ${utils.formatMatchResult(match)}</span>
                    </div>
                `;
                userMatchesList.appendChild(card);
            });
        } catch (error) {
            console.error('Error loading user matches:', error);
        }
    },

    viewMatch(matchId) {
        state.currentMatch = matchId;
        document.getElementById('matchModal').classList.add('active');
        this.loadMatchDetails(matchId);
    },

    async loadMatchDetails(matchId) {
        try {
            const match = await api.getMatch(matchId);
            const status = utils.getMatchStatus(match);
            document.getElementById('matchDetails').innerHTML = `
                <div class="match-header">
                    <div>
                        <h3>Partido ${match.id}</h3>
                        <p>${utils.formatDate(match.createdAt)}</p>
                        <span class="match-status ${status.class}">${status.text}</span>
                    </div>
                </div>
                <div class="match-players">
                    <div class="player-pair">
                        <h4>Pareja 1</h4>
                        <div class="player-list">
                            ${match.players?.filter(p => p.pair === 1).map(p => `<div class="player-item">${p.user.name}</div>`).join('') || '<div class="player-item">Sin jugadores</div>'}
                        </div>
                    </div>
                    <div class="player-pair">
                        <h4>Pareja 2</h4>
                        <div class="player-list">
                            ${match.players?.filter(p => p.pair === 2).map(p => `<div class="player-item">${p.user.name}</div>`).join('') || '<div class="player-item">Sin jugadores</div>'}
                        </div>
                    </div>
                </div>
                ${match.sets && match.sets.length > 0 ? `
                    <div class="mt-20">
                        <h4>Sets</h4>
                        <div class="list-container">
                            ${match.sets.map(set => `
                                <div class="list-item">
                                    <div>
                                        <h4>Set ${set.setNumber}</h4>
                                        <p>${set.gamesPair1} - ${set.gamesPair2}</p>
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                ` : ''}
            `;
        } catch (error) {
            console.error('Error loading match details:', error);
        }
    },

    recordResult(matchId) {
        state.currentMatch = matchId;
        document.getElementById('matchModal').classList.add('active');
        this.loadMatchDetails(matchId);
        document.getElementById('matchResultForm').style.display = 'block';
    },

    async submitResult() {
        const matchId = state.currentMatch;
        const set1Pair1 = parseInt(document.getElementById('set1Pair1').value) || 0;
        const set1Pair2 = parseInt(document.getElementById('set1Pair2').value) || 0;
        const set2Pair1 = parseInt(document.getElementById('set2Pair1').value) || 0;
        const set2Pair2 = parseInt(document.getElementById('set2Pair2').value) || 0;
        const set3Pair1 = parseInt(document.getElementById('set3Pair1').value) || 0;
        const set3Pair2 = parseInt(document.getElementById('set3Pair2').value) || 0;

        const result = {
            sets: [
                { setNumber: 1, gamesPair1: set1Pair1, gamesPair2: set1Pair2 },
                { setNumber: 2, gamesPair1: set2Pair1, gamesPair2: set2Pair2 }
            ]
        };
        if (set3Pair1 > 0 || set3Pair2 > 0) {
            result.sets.push({ setNumber: 3, gamesPair1: set3Pair1, gamesPair2: set3Pair2 });
        }

        try {
            await api.updateMatchResult(matchId, result);
            utils.showToast('Resultado registrado exitosamente', 'success');
            document.getElementById('matchResultForm').style.display = 'none';
            this.loadMatchDetails(matchId);
        } catch (error) {
            utils.showToast('Error al registrar el resultado', 'error');
        }
    }
};
