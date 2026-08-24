-- ============================================
-- Database: designflow_db
-- Description: Schema para la aplicación DesignFlow
-- ============================================

-- Drop tables si existen (orden inverso por dependencias)
DROP TABLE IF EXISTS pautas CASCADE;
DROP TABLE IF EXISTS pages CASCADE;
DROP TABLE IF EXISTS editions CASCADE;

-- ============================================
-- Table: editions
-- ============================================
CREATE TABLE editions (
    id SERIAL PRIMARY KEY,
    no_paginas INTEGER NOT NULL CHECK (no_paginas > 0),
    ancho_mm DECIMAL(10, 2) NOT NULL CHECK (ancho_mm > 0),
    alto_mm DECIMAL(10, 2) NOT NULL CHECK (alto_mm > 0),
    cuadros_ancho INTEGER NOT NULL CHECK (cuadros_ancho > 0),
    cuadros_alto INTEGER NOT NULL CHECK (cuadros_alto > 0),
    facing_pages BOOLEAN NOT NULL DEFAULT false,
    margen_superior_mm DECIMAL(10, 2) NOT NULL DEFAULT 0,
    margen_inferior_mm DECIMAL(10, 2) NOT NULL DEFAULT 0,
    margen_izquierdo_mm DECIMAL(10, 2) NOT NULL DEFAULT 0,
    margen_derecho_mm DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- Table: pages
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
-- ============================================
CREATE TABLE pautas (
    id SERIAL PRIMARY KEY,
    descripcion_pauta VARCHAR(255) NOT NULL,
    cuadros_alto INTEGER NOT NULL CHECK (cuadros_alto > 0),
    cuadros_ancho INTEGER NOT NULL CHECK (cuadros_ancho > 0),
    ubicacion_cuadros_x INTEGER DEFAULT NULL,
    ubicacion_cuadros_y INTEGER DEFAULT NULL,
    pagina_id INTEGER DEFAULT NULL,
    content_type VARCHAR(10) NOT NULL DEFAULT 'text',
    image_base64 TEXT DEFAULT NULL,
    cover_date DATE DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_pautas_page FOREIGN KEY (pagina_id)
        REFERENCES pages(id) ON DELETE SET NULL
);

-- ============================================
-- Índices
-- ============================================
CREATE INDEX idx_pages_edicion_id ON pages(edicion_id);
CREATE INDEX idx_pautas_pagina_id ON pautas(pagina_id);

-- ============================================
-- Datos de prueba
-- ============================================

-- Edición 1: Periódico tabloide (265×370mm), grilla 5×8, sin facing pages
INSERT INTO editions (id, no_paginas, ancho_mm, alto_mm, cuadros_ancho, cuadros_alto, facing_pages, margen_superior_mm, margen_inferior_mm, margen_izquierdo_mm, margen_derecho_mm)
VALUES (1, 4, 265.0, 370.0, 5, 8, false, 10.0, 10.0, 10.0, 10.0);

INSERT INTO pages (id, no_pagina, edicion_id) VALUES
(1, 1, 1), (2, 2, 1), (3, 3, 1), (4, 4, 1);

-- Página 1: Portada
INSERT INTO pautas (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
('Cabezal',           1, 5, 0, 0, 1),
('Foto principal',    4, 3, 0, 1, 1),
('Llamada lateral',   2, 2, 3, 1, 1),
('Avance deportes',   2, 2, 3, 3, 1),
('Cintillo inferior', 1, 5, 0, 7, 1);

-- Página 2: Noticias locales
INSERT INTO pautas (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
('Nota principal',       4, 3, 0, 0, 2),
('Foto nota principal',  2, 2, 3, 0, 2),
('Breves',               2, 2, 3, 2, 2),
('Publicidad media',     2, 5, 0, 6, 2);

-- Página 3: Deportes
INSERT INTO pautas (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
('Encabezado deportes',  1, 5, 0, 0, 3),
('Nota futbol',          3, 3, 0, 1, 3),
('Tabla posiciones',     3, 2, 3, 1, 3),
('Nota secundaria',      2, 3, 0, 4, 3),
('Publicidad esquina',   2, 2, 3, 4, 3),
('Resultados',           2, 5, 0, 6, 3);

-- Página 4: Contraportada
INSERT INTO pautas (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
('Publicidad página completa', 8, 5, 0, 0, 4);

-- Edición 2: Revista A4 (210×297mm), grilla 4×6, con facing pages
INSERT INTO editions (id, no_paginas, ancho_mm, alto_mm, cuadros_ancho, cuadros_alto, facing_pages, margen_superior_mm, margen_inferior_mm, margen_izquierdo_mm, margen_derecho_mm)
VALUES (2, 8, 210.0, 297.0, 4, 6, true, 15.0, 15.0, 20.0, 15.0);

INSERT INTO pages (id, no_pagina, edicion_id) VALUES
(5, 1, 2), (6, 2, 2), (7, 3, 2), (8, 4, 2),
(9, 5, 2), (10, 6, 2), (11, 7, 2), (12, 8, 2);

-- Página 1: Portada revista
INSERT INTO pautas (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
('Logo',             1, 4, 0, 0, 5),
('Imagen portada',   4, 4, 0, 1, 5),
('Titulo principal', 1, 4, 0, 5, 5);

-- Página 2: Editorial
INSERT INTO pautas (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
('Titulo editorial', 1, 4, 0, 0, 6),
('Texto editorial',  4, 3, 0, 1, 6),
('Foto editor',      2, 1, 3, 1, 6),
('Staff',            1, 4, 0, 5, 6);

-- Página 3: Reportaje
INSERT INTO pautas (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
('Titular reportaje',  1, 4, 0, 0, 7),
('Foto reportaje',     3, 2, 0, 1, 7),
('Texto reportaje',    3, 2, 2, 1, 7),
('Pie de foto',        1, 2, 0, 4, 7),
('Continuación texto', 1, 2, 2, 4, 7);

-- Edición 3: Periódico broadsheet con facing (380×560mm), grilla 6×10
INSERT INTO editions (id, no_paginas, ancho_mm, alto_mm, cuadros_ancho, cuadros_alto, facing_pages, margen_superior_mm, margen_inferior_mm, margen_izquierdo_mm, margen_derecho_mm)
VALUES (3, 2, 380.0, 560.0, 6, 10, true, 12.0, 12.0, 15.0, 15.0);

INSERT INTO pages (id, no_pagina, edicion_id) VALUES
(13, 1, 3), (14, 2, 3);

-- Página 1: Portada broadsheet
INSERT INTO pautas (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
('Masthead',              1, 6, 0, 0, 13),
('Nota apertura',         4, 4, 0, 1, 13),
('Foto apertura',         4, 2, 4, 1, 13),
('Columna de opinión',    3, 2, 0, 5, 13),
('Avances interiores',    3, 2, 2, 5, 13),
('Pronóstico del tiempo', 3, 2, 4, 5, 13),
('Cintillo publicidad',   2, 6, 0, 8, 13);

-- Página 2: Interiores broadsheet
INSERT INTO pautas (descripcion_pauta, cuadros_alto, cuadros_ancho, ubicacion_cuadros_x, ubicacion_cuadros_y, pagina_id) VALUES
('Nota principal interior', 5, 4, 0, 0, 14),
('Foto interior',           3, 2, 4, 0, 14),
('Recuadro datos',          2, 2, 4, 3, 14),
('Nota secundaria',         3, 3, 0, 5, 14),
('Publicidad cuarto',       3, 3, 3, 5, 14),
('Pie de página',           2, 6, 0, 8, 14);

-- ============================================
-- Reset sequences
-- ============================================
SELECT setval('editions_id_seq', (SELECT MAX(id) FROM editions));
SELECT setval('pages_id_seq', (SELECT MAX(id) FROM pages));
SELECT setval('pautas_id_seq', (SELECT MAX(id) FROM pautas));
