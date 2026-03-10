# Backend API Documentation

This document describes the backend API for the Knowledge Base System.
The API is designed to support a DeepSeek-style web interface with streaming chat responses and conversation management.

## Base URL
`http://localhost:8080`

## Authentication
Currently, the system uses a simple username-based authentication.
- **POST Requests**: Include `username` in the JSON body.
- **GET/DELETE Requests**: Include `username` as a query parameter (e.g., `?username=alice`).

### 1. Register
Create a new user account.

- **Endpoint**: `POST /api/register`
- **Request Body**:
  ```json
  {
    "username": "user1",
    "password": "password123"
  }
  ```
- **Response**:
  - `201 Created`: `{"message": "User registered successfully"}`
  - `409 Conflict`: `{"error": "User already exists"}`

### 2. Login
Verify user credentials.

- **Endpoint**: `POST /api/login`
- **Request Body**:
  ```json
  {
    "username": "user1",
    "password": "password123"
  }
  ```
- **Response**:
  - `200 OK`: `{"message": "Login successful"}`
  - `401 Unauthorized`: `{"error": "Invalid credentials"}`

---

## User Profile Management
Manage user profile information, including nickname, email, phone, and avatar.

### 2.1 Get User Profile
Get the profile information for a specific user.

- **Endpoint**: `GET /api/user/profile?username={username}`
- **Response**:
  - `200 OK`:
    ```json
    {
      "username": "user1",
      "email": "user1@example.com",
      "phone": "13800138000",
      "avatar": "/uploads/user1_1712345678_avatar.png",
      "nickname": "Student Alpha"
    }
    ```
  - `404 Not Found`: `{"error": "user not found"}`

### 2.2 Update User Profile
Update nickname, email, and phone number.

- **Endpoint**: `POST /api/user/update`
- **Request Body**:
  ```json
  {
    "username": "user1",
    "nickname": "Student Alpha",
    "email": "user1@example.com",
    "phone": "13800138000"
  }
  ```
- **Response**:
  - `200 OK`: `{"message": "Profile updated successfully"}`
  - `400 Bad Request`: `{"error": "Username is required"}`

### 2.3 Upload Avatar
Upload a new avatar image.

- **Endpoint**: `POST /api/user/avatar`
- **Content-Type**: `multipart/form-data`
- **Form Data**:
  - `username`: `user1`
  - `avatar`: (File)
- **Response**:
  - `200 OK`: `{"message": "Avatar uploaded successfully", "error": "/uploads/user1_1712345678_avatar.png"}`
  - `400 Bad Request`: `{"error": "Avatar file is required"}`

---

## Conversation Management
Manage chat sessions (conversations).

### 3. List Conversations
Get a list of all conversations for a user (summary only, no messages).

- **Endpoint**: `GET /api/conversations?username={username}`
- **Response**:
  - `200 OK`: Array of Conversation objects.
  ```json
  [
    {
      "id": "171234567890",
      "user_id": "user1",
      "title": "New Chat",
      "created_at": "2024-05-20T10:00:00Z",
      "updated_at": "2024-05-20T10:05:00Z"
    }
  ]
  ```

### 4. Create Conversation
Start a new chat session.

- **Endpoint**: `POST /api/conversations`
- **Request Body**:
  ```json
  {
    "username": "user1",
    "title": "Discussion about Exams"  // Optional, defaults to "New Chat"
  }
  ```
- **Response**:
  - `201 Created`: The created Conversation object.

### 5. Get Conversation Detail
Get the full history of a specific conversation, including all messages.

- **Endpoint**: `GET /api/conversations/{conversation_id}?username={username}`
- **Response**:
  - `200 OK`: Conversation object with `messages`.
  ```json
  {
    "id": "171234567890",
    "title": "Discussion about Exams",
    "messages": [
      {
        "id": "msg_1",
        "role": "user",
        "content": "When is the exam?",
        "timestamp": "..."
      },
      {
        "id": "msg_2",
        "role": "assistant",
        "content": "The exam is on June 1st.",
        "timestamp": "..."
      }
    ]
  }
  ```

### 6. Delete Conversation
Delete a specific conversation.

- **Endpoint**: `DELETE /api/conversations/{conversation_id}?username={username}`
- **Response**:
  - `200 OK`: `{"message": "Conversation deleted"}`

### 6.1 Rename Conversation
Rename a conversation's title manually.

- **Endpoint**: `POST /api/conversations/rename`
- **Request Body**:
  ```json
  {
    "conversation_id": "171234567890",
    "new_title": "My Updated Title"
  }
  ```
- **Response**:
  - `200 OK`: `{"message": "Conversation renamed successfully"}`
  - `404 Not Found`: `{"error": "conversation not found"}`

---

## Chat (Streaming)
Send a message and receive a streaming response (Server-Sent Events).

### 7. Send Message
- **Endpoint**: `POST /api/chat`
- **Headers**:
  - `Content-Type: application/json`
  - `Accept: text/event-stream`
- **Request Body**:
  ```json
  {
    "username": "user1",
    "conversation_id": "171234567890",
    "content": "Tell me about the campus map."
  }
  ```
- **Response**: **Server-Sent Events (SSE)**
  The server streams JSON objects prefixed with `data: `.

  **Event Types:**
  1. `status`: Updates on what the system is doing (Thinking/Searching).
     ```json
     data: {"type": "status", "content": "Searching school website..."}
     ```
  2. `token`: A chunk of the final answer text (append these to show the answer).
     ```json
     data: {"type": "token", "content": "The campus"}
     data: {"type": "token", "content": " map is"}
     data: {"type": "token", "content": " located..."}
     ```
  3. `error`: If something goes wrong.
     ```json
     data: {"type": "error", "content": "Failed to connect to LLM"}
     ```

  **Example Stream:**
  ```
  data: {"type": "status", "content": "Analyzing question..."}

  data: {"type": "status", "content": "Retrieving knowledge..."}

  data: {"type": "token", "content": "According"}

  data: {"type": "token", "content": " to the"}

  data: {"type": "token", "content": " documents..."}
  ```

## Frontend Implementation Tips
1. Use `EventSource` is not suitable for POST requests with body. Instead, use `fetch` with a reader or a library like `@microsoft/fetch-event-source` to handle POST SSE.
2. Maintain a local list of conversations and update the active one.
3. Display `status` events as "Thinking..." indicators (like DeepSeek).
4. Append `token` events to the message bubble in real-time.
