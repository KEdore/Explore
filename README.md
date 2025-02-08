# Explore Service

This is a gRPC-based service that manages user interactions (likes/passes) in a dating application. The service provides endpoints for recording decisions, listing users who liked a particular user, and counting likes.

## Architecture Decisions

1. **Database Choice**: MySQL was chosen for this implementation because:
   - It provides ACID compliance which is crucial for maintaining consistency in user interactions
   - Supports efficient indexing for our query patterns
   - Handles high write throughput well with proper configuration
   - Supports transaction isolation levels needed for concurrent operations

2. **Scalability Considerations**:
   - Implemented pagination for listing endpoints to handle users with many likes
   - Used database connection pooling to manage resources efficiently
   - Indexes on (recipient_id, liked) and (actor_id, recipient_id) for efficient queries
   - Timestamp-based pagination for consistent results with new data

3. **Performance Optimizations**:
   - Used prepared statements to reduce query parsing overhead
   - Implemented efficient pagination using timestamps
   - Limited result sets to 50 records per page
   - Used EXISTS clause for mutual like checking instead of JOIN where appropriate

## Database Schema

```sql
CREATE TABLE decisions (
    actor_id VARCHAR(255) NOT NULL,
    recipient_id VARCHAR(255) NOT NULL,
    liked BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (actor_id, recipient_id),
    INDEX idx_recipient_liked (recipient_id, liked),
    INDEX idx_created_at (created_at)
);
```

## Running the Service

There are two primary ways to run the service: using plain Docker commands or with Docker Compose.

1. Build and run using Docker:
```bash
docker build -t explore .
docker run -e DB_USER=myuser \
           -e DB_PASS=mypass \
           -e DB_NAME=mydb \
           -e DB_HOST=mydbhost:3306 \
           -e SERVER_ADDRESS=":50051" \
           -p 50051:50051 \
           explore

```

- Note: The service listens on port 50051 by default. Adjust the -p flag as needed.
- If no DB_HOST is specified, the application defaults to localhost:3306.


2. Build and run using Docker compose:
```bash
docker-compose up --build
```

This command will:

Build the app image using the Dockerfile.
Start the MySQL container.
Start the Explore Service container, which connects to MySQL using the hostname mysql on the Docker network.

Accessing the Service:

The gRPC service is exposed on container port 50051 and mapped to host port 9090.
You can interact with the service via localhost:9090.


### Required Environment Variables
- DB_USER: MySQL username
- DB_PASS: MySQL password
- DB_NAME: MySQL database name
- DB_HOST: MySQL host (defaults to localhost:3306 if not set; in Docker Compose, this is set to mysql)
- SERVER_ADDRESS: The address the gRPC server listens on (defaults to :50051)


## Testing

Run the tests using:
```bash
go test -v ./...
```

or ```bash
make test
```

## Assumptions

1. User IDs are strings and are already validated upstream
2. A decision can be overwritten at any time
3. Pagination: An offset-based pagination token (numeric, represented as a string) is used. Although sufficient for moderate volumes, a cursor-based approach could be explored for extreme scale.
4. Database Availability: The service assumes a valid database connection and handles transient errors via retries at the database driver level.
5. No Decision Deletion: The service does not implement deletion of decisions as it is not required by the current specifications.

## Future Improvements

Caching Layer:
Add a caching layer (e.g., Redis) for frequently accessed data.

Rate Limiting:
Implement rate limiting to prevent abuse.

Batch Processing:
For very high volumes, add batch processing mechanisms to update or query decisions.

Input Validation Middleware:
Add middleware to validate incoming requests more robustly.

Enhanced Observability:
Integrate logging, tracing, and metrics (e.g., via Prometheus) to monitor performance and issues.

Pagination Improvements:
Consider moving to cursor-based pagination for better performance on very large datasets