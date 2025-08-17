// Configuration constants
const CONFIG = {
  MAP: {
    DEFAULT_LAT: 51.505,
    DEFAULT_LNG: -0.09,
    DEFAULT_ZOOM: 13,
    TILE_LAYER: 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
    ATTRIBUTION: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
  },
  API: {
    BASE_URL: '',
    ENDPOINTS: {
      LOGIN: '/api/login',
      REGISTER: '/api/register_user',
      LOGOUT: '/api/login_protected/logout',
      CREATE_NOTE: '/api/login_protected/create_note',
      DELETE_NOTE: '/api/login_protected/edit_permission',
      GET_NOTES: '/api/view_permission/get_within_square'
    }
  },
  NOTE: {
    LONG_NOTE_LINES: 3,
    LONG_NOTE_CHARS: 200
  },
  UI: {
    MESSAGE_DURATION: 3000,
    ANIMATION_DURATION: 300
  }
};

export default CONFIG;
