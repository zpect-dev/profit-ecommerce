CREATE TABLE IF NOT EXISTS tipo_cli (
    tip_cli VARCHAR(50) PRIMARY KEY,
    des_tipo VARCHAR(255) NOT NULL,
    precio_a VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS clientes (
    co_cli VARCHAR(50) PRIMARY KEY,
    tipo VARCHAR(50) NOT NULL,
    cli_des VARCHAR(255) NOT NULL,
    rif VARCHAR(50),
    inactivo BOOLEAN DEFAULT FALSE,
    login NUMERIC(15,2) DEFAULT 0,
    mont_cre NUMERIC(15,2) DEFAULT 0,
    CONSTRAINT fk_tipo_cli FOREIGN KEY (tipo) REFERENCES tipo_cli(tip_cli) ON DELETE RESTRICT
);
