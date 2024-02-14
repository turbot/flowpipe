-- Create a new database and user for the Flowpipe test suite
CREATE USER IF NOT EXISTS 'flowpipe'@'localhost' IDENTIFIED BY 'password';

-- Create the 'flowpipe-test' database owned by 'flowpipe'
CREATE DATABASE IF NOT EXISTS `flowpipe-test`;

-- Grant the user all privileges on the new database
GRANT ALL PRIVILEGES ON `flowpipe-test`.* TO 'flowpipe'@'localhost';
FLUSH PRIVILEGES;
