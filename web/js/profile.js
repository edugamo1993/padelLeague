// Profile Module
const profile = {
    async load(forceShowForm = false) {
        if (state.userType === 'club') {
            utils.showToast('El modo club no permite ver perfil de usuario', 'warning');
            router.navigate('clubDashboard');
            return;
        }
        await this.loadProfile();
        await this.checkProfileCompletion(forceShowForm);
    },

    async loadProfile() {
        try {
            const history = await api.getUserHistory(state.user.id);
            document.getElementById('profileName').textContent = state.user.name;
            document.getElementById('profileEmail').textContent = state.user.email;
            document.getElementById('statMatches').textContent = history.matches.length;
            document.getElementById('statWins').textContent = history.matches.filter(m => m.won).length;
            document.getElementById('statLosses').textContent = history.matches.filter(m => !m.won).length;
        } catch (error) {
            console.error('Error loading profile:', error);
        }
    },

    async checkProfileCompletion(forceShowForm = false) {
        try {
            const profileData = await api.getProfile();

            document.getElementById('profileName').textContent = profileData.name;
            document.getElementById('profileEmail').textContent = profileData.email;
            document.getElementById('profileLastName').textContent = profileData.last_name || 'No especificado';
            document.getElementById('profilePhone').textContent = profileData.phone || 'No especificado';
            document.getElementById('profileCity').textContent = profileData.city || 'No especificado';
            document.getElementById('profileLevel').textContent = profileData.padel_level || 'No especificado';

            document.getElementById('profileNameInput').value = profileData.name || '';
            document.getElementById('profileLastNameInput').value = profileData.last_name || '';
            document.getElementById('profilePhoneInput').value = profileData.phone || '';
            document.getElementById('profileBirthDateInput').value = profileData.birth_date || '';
            document.getElementById('profileCityInput').value = profileData.city || '';
            document.getElementById('profileLevelInput').value = profileData.padel_level || '';

            const showForm = forceShowForm || (profileData.is_google_user && !profileData.hasProfile);
            document.getElementById('profileForm').style.display = showForm ? 'block' : 'none';
            document.getElementById('profileComplete').style.display = showForm ? 'none' : 'block';
        } catch (error) {
            console.error('Error checking profile completion:', error);
            document.getElementById('profileForm').style.display = 'block';
            document.getElementById('profileComplete').style.display = 'none';
        }
    },

    async saveProfile() {
        const profileData = {
            name: document.getElementById('profileNameInput').value,
            last_name: document.getElementById('profileLastNameInput').value,
            phone: document.getElementById('profilePhoneInput').value,
            birth_date: document.getElementById('profileBirthDateInput').value,
            city: document.getElementById('profileCityInput').value,
            padel_level: document.getElementById('profileLevelInput').value
        };

        try {
            await api.updateProfile(profileData);
            utils.showToast('Perfil actualizado exitosamente', 'success');
            this.load();
        } catch (error) {
            utils.showToast('Error al actualizar el perfil', 'error');
        }
    }
};
