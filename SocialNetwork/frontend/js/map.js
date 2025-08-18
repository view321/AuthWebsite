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

    // Only display parent notes (not replies) on the map
    const parentNotes = notes.filter(note => note.parent_id === 0);
    
    parentNotes.forEach(note => {
      const marker = L.marker([note.lattitude, note.longitude])
        .addTo(this.notesLayerGroup);
      
      // Add click event to show note modal with replies
      marker.on('click', () => {
        const allReplies = this.findAllReplies(note.id, notes);
        this.showNoteModal(note.text, note.user_id, note.id, allReplies);
      });
    });
  }

  findAllReplies(noteId, allNotes) {
    const directReplies = allNotes.filter(note => note.parent_id === noteId);
    let allDescendants = [...directReplies];

    directReplies.forEach(reply => {
      const descendants = this.findAllReplies(reply.id, allNotes);
      allDescendants = allDescendants.concat(descendants);
    });

    return allDescendants;
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

  showNoteModal(noteText, noteAuthor, noteId, replies = []) {
    const noteModal = document.getElementById('note-modal');
    const noteModalAuthor = document.getElementById('note-modal-author');
    const noteModalBody = document.getElementById('note-modal-body');
    const noteModalActions = document.getElementById('note-modal-actions');
    
    noteModalAuthor.textContent = `${noteAuthor}`;
    
    // Build content with original note and threaded replies
    let content = `<div class="original-note">${noteText}</div>`;
    
    if (replies.length > 0) {
      content += `<div class="replies-section">
        <div class="replies-header">
          <i class="fas fa-comments"></i> ${replies.length} ${replies.length === 1 ? 'reply' : 'replies'}
        </div>
        <div class="replies-list">`;
      
      // Build threaded reply structure
      const threadedReplies = this.buildReplyThread(replies, noteId);
      content += this.renderReplyThread(threadedReplies, 0);
      
      content += `</div></div>`;
    }
    
    // Add main reply form if authenticated
    if (apiService.isAuthenticated) {
      content += `
        <div class="reply-form-section">
          <div class="reply-form compact">
            <textarea id="reply-text-${noteId}" placeholder="Write a reply..." rows="2"></textarea>
            <button class="btn btn-sm btn-primary reply-submit-btn" data-parent-id="${noteId}" data-target-id="reply-text-${noteId}">
              <i class="fas fa-paper-plane"></i> Reply
            </button>
          </div>
        </div>`;
    }
    
    noteModalBody.innerHTML = content;
    
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
    
    // Add event listeners for reply interactions
    if (apiService.isAuthenticated) {
      setTimeout(() => {
        // Reply submission buttons
        const replyBtns = noteModalBody.querySelectorAll('.reply-submit-btn');
        replyBtns.forEach(btn => {
          btn.addEventListener('click', () => {
            const parentId = btn.getAttribute('data-parent-id');
            const targetId = btn.getAttribute('data-target-id');
            this.handleReplySubmission(parentId, targetId);
          });
        });

        // Reply-to-reply buttons
        const replyToReplyBtns = noteModalBody.querySelectorAll('.reply-to-reply-btn');
        replyToReplyBtns.forEach(btn => {
          btn.addEventListener('click', () => {
            const replyId = btn.getAttribute('data-parent-id');
            this.toggleReplyForm(replyId);
          });
        });

        // Cancel reply buttons
        const cancelBtns = noteModalBody.querySelectorAll('.cancel-reply-btn');
        cancelBtns.forEach(btn => {
          btn.addEventListener('click', () => {
            const replyId = btn.getAttribute('data-reply-id');
            this.hideReplyForm(replyId);
          });
        });
      }, 100);
    }
    
    noteModal.classList.add('show');
  }

  buildReplyThread(replies, rootId) {
    const replyMap = new Map();
    
    // Create a map of all replies by their ID
    replies.forEach(reply => {
      replyMap.set(reply.id, { ...reply, children: [] });
    });

    const threadedReplies = [];

    // Build the thread structure
    replies.forEach(reply => {
      if (reply.parent_id === rootId) {
        // This is a root-level reply
        threadedReplies.push(replyMap.get(reply.id));
      } else {
        // This is a nested reply
        const parent = replyMap.get(reply.parent_id);
        if (parent) {
          parent.children.push(replyMap.get(reply.id));
        }
      }
    });

    return threadedReplies;
  }

  renderReplyThread(replies, depth) {
    let html = '';
    
    replies.forEach(reply => {
      const indentClass = depth > 0 ? `reply-indent-${Math.min(depth, 3)}` : '';
      
      html += `
        <div class="reply-item ${indentClass}" data-reply-id="${reply.id}">
          <div class="reply-header">
            <span class="reply-author">${reply.user_id}</span>
            ${apiService.isAuthenticated ? `<button class="reply-to-reply-btn" data-parent-id="${reply.id}"><i class="fas fa-reply"></i></button>` : ''}
          </div>
          <div class="reply-text">${reply.text}</div>
          <div class="reply-form-container" id="reply-form-${reply.id}" style="display: none;">
            <div class="reply-form compact nested">
              <textarea id="reply-text-${reply.id}" placeholder="Reply to ${reply.user_id}..." rows="2"></textarea>
              <div class="reply-actions">
                <button class="btn btn-xs btn-secondary cancel-reply-btn" data-reply-id="${reply.id}">Cancel</button>
                <button class="btn btn-xs btn-primary reply-submit-btn" data-parent-id="${reply.id}" data-target-id="reply-text-${reply.id}">Reply</button>
              </div>
            </div>
          </div>
        </div>`;
      
      // Render children with increased depth
      if (reply.children && reply.children.length > 0) {
        html += this.renderReplyThread(reply.children, depth + 1);
      }
    });
    
    return html;
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

  async handleReplySubmission(parentId, targetId) {
    const replyText = document.getElementById(targetId).value.trim();
    
    if (!replyText) {
      showMessage('Please enter a reply message', 'error');
      return;
    }

    const replyBtn = document.querySelector(`[data-target-id="${targetId}"]`);
    const originalText = replyBtn.innerHTML;
    replyBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
    replyBtn.disabled = true;

    try {
      const replyData = {
        text: replyText,
        longitude: 0,
        lattitude: 0,
        public: false,
        allowed_users: [],
        parent_id: parseInt(parentId)
      };

      await apiService.createNote(replyData);
      showMessage('Reply added!', 'success');
      
      // Clear the reply form
      document.getElementById(targetId).value = '';
      
      // Hide nested reply form if it exists
      const replyForm = document.getElementById(`reply-form-${parentId}`);
      if (replyForm) {
        replyForm.style.display = 'none';
      }
      
      // Refresh the notes
      await this.fetchNotes();
      document.getElementById('note-modal').classList.remove('show');
      
    } catch (error) {
      console.error('Error creating reply:', error);
      showMessage('Failed to add reply', 'error');
    } finally {
      replyBtn.innerHTML = originalText;
      replyBtn.disabled = false;
    }
  }

  toggleReplyForm(replyId) {
    const replyForm = document.getElementById(`reply-form-${replyId}`);
    if (replyForm) {
      const isVisible = replyForm.style.display !== 'none';
      replyForm.style.display = isVisible ? 'none' : 'block';
      
      if (!isVisible) {
        // Focus on the textarea when showing
        setTimeout(() => {
          const textarea = document.getElementById(`reply-text-${replyId}`);
          if (textarea) textarea.focus();
        }, 100);
      }
    }
  }

  hideReplyForm(replyId) {
    const replyForm = document.getElementById(`reply-form-${replyId}`);
    if (replyForm) {
      replyForm.style.display = 'none';
      // Clear the textarea
      const textarea = document.getElementById(`reply-text-${replyId}`);
      if (textarea) textarea.value = '';
    }
  }

  getCurrentLocation() {
    return this.currentNoteLocation;
  }
}

export default new MapManager();
