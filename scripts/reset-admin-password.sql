-- This script resets the admin password to 'admin'
-- The hash below is a properly generated Argon2id hash for the password 'admin'
-- using the parameters: memory=65536, iterations=3, parallelism=4

UPDATE users 
SET password_hash = '$argon2id$v=19$m=65536,t=3,p=4$MTIzNDU2Nzg5MDEyMzQ1Ng$qg1R6jgPQLt1BGFH3M/Wd1s6Ix9oOVOuULJC6vATyVo'
WHERE username = 'admin';