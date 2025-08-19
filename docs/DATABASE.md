# Database Documentation

This document describes the database schema, relationships, and design decisions for the Exotic Travel Booking Platform.

## Database Technology

- **Database**: PostgreSQL 15+
- **Connection Pooling**: Custom Go connection pool with optimization
- **Migrations**: Go-based migration system
- **Indexing**: Optimized indexes for performance
- **Backup**: Automated backup and recovery procedures

## Schema Overview

The database follows a normalized design with clear relationships between entities. Here's the entity relationship diagram:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    Users    │    │Destinations │    │   Bookings  │
│             │    │             │    │             │
│ id (PK)     │    │ id (PK)     │    │ id (PK)     │
│ name        │    │ name        │    │ user_id (FK)│
│ email       │    │ description │    │ dest_id (FK)│
│ password    │    │ country     │    │ start_date  │
│ role        │    │ city        │    │ end_date    │
│ created_at  │    │ price       │    │ guests      │
│ updated_at  │    │ duration    │    │ total_price │
└─────────────┘    │ max_guests  │    │ status      │
                   │ images      │    │ created_at  │
                   │ features    │    └─────────────┘
                   │ created_at  │           │
                   │ updated_at  │           │
                   └─────────────┘           │
                          │                 │
                          └─────────────────┘
                                  │
                          ┌─────────────┐
                          │   Reviews   │
                          │             │
                          │ id (PK)     │
                          │ user_id (FK)│
                          │ dest_id (FK)│
                          │ booking_id  │
                          │ rating      │
                          │ comment     │
                          │ created_at  │
                          └─────────────┘
```

## Table Definitions

### Users Table

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    phone VARCHAR(20),
    role VARCHAR(50) DEFAULT 'user' NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at);
```

**Fields:**
- `id`: Primary key, auto-incrementing
- `name`: User's full name
- `email`: Unique email address for authentication
- `password_hash`: Argon2id hashed password
- `phone`: Optional phone number
- `role`: User role (user, admin, moderator)
- `email_verified`: Email verification status
- `created_at`: Account creation timestamp
- `updated_at`: Last update timestamp

### Destinations Table

```sql
CREATE TABLE destinations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    country VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    duration INTEGER NOT NULL, -- days
    max_guests INTEGER NOT NULL,
    images TEXT[] DEFAULT '{}',
    features TEXT[] DEFAULT '{}',
    rating DECIMAL(3,2) DEFAULT 0.0,
    review_count INTEGER DEFAULT 0,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_destinations_country ON destinations(country);
CREATE INDEX idx_destinations_city ON destinations(city);
CREATE INDEX idx_destinations_price ON destinations(price);
CREATE INDEX idx_destinations_rating ON destinations(rating);
CREATE INDEX idx_destinations_active ON destinations(active);
CREATE INDEX idx_destinations_created_at ON destinations(created_at);

-- Full-text search index
CREATE INDEX idx_destinations_search ON destinations 
USING gin(to_tsvector('english', name || ' ' || description));
```

**Fields:**
- `id`: Primary key, auto-incrementing
- `name`: Destination name
- `description`: Detailed description
- `country`: Country name
- `city`: City name
- `price`: Price per person in USD
- `duration`: Trip duration in days
- `max_guests`: Maximum number of guests
- `images`: Array of image URLs
- `features`: Array of destination features
- `rating`: Average rating (calculated)
- `review_count`: Number of reviews (calculated)
- `active`: Whether destination is available for booking

### Bookings Table

```sql
CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    destination_id BIGINT NOT NULL REFERENCES destinations(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    guests INTEGER NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' NOT NULL,
    special_requests TEXT,
    payment_intent_id VARCHAR(255),
    payment_status VARCHAR(50) DEFAULT 'pending',
    cancellation_reason TEXT,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT chk_booking_dates CHECK (end_date > start_date),
    CONSTRAINT chk_booking_guests CHECK (guests > 0),
    CONSTRAINT chk_booking_price CHECK (total_price > 0)
);

-- Indexes
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_destination_id ON bookings(destination_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_dates ON bookings(start_date, end_date);
CREATE INDEX idx_bookings_created_at ON bookings(created_at);

-- Unique constraint to prevent double booking
CREATE UNIQUE INDEX idx_bookings_no_overlap ON bookings(destination_id, start_date, end_date)
WHERE status IN ('confirmed', 'pending');
```

**Fields:**
- `id`: Primary key, auto-incrementing
- `user_id`: Foreign key to users table
- `destination_id`: Foreign key to destinations table
- `start_date`: Booking start date
- `end_date`: Booking end date
- `guests`: Number of guests
- `total_price`: Total booking price
- `status`: Booking status (pending, confirmed, cancelled, completed)
- `special_requests`: Optional special requests
- `payment_intent_id`: Stripe payment intent ID
- `payment_status`: Payment status (pending, paid, failed, refunded)
- `cancellation_reason`: Reason for cancellation
- `cancelled_at`: Cancellation timestamp

### Reviews Table

```sql
CREATE TABLE reviews (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    destination_id BIGINT NOT NULL REFERENCES destinations(id) ON DELETE CASCADE,
    booking_id BIGINT REFERENCES bookings(id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    helpful_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, destination_id, booking_id)
);

-- Indexes
CREATE INDEX idx_reviews_user_id ON reviews(user_id);
CREATE INDEX idx_reviews_destination_id ON reviews(destination_id);
CREATE INDEX idx_reviews_rating ON reviews(rating);
CREATE INDEX idx_reviews_created_at ON reviews(created_at);
```

**Fields:**
- `id`: Primary key, auto-incrementing
- `user_id`: Foreign key to users table
- `destination_id`: Foreign key to destinations table
- `booking_id`: Optional foreign key to bookings table
- `rating`: Rating from 1 to 5 stars
- `comment`: Optional review comment
- `helpful_count`: Number of helpful votes
- `created_at`: Review creation timestamp
- `updated_at`: Last update timestamp

### Audit Log Table

```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id BIGINT,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

## Database Functions and Triggers

### Update Rating Function

```sql
CREATE OR REPLACE FUNCTION update_destination_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE destinations 
    SET 
        rating = (
            SELECT COALESCE(AVG(rating), 0) 
            FROM reviews 
            WHERE destination_id = COALESCE(NEW.destination_id, OLD.destination_id)
        ),
        review_count = (
            SELECT COUNT(*) 
            FROM reviews 
            WHERE destination_id = COALESCE(NEW.destination_id, OLD.destination_id)
        ),
        updated_at = NOW()
    WHERE id = COALESCE(NEW.destination_id, OLD.destination_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger to update ratings when reviews change
CREATE TRIGGER trigger_update_destination_rating
    AFTER INSERT OR UPDATE OR DELETE ON reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_destination_rating();
```

### Updated At Trigger

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables with updated_at column
CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_destinations_updated_at
    BEFORE UPDATE ON destinations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_bookings_updated_at
    BEFORE UPDATE ON bookings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_reviews_updated_at
    BEFORE UPDATE ON reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

## Performance Optimizations

### Connection Pooling
- **Max Connections**: 25
- **Idle Connections**: 5
- **Connection Lifetime**: 1 hour
- **Connection Idle Time**: 15 minutes

### Query Optimization
- **Prepared Statements**: Cached prepared statements for repeated queries
- **Batch Operations**: Optimized batch inserts and updates
- **Index Usage**: Strategic indexes for common query patterns
- **Query Analysis**: EXPLAIN ANALYZE for slow query identification

### Caching Strategy
- **Redis Integration**: Frequently accessed data cached in Redis
- **Cache TTL**: Configurable cache expiration times
- **Cache Invalidation**: Automatic cache invalidation on data changes

## Backup and Recovery

### Automated Backups
```bash
# Daily full backup
pg_dump -h localhost -U postgres exotic_travel_prod > backup_$(date +%Y%m%d).sql

# Point-in-time recovery setup
# Enable WAL archiving in postgresql.conf
archive_mode = on
archive_command = 'cp %p /backup/wal/%f'
```

### Recovery Procedures
```bash
# Restore from backup
psql -h localhost -U postgres -d exotic_travel_prod < backup_20240101.sql

# Point-in-time recovery
pg_basebackup -h localhost -D /backup/base -U postgres -v -P -W
```

## Migration System

### Migration Files
```
migrations/
├── 001_create_users_table.up.sql
├── 001_create_users_table.down.sql
├── 002_create_destinations_table.up.sql
├── 002_create_destinations_table.down.sql
└── ...
```

### Migration Commands
```bash
# Run migrations
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down

# Check migration status
go run cmd/migrate/main.go status
```

## Security Considerations

### Data Protection
- **Password Hashing**: Argon2id with secure parameters
- **Sensitive Data**: Encrypted at rest and in transit
- **Access Control**: Row-level security where applicable
- **Audit Trail**: Comprehensive audit logging

### Database Security
- **Connection Encryption**: SSL/TLS for all connections
- **User Permissions**: Principle of least privilege
- **Network Security**: Database firewall and VPN access
- **Regular Updates**: Automated security patch management

## Monitoring and Maintenance

### Performance Monitoring
- **Query Performance**: Slow query logging and analysis
- **Connection Monitoring**: Connection pool metrics
- **Resource Usage**: CPU, memory, and disk monitoring
- **Index Usage**: Index effectiveness analysis

### Maintenance Tasks
- **VACUUM**: Regular table maintenance
- **ANALYZE**: Statistics updates for query planner
- **REINDEX**: Index rebuilding when necessary
- **Log Rotation**: Automated log file management
