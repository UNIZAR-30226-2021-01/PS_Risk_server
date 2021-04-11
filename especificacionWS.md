# COMUNICACIÓN CON WEBSOCKETS
Cuando un cliente se conecte a una sala se establecerá una conexión mediante WebSockets. Tanto los parámetros como resultados se envían en formato JSON. Las conexiones son a la URL base https://risk-servidor.herokuapp.com/.

## Establecer una conexión
Cuando un cliente quiera establecer una conexión WebSocket con una sala debe hacerlo en las siguientes URL.

- ## /crearSala
    Cuando un cliente quiera crear una nueva sala debe establecer una conexión esta URL y enviar el siguiente mensaje.
        
        { 
            "idUsuario":int, 
            "clave":string, 
            "tiempoTurno":int, 
            "nombreSala":string 
        }
    
    **idUsuario:** identificador numérico del usuario que quiere crear la sala.
    
    **clave:** hash SHA256 de la clave del usuario.

    **tiempoTurno:** tiempo de turno para establecer en la sala.

    **nombreSala:** nombre para establecer en la sala.
    
- ## /aceptarSala
    Cuando un cliente solicite unirse a una sala ya existente debe establecer una conexión a esta URL y enviar el siguiente mensaje.

        { 
            "idUsuario":int, 
            "clave":string, 
            "idSala":int 
        }
    
    **idUsuario:** identificador numérico del usuario que quiere entrar a la sala.
    
    **clave:** hash SHA256 de la clave del usuario.

    **idSala:** identificador numérico de la sala a la que se quiere unir.

## Mensajes
Durante el transcurso de la conexión tanto el servidor como los clientes se pueden enviar mensajes en formato JSON. Los mensajes estan detallados a continuación.

- ## Errores
    Si ocurre algun error durante el transcurso de la conexión se envía un error en formato JSON.

        {
            "_tipoMensaje":"e",
            "code": int,
            "err": string,
        }

    **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para los errores su valor es "e".

    **code:** Indica el tipo de error y la accion a tomar por el cliente. Puede tomar los siguientes valores:
    - 1: Ha ocurrido un error y la operación solicitada no se ha llevado a cabo, no se requiere ninguna acción.
    - 2: Ha ocurrido un error verificando al usuario y la operación solicitada no se ha llevado a cabo, se requiere cerrar la sesión del usuario.
    - 3: Ha ocurrido un error en una la sala de espera, se requiere cerrar la conexión con la sala.

    **err:** Explica que error ha ocurrido.

- ## Información de sala
    El servidor puede enviar a los distintos participantes de una sala el siguiente mensaje cuando necesite comunicar cambios en la información de la sala.
    
        { 
            "_tipoMensaje":"d", 
            "tiempoTurno":int, 
            "nombreSala":string, 
            "idSala":int
            "jugadores": [ { "id":int, "nombre":string, "icono":int, "aspecto":int } ] 
        }

    **_tipoMensaje:** se utiliza para ayudar a la decodificación por parte de los clientes. Para este mensaje su valor es "d".

    **tiempoTurno:** tiempo de turno configurado en la sala.

    **nombreSala:** nombre de la sala.

    **idSala:** identificador numérico de la sala.

    **jugadores:** lista de jugadores que se encuetren en la sala.
    - **id:** identificador numérico del jugador.
    - **nombre:** nombre del jugador.
    - **icono:** icono del jugador.
    - **aspecto:** aspecto del jugador.

- ## Invitar un participante a la sala
    El creador de la sala puede solicitar que se invite a un usuario a la sala. Debe enviar el siguiente mensaje.
        
        { 
            "idInvitado":int, 
            "tipo":"Invitar" 
        } 

    **idInvitado:** indentificador numérico del usuario a invitar a la sala.
    
    **tipo:** se utiliza para ayudar a la decodificación por parte del servidor. Para este mensaje su valor es "Invitar"

- ## Iniciar partida
    El creador de la sala puede solicitar el inicio de la partida. Debe enviar el siguiente mensaje.

        {
            "tipo":"Iniciar"
        }

    **tipo:** se utiliza para ayudar a la decodificación por parte del servidor. Para este mensaje su valor es "Iniciar"

<!--
wss::// /unirsePartida
    JSON { "idUsuario":int, "clave":string, "idSala":int }

MENSAJE CUANDO COMIENZA PARTIDA
    <- partidaCompleta {
        "_tipoMensaje": "p",
        "tiempoTurno": int,
        "nombreSala": string,
        "turnoActual": int,
        "fase": int,
        "jugadores": [ { "id":int, "nombre":string, "icono":int, "aspecto":int, "sigueVivo":bool } ] ,
        "listaTerritorios": [ { "numJugador":int, "tropas":int } ]
    }

MENSAJES DURANTE LA PARTIDA
    <- partidaCompleta

    <-> AccionPartida {
    	"_tipoMensaje": "a",
        "origen":int,
        "objetivo":int,
        "numTropasOrigen":int,
        "numTropasObjetivo":int,
        "tipo":string -> [Reclutar, Mover, Atacar]
    }

    <- { "_tipoMensaje":"e", "code":int , "err":string }

    <- FinPartida {
    	"_tipoMensaje":"f"
        "ganador":string,
        "riskos":int
    }
-->