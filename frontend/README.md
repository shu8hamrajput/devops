# Splitwise Frontend

React-based frontend application for the Splitwise expense sharing backend.

## Features

- **Users Management**: Create and view users
- **Groups Management**: Create groups and add users to them
- **Expenses Management**: Create and view expenses for groups

## Getting Started

### Prerequisites

- Node.js (v16 or higher)
- npm or yarn
- The backend server running on `http://localhost:8080`

### Installation

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

The frontend will be available at `http://localhost:3000`

### Building for Production

```bash
npm run build
```

The built files will be in the `dist` directory.

## Project Structure

```
frontend/
├── src/
│   ├── components/      # React components
│   │   ├── Users.jsx
│   │   ├── Groups.jsx
│   │   └── Expenses.jsx
│   ├── services/        # API service layer
│   │   └── api.js
│   ├── App.jsx          # Main app component
│   ├── App.css
│   ├── main.jsx         # Entry point
│   └── index.css        # Global styles
├── index.html
├── package.json
└── vite.config.js
```

## API Integration

The frontend communicates with the backend API running on `http://localhost:8080`. The API service layer is configured in `src/services/api.js`.

### CORS Configuration

If you encounter CORS errors when making API requests, you may need to add CORS middleware to your Go backend. You can use the `github.com/rs/cors` package:

```go
import "github.com/rs/cors"

// In main.go, wrap your router:
c := cors.New(cors.Options{
    AllowedOrigins: []string{"http://localhost:3000"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{"*"},
})
handler := c.Handler(router)
log.Fatal(http.ListenAndServe(port, handler))
```

### Available Endpoints

- **Users**: `/users` (POST, GET), `/users/{id}` (GET)
- **Groups**: `/groups` (POST, GET), `/groups/{id}/users` (POST)
- **Expenses**: `/expenses` (POST), `/expenses/{id}` (GET), `/groups/{group_id}/expenses` (GET)

## Development

The app uses Vite as the build tool for fast development and hot module replacement.

### Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

## Technologies Used

- React 18
- Vite
- Axios (for API calls)
- CSS3 (for styling)

