Here's the `README.md` formatted for your Go project, documenting the API and its functionality:

```markdown
# GoTasker

GoTasker is a simple task management web application built with **Go** and **Gin**. It allows users to manage tasks, track progress, and organize their work. The app uses **JWT** for authentication, **GORM** for database management, and **PostgreSQL** for data storage.

## Features

- **User Authentication**: Register, login, and logout with JWT token-based authentication.
- **Task Management**: Create, update, delete, and retrieve tasks with due dates, statuses, and priorities.
- **Task Statistics**: View stats like total, completed, and pending tasks.
- **Due Soon Tasks**: Retrieve tasks that are due within the next 3 days.
- **Pagination and Search**: Tasks can be paginated and filtered based on title, status, or due date.

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [API Endpoints](#api-endpoints)
- [Database Schema](#database-schema)
- [License](#license)

## Requirements

- Go 1.16 or later
- PostgreSQL 13 or later
- Gin Framework
- GORM (ORM for Go)
- JWT (JSON Web Token) for authentication

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/gotasker.git
   cd gotasker
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up PostgreSQL and create a database:
   - Ensure PostgreSQL is installed and running.
   - Create a database for GoTasker (e.g., `gotasker`).

4. Configure the database connection:
   - Edit the `dsn` in the `main.go` file:
   ```go
   dsn := "host=localhost user=postgres password=admin dbname=gotasker port=5432 sslmode=disable"
   ```
   - Replace the database connection details as needed.

5. Run the application:
   ```bash
   go run main.go
   ```
   The server will start at `http://localhost:8080`.

## API Endpoints

### User Authentication

#### Register

- **Endpoint**: `POST /register`
- **Description**: Registers a new user.
- **Request Body**:
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```
- **Response**:
  ```json
  {
    "message": "registered"
  }
  ```

#### Login

- **Endpoint**: `POST /login`
- **Description**: Logs in a user and returns a JWT token.
- **Request Body**:
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```
- **Response**:
  ```json
  {
    "token": "jwt_token"
  }
  ```

### Task Management

#### Create Task

- **Endpoint**: `POST /tasks`
- **Description**: Creates a new task.
- **Request Body**:
  ```json
  {
    "title": "string",
    "status": "string",
    "due_date": "yyyy-mm-ddT00:00:00Z"
  }
  ```
- **Response**:
  ```json
  {
    "id": 1,
    "title": "string",
    "status": "string",
    "due_date": "yyyy-mm-ddT00:00:00Z",
    "user_id": 1
  }
  ```

#### Get Tasks

- **Endpoint**: `GET /tasks`
- **Description**: Retrieves a list of tasks, supports pagination and filtering.
- **Query Parameters**:
  - `page`: Page number (default: 1)
  - `limit`: Number of tasks per page (default: 10)
  - `search`: Search term for task titles
  - `status`: Filter by task status
  - `due`: Filter by due date (format: `yyyy-mm-dd`)
- **Response**:
  ```json
  {
    "tasks": [/* list of tasks */],
    "total": 10,
    "page": 1,
    "limit": 10,
    "total_pages": 1
  }
  ```

#### Update Task

- **Endpoint**: `PUT /tasks/:id`
- **Description**: Updates an existing task.
- **Request Body**:
  ```json
  {
    "title": "string",
    "status": "string",
    "due_date": "yyyy-mm-ddT00:00:00Z"
  }
  ```
- **Response**:
  ```json
  {
    "id": 1,
    "title": "updated title",
    "status": "updated status",
    "due_date": "yyyy-mm-ddT00:00:00Z"
  }
  ```

#### Delete Task

- **Endpoint**: `DELETE /tasks/:id`
- **Description**: Deletes a task.
- **Response**:
  ```json
  {
    "message": "task deleted"
  }
  ```

#### Get Due Soon Tasks

- **Endpoint**: `GET /tasks/due-soon`
- **Description**: Retrieves tasks that are due in the next 3 days.
- **Response**:
  ```json
  [
    {
      "id": 1,
      "title": "string",
      "status": "todo",
      "due_date": "yyyy-mm-ddT00:00:00Z"
    }
  ]
  ```

#### Get Task Stats

- **Endpoint**: `GET /tasks/stats`
- **Description**: Retrieves statistics for tasks, including total, completed, and pending tasks.
- **Response**:
  ```json
  {
    "total_tasks": 10,
    "completed": 5,
    "pending": 5
  }
  ```

## Database Schema

### User

- `id`: Primary Key
- `username`: Unique
- `password`: Hashed password
- `theme`: Light/Dark theme (default: light)

### Task

- `id`: Primary Key
- `title`: Task title
- `user_id`: Foreign key to User
- `status`: Task status (default: todo)
- `due_date`: Task due date
- `priority`: Task priority (default: medium)

### TaskHistory

- `id`: Primary Key
- `task_id`: Foreign key to Task
- `changed_by`: User ID of the person who made the change
- `field`: Field that was changed
- `old_value`: Old value of the field
- `new_value`: New value of the field
- `change_time`: Timestamp of the change

### Tag

- `id`: Primary Key
- `name`: Tag name
- `user_id`: Foreign key to User

### TaskTag

- `task_id`: Foreign key to Task
- `tag_id`: Foreign key to Tag

### TaskTimeLog

- `id`: Primary Key
- `task_id`: Foreign key to Task
- `user_id`: Foreign key to User
- `duration`: Time spent on the task
- `note`: Additional notes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```

This `README.md` document includes sections for installation, API endpoints, database schema, and a license. You can customize it as needed for your specific project!
