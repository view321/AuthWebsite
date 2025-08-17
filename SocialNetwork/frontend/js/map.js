import CONFIG from './config.js';
import { isLongNote } from './utils.js';
import { showMessage } from './utils.js';
import apiService from './api.js';
import authManager from './auth.js';

class MapManager {
  constructor() {
    this.map = null;
    this.notesLayerGroup = null;
    this.currentNoteLocation = null;
    this.onNoteCreate = null;
  }

  initialize() {
    this.map = L.map('map').setView([
      CONFIG.MAP.DEFAULT_LAT, 
      CONFIG.MAP.DEFAULT_LNG
    ], CONFIG.MAP.DEFAULT_ZOOM);

    L.tileLayer(CONFIG.MAP.TILE_LAYER, {
      attribution: CONFIG.MAP.ATTRIBUTION
    }).addTo(this.map);

    this.notesLayerGroup = L.layerGroup().addTo(this.map);

    this.setupEventListeners();
    this.fetchNotes();
  }

  setupEventListeners() {
    // Right-click to create note
    this.map.on('contextmenu', (e) => {
      L.DomEvent.preventDefault(e);
      this.handleRightClick(e);
    });

    // Fetch notes when map moves
    this.map.on('moveend', () => {
      this.fetchNotes();
    });
  }

  handleRightClick(e) {
    this.currentNoteLocation = e.latlng;
    
    // Clear form and show modal
    document.getElementById('note-text').value = '';
    document.getElementById('note-privacy').value = 'true';
    document.getElementById('note-allowed-users').value = '';
    
    document.getElementById('create-note-modal').classList.add('show');
    
    // Focus on textarea
    setTimeout(() => {
      document.getElementById('note-text').focus();
    }, 100);
  }

  async fetchNotes() {
    try {
      const bounds = this.map.getBounds();
      const notes = await apiService.getNotes(bounds);
      this.displayNotes(notes);
    } catch (error) {
      console.error('Error fetching notes:', error);
    }
  }

  displayNotes(notes) {
    this.notesLayerGroup.clearLayers();

    notes.forEach(note => {
      const marker = L.marker([note.lattitude, note.longitude])
        .addTo(this.notesLayerGroup);
      
      // Add click event to show note modal
      marker.on('click', () => {
        this.showNoteModal(note.text, note.user_id, note.id);
      });
    });
  }

  async deleteNote(noteId) {
    if (!confirm('Are you sure you want to delete this note?')) {
      return;
    }

    try {
      await apiService.deleteNote(noteId);
      showMessage('Note deleted successfully!', 'success');
      
      // Close the note modal if it's open
      const noteModal = document.getElementById('note-modal');
      if (noteModal.classList.contains('show')) {
        noteModal.classList.remove('show');
      }
      
      this.fetchNotes();
    } catch (error) {
      console.error('Error deleting note:', error);
      showMessage('Failed to delete note. Please try again.', 'error');
    }
  }

  showNoteModal(noteText, noteAuthor, noteId) {
    const noteModal = document.getElementById('note-modal');
    const noteModalAuthor = document.getElementById('note-modal-author');
    const noteModalBody = document.getElementById('note-modal-body');
    const noteModalActions = document.getElementById('note-modal-actions');
    
    noteModalAuthor.textContent = `By: ${noteAuthor}`;
    noteModalBody.textContent = noteText;
    
    // Add delete button if user owns the note
    if (apiService.isAuthenticated && noteAuthor === authManager.getCurrentUser()) {
      noteModalActions.innerHTML = `
        <button class="btn btn-danger delete-note-btn" data-note-id="${noteId}">
          <i class="fas fa-trash"></i> Delete Note
        </button>
      `;
      
      // Add event listener for delete button
      setTimeout(() => {
        const deleteBtn = noteModalActions.querySelector('.delete-note-btn');
        if (deleteBtn) {
          deleteBtn.addEventListener('click', () => {
            const noteId = deleteBtn.getAttribute('data-note-id');
            this.deleteNote(noteId);
            noteModal.classList.remove('show');
          });
        }
      }, 100);
    } else {
      noteModalActions.innerHTML = '';
    }
    
    noteModal.classList.add('show');
  }

  async createNote(noteData) {
    try {
      await apiService.createNote(noteData);
      showMessage('Note created successfully!', 'success');
      this.fetchNotes();
      return true;
    } catch (error) {
      console.error('Error creating note:', error);
      showMessage('Failed to create note. Please try again.', 'error');
      return false;
    }
  }

  getCurrentLocation() {
    return this.currentNoteLocation;
  }
}

export default new MapManager();
