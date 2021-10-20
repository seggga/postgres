CREATE USER gotuber
WITH PASSWORD 'Passw0rd';

CREATE DATABASE go_tube
    WITH OWNER gotuber
    TEMPLATE = 'template0'
    ENCODING = 'utf-8'
    LC_COLLATE = 'C.UTF-8'
    LC_CTYPE = 'C.UTF-8';