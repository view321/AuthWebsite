import { showMessage, createLoadingSpinner } from './utils.js';
import mapManager from './map.js';

class ModalManager {
  constructor() {
    this.setupEventListeners();
  }

  setupEventListeners() {
    // Note modal controls
    const noteModal = document.getElementById('note-modal');
    const noteModalClose = document.getElementById('note-modal-close');

    noteModalClose.addEventListener('click', () => {
      noteModal.classList.remove('show');
    });

    noteModal.addEventListener('click', (e) => {
      if (e.target === noteModal) {
        noteModal.classList.remove('show');
      }
    });

    // Create note modal controls
    const createNoteModal = document.getElementById('create-note-modal');
    const createNoteClose = document.getElementById('create-note-close');
    const cancelNoteBtn = document.getElementById('cancel-note-btn');
    const sendNoteBtn = document.getElementById('send-note-btn');

    createNoteClose.addEventListener('click', () => {
      createNoteModal.classList.remove('show');
    });

    cancelNoteBtn.addEventListener('click', () => {
      createNoteModal.classList.remove('show');
    });

    createNoteModal.addEventListener('click', (e) => {
      if (e.target === createNoteModal) {
        createNoteModal.classList.remove('show');
      }
    });

    // Send note button
    sendNoteBtn.addEventListener('click', () => this.handleCreateNote());
  }

  async handleCreateNote() {
    const message = document.getElementById('note-text').value;
    const publicValue = document.getElementById('note-privacy').value === 'true';
    const allowedUsers = document.getElementById('note-allowed-users').value
      .split(',')
      .map(s => s.trim())
      .filter(s => s !== '');

    if (!message.trim()) {
      showMessage('Please enter a message', 'error');
      return;
    }

    const currentLocation = mapManager.getCurrentLocation();
    if (!currentLocation) {
      showMessage('No location selected', 'error');
      return;
    }

    const noteData = {
      text: message,
      longitude: currentLocation.lng,
      lattitude: currentLocation.lat,
      public: publicValue,
      allowed_users: allowedUsers
    };

    const sendNoteBtn = document.getElementById('send-note-btn');
    const originalText = sendNoteBtn.innerHTML;
    sendNoteBtn.innerHTML = createLoadingSpinner().outerHTML;
    sendNoteBtn.disabled = true;

    try {
      const success = await mapManager.createNote(noteData);
      if (success) {
        document.getElementById('create-note-modal').classList.remove('show');
      }
    } catch (error) {
      console.error('Error creating note:', error);
    } finally {
      sendNoteBtn.innerHTML = originalText;
      sendNoteBtn.disabled = false;
    }
  }

  showModal(modalId) {
    document.getElementById(modalId).classList.add('show');
  }

  hideModal(modalId) {
    document.getElementById(modalId).classList.remove('show');
  }
}

export default new ModalManager();
