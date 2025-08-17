import mapManager from './map.js';
import authManager from './auth.js';
import modalManager from './modals.js';
import instructionsManager from './instructions.js';

class App {
  constructor() {
    this.initialized = false;
  }

  async initialize() {
    if (this.initialized) return;

    try {
      // Check authentication status first
      await authManager.checkAuthStatus();
      
      // Initialize map first
      mapManager.initialize();
      
      // Initialize instructions manager
      instructionsManager.initialize();
      
      // Initialize other managers
      // (modalManager is auto-initialized in its constructor)
      
      this.initialized = true;
      console.log('GeoNotes application initialized successfully');
    } catch (error) {
      console.error('Failed to initialize application:', error);
    }
  }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  const app = new App();
  app.initialize();
});

export default App;
