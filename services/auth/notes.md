goose -dir internal/adapters/db/migrations create create_users_table sql

goose -dir internal/adapters/db/migrations up

goose -dir internal/adapters/db/migrations status

goose -dir internal/adapters/db/migrations down

openssl pkey -in keys/access_private.pem -pubout -out keys/access_public.pem

cd services/auth && openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out keys/access_private.pem

mailpit mail url http://localhost:8025
