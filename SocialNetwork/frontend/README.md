# GeoNotes Frontend

A modern, modular frontend for the GeoNotes application built with vanilla JavaScript using ES6 modules.

## 📁 Directory Structure

```
frontend/
├── index.html          # Main HTML file
├── css/
│   └── styles.css      # All application styles
├── js/
│   ├── app.js          # Main application entry point
│   ├── config.js       # Configuration constants
│   ├── utils.js        # Utility functions
│   ├── api.js          # API service layer
│   ├── map.js          # Map management
│   ├── auth.js         # Authentication management
│   └── modals.js       # Modal management
└── README.md           # This file
```

## 🏗️ Architecture

The frontend is organized into modular components for better maintainability:

### Core Modules

- **`app.js`** - Main application entry point that initializes all modules
- **`config.js`** - Centralized configuration for API endpoints, map settings, and UI constants
- **`utils.js`** - Reusable utility functions (JWT parsing, message display, etc.)

### Feature Modules

- **`api.js`** - API service layer handling all HTTP requests
- **`map.js`** - Map management, note display, and map interactions
- **`auth.js`** - Authentication flow (login, register, logout)
- **`modals.js`** - Modal management and form handling

## 🚀 Getting Started

1. **Prerequisites**: Modern browser with ES6 module support
2. **Dependencies**: 
   - Leaflet.js (for maps)
   - Font Awesome (for icons)
   - Inter font (Google Fonts)

3. **Setup**: Simply open `index.html` in a web browser

## 🔧 Development

### Adding New Features

1. **Create a new module** in the `js/` directory
2. **Import dependencies** from existing modules
3. **Export the module** for use in other files
4. **Initialize** in `app.js` if needed

### Styling

- All styles are in `css/styles.css`
- Uses CSS custom properties for theming
- Responsive design with mobile-first approach
- Modern CSS features (Grid, Flexbox, CSS Variables)

### API Integration

- All API calls go through the `api.js` service
- Centralized error handling
- Automatic token management
- Consistent request/response patterns

## 📱 Features

- **Interactive Map**: Leaflet.js integration with custom markers
- **Authentication**: Login/register with JWT tokens
- **Note Management**: Create, read, delete location-based notes
- **Privacy Controls**: Public/private notes with user sharing
- **Responsive Design**: Works on desktop and mobile devices
- **Modern UI**: Clean, accessible interface with smooth animations

## 🛠️ Technical Details

### Module System
- Uses ES6 modules with `import`/`export`
- No bundler required - runs directly in modern browsers
- Clear separation of concerns

### State Management
- Minimal global state
- Module-specific state management
- Event-driven communication between modules

### Error Handling
- Centralized error handling in API service
- User-friendly error messages
- Graceful degradation

### Performance
- Lazy loading of map tiles
- Efficient DOM manipulation
- Minimal re-renders

## 🔒 Security

- JWT token-based authentication
- Secure API communication
- Input validation and sanitization
- XSS protection through proper DOM manipulation

## 🌐 Browser Support

- Modern browsers with ES6 module support
- Chrome 61+, Firefox 60+, Safari 10.1+, Edge 16+
- Mobile browsers with ES6 support

## 📝 Contributing

1. Follow the modular architecture
2. Add new features as separate modules
3. Update documentation for new features
4. Test across different browsers
5. Maintain consistent code style

## 🐛 Troubleshooting

### Common Issues

1. **Module loading errors**: Ensure server supports ES6 modules
2. **CORS issues**: Configure backend to allow frontend origin
3. **Map not loading**: Check Leaflet.js CDN availability
4. **Authentication issues**: Verify JWT token format and expiration

### Debug Mode

Enable browser developer tools to see detailed error messages and module loading status.
