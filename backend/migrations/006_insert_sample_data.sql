-- Insert sample admin user (password: admin123)
INSERT INTO users (email, password_hash, first_name, last_name, role) VALUES
('admin@exotic-travel.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'User', 'admin')
ON CONFLICT (email) DO NOTHING;

-- Insert sample regular user (password: user123)
INSERT INTO users (email, password_hash, first_name, last_name, role) VALUES
('user@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'John', 'Doe', 'user')
ON CONFLICT (email) DO NOTHING;

-- Insert sample destinations
INSERT INTO destinations (name, description, country, city, price, duration, max_guests, images, features) VALUES
(
    'Maldives Paradise Resort',
    'Experience luxury in overwater bungalows surrounded by crystal-clear turquoise waters. This exclusive resort offers world-class diving, spa treatments, and gourmet dining with stunning ocean views.',
    'Maldives',
    'Malé',
    2500.00,
    7,
    4,
    ARRAY['https://images.unsplash.com/photo-1506905925346-21bda4d32df4', 'https://images.unsplash.com/photo-1540979388789-6cee28a1cdc9'],
    ARRAY['Overwater Bungalows', 'Private Beach', 'Spa & Wellness', 'Scuba Diving', 'Fine Dining', 'Airport Transfer']
),
(
    'Amazon Rainforest Adventure',
    'Embark on an unforgettable journey deep into the Amazon rainforest. Stay in eco-lodges, spot exotic wildlife, and learn about indigenous cultures while contributing to conservation efforts.',
    'Brazil',
    'Manaus',
    1800.00,
    10,
    8,
    ARRAY['https://images.unsplash.com/photo-1544735716-392fe2489ffa', 'https://images.unsplash.com/photo-1516026672322-bc52d61a55d5'],
    ARRAY['Eco Lodge', 'Wildlife Spotting', 'Canoe Expeditions', 'Indigenous Culture', 'Conservation Program', 'Expert Guides']
),
(
    'Sahara Desert Glamping',
    'Sleep under the stars in luxury desert camps while exploring the vast Sahara. Enjoy camel trekking, traditional Berber cuisine, and breathtaking sunrises over endless sand dunes.',
    'Morocco',
    'Merzouga',
    1200.00,
    5,
    6,
    ARRAY['https://images.unsplash.com/photo-1509316975850-ff9c5deb0cd9', 'https://images.unsplash.com/photo-1518548419970-58e3b4079ab2'],
    ARRAY['Luxury Tents', 'Camel Trekking', 'Stargazing', 'Traditional Cuisine', 'Berber Culture', 'Sandboarding']
),
(
    'Antarctic Expedition Cruise',
    'Journey to the last frontier on Earth aboard a luxury expedition vessel. Witness massive icebergs, encounter penguins and whales, and explore the pristine Antarctic wilderness.',
    'Antarctica',
    'Antarctic Peninsula',
    8500.00,
    14,
    12,
    ARRAY['https://images.unsplash.com/photo-1518837695005-2083093ee35b', 'https://images.unsplash.com/photo-1544966503-7cc5ac882d5f'],
    ARRAY['Expedition Cruise', 'Wildlife Viewing', 'Zodiac Landings', 'Expert Naturalists', 'Photography Workshops', 'All Meals Included']
);

-- Insert sample reviews
INSERT INTO reviews (user_id, destination_id, rating, comment) VALUES
(2, 1, 5, 'Absolutely incredible experience! The overwater bungalow was perfect and the staff was amazing.'),
(2, 2, 4, 'Great adventure, saw so many animals! The guides were very knowledgeable about the rainforest.');
