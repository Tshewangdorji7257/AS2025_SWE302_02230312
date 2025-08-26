
---

**Module:** Software Testing & Quality Assurance

**Practical 2:** Writing Unit and Coverage Tests for a Go CRUD Server

---

## Objective

The goal of this practical was to learn how to write **unit tests** for a simple Go HTTP server and to measure **test coverage** using Go’s standard tools.

---

## What I Did

### 1. Project Setup

* Created a new Go project folder `go-crud-testing`.
* Initialized the module with `go mod init crud-testing`.
* Added three main files:

  * **main.go** → Starts the HTTP server.
  * **handlers.go** → Contains the CRUD logic for managing in-memory users.
  * **handlers\_test.go** → Contains unit tests for CRUD operations.

---

### 2. Server Implementation

* Built an in-memory **Users API** with CRUD operations:

  * **Create User** → `POST /users`
  * **Get All Users** → `GET /users`
  * **Get User by ID** → `GET /users/{id}`
  * **Update User** → `PUT /users/{id}`
  * **Delete User** → `DELETE /users/{id}`
* Used **Chi router** for routing and `sync.Mutex` for safe concurrent access.

---

### 3. Writing Unit Tests

* Used **`testing`** and **`httptest`** packages.
* Wrote tests for:

  * Creating a user (`TestCreateUserHandler`).
  * Fetching an existing and non-existing user (`TestGetUserHandler`).
  * Deleting a user (`TestDeleteUserHandler`).
* Added a helper function `resetState()` to reset the in-memory storage before each test.

---

### 4. Running Tests

* Ran tests with:

  ```bash
  go test -v
  ```
  ![alt text](<Screenshot 2025-08-25 114245.png>) 

* All tests passed successfully.


---

### 5. Code Coverage

* Checked coverage with:

  ```bash
  go test -v -cover
  ```
  ![alt text](<Screenshot 2025-08-25 114334.png>) 

* Achieved around **#38.6% coverage**.

* Generated detailed HTML coverage report with:

  ```bash
  go test -coverprofile=coverage.out
  go tool cover -html=coverage.out
  ```
![alt text](<Screenshot 2025-08-25 114401.png>)
![alt text](<Screenshot 2025-08-25 114519.png>)

* The report highlighted tested (green) and untested (red) code lines.

---

## Results

* Successfully implemented CRUD API with in-memory storage.
* Wrote unit tests for main handlers.
* Verified correctness through **status codes and JSON responses**.
* Achieved good code coverage (\~38.6%).
* Learned how to visualize coverage using Go’s `cover` tool.

---

## Conclusion

This practical helped me understand how to:

* Test HTTP handlers in Go using the standard library.
* Use mock requests and recorders for unit testing.
* Measure and improve test coverage.
* Ensure code reliability by covering different success and error cases.

---



