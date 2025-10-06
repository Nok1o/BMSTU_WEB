package test

const TestDbConnStringEnvVar = "DATABASE_TEST_URL"
const TestRedisConnStringEnvVar = "REDIS_TEST_URL"
const TestMinioConnStringEnvVar = "MINIO_TEST_URL"

const DefaultDatabaseTestUrl = "postgres://quickflow_admin:SuperSecurePassword1@localhost:5432/quickflow_db?sslmode=disable"
