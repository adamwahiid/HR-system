# Postman Testing Checklist

Use this checklist to manually verify every route in your Employee Management System.

---

## Environment Setup
First, make sure your `.env` file exists and the Go server is running (`go run .`).
In Postman, it is helpful to set up an environment variable called `{{token}}` to easily paste your JWT after login.

---

## 1. Authentication

### Login
- **Method:** `POST`
- **URL:** `http://localhost:8080/login`
- **Auth:** None
- **Body (JSON):**
  ```json
  {
      "email": "admin@company.com",
      "password": "password123"
  }
  ```
- **Expected Success:** `200 OK` with a `"token"` field.
- **Important Permission Test:** Try an invalid password to ensure you get a `401 Unauthorized`.

### Logout
- **Method:** `POST`
- **URL:** `http://localhost:8080/logout`
- **Auth:** Bearer Token
- **Expected Success:** `200 OK` with message "Logged out successfully."

---

## 2. Admin & HR (Account Creation)

### Create User (Worker/Manager/Board Member)
- **Method:** `POST`
- **URL:** `http://localhost:8080/admin/users`
- **Auth:** Bearer Token (must be `admin` or `hr`)
- **Body (JSON):**
  ```json
  {
      "name": "New Worker",
      "email": "worker1@company.com",
      "password": "password123",
      "role_id": 4, 
      "salary": 50000.0,
      "manager_id": 1
  }
  ```
- **Expected Success:** `201 Created`
- **Important Permission Test:** Try calling this route using a Worker's token. It must return `403 Forbidden`.

### Create User (New Admin/HR)
- **Method:** `POST`
- **URL:** `http://localhost:8080/admin/users`
- **Auth:** Bearer Token (must be `admin` or `hr`)
- **Body (JSON):**
  ```json
  {
      "name": "New Admin",
      "email": "admin2@company.com",
      "password": "password123",
      "role_id": 2
  }
  ```
- **Expected Success:** `201 Created`
- **Important Fix Test:** Verify that this succeeds and you do *not* get an "unsupported role" error (testing our recent fix!).

---

## 3. Workers

### Get All Workers
- **Method:** `GET`
- **URL:** `http://localhost:8080/workers`
- **Auth:** Bearer Token (Any valid token)
- **Expected Success:** `200 OK` with JSON array of workers.

### Get Worker By ID
- **Method:** `GET`
- **URL:** `http://localhost:8080/workers/1`
- **Auth:** Bearer Token (Any valid token)
- **Expected Success:** `200 OK` with the specific worker's details.

### Update Worker
- **Method:** `PUT`
- **URL:** `http://localhost:8080/workers/1`
- **Auth:** Bearer Token (Admin, HR, Worker's specific Manager, or the Worker themselves)
- **Body (JSON):**
  ```json
  {
      "name": "Updated Name",
      "email": "updated@company.com",
      "salary": 60000.0,
      "manager_id": 2
  }
  ```
- **Expected Success:** `200 OK`
- **Important Permission Test:** As a regular worker, try updating your `"salary"`. The API should succeed in updating your name, but it should completely ignore your attempt to change the salary.

### Delete Worker
- **Method:** `DELETE`
- **URL:** `http://localhost:8080/workers/1`
- **Auth:** Bearer Token (must be `admin` or `hr`)
- **Expected Success:** `200 OK`
- **Important Permission Test:** Try deleting yourself using a Worker's token. It must return `403 Forbidden`.

---

## 4. Managers

### Get All Managers
- **Method:** `GET`
- **URL:** `http://localhost:8080/managers`
- **Auth:** Bearer Token (Any valid token)
- **Expected Success:** `200 OK`

### Get Manager By ID
- **Method:** `GET`
- **URL:** `http://localhost:8080/managers/1`
- **Auth:** Bearer Token (Any valid token)
- **Expected Success:** `200 OK`

### Update Manager
- **Method:** `PUT`
- **URL:** `http://localhost:8080/managers/1`
- **Auth:** Bearer Token (Admin, HR, Manager's specific Board Member, or the Manager themselves)
- **Body (JSON):**
  ```json
  {
      "name": "Updated Manager",
      "email": "manager_new@company.com"
  }
  ```
- **Expected Success:** `200 OK`
- **Important Permission Test:** Send an invalid `board_mem_id` as an Admin (e.g. `999`). It must return `400 Bad Request` with "Provided board_mem_id does not exist".

### Delete Manager
- **Method:** `DELETE`
- **URL:** `http://localhost:8080/managers/1`
- **Auth:** Bearer Token (must be `admin` or `hr`)
- **Expected Success:** `200 OK`

---

## 5. Board Members

### Get All Board Members
- **Method:** `GET`
- **URL:** `http://localhost:8080/board-members`
- **Auth:** Bearer Token (Any valid token)
- **Expected Success:** `200 OK`

### Get Board Member By ID
- **Method:** `GET`
- **URL:** `http://localhost:8080/board-members/1`
- **Auth:** Bearer Token (Any valid token)
- **Expected Success:** `200 OK`

### Update Board Member
- **Method:** `PUT`
- **URL:** `http://localhost:8080/board-members/1`
- **Auth:** Bearer Token (Admin, HR, or the Board Member themselves)
- **Body (JSON):**
  ```json
  {
      "name": "Updated Board Member",
      "email": "board_new@company.com"
  }
  ```
- **Expected Success:** `200 OK`
- **Important Permission Test:** After successfully updating the email here, manually check your database `users` table. The email there MUST have changed to match the new one (proving the transaction works).

### Delete Board Member
- **Method:** `DELETE`
- **URL:** `http://localhost:8080/board-members/1`
- **Auth:** Bearer Token (must be `admin` or `hr`)
- **Expected Success:** `200 OK`
