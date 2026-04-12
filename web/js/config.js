// API Configuration
const API_BASE_URL = 'http://localhost:8080';
const TOKEN_KEY = 'padel_token';

// Application State
const state = {
    user: null,
    currentClub: null,
    currentLeague: null,
    currentGroup: null,
    currentMatch: null,
    userType: 'player' // 'player' | 'club'
};
