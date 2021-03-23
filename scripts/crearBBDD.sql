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
    nombre VARCHAR(20),
    json_estado JSON
);

CREATE TABLE notificacionTurno (
    id_recibe INT,
    id_envia INT,
    FOREIGN KEY (id_recibe) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    FOREIGN KEY (id_envia) REFERENCES partida(id_partida) ON DELETE CASCADE,
    PRIMARY KEY (id_recibe, id_envia)
);

CREATE TABLE solicitudAmistad (
    id_recibe INT,
    id_envia INT CHECK (id_recibe != id_envia),
    FOREIGN KEY (id_recibe) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    FOREIGN KEY (id_envia) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    PRIMARY KEY (id_recibe, id_envia)
);

CREATE TABLE invitacionPartida (
    id_recibe INT,
    id_envia INT,
    FOREIGN KEY (id_recibe) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    FOREIGN KEY (id_envia) REFERENCES partida(id_partida) ON DELETE CASCADE,
    PRIMARY KEY (id_recibe, id_envia)
);

CREATE TABLE juega (
    id_partida INT,
    id_usuario INT,
    FOREIGN KEY (id_partida) REFERENCES partida(id_partida) ON DELETE CASCADE,
    FOREIGN KEY (id_usuario) REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    PRIMARY KEY (id_partida, id_usuario)
);

INSERT INTO icono (id_icono, precio) VALUES (0, 0);
INSERT INTO aspecto (id_aspecto, precio) VALUES (0, 0);