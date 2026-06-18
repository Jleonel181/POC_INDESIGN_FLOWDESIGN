-- ============================================
-- Database: designflow_db
-- Description: Schema for DesignFlow application
-- ============================================

-- Drop tables if exist (for development purposes)
DROP TABLE IF EXISTS pautas CASCADE;
DROP TABLE IF EXISTS pages CASCADE;
DROP TABLE IF EXISTS editions CASCADE;

-- ============================================
-- Table: editions
-- Description: Stores edition information
-- ============================================
CREATE TABLE editions (
    id SERIAL PRIMARY KEY,
    no_paginas INTEGER NOT NULL CHECK (no_paginas > 0),
    ancho_mm DECIMAL(10, 2) NOT NULL CHECK (ancho_mm > 0),
    alto_mm DECIMAL(10, 2) NOT NULL CHECK (alto_mm > 0),
    cuadros_ancho INTEGER NOT NULL CHECK (cuadros_ancho > 0),
    cuadros_alto INTEGER NOT NULL CHECK (cuadros_alto > 0),
    facing_pages BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- Table: pages
-- Description: Stores page information related to editions
-- ============================================
CREATE TABLE pages (
    id SERIAL PRIMARY KEY,
    no_pagina INTEGER NOT NULL CHECK (no_pagina > 0),
    edicion_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_pages_edition FOREIGN KEY (edicion_id) 
        REFERENCES editions(id) ON DELETE CASCADE,
    CONSTRAINT unique_page_per_edition UNIQUE (edicion_id, no_pagina)
);

-- ============================================
-- Table: pautas
-- Description: Stores pauta (layout guidelines) information
-- ============================================
CREATE TABLE pautas (
    id SERIAL PRIMARY KEY,
    descripcion_pauta VARCHAR(255) NOT NULL,
    cuadros_alto INTEGER NOT NULL CHECK (cuadros_alto >= 0),
    cuadros_ancho INTEGER NOT NULL CHECK (cuadros_ancho >= 0),
    ubicacion_cuadros_x INTEGER NOT NULL CHECK (ubicacion_cuadros_x >= 0),
    ubicacion_cuadros_y INTEGER NOT NULL CHECK (ubicacion_cuadros_y >= 0),
    pagina_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_pautas_page FOREIGN KEY (pagina_id) 
        REFERENCES pages(id) ON DELETE CASCADE
);

-- ============================================
-- Indexes for performance optimization
-- ============================================
CREATE INDEX idx_pages_edicion_id ON pages(edicion_id);
CREATE INDEX idx_pautas_pagina_id ON pautas(pagina_id);

-- ============================================
-- Seed data (for development/testing)
-- ============================================

-- Insert test edition
INSERT INTO editions (id, no_paginas, ancho_mm, alto_mm, cuadros_ancho, cuadros_alto, facing_pages, margen_superior_mm, margen_inferior_mm, margen_izquierdo_mm, margen_derecho_mm) 
VALUES (1, 10, 210.0, 297.0, 10, 15, false, 0, 0, 0, 0);

-- Insert test pages
INSERT INTO pages (id, no_pagina, edicion_id) VALUES
(1, 1, 1),
(2, 2, 1),
(3, 3, 1);

-- Insert test pautas
INSERT INTO pautas (id, descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
(1, 'Pauta 1', 2, 5, 0, 0, 1),
(2, 'Pauta 2', 1, 3, 0, 2, 1),
(3, 'Pauta 3', 1, 3, 0, 7, 1);

-- Reset sequences
SELECT setval('editions_id_seq', (SELECT MAX(id) FROM editions));
SELECT setval('pages_id_seq', (SELECT MAX(id) FROM pages));
SELECT setval('pautas_id_seq', (SELECT MAX(id) FROM pautas));

-- ============================================
-- Grant permissions (optional, for specific user)
-- ============================================
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO designflow;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO designflow;
