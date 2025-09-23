# GoBE Web Dashboard

Modern, responsive web dashboard for the GoBE backend system.

## Features

### 🎨 Modern Design

- **Dark theme** with glassmorphism effects
- **Responsive design** that works on all devices
- **Animated components** with smooth transitions
- **Professional typography** using Inter font

### 🚀 Real-time Functionality

- **Live system status** via MCP protocol integration
- **Auto-refresh** every 30 seconds
- **Interactive testing** of all system components
- **Toast notifications** for user feedback

### 🛠️ Technical Features

- **MCP Integration**: Tests `/mcp/tools` and `/mcp/exec` endpoints
- **API Documentation**: Shows available REST endpoints
- **System Metrics**: Real-time Go runtime statistics
- **Health Monitoring**: Database, integrations, and service status
- **Service Worker**: Offline caching for better performance

## File Structure

```
web/
├── index.html          # Main landing page
├── style.css           # Modern CSS with CSS variables
├── app.js              # Modular JavaScript functionality
├── sw.js               # Service Worker for caching
└── README.md           # This documentation
```

## Features Breakdown

### Dashboard Cards

1. **System Status** - Real-time metrics from MCP `system.status` tool
2. **MCP Protocol** - Model Context Protocol tools and testing
3. **REST API** - Available endpoints with HTTP methods
4. **Database** - Connection status and health checks
5. **Security** - TLS, JWT, Rate Limiting, CORS status
6. **Integrations** - WhatsApp, Telegram, RabbitMQ status

### Interactive Actions

- **Health Check** - Opens `/health` endpoint
- **Test System.Status** - Tests MCP system.status tool
- **Run Full Test** - Comprehensive system test suite
- **View Logs** - Placeholder for log viewer
- **Download Config** - Generates and downloads configuration JSON
- **Documentation** - Links to GitHub repository

### Advanced Features

#### Keyboard Shortcuts

- `Ctrl+R` - Refresh system status
- `Ctrl+T` - Run comprehensive system test

#### JavaScript API

```javascript
// Access dashboard instance
const dashboard = new GoBEDashboard();

// Available methods
dashboard.refreshSystemStatus()
dashboard.testMCPTools()
dashboard.runSystemTest()
dashboard.downloadConfig()
```

#### CSS Variables

```css
:root {
    --primary-gradient: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    --accent-color: #00ff88;
    --bg-dark: #0d1117;
    /* ... more variables */
}
```

## Integration with GoBE

### MCP Protocol Integration

The dashboard connects to GoBE's MCP endpoints:

- `GET /mcp/tools` - Lists available MCP tools
- `POST /mcp/exec` - Executes MCP tools with arguments

### System Status

Uses the `system.status` MCP tool to get:

- Server version and uptime
- Go runtime metrics (memory, goroutines)
- System health information

### API Testing

Tests various GoBE endpoints:

- `/health` - Basic health check
- `/api/v1/system/metrics` - System metrics
- `/api/v1/{integration}/ping` - Integration status

## Performance Optimizations

### Loading Strategy

- **Critical CSS** inlined in HTML head
- **External CSS** loaded asynchronously
- **JavaScript modules** for better bundling

### Caching

- **Service Worker** caches static assets
- **Version-based** cache invalidation
- **Offline fallback** for better UX

### Responsive Design

- **CSS Grid** for dashboard layout
- **Flexbox** for component alignment
- **Media queries** for mobile optimization

## Browser Support

- ✅ Chrome 60+
- ✅ Firefox 55+
- ✅ Safari 11+
- ✅ Edge 79+

### Progressive Enhancement

- Works without JavaScript (basic view)
- Enhanced functionality with JavaScript enabled
- Offline support via Service Worker

## Development

### Local Testing

```bash
# Start GoBE server
./gobe start

# Open browser
open http://localhost:3666
```

### Customization

- Modify CSS variables in `style.css`
- Add new functionality in `app.js`
- Update layout in `index.html`

## Security Considerations

- **No sensitive data** exposed in frontend
- **CORS-compliant** requests to backend
- **Content Security Policy** ready
- **Input validation** on all forms

## Future Enhancements

- Real-time WebSocket connections
- Advanced metrics visualizations
- User authentication integration
- Theme customization options
- Multi-language support

---

***Built with ❤️ for the Kubex Ecosystem***
