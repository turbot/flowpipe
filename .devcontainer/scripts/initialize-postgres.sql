-- Create the 'flowpipe' user with a password
CREATE USER flowpipe WITH PASSWORD 'password';

-- Create the 'flowpipe-test' database owned by 'flowpipe'
CREATE DATABASE "flowpipe-test" OWNER flowpipe;

-- Grant all privileges on the database 'flowpipe-test' to the user 'flowpipe'
GRANT ALL PRIVILEGES ON DATABASE "flowpipe-test" TO flowpipe;
