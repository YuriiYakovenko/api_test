//create table
CREATE TABLE users (
   User_id uuid DEFAULT uuid_generate_v4 (),  
   Email CHAR(256) NOT NULL,  
   Fullname CHAR(256) NOT NULL,  
   Password CHAR(256) NOT NULL, 
   CreatedAt TIMESTAMPTZ DEFAULT Now(),
   PRIMARY KEY (User_id));