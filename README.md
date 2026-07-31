# Splitwise Clone API 💸

A fully functional, high-performance backend API for an expense-sharing application, built using **Go** and **PostgreSQL**. This project handles user relationships, group expenses, and complex debt tracking algorithms securely.

## 🚀 Features

* **Secure Authentication:** User registration and login protected by bcrypt password hashing and JWT (JSON Web Tokens).
* **User & Social Management:** Dedicated user profiles, friend lists, and the ability to create multi-user expense groups.
* **Algorithmic Splitting Engine:** Safely log expenses with support for four distinct split types:
  * **EQUAL:** Divides the total evenly among all selected participants.
  * **EXACT:** Users specify exactly how much each person owes (with automatic even splitting for remaining unspecified balances).
  * **PERCENTAGE:** Divides the bill by percentages (e.g., 60/40).
  * **RATIO:** Divides by shares (e.g., 2 shares to 1 share).
* **Balances Dashboard:** Instantly calculates condensed "Who I owe" and "Who owes me" metrics using optimized SQL aggregations.
* **Transactional Integrity:** All multi-step database operations (like logging an expense and multiple splits) are wrapped in ACID-compliant SQL transactions to prevent corrupted states.

## 🛠 Tech Stack

* **Language:** Go (1.22+)
* **Database:** PostgreSQL
* **Driver & Pool:** github.com/jackc/pgx/v5/pgxpool
* **Security:** golang.org/x/crypto/bcrypt, github.com/golang-jwt/jwt/v5
* **Environment:** github.com/joho/godotenv

---

## ⚙️ Local Setup Instructions

Follow these steps to run the API on your local machine.

### 1. Clone the Repository

git clone https://github.com/YOUR_USERNAME/splitwise-clone.git
cd splitwise-clone
go mod tidy


### 2. Configure Environment Variables
Create a file named `.env` in the root directory of the project. It must contain the following keys:

DATABASE_URL=postgres://postgres:YOUR_POSTGRES_PASSWORD@localhost:5432/splitwise
JWT_SECRET=super_secret_key_change_this_in_production
PORT=8080

*(Note: Replace `YOUR_POSTGRES_PASSWORD` with your actual local PostgreSQL admin password).*

### 3. Database Setup (SQL)
Log into your local PostgreSQL instance (e.g., using `psql` or DBeaver) and run the following commands to create the database and initialize the required tables:

CREATE DATABASE splitwise;
\c splitwise;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone_number VARCHAR(20),
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE group_members (
    group_id INT REFERENCES groups(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE expenses (
    id SERIAL PRIMARY KEY,
    group_id INT REFERENCES groups(id) NULL, 
    payer_id INT REFERENCES users(id),
    description VARCHAR(255) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE expense_splits (
    id SERIAL PRIMARY KEY,
    expense_id INT REFERENCES expenses(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id),
    amount_owed DECIMAL(10,2) NOT NULL
);

CREATE TABLE friends (
    user_id_1 INT REFERENCES users(id),
    user_id_2 INT REFERENCES users(id),
    PRIMARY KEY (user_id_1, user_id_2)
);


### 4. Run the Server
Once the database is configured and the `.env` file is saved, start the application:

go run main.go

The server will start listening on `http://localhost:8080`.

---

## 📡 API Endpoints Reference

**Public Routes (No Auth Required)**
* `POST /api/users/register` - Create a new user account.
* `POST /api/users/login` - Authenticate and receive a JWT.

**Protected Routes (Requires `Authorization: Bearer <token>` Header)**
* `GET /api/users/me` - Fetch the logged-in user's profile and active group count.
* `POST /api/friends` - Add a friend.
* `POST /api/groups` - Create a new expense group.
* `POST /api/expenses` - Log a new expense (calculates splits and executes DB transaction).
* `GET /api/dashboard/balances` - View calculated debts (Who you owe & Who owes you).

---

## 🎥 Deliverables

* **Project Demo Video:** [https://drive.google.com/file/d/1fmUdgpTx49iaXjbKeiGjXZK_Qc0j4FjV/view?usp=sharing] (A full demonstration of the application explaining all the features via Insomnia).
