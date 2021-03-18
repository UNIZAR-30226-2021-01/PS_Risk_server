CREATE TABLE aspecto (
    id_aspecto SERIAL PRIMARY KEY,
    precio INT NOT NULL CHECK (precio >= 0)
);

CREATE TABLE icono (
    id_icono SERIAL PRIMARY KEY,
    precio INT NOT NULL CHECK (precio >= 0)
);

CREATE TABLE usuario (
    id_usuario SERIAL PRIMARY KEY,
    aspecto INT NOT NULL,
    icono INT NOT NULL,
    nombre VARCHAR(20) UNIQUE NOT NULL,
    correo VARCHAR(20) UNIQUE NOT NULL,
    clave VARCHAR(64) NOT NULL,
    riskos INT NOT NULL CHECK (riskos >= 0),
    recibeCorreos BOOLEAN NOT NULL,
    FOREIGN KEY (aspecto) REFERENCES aspecto(id_aspecto),
    FOREIGN KEY (icono) REFERENCES icono(id_icono) 
);

CREATE TABLE aspectosComprados (
	id_usuario INT,
	id_aspecto INT,
	FOREIGN KEY (id_usuario) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
	FOREIGN KEY (id_aspecto) REFERENCES aspecto(id_aspecto) ON DELETE CASCADE,
	PRIMARY KEY (id_usuario, id_aspecto)
);

CREATE TABLE iconosComprados (
	id_usuario INT,
	id_icono INT,
	FOREIGN KEY (id_usuario) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
	FOREIGN KEY (id_icono) REFERENCES icono(id_icono) ON DELETE CASCADE,
	PRIMARY KEY (id_usuario, id_icono)
);

CREATE TABLE esAmigo (
    id_usuario1 INT,
    id_usuario2 INT CHECK (id_usuario1 != id_usuario2),
    FOREIGN KEY (id_usuario1) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    FOREIGN KEY (id_usuario2) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    PRIMARY KEY (id_usuario1, id_usuario2)
);

CREATE TABLE partida (
    id_partida SERIAL PRIMARY KEY,
    json_estado JSON
);

CREATE TABLE notificacion (
    id_notificacion SERIAL PRIMARY KEY,
    id_usuarioEnvia INT,
    id_usuarioRecibe INT NOT NULL CHECK (id_usuarioRecibe != id_usuarioEnvia),
    id_partida INT,
    tipo INT NOT NULL,
    FOREIGN KEY (id_usuarioEnvia) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    FOREIGN KEY (id_usuarioRecibe) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    FOREIGN KEY (id_partida) REFERENCES partida(id_partida) ON DELETE CASCADE
);

CREATE TABLE juega (
    id_partida INT,
    id_usuario INT,
    FOREIGN KEY (id_partida) REFERENCES partida(id_partida) ON DELETE CASCADE,
    FOREIGN KEY (id_usuario) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    PRIMARY KEY (id_partida, id_usuario)
);