DROP TABLE IF EXISTS employee;

CREATE TABLE employee (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    preferences JSON
);

INSERT INTO employee (id, name, email, preferences) VALUES
(1, 'John', 'john@example.com', '{"theme": "dark", "notifications": true}'),
(2, 'Adam', 'adam@example.com', '{"theme": "dark", "notifications": true}'),
(3, 'Alice', 'alice@example.com', '{"theme": "dark", "notifications": false}'),
(4, 'Bob', 'bob@example.com', '{"theme": "light", "notifications": false}'),
(5, 'Alex', 'alex@example.com', '{"theme": "dark", "notifications": true}'),
(6, 'Carey', 'carey@example.com', '{"theme": "light", "notifications": false}'),
(7, 'Cody', 'cody@example.com', '{"theme": "light", "notifications": false}'),
(8, 'Andrew', 'andrew@example.com', '{"theme": "dark", "notifications": true}'),
(9, 'Alexandra', 'alexandra@example.com', '{"theme": "light", "notifications": true}'),
(10, 'Jon', 'jon@example.com', '{"theme": "dark", "notifications": true}'),
(11, 'Jennifer', 'jennifer@example.com', '{"theme": "light", "notifications": false}'),
(12, 'Alan', 'alan@example.com', '{"theme": "dark", "notifications": true}'),
(13, 'Mia', 'mia@example.com', '{"theme": "light", "notifications": true}'),
(14, 'Aaron', 'aaron@example.com', '{"theme": "light", "notifications": true}'),
(15, 'Adrian', 'adrian@example.com', '{"theme": "dark", "notifications": true}');
