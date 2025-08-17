import CONFIG from './config.js';

// Utility functions
export const showMessage = (message, type = 'success') => {
  const messageEl = document.createElement('div');
  messageEl.className = `message ${type}`;
  messageEl.textContent = message;
  document.body.appendChild(messageEl);
  
  setTimeout(() => {
    messageEl.remove();
  }, CONFIG.UI.MESSAGE_DURATION);
};

export const getCurrentUsername = () => {
  // This will be called from authManager.getCurrentUser()
  // The actual implementation will be handled by the auth manager
  return null;
};

export const isLongNote = (text) => {
  return text.split('\n').length > CONFIG.NOTE.LONG_NOTE_LINES || 
         text.length > CONFIG.NOTE.LONG_NOTE_CHARS;
};

export const createLoadingSpinner = () => {
  const spinner = document.createElement('div');
  spinner.className = 'loading';
  return spinner;
};

export const debounce = (func, wait) => {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
};
