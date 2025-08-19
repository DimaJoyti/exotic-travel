-- Create destinations table
CREATE TABLE IF NOT EXISTS destinations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    country VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    duration INTEGER NOT NULL CHECK (duration > 0),
    max_guests INTEGER NOT NULL CHECK (max_guests > 0),
    images TEXT[] DEFAULT '{}',
    features TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_destinations_country ON destinations(country);
CREATE INDEX IF NOT EXISTS idx_destinations_city ON destinations(city);
CREATE INDEX IF NOT EXISTS idx_destinations_price ON destinations(price);
CREATE INDEX IF NOT EXISTS idx_destinations_duration ON destinations(duration);
CREATE INDEX IF NOT EXISTS idx_destinations_max_guests ON destinations(max_guests);

-- Create full-text search index for name and description
CREATE INDEX IF NOT EXISTS idx_destinations_search ON destinations 
    USING gin(to_tsvector('english', name || ' ' || description));

-- Create trigger for destinations table
CREATE TRIGGER update_destinations_updated_at 
    BEFORE UPDATE ON destinations 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
