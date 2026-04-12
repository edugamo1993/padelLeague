// Application entry point
document.addEventListener('DOMContentLoaded', () => {
    // Hide loader after a brief delay
    setTimeout(() => {
        document.getElementById('loader').classList.add('hidden');
    }, 1000);

    // Initialize modules
    auth.init();
    router.init();
});
