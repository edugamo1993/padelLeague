// Utility Functions
const utils = {
    formatDate: (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('es-ES', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    },

    showToast: (message, type = 'info') => {
        const toastContainer = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${type} show`;
        toast.textContent = message;
        toastContainer.appendChild(toast);
        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    },

    showErrorModal: (message) => {
        document.getElementById('errorMessage').textContent = message;
        document.getElementById('errorModal').classList.add('active');
    },

    hideErrorModal: () => {
        document.getElementById('errorModal').classList.remove('active');
    },

    formatScore: (sets) => {
        if (!sets || sets.length === 0) return 'Sin jugar';
        return sets.map(set => `${set.gamesPair1}-${set.gamesPair2}`).join(', ');
    },

    getMatchStatus: (match) => {
        if (match.status === 'finished') {
            return { class: 'finished', text: 'Finalizado' };
        }
        return { class: 'pending', text: 'Pendiente' };
    },

    formatMatchResult: (match) => {
        if (match.status !== 'finished') return 'Sin jugar';
        const sets = match.sets || [];
        const pair1Wins = sets.filter(s => s.gamesPair1 > s.gamesPair2).length;
        const pair2Wins = sets.filter(s => s.gamesPair2 > s.gamesPair1).length;
        return `${pair1Wins}-${pair2Wins}`;
    }
};
