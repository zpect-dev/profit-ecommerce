CREATE TABLE IF NOT EXISTS lin_art (
                                       co_lin VARCHAR(6) PRIMARY KEY,
    lin_des VARCHAR(60) NOT NULL
    );

CREATE TABLE IF NOT EXISTS sub_lin (
                                       co_subl VARCHAR(6) PRIMARY KEY,
    subl_des VARCHAR(60) NOT NULL ,
    co_lin VARCHAR(6) NOT NULL,

    CONSTRAINT fk_sub_lin_lin_art FOREIGN KEY (co_lin) REFERENCES lin_art(co_lin)
    );

CREATE TABLE IF NOT EXISTS cat_art (
                                       co_cat VARCHAR(6) PRIMARY KEY,
    cat_des VARCHAR(60) NOT NULL
    );

CREATE TABLE IF NOT EXISTS art (

    -- Datos espejo profit
    co_art VARCHAR(30) PRIMARY KEY,
    art_des VARCHAR(120) NOT NULL,
    stock_act DECIMAL(18, 5) NOT NULL DEFAULT 0,
    prec_vta1 DECIMAL(18, 5) NOT NULL DEFAULT 0,
    prec_vta2 DECIMAL(18, 5) NOT NULL DEFAULT 0,
    prec_vta3 DECIMAL(18, 5) NOT NULL DEFAULT 0,
    prec_vta4 DECIMAL(18, 5) NOT NULL DEFAULT 0,
    prec_vta5 DECIMAL(18, 5) NOT NULL DEFAULT 0,
    tipo_imp CHAR(1) NOT NULL,

    -- Relaciones
    co_lin VARCHAR(6),
    co_cat VARCHAR(6),
    co_subl VARCHAR(6),
    campo4 VARCHAR(60),

    -- Datos para la web
    image_url VARCHAR(255),
    is_active   BOOLEAN DEFAULT true,
    last_sync   TIMESTAMP DEFAULT NOW(),

    -- Constraints
    CONSTRAINT fk_art_line_art FOREIGN KEY (co_lin) REFERENCES lin_art(co_lin),
    CONSTRAINT fk_art_sub_lin FOREIGN KEY (co_subl) REFERENCES sub_lin(co_subl),
    CONSTRAINT fk_art_cat_art FOREIGN KEY (co_cat) REFERENCES cat_art(co_cat)
    );

CREATE TABLE IF NOT EXISTS descuen (
    co_desc VARCHAR(30),
    tipo_cli VARCHAR(6),
    tipo_desc VARCHAR(1),
    porc1 DECIMAL(18, 2)
    );

CREATE INDEX IF NOT EXISTS idx_art_line ON art(co_lin);
CREATE INDEX IF NOT EXISTS idx_art_art_des ON art(art_des);