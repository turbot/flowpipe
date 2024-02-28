
-- Drop the table if it already exists
DROP TABLE IF EXISTS employee;

-- Create the table with extended data types for PostgreSQL
CREATE TABLE employee (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    preferences JSON,
    salary NUMERIC(10, 2),
    birth_date DATE,
    -- NOTE: Problem with hire_datetime, it is not accepting the time zone
    hire_datetime TIMESTAMP WITH TIME ZONE,
    part_time BOOLEAN,
    biography TEXT,
    profile_picture BYTEA,
    last_login TIMESTAMP,
    vacation_days SMALLINT,
    contract_length INTEGER,
    employee_number BIGINT,
    office_location POINT,
    working_hours INTERVAL,
    yearly_bonus DOUBLE PRECISION,
    employee_code CHAR(10),
    health_status TEXT CHECK (health_status IN ('excellent', 'good', 'fair', 'poor')),
    security_level SMALLINT,
    resume BYTEA,
    linkedin_url TEXT,
    personal_website TEXT,
    notes TEXT,
    department_id SMALLINT,
    fingerprint BYTEA,
    schedule TEXT[] NOT NULL,
    last_performance_review DATE,
    nationality CHAR(3),
    languages JSONB,
    hire_date_year INTEGER
);

-- Insert data into the table
INSERT INTO employee (name, email, preferences, salary, birth_date, hire_datetime, part_time, biography, last_login, vacation_days, contract_length, employee_number, working_hours, yearly_bonus, employee_code, health_status, security_level, linkedin_url, personal_website, notes, department_id, last_performance_review, nationality, languages, hire_date_year, schedule) VALUES
('John', 'john@example.com', '{"theme": "dark", "notifications": true}', 50000.00, '1980-01-01', '2020-01-01 08:30:00+00', true, 'John has been a part of our company for over a decade...', '2023-01-01 12:00:00', 10, 12, 100001, '8 hours', 3000.00, 'EMP00001', 'good', 3, 'https://linkedin.com/in/john', 'https://johnsblog.com', 'John has consistently performed well.', 1, '2023-01-01', 'USA', '{"English": "fluent", "Spanish": "intermediate"}', 2020, ARRAY['morning', 'afternoon']),
('Adam', 'adam@example.com', '{"theme": "dark", "notifications": true}', 52000.00, '1982-05-12', '2020-03-15 09:00:00+00', false, 'Adam is a recent addition to the team...', '2023-02-02 14:30:00', 15, 24, 100002, '9 hours', 2500.00, 'EMP00002', 'excellent', 4, 'https://linkedin.com/in/adam', 'https://adamportfolio.com', 'Adam brings fresh perspectives.', 2, '2023-02-02', 'CAN', '{"French": "fluent", "English": "fluent"}', 2020, ARRAY['morning', 'night']),
('Diana', 'diana@example.com', '{"theme": "dark", "notifications": true}', 55000.00, '1990-04-05', '2021-04-15 09:30:00+00', false, 'Diana is known for her attention to detail...', '2024-01-20 10:00:00', 20, 36, 100016, '8 hours', 4500.00, 'EMP00016', 'excellent', 5, 'https://linkedin.com/in/diana', 'https://dianasportfolio.com', 'Diana has led several successful projects.', 3, '2024-01-20', 'GBR', '{"English": "fluent", "French": "basic"}', 2021, ARRAY['afternoon', 'night']);
