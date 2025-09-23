# ChatApp 🚀  

Real-time web chat application built with **Go**, **WebSocket**, and **JWT Authentication**.  
Supports user registration, login, profile management, and messaging in real-time.  

---

## ✨ Features
- 🔐 **Authentication**: JWT-based access & refresh tokens  
- 👤 **User Profile**: username and account info decoded directly from JWT  
- 💬 **Chat**: real-time messaging using WebSocket (with broadcast to all connected clients)  
- 🗄️ **Database**: PostgreSQL for user storage  
- 🖥️ **Frontend**: HTML, CSS, Bootstrap, JavaScript  
- 🐳 **Containerization**: Dockerized backend and database  

---

## 🛠️ Tech Stack
- **Language**: Go  
- **Backend**: Gin, Gorilla WebSocket, JWT, logrus  
- **Database**: PostgreSQL  
- **Frontend**: HTML, CSS, Bootstrap, Vanilla JS  
- **Auth**: JWT (access & refresh tokens)  
- **DevOps**: Docker  

---
## ⚡ How to Run

### 1. Clone the repository
```bash
git clone https://github.com/Asladck/WebChat.git
cd WebChat
```
### 2. Run with Docker
```bash
docker-compose up --build
```
### 3. Open in browser

Home page → http://localhost:9090

Sign in → /sign-in

Sign up → /sign-up

Chat → /chat

Profile → /profile

## 🔑 JWT Authentication

Access token contains user info (username, exp, iat).

Stored in localStorage after login.

Auto redirect to /sign-in if no valid token.
