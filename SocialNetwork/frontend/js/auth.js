import apiService from './api.js';
import { showMessage, createLoadingSpinner } from './utils.js';
import mapManager from './map.js';

class AuthManager {
  constructor() {
    console.log('🔍 AuthManager constructor called'); // Debug log
    this.isLoggedIn = false;
    this.currentUser = null;
    this.setupEventListeners();
  }

  setupEventListeners() {
    console.log('🔍 Setting up event listeners'); // Debug log
    // Auth modal controls
    const profileButton = document.getElementById('profile-button');
    const logoutButton = document.getElementById('logout-button');
    const authModal = document.getElementById('auth-modal');
    const closeModal = document.getElementById('close-modal');

    profileButton.addEventListener('click', () => {
      authModal.classList.add('show');
    });

    closeModal.addEventListener('click', () => {
      authModal.classList.remove('show');
    });

    authModal.addEventListener('click', (e) => {
      if (e.target === authModal) {
        authModal.classList.remove('show');
      }
    });

    // Login and register buttons
    const loginButton = document.getElementById('login-button');
    const registerButton = document.getElementById('register-button');

    loginButton.addEventListener('click', () => {
      console.log('🔍 Login button clicked'); // Debug log
      this.handleLogin();
    });
    registerButton.addEventListener('click', () => this.handleRegister());

    // Logout button
    logoutButton.addEventListener('click', () => this.handleLogout());

    // Handle Enter key in auth form
    document.getElementById('auth-form').addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        console.log('🔍 Enter key pressed, clicking login button'); // Debug log
        loginButton.click();
      }
    });
  }

  async handleLogin() {
    console.log('🔍 handleLogin called'); // Debug log
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!username || !password) {
      showMessage('Please fill in all fields', 'error');
      return;
    }

    const loginButton = document.getElementById('login-button');
    const originalText = loginButton.innerHTML;
    loginButton.innerHTML = createLoadingSpinner().outerHTML;
    loginButton.disabled = true;

    try {
      console.log('🔍 Making login request...'); // Debug log
      const response = await apiService.login(username, password);
      console.log('🔍 Login response received:', response); // Debug log
      await this.loginSuccess(response);
    } catch (error) {
      console.error('Login error:', error);
      showMessage('Login failed. Please check your credentials.', 'error');
    } finally {
      loginButton.innerHTML = originalText;
      loginButton.disabled = false;
    }
  }

  async handleRegister() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!username || !password) {
      showMessage('Please fill in all fields', 'error');
      return;
    }

    const registerButton = document.getElementById('register-button');
    const originalText = registerButton.innerHTML;
    registerButton.innerHTML = createLoadingSpinner().outerHTML;
    registerButton.disabled = true;

    try {
      await apiService.register(username, password);
      showMessage('Registration successful! You can now login.', 'success');
      document.getElementById('username').value = '';
      document.getElementById('password').value = '';
    } catch (error) {
      console.error('Registration error:', error);
      showMessage('Registration failed. Username might already exist.', 'error');
    } finally {
      registerButton.innerHTML = originalText;
      registerButton.disabled = false;
    }
  }

  async handleLogout() {
    try {
      await apiService.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      this.logoutSuccess();
    }
  }

  async loginSuccess(response) {
    this.isLoggedIn = true;
    apiService.setAuthenticated(true);
    
    // Extract user info from login response
    this.currentUser = response.user || response.username || response.name;
    
    showMessage('Login successful!', 'success');
    document.getElementById('auth-modal').classList.remove('show');
    document.getElementById('profile-button').style.display = 'none';
    document.getElementById('logout-button').style.display = 'block';
    
    // Refresh notes to show user's private notes
    mapManager.fetchNotes();
  }

  logoutSuccess() {
    this.isLoggedIn = false;
    apiService.setAuthenticated(false);
    
    document.getElementById('logout-button').style.display = 'none';
    document.getElementById('profile-button').style.display = 'block';
    showMessage('Logged out successfully!', 'success');
    
    // Refresh notes to show only public notes
    mapManager.fetchNotes();
  }

  isUserLoggedIn() {
    return this.isLoggedIn;
  }

  async checkAuthStatus() {
    try {
      const response = await apiService.checkAuthStatus();
      this.isLoggedIn = true;
      apiService.setAuthenticated(true);
      
      // Extract user info from auth check response
      this.currentUser = response.user || response.username || response.name;
      
      // Update UI
      document.getElementById('profile-button').style.display = 'none';
      document.getElementById('logout-button').style.display = 'block';
      
      return true;
    } catch (error) {
      this.isLoggedIn = false;
      this.currentUser = null;
      apiService.setAuthenticated(false);
      
      // Update UI
      document.getElementById('logout-button').style.display = 'none';
      document.getElementById('profile-button').style.display = 'block';
      
      return false;
    }
  }

  async updateCurrentUser() {
    try {
      const userResponse = await apiService.getCurrentUser();
      // The login endpoint returns user info when JWT cookie is valid
      this.currentUser = userResponse.user || userResponse.username || userResponse.name;
      return this.currentUser;
    } catch (error) {
      console.error('Error updating current user:', error);
      this.currentUser = null;
      return null;
    }
  }

  // Method to manually refresh user info if needed
  async refreshUserInfo() {
    return await this.updateCurrentUser();
  }

  getCurrentUser() {
    return this.currentUser;
  }
}

export default new AuthManager();
