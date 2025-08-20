import apiService from './api.js';
import { showMessage, createLoadingSpinner } from './utils.js';
import mapManager from './map.js';

class AuthManager {
  constructor() {
    console.log('🔍 AuthManager constructor called'); // Debug log
    this.isLoggedIn = false;
    this.currentUser = null;
    this.registrationStep = 'login'; // 'login', 'register', 'verify'
    this.pendingRegistration = null;
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
    const sendVerificationButton = document.getElementById('send-verification-button');
    const verifyRegisterButton = document.getElementById('verify-register-button');
    const backToLoginButton = document.getElementById('back-to-login-button');

    loginButton.addEventListener('click', () => {
      console.log('🔍 Login button clicked'); // Debug log
      this.handleLogin();
    });
    registerButton.addEventListener('click', () => this.showRegistrationForm());
    sendVerificationButton.addEventListener('click', () => this.handleSendVerification());
    verifyRegisterButton.addEventListener('click', () => this.handleVerifyAndRegister());
    backToLoginButton.addEventListener('click', () => this.showLoginForm());

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

  showRegistrationForm() {
    this.registrationStep = 'register';
    
    // Show email field and hide login-specific elements
    document.getElementById('email-group').style.display = 'block';
    document.getElementById('login-button').style.display = 'none';
    document.getElementById('register-button').style.display = 'none';
    document.getElementById('send-verification-button').style.display = 'block';
    document.getElementById('back-to-login-button').style.display = 'block';
    
    // Update header text
    document.querySelector('.auth-header h2').textContent = 'Create Account';
    document.querySelector('.auth-header p').textContent = 'Enter your details to create a new account';
  }

  showLoginForm() {
    this.registrationStep = 'login';
    
    // Hide registration-specific elements
    document.getElementById('email-group').style.display = 'none';
    document.getElementById('verification-group').style.display = 'none';
    document.getElementById('send-verification-button').style.display = 'none';
    document.getElementById('verify-register-button').style.display = 'none';
    document.getElementById('back-to-login-button').style.display = 'none';
    
    // Show login elements
    document.getElementById('login-button').style.display = 'block';
    document.getElementById('register-button').style.display = 'block';
    
    // Reset header text
    document.querySelector('.auth-header h2').textContent = 'Welcome to GeoNotes';
    document.querySelector('.auth-header p').textContent = 'Sign in to create and manage your location-based notes';
    
    // Clear form
    this.clearForm();
  }

  showVerificationForm() {
    this.registrationStep = 'verify';
    
    // Show verification field and hide send verification button
    document.getElementById('verification-group').style.display = 'block';
    document.getElementById('send-verification-button').style.display = 'none';
    document.getElementById('verify-register-button').style.display = 'block';
    
    // Update header text
    document.querySelector('.auth-header h2').textContent = 'Verify Email';
    document.querySelector('.auth-header p').textContent = 'Enter the verification code sent to your email';
    
    // Focus on verification code input
    setTimeout(() => {
      document.getElementById('verification-code').focus();
    }, 100);
  }

  async handleSendVerification() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const email = document.getElementById('email').value;

    if (!username || !password || !email) {
      showMessage('Please fill in all fields', 'error');
      return;
    }

    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      showMessage('Please enter a valid email address', 'error');
      return;
    }

    const sendButton = document.getElementById('send-verification-button');
    const originalText = sendButton.innerHTML;
    sendButton.innerHTML = createLoadingSpinner().outerHTML;
    sendButton.disabled = true;

    try {
      await apiService.sendEmailVerification(email);
      
      // Store registration data for later use
      this.pendingRegistration = { username, password, email };
      
      showMessage('Verification code sent to your email!', 'success');
      this.showVerificationForm();
    } catch (error) {
      console.error('Email verification error:', error);
      showMessage('Failed to send verification code. Please try again.', 'error');
    } finally {
      sendButton.innerHTML = originalText;
      sendButton.disabled = false;
    }
  }

  async handleVerifyAndRegister() {
    const verificationCode = document.getElementById('verification-code').value;

    if (!verificationCode) {
      showMessage('Please enter the verification code', 'error');
      return;
    }

    if (!this.pendingRegistration) {
      showMessage('Registration data not found. Please start over.', 'error');
      this.showLoginForm();
      return;
    }

    const verifyButton = document.getElementById('verify-register-button');
    const originalText = verifyButton.innerHTML;
    verifyButton.innerHTML = createLoadingSpinner().outerHTML;
    verifyButton.disabled = true;

    try {
      const { username, password, email } = this.pendingRegistration;
      const response = await apiService.register(username, password, email, verificationCode);
      
      showMessage('Registration successful! You are now logged in.', 'success');
      
      // Auto-login after successful registration
      await this.loginSuccess(response);
      
      // Clear pending registration data
      this.pendingRegistration = null;
      
    } catch (error) {
      console.error('Registration error:', error);
      if (error.message.includes('wrong email key')) {
        showMessage('Invalid verification code. Please check your email and try again.', 'error');
      } else if (error.message.includes('already exists')) {
        showMessage('Username already exists. Please choose a different username.', 'error');
        this.showRegistrationForm();
      } else {
        showMessage('Registration failed. Please try again.', 'error');
      }
    } finally {
      verifyButton.innerHTML = originalText;
      verifyButton.disabled = false;
    }
  }

  clearForm() {
    document.getElementById('username').value = '';
    document.getElementById('password').value = '';
    document.getElementById('email').value = '';
    document.getElementById('verification-code').value = '';
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
