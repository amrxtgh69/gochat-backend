# gochat-backend

## Setup Backend

1. clone the repo and cd into it
2. run ```go mod tidy```
3. download mysql and setup
4. make a mysql user with username root and password root
5. login to the mysql shell with the password you have created ```mysql -u root -p```
6. If database is not created already, copy and paste this command below:
```
CREATE DATABASE myapp;
USE myapp;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    full_name VARCHAR(100),
    user_name VARCHAR(50) UNIQUE,
    password VARCHAR(255)
);
```
7. exit the mysql shell
6. run the command ```go run cmd/gochat/main.go```
7. open a new terminal and use the GoChat client
