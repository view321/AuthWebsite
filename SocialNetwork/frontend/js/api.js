import CONFIG from './config.js';
import { showMessage } from './utils.js';

class ApiService {
  constructor() {
    this.isAuthenticated = false;
  }

  setAuthenticated(status) {
    this.isAuthenticated = status;
  }

  getHeaders() {
    return {
      'Content-Type': 'application/json'
    };
  }

  async request(endpoint, options = {}) {
    console.log('🔍 API request to:', endpoint, 'method:', options.method || 'GET'); // Debug log
    try {
      const response = await fetch(CONFIG.API.BASE_URL + endpoint, {
        ...options,
        headers: this.getHeaders(),
        credentials: "include" // Include cookies in requests
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }

  // Authentication methods
  async login(username, password) {
    console.log('🔍 API login called with username:', username); // Debug log
    return this.request(CONFIG.API.ENDPOINTS.LOGIN, {
      method: 'POST',
      body: JSON.stringify({ username, password })
    });
  }

  async sendEmailVerification(email) {
    return this.request('/api/send_email_key', {
      method: 'POST',
      body: JSON.stringify({ email })
    });
  }

  async register(username, password, email, emailKey) {
    return this.request(CONFIG.API.ENDPOINTS.REGISTER, {
      method: 'POST',
      body: JSON.stringify({ username, password, email, email_key: emailKey })
    });
  }

  async logout() {
    return this.request(CONFIG.API.ENDPOINTS.LOGOUT, {
      method: 'POST'
    });
  }

  // Note methods
  async createNote(noteData) {
    return this.request(CONFIG.API.ENDPOINTS.CREATE_NOTE, {
      method: 'POST',
      body: JSON.stringify(noteData)
    });
  }

  async deleteNote(noteId) {
    return this.request(`${CONFIG.API.ENDPOINTS.DELETE_NOTE}/${noteId}/delete`, {
      method: 'DELETE'
    });
  }

  async getNotes(bounds) {
    return this.request(CONFIG.API.ENDPOINTS.GET_NOTES, {
      method: 'POST',
      body: JSON.stringify({
        upper_lattitude: bounds.getNorth(),
        lower_lattitude: bounds.getSouth(),
        upper_longitude: bounds.getEast(),
        lower_longitude: bounds.getWest()
      })
    });
  }

  // Check authentication status by making a login request with existing cookie
    async checkAuthStatus() {
    try {
      // This endpoint will only succeed if the JWT cookie is valid.
      const response = await this.request('/api/login_protected/check_session', {
        method: 'GET'
      });
      this.setAuthenticated(true);
      return response;
    } catch (error) {
      this.setAuthenticated(false);
      // It's expected to fail if not logged in, so we don't re-throw the error.
      return null;
    }
  }

  // Get current user info from the session check endpoint
  async getCurrentUser() {
    // This can now be an alias for checkAuthStatus
    return this.checkAuthStatus();
  }
}

export default new ApiService();
