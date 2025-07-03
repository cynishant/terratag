# Terratag UI

A modern React-based web interface for managing tag standardization and running validation/tagging operations with Terratag.

## Features

### ✨ Tag Standardization Management
- **Create & Edit Standards**: Visual form-based editor with support for YAML editing
- **Multi-Cloud Support**: AWS, Azure, GCP, and generic cloud provider configurations
- **Version Control**: Track versions of your tag standards
- **Export/Import**: Download standards as YAML files
- **Provider Filtering**: Filter standards by cloud provider

### 🎯 Tag Validation & Application
- **Operation Management**: Create and monitor validation and tagging operations
- **Real-time Status**: Track operation progress with live updates
- **Detailed Results**: View comprehensive results with file-by-file breakdown
- **Error Handling**: Clear error messages and resolution guidance

### 📊 Results & Analytics
- **Operation History**: View past operations and their outcomes
- **Statistics Dashboard**: Compliance rates, resource counts, and violation summaries
- **Detailed Logs**: Access operation logs for debugging and auditing

## Technology Stack

- **Frontend**: React 18 + TypeScript
- **Styling**: Tailwind CSS with custom design system
- **Icons**: Lucide React for modern, consistent icons
- **State Management**: Zustand for lightweight, efficient state management
- **API Client**: Axios with type-safe API integration
- **Routing**: React Router for SPA navigation

## UI/UX Design Principles

### Modern Professional Interface
- Clean, minimal design following modern UI patterns
- Consistent spacing, typography, and color scheme
- Responsive design that works on desktop, tablet, and mobile
- Accessible components with proper ARIA labels and keyboard navigation

### User Experience
- **Intuitive Navigation**: Clear menu structure and breadcrumbs
- **Progressive Disclosure**: Information revealed as needed
- **Immediate Feedback**: Loading states, success/error messages
- **Efficient Workflows**: Streamlined processes for common tasks

### Design System
- **Colors**: Primary blue theme with semantic color usage
- **Typography**: Clean, readable font hierarchy
- **Components**: Reusable UI components with consistent styling
- **Layout**: Grid-based layout with proper spacing and alignment

## Getting Started

### Prerequisites
- Node.js 18+ and npm
- Go 1.23+ (for API server)
- Make (optional, for convenience commands)

### Quick Start

1. **Setup the environment**:
   ```bash
   make setup
   ```

2. **Start development servers**:
   ```bash
   make dev
   ```
   This starts both the API server (port 8080) and UI dev server (port 3000).

3. **Access the UI**:
   - Development: http://localhost:3000
   - Production: http://localhost:8080 (when API is serving static files)

### Manual Setup

1. **Install UI dependencies**:
   ```bash
   cd web/ui
   npm install
   ```

2. **Start the API server**:
   ```bash
   go run cmd/api/main.go
   ```

3. **Start the UI development server**:
   ```bash
   cd web/ui
   REACT_APP_API_URL=http://localhost:8080/api/v1 npm start
   ```

## Building for Production

### Build UI for production:
```bash
make build-ui
# or
cd web/ui && npm run build
```

### Build complete application:
```bash
make build
```

This creates:
- `bin/terratag-api`: API server binary
- `bin/terratag`: CLI tool binary  
- `web/ui/build/`: Production UI build

## Project Structure

```
web/ui/
├── public/                 # Static assets
├── src/
│   ├── api/               # API client and HTTP utilities
│   ├── components/        # Reusable UI components
│   │   ├── common/        # Generic components (Button, Modal, etc.)
│   │   ├── layout/        # Layout components (Header, Navigation)
│   │   └── standards/     # Tag standards specific components
│   ├── pages/             # Page components
│   ├── store/             # State management (Zustand)
│   ├── types/             # TypeScript type definitions
│   └── utils/             # Utility functions
├── package.json
├── tailwind.config.js     # Tailwind CSS configuration
└── tsconfig.json          # TypeScript configuration
```

## Available Scripts

In the `web/ui` directory:

- `npm start`: Start development server
- `npm run build`: Build for production
- `npm test`: Run tests
- `npm run lint`: Run ESLint
- `npm run eject`: Eject from Create React App (not recommended)

## Configuration

### Environment Variables

The UI can be configured with the following environment variables:

- `REACT_APP_API_URL`: API server URL (default: http://localhost:8080/api/v1)

### API Integration

The UI communicates with the Terratag API server through REST endpoints:

- `GET /api/v1/standards`: List tag standards
- `POST /api/v1/standards`: Create tag standard
- `PUT /api/v1/standards/:id`: Update tag standard
- `DELETE /api/v1/standards/:id`: Delete tag standard
- `POST /api/v1/operations`: Create operation
- `GET /api/v1/operations/:id/summary`: Get operation results

## Development Guidelines

### Component Development
- Use TypeScript for all components
- Follow React functional component patterns
- Implement proper error boundaries
- Use proper loading states

### Styling Guidelines
- Use Tailwind CSS utility classes
- Follow the established design system
- Ensure responsive design
- Maintain accessibility standards

### State Management
- Use Zustand for global state
- Keep component state local when possible
- Implement proper error handling
- Use optimistic updates where appropriate

## Testing

### Unit Tests
```bash
cd web/ui
npm test
```

### Test Coverage
```bash
cd web/ui
npm test -- --coverage
```

## Deployment

### Docker Deployment
The UI is included in the main Terratag Docker image:

```bash
docker build -t terratag .
docker run -p 8080:8080 terratag
```

### Static File Serving
The API server serves the built UI files from `/` when running in production mode.

## Contributing

1. Follow the existing code style and patterns
2. Add tests for new functionality
3. Update documentation as needed
4. Ensure all lint checks pass
5. Test on multiple browsers and screen sizes

## Browser Support

- Chrome 88+
- Firefox 85+
- Safari 14+
- Edge 88+

## Performance

The UI is optimized for performance with:
- Code splitting for lazy loading
- Optimized bundle sizes
- Efficient state management
- Minimal re-renders
- Compressed assets

## Accessibility

The UI follows WCAG 2.1 guidelines:
- Proper semantic HTML
- Keyboard navigation support
- Screen reader compatibility
- High contrast color scheme
- Focus management