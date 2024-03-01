
-- Drop the table if it already exists
DROP TABLE IF EXISTS employee;
DROP TABLE IF EXISTS department;

-- Create the table with extended data types
CREATE TABLE employee (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    preferences JSON,
    salary DECIMAL(10, 2),
    birth_date DATE,
    hire_datetime DATETIME,
    part_time BOOLEAN,
    biography TEXT,
    profile_picture BLOB,
    last_login TIMESTAMP,
    vacation_days SMALLINT,
    contract_length MEDIUMINT,
    employee_number BIGINT,
    office_location POINT,
    working_hours TIME,
    yearly_bonus DOUBLE,
    employee_code CHAR(10),
    health_status ENUM('excellent', 'good', 'fair', 'poor'),
    security_level TINYINT,
    resume MEDIUMBLOB,
    linkedin_url VARCHAR(255),
    personal_website VARCHAR(255),
    notes LONGTEXT,
    department_id TINYINT,
    fingerprint VARBINARY(128),
    schedule SET('morning', 'afternoon', 'night'),
    last_performance_review YEAR,
    nationality CHAR(3),
    languages JSON,
    hire_date_year YEAR(4)
);

CREATE TABLE department (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255)
);

-- Insert data into the table
INSERT INTO employee (id, name, email, preferences, salary, birth_date, hire_datetime, part_time, biography, last_login, vacation_days, contract_length, employee_number, working_hours, yearly_bonus, employee_code, health_status, security_level, linkedin_url, personal_website, notes, department_id, last_performance_review, nationality, languages, hire_date_year, office_location) VALUES
(1, 'John', 'john@example.com', '{"theme": "dark", "notifications": true}', 50000.00, '1980-01-01', '2020-01-01 08:30:00', true, 'John has been a part of our company for over a decade...', '2023-01-01 12:00:00', 10, 12, 100001, '09:00:00', 3000.00, 'EMP00001', 'good', 3, 'https://linkedin.com/in/john', 'https://johnsblog.com', 'John has consistently performed well.', 1, 2023, 'USA', '{"English": "fluent", "Spanish": "intermediate"}', 2020, POINT(-74.0060, 40.7128)),
(2, 'Adam', 'adam@example.com', '{"theme": "dark", "notifications": true}', 52000.00, '1982-05-12', '2020-03-15 09:00:00', false, 'Adam is a recent addition to the team...', '2023-02-02 14:30:00', 15, 24, 100002, '10:00:00', 2500.00, 'EMP00002', 'excellent', 4, 'https://linkedin.com/in/adam', 'https://adamportfolio.com', 'Adam brings fresh perspectives.', 2, 2023, 'CAN', '{"French": "fluent", "English": "fluent"}', 2020, POINT(-4.0060, 12.7128)),
(16, 'Diana', 'diana@example.com', '{"theme": "dark", "notifications": true}', 55000.00, '1990-04-05', '2021-04-15 09:30:00', false, 'Diana is known for her attention to detail...', '2024-01-20 10:00:00', 20, 36, 100016, '08:00:00', 4500.00, 'EMP00016', 'excellent', 5, 'https://linkedin.com/in/diana', 'https://dianasportfolio.com', 'Diana has led several successful projects.', 3, 2024, 'GBR', '{"English": "fluent", "French": "basic"}', 2021, POINT(14.0060, -80.7128));
