// API Service
const api = {
    async request(endpoint, options = {}) {
        const url = `${API_BASE_URL}${endpoint}`;
        const token = localStorage.getItem(TOKEN_KEY);

        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...(token && { 'Authorization': `Bearer ${token}` })
            },
            ...options
        };

        try {
            const response = await fetch(url, config);

            if (!response.ok) {
                if (response.status === 401) {
                    localStorage.removeItem(TOKEN_KEY);
                    window.location.reload();
                }
                let errorMessage = `HTTP error! status: ${response.status}`;
                try {
                    const errorData = await response.json();
                    if (errorData.error) errorMessage = errorData.error;
                } catch (e) { /* usa mensaje por defecto */ }
                throw new Error(errorMessage);
            }

            return await response.json();
        } catch (error) {
            console.error('API Error:', error);
            throw error;
        }
    },

    // Auth
    async login(email, password, role = 'player') {
        return this.request('/auth/login', { method: 'POST', body: JSON.stringify({ email, password, role }) });
    },
    async register(name, email, password, role = 'player') {
        return this.request('/auth/register', { method: 'POST', body: JSON.stringify({ name, email, password, role }) });
    },

    // Profile
    async getProfile() { return this.request('/profile'); },
    async updateProfile(profileData) {
        return this.request('/profile', { method: 'PUT', body: JSON.stringify(profileData) });
    },

    // Users
    async getUserHistory(userId) { return this.request(`/users/${userId}/history`); },
    async searchUsersByPhone(phone) { return this.request(`/users/search?phone=${encodeURIComponent(phone)}`); },

    // Clubs
    async getClubs() { return this.request('/clubs'); },
    async getUserClubs() { return this.request('/user/clubs'); },
    async createClub(name, location) {
        return this.request('/clubs', { method: 'POST', body: JSON.stringify({ name, location }) });
    },
    async joinClub(clubId) {
        return this.request(`/clubs/${clubId}/join`, { method: 'POST' });
    },

    // Leagues
    async getLeagues() { return this.request('/leagues'); },
    async createLeague(clubId, name, season) {
        return this.request('/leagues', { method: 'POST', body: JSON.stringify({ clubId, name, season }) });
    },
    async deleteLeague(leagueId) {
        return this.request(`/leagues/${leagueId}`, { method: 'DELETE' });
    },
    async getLeagueCategories(leagueId) { return this.request(`/leagues/${leagueId}/categories`); },
    async getLeagueStandings(leagueId) { return this.request(`/leagues/${leagueId}/standings`); },
    async getLeagueRounds(leagueId) { return this.request(`/leagues/${leagueId}/rounds`); },

    // Groups
    async getLeagueGroups(leagueId) { return this.request(`/leagues/${leagueId}/groups`); },
    async createGroup(leagueId, name, minPlayers) {
        return this.request(`/leagues/${leagueId}/groups`, { method: 'POST', body: JSON.stringify({ name, minPlayers }) });
    },
    async updateGroup(groupId, name, minPlayers) {
        return this.request(`/groups/${groupId}`, { method: 'PUT', body: JSON.stringify({ name, minPlayers }) });
    },
    async deleteGroup(groupId) { return this.request(`/groups/${groupId}`, { method: 'DELETE' }); },
    async getGroupMembers(groupId) { return this.request(`/groups/${groupId}/members`); },
    async addGroupMember(groupId, memberData) {
        return this.request(`/groups/${groupId}/members`, { method: 'POST', body: JSON.stringify(memberData) });
    },
    async removeGroupMember(groupId, memberId) {
        return this.request(`/groups/${groupId}/members/${memberId}`, { method: 'DELETE' });
    },

    // Matches
    async getMatches() { return this.request('/matches'); },
    async getMatch(matchId) { return this.request(`/matches/${matchId}`); },
    async updateMatchResult(matchId, result) {
        return this.request(`/matches/${matchId}/result`, { method: 'PUT', body: JSON.stringify(result) });
    }
};
