class InstructionsManager {
  constructor() {
    this.instructions = document.getElementById('instructions');
    this.instructionsClose = document.getElementById('instructions-close');
    this.instructionsToggle = document.getElementById('instructions-toggle');
    this.isHidden = false;
    
    this.setupEventListeners();
  }

  setupEventListeners() {
    // Close button
    this.instructionsClose.addEventListener('click', () => {
      this.hide();
    });

    // Toggle button
    this.instructionsToggle.addEventListener('click', () => {
      this.show();
    });
  }

  hide() {
    this.instructions.classList.add('hidden');
    this.instructionsToggle.classList.add('show');
    this.isHidden = true;
    
    // Store preference in localStorage
    localStorage.setItem('instructionsHidden', 'true');
  }

  show() {
    this.instructions.classList.remove('hidden');
    this.instructionsToggle.classList.remove('show');
    this.isHidden = false;
    
    // Store preference in localStorage
    localStorage.setItem('instructionsHidden', 'false');
  }

  toggle() {
    if (this.isHidden) {
      this.show();
    } else {
      this.hide();
    }
  }

  // Check if user previously closed instructions
  checkPreviousState() {
    const wasHidden = localStorage.getItem('instructionsHidden') === 'true';
    if (wasHidden) {
      this.hide();
    }
  }

  // Initialize the instructions state
  initialize() {
    this.checkPreviousState();
  }
}

export default new InstructionsManager();
