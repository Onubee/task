CREATE TABLE IF NOT EXISTS brands (
                                      id SERIAL PRIMARY KEY,
                                      name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS categories (
                                          id SERIAL PRIMARY KEY,
                                          name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
                                        id INTEGER PRIMARY KEY,
                                        name TEXT NOT NULL,
                                        brand_id INTEGER REFERENCES brands(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    price NUMERIC(12,2) NOT NULL,
    stock INTEGER NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (name, brand_id)
    );

CREATE TABLE IF NOT EXISTS clients (
                                       id INTEGER PRIMARY KEY,
                                       first_name TEXT NOT NULL,
                                       last_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS client_products (
                                               client_id INTEGER REFERENCES clients(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    PRIMARY KEY (client_id, product_id)
    );

CREATE TABLE IF NOT EXISTS tasks (
                                     id SERIAL PRIMARY KEY,
                                     started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                     finished_at TIMESTAMP,
                                     status TEXT NOT NULL CHECK (status IN ('in_progress', 'completed', 'stopped', 'cancelled', 'completed_with_errors'))
    );

CREATE INDEX idx_products_brand ON products(brand_id);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_client_products_client ON client_products(client_id);
CREATE INDEX idx_tasks_status ON tasks(status);