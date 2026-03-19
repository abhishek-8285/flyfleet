GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO fleet_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO fleet_user;
DELETE FROM users WHERE email='admin@flyfleet.com';
INSERT INTO users (id, email, password_hash, role) VALUES (gen_random_uuid(), 'admin@flyfleet.com', '$argon2id$v=19$m=65536,t=3,p=4$yLdGBVfwF8NQ+5Z70RPokQ$cBQHQPAo3Qz1RnqNQZ+woumaJU5JElTnTqrlCwr4nAc', 'admin');
