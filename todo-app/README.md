# Todo List Application

A modern, full-stack todo list application built with Go backend using Godis for data storage and React frontend with a beautiful UI.

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React Frontend │    │  Go Backend     │    │     Godis       │
│                 │    │                 │    │                 │
│ • Zustand Store │◄──►│ • Gin Web       │◄──►│ • In-Memory     │
│ • Tailwind CSS  │    │ • REST API      │    │ • Redis Protocol│
│ • Lucide Icons  │    │ • Business Logic│    │ • Data Storage  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## ✨ Features

### Frontend Features
- **Modern UI**: Clean, responsive design with Tailwind CSS
- **Real-time Updates**: Immediate UI feedback with optimistic updates
- **Smart Filtering**: Filter by status, priority, and sort by multiple criteria
- **Form Validation**: Client-side validation with helpful error messages
- **Progressive Enhancement**: Works without JavaScript for basic functionality
- **Responsive Design**: Optimized for desktop, tablet, and mobile devices

### Backend Features
- **RESTful API**: Clean REST endpoints following best practices
- **Data Validation**: Server-side validation for all inputs
- **Error Handling**: Comprehensive error handling and logging
- **CORS Support**: Cross-origin requests for frontend integration
- **Performance**: Efficient data operations with Redis-style storage

### Todo Management
- ✅ Create, read, update, and delete todos
- 🎯 Priority levels (High, Medium, Low)
- 📅 Due date tracking with overdue indicators
- ✔️ Mark todos as complete/incomplete
- 📊 Statistics dashboard
- 🔍 Advanced filtering and sorting
- 📱 Mobile-friendly interface

## 🚀 Quick Start

### Prerequisites
- Go 1.18+ installed
- Node.js 18+ installed (for frontend development)
- Git

### Backend Setup

1. **Navigate to backend directory**:
   ```bash
   cd todo-app/backend
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Run the backend server**:
   ```bash
   go run main.go
   ```

   The backend server will start on `http://localhost:8080`

### Frontend Setup

1. **Navigate to frontend directory**:
   ```bash
   cd todo-app/frontend
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Start the development server**:
   ```bash
   npm run dev
   ```

   The frontend will be available at `http://localhost:3000`

### Using Docker (Optional)

Coming soon: Docker Compose setup for easy deployment.

## 📁 Project Structure

```
todo-app/
├── backend/                 # Go backend application
│   ├── handlers/           # HTTP request handlers
│   ├── models/             # Data models and types
│   ├── repository/         # Data access layer
│   ├── service/            # Business logic layer
│   ├── main.go            # Application entry point
│   └── go.mod             # Go module definition
├── frontend/               # React frontend application
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── services/      # API service layer
│   │   ├── store/         # Zustand state management
│   │   ├── styles/        # CSS styles
│   │   └── types/         # TypeScript type definitions
│   ├── public/            # Static assets
│   └── package.json       # Node.js dependencies
└── docs/                  # Documentation
```

## 🛠️ Technology Stack

### Backend
- **Go 1.18+**: Programming language
- **Gin**: Web framework
- **Godis**: Redis-compatible in-memory database
- **UUID**: Unique identifier generation

### Frontend
- **React 18**: UI framework
- **Vite**: Build tool and dev server
- **Zustand**: State management
- **Tailwind CSS**: Utility-first CSS framework
- **Lucide React**: Beautiful icons
- **Axios**: HTTP client

## 📚 API Documentation

### Base URL
```
http://localhost:8080/api
```

### Endpoints

#### Get All Todos
```http
GET /todos?completed=true&priority=high&sortBy=createdAt&order=desc
```

#### Create Todo
```http
POST /todos
Content-Type: application/json

{
  "title": "Complete project",
  "description": "Finish the todo app implementation",
  "priority": "high",
  "dueDate": "2024-12-31T23:59:59Z"
}
```

#### Get Todo by ID
```http
GET /todos/{id}
```

#### Update Todo
```http
PUT /todos/{id}
Content-Type: application/json

{
  "title": "Updated title",
  "completed": true
}
```

#### Delete Todo
```http
DELETE /todos/{id}
```

#### Toggle Todo Status
```http
PATCH /todos/{id}/toggle
```

#### Get Statistics
```http
GET /stats
```

### Response Format

#### Success Response
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Sample Todo",
    "description": "This is a sample todo",
    "completed": false,
    "priority": "medium",
    "dueDate": "2024-12-31T23:59:59Z",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "message": "Todo created successfully"
}
```

#### Error Response
```json
{
  "success": false,
  "error": "Todo not found",
  "code": "NOT_FOUND"
}
```

## 🧪 Testing

### Backend Tests
```bash
cd backend
go test ./...
```

### Frontend Tests
```bash
cd frontend
npm test
```

## 🚀 Deployment

### Backend Deployment
1. Build the binary:
   ```bash
   cd backend
   go build -o todo-backend main.go
   ```

2. Run the binary:
   ```bash
   ./todo-backend
   ```

### Frontend Deployment
1. Build for production:
   ```bash
   cd frontend
   npm run build
   ```

2. Serve the built files using any static file server.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin feature-name`
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Godis](https://github.com/HDT3213/godis) - For the excellent Redis implementation in Go
- [Gin](https://github.com/gin-gonic/gin) - For the fast HTTP web framework
- [React](https://reactjs.org/) - For the powerful UI library
- [Tailwind CSS](https://tailwindcss.com/) - For the utility-first CSS framework
- [Zustand](https://github.com/pmndrs/zustand) - For the simple state management

## 📞 Support

If you have any questions or run into issues, please:

1. Check the existing issues in the repository
2. Create a new issue with detailed information
3. Contact the maintainers

---

Made with ❤️ using Go and React